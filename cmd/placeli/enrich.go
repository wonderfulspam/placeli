package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/maps"
)

var (
	enrichAPIKey     string
	enrichPhotoDir   string
	enrichPlaceID    string
	enrichPhotos     bool
	enrichReviews    bool
	enrichPhotoWidth int
	enrichTimeout    int
)

var enrichCmd = &cobra.Command{
	Use:   "enrich",
	Short: "Enrich places with Google Maps API data",
	Long: `Enrich your saved places with additional data from Google Maps API including:
  - Updated ratings and review counts
  - Current business hours
  - Phone numbers and websites
  - Photos (with --photos flag)
  - Latest reviews (with --reviews flag)

Requires a Google Maps API key with Places API enabled.
Set the API key using the --api-key flag or GOOGLE_MAPS_API_KEY environment variable.

Examples:
  placeli enrich --api-key=YOUR_API_KEY
  placeli enrich --place-id=ChIJN1t_tDeuEmsRUsoyG83frY4
  placeli enrich --photos --photo-dir=./photos
  placeli enrich --reviews --api-key=YOUR_API_KEY`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey := enrichAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_MAPS_API_KEY")
		}
		if apiKey == "" {
			return fmt.Errorf("Google Maps API key required. Use --api-key flag or set GOOGLE_MAPS_API_KEY environment variable")
		}

		if enrichPhotos && enrichPhotoDir == "" {
			homeDir, _ := os.UserHomeDir()
			enrichPhotoDir = filepath.Join(homeDir, ".placeli", "photos")
		}

		service := maps.NewEnrichmentService(apiKey, db, enrichPhotoDir)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(enrichTimeout)*time.Second)
		defer cancel()

		opts := maps.EnrichmentOptions{
			FetchPhotos:   enrichPhotos,
			FetchReviews:  enrichReviews,
			PhotoMaxWidth: enrichPhotoWidth,
		}

		if enrichPlaceID != "" {
			place, err := db.GetPlace(enrichPlaceID)
			if err != nil {
				return fmt.Errorf("failed to find place with ID %s: %w", enrichPlaceID, err)
			}

			fmt.Printf("Enriching %s...\n", place.Name)
			if err := service.EnrichPlace(ctx, place, opts); err != nil {
				return fmt.Errorf("failed to enrich place: %w", err)
			}

			fmt.Printf("Successfully enriched %s\n", place.Name)
			return nil
		}

		fmt.Printf("Starting enrichment of all places...\n")
		if enrichPhotos {
			fmt.Printf("Photos will be saved to: %s\n", enrichPhotoDir)
		}

		return service.EnrichAllPlaces(ctx, opts)
	},
}

func init() {
	enrichCmd.Flags().StringVar(&enrichAPIKey, "api-key", "", "Google Maps API key")
	enrichCmd.Flags().StringVar(&enrichPhotoDir, "photo-dir", "", "directory to save photos (default: ~/.placeli/photos)")
	enrichCmd.Flags().StringVar(&enrichPlaceID, "place-id", "", "enrich specific place by ID")
	enrichCmd.Flags().BoolVar(&enrichPhotos, "photos", false, "download place photos")
	enrichCmd.Flags().BoolVar(&enrichReviews, "reviews", false, "fetch latest reviews")
	enrichCmd.Flags().IntVar(&enrichPhotoWidth, "photo-width", 800, "maximum photo width in pixels")
	enrichCmd.Flags().IntVar(&enrichTimeout, "timeout", 300, "timeout in seconds for enrichment process")

	rootCmd.AddCommand(enrichCmd)
}
