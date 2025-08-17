package database

import (
	"github.com/user/placeli/internal/constants"
	"github.com/user/placeli/internal/models"
)

// ForEachPlace applies a function to all places, with optional filtering
func (db *DB) ForEachPlace(filter string, fn func(*models.Place) error) error {
	var places []*models.Place
	var err error

	if filter != "" {
		places, err = db.SearchPlaces(filter)
	} else {
		places, err = db.ListPlaces(constants.DefaultPlaceLimit, 0)
	}

	if err != nil {
		return err
	}

	for _, place := range places {
		if err := fn(place); err != nil {
			return err
		}
	}

	return nil
}

// CountPlaces returns the total number of places in the database
func (db *DB) CountPlaces() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM places").Scan(&count)
	return count, err
}
