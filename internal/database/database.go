package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/user/placeli/internal/models"
	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS places (
		id TEXT PRIMARY KEY,
		place_id TEXT,
		name TEXT NOT NULL,
		address TEXT,
		lat REAL,
		lng REAL,
		categories TEXT,
		data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		imported_at DATETIME,
		source_hash TEXT
	);

	CREATE TABLE IF NOT EXISTS user_data (
		place_id TEXT PRIMARY KEY,
		notes TEXT,
		tags TEXT,
		custom_fields TEXT,
		FOREIGN KEY (place_id) REFERENCES places(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_places_name ON places(name);
	CREATE INDEX IF NOT EXISTS idx_places_coordinates ON places(lat, lng);
	CREATE INDEX IF NOT EXISTS idx_places_place_id ON places(place_id);
	CREATE INDEX IF NOT EXISTS idx_user_data_tags ON user_data(tags);
	CREATE INDEX IF NOT EXISTS idx_places_source_hash ON places(source_hash);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return err
	}

	// Migration: Add new columns if they don't exist
	migrations := []string{
		"ALTER TABLE places ADD COLUMN imported_at DATETIME",
		"ALTER TABLE places ADD COLUMN source_hash TEXT",
	}

	for _, migration := range migrations {
		// Ignore errors if columns already exist
		_, _ = db.conn.Exec(migration)
	}

	return nil
}

func (db *DB) SavePlace(place *models.Place) error {
	now := time.Now()
	if place.CreatedAt.IsZero() {
		place.CreatedAt = now
	}
	place.UpdatedAt = now

	categoriesJSON, _ := json.Marshal(place.Categories)

	data := placeData{
		Photos:      place.Photos,
		Reviews:     place.Reviews,
		Rating:      place.Rating,
		UserRatings: place.UserRatings,
		PriceLevel:  place.PriceLevel,
		Hours:       place.Hours,
		Phone:       place.Phone,
		Website:     place.Website,
	}
	dataJSON, _ := json.Marshal(data)

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO places
		(id, place_id, name, address, lat, lng, categories, data, created_at, updated_at, imported_at, source_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		place.ID, place.PlaceID, place.Name, place.Address,
		place.Coordinates.Lat, place.Coordinates.Lng,
		string(categoriesJSON), string(dataJSON),
		place.CreatedAt, place.UpdatedAt, place.ImportedAt, place.SourceHash)
	if err != nil {
		return err
	}

	tagsJSON, _ := json.Marshal(place.UserTags)
	customFieldsJSON, _ := json.Marshal(place.CustomFields)

	_, err = tx.Exec(`
		INSERT OR REPLACE INTO user_data (place_id, notes, tags, custom_fields)
		VALUES (?, ?, ?, ?)`,
		place.ID, place.UserNotes, string(tagsJSON), string(customFieldsJSON))
	if err != nil {
		return err
	}

	return tx.Commit()
}

type placeData struct {
	Photos      []models.Photo  `json:"photos"`
	Reviews     []models.Review `json:"reviews"`
	Rating      float32         `json:"rating"`
	UserRatings int             `json:"user_ratings"`
	PriceLevel  int             `json:"price_level"`
	Hours       string          `json:"hours"`
	Phone       string          `json:"phone"`
	Website     string          `json:"website"`
}

func scanPlace(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Place, error) {
	var place models.Place
	var categoriesJSON, dataJSON string
	var tagsJSON, customFieldsJSON sql.NullString
	var importedAt sql.NullTime
	var sourceHash sql.NullString

	err := scanner.Scan(
		&place.ID, &place.PlaceID, &place.Name, &place.Address,
		&place.Coordinates.Lat, &place.Coordinates.Lng,
		&categoriesJSON, &dataJSON, &place.CreatedAt, &place.UpdatedAt,
		&importedAt, &sourceHash,
		&place.UserNotes, &tagsJSON, &customFieldsJSON)
	if err != nil {
		return nil, err
	}

	if importedAt.Valid {
		place.ImportedAt = &importedAt.Time
	}
	if sourceHash.Valid {
		place.SourceHash = sourceHash.String
	}

	return unmarshalPlace(&place, categoriesJSON, dataJSON, tagsJSON, customFieldsJSON)
}

func unmarshalPlace(place *models.Place, categoriesJSON, dataJSON string, tagsJSON, customFieldsJSON sql.NullString) (*models.Place, error) {
	if err := json.Unmarshal([]byte(categoriesJSON), &place.Categories); err != nil {
		place.Categories = []string{}
	}

	var data placeData
	if err := json.Unmarshal([]byte(dataJSON), &data); err == nil {
		place.Photos = data.Photos
		place.Reviews = data.Reviews
		place.Rating = data.Rating
		place.UserRatings = data.UserRatings
		place.PriceLevel = data.PriceLevel
		place.Hours = data.Hours
		place.Phone = data.Phone
		place.Website = data.Website
	}

	if tagsJSON.Valid {
		if err := json.Unmarshal([]byte(tagsJSON.String), &place.UserTags); err != nil {
			place.UserTags = []string{}
		}
	}
	if customFieldsJSON.Valid {
		if err := json.Unmarshal([]byte(customFieldsJSON.String), &place.CustomFields); err != nil {
			place.CustomFields = make(map[string]interface{})
		}
	}

	return place, nil
}

func (db *DB) GetPlace(id string) (*models.Place, error) {
	row := db.conn.QueryRow(`
		SELECT
			p.id, p.place_id, p.name, p.address, p.lat, p.lng,
			p.categories, p.data, p.created_at, p.updated_at,
			p.imported_at, p.source_hash,
			ud.notes, ud.tags, ud.custom_fields
		FROM places p
		LEFT JOIN user_data ud ON p.id = ud.place_id
		WHERE p.id = ?`, id)

	return scanPlace(row)
}

func (db *DB) ListPlaces(limit, offset int) ([]*models.Place, error) {
	rows, err := db.conn.Query(`
		SELECT
			p.id, p.place_id, p.name, p.address, p.lat, p.lng,
			p.categories, p.data, p.created_at, p.updated_at,
			p.imported_at, p.source_hash,
			ud.notes, ud.tags, ud.custom_fields
		FROM places p
		LEFT JOIN user_data ud ON p.id = ud.place_id
		ORDER BY p.updated_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
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

func (db *DB) DeletePlace(id string) error {
	_, err := db.conn.Exec("DELETE FROM places WHERE id = ?", id)
	return err
}

func (db *DB) SearchPlaces(query string) ([]*models.Place, error) {
	rows, err := db.conn.Query(`
		SELECT
			p.id, p.place_id, p.name, p.address, p.lat, p.lng,
			p.categories, p.data, p.created_at, p.updated_at,
			p.imported_at, p.source_hash,
			ud.notes, ud.tags, ud.custom_fields
		FROM places p
		LEFT JOIN user_data ud ON p.id = ud.place_id
		WHERE p.name LIKE ? OR p.address LIKE ? OR ud.notes LIKE ?
		ORDER BY p.updated_at DESC`, "%"+query+"%", "%"+query+"%", "%"+query+"%")
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
