package models

import (
	"testing"
	"time"
)

func TestPlace_ToJSON(t *testing.T) {
	place := &Place{
		ID:      "test-id",
		PlaceID: "place-123",
		Name:    "Test Place",
		Address: "123 Test St",
		Coordinates: Coordinates{
			Lat: 40.7128,
			Lng: -74.0060,
		},
		Categories: []string{"restaurant", "food"},
		UserNotes:  "Great place!",
		UserTags:   []string{"favorite", "visited"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := place.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var decoded Place
	err = decoded.FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if decoded.Name != place.Name {
		t.Errorf("Expected name %q, got %q", place.Name, decoded.Name)
	}
	if decoded.Coordinates.Lat != place.Coordinates.Lat {
		t.Errorf("Expected lat %f, got %f", place.Coordinates.Lat, decoded.Coordinates.Lat)
	}
}

func TestPlace_HasTag(t *testing.T) {
	place := &Place{
		UserTags: []string{"favorite", "visited", "restaurant"},
	}

	if !place.HasTag("favorite") {
		t.Error("Expected place to have 'favorite' tag")
	}
	if place.HasTag("nonexistent") {
		t.Error("Expected place to not have 'nonexistent' tag")
	}
}

func TestPlace_AddTag(t *testing.T) {
	place := &Place{
		UserTags: []string{"favorite"},
	}

	place.AddTag("visited")
	if !place.HasTag("visited") {
		t.Error("Expected place to have 'visited' tag after adding")
	}

	place.AddTag("favorite")
	count := 0
	for _, tag := range place.UserTags {
		if tag == "favorite" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected 'favorite' tag to appear once, got %d times", count)
	}
}

func TestPlace_RemoveTag(t *testing.T) {
	place := &Place{
		UserTags: []string{"favorite", "visited", "restaurant"},
	}

	place.RemoveTag("visited")
	if place.HasTag("visited") {
		t.Error("Expected 'visited' tag to be removed")
	}
	if !place.HasTag("favorite") {
		t.Error("Expected 'favorite' tag to remain")
	}
}
