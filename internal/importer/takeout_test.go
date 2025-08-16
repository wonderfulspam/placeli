package importer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/placeli/internal/models"
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

func TestImportFromTakeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "takeout-import-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create directory structure similar to Google Takeout
	savedPlacesDir := filepath.Join(tmpDir, "Takeout", "Maps (your places)", "Saved Places")
	err = os.MkdirAll(savedPlacesDir, 0755)
	require.NoError(t, err)

	// Create test data for first list
	testData1 := TakeoutList{
		Name: "Restaurants",
		Features: []TakeoutPlace{
			{
				Geometry: TakeoutGeometry{
					Coordinates: []float64{-74.0060, 40.7128},
					Type:        "Point",
				},
				Properties: TakeoutProperties{
					Name:        "Joe's Pizza",
					Address:     "123 Main St, NY",
					PlaceID:     "ChIJpizza123",
					Rating:      4.5,
					ReviewCount: 200,
					Categories:  []string{"restaurant", "pizza"},
				},
			},
		},
	}

	// Create test data for second list
	testData2 := TakeoutList{
		Name: "Coffee Shops",
		Features: []TakeoutPlace{
			{
				Geometry: TakeoutGeometry{
					Coordinates: []float64{-73.9857, 40.7484},
					Type:        "Point",
				},
				Properties: TakeoutProperties{
					Name:        "Blue Bottle Coffee",
					Address:     "456 Coffee Ave, NY",
					PlaceID:     "ChIJcoffee456",
					Rating:      4.2,
					ReviewCount: 150,
					Categories:  []string{"cafe", "coffee"},
				},
			},
		},
	}

	// Write first list file
	data1, err := json.Marshal(testData1)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(savedPlacesDir, "Restaurants.json"), data1, 0644)
	require.NoError(t, err)

	// Write second list file
	data2, err := json.Marshal(testData2)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(savedPlacesDir, "Coffee Shops.json"), data2, 0644)
	require.NoError(t, err)

	// Write a non-matching file (should be ignored)
	err = os.WriteFile(filepath.Join(tmpDir, "other.json"), []byte("{}"), 0644)
	require.NoError(t, err)

	// Test ImportFromTakeout
	places, err := ImportFromTakeout(tmpDir)
	require.NoError(t, err)

	assert.Len(t, places, 2, "Should import 2 places")

	// Find pizza place
	var pizzaPlace *models.Place
	for _, place := range places {
		if strings.Contains(place.Name, "Pizza") {
			pizzaPlace = place
			break
		}
	}
	require.NotNil(t, pizzaPlace, "Should find pizza place")

	assert.Equal(t, "Joe's Pizza", pizzaPlace.Name)
	assert.Equal(t, "ChIJpizza123", pizzaPlace.PlaceID)
	assert.Equal(t, float32(4.5), pizzaPlace.Rating)
	assert.Equal(t, 200, pizzaPlace.UserRatings)
	assert.Contains(t, pizzaPlace.Categories, "restaurant")
	assert.Contains(t, pizzaPlace.Categories, "pizza")
	assert.Equal(t, float64(40.7128), pizzaPlace.Coordinates.Lat)
	assert.Equal(t, float64(-74.0060), pizzaPlace.Coordinates.Lng)
}

func TestImportFromTakeout_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "takeout-empty-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	places, err := ImportFromTakeout(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, places, "Should return empty slice for empty directory")
}

func TestImportFromTakeout_NonexistentDirectory(t *testing.T) {
	places, err := ImportFromTakeout("/nonexistent/path")
	assert.Error(t, err)
	assert.Nil(t, places)
}

func TestParseListFile_MalformedJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "takeout-malformed-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Write malformed JSON
	malformedFile := filepath.Join(tmpDir, "malformed.json")
	err = os.WriteFile(malformedFile, []byte("{invalid json"), 0644)
	require.NoError(t, err)

	places, err := parseListFile(malformedFile)
	assert.Error(t, err)
	assert.Nil(t, places)
}

func TestParseListFile_NonexistentFile(t *testing.T) {
	places, err := parseListFile("/nonexistent/file.json")
	assert.Error(t, err)
	assert.Nil(t, places)
}

func TestConvertTakeoutPlace_EmptyCoordinates(t *testing.T) {
	tp := TakeoutPlace{
		Geometry: TakeoutGeometry{
			Coordinates: []float64{}, // Empty coordinates
			Type:        "Point",
		},
		Properties: TakeoutProperties{
			Name:    "Test Place",
			Address: "123 Test St",
		},
	}

	place := convertTakeoutPlace(tp)
	assert.Equal(t, "Test Place", place.Name)
	assert.Equal(t, float64(0), place.Coordinates.Lat)
	assert.Equal(t, float64(0), place.Coordinates.Lng)
}

func TestConvertTakeoutPlace_NoPlaceID(t *testing.T) {
	tp := TakeoutPlace{
		Geometry: TakeoutGeometry{
			Coordinates: []float64{-74.0060, 40.7128},
			Type:        "Point",
		},
		Properties: TakeoutProperties{
			Name:    "Test Place",
			Address: "123 Test St",
			// No PlaceID provided
		},
	}

	place := convertTakeoutPlace(tp)
	assert.Equal(t, "Test Place", place.Name)
	assert.True(t, strings.HasPrefix(place.PlaceID, "takeout_"), "Should generate takeout_ prefixed PlaceID")
	assert.Len(t, place.PlaceID, 16, "Generated PlaceID should be 16 characters")
}

func TestConvertTakeoutPlace_WithHours(t *testing.T) {
	hours := map[string]interface{}{
		"monday":    "9:00 AM - 5:00 PM",
		"tuesday":   "9:00 AM - 5:00 PM",
		"wednesday": "9:00 AM - 5:00 PM",
	}

	tp := TakeoutPlace{
		Properties: TakeoutProperties{
			Name:    "Test Place",
			Address: "123 Test St",
			Hours:   hours,
		},
	}

	place := convertTakeoutPlace(tp)
	assert.Equal(t, "Test Place", place.Name)
	assert.NotEmpty(t, place.Hours, "Hours should be serialized to JSON string")
	assert.Contains(t, place.Hours, "monday", "Hours should contain serialized data")
}

func TestConvertTakeoutPlace_CustomFields(t *testing.T) {
	tp := TakeoutPlace{
		Properties: TakeoutProperties{
			Name:          "Test Place",
			GoogleMapsURL: "https://maps.google.com/?cid=123",
			Description:   "Great place to visit",
		},
	}

	place := convertTakeoutPlace(tp)
	assert.Equal(t, "Test Place", place.Name)
	assert.Equal(t, "Great place to visit", place.UserNotes)
	
	assert.Contains(t, place.CustomFields, "google_maps_url")
	assert.Equal(t, "https://maps.google.com/?cid=123", place.CustomFields["google_maps_url"])
	
	assert.Contains(t, place.CustomFields, "imported_from")
	assert.Equal(t, "takeout", place.CustomFields["imported_from"])
	
	assert.Contains(t, place.CustomFields, "import_date")
	assert.IsType(t, "", place.CustomFields["import_date"])
}

func TestGenerateID(t *testing.T) {
	id1 := generateID("test-place-id")
	id2 := generateID("test-place-id")
	id3 := generateID("different-place-id")

	assert.Len(t, id1, 12, "Generated ID should be 12 characters")
	assert.Equal(t, id1, id2, "Same input should generate same ID")
	assert.NotEqual(t, id1, id3, "Different input should generate different ID")
}
