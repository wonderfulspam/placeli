package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/placeli/internal/models"
)

func createTestPlaces() []*models.Place {
	now := time.Now()
	return []*models.Place{
		{
			ID:      "place1",
			PlaceID: "ChIJN1t_tDeuEmsRUsoyG83frY4",
			Name:    "Joe's Pizza",
			Address: "123 Main St, Brooklyn, NY 11201",
			Coordinates: models.Coordinates{
				Lat: 40.6892,
				Lng: -74.0445,
			},
			Categories:  []string{"Restaurant", "Pizza"},
			Rating:      4.2,
			UserRatings: 523,
			PriceLevel:  2,
			Hours:       "Mon-Thu 11AM-10PM, Fri-Sat 11AM-11PM, Sun 12PM-10PM",
			Phone:       "(718) 555-0123",
			Website:     "joespizzabrooklyn.com",
			UserNotes:   "Great pizza, a bit crowded on weekends",
			UserTags:    []string{"favorite", "pizza"},
			Photos: []models.Photo{
				{
					Reference: "photo1",
					LocalPath: "/photos/pizza.jpg",
					Width:     800,
					Height:    600,
				},
			},
			Reviews: []models.Review{
				{
					Author: "Sarah M.",
					Rating: 5,
					Text:   "Absolutely the best pizza in Brooklyn!",
					Time:   now.Add(-2 * 24 * time.Hour),
				},
			},
			CustomFields: map[string]interface{}{
				"visited_date": "2023-12-01",
				"priority":     "high",
			},
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
		},
		{
			ID:      "place2",
			PlaceID: "ChIJOwg_06VPwokRYv534QaPC8g",
			Name:    "Central Park",
			Address: "New York, NY 10024",
			Coordinates: models.Coordinates{
				Lat: 40.7829,
				Lng: -73.9654,
			},
			Categories:  []string{"Park", "Tourist Attraction"},
			Rating:      4.8,
			UserRatings: 12845,
			UserNotes:   "Beautiful park for morning runs",
			UserTags:    []string{"exercise", "nature"},
			CreatedAt:   now.Add(-60 * 24 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
		},
	}
}

func TestExportCSV(t *testing.T) {
	places := createTestPlaces()
	var buf bytes.Buffer

	err := ExportCSV(places, &buf)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	assert.Equal(t, 3, len(lines)) // header + 2 places

	assert.Contains(t, lines[0], "ID,PlaceID,Name,Address")
	assert.Contains(t, lines[1], "Joe's Pizza")
	assert.Contains(t, lines[1], "40.689200")
	assert.Contains(t, lines[1], "Restaurant; Pizza")
	assert.Contains(t, lines[2], "Central Park")
}

func TestExportJSON(t *testing.T) {
	places := createTestPlaces()
	var buf bytes.Buffer

	err := ExportJSON(places, &buf)
	require.NoError(t, err)

	var result []*models.Place
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, 2, len(result))
	assert.Equal(t, "Joe's Pizza", result[0].Name)
	assert.Equal(t, "Central Park", result[1].Name)
	assert.Equal(t, 40.6892, result[0].Coordinates.Lat)
}

func TestExportGeoJSON(t *testing.T) {
	places := createTestPlaces()
	var buf bytes.Buffer

	err := ExportGeoJSON(places, &buf)
	require.NoError(t, err)

	var result GeoJSONFeatureCollection
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "FeatureCollection", result.Type)
	assert.Equal(t, 2, len(result.Features))

	feature := result.Features[0]
	assert.Equal(t, "Feature", feature.Type)
	assert.Equal(t, "Point", feature.Geometry.Type)
	assert.Equal(t, -74.0445, feature.Geometry.Coordinates[0]) // longitude first in GeoJSON
	assert.Equal(t, 40.6892, feature.Geometry.Coordinates[1])  // latitude second
	assert.Equal(t, "Joe's Pizza", feature.Properties["name"])
	assert.Equal(t, []interface{}{"favorite", "pizza"}, feature.Properties["user_tags"])
}

func TestExportMarkdown(t *testing.T) {
	places := createTestPlaces()
	var buf bytes.Buffer

	err := ExportMarkdown(places, &buf)
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "# Places Export")
	assert.Contains(t, output, "Total places: 2")
	assert.Contains(t, output, "## Joe's Pizza")
	assert.Contains(t, output, "## Central Park")
	assert.Contains(t, output, "**Address:** 123 Main St, Brooklyn, NY 11201")
	assert.Contains(t, output, "**Rating:** ⭐⭐⭐⭐½ 4.2 (523 reviews)")
	assert.Contains(t, output, "**Price Level:** $$")
	assert.Contains(t, output, "**Tags:** `favorite`, `pizza`")
	assert.Contains(t, output, "### Notes")
	assert.Contains(t, output, "Great pizza, a bit crowded on weekends")
	assert.Contains(t, output, "### Photos")
	assert.Contains(t, output, "### Reviews")
	assert.Contains(t, output, "**Sarah M.** ⭐⭐⭐⭐⭐")
}

func TestExportWithFormat(t *testing.T) {
	places := createTestPlaces()

	testCases := []struct {
		format   Format
		expected string
	}{
		{FormatCSV, "ID,PlaceID,Name"},
		{FormatJSON, `"name": "Joe's Pizza"`},
		{FormatGeoJSON, `"FeatureCollection"`},
		{FormatMarkdown, "# Places Export"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.format), func(t *testing.T) {
			var buf bytes.Buffer
			err := Export(places, tc.format, &buf)
			require.NoError(t, err)
			assert.Contains(t, buf.String(), tc.expected)
		})
	}
}

func TestValidateFormat(t *testing.T) {
	validFormats := []string{"csv", "json", "geojson", "markdown", "md"}
	for _, format := range validFormats {
		assert.NoError(t, ValidateFormat(format))
		assert.NoError(t, ValidateFormat(strings.ToUpper(format)))
	}

	invalidFormats := []string{"xml", "pdf", "invalid"}
	for _, format := range invalidFormats {
		assert.Error(t, ValidateFormat(format))
	}
}

func TestGetSupportedFormats(t *testing.T) {
	formats := GetSupportedFormats()
	expected := []string{"csv", "geojson", "json", "markdown"}
	assert.Equal(t, expected, formats)
}

func TestExportUnsupportedFormat(t *testing.T) {
	places := createTestPlaces()
	var buf bytes.Buffer

	err := Export(places, "invalid", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported export format")
}

func TestExportEmptyPlaces(t *testing.T) {
	var places []*models.Place
	var buf bytes.Buffer

	err := ExportCSV(places, &buf)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 1, len(lines)) // just header
}

func TestExportMarkdownWithEmptyFields(t *testing.T) {
	place := &models.Place{
		ID:   "minimal",
		Name: "Minimal Place",
		Coordinates: models.Coordinates{
			Lat: 0,
			Lng: 0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var buf bytes.Buffer
	err := ExportMarkdown([]*models.Place{place}, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "## Minimal Place")
	assert.NotContains(t, output, "**Address:**")
	assert.NotContains(t, output, "**Rating:**")
	assert.NotContains(t, output, "### Notes")
}
