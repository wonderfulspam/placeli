package sources

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/user/placeli/internal/models"
)

// TakeoutImporter handles imports from Google Takeout exports
type TakeoutImporter struct{}

// Types for parsing Google Takeout "Maps (your places)" format
type TakeoutList struct {
	Name     string         `json:"name"`
	Features []TakeoutPlace `json:"features"`
}

type TakeoutPlace struct {
	Geometry   TakeoutGeometry   `json:"geometry"`
	Properties TakeoutProperties `json:"properties"`
}

type TakeoutGeometry struct {
	Coordinates []float64 `json:"coordinates"`
	Type        string    `json:"type"`
}

type TakeoutLocation struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	CountryCode string `json:"country_code"`
}

type TakeoutProperties struct {
	Date          string           `json:"date"`
	GoogleMapsURL string           `json:"google_maps_url"`
	Location      *TakeoutLocation `json:"location"`
	Comment       string           `json:"Comment"`
	// Legacy fields for older export formats
	Name            string                 `json:"name"`
	Address         string                 `json:"address"`
	Categories      []string               `json:"categories"`
	Rating          float32                `json:"rating"`
	ReviewCount     int                    `json:"review_count"`
	PriceLevel      int                    `json:"price_level"`
	Phone           string                 `json:"phone"`
	Website         string                 `json:"website"`
	Hours           map[string]interface{} `json:"hours"`
	Description     string                 `json:"description"`
	PlaceID         string                 `json:"place_id"`
	OtherCandidates []interface{}          `json:"Other candidates"`
}

// Helper functions moved from internal/importer/takeout.go

// generateID generates a unique ID from a place ID
func generateID(placeID string) string {
	hash := sha256.Sum256([]byte(placeID))
	return fmt.Sprintf("%x", hash)[:12]
}

// generateHash generates SHA256 hash of data
func generateHash(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// extractFromGoogleMapsURL extracts coordinates and place name from a Google Maps URL
func extractFromGoogleMapsURL(gmapsURL string) (lat, lng float64, placeName string) {
	if gmapsURL == "" {
		return 0, 0, ""
	}

	parsedURL, err := url.Parse(gmapsURL)
	if err != nil {
		return 0, 0, ""
	}

	q := parsedURL.Query().Get("q")
	if q == "" {
		return 0, 0, ""
	}

	// Try to parse as coordinates (e.g., "56.14993739674781,12.572669684886932")
	coordPattern := regexp.MustCompile(`^(-?\d+\.?\d*),(-?\d+\.?\d*)$`)
	if matches := coordPattern.FindStringSubmatch(q); len(matches) == 3 {
		lat, _ = strconv.ParseFloat(matches[1], 64)
		lng, _ = strconv.ParseFloat(matches[2], 64)
		return lat, lng, ""
	}

	// Otherwise, treat as place name (decode URL encoding)
	placeName, _ = url.QueryUnescape(q)
	// Remove any ftid parameter info
	if idx := strings.Index(placeName, "&"); idx > 0 {
		placeName = placeName[:idx]
	}
	return 0, 0, placeName
}

// convertTakeoutPlace converts a TakeoutPlace to our models.Place
func convertTakeoutPlace(tp TakeoutPlace) *models.Place {
	var coordinates models.Coordinates
	if len(tp.Geometry.Coordinates) >= 2 {
		coordinates = models.Coordinates{
			Lng: tp.Geometry.Coordinates[0],
			Lat: tp.Geometry.Coordinates[1],
		}
	}

	// Extract name and address from either nested location or legacy fields
	var name, address string
	if tp.Properties.Location != nil {
		name = tp.Properties.Location.Name
		address = tp.Properties.Location.Address
	} else {
		// Fall back to legacy fields
		name = tp.Properties.Name
		address = tp.Properties.Address
	}

	// If no location info but we have a Google Maps URL, try to extract from it
	if name == "" && address == "" && tp.Properties.GoogleMapsURL != "" {
		lat, lng, placeName := extractFromGoogleMapsURL(tp.Properties.GoogleMapsURL)
		if lat != 0 && lng != 0 {
			// We have coordinates from the URL
			coordinates.Lat = lat
			coordinates.Lng = lng
			name = fmt.Sprintf("Saved Place (%.6f, %.6f)", lat, lng)
		} else if placeName != "" {
			// We have a place name from the URL
			name = placeName
		}
	}

	// Skip only if we have absolutely no useful data
	if name == "" && address == "" && coordinates.Lat == 0 && coordinates.Lng == 0 {
		return nil
	}

	var placeID string
	if tp.Properties.PlaceID != "" {
		placeID = tp.Properties.PlaceID
	} else {
		hash := sha256.Sum256([]byte(name + address))
		placeID = fmt.Sprintf("takeout_%x", hash)[:16]
	}

	var hours string
	if tp.Properties.Hours != nil {
		hoursData, _ := json.Marshal(tp.Properties.Hours)
		hours = string(hoursData)
	}

	// Generate source hash for duplicate detection
	sourceData := fmt.Sprintf("%s|%s|%f,%f|%s",
		name,
		address,
		coordinates.Lat,
		coordinates.Lng,
		placeID)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(sourceData)))

	now := time.Now()
	place := &models.Place{
		ID:          generateID(placeID),
		PlaceID:     placeID,
		Name:        name,
		Address:     address,
		Coordinates: coordinates,
		Categories:  tp.Properties.Categories,
		Rating:      tp.Properties.Rating,
		UserRatings: tp.Properties.ReviewCount,
		PriceLevel:  tp.Properties.PriceLevel,
		Hours:       hours,
		Phone:       tp.Properties.Phone,
		Website:     tp.Properties.Website,
		Photos:      []models.Photo{},
		Reviews:     []models.Review{},
		UserNotes:   tp.Properties.Description,
		UserTags:    []string{},
		CustomFields: map[string]interface{}{
			"google_maps_url": tp.Properties.GoogleMapsURL,
			"imported_from":   "takeout",
			"import_date":     now.Format(time.RFC3339),
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		ImportedAt: &now,
		SourceHash: sourceHash,
	}

	// Add country code if available
	if tp.Properties.Location != nil && tp.Properties.Location.CountryCode != "" {
		place.CustomFields["country_code"] = tp.Properties.Location.CountryCode
	}

	// Add import date if available
	if tp.Properties.Date != "" {
		place.CustomFields["saved_date"] = tp.Properties.Date
	}

	return place
}

