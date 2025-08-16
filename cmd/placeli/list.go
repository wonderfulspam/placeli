package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/models"
)

var (
	listLimit  int
	listOffset int
	listFormat string
	listQuery  string
)

const (
	tableNameWidth     = 28
	tableAddressWidth  = 38
	tableCategoryWidth = 13
	tableTotalWidth    = 100
)

var queryCmd = &cobra.Command{
	Use:     "query",
	Aliases: []string{"list"}, // Keep backward compatibility
	Short:   "Query saved places",
	Long:    "Query places from the local database with optional filtering and formatting.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var places, err = func() (interface{}, error) {
			if listQuery != "" {
				return db.SearchPlaces(listQuery)
			}
			return db.ListPlaces(listLimit, listOffset)
		}()

		if err != nil {
			return fmt.Errorf("failed to retrieve places: %w", err)
		}

		switch listFormat {
		case "json":
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(places)

		case "table":
			return printTable(places)

		default:
			return printSimple(places)
		}
	},
}

func printSimple(places interface{}) error {
	placeList, ok := places.([]*models.Place)
	if !ok {
		return fmt.Errorf("invalid places data")
	}

	for i, place := range placeList {
		fmt.Printf("%d. %s\n", i+1, place.Name)
		if place.Address != "" {
			fmt.Printf("   📍 %s\n", place.Address)
		}
		if place.Rating > 0 {
			stars := strings.Repeat("⭐", int(place.Rating))
			fmt.Printf("   %s %.1f (%d reviews)\n", stars, place.Rating, place.UserRatings)
		}
		if len(place.Categories) > 0 {
			fmt.Printf("   🏷️  %s\n", strings.Join(place.Categories, ", "))
		}
		if len(place.UserTags) > 0 {
			fmt.Printf("   🏷️  Tags: %s\n", strings.Join(place.UserTags, ", "))
		}
		if place.UserNotes != "" {
			fmt.Printf("   📝 %s\n", place.UserNotes)
		}
		fmt.Println()
	}

	fmt.Printf("Showing %d places\n", len(placeList))
	return nil
}

func printTable(places interface{}) error {
	placeList, ok := places.([]*models.Place)
	if !ok {
		return fmt.Errorf("invalid places data")
	}

	fmt.Printf("%-4s %-30s %-40s %-8s %-15s\n", "ID", "Name", "Address", "Rating", "Categories")
	fmt.Println(strings.Repeat("-", tableTotalWidth))

	for i, place := range placeList {
		name := truncate(place.Name, tableNameWidth)
		address := truncate(place.Address, tableAddressWidth)
		rating := ""
		if place.Rating > 0 {
			rating = fmt.Sprintf("%.1f", place.Rating)
		}
		categories := truncate(strings.Join(place.Categories, ","), tableCategoryWidth)

		fmt.Printf("%-4s %-30s %-40s %-8s %-15s\n",
			strconv.Itoa(i+1), name, address, rating, categories)
	}

	fmt.Printf("\nShowing %d places\n", len(placeList))
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	queryCmd.Flags().IntVarP(&listLimit, "limit", "l", 50, "maximum number of places to show")
	queryCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, "number of places to skip")
	queryCmd.Flags().StringVarP(&listFormat, "format", "f", "simple", "output format (simple, table, json)")
	queryCmd.Flags().StringVarP(&listQuery, "query", "q", "", "search query")

	rootCmd.AddCommand(queryCmd)
}
