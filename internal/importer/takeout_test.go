package importer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseListFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "takeout-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testData := TakeoutList{
		Name: "My Favorites",
		Features: []TakeoutPlace{
			{
				Geometry: TakeoutGeometry{
					Coordinates: []float64{-74.0060, 40.7128},
					Type:        "Point",
				},
				Properties: TakeoutProperties{
					Name:          "Test Restaurant",
					Address:       "123 Test St, New York, NY",
					GoogleMapsURL: "https://maps.google.com/?cid=123",
					Categories:    []string{"restaurant", "food"},
					Rating:        4.5,
					ReviewCount:   100,
					PriceLevel:    2,
					Phone:         "(555) 123-4567",
					Website:       "https://testrestaurant.com",
					Description:   "Great food",
					PlaceID:       "ChIJtest123",
				},
			},
		},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tmpDir, "test_list.json")
	err = os.WriteFile(testFile, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	places, err := parseListFile(testFile)
	if err != nil {
		t.Fatalf("parseListFile failed: %v", err)
	}

	if len(places) != 1 {
		t.Fatalf("Expected 1 place, got %d", len(places))
	}

	place := places[0]
	if place.Name != "Test Restaurant" {
		t.Errorf("Expected name 'Test Restaurant', got %q", place.Name)
	}
	if place.Rating != 4.5 {
		t.Errorf("Expected rating 4.5, got %f", place.Rating)
	}
	if place.Coordinates.Lat != 40.7128 {
		t.Errorf("Expected lat 40.7128, got %f", place.Coordinates.Lat)
	}
	if place.UserNotes != "Great food" {
		t.Errorf("Expected notes 'Great food', got %q", place.UserNotes)
	}
}

func TestConvertTakeoutPlace(t *testing.T) {
	tp := TakeoutPlace{
		Geometry: TakeoutGeometry{
			Coordinates: []float64{-74.0060, 40.7128},
			Type:        "Point",
		},
		Properties: TakeoutProperties{
			Name:        "Test Place",
			Address:     "123 Main St",
			Categories:  []string{"restaurant"},
			Rating:      4.0,
			ReviewCount: 50,
			PlaceID:     "ChIJtest456",
		},
	}

	place := convertTakeoutPlace(tp)

	if place.Name != "Test Place" {
		t.Errorf("Expected name 'Test Place', got %q", place.Name)
	}
	if place.PlaceID != "ChIJtest456" {
		t.Errorf("Expected PlaceID 'ChIJtest456', got %q", place.PlaceID)
	}
	if place.Coordinates.Lng != -74.0060 {
		t.Errorf("Expected lng -74.0060, got %f", place.Coordinates.Lng)
	}
	if place.Rating != 4.0 {
		t.Errorf("Expected rating 4.0, got %f", place.Rating)
	}
}
