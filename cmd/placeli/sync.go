package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/importer"
	"github.com/user/placeli/internal/models"
)

var (
	syncDryRun bool
	syncForce  bool
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "show what would be synced without making changes")
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "overwrite existing places with newer data")
}

var syncCmd = &cobra.Command{
	Use:   "sync <takeout-path>",
	Short: "Sync places from Google Takeout with existing database",
	Long: `Sync places from Google Takeout, intelligently merging new data without creating duplicates.

This command will:
- Add new places that don't exist in the database
- Update existing places with newer information (if --force is used)
- Skip places that already exist (unless --force is used)
- Preserve user-added notes and tags

The sync process is more conservative than import to prevent data loss.

Examples:
  placeli sync ~/Downloads/takeout-20240115
  placeli sync --dry-run ~/Downloads/takeout-20240115
  placeli sync --force ~/Downloads/takeout-20240115`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		takeoutPath := args[0]

		// Validate path
		if !filepath.IsAbs(takeoutPath) {
			abs, err := filepath.Abs(takeoutPath)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}
			takeoutPath = abs
		}

		fmt.Printf("Syncing from: %s\n", takeoutPath)
		if syncDryRun {
			fmt.Println("DRY RUN: No changes will be made")
		}

		// Import places from takeout
		newPlaces, err := importer.ImportFromTakeout(takeoutPath)
		if err != nil {
			return fmt.Errorf("failed to read takeout data: %w", err)
		}

		if len(newPlaces) == 0 {
			fmt.Println("No places found in takeout data")
			return nil
		}

		fmt.Printf("Found %d places in takeout data\n\n", len(newPlaces))

		// Sync with existing database
		added, updated, skipped, err := syncPlaces(newPlaces, syncDryRun, syncForce)
		if err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		// Report results
		fmt.Printf("\nSync complete:\n")
		fmt.Printf("  Added: %d new places\n", added)
		if syncForce {
			fmt.Printf("  Updated: %d existing places\n", updated)
		}
		fmt.Printf("  Skipped: %d duplicates\n", skipped)

		if syncDryRun {
			fmt.Println("\nRun without --dry-run to apply changes")
		}

		return nil
	},
}

type SyncResult struct {
	Action string      // "add", "update", "skip"
	Place  *models.Place
	Reason string
}

func syncPlaces(newPlaces []*models.Place, dryRun, force bool) (added, updated, skipped int, err error) {
	for i, place := range newPlaces {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(newPlaces), place.Name)

		result, err := syncSinglePlace(place, dryRun, force)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		switch result.Action {
		case "add":
			added++
			if !dryRun {
				fmt.Printf("  ✓ Added new place\n")
			} else {
				fmt.Printf("  ✓ Would add new place\n")
			}
		case "update":
			updated++
			if !dryRun {
				fmt.Printf("  ✓ Updated existing place\n")
			} else {
				fmt.Printf("  ✓ Would update existing place\n")
			}
		case "skip":
			skipped++
			fmt.Printf("  - Skipped: %s\n", result.Reason)
		}
	}

	return added, updated, skipped, nil
}

func syncSinglePlace(place *models.Place, dryRun, force bool) (*SyncResult, error) {
	// Check if place already exists by PlaceID or name+address
	existing, err := findExistingPlace(place)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		// New place - always add
		if !dryRun {
			if err := db.SavePlace(place); err != nil {
				return nil, fmt.Errorf("failed to save place: %w", err)
			}
		}
		return &SyncResult{Action: "add", Place: place}, nil
	}

	// Place exists
	if !force {
		return &SyncResult{
			Action: "skip", 
			Place:  place, 
			Reason: "already exists (use --force to update)",
		}, nil
	}

	// Update existing place with new data, preserving user data
	mergedPlace := mergePlace(existing, place)
	if !dryRun {
		if err := db.SavePlace(mergedPlace); err != nil {
			return nil, fmt.Errorf("failed to update place: %w", err)
		}
	}

	return &SyncResult{Action: "update", Place: mergedPlace}, nil
}

func findExistingPlace(place *models.Place) (*models.Place, error) {
	// First try to find by PlaceID
	if place.PlaceID != "" {
		existing, err := db.GetPlace(place.ID)
		if err == nil {
			return existing, nil
		}
	}

	// If not found by PlaceID, search by name and address
	if place.Name != "" {
		searchQuery := place.Name
		if place.Address != "" {
			searchQuery += " " + place.Address
		}
		
		places, err := db.SearchPlaces(searchQuery)
		if err != nil {
			return nil, err
		}

		// Look for exact match
		for _, existing := range places {
			if existing.Name == place.Name && existing.Address == place.Address {
				return existing, nil
			}
		}
	}

	return nil, nil // Not found
}

func mergePlace(existing, new *models.Place) *models.Place {
	// Start with the new place data
	merged := *new
	
	// Preserve the original ID and timestamps
	merged.ID = existing.ID
	merged.CreatedAt = existing.CreatedAt
	merged.UpdatedAt = time.Now()
	
	// Preserve user data (notes, tags, custom fields)
	merged.UserNotes = existing.UserNotes
	merged.UserTags = existing.UserTags
	
	// Merge custom fields, preserving user-added fields
	if existing.CustomFields != nil {
		if merged.CustomFields == nil {
			merged.CustomFields = make(map[string]interface{})
		}
		
		// Preserve user-added custom fields (not from import)
		for key, value := range existing.CustomFields {
			if key != "google_maps_url" && key != "imported_from" && key != "import_date" {
				merged.CustomFields[key] = value
			}
		}
	}
	
	// Update import metadata
	if merged.CustomFields == nil {
		merged.CustomFields = make(map[string]interface{})
	}
	merged.CustomFields["last_sync"] = time.Now().Format(time.RFC3339)
	
	return &merged
}