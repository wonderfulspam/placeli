package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/importer"
)

var importCmd = &cobra.Command{
	Use:   "import <path>",
	Short: "Import places from Google Takeout",
	Long:  "Import saved places from a Google Takeout export directory.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		takeoutPath := args[0]

		if _, err := os.Stat(takeoutPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", takeoutPath)
		}

		fmt.Printf("Importing places from %s...\n", takeoutPath)
		places, err := importer.ImportFromTakeout(takeoutPath)
		if err != nil {
			return fmt.Errorf("failed to import from takeout: %w", err)
		}

		fmt.Printf("Found %d places to import\n", len(places))

		imported := 0
		for _, place := range places {
			if err := db.SavePlace(place); err != nil {
				fmt.Printf("Warning: failed to save place %s: %v\n", place.Name, err)
				continue
			}
			imported++
		}

		fmt.Printf("Successfully imported %d places\n", imported)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
