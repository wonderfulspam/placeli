package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/placeli/internal/logger"
	"github.com/user/placeli/internal/models"
)

var (
	fieldType     string
	fieldTemplate string
)

func init() {
	rootCmd.AddCommand(fieldsCmd)

	// Add subcommands
	fieldsCmd.AddCommand(fieldsListCmd)
	fieldsCmd.AddCommand(fieldsAddCmd)
	fieldsCmd.AddCommand(fieldsRemoveCmd)
	fieldsCmd.AddCommand(fieldsSetCmd)
	fieldsCmd.AddCommand(fieldsTemplatesCmd)

	// Flags for add command
	fieldsAddCmd.Flags().StringVar(&fieldType, "type", "text", "field type: text, number, date, boolean, list")
	fieldsTemplatesCmd.Flags().StringVar(&fieldTemplate, "template", "", "apply a field template: travel, business, personal")
}

var fieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "Manage custom fields for places",
	Long: `Manage custom fields to add your own metadata to places.

Custom fields allow you to store additional information beyond what's provided by Google Maps.

Available subcommands:
  list      - Show all custom fields in use
  add       - Add a custom field to a place
  remove    - Remove a custom field from a place
  set       - Set the value of a custom field
  templates - Apply predefined field templates

Field types:
  text     - Free text (default)
  number   - Numeric values
  date     - Date in YYYY-MM-DD format
  boolean  - true/false values
  list     - Comma-separated list of values

Examples:
  placeli fields list
  placeli fields add <place-id> "visit_date" --type=date
  placeli fields set <place-id> "visit_date" "2024-01-15"
  placeli fields templates --template=travel`,
}

var fieldsListCmd = &cobra.Command{
	Use:   "list [place-id]",
	Short: "List custom fields",
	Long:  `List all custom fields. If place-id is provided, show fields for that place only.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			// Show fields for specific place
			placeID := args[0]
			place, err := db.GetPlace(placeID)
			if err != nil {
				return fmt.Errorf("failed to get place: %w", err)
			}

			if len(place.CustomFields) == 0 {
				fmt.Printf("No custom fields found for place: %s\n", place.Name)
				return nil
			}

			fmt.Printf("Custom fields for %s:\n\n", place.Name)
			for key, value := range place.CustomFields {
				fmt.Printf("%-20s: %v\n", key, value)
			}
		} else {
			// Show all field names used across all places
			fieldNames, err := getAllFieldNames()
			if err != nil {
				return fmt.Errorf("failed to get field names: %w", err)
			}

			if len(fieldNames) == 0 {
				fmt.Println("No custom fields found")
				return nil
			}

			fmt.Printf("Found %d custom fields in use:\n\n", len(fieldNames))
			for field, count := range fieldNames {
				fmt.Printf("%-20s (used in %d places)\n", field, count)
			}
		}

		return nil
	},
}

var fieldsAddCmd = &cobra.Command{
	Use:   "add <place-id> <field-name>",
	Short: "Add a custom field to a place",
	Long:  `Add a custom field to a place with a default value based on the field type.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		placeID := args[0]
		fieldName := strings.TrimSpace(args[1])

		if fieldName == "" {
			return fmt.Errorf("field name cannot be empty")
		}

		place, err := db.GetPlace(placeID)
		if err != nil {
			return fmt.Errorf("failed to get place: %w", err)
		}

		if place.CustomFields == nil {
			place.CustomFields = make(map[string]interface{})
		}

		// Set default value based on type
		var defaultValue interface{}
		switch fieldType {
		case "text":
			defaultValue = ""
		case "number":
			defaultValue = 0
		case "date":
			defaultValue = time.Now().Format("2006-01-02")
		case "boolean":
			defaultValue = false
		case "list":
			defaultValue = []string{}
		default:
			return fmt.Errorf("unknown field type: %s", fieldType)
		}

		place.CustomFields[fieldName] = defaultValue

		if err := db.SavePlace(place); err != nil {
			return fmt.Errorf("failed to save place: %w", err)
		}

		logger.Info("Added custom field", "place", place.Name, "field", fieldName, "type", fieldType)
		fmt.Printf("Added field '%s' (%s) to %s\n", fieldName, fieldType, place.Name)

		return nil
	},
}

var fieldsRemoveCmd = &cobra.Command{
	Use:   "remove <place-id> <field-name>",
	Short: "Remove a custom field from a place",
	Long:  `Remove a custom field from a place.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		placeID := args[0]
		fieldName := strings.TrimSpace(args[1])

		place, err := db.GetPlace(placeID)
		if err != nil {
			return fmt.Errorf("failed to get place: %w", err)
		}

		if place.CustomFields == nil {
			fmt.Printf("No custom fields found for %s\n", place.Name)
			return nil
		}

		if _, exists := place.CustomFields[fieldName]; !exists {
			fmt.Printf("Field '%s' not found for %s\n", fieldName, place.Name)
			return nil
		}

		delete(place.CustomFields, fieldName)

		if err := db.SavePlace(place); err != nil {
			return fmt.Errorf("failed to save place: %w", err)
		}

		logger.Info("Removed custom field", "place", place.Name, "field", fieldName)
		fmt.Printf("Removed field '%s' from %s\n", fieldName, place.Name)

		return nil
	},
}

var fieldsSetCmd = &cobra.Command{
	Use:   "set <place-id> <field-name> <value>",
	Short: "Set the value of a custom field",
	Long:  `Set the value of a custom field. The value will be parsed according to the field type.`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		placeID := args[0]
		fieldName := strings.TrimSpace(args[1])
		valueStr := args[2]

		place, err := db.GetPlace(placeID)
		if err != nil {
			return fmt.Errorf("failed to get place: %w", err)
		}

		if place.CustomFields == nil {
			place.CustomFields = make(map[string]interface{})
		}

		// Parse value based on current type or infer type
		var value interface{}
		if existing, exists := place.CustomFields[fieldName]; exists {
			// Parse according to existing type
			value, err = parseFieldValue(valueStr, existing)
			if err != nil {
				return fmt.Errorf("failed to parse value: %w", err)
			}
		} else {
			// Try to infer type from value
			value = inferFieldValue(valueStr)
		}

		place.CustomFields[fieldName] = value

		if err := db.SavePlace(place); err != nil {
			return fmt.Errorf("failed to save place: %w", err)
		}

		logger.Info("Set custom field", "place", place.Name, "field", fieldName, "value", value)
		fmt.Printf("Set field '%s' = %v for %s\n", fieldName, value, place.Name)

		return nil
	},
}

var fieldsTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Apply predefined field templates",
	Long: `Apply predefined field templates to add common custom fields.

Available templates:
  travel   - Fields for travel planning (visit_date, rating_personal, notes_private)
  business - Fields for business tracking (last_visit, expense_category, client_rating)
  personal - Fields for personal use (favorite, last_meal, companion)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if fieldTemplate == "" {
			fmt.Println("Available templates:")
			fmt.Println("  travel   - Travel planning fields")
			fmt.Println("  business - Business tracking fields")
			fmt.Println("  personal - Personal use fields")
			fmt.Println("\nUse --template=<name> to apply a template")
			return nil
		}

		count, err := applyFieldTemplate(fieldTemplate)
		if err != nil {
			return fmt.Errorf("failed to apply template: %w", err)
		}

		fmt.Printf("Applied '%s' template to %d places\n", fieldTemplate, count)

		return nil
	},
}

// Helper functions for field management

func getAllFieldNames() (map[string]int, error) {
	fieldCounts := make(map[string]int)

	err := db.ForEachPlace("", func(place *models.Place) error {
		for fieldName := range place.CustomFields {
			// Skip system fields
			if !isSystemField(fieldName) {
				fieldCounts[fieldName]++
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return fieldCounts, nil
}

func isSystemField(fieldName string) bool {
	systemFields := []string{"google_maps_url", "imported_from", "import_date", "last_sync"}
	for _, sysField := range systemFields {
		if fieldName == sysField {
			return true
		}
	}
	return false
}

func parseFieldValue(valueStr string, existing interface{}) (interface{}, error) {
	switch existing.(type) {
	case string:
		return valueStr, nil
	case float64, int:
		return strconv.ParseFloat(valueStr, 64)
	case bool:
		return strconv.ParseBool(valueStr)
	case []interface{}:
		// Parse as comma-separated list
		items := strings.Split(valueStr, ",")
		var result []interface{}
		for _, item := range items {
			result = append(result, strings.TrimSpace(item))
		}
		return result, nil
	default:
		return valueStr, nil
	}
}

func inferFieldValue(valueStr string) interface{} {
	// Try to parse as number
	if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return val
	}

	// Try to parse as boolean
	if val, err := strconv.ParseBool(valueStr); err == nil {
		return val
	}

	// Try to parse as date
	if _, err := time.Parse("2006-01-02", valueStr); err == nil {
		return valueStr
	}

	// Default to string
	return valueStr
}

func applyFieldTemplate(template string) (int, error) {
	var fields map[string]interface{}

	switch template {
	case "travel":
		fields = map[string]interface{}{
			"visit_date":      "",
			"rating_personal": 0,
			"notes_private":   "",
			"planned_visit":   false,
		}
	case "business":
		fields = map[string]interface{}{
			"last_visit":       "",
			"expense_category": "",
			"client_rating":    0,
			"meeting_notes":    "",
		}
	case "personal":
		fields = map[string]interface{}{
			"favorite":    false,
			"last_meal":   "",
			"companion":   "",
			"mood_rating": 0,
		}
	default:
		return 0, fmt.Errorf("unknown template: %s", template)
	}

	count := 0

	err := db.ForEachPlace("", func(place *models.Place) error {
		if place.CustomFields == nil {
			place.CustomFields = make(map[string]interface{})
		}

		// Add template fields that don't already exist
		added := false
		for fieldName, defaultValue := range fields {
			if _, exists := place.CustomFields[fieldName]; !exists {
				place.CustomFields[fieldName] = defaultValue
				added = true
			}
		}

		if added {
			if err := db.SavePlace(place); err != nil {
				return fmt.Errorf("failed to update place %s: %w", place.ID, err)
			}
			count++
		}
		return nil
	})

	if err != nil {
		return count, err
	}

	return count, nil
}
