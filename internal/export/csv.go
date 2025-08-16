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

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record for place %s: %w", place.ID, err)
		}
	}

	return nil
}
