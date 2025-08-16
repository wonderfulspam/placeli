package importer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/placeli/internal/models"
)

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

type TakeoutProperties struct {
	Name            string                 `json:"name"`
	Address         string                 `json:"address"`
	GoogleMapsURL   string                 `json:"Google Maps URL"`
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

func ImportFromTakeout(takeoutPath string) ([]*models.Place, error) {
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
		places = append(places, place)
	}

	return places, nil
}

func convertTakeoutPlace(tp TakeoutPlace) *models.Place {
	var coordinates models.Coordinates
	if len(tp.Geometry.Coordinates) >= 2 {
		coordinates = models.Coordinates{
			Lng: tp.Geometry.Coordinates[0],
			Lat: tp.Geometry.Coordinates[1],
		}
	}

	var placeID string
	if tp.Properties.PlaceID != "" {
		placeID = tp.Properties.PlaceID
	} else {
		hash := sha256.Sum256([]byte(tp.Properties.Name + tp.Properties.Address))
		placeID = fmt.Sprintf("takeout_%x", hash)[:16]
	}

	var hours string
	if tp.Properties.Hours != nil {
		hoursData, _ := json.Marshal(tp.Properties.Hours)
		hours = string(hoursData)
	}

	// Generate source hash for duplicate detection
	sourceData := fmt.Sprintf("%s|%s|%f,%f|%s",
		tp.Properties.Name,
		tp.Properties.Address,
		coordinates.Lat,
		coordinates.Lng,
		placeID)
	sourceHash := fmt.Sprintf("%x", sha256.Sum256([]byte(sourceData)))

	now := time.Now()
	place := &models.Place{
		ID:          generateID(placeID),
		PlaceID:     placeID,
		Name:        tp.Properties.Name,
		Address:     tp.Properties.Address,
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

	return place
}

func generateID(placeID string) string {
	hash := sha256.Sum256([]byte(placeID))
	return fmt.Sprintf("%x", hash)[:12]
}
