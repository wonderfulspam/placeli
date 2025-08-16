package sources

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/user/placeli/internal/models"
)

// FoursquareImporter handles Foursquare/Swarm check-in exports
type FoursquareImporter struct{}

// FoursquareExport represents the structure of a Foursquare export
type FoursquareExport struct {
	Checkins []FoursquareCheckin `json:"checkins"`
}

// FoursquareCheckin represents a single check-in from Foursquare/Swarm
type FoursquareCheckin struct {
	ID           string             `json:"id"`
	CreatedAt    int64              `json:"createdAt"`
	Type         string             `json:"type"`
	TimeZone     string             `json:"timeZone"`
	Venue        FoursquareVenue    `json:"venue"`
	Photos       []FoursquarePhoto  `json:"photos"`
	Comments     string             `json:"comments"`
	Source       FoursquareSource   `json:"source"`
}

// FoursquareVenue represents venue information
type FoursquareVenue struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	Contact    FoursquareContact        `json:"contact"`
	Location   FoursquareLocation       `json:"location"`
	Categories []FoursquareCategory     `json:"categories"`
	URL        string                   `json:"url"`
	Stats      FoursquareStats          `json:"stats"`
	Rating     float64                  `json:"rating"`
	Price      FoursquarePrice          `json:"price"`
}

// FoursquareContact represents contact information
type FoursquareContact struct {
	Phone       string `json:"phone"`
	FormattedPhone string `json:"formattedPhone"`
	Twitter     string `json:"twitter"`
	Instagram   string `json:"instagram"`
	Facebook    string `json:"facebook"`
}

// FoursquareLocation represents location information
type FoursquareLocation struct {
	Address     string  `json:"address"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Distance    int     `json:"distance"`
	PostalCode  string  `json:"postalCode"`
	CC          string  `json:"cc"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Country     string  `json:"country"`
	FormattedAddress []string `json:"formattedAddress"`
}

// FoursquareCategory represents venue category
type FoursquareCategory struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Primary bool  `json:"primary"`
	Icon   FoursquareIcon `json:"icon"`
}

// FoursquareIcon represents category icon
type FoursquareIcon struct {
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}

// FoursquareStats represents venue statistics
type FoursquareStats struct {
	CheckinsCount int `json:"checkinsCount"`
	UsersCount    int `json:"usersCount"`
	TipCount      int `json:"tipCount"`
}

// FoursquarePrice represents price information
type FoursquarePrice struct {
	Tier     int    `json:"tier"`
	Message  string `json:"message"`
	Currency string `json:"currency"`
}

// FoursquarePhoto represents check-in photos
type FoursquarePhoto struct {
	ID     string `json:"id"`
	Source FoursquareSource `json:"source"`
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// FoursquareSource represents the source of data
type FoursquareSource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (fi *FoursquareImporter) Name() string {
	return "Foursquare/Swarm"
}

func (fi *FoursquareImporter) SupportedFormats() []string {
	return []string{"json"}
}

func (fi *FoursquareImporter) ImportFromFile(filePath string) ([]*models.Place, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return fi.ImportFromData(data, "json")
}

func (fi *FoursquareImporter) ImportFromData(data []byte, format string) ([]*models.Place, error) {
	if format != "json" {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return fi.parseJSON(data)
}

func (fi *FoursquareImporter) parseJSON(data []byte) ([]*models.Place, error) {
	var export FoursquareExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("failed to parse Foursquare JSON: %w", err)
	}

	// Group check-ins by venue to avoid duplicates
	venueMap := make(map[string]*models.Place)
	venueCheckins := make(map[string][]FoursquareCheckin)

	for _, checkin := range export.Checkins {
		venueID := checkin.Venue.ID
		if venueID == "" {
			continue // Skip check-ins without venue ID
		}

		// Track check-ins per venue for metadata
		venueCheckins[venueID] = append(venueCheckins[venueID], checkin)

		// Create or update place for this venue
		if _, exists := venueMap[venueID]; !exists {
			place := fi.convertFoursquareVenue(checkin.Venue, checkin)
			if place != nil {
				venueMap[venueID] = place
			}
		}
	}

	// Add check-in statistics to places
	for venueID, place := range venueMap {
		checkins := venueCheckins[venueID]
		fi.addCheckinMetadata(place, checkins)
	}

	// Convert map to slice
	var places []*models.Place
	for _, place := range venueMap {
		places = append(places, place)
	}

	return places, nil
}

func (fi *FoursquareImporter) convertFoursquareVenue(venue FoursquareVenue, checkin FoursquareCheckin) *models.Place {
	if venue.Name == "" {
		return nil // Skip venues without names
	}

	coords := models.Coordinates{
		Lat: venue.Location.Lat,
		Lng: venue.Location.Lng,
	}

	// Generate unique IDs
	sourceData := fmt.Sprintf("foursquare|%s|%s|%f,%f", 
		venue.ID,
		venue.Name,
		coords.Lat, 
		coords.Lng)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(sourceData)))
	placeID := fmt.Sprintf("4sq_%s", venue.ID)

	now := time.Now()
	place := &models.Place{
		ID:          fi.generateID(placeID),
		PlaceID:     placeID,
		Name:        venue.Name,
		Address:     fi.buildAddress(venue.Location),
		Coordinates: coords,
		Categories:  fi.extractCategories(venue.Categories),
		Rating:      float32(venue.Rating),
		UserRatings: venue.Stats.CheckinsCount,
		PriceLevel:  venue.Price.Tier,
		Phone:       venue.Contact.FormattedPhone,
		Website:     venue.URL,
		UserNotes:   checkin.Comments,
		UserTags:    []string{},
		Photos:      fi.convertPhotos(checkin.Photos),
		Reviews:     []models.Review{},
		CustomFields: map[string]interface{}{
			"imported_from":     "foursquare",
			"import_date":       now.Format(time.RFC3339),
			"foursquare_id":     venue.ID,
			"checkins_count":    venue.Stats.CheckinsCount,
			"users_count":       venue.Stats.UsersCount,
			"tips_count":        venue.Stats.TipCount,
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		ImportedAt: &now,
		SourceHash: sourceHash,
	}

	// Add social media links if available
	if venue.Contact.Twitter != "" {
		place.CustomFields["twitter"] = venue.Contact.Twitter
	}
	if venue.Contact.Instagram != "" {
		place.CustomFields["instagram"] = venue.Contact.Instagram
	}
	if venue.Contact.Facebook != "" {
		place.CustomFields["facebook"] = venue.Contact.Facebook
	}

	// Add price information
	if venue.Price.Message != "" {
		place.CustomFields["price_message"] = venue.Price.Message
	}
	if venue.Price.Currency != "" {
		place.CustomFields["currency"] = venue.Price.Currency
	}

	return place
}

func (fi *FoursquareImporter) buildAddress(location FoursquareLocation) string {
	if len(location.FormattedAddress) > 0 {
		return strings.Join(location.FormattedAddress, ", ")
	}
	
	var parts []string
	if location.Address != "" {
		parts = append(parts, location.Address)
	}
	if location.City != "" {
		parts = append(parts, location.City)
	}
	if location.State != "" {
		parts = append(parts, location.State)
	}
	if location.PostalCode != "" {
		parts = append(parts, location.PostalCode)
	}
	if location.Country != "" {
		parts = append(parts, location.Country)
	}
	
	return strings.Join(parts, ", ")
}

func (fi *FoursquareImporter) extractCategories(categories []FoursquareCategory) []string {
	var result []string
	for _, category := range categories {
		if category.Name != "" {
			result = append(result, category.Name)
		}
	}
	return result
}

func (fi *FoursquareImporter) convertPhotos(photos []FoursquarePhoto) []models.Photo {
	var result []models.Photo
	for _, photo := range photos {
		result = append(result, models.Photo{
			Reference: photo.ID,
			LocalPath: "", // Photos would need to be downloaded separately
			Width:     photo.Width,
			Height:    photo.Height,
		})
	}
	return result
}

func (fi *FoursquareImporter) addCheckinMetadata(place *models.Place, checkins []FoursquareCheckin) {
	if len(checkins) == 0 {
		return
	}

	// Find first and last check-in
	var firstCheckin, lastCheckin int64 = checkins[0].CreatedAt, checkins[0].CreatedAt
	for _, checkin := range checkins {
		if checkin.CreatedAt < firstCheckin {
			firstCheckin = checkin.CreatedAt
		}
		if checkin.CreatedAt > lastCheckin {
			lastCheckin = checkin.CreatedAt
		}
	}

	place.CustomFields["my_checkins_count"] = len(checkins)
	place.CustomFields["first_visit"] = time.Unix(firstCheckin, 0).Format("2006-01-02")
	place.CustomFields["last_visit"] = time.Unix(lastCheckin, 0).Format("2006-01-02")

	// Collect unique comments
	var comments []string
	for _, checkin := range checkins {
		if checkin.Comments != "" {
			comments = append(comments, checkin.Comments)
		}
	}
	if len(comments) > 0 {
		place.CustomFields["checkin_comments"] = comments
	}
}

func (fi *FoursquareImporter) generateID(placeID string) string {
	hash := sha256.Sum256([]byte(placeID))
	return fmt.Sprintf("%x", hash)[:12]
}