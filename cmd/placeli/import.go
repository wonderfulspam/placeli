package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/importer/sources"
	"github.com/user/placeli/internal/logger"
	"github.com/user/placeli/internal/models"
)

var (
	importSource  string
	importDryRun  bool
	importForce   bool
	importNoMerge bool
)

func init() {
	rootCmd.AddCommand(importCmd)

	// Add subcommands
	importCmd.AddCommand(importSourcesCmd)
	importCmd.AddCommand(importFromCmd)

	// Flags for import command
	importFromCmd.Flags().StringVar(&importSource, "source", "auto", "import source: auto, apple, osm, foursquare, takeout")
	importFromCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "show what would be imported without making changes")
	importFromCmd.Flags().BoolVar(&importForce, "force", false, "overwrite existing places (default: skip duplicates)")
	importFromCmd.Flags().BoolVar(&importNoMerge, "no-merge", false, "disable merging and treat all places as new")
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import places from various sources",
	Long: `Import places from multiple sources including Apple Maps, OpenStreetMap,
Foursquare/Swarm, and Google Takeout.

By default, import will:
- Detect and skip duplicate places
- Preserve existing user data (notes, tags, custom fields)
- Only update places when --force is used

Supported formats:
  Apple Maps     - KML, KMZ, GPX files
  OpenStreetMap  - JSON, CSV exports
  Foursquare     - JSON export files
  Google Takeout - ZIP archives, JSON (Maps), CSV (Saved)

Available subcommands:
  sources - List available import sources
  from    - Import from a specific file or directory

Examples:
  placeli import sources
  placeli import from ~/Downloads/takeout.zip
  placeli import from ~/Downloads/places.kml
  placeli import from ~/Downloads/saved-places.csv --source=takeout
  placeli import from ~/Downloads/places.json --force`,
}

var importSourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List available import sources",
	Long:  `Show all available import sources and their supported formats.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sm := sources.NewSourceManager()

		fmt.Println("Available import sources:")

		for name, source := range sm.ListSources() {
			fmt.Printf("📁 %s (%s)\n", source.Name(), name)
			fmt.Printf("   Supported formats: %s\n", strings.Join(source.SupportedFormats(), ", "))
			fmt.Println()
		}

		fmt.Println("Use 'placeli import from <file>' to import from any supported format.")
		fmt.Println("Add --source=<name> to force a specific source.")

		return nil
	},
}

var importFromCmd = &cobra.Command{
	Use:   "from <file-path>",
	Short: "Import places from a file",
	Long: `Import places from a file or directory using automatic source detection or a specified source.

The import process will:
- Detect the appropriate source based on file format
- Parse places from the file
- Check for duplicates using source hashes (unless --no-merge is used)
- Skip existing places (unless --force is used)
- Preserve existing user data (notes, tags, custom fields) when updating

Use --dry-run to preview what would be imported.
Use --source to force a specific import source.
Use --force to update existing places with new data.
Use --no-merge to disable duplicate detection and import all as new.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Validate path
		if !filepath.IsAbs(filePath) {
			abs, err := filepath.Abs(filePath)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}
			filePath = abs
		}

		logger.Info("Starting import",
			"file", filePath,
			"source", importSource,
			"dry_run", importDryRun,
			"force", importForce,
			"no_merge", importNoMerge)

		sm := sources.NewSourceManager()

		var places []*models.Place
		var err error
		var sourceName string

		if importSource == "auto" {
			// Auto-detect source
			source := sm.DetectSource(filePath)
			if source == nil {
				// Try each source to see if any can handle the file
				for name, src := range sm.ListSources() {
					places, err = src.ImportFromFile(filePath)
					if err == nil && len(places) > 0 {
						source = src
						sourceName = name
						break
					}
				}
				if source == nil {
					return fmt.Errorf("could not detect source for file: %s", filePath)
				}
			} else {
				sourceName = getSourceName(sm, source)
				places, err = source.ImportFromFile(filePath)
			}
		} else {
			// Use specified source
			source := sm.GetSource(importSource)
			if source == nil {
				return fmt.Errorf("unknown source: %s", importSource)
			}
			sourceName = importSource
			places, err = source.ImportFromFile(filePath)
		}

		if err != nil {
			return fmt.Errorf("failed to import from %s: %w", sourceName, err)
		}

		if len(places) == 0 {
			fmt.Printf("No places found in %s\n", filePath)
			return nil
		}

		logger.Info("Parsed places", "count", len(places), "source", sourceName)

		// Process the places (check for duplicates, save to database)
		var added, updated, skipped int
		if importNoMerge {
			// Simple import without duplicate checking
			added, updated, skipped, err = simpleImportPlaces(places, importDryRun)
		} else {
			// Smart import with duplicate detection and merging
			added, updated, skipped, err = smartImportPlaces(places, importDryRun, importForce)
		}

		if err != nil {
			return fmt.Errorf("failed to process places: %w", err)
		}

		// Report results
		fmt.Printf("\nImport complete from %s:\n", sourceName)
		fmt.Printf("  Added:   %d places\n", added)
		fmt.Printf("  Updated: %d places\n", updated)
		fmt.Printf("  Skipped: %d places\n", skipped)

		if importDryRun {
			fmt.Println("\nRun without --dry-run to apply changes")
		}

		logger.Info("Import complete",
			"source", sourceName,
			"added", added,
			"updated", updated,
			"skipped", skipped)

		return nil
	},
}

func getSourceName(sm *sources.SourceManager, source sources.ImportSource) string {
	for name, src := range sm.ListSources() {
		if src == source {
			return name
		}
	}
	return "unknown"
}

func simpleImportPlaces(places []*models.Place, dryRun bool) (added, updated, skipped int, err error) {
	for i, place := range places {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(places), place.Name)

		if !dryRun {
			if err := db.SavePlace(place); err != nil {
				fmt.Printf("  Error saving place: %v\n", err)
				skipped++
				continue
			}
		}

		added++
		if dryRun {
			fmt.Printf("  ✓ Would add new place\n")
		} else {
			fmt.Printf("  ✓ Added new place\n")
		}
	}

	return added, updated, skipped, nil
}

func smartImportPlaces(places []*models.Place, dryRun, force bool) (added, updated, skipped int, err error) {
	for i, place := range places {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(places), place.Name)

		// Check for existing place by source hash
		existing, err := findExistingPlace(place)
		if err != nil {
			fmt.Printf("  Error checking for duplicates: %v\n", err)
			continue
		}

		if existing == nil {
			// New place
			if !dryRun {
				if err := db.SavePlace(place); err != nil {
					fmt.Printf("  Error saving place: %v\n", err)
					continue
				}
			}

			added++
			if dryRun {
				fmt.Printf("  ✓ Would add new place\n")
			} else {
				fmt.Printf("  ✓ Added new place\n")
			}
		} else if force {
			// Update existing place
			mergedPlace := mergeImportedPlace(existing, place)

			if !dryRun {
				if err := db.SavePlace(mergedPlace); err != nil {
					fmt.Printf("  Error updating place: %v\n", err)
					continue
				}
			}

			updated++
			if dryRun {
				fmt.Printf("  ✓ Would update existing place\n")
			} else {
				fmt.Printf("  ✓ Updated existing place\n")
			}
		} else {
			// Skip existing place
			skipped++
			fmt.Printf("  - Skipped: already exists (use --force to update)\n")
		}
	}

	return added, updated, skipped, nil
}

func findExistingPlace(place *models.Place) (*models.Place, error) {
	// First try to find by source hash (most reliable)
	if place.SourceHash != "" {
		existing, err := db.FindPlaceBySourceHash(place.SourceHash)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	// Try to find potential duplicates by coordinates and place_id
	candidates, err := db.FindDuplicateCandidates(place)
	if err != nil {
		return nil, err
	}

	// Return the first candidate if any found
	if len(candidates) > 0 {
		return candidates[0], nil
	}

	return nil, nil
}

func mergeImportedPlace(existing, imported *models.Place) *models.Place {
	// Start with imported data
	merged := *imported

	// Preserve existing user data
	merged.ID = existing.ID
	merged.CreatedAt = existing.CreatedAt
	merged.UpdatedAt = time.Now()

	// Preserve user data
	merged.UserNotes = existing.UserNotes
	merged.UserTags = existing.UserTags

	// Merge custom fields, preserving user-added fields
	if existing.CustomFields != nil {
		if merged.CustomFields == nil {
			merged.CustomFields = make(map[string]interface{})
		}

		// Preserve user-added custom fields (not from import systems)
		systemPrefixes := []string{"google_", "osm_", "apple_", "foursquare_", "gpx_", "imported_from", "import_date"}
		for key, value := range existing.CustomFields {
			isSystemField := false
			for _, prefix := range systemPrefixes {
				if strings.HasPrefix(key, prefix) || key == prefix {
					isSystemField = true
					break
				}
			}

			// Keep user fields
			if !isSystemField {
				merged.CustomFields[key] = value
			}
		}
	}

	// Update import metadata
	if merged.CustomFields == nil {
		merged.CustomFields = make(map[string]interface{})
	}
	merged.CustomFields["last_import"] = time.Now().Format(time.RFC3339)

	return &merged
}
