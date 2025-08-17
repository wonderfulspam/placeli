package database

import (
	"database/sql"

	"github.com/user/placeli/internal/models"
)

// FindPlaceBySourceHash finds a place by its source hash to detect duplicates
func (db *DB) FindPlaceBySourceHash(hash string) (*models.Place, error) {
	row := db.conn.QueryRow(`
		SELECT
			p.id, p.place_id, p.name, p.address, p.lat, p.lng,
			p.categories, p.data, p.created_at, p.updated_at,
			p.imported_at, p.source_hash,
			ud.notes, ud.tags, ud.custom_fields
		FROM places p
		LEFT JOIN user_data ud ON p.id = ud.place_id
		WHERE p.source_hash = ?`, hash)

	place, err := scanPlace(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return place, err
}

// FindDuplicateCandidates finds potential duplicate places based on coordinates or place_id
func (db *DB) FindDuplicateCandidates(place *models.Place) ([]*models.Place, error) {
	// Look for potential duplicates based on coordinates or place_id
	// Note: Zero coordinates (0,0) are treated as "no coordinates" and should not match each other
	rows, err := db.conn.Query(`
		SELECT
			p.id, p.place_id, p.name, p.address, p.lat, p.lng,
			p.categories, p.data, p.created_at, p.updated_at,
			p.imported_at, p.source_hash,
			ud.notes, ud.tags, ud.custom_fields
		FROM places p
		LEFT JOIN user_data ud ON p.id = ud.place_id
		WHERE (p.place_id = ? AND p.place_id != '')
		   OR (ABS(p.lat - ?) < 0.0001 AND ABS(p.lng - ?) < 0.0001
		       AND p.lat != 0 AND p.lng != 0 AND ? != 0 AND ? != 0)`,
		place.PlaceID, place.Coordinates.Lat, place.Coordinates.Lng,
		place.Coordinates.Lat, place.Coordinates.Lng)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []*models.Place
	for rows.Next() {
		candidate, err := scanPlace(rows)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

// GetPlacesWithoutSourceHash returns places that don't have a source hash (legacy imports)
func (db *DB) GetPlacesWithoutSourceHash() ([]*models.Place, error) {
	rows, err := db.conn.Query(`
		SELECT
			p.id, p.place_id, p.name, p.address, p.lat, p.lng,
			p.categories, p.data, p.created_at, p.updated_at,
			p.imported_at, p.source_hash,
			ud.notes, ud.tags, ud.custom_fields
		FROM places p
		LEFT JOIN user_data ud ON p.id = ud.place_id
		WHERE p.source_hash IS NULL OR p.source_hash = ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []*models.Place
	for rows.Next() {
		place, err := scanPlace(rows)
		if err != nil {
			return nil, err
		}
		places = append(places, place)
	}

	return places, nil
}
