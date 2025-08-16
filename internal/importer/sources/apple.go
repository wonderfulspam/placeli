package sources

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/user/placeli/internal/models"
)

// AppleImporter handles KML and GPX files from Apple Maps
type AppleImporter struct{}

// KML structures for parsing Apple Maps exports
type KML struct {
	XMLName   xml.Name    `xml:"kml"`
	Document  KMLDocument `xml:"Document"`
	Placemarks []KMLPlacemark `xml:"Document>Placemark"`
}

type KMLDocument struct {
	Name        string        `xml:"name"`
	Placemarks  []KMLPlacemark `xml:"Placemark"`
}

type KMLPlacemark struct {
	Name        string      `xml:"name"`
	Description string      `xml:"description"`
	Point       KMLPoint    `xml:"Point"`
	Address     string      `xml:"address"`
	PhoneNumber string      `xml:"phoneNumber"`
	ExtendedData KMLExtendedData `xml:"ExtendedData"`
}

type KMLPoint struct {
	Coordinates string `xml:"coordinates"`
}

type KMLExtendedData struct {
	Data []KMLData `xml:"Data"`
}

type KMLData struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value"`
}

// GPX structures for parsing GPX files
type GPX struct {
	XMLName    xml.Name    `xml:"gpx"`
	Waypoints  []GPXWaypoint `xml:"wpt"`
}

type GPXWaypoint struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Name string  `xml:"name"`
	Desc string  `xml:"desc"`
	Type string  `xml:"type"`
	Time string  `xml:"time"`
}

func (ai *AppleImporter) Name() string {
	return "Apple Maps"
}

func (ai *AppleImporter) SupportedFormats() []string {
	return []string{"kml", "kmz", "gpx"}
}

func (ai *AppleImporter) ImportFromFile(filePath string) ([]*models.Place, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine format from file extension
	format := "kml"
	if strings.HasSuffix(strings.ToLower(filePath), ".gpx") {
		format = "gpx"
	} else if strings.HasSuffix(strings.ToLower(filePath), ".kmz") {
		return nil, fmt.Errorf("KMZ files not yet supported, please extract to KML first")
	}

	return ai.ImportFromData(data, format)
}

