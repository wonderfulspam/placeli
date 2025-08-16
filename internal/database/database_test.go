package database

import (
	"os"
	"testing"
	"time"

	"github.com/user/placeli/internal/models"
)

func TestDB_SaveAndGetPlace(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	place := &models.Place{
		ID:      "test-id",
		PlaceID: "place-123",
		Name:    "Test Restaurant",
		Address: "123 Test St, Test City",
		Coordinates: models.Coordinates{
			Lat: 40.7128,
			Lng: -74.0060,
		},
		Categories:   []string{"restaurant", "food"},
		Rating:       4.5,
		UserRatings:  100,
		PriceLevel:   2,
		Hours:        "Mon-Sun 9AM-10PM",
		Phone:        "(555) 123-4567",
		Website:      "https://testrestaurant.com",
		UserNotes:    "Great food!",
		UserTags:     []string{"favorite", "visited"},
		CustomFields: map[string]interface{}{"priority": "high"},
	}

	err = db.SavePlace(place)
	if err != nil {
		t.Fatalf("SavePlace failed: %v", err)
	}

	retrieved, err := db.GetPlace("test-id")
	if err != nil {
		t.Fatalf("GetPlace failed: %v", err)
	}

	if retrieved.Name != place.Name {
		t.Errorf("Expected name %q, got %q", place.Name, retrieved.Name)
	}
	if retrieved.Rating != place.Rating {
		t.Errorf("Expected rating %f, got %f", place.Rating, retrieved.Rating)
	}
	if len(retrieved.UserTags) != len(place.UserTags) {
		t.Errorf("Expected %d tags, got %d", len(place.UserTags), len(retrieved.UserTags))
	}
}

func TestDB_ListPlaces(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	place1 := &models.Place{
		ID:   "place1",
		Name: "Place One",
	}
	place2 := &models.Place{
		ID:   "place2",
		Name: "Place Two",
	}

	if err := db.SavePlace(place1); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	if err := db.SavePlace(place2); err != nil {
		t.Fatal(err)
	}

	places, err := db.ListPlaces(10, 0)
	if err != nil {
		t.Fatalf("ListPlaces failed: %v", err)
	}

	if len(places) != 2 {
		t.Errorf("Expected 2 places, got %d", len(places))
	}

	if places[0].Name != "Place Two" {
		t.Errorf("Expected first place to be 'Place Two' (most recent), got %q", places[0].Name)
	}
}

func TestDB_SearchPlaces(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	place1 := &models.Place{
		ID:      "place1",
		Name:    "Pizza Palace",
		Address: "123 Main St",
	}
	place2 := &models.Place{
		ID:        "place2",
		Name:      "Burger Joint",
		Address:   "456 Side St",
		UserNotes: "Best pizza in town",
	}

	if err := db.SavePlace(place1); err != nil {
		t.Fatal(err)
	}
	if err := db.SavePlace(place2); err != nil {
		t.Fatal(err)
	}

	results, err := db.SearchPlaces("pizza")
	if err != nil {
		t.Fatalf("SearchPlaces failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'pizza' search, got %d", len(results))
	}
}

func TestDB_DeletePlace(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := New(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	place := &models.Place{
		ID:   "test-delete",
		Name: "To Be Deleted",
	}

	if err := db.SavePlace(place); err != nil {
		t.Fatal(err)
	}

	_, err = db.GetPlace("test-delete")
	if err != nil {
		t.Fatal("Place should exist before deletion")
	}

	err = db.DeletePlace("test-delete")
	if err != nil {
		t.Fatalf("DeletePlace failed: %v", err)
	}

	_, err = db.GetPlace("test-delete")
	if err == nil {
		t.Error("Place should not exist after deletion")
	}
}
