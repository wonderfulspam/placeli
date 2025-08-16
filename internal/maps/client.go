package maps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/user/placeli/internal/models"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

type PlaceDetailsResponse struct {
	Result PlaceDetails `json:"result"`
	Status string       `json:"status"`
}

type PlaceDetails struct {
	PlaceID     string        `json:"place_id"`
	Name        string        `json:"name"`
	Rating      float32       `json:"rating"`
	UserRatings int           `json:"user_ratings_total"`
	PriceLevel  int           `json:"price_level"`
	Website     string        `json:"website"`
	Phone       string        `json:"formatted_phone_number"`
	Hours       *OpeningHours `json:"opening_hours"`
	Photos      []PlacePhoto  `json:"photos"`
	Reviews     []PlaceReview `json:"reviews"`
	Geometry    PlaceGeometry `json:"geometry"`
	Types       []string      `json:"types"`
	Address     string        `json:"formatted_address"`
}

type OpeningHours struct {
	OpenNow     bool     `json:"open_now"`
	WeekdayText []string `json:"weekday_text"`
}

type PlacePhoto struct {
	PhotoReference string `json:"photo_reference"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
}

type PlaceReview struct {
	AuthorName       string `json:"author_name"`
	AuthorURL        string `json:"author_url"`
	ProfilePhotoURL  string `json:"profile_photo_url"`
	Rating           int    `json:"rating"`
	RelativeTimeDesc string `json:"relative_time_description"`
	Text             string `json:"text"`
	Time             int64  `json:"time"`
}

type PlaceGeometry struct {
	Location PlaceLocation `json:"location"`
}

type PlaceLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://maps.googleapis.com/maps/api",
	}
}

func (c *Client) GetPlaceDetails(ctx context.Context, placeID string) (*PlaceDetails, error) {
	if placeID == "" {
		return nil, fmt.Errorf("place ID cannot be empty")
	}

	params := url.Values{
		"place_id": {placeID},
		"key":      {c.apiKey},
		"fields": {
			"place_id,name,rating,user_ratings_total,price_level,website," +
				"formatted_phone_number,opening_hours,photos,reviews," +
				"geometry,types,formatted_address",
		},
	}

	url := fmt.Sprintf("%s/place/details/json?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response PlaceDetailsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Status != "OK" {
		return nil, fmt.Errorf("API returned status: %s", response.Status)
	}

	return &response.Result, nil
}

func (c *Client) DownloadPhoto(ctx context.Context, photoReference string, maxWidth int) ([]byte, error) {
	if photoReference == "" {
		return nil, fmt.Errorf("photo reference cannot be empty")
	}

	params := url.Values{
		"photoreference": {photoReference},
		"key":            {c.apiKey},
		"maxwidth":       {fmt.Sprintf("%d", maxWidth)},
	}

	url := fmt.Sprintf("%s/place/photo?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("photo download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (d *PlaceDetails) ToPlace() *models.Place {
	place := &models.Place{
		PlaceID:     d.PlaceID,
		Name:        d.Name,
		Address:     d.Address,
		Rating:      d.Rating,
		UserRatings: d.UserRatings,
		PriceLevel:  d.PriceLevel,
		Website:     d.Website,
		Phone:       d.Phone,
		Categories:  d.Types,
		Coordinates: models.Coordinates{
			Lat: d.Geometry.Location.Lat,
			Lng: d.Geometry.Location.Lng,
		},
	}

	if d.Hours != nil && len(d.Hours.WeekdayText) > 0 {
		place.Hours = fmt.Sprintf("%v", d.Hours.WeekdayText)
	}

	for _, photo := range d.Photos {
		place.Photos = append(place.Photos, models.Photo{
			Reference: photo.PhotoReference,
			Width:     photo.Width,
			Height:    photo.Height,
		})
	}

	for _, review := range d.Reviews {
		place.Reviews = append(place.Reviews, models.Review{
			Author:       review.AuthorName,
			Rating:       review.Rating,
			Text:         review.Text,
			Time:         time.Unix(review.Time, 0),
			ProfilePhoto: review.ProfilePhotoURL,
		})
	}

	return place
}
