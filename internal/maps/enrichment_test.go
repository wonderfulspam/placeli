package maps

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/placeli/internal/models"
)

func TestMergeDetails(t *testing.T) {
	service := &EnrichmentService{}

	place := &models.Place{
		Name:        "Old Pizza Place",
		Rating:      3.5,
		UserRatings: 100,
		PriceLevel:  1,
		Website:     "old-website.com",
		Phone:       "555-0000",
		Hours:       "Mon-Fri 9-5",
		Categories:  []string{"restaurant"},
		Reviews:     []models.Review{{Author: "Old Review", Text: "Meh"}},
	}

	details := &PlaceDetails{
		Name:        "New Pizza Place",
		Rating:      4.2,
		UserRatings: 523,
		PriceLevel:  2,
		Website:     "new-website.com",
		Phone:       "555-1234",
		Types:       []string{"restaurant", "pizza", "food"},
		Hours: &OpeningHours{
			WeekdayText: []string{"Monday: 11:00 AM – 10:00 PM", "Tuesday: 11:00 AM – 10:00 PM"},
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
	}

	opts := EnrichmentOptions{
		FetchReviews: true,
	}

	service.mergeDetails(place, details, opts)

	assert.Equal(t, float32(4.2), place.Rating)
	assert.Equal(t, 523, place.UserRatings)
	assert.Equal(t, 2, place.PriceLevel)
	assert.Equal(t, "new-website.com", place.Website)
	assert.Equal(t, "555-1234", place.Phone)
	assert.Contains(t, place.Hours, "Monday: 11:00 AM – 10:00 PM")

	assert.Equal(t, 3, len(place.Categories))
	assert.Contains(t, place.Categories, "restaurant")
	assert.Contains(t, place.Categories, "pizza")
	assert.Contains(t, place.Categories, "food")

	require.Equal(t, 1, len(place.Reviews))
	assert.Equal(t, "Sarah M.", place.Reviews[0].Author)
	assert.Equal(t, "Great pizza!", place.Reviews[0].Text)
}

func TestMergeDetails_NoOverwrite(t *testing.T) {
	service := &EnrichmentService{}

	place := &models.Place{
		Rating:      4.5,
		UserRatings: 1000,
		PriceLevel:  3,
		Website:     "existing-website.com",
		Phone:       "555-1111",
	}

	details := &PlaceDetails{
		Rating:      0,
		UserRatings: 0,
		PriceLevel:  0,
		Website:     "",
		Phone:       "",
	}

	opts := EnrichmentOptions{}

	service.mergeDetails(place, details, opts)

	assert.Equal(t, float32(4.5), place.Rating)
	assert.Equal(t, 1000, place.UserRatings)
	assert.Equal(t, 3, place.PriceLevel)
	assert.Equal(t, "existing-website.com", place.Website)
	assert.Equal(t, "555-1111", place.Phone)
}

func TestMergeDetails_UniqueCategories(t *testing.T) {
	service := &EnrichmentService{}

	place := &models.Place{
		Categories: []string{"restaurant", "pizza"},
	}

	details := &PlaceDetails{
		Types: []string{"restaurant", "food", "establishment", "pizza"},
	}

	opts := EnrichmentOptions{}

	service.mergeDetails(place, details, opts)

	assert.Equal(t, 4, len(place.Categories))
	assert.Contains(t, place.Categories, "restaurant")
	assert.Contains(t, place.Categories, "pizza")
	assert.Contains(t, place.Categories, "food")
	assert.Contains(t, place.Categories, "establishment")

	categoryCount := make(map[string]int)
	for _, cat := range place.Categories {
		categoryCount[cat]++
	}

	for cat, count := range categoryCount {
		assert.Equal(t, 1, count, "Category %s appears %d times", cat, count)
	}
}

func TestMergeDetails_SkipReviewsWhenNotRequested(t *testing.T) {
	service := &EnrichmentService{}

	place := &models.Place{
		Reviews: []models.Review{{Author: "Original Review", Text: "Original"}},
	}

	details := &PlaceDetails{
		Reviews: []PlaceReview{
			{AuthorName: "New Review", Text: "New"},
		},
	}

	opts := EnrichmentOptions{
		FetchReviews: false,
	}

	service.mergeDetails(place, details, opts)

	require.Equal(t, 1, len(place.Reviews))
	assert.Equal(t, "Original Review", place.Reviews[0].Author)
	assert.Equal(t, "Original", place.Reviews[0].Text)
}

func TestEnrichPlace_NoPlaceID(t *testing.T) {
	service := &EnrichmentService{}
	place := &models.Place{
		Name:    "Test Place",
		PlaceID: "",
	}

	err := service.EnrichPlace(context.Background(), place, EnrichmentOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has no PlaceID for enrichment")
}