// parseListFile parses a JSON file containing takeout places
func parseListFile(filePath string) ([]*models.Place, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var takeoutList TakeoutList
	if err := json.Unmarshal(data, &takeoutList); err != nil {
		return nil, err
	}

	var places []*models.Place
	for _, feature := range takeoutList.Features {
		place := convertTakeoutPlace(feature)
		if place != nil {
			places = append(places, place)
		}
	}

	return places, nil
}

// importFromTakeout imports places from an extracted takeout directory
func importFromTakeout(takeoutPath string) ([]*models.Place, error) {
	var places []*models.Place

	err := filepath.Walk(takeoutPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".json") && strings.Contains(path, "Saved Places") {
			listPlaces, err := parseListFile(path)
			if err != nil {
				fmt.Printf("Warning: Could not parse %s: %v\n", path, err)
				return nil
			}
			places = append(places, listPlaces...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return places, nil
}

// Name returns the name of this import source
func (t *TakeoutImporter) Name() string {
	return "Google Takeout"
}

// SupportedFormats returns the file formats this source can handle
func (t *TakeoutImporter) SupportedFormats() []string {
	return []string{"zip", "json", "csv", "directory"}
}

// ImportFromFile imports places from a Google Takeout export
func (t *TakeoutImporter) ImportFromFile(filePath string) ([]*models.Place, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Handle directory (extracted takeout)
	if info.IsDir() {
		return importFromTakeout(filePath)
	}

	// Handle ZIP file
	if strings.HasSuffix(strings.ToLower(filePath), ".zip") {
		return t.importFromZip(filePath)
	}

	// Handle direct JSON file (Saved Places.json)
	if strings.HasSuffix(strings.ToLower(filePath), ".json") {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return t.ImportFromData(data, "json")
	}

	// Handle CSV file (Saved Places CSV from Google Takeout "Saved")
	if strings.HasSuffix(strings.ToLower(filePath), ".csv") {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		return t.ImportFromData(data, "csv")
	}

	return nil, fmt.Errorf("unsupported file type: %s", filePath)
}

// ImportFromData imports places from raw data
func (t *TakeoutImporter) ImportFromData(data []byte, format string) ([]*models.Place, error) {
	switch format {
	case "json":
		return t.importFromJSON(data)
	case "csv":
		return t.importFromCSV(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// importFromJSON handles JSON format imports
func (t *TakeoutImporter) importFromJSON(data []byte) ([]*models.Place, error) {
	// Try to parse as Takeout "Maps (your places)" format
	var takeoutList TakeoutList
	if err := json.Unmarshal(data, &takeoutList); err == nil && len(takeoutList.Features) > 0 {
		return t.parseTakeoutPlaces(takeoutList.Features), nil
	}

	// Try to parse as Takeout "Saved" format (Lists)
	var savedLists SavedLists
	if err := json.Unmarshal(data, &savedLists); err == nil && len(savedLists.Lists) > 0 {
		return t.parseSavedLists(savedLists.Lists), nil
	}

	// Try single list format
	var singleList SavedList
	if err := json.Unmarshal(data, &singleList); err == nil && len(singleList.Places) > 0 {
		return t.parseSavedList(&singleList), nil
	}

	return nil, fmt.Errorf("unrecognized Google Takeout JSON format")
}

// importFromCSV handles CSV format imports (Google Takeout "Saved" export)
func (t *TakeoutImporter) importFromCSV(data []byte) ([]*models.Place, error) {
	reader := csv.NewReader(bytes.NewReader(data))

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Map headers to indices
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.TrimSpace(h)] = i
	}

	var places []*models.Place

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	for _, record := range records {
		place := t.parseCSVRecord(record, headerMap)
		if place != nil {
			places = append(places, place)
		}
	}

	return places, nil
}

// parseCSVRecord parses a single CSV record into a Place
func (t *TakeoutImporter) parseCSVRecord(record []string, headers map[string]int) *models.Place {
	// Common CSV column names from Google Takeout "Saved"
	name := t.getCSVField(record, headers, "Title", "Name", "Place Name")
	address := t.getCSVField(record, headers, "Address", "Location")
	url := t.getCSVField(record, headers, "URL", "Link", "Google Maps URL")
	note := t.getCSVField(record, headers, "Comment", "Note", "Description")
	listName := t.getCSVField(record, headers, "List", "Collection", "Folder")

	// Try to get coordinates
	lat := t.getCSVFloat(record, headers, "Latitude", "Lat")
	lng := t.getCSVFloat(record, headers, "Longitude", "Lng", "Long")

	// Skip empty records
	if name == "" && address == "" && url == "" {
		return nil
	}

	// Extract coordinates from URL if not provided
	if lat == 0 && lng == 0 && url != "" {
		extractedLat, extractedLng, extractedName := extractFromGoogleMapsURL(url)
		if extractedLat != 0 && extractedLng != 0 {
			lat = extractedLat
			lng = extractedLng
		}
		if name == "" && extractedName != "" {
			name = extractedName
		}
	}

	// Generate a place ID
	placeID := fmt.Sprintf("saved_%x", generateHash([]byte(name+address+url)))[:16]

	now := time.Now()
	place := &models.Place{
		ID:      generateID(placeID),
		PlaceID: placeID,
		Name:    name,
		Address: address,
		Coordinates: models.Coordinates{
			Lat: lat,
			Lng: lng,
		},
		UserNotes: note,
		UserTags:  []string{},
		CustomFields: map[string]interface{}{
			"imported_from": "takeout_saved_csv",
			"import_date":   now.Format(time.RFC3339),
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		ImportedAt: &now,
	}

	if url != "" {
		place.CustomFields["google_maps_url"] = url
	}
	if listName != "" {
		place.UserTags = append(place.UserTags, listName)
		place.CustomFields["original_list"] = listName
	}

	// Generate source hash for duplicate detection
	sourceData := fmt.Sprintf("%s|%s|%f,%f|%s",
		name,
		address,
		lat,
		lng,
		placeID)
	place.SourceHash = fmt.Sprintf("%x", generateHash([]byte(sourceData)))

	return place
}

// getCSVField tries multiple column names and returns the first non-empty value
func (t *TakeoutImporter) getCSVField(record []string, headers map[string]int, names ...string) string {
	for _, name := range names {
		if idx, ok := headers[name]; ok && idx < len(record) {
			if val := strings.TrimSpace(record[idx]); val != "" {
				return val
			}
		}
	}
	return ""
}

// getCSVFloat tries multiple column names and returns the first valid float
func (t *TakeoutImporter) getCSVFloat(record []string, headers map[string]int, names ...string) float64 {
	for _, name := range names {
		if idx, ok := headers[name]; ok && idx < len(record) {
			if val, err := strconv.ParseFloat(strings.TrimSpace(record[idx]), 64); err == nil {
				return val
			}
		}
	}
	return 0
}

// importFromZip extracts and imports places from a Takeout ZIP file
func (t *TakeoutImporter) importFromZip(zipPath string) ([]*models.Place, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	var allPlaces []*models.Place

	for _, file := range reader.File {
		// Look for relevant JSON files
		if !strings.HasSuffix(file.Name, ".json") {
			continue
		}

		// Check for Maps data
		if strings.Contains(file.Name, "Maps") || strings.Contains(file.Name, "Saved") {
			rc, err := file.Open()
			if err != nil {
				continue
			}

			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			places, err := t.ImportFromData(data, "json")
			if err == nil && len(places) > 0 {
				allPlaces = append(allPlaces, places...)
			}
		}
	}

	return allPlaces, nil
}

// parseTakeoutPlaces converts Takeout places to our model
func (t *TakeoutImporter) parseTakeoutPlaces(features []TakeoutPlace) []*models.Place {
	var places []*models.Place
	for _, feature := range features {
		place := convertTakeoutPlace(feature)
		if place != nil {
			places = append(places, place)
		}
	}
	return places
}

// SavedLists represents the structure of Takeout "Saved" lists
type SavedLists struct {
	Lists []SavedList `json:"lists"`
}

// SavedList represents a single saved list
type SavedList struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Places      []SavedPlace `json:"places"`
	UpdatedAt   string       `json:"updated_at"`
	CreatedAt   string       `json:"created_at"`
}

// SavedPlace represents a place in a saved list
type SavedPlace struct {
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	PlaceID     string      `json:"place_id"`
	GoogleURL   string      `json:"google_maps_url"`
	Coordinates Coordinates `json:"coordinates"`
	Categories  []string    `json:"categories"`
	Note        string      `json:"note"`
	AddedAt     string      `json:"added_at"`
}

// Coordinates for saved places
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// parseSavedLists converts multiple saved lists to places
func (t *TakeoutImporter) parseSavedLists(lists []SavedList) []*models.Place {
	var places []*models.Place
	for _, list := range lists {
		listPlaces := t.parseSavedList(&list)
		places = append(places, listPlaces...)
	}
	return places
}

// parseSavedList converts a saved list to places
func (t *TakeoutImporter) parseSavedList(list *SavedList) []*models.Place {
	var places []*models.Place
	for _, savedPlace := range list.Places {
		place := t.convertSavedPlace(&savedPlace, list.Name)
		if place != nil {
			places = append(places, place)
		}
	}
	return places
}

// convertSavedPlace converts a saved place to our model
func (t *TakeoutImporter) convertSavedPlace(sp *SavedPlace, listName string) *models.Place {
	if sp.Name == "" && sp.Address == "" {
		return nil
	}

	placeID := sp.PlaceID
	if placeID == "" {
		placeID = fmt.Sprintf("saved_%x", generateHash([]byte(sp.Name+sp.Address)))[:16]
	}

	now := time.Now()
	place := &models.Place{
		ID:      generateID(placeID),
		PlaceID: placeID,
		Name:    sp.Name,
		Address: sp.Address,
		Coordinates: models.Coordinates{
			Lat: sp.Coordinates.Latitude,
			Lng: sp.Coordinates.Longitude,
		},
		Categories: sp.Categories,
		UserNotes:  sp.Note,
		UserTags:   []string{listName}, // Use list name as a tag
		CreatedAt:  now,
		UpdatedAt:  now,
		ImportedAt: &now,
		CustomFields: map[string]interface{}{
			"google_maps_url": sp.GoogleURL,
			"imported_from":   "takeout_saved",
			"original_list":   listName,
			"import_date":     now.Format(time.RFC3339),
		},
	}

	if sp.AddedAt != "" {
		place.CustomFields["added_at"] = sp.AddedAt
	}

	// Generate source hash for duplicate detection
	sourceData := fmt.Sprintf("%s|%s|%f,%f|%s",
		sp.Name,
		sp.Address,
		sp.Coordinates.Latitude,
		sp.Coordinates.Longitude,
		placeID)
	place.SourceHash = fmt.Sprintf("%x", generateHash([]byte(sourceData)))

	return place
}
