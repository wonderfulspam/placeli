package sources

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/user/placeli/internal/models"
)

// OSMImporter handles OpenStreetMap data exports
type OSMImporter struct{}

// OSMNode represents a basic OSM node with POI data
type OSMNode struct {
	ID   int64             `json:"id"`
	Lat  float64           `json:"lat"`
	Lon  float64           `json:"lon"`
	Tags map[string]string `json:"tags"`
}

// OSMExport represents an OSM export file
type OSMExport struct {
	Version   string    `json:"version"`
	Generator string    `json:"generator"`
	Elements  []OSMNode `json:"elements"`
}

func (oi *OSMImporter) Name() string {
	return "OpenStreetMap"
}

func (oi *OSMImporter) SupportedFormats() []string {
	return []string{"json", "csv"}
}

func (oi *OSMImporter) ImportFromFile(filePath string) ([]*models.Place, error) {
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
	format := "json"
	if strings.HasSuffix(strings.ToLower(filePath), ".csv") {
		format = "csv"
	}

	return oi.ImportFromData(data, format)
}

func (oi *OSMImporter) ImportFromData(data []byte, format string) ([]*models.Place, error) {
	switch format {
	case "json":
		return oi.parseJSON(data)
	case "csv":
		return oi.parseCSV(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (oi *OSMImporter) parseJSON(data []byte) ([]*models.Place, error) {
	var osmExport OSMExport
	if err := json.Unmarshal(data, &osmExport); err != nil {
		return nil, fmt.Errorf("failed to parse OSM JSON: %w", err)
	}

	var places []*models.Place
	for _, element := range osmExport.Elements {
		place := oi.convertOSMNode(element)
		if place != nil {
			places = append(places, place)
		}
	}

	return places, nil
}

func (oi *OSMImporter) parseCSV(data []byte) ([]*models.Place, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	// Expect header row with columns: name, lat, lon, type, description
	header := records[0]
	nameCol, latCol, lonCol, typeCol, descCol := -1, -1, -1, -1, -1
	
	for i, col := range header {
		switch strings.ToLower(col) {
		case "name", "title":
			nameCol = i
		case "lat", "latitude":
			latCol = i
		case "lon", "lng", "longitude":
			lonCol = i
		case "type", "category":
			typeCol = i
		case "description", "desc", "notes":
			descCol = i
		}
	}

	if nameCol == -1 || latCol == -1 || lonCol == -1 {
		return nil, fmt.Errorf("CSV must have name, lat, and lon columns")
	}

	var places []*models.Place
	for i, record := range records[1:] { // Skip header
		if len(record) <= nameCol || len(record) <= latCol || len(record) <= lonCol {
			continue // Skip incomplete records
		}

		name := strings.TrimSpace(record[nameCol])
		if name == "" {
			continue
		}

		lat, err1 := strconv.ParseFloat(record[latCol], 64)
		lon, err2 := strconv.ParseFloat(record[lonCol], 64)
		if err1 != nil || err2 != nil {
			continue // Skip records with invalid coordinates
		}

		// Create OSM node structure for conversion
		osmNode := OSMNode{
			ID:   int64(i), // Use row index as temporary ID
			Lat:  lat,
			Lon:  lon,
			Tags: map[string]string{
				"name": name,
			},
		}

		if typeCol != -1 && len(record) > typeCol {
			osmNode.Tags["amenity"] = record[typeCol]
		}

		if descCol != -1 && len(record) > descCol {
			osmNode.Tags["description"] = record[descCol]
		}

		place := oi.convertOSMNode(osmNode)
		if place != nil {
			places = append(places, place)
		}
	}

	return places, nil
}

func (oi *OSMImporter) convertOSMNode(node OSMNode) *models.Place {
	name := node.Tags["name"]
	if name == "" {
		return nil // Skip nodes without names
	}

	coords := models.Coordinates{
		Lat: node.Lat,
		Lng: node.Lon,
	}

	// Generate unique IDs
	sourceData := fmt.Sprintf("osm|%d|%s|%f,%f", 
		node.ID,
		name,
		coords.Lat, 
		coords.Lng)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(sourceData)))
	placeID := fmt.Sprintf("osm_%d", node.ID)

	now := time.Now()
	place := &models.Place{
		ID:          oi.generateID(placeID),
		PlaceID:     placeID,
		Name:        name,
		Address:     oi.buildAddress(node.Tags),
		Coordinates: coords,
		Categories:  oi.extractCategories(node.Tags),
		Phone:       node.Tags["phone"],
		Website:     node.Tags["website"],
		UserNotes:   node.Tags["description"],
		UserTags:    []string{},
		Photos:      []models.Photo{},
		Reviews:     []models.Review{},
		CustomFields: map[string]interface{}{
			"imported_from": "openstreetmap",
			"import_date":   now.Format(time.RFC3339),
			"osm_id":        node.ID,
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		ImportedAt: &now,
		SourceHash: sourceHash,
	}

	// Add OSM-specific tags as custom fields
	for key, value := range node.Tags {
		if key != "name" && key != "phone" && key != "website" && 
		   key != "description" && !strings.HasPrefix(key, "addr:") {
			place.CustomFields[fmt.Sprintf("osm_%s", key)] = value
		}
	}

	return place
}

func (oi *OSMImporter) buildAddress(tags map[string]string) string {
	var addressParts []string
	
	// Common OSM address tags
	if housenumber := tags["addr:housenumber"]; housenumber != "" {
		addressParts = append(addressParts, housenumber)
	}
	if street := tags["addr:street"]; street != "" {
		addressParts = append(addressParts, street)
	}
	if city := tags["addr:city"]; city != "" {
		addressParts = append(addressParts, city)
	}
	if postcode := tags["addr:postcode"]; postcode != "" {
		addressParts = append(addressParts, postcode)
	}
	if country := tags["addr:country"]; country != "" {
		addressParts = append(addressParts, country)
	}
	
	return strings.Join(addressParts, ", ")
}

func (oi *OSMImporter) extractCategories(tags map[string]string) []string {
	var categories []string
	
	// Common OSM category tags
	categoryTags := []string{"amenity", "shop", "tourism", "leisure", "craft", "office"}
	
	for _, tagKey := range categoryTags {
		if value := tags[tagKey]; value != "" {
			// Convert OSM values to readable categories
			category := oi.humanizeCategory(tagKey, value)
			categories = append(categories, category)
		}
	}
	
	return categories
}

func (oi *OSMImporter) humanizeCategory(tagKey, value string) string {
	// Convert OSM tag values to human-readable categories
	switch tagKey {
	case "amenity":
		switch value {
		case "restaurant":
			return "Restaurant"
		case "cafe":
			return "Cafe"
		case "bar", "pub":
			return "Bar"
		case "bank":
			return "Bank"
		case "hospital":
			return "Hospital"
		case "school":
			return "School"
		default:
			return strings.Title(value)
		}
	case "shop":
		return "Shopping - " + strings.Title(value)
	case "tourism":
		return "Tourism - " + strings.Title(value)
	case "leisure":
		return "Leisure - " + strings.Title(value)
	default:
		return strings.Title(value)
	}
}

func (oi *OSMImporter) generateID(placeID string) string {
	hash := sha256.Sum256([]byte(placeID))
	return fmt.Sprintf("%x", hash)[:12]
}