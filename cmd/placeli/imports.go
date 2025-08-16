package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/importer/sources"
	"github.com/user/placeli/internal/logger"
	"github.com/user/placeli/internal/models"
)

var (
	importSource   string
	importDryRun   bool
	importOverwrite bool
)

func init() {
	rootCmd.AddCommand(importsCmd)
	
	// Add subcommands
	importsCmd.AddCommand(importSourcesCmd)
	importsCmd.AddCommand(importFromCmd)
	
	// Flags for import command
	importFromCmd.Flags().StringVar(&importSource, "source", "auto", "import source: auto, apple, osm, foursquare, takeout")
	importFromCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "show what would be imported without making changes")
	importFromCmd.Flags().BoolVar(&importOverwrite, "overwrite", false, "overwrite existing places with same source hash")
}

var importsCmd = &cobra.Command{
	Use:   "imports",
	Short: "Import places from various sources",
	Long: `Import places from multiple sources including Apple Maps, OpenStreetMap, 
Foursquare/Swarm, and Google Takeout.

Supported formats:
  Apple Maps    - KML, GPX files
  OpenStreetMap - JSON, CSV exports  
  Foursquare    - JSON export files
  Google Takeout - JSON saved places

Available subcommands:
  sources - List available import sources
  from    - Import from a specific file

Examples:
  placeli imports sources
  placeli imports from ~/Downloads/places.kml
  placeli imports from ~/Downloads/foursquare.json --source=foursquare
  placeli imports from ~/Downloads/osm-data.csv --source=osm`,
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
		
		fmt.Println("Use 'placeli imports from <file>' to import from any supported format.")
		fmt.Println("Add --source=<name> to force a specific source.")
		
		return nil
	},
}

var importFromCmd = &cobra.Command{
	Use:   "from <file-path>",
	Short: "Import places from a file",
	Long: `Import places from a file using automatic source detection or a specified source.

The import process will:
- Detect the appropriate source based on file format
- Parse places from the file  
- Check for duplicates using source hashes
- Add new places to the database
- Preserve existing user data (notes, tags, custom fields)

Use --dry-run to preview what would be imported.
Use --source to force a specific import source.
Use --overwrite to replace existing places with the same source hash.`,
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
			"overwrite", importOverwrite)
		
		sm := sources.NewSourceManager()
		
		var places []*models.Place
		var err error
		var sourceName string
		
		if importSource == "auto" {
			// Auto-detect source
			source := sm.DetectSource(filePath)
			if source == nil {
				return fmt.Errorf("could not detect source for file: %s", filePath)
			}
			sourceName = source.Name()
			places, err = source.ImportFromFile(filePath)
		} else {
			// Use specified source
			source := sm.GetSource(importSource)
			if source == nil {
				return fmt.Errorf("unknown source: %s", importSource)
			}
			sourceName = source.Name()
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
		added, updated, skipped, err := processImportedPlaces(places, importDryRun, importOverwrite)
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

func processImportedPlaces(places []*models.Place, dryRun, overwrite bool) (added, updated, skipped int, err error) {
	for i, place := range places {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(places), place.Name)
		
		// Check for existing place by source hash
		existing, err := db.FindPlaceBySourceHash(place.SourceHash)
		if err != nil {
			fmt.Printf("  Error checking for duplicates: %v\n", err)
			continue
		}
		
		if existing != nil {
			// Place exists
			if !overwrite {
				skipped++
				fmt.Printf("  - Skipped: already exists (use --overwrite to update)\n")
				continue
			}
			
			// Merge with existing place, preserving user data
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
		}
	}
	
	return added, updated, skipped, nil
}

func mergeImportedPlace(existing, imported *models.Place) *models.Place {
	// Start with imported data
	merged := *imported
	
	// Preserve existing user data
	merged.ID = existing.ID
	merged.CreatedAt = existing.CreatedAt
	merged.UserNotes = existing.UserNotes
	merged.UserTags = existing.UserTags
	
	// Merge custom fields, preserving user-added fields
	if existing.CustomFields != nil {
		if merged.CustomFields == nil {
			merged.CustomFields = make(map[string]interface{})
		}
		
		// Preserve user-added custom fields (not from import systems)
		systemPrefixes := []string{"google_", "osm_", "apple_", "foursquare_", "gpx_"}
		for key, value := range existing.CustomFields {
			isSystemField := false
			for _, prefix := range systemPrefixes {
				if strings.HasPrefix(key, prefix) {
					isSystemField = true
					break
				}
			}
			
			// Keep user fields and certain system fields
			if !isSystemField || key == "imported_from" || key == "import_date" {
				merged.CustomFields[key] = value
			}
		}
	}
	
	return &merged
}