package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/logger"
	"github.com/user/placeli/internal/models"
	"github.com/user/placeli/internal/tui/mapview"
)

var (
	listLimit    int
	listOffset   int
	listFormat   string
	listSearch   string
	listMap      bool
	listMapSize  int
)

func init() {
	rootCmd.AddCommand(listCmd)
	
	listCmd.Flags().IntVar(&listLimit, "limit", 20, "maximum number of places to show")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "number of places to skip")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "output format: table, simple, json, map")
	listCmd.Flags().StringVar(&listSearch, "search", "", "search query to filter places")
	listCmd.Flags().BoolVar(&listMap, "map", false, "show mini-map of places")
	listCmd.Flags().IntVar(&listMapSize, "map-size", 20, "size of mini-map (width)")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List places in the database",
	Long: `List places in the database with various output formats and filtering options.

Output formats:
  table  - Formatted table with columns (default)
  simple - Simple line-by-line output
  json   - JSON array of places
  map    - ASCII map view only

The --map flag adds a mini-map to table and simple formats.
Use --search to filter places by name, address, or notes.

Examples:
  placeli list                           # List first 20 places
  placeli list --limit=50               # List first 50 places
  placeli list --search="coffee"        # Search for coffee places
  placeli list --format=json            # Output as JSON
  placeli list --map                    # Include mini-map
  placeli list --format=map             # Show only map`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Listing places", 
			"limit", listLimit, 
			"offset", listOffset,
			"format", listFormat,
			"search", listSearch,
			"map", listMap)
		
		// Get places
		var places []*models.Place
		var err error
		
		if listSearch != "" {
			places, err = db.SearchPlaces(listSearch)
			if err != nil {
				return fmt.Errorf("failed to search places: %w", err)
			}
			
			// Apply offset and limit to search results
			if listOffset >= len(places) {
				places = []*models.Place{}
			} else {
				end := listOffset + listLimit
				if end > len(places) {
					end = len(places)
				}
				places = places[listOffset:end]
			}
		} else {
			places, err = db.ListPlaces(listLimit, listOffset)
			if err != nil {
				return fmt.Errorf("failed to list places: %w", err)
			}
		}
		
		// Display results
		switch listFormat {
		case "table":
			return displayTable(places)
		case "simple":
			return displaySimple(places)
		case "json":
			return displayJSON(places)
		case "map":
			return displayMapOnly(places)
		default:
			return fmt.Errorf("unknown format: %s", listFormat)
		}
	},
}

func displayTable(places []*models.Place) error {
	if len(places) == 0 {
		fmt.Println("No places found")
		return nil
	}
	
	// Show mini-map if requested
	if listMap {
		err := displayMiniMap(places)
		if err != nil {
			fmt.Printf("Warning: could not display map: %v\n", err)
		}
		fmt.Println()
	}
	
	// Create table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	// Header
	fmt.Fprintln(w, "NAME\tADDRESS\tRATING\tCATEGORIES\tTAGS")
	fmt.Fprintln(w, "----\t-------\t------\t----------\t----")
	
	// Rows
	for _, place := range places {
		name := place.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		
		address := place.Address
		if len(address) > 40 {
			address = address[:37] + "..."
		}
		
		rating := ""
		if place.Rating > 0 {
			stars := strings.Repeat("⭐", int(place.Rating))
			rating = fmt.Sprintf("%.1f %s", place.Rating, stars)
		}
		
		categories := strings.Join(place.Categories, ", ")
		if len(categories) > 20 {
			categories = categories[:17] + "..."
		}
		
		tags := strings.Join(place.UserTags, ", ")
		if len(tags) > 15 {
			tags = tags[:12] + "..."
		}
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
			name, address, rating, categories, tags)
	}
	
	return w.Flush()
}

func displaySimple(places []*models.Place) error {
	if len(places) == 0 {
		fmt.Println("No places found")
		return nil
	}
	
	// Show mini-map if requested
	if listMap {
		err := displayMiniMap(places)
		if err != nil {
			fmt.Printf("Warning: could not display map: %v\n", err)
		}
		fmt.Println()
	}
	
	for i, place := range places {
		fmt.Printf("%d. %s\n", i+1+listOffset, place.Name)
		
		if place.Address != "" {
			fmt.Printf("   📍 %s\n", place.Address)
		}
		
		if place.Rating > 0 {
			stars := strings.Repeat("⭐", int(place.Rating))
			fmt.Printf("   ⭐ %.1f %s\n", place.Rating, stars)
		}
		
		if len(place.Categories) > 0 {
			fmt.Printf("   🏷️  %s\n", strings.Join(place.Categories, ", "))
		}
		
		if len(place.UserTags) > 0 {
			fmt.Printf("   🔖 %s\n", strings.Join(place.UserTags, ", "))
		}
		
		if place.UserNotes != "" {
			notes := place.UserNotes
			if len(notes) > 100 {
				notes = notes[:97] + "..."
			}
			fmt.Printf("   📝 %s\n", notes)
		}
		
		fmt.Println()
	}
	
	return nil
}

func displayJSON(places []*models.Place) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(places)
}

func displayMapOnly(places []*models.Place) error {
	if len(places) == 0 {
		fmt.Println("No places to display on map")
		return nil
	}
	
	// Create a larger map for map-only view
	config := mapview.MapConfig{
		Width:      80,
		Height:     30,
		ZoomLevel:  10,
		ShowLabels: true,
	}
	
	mapView := mapview.NewMapView(places, config)
	mapView.FitBounds()
	
	fmt.Print(mapView.Render())
	
	return nil
}

func displayMiniMap(places []*models.Place) error {
	if len(places) == 0 {
		return nil
	}
	
	// Create small map
	size := listMapSize
	if size < 10 {
		size = 10
	}
	if size > 50 {
		size = 50
	}
	
	config := mapview.MapConfig{
		Width:      size,
		Height:     size/2,
		ZoomLevel:  10,
		ShowLabels: false,
	}
	
	mapView := mapview.NewMapView(places, config)
	mapView.FitBounds()
	
	fmt.Printf("Mini-map (%d places):\n", len(places))
	fmt.Print(mapView.Render())
	
	return nil
}