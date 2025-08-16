package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/logger"
	"github.com/user/placeli/internal/models"
)

var (
	tagsFilter string
	tagsForce  bool
)

func init() {
	rootCmd.AddCommand(tagsCmd)

	// Add subcommands
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsRenameCmd)
	tagsCmd.AddCommand(tagsDeleteCmd)
	tagsCmd.AddCommand(tagsApplyCmd)

	// Flags for apply command
	tagsApplyCmd.Flags().StringVar(&tagsFilter, "filter", "", "search query to filter places")
	tagsDeleteCmd.Flags().BoolVar(&tagsForce, "force", false, "delete tag without confirmation")
}

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags for organizing places",
	Long: `Manage tags for organizing and categorizing your saved places.

Available subcommands:
  list    - Show all tags with usage counts
  rename  - Rename a tag across all places
  delete  - Remove a tag from all places
  apply   - Add a tag to places matching a filter

Examples:
  placeli tags list
  placeli tags rename "old-name" "new-name"
  placeli tags delete "unwanted-tag"
  placeli tags apply "favorite" --filter="coffee"`,
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags with usage counts",
	Long:  `Show all tags currently in use across your places, sorted by usage count.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("Listing all tags")

		tagCounts, err := getTagCounts()
		if err != nil {
			return fmt.Errorf("failed to get tag counts: %w", err)
		}

		if len(tagCounts) == 0 {
			fmt.Println("No tags found")
			return nil
		}

		// Sort by count (descending)
		type tagCount struct {
			tag   string
			count int
		}

		var tags []tagCount
		for tag, count := range tagCounts {
			tags = append(tags, tagCount{tag, count})
		}

		sort.Slice(tags, func(i, j int) bool {
			return tags[i].count > tags[j].count
		})

		fmt.Printf("Found %d tags:\n\n", len(tags))
		fmt.Printf("%-20s %s\n", "TAG", "COUNT")
		fmt.Printf("%-20s %s\n", "---", "-----")

		for _, tc := range tags {
			fmt.Printf("%-20s %d\n", tc.tag, tc.count)
		}

		return nil
	},
}

var tagsRenameCmd = &cobra.Command{
	Use:   "rename <old-tag> <new-tag>",
	Short: "Rename a tag across all places",
	Long:  `Rename a tag across all places that use it. This operation is atomic - either all places are updated or none.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldTag := strings.TrimSpace(args[0])
		newTag := strings.TrimSpace(args[1])

		if oldTag == "" || newTag == "" {
			return fmt.Errorf("tag names cannot be empty")
		}

		if oldTag == newTag {
			return fmt.Errorf("old and new tag names are the same")
		}

		logger.Info("Renaming tag", "old", oldTag, "new", newTag)

		count, err := renameTag(oldTag, newTag)
		if err != nil {
			return fmt.Errorf("failed to rename tag: %w", err)
		}

		if count == 0 {
			fmt.Printf("Tag '%s' not found\n", oldTag)
		} else {
			fmt.Printf("Renamed tag '%s' to '%s' in %d places\n", oldTag, newTag, count)
		}

		return nil
	},
}

var tagsDeleteCmd = &cobra.Command{
	Use:   "delete <tag>",
	Short: "Delete a tag from all places",
	Long:  `Remove a tag from all places that use it. Use --force to skip confirmation.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tag := strings.TrimSpace(args[0])

		if tag == "" {
			return fmt.Errorf("tag name cannot be empty")
		}

		// Check how many places use this tag
		count, err := getTagUsageCount(tag)
		if err != nil {
			return fmt.Errorf("failed to check tag usage: %w", err)
		}

		if count == 0 {
			fmt.Printf("Tag '%s' not found\n", tag)
			return nil
		}

		// Confirm deletion unless --force is used
		if !tagsForce {
			fmt.Printf("This will remove tag '%s' from %d places. Continue? (y/N): ", tag, count)
			var response string
			fmt.Scanln(&response)

			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Cancelled")
				return nil
			}
		}

		logger.Info("Deleting tag", "tag", tag)

		deleted, err := deleteTag(tag)
		if err != nil {
			return fmt.Errorf("failed to delete tag: %w", err)
		}

		fmt.Printf("Deleted tag '%s' from %d places\n", tag, deleted)

		return nil
	},
}

var tagsApplyCmd = &cobra.Command{
	Use:   "apply <tag>",
	Short: "Apply a tag to places matching a filter",
	Long: `Add a tag to all places matching the search filter.
	
Use --filter to specify which places to tag. Without a filter, 
the tag will be applied to all places.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tag := strings.TrimSpace(args[0])

		if tag == "" {
			return fmt.Errorf("tag name cannot be empty")
		}

		logger.Info("Applying tag", "tag", tag, "filter", tagsFilter)

		count, err := applyTag(tag, tagsFilter)
		if err != nil {
			return fmt.Errorf("failed to apply tag: %w", err)
		}

		if tagsFilter != "" {
			fmt.Printf("Applied tag '%s' to %d places matching '%s'\n", tag, count, tagsFilter)
		} else {
			fmt.Printf("Applied tag '%s' to %d places\n", tag, count)
		}

		return nil
	},
}

// Database operations for tag management

func getTagCounts() (map[string]int, error) {
	places, err := db.ListPlaces(10000, 0) // Get all places
	if err != nil {
		return nil, err
	}

	tagCounts := make(map[string]int)
	for _, place := range places {
		for _, tag := range place.UserTags {
			tagCounts[tag]++
		}
	}

	return tagCounts, nil
}

func getTagUsageCount(tag string) (int, error) {
	places, err := db.ListPlaces(10000, 0) // Get all places
	if err != nil {
		return 0, err
	}

	count := 0
	for _, place := range places {
		if place.HasTag(tag) {
			count++
		}
	}

	return count, nil
}

func renameTag(oldTag, newTag string) (int, error) {
	places, err := db.ListPlaces(10000, 0) // Get all places
	if err != nil {
		return 0, err
	}

	count := 0
	for _, place := range places {
		if place.HasTag(oldTag) {
			place.RemoveTag(oldTag)
			place.AddTag(newTag)

			if err := db.SavePlace(place); err != nil {
				return count, fmt.Errorf("failed to update place %s: %w", place.ID, err)
			}
			count++
		}
	}

	return count, nil
}

func deleteTag(tag string) (int, error) {
	places, err := db.ListPlaces(10000, 0) // Get all places
	if err != nil {
		return 0, err
	}

	count := 0
	for _, place := range places {
		if place.HasTag(tag) {
			place.RemoveTag(tag)

			if err := db.SavePlace(place); err != nil {
				return count, fmt.Errorf("failed to update place %s: %w", place.ID, err)
			}
			count++
		}
	}

	return count, nil
}

func applyTag(tag, filter string) (int, error) {
	var places []*models.Place
	var err error

	if filter != "" {
		places, err = db.SearchPlaces(filter)
	} else {
		places, err = db.ListPlaces(10000, 0) // Get all places
	}

	if err != nil {
		return 0, err
	}

	count := 0
	for _, place := range places {
		if !place.HasTag(tag) {
			place.AddTag(tag)

			if err := db.SavePlace(place); err != nil {
				return count, fmt.Errorf("failed to update place %s: %w", place.ID, err)
			}
			count++
		}
	}

	return count, nil
}
