package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/export"
)

var (
	exportLimit int
)

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().IntVar(&exportLimit, "limit", 0, "maximum number of places to export (0 = all)")
}

var exportCmd = &cobra.Command{
	Use:   "export <format> <output-file>",
	Short: "Export places to various formats",
	Long: `Export your saved places to different formats including CSV, GeoJSON, JSON, and Markdown.

Supported formats:
  csv      - Comma-separated values for spreadsheet applications
  geojson  - Geographic data format for mapping applications
  json     - Raw JSON data
  markdown - Human-readable documentation format

Examples:
  placeli export csv places.csv
  placeli export geojson places.geojson
  placeli export markdown places.md
  placeli export json places.json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		format := args[0]
		outputFile := args[1]

		if err := export.ValidateFormat(format); err != nil {
			return err
		}

		// Use a large limit if exportLimit is 0 (export all)
		limit := exportLimit
		if limit == 0 {
			limit = 50000 // Use reasonable upper bound instead of hardcoded 10000
		}
		places, err := db.ListPlaces(limit, 0)
		if err != nil {
			return fmt.Errorf("failed to retrieve places: %w", err)
		}

		if len(places) == 0 {
			return fmt.Errorf("no places found to export")
		}

		if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()

		if err := export.Export(places, export.Format(format), file); err != nil {
			return fmt.Errorf("failed to export places: %w", err)
		}

		fmt.Printf("Successfully exported %d places to %s (%s format)\n",
			len(places), outputFile, strings.ToUpper(format))

		return nil
	},
}
