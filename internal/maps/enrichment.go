package maps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/placeli/internal/database"
	"github.com/user/placeli/internal/models"
)

type EnrichmentService struct {
	client   *Client
	db       *database.DB
	photoDir string
}

type EnrichmentOptions struct {
	FetchPhotos   bool
	FetchReviews  bool
	PhotoMaxWidth int
}

func NewEnrichmentService(apiKey string, db *database.DB, photoDir string) *EnrichmentService {
	return &EnrichmentService{
		client:   NewClient(apiKey),
		db:       db,
		photoDir: photoDir,
	}
}

func (s *EnrichmentService) EnrichPlace(ctx context.Context, place *models.Place, opts EnrichmentOptions) error {
	if place.PlaceID == "" {
		return fmt.Errorf("place %s has no PlaceID for enrichment", place.Name)
	}

	details, err := s.client.GetPlaceDetails(ctx, place.PlaceID)
	if err != nil {
		return fmt.Errorf("failed to get place details for %s: %w", place.Name, err)
	}

	s.mergeDetails(place, details, opts)

	if opts.FetchPhotos && len(details.Photos) > 0 {
		if err := s.downloadPhotos(ctx, place, details.Photos, opts.PhotoMaxWidth); err != nil {
			return fmt.Errorf("failed to download photos for %s: %w", place.Name, err)
		}
	}

	place.UpdatedAt = time.Now()
	if err := s.db.SavePlace(place); err != nil {
		return fmt.Errorf("failed to save enriched place %s: %w", place.Name, err)
	}

	return nil
}

func (s *EnrichmentService) EnrichAllPlaces(ctx context.Context, opts EnrichmentOptions) error {
	places, err := s.db.ListPlaces(10000, 0)
	if err != nil {
		return fmt.Errorf("failed to retrieve places: %w", err)
	}

	var enriched, skipped, failed int

	for _, place := range places {
		if place.PlaceID == "" {
			skipped++
			fmt.Printf("Skipping %s (no PlaceID)\n", place.Name)
			continue
		}

		fmt.Printf("Enriching %s...\n", place.Name)
		if err := s.EnrichPlace(ctx, place, opts); err != nil {
			failed++
			fmt.Printf("Failed to enrich %s: %v\n", place.Name, err)
			continue
		}

		enriched++
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Enrichment complete: %d enriched, %d skipped, %d failed\n",
		enriched, skipped, failed)
	return nil
}

func (s *EnrichmentService) mergeDetails(place *models.Place, details *PlaceDetails, opts EnrichmentOptions) {
	if details.Rating > 0 && details.Rating != place.Rating {
		place.Rating = details.Rating
	}
	if details.UserRatings > 0 && details.UserRatings != place.UserRatings {
		place.UserRatings = details.UserRatings
	}
	if details.PriceLevel > 0 && details.PriceLevel != place.PriceLevel {
		place.PriceLevel = details.PriceLevel
	}
	if details.Website != "" && details.Website != place.Website {
		place.Website = details.Website
	}
	if details.Phone != "" && details.Phone != place.Phone {
		place.Phone = details.Phone
	}
	if details.Hours != nil && len(details.Hours.WeekdayText) > 0 {
		hoursText := strings.Join(details.Hours.WeekdayText, ", ")
		if hoursText != place.Hours {
			place.Hours = hoursText
		}
	}

	if opts.FetchReviews && len(details.Reviews) > 0 {
		place.Reviews = nil
		for _, review := range details.Reviews {
			place.Reviews = append(place.Reviews, models.Review{
				Author:       review.AuthorName,
				Rating:       review.Rating,
				Text:         review.Text,
				Time:         time.Unix(review.Time, 0),
				ProfilePhoto: review.ProfilePhotoURL,
			})
		}
	}

	if len(details.Types) > 0 {
		uniqueTypes := make(map[string]bool)
		var newCategories []string

		for _, category := range place.Categories {
			if !uniqueTypes[category] {
				newCategories = append(newCategories, category)
				uniqueTypes[category] = true
			}
		}

		for _, typ := range details.Types {
			if !uniqueTypes[typ] {
				newCategories = append(newCategories, typ)
				uniqueTypes[typ] = true
			}
		}

		place.Categories = newCategories
	}
}

func (s *EnrichmentService) downloadPhotos(ctx context.Context, place *models.Place, photos []PlacePhoto, maxWidth int) error {
	if s.photoDir == "" {
		return fmt.Errorf("photo directory not configured")
	}

	if err := os.MkdirAll(s.photoDir, 0755); err != nil {
		return fmt.Errorf("failed to create photo directory: %w", err)
	}

	place.Photos = nil

	for i, photo := range photos {
		if i >= 5 {
			break
		}

		photoData, err := s.client.DownloadPhoto(ctx, photo.PhotoReference, maxWidth)
		if err != nil {
			fmt.Printf("Failed to download photo %d for %s: %v\n", i+1, place.Name, err)
			continue
		}

		filename := fmt.Sprintf("%s_%d.jpg",
			strings.ReplaceAll(strings.ToLower(place.Name), " ", "_"), i+1)
		photoPath := filepath.Join(s.photoDir, filename)

		if err := os.WriteFile(photoPath, photoData, 0644); err != nil {
			fmt.Printf("Failed to save photo %s: %v\n", photoPath, err)
			continue
		}

		place.Photos = append(place.Photos, models.Photo{
			Reference: photo.PhotoReference,
			LocalPath: photoPath,
			Width:     photo.Width,
			Height:    photo.Height,
		})

		time.Sleep(50 * time.Millisecond)
	}

	return nil
}
