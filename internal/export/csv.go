package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/user/placeli/internal/models"
)

func ExportCSV(places []*models.Place, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Collect all custom field names
	customFields := getAllCustomFieldNames(places)

	headers := []string{
		"ID",
		"PlaceID",
		"Name",
		"Address",
		"Latitude",
		"Longitude",
		"Categories",
		"Rating",
		"UserRatings",
		"PriceLevel",
		"Hours",
		"Phone",
		"Website",
		"UserNotes",
		"UserTags",
		"CreatedAt",
		"UpdatedAt",
	}

	// Add custom field headers
	for _, fieldName := range customFields {
		headers = append(headers, "custom_"+fieldName)
	}

	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	for _, place := range places {
		record := []string{
			place.ID,
			place.PlaceID,
			place.Name,
			place.Address,
			fmt.Sprintf("%.6f", place.Coordinates.Lat),
			fmt.Sprintf("%.6f", place.Coordinates.Lng),
			strings.Join(place.Categories, "; "),
			fmt.Sprintf("%.1f", place.Rating),
			strconv.Itoa(place.UserRatings),
			strconv.Itoa(place.PriceLevel),
			place.Hours,
			place.Phone,
			place.Website,
			place.UserNotes,
			strings.Join(place.UserTags, "; "),
			place.CreatedAt.Format("2006-01-02 15:04:05"),
			place.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		// Add custom field values
		for _, fieldName := range customFields {
			value := ""
			if place.CustomFields != nil {
				if val, exists := place.CustomFields[fieldName]; exists && val != nil {
					value = formatCustomFieldValue(val)
				}
			}
			record = append(record, value)
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record for place %s: %w", place.ID, err)
		}
	}

	return nil
}

func getAllCustomFieldNames(places []*models.Place) []string {
	fieldNames := make(map[string]bool)
	systemFields := map[string]bool{
		"google_maps_url": true,
		"imported_from":   true,
		"import_date":     true,
		"last_sync":       true,
	}

	for _, place := range places {
		for fieldName := range place.CustomFields {
			if !systemFields[fieldName] {
				fieldNames[fieldName] = true
			}
		}
	}

	// Convert to sorted slice
	var result []string
	for fieldName := range fieldNames {
		result = append(result, fieldName)
	}

	return result
}

func formatCustomFieldValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.1f", v)
	case bool:
		return strconv.FormatBool(v)
	case []interface{}:
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf("%v", item))
		}
		return strings.Join(items, "; ")
	default:
		if value != nil {
			return fmt.Sprintf("%v", value)
		}
		return ""
	}
}
