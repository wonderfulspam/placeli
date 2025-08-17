package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

// testUpdate simulates key input without running the full program
func testUpdate(t *testing.T, model BrowseModel, keySequence []string) BrowseModel {
	t.Helper()

	for _, key := range keySequence {
		var keyMsg tea.KeyMsg
		switch key {
		case "j":
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
		case "k":
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
		case " ":
			keyMsg = tea.KeyMsg{Type: tea.KeySpace}
		case "/":
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
		default:
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}

		updatedModel, _ := model.Update(keyMsg)
		model = updatedModel.(BrowseModel)
	}

	return model
}

func TestBrowseModelNavigation(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add test places
	places := []*models.Place{
		{ID: "1", PlaceID: "place-1", Name: "First Place", Address: "1 Test St"},
		{ID: "2", PlaceID: "place-2", Name: "Second Place", Address: "2 Test St"},
		{ID: "3", PlaceID: "place-3", Name: "Third Place", Address: "3 Test St"},
	}

	for _, place := range places {
		if err := db.SavePlace(place); err != nil {
			t.Fatalf("Failed to save place: %v", err)
		}
	}

	model := NewBrowseModel(db)

	// Load places first
	cmd := model.loadPlaces()
	msg := cmd()

	if placesMsg, ok := msg.(placesLoadedMsg); ok {
		model.places = placesMsg.places
	} else {
		t.Fatalf("Failed to load places: %v", msg)
	}

	// Test navigation down
	t.Run("navigate down", func(t *testing.T) {
		finalModel := testUpdate(t, model, []string{"j", "j"})

		if finalModel.cursor != 2 {
			t.Errorf("Expected cursor at position 2, got %d", finalModel.cursor)
		}

		output := finalModel.View()
		if !strings.Contains(output, "Third Place") {
			t.Errorf("Expected output to contain 'Third Place', got: %s", output)
		}
	})

	// Test navigation up
	t.Run("navigate up from bottom", func(t *testing.T) {
		testModel := model
		testModel.cursor = 2 // Start at bottom

		finalModel := testUpdate(t, testModel, []string{"k"})

		if finalModel.cursor != 1 {
			t.Errorf("Expected cursor at position 1, got %d", finalModel.cursor)
		}

		output := finalModel.View()
		if !strings.Contains(output, "Second Place") {
			t.Errorf("Expected output to contain 'Second Place', got: %s", output)
		}
	})

	// Test selection
	t.Run("select items", func(t *testing.T) {
		finalModel := testUpdate(t, model, []string{" ", "j", " "})

		if len(finalModel.selected) != 2 {
			t.Errorf("Expected 2 selected items, got %d", len(finalModel.selected))
		}

		if _, ok := finalModel.selected[0]; !ok {
			t.Errorf("Expected item 0 to be selected")
		}

		if _, ok := finalModel.selected[1]; !ok {
			t.Errorf("Expected item 1 to be selected")
		}

		// Check for checkmarks in output
		output := finalModel.View()
		checkCount := strings.Count(output, "✓")
		if checkCount < 2 {
			t.Errorf("Expected at least 2 checkmarks in output, got %d. Output: %s", checkCount, output)
		}
	})
}

func TestBrowseModelSearch(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add test places
	places := []*models.Place{
		{ID: "1", PlaceID: "place-1", Name: "Coffee Shop", Address: "1 Test St"},
		{ID: "2", PlaceID: "place-2", Name: "Restaurant", Address: "2 Test St"},
		{ID: "3", PlaceID: "place-3", Name: "Coffee House", Address: "3 Test St"},
	}

	for _, place := range places {
		if err := db.SavePlace(place); err != nil {
			t.Fatalf("Failed to save place: %v", err)
		}
	}

	model := NewBrowseModel(db)

	// Test search mode activation
	t.Run("enter search mode", func(t *testing.T) {
		finalModel := testUpdate(t, model, []string{"/"})

		if !finalModel.searchMode {
			t.Errorf("Expected search mode to be active")
		}

		output := finalModel.View()
		if !strings.Contains(output, "Search:") {
			t.Errorf("Expected output to contain search prompt, got: %s", output)
		}
	})
}

// TestBrowseWithoutTTY tests that the browse functionality works without a TTY
func TestBrowseWithoutTTY(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add a test place
	place := &models.Place{
		ID:         "test-1",
		PlaceID:    "place-1",
		Name:       "Test Place",
		Address:    "123 Test St",
		Rating:     4.5,
		Categories: []string{"restaurant", "food"},
	}

	err = db.SavePlace(place)
	if err != nil {
		t.Fatalf("Failed to save place: %v", err)
	}

	model := NewBrowseModel(db)

	// Load places first
	cmd := model.loadPlaces()
	msg := cmd()

	if placesMsg, ok := msg.(placesLoadedMsg); ok {
		model.places = placesMsg.places
	} else {
		t.Fatalf("Failed to load places: %v", msg)
	}

	// Test that we can create and render the view without TTY
	output := model.View()

	if !strings.Contains(output, "placeli browse") {
		t.Errorf("Expected output to contain title, got: %s", output)
	}

	if !strings.Contains(output, "Test Place") {
		t.Errorf("Expected output to contain place name, got: %s", output)
	}

	if !strings.Contains(output, "123 Test St") {
		t.Errorf("Expected output to contain address, got: %s", output)
	}
}

