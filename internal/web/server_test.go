package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/placeli/internal/database"
	"github.com/user/placeli/internal/models"
)

func setupTestServer(t *testing.T) (*Server, *database.DB) {
	db, err := database.New(":memory:")
	require.NoError(t, err)

	server, err := NewServer(db, 8080, "test-api-key")
	require.NoError(t, err)

	place := &models.Place{
		ID:      "test-place-1",
		PlaceID: "test-google-place-id",
		Name:    "Test Place",
		Address: "123 Test St",
		Coordinates: models.Coordinates{
			Lat: 37.7749,
			Lng: -122.4194,
		},
		Rating:     4.5,
		Categories: []string{"Restaurant"},
		UserNotes:  "Great food",
		UserTags:   []string{"favorite"},
	}
	err = db.SavePlace(place)
	require.NoError(t, err)

	return server, db
}

func TestHandleIndex(t *testing.T) {
	server, db := setupTestServer(t)
	defer db.Close()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Placeli")
	assert.Contains(t, w.Body.String(), "test-api-key")
}

func TestHandleAPIPlaces(t *testing.T) {
	server, db := setupTestServer(t)
	defer db.Close()

	tests := []struct {
		name     string
		url      string
		wantCode int
		wantLen  int
	}{
		{
			name:     "Get all places",
			url:      "/api/places",
			wantCode: http.StatusOK,
			wantLen:  1,
		},
		{
			name:     "Search places",
			url:      "/api/places?search=Test",
			wantCode: http.StatusOK,
			wantLen:  1,
		},
		{
			name:     "Search no results",
			url:      "/api/places?search=NotFound",
			wantCode: http.StatusOK,
			wantLen:  0,
		},
		{
			name:     "With limit",
			url:      "/api/places?limit=10",
			wantCode: http.StatusOK,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			server.handleAPIPlaces(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				var places []*models.Place
				err := json.Unmarshal(w.Body.Bytes(), &places)
				require.NoError(t, err)
				assert.Len(t, places, tt.wantLen)
			}
		})
	}
}

func TestHandleAPIPlacesInvalidMethod(t *testing.T) {
	server, db := setupTestServer(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/places", nil)
	w := httptest.NewRecorder()

	server.handleAPIPlaces(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleAPIPlace(t *testing.T) {
	server, db := setupTestServer(t)
	defer db.Close()

	places, err := db.ListPlaces(10, 0)
	require.NoError(t, err)
	require.Len(t, places, 1)
	placeID := places[0].ID

	t.Run("Get place", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/place/"+placeID, nil)
		w := httptest.NewRecorder()

		server.handleAPIPlace(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var place models.Place
		err := json.Unmarshal(w.Body.Bytes(), &place)
		require.NoError(t, err)
		assert.Equal(t, "Test Place", place.Name)
	})

	t.Run("Update place", func(t *testing.T) {
		updateData := `{"name":"Updated Place","address":"456 New St"}`
		req := httptest.NewRequest("PUT", "/api/place/"+placeID, strings.NewReader(updateData))
		w := httptest.NewRecorder()

		server.handleAPIPlace(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var place models.Place
		err := json.Unmarshal(w.Body.Bytes(), &place)
		require.NoError(t, err)
		assert.Equal(t, "Updated Place", place.Name)
		assert.Equal(t, "456 New St", place.Address)
	})

	t.Run("Invalid place ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/place/", nil)
		w := httptest.NewRecorder()

		server.handleAPIPlace(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Place not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/place/99999", nil)
		w := httptest.NewRecorder()

		server.handleAPIPlace(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Invalid method", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/place/1", nil)
		w := httptest.NewRecorder()

		server.handleAPIPlace(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestNewServer(t *testing.T) {
	db, err := database.New(":memory:")
	require.NoError(t, err)
	defer db.Close()

	server, err := NewServer(db, 8080, "api-key")
	require.NoError(t, err)

	assert.NotNil(t, server)
	assert.Equal(t, 8080, server.port)
	assert.Equal(t, "api-key", server.apiKey)
	assert.NotNil(t, server.db)
	assert.NotNil(t, server.tmpl)
}