func (ai *AppleImporter) ImportFromData(data []byte, format string) ([]*models.Place, error) {
	switch format {
	case "kml":
		return ai.parseKML(data)
	case "gpx":
		return ai.parseGPX(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (ai *AppleImporter) parseKML(data []byte) ([]*models.Place, error) {
	var kml KML
	if err := xml.Unmarshal(data, &kml); err != nil {
		return nil, fmt.Errorf("failed to parse KML: %w", err)
	}

	var places []*models.Place
	
	// Handle placemarks directly under document
	for _, placemark := range kml.Document.Placemarks {
		place := ai.convertKMLPlacemark(placemark)
		if place != nil {
			places = append(places, place)
		}
	}
	
	// Handle placemarks at root level
	for _, placemark := range kml.Placemarks {
		place := ai.convertKMLPlacemark(placemark)
		if place != nil {
			places = append(places, place)
		}
	}

	return places, nil
}

func (ai *AppleImporter) parseGPX(data []byte) ([]*models.Place, error) {
	var gpx GPX
	if err := xml.Unmarshal(data, &gpx); err != nil {
		return nil, fmt.Errorf("failed to parse GPX: %w", err)
	}

	var places []*models.Place
	for _, waypoint := range gpx.Waypoints {
		place := ai.convertGPXWaypoint(waypoint)
		if place != nil {
			places = append(places, place)
		}
	}

	return places, nil
}

func (ai *AppleImporter) convertKMLPlacemark(placemark KMLPlacemark) *models.Place {
	if placemark.Name == "" {
		return nil // Skip placemarks without names
	}

	// Parse coordinates
	coords := ai.parseCoordinates(placemark.Point.Coordinates)
	if coords.Lat == 0 && coords.Lng == 0 {
		return nil // Skip placemarks without valid coordinates
	}

	// Generate unique IDs
	sourceData := fmt.Sprintf("apple|%s|%s|%f,%f", 
		placemark.Name, 
		placemark.Address,
		coords.Lat, 
		coords.Lng)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(sourceData)))
	placeID := fmt.Sprintf("apple_%s", sourceHash[:16])

	now := time.Now()
	place := &models.Place{
		ID:          ai.generateID(placeID),
		PlaceID:     placeID,
		Name:        placemark.Name,
		Address:     placemark.Address,
		Coordinates: coords,
		Categories:  ai.extractCategories(placemark),
		Phone:       placemark.PhoneNumber,
		UserNotes:   placemark.Description,
		UserTags:    []string{},
		Photos:      []models.Photo{},
		Reviews:     []models.Review{},
		CustomFields: map[string]interface{}{
			"imported_from": "apple_maps",
			"import_date":   now.Format(time.RFC3339),
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		ImportedAt: &now,
		SourceHash: sourceHash,
	}

	// Add extended data as custom fields
	for _, data := range placemark.ExtendedData.Data {
		if data.Name != "" && data.Value != "" {
			place.CustomFields[fmt.Sprintf("apple_%s", data.Name)] = data.Value
		}
	}

	return place
}

func (ai *AppleImporter) convertGPXWaypoint(waypoint GPXWaypoint) *models.Place {
	if waypoint.Name == "" {
		return nil // Skip waypoints without names
	}

	coords := models.Coordinates{
		Lat: waypoint.Lat,
		Lng: waypoint.Lon,
	}

	// Generate unique IDs
	sourceData := fmt.Sprintf("gpx|%s|%s|%f,%f", 
		waypoint.Name, 
		waypoint.Desc,
		coords.Lat, 
		coords.Lng)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(sourceData)))
	placeID := fmt.Sprintf("gpx_%s", sourceHash[:16])

	now := time.Now()
	place := &models.Place{
		ID:          ai.generateID(placeID),
		PlaceID:     placeID,
		Name:        waypoint.Name,
		Address:     "", // GPX typically doesn't have addresses
		Coordinates: coords,
		Categories:  ai.extractGPXCategories(waypoint.Type),
		UserNotes:   waypoint.Desc,
		UserTags:    []string{},
		Photos:      []models.Photo{},
		Reviews:     []models.Review{},
		CustomFields: map[string]interface{}{
			"imported_from": "gpx",
			"import_date":   now.Format(time.RFC3339),
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		ImportedAt: &now,
		SourceHash: sourceHash,
	}

	// Add GPX-specific fields
	if waypoint.Type != "" {
		place.CustomFields["gpx_type"] = waypoint.Type
	}
	if waypoint.Time != "" {
		place.CustomFields["gpx_time"] = waypoint.Time
	}

	return place
}

func (ai *AppleImporter) parseCoordinates(coordStr string) models.Coordinates {
	// KML coordinates are in longitude,latitude,altitude format
	parts := strings.Split(strings.TrimSpace(coordStr), ",")
	if len(parts) < 2 {
		return models.Coordinates{}
	}

	lng, err1 := strconv.ParseFloat(parts[0], 64)
	lat, err2 := strconv.ParseFloat(parts[1], 64)
	
	if err1 != nil || err2 != nil {
		return models.Coordinates{}
	}

	return models.Coordinates{Lat: lat, Lng: lng}
}

func (ai *AppleImporter) extractCategories(placemark KMLPlacemark) []string {
	var categories []string
	
	// Look for category information in extended data
	for _, data := range placemark.ExtendedData.Data {
		if strings.ToLower(data.Name) == "category" || 
		   strings.ToLower(data.Name) == "type" {
			if data.Value != "" {
				categories = append(categories, data.Value)
			}
		}
	}
	
	// If no categories found, try to infer from name or description
	if len(categories) == 0 {
		if strings.Contains(strings.ToLower(placemark.Name), "restaurant") ||
		   strings.Contains(strings.ToLower(placemark.Description), "restaurant") {
			categories = append(categories, "Restaurant")
		}
	}
	
	return categories
}

func (ai *AppleImporter) extractGPXCategories(waypointType string) []string {
	if waypointType == "" {
		return []string{}
	}
	return []string{waypointType}
}

func (ai *AppleImporter) generateID(placeID string) string {
	hash := sha256.Sum256([]byte(placeID))
	return fmt.Sprintf("%x", hash)[:12]
}