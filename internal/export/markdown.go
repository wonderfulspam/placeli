package export

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/user/placeli/internal/models"
)

func ExportMarkdown(places []*models.Place, writer io.Writer) error {
	fmt.Fprintf(writer, "# Places Export\n\n")
	fmt.Fprintf(writer, "Generated on %s\n\n", time.Now().Format("January 2, 2006 at 3:04 PM"))
	fmt.Fprintf(writer, "Total places: %d\n\n", len(places))

	for i, place := range places {
		if i > 0 {
			fmt.Fprintf(writer, "---\n\n")
		}

		fmt.Fprintf(writer, "## %s\n\n", place.Name)

		if place.Address != "" {
			fmt.Fprintf(writer, "**Address:** %s\n\n", place.Address)
		}

		if place.Coordinates.Lat != 0 || place.Coordinates.Lng != 0 {
			fmt.Fprintf(writer, "**Coordinates:** %.6f, %.6f\n\n", place.Coordinates.Lat, place.Coordinates.Lng)
		}

		if len(place.Categories) > 0 {
			fmt.Fprintf(writer, "**Categories:** %s\n\n", strings.Join(place.Categories, ", "))
		}

		if place.Rating > 0 {
			stars := strings.Repeat("⭐", int(place.Rating))
			if place.Rating != float32(int(place.Rating)) {
				stars += "½"
			}
			fmt.Fprintf(writer, "**Rating:** %s %.1f", stars, place.Rating)
			if place.UserRatings > 0 {
				fmt.Fprintf(writer, " (%d reviews)", place.UserRatings)
			}
			fmt.Fprintf(writer, "\n\n")
		}

		if place.PriceLevel > 0 {
			priceSymbols := strings.Repeat("$", place.PriceLevel)
			fmt.Fprintf(writer, "**Price Level:** %s\n\n", priceSymbols)
		}

		if place.Hours != "" {
			fmt.Fprintf(writer, "**Hours:** %s\n\n", place.Hours)
		}

		if place.Phone != "" {
			fmt.Fprintf(writer, "**Phone:** %s\n\n", place.Phone)
		}

		if place.Website != "" {
			fmt.Fprintf(writer, "**Website:** [%s](%s)\n\n", place.Website, place.Website)
		}

		if place.UserNotes != "" {
			fmt.Fprintf(writer, "### Notes\n\n%s\n\n", place.UserNotes)
		}

		if len(place.UserTags) > 0 {
			fmt.Fprintf(writer, "**Tags:** ")
			for i, tag := range place.UserTags {
				if i > 0 {
					fmt.Fprintf(writer, ", ")
				}
				fmt.Fprintf(writer, "`%s`", tag)
			}
			fmt.Fprintf(writer, "\n\n")
		}

		if len(place.Photos) > 0 {
			fmt.Fprintf(writer, "### Photos\n\n")
			for _, photo := range place.Photos {
				if photo.LocalPath != "" {
					fmt.Fprintf(writer, "- ![Photo](%s)\n", photo.LocalPath)
				} else if photo.Reference != "" {
					fmt.Fprintf(writer, "- Photo Reference: `%s`\n", photo.Reference)
				}
			}
			fmt.Fprintf(writer, "\n")
		}

		if len(place.Reviews) > 0 {
			fmt.Fprintf(writer, "### Reviews\n\n")
			for _, review := range place.Reviews {
				stars := strings.Repeat("⭐", review.Rating)
				fmt.Fprintf(writer, "**%s** %s\n\n", review.Author, stars)
				if review.Text != "" {
					fmt.Fprintf(writer, "> %s\n\n", review.Text)
				}
				if !review.Time.IsZero() {
					fmt.Fprintf(writer, "*%s*\n\n", review.Time.Format("January 2, 2006"))
				}
			}
		}

		if len(place.CustomFields) > 0 {
			fmt.Fprintf(writer, "### Custom Fields\n\n")
			for key, value := range place.CustomFields {
				fmt.Fprintf(writer, "- **%s:** %v\n", key, value)
			}
			fmt.Fprintf(writer, "\n")
		}

		fmt.Fprintf(writer, "*Created: %s*\n", place.CreatedAt.Format("January 2, 2006"))
		fmt.Fprintf(writer, "*Updated: %s*\n\n", place.UpdatedAt.Format("January 2, 2006"))
	}

	return nil
}
