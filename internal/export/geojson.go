package export

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/user/placeli/internal/models"
)

type GeoJSONFeatureCollection struct {
	Type     string            `json:"type"`
	Features []*GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   *GeoJSONGeometry       `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func ExportGeoJSON(places []*models.Place, writer io.Writer) error {
	features := make([]*GeoJSONFeature, 0, len(places))

	for _, place := range places {
		feature := &GeoJSONFeature{
			Type: "Feature",
			Geometry: &GeoJSONGeometry{
				Type:        "Point",
				Coordinates: []float64{place.Coordinates.Lng, place.Coordinates.Lat},
			},
			Properties: map[string]interface{}{
				"id":           place.ID,
				"place_id":     place.PlaceID,
				"name":         place.Name,
				"address":      place.Address,
				"categories":   place.Categories,
				"rating":       place.Rating,
				"user_ratings": place.UserRatings,
				"price_level":  place.PriceLevel,
				"hours":        place.Hours,
				"phone":        place.Phone,
				"website":      place.Website,
				"user_notes":   place.UserNotes,
				"user_tags":    place.UserTags,
				"created_at":   place.CreatedAt.Format("2006-01-02T15:04:05Z"),
				"updated_at":   place.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			},
		}

		if len(place.Photos) > 0 {
			photos := make([]map[string]interface{}, len(place.Photos))
			for i, photo := range place.Photos {
				photos[i] = map[string]interface{}{
					"reference":  photo.Reference,
					"local_path": photo.LocalPath,
					"width":      photo.Width,
					"height":     photo.Height,
				}
			}
			feature.Properties["photos"] = photos
		}

		if len(place.Reviews) > 0 {
			reviews := make([]map[string]interface{}, len(place.Reviews))
			for i, review := range place.Reviews {
				reviews[i] = map[string]interface{}{
					"author":        review.Author,
					"rating":        review.Rating,
					"text":          review.Text,
					"time":          review.Time.Format("2006-01-02T15:04:05Z"),
					"profile_photo": review.ProfilePhoto,
				}
			}
			feature.Properties["reviews"] = reviews
		}

		if len(place.CustomFields) > 0 {
			feature.Properties["custom_fields"] = place.CustomFields
		}

		features = append(features, feature)
	}

	collection := &GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(collection); err != nil {
		return fmt.Errorf("failed to encode GeoJSON: %w", err)
	}

	return nil
}