// TestViewRenderingWithNavigation tests specific rendering behavior during navigation
func TestViewRenderingWithNavigation(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add test places with custom fields to reproduce the issue
	places := []*models.Place{
		{
			ID:         "1",
			PlaceID:    "place-1",
			Name:       "First Restaurant",
			Address:    "1 First St",
			Categories: []string{"restaurant"},
			CustomFields: map[string]interface{}{
				"price_range": "$$$",
				"has_parking": true,
			},
		},
		{
			ID:         "2",
			PlaceID:    "place-2",
			Name:       "Second Cafe",
			Address:    "2 Second Ave",
			Categories: []string{"cafe"},
			CustomFields: map[string]interface{}{
				"wifi_available":  true,
				"outdoor_seating": false,
			},
		},
		{
			ID:         "3",
			PlaceID:    "place-3",
			Name:       "Third Bakery",
			Address:    "3 Third Blvd",
			Categories: []string{"bakery"},
			CustomFields: map[string]interface{}{
				"fresh_bread": true,
				"gluten_free": true,
			},
		},
	}

	for _, place := range places {
		if err := db.SavePlace(place); err != nil {
			t.Fatalf("Failed to save place: %v", err)
		}
	}

	model := NewBrowseModel(db)

	// Load places
	cmd := model.loadPlaces()
	msg := cmd()
	if placesMsg, ok := msg.(placesLoadedMsg); ok {
		model.places = placesMsg.places
	} else {
		t.Fatalf("Failed to load places: %v", msg)
	}

	// Test cursor at position 0
	t.Run("cursor at position 0", func(t *testing.T) {
		model.cursor = 0
		output := model.View()

		t.Logf("Output at cursor 0:\n%s", output)

		// Should show First Restaurant details
		if !strings.Contains(output, "First Restaurant") {
			t.Errorf("Expected 'First Restaurant' in output")
		}
		if !strings.Contains(output, "1 First St") {
			t.Errorf("Expected '1 First St' in output")
		}
		if !strings.Contains(output, "price_range:$$$") {
			t.Errorf("Expected 'price_range:$$$' in output")
		}

		// Should NOT show other places' details prominently
		if strings.Count(output, "Second Cafe") > 1 || strings.Count(output, "Third Bakery") > 1 {
			t.Errorf("Other place names appear too frequently, suggesting rendering issue")
		}
	})

	// Test cursor at position 1
	t.Run("cursor at position 1", func(t *testing.T) {
		model.cursor = 1
		output := model.View()

		t.Logf("Output at cursor 1:\n%s", output)

		// Should show Second Cafe details
		if !strings.Contains(output, "Second Cafe") {
			t.Errorf("Expected 'Second Cafe' in output")
		}
		if !strings.Contains(output, "2 Second Ave") {
			t.Errorf("Expected '2 Second Ave' in output")
		}
		if !strings.Contains(output, "wifi_available:true") {
			t.Errorf("Expected 'wifi_available:true' in output")
		}
	})

	// Test navigation sequence to see if output changes appropriately
	t.Run("navigation sequence", func(t *testing.T) {
		// Start at position 0
		model.cursor = 0
		output1 := model.View()

		// Move to position 1
		finalModel := testUpdate(t, model, []string{"j"})
		output2 := finalModel.View()

		t.Logf("Output before navigation:\n%s", output1)
		t.Logf("Output after navigation:\n%s", output2)

		// Outputs should be different - this is the key test
		if output1 == output2 {
			t.Errorf("Output should change when cursor moves, but it didn't")
		}

		// First output should show first place prominently
		if !strings.Contains(output1, "First Restaurant") {
			t.Errorf("First output should contain 'First Restaurant'")
		}

		// Second output should show second place prominently
		if !strings.Contains(output2, "Second Cafe") {
			t.Errorf("Second output should contain 'Second Cafe'")
		}
	})
}

// TestKeyHandling tests actual key message handling
func TestKeyHandling(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add test places
	places := []*models.Place{
		{ID: "1", PlaceID: "place-1", Name: "First Place"},
		{ID: "2", PlaceID: "place-2", Name: "Second Place"},
		{ID: "3", PlaceID: "place-3", Name: "Third Place"},
	}

	for _, place := range places {
		if err := db.SavePlace(place); err != nil {
			t.Fatalf("Failed to save place: %v", err)
		}
	}

	model := NewBrowseModel(db)

	// Load places
	cmd := model.loadPlaces()
	msg := cmd()
	if placesMsg, ok := msg.(placesLoadedMsg); ok {
		model.places = placesMsg.places
	} else {
		t.Fatalf("Failed to load places: %v", msg)
	}

	// Test different key types to understand what's happening
	testCases := []struct {
		name         string
		keyMsg       tea.KeyMsg
		expectCursor int
	}{
		{
			name: "j key as runes",
			keyMsg: tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("j"),
			},
			expectCursor: 1,
		},
		{
			name: "down arrow",
			keyMsg: tea.KeyMsg{
				Type: tea.KeyDown,
			},
			expectCursor: 1,
		},
		{
			name: "k key as runes",
			keyMsg: tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune("k"),
			},
			expectCursor: 0, // Should move back up
		},
		{
			name: "up arrow",
			keyMsg: tea.KeyMsg{
				Type: tea.KeyUp,
			},
			expectCursor: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset cursor
			model.cursor = 0

			t.Logf("Testing key: %s, String(): %s, Type: %v", tc.name, tc.keyMsg.String(), tc.keyMsg.Type)

			updatedModel, _ := model.Update(tc.keyMsg)
			finalModel := updatedModel.(BrowseModel)

			if finalModel.cursor != tc.expectCursor {
				t.Errorf("Expected cursor %d, got %d for key %s (String: %s)",
					tc.expectCursor, finalModel.cursor, tc.name, tc.keyMsg.String())
			}
		})
	}
}
