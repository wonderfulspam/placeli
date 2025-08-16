package maps

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")
	assert.NotNil(t, client)
	assert.Equal(t, "test-api-key", client.apiKey)
	assert.Equal(t, "https://maps.googleapis.com/maps/api", client.baseURL)
	assert.NotNil(t, client.httpClient)
}

func TestGetPlaceDetails_Success(t *testing.T) {
	mockResponse := PlaceDetailsResponse{
		Status: "OK",
		Result: PlaceDetails{
			PlaceID:     "ChIJN1t_tDeuEmsRUsoyG83frY4",
			Name:        "Joe's Pizza",
			Rating:      4.2,
			UserRatings: 523,
			PriceLevel:  2,
			Website:     "https://joespizza.com",
			Phone:       "(718) 555-0123",
			Address:     "123 Main St, Brooklyn, NY 11201",
			Types:       []string{"restaurant", "food", "establishment"},
			Geometry: PlaceGeometry{
				Location: PlaceLocation{
					Lat: 40.6892,
					Lng: -74.0445,
				},
			},
			Hours: &OpeningHours{
				OpenNow:     true,
				WeekdayText: []string{"Monday: 11:00 AM – 10:00 PM", "Tuesday: 11:00 AM – 10:00 PM"},
			},
			Photos: []PlacePhoto{
				{
					PhotoReference: "photo_ref_1",
					Width:          800,
					Height:         600,
				},
			},
			Reviews: []PlaceReview{
				{
					AuthorName:      "Sarah M.",
					Rating:          5,
					Text:            "Great pizza!",
					Time:            time.Now().Unix(),
					ProfilePhotoURL: "https://example.com/photo.jpg",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Query().Get("place_id"), "ChIJN1t_tDeuEmsRUsoyG83frY4")
		assert.Contains(t, r.URL.Query().Get("key"), "test-key")
		assert.Contains(t, r.URL.Query().Get("fields"), "place_id,name,rating")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.baseURL = server.URL

	details, err := client.GetPlaceDetails(context.Background(), "ChIJN1t_tDeuEmsRUsoyG83frY4")
	require.NoError(t, err)
	assert.Equal(t, "Joe's Pizza", details.Name)
	assert.Equal(t, float32(4.2), details.Rating)
	assert.Equal(t, 523, details.UserRatings)
	assert.Equal(t, 2, details.PriceLevel)
	assert.Equal(t, 1, len(details.Photos))
	assert.Equal(t, 1, len(details.Reviews))
}

func TestGetPlaceDetails_EmptyPlaceID(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.GetPlaceDetails(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "place ID cannot be empty")
}

func TestGetPlaceDetails_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.baseURL = server.URL

	_, err := client.GetPlaceDetails(context.Background(), "test-place-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 400")
}

func TestGetPlaceDetails_InvalidStatus(t *testing.T) {
	mockResponse := PlaceDetailsResponse{
		Status: "INVALID_REQUEST",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.baseURL = server.URL

	_, err := client.GetPlaceDetails(context.Background(), "test-place-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API returned status: INVALID_REQUEST")
}

func TestDownloadPhoto_Success(t *testing.T) {
	testPhotoData := []byte("fake-image-data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "photo_ref_123", r.URL.Query().Get("photoreference"))
		assert.Equal(t, "800", r.URL.Query().Get("maxwidth"))
		assert.Equal(t, "test-key", r.URL.Query().Get("key"))

		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(testPhotoData)
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.baseURL = server.URL

	data, err := client.DownloadPhoto(context.Background(), "photo_ref_123", 800)
	require.NoError(t, err)
	assert.Equal(t, testPhotoData, data)
}

func TestDownloadPhoto_EmptyReference(t *testing.T) {
	client := NewClient("test-key")
	_, err := client.DownloadPhoto(context.Background(), "", 800)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "photo reference cannot be empty")
}

func TestPlaceDetailsToPlace(t *testing.T) {
	details := &PlaceDetails{
		PlaceID:     "ChIJN1t_tDeuEmsRUsoyG83frY4",
		Name:        "Joe's Pizza",
		Rating:      4.2,
		UserRatings: 523,
		PriceLevel:  2,
		Website:     "https://joespizza.com",
		Phone:       "(718) 555-0123",
		Address:     "123 Main St, Brooklyn, NY 11201",
		Types:       []string{"restaurant", "food"},
		Geometry: PlaceGeometry{
			Location: PlaceLocation{
				Lat: 40.6892,
				Lng: -74.0445,
			},
		},
		Hours: &OpeningHours{
			WeekdayText: []string{"Monday: 11:00 AM – 10:00 PM"},
		},
		Photos: []PlacePhoto{
			{
				PhotoReference: "photo_ref_1",
				Width:          800,
				Height:         600,
			},
		},
		Reviews: []PlaceReview{
			{
				AuthorName:      "Sarah M.",
				Rating:          5,
				Text:            "Great pizza!",
				Time:            1642780800, // 2022-01-21
				ProfilePhotoURL: "https://example.com/photo.jpg",
			},
		},
	}

	place := details.ToPlace()

	assert.Equal(t, "ChIJN1t_tDeuEmsRUsoyG83frY4", place.PlaceID)
	assert.Equal(t, "Joe's Pizza", place.Name)
	assert.Equal(t, "123 Main St, Brooklyn, NY 11201", place.Address)
	assert.Equal(t, float32(4.2), place.Rating)
	assert.Equal(t, 523, place.UserRatings)
	assert.Equal(t, 2, place.PriceLevel)
	assert.Equal(t, "https://joespizza.com", place.Website)
	assert.Equal(t, "(718) 555-0123", place.Phone)
	assert.Equal(t, 40.6892, place.Coordinates.Lat)
	assert.Equal(t, -74.0445, place.Coordinates.Lng)
	assert.Equal(t, []string{"restaurant", "food"}, place.Categories)
	assert.Contains(t, place.Hours, "Monday: 11:00 AM – 10:00 PM")

	require.Equal(t, 1, len(place.Photos))
	assert.Equal(t, "photo_ref_1", place.Photos[0].Reference)
	assert.Equal(t, 800, place.Photos[0].Width)
	assert.Equal(t, 600, place.Photos[0].Height)

	require.Equal(t, 1, len(place.Reviews))
	assert.Equal(t, "Sarah M.", place.Reviews[0].Author)
	assert.Equal(t, 5, place.Reviews[0].Rating)
	assert.Equal(t, "Great pizza!", place.Reviews[0].Text)
	assert.Equal(t, "https://example.com/photo.jpg", place.Reviews[0].ProfilePhoto)
}

func TestPlaceDetailsToPlace_EmptyData(t *testing.T) {
	details := &PlaceDetails{
		PlaceID: "test-place-id",
		Name:    "Test Place",
	}

	place := details.ToPlace()

	assert.Equal(t, "test-place-id", place.PlaceID)
	assert.Equal(t, "Test Place", place.Name)
	assert.Empty(t, place.Photos)
	assert.Empty(t, place.Reviews)
	assert.Empty(t, place.Categories)
	assert.Empty(t, place.Hours)
}
