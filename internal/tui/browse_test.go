package tui

import (
	"testing"

	"github.com/user/placeli/internal/database"
	"github.com/user/placeli/internal/models"
)

func TestNewBrowseModel(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	model := NewBrowseModel(db)

	if model.db != db {
		t.Errorf("Expected db to be set")
	}

	if model.selected == nil {
		t.Errorf("Expected selected map to be initialized")
	}

	if model.cursor != 0 {
		t.Errorf("Expected cursor to start at 0, got %d", model.cursor)
	}
}

func TestBrowseModelLoadPlaces(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add a test place
	place := &models.Place{
		ID:      "test-1",
		PlaceID: "place-1",
		Name:    "Test Place",
		Address: "123 Test St",
	}

	err = db.SavePlace(place)
	if err != nil {
		t.Fatalf("Failed to save place: %v", err)
	}

	model := NewBrowseModel(db)

	// Test loadPlaces command
	cmd := model.loadPlaces()
	if cmd == nil {
		t.Errorf("Expected loadPlaces to return a command")
	}

	// Execute the command to get the message
	msg := cmd()

	switch msg := msg.(type) {
	case placesLoadedMsg:
		if len(msg.places) != 1 {
			t.Errorf("Expected 1 place, got %d", len(msg.places))
		}
		if msg.places[0].Name != "Test Place" {
			t.Errorf("Expected place name 'Test Place', got %s", msg.places[0].Name)
		}
	case errMsg:
		t.Errorf("Expected places loaded message, got error: %v", msg.err)
	default:
		t.Errorf("Expected placesLoadedMsg, got %T", msg)
	}
}
