# placeli - A Terminal-Based Google Maps List Manager

## Project Overview

placeli is a terminal-based tool for managing your Google Maps lists. It provides a local, offline-first copy of your saved places, allowing you to interact with your data in a fast, keyboard-driven interface. It's designed for users who want to efficiently manage their saved places, add notes, and export their data to various formats.

## Core Concepts

* **Local First:** Your Google Maps data is synced to a local database, allowing for fast access and offline use.
* **Terminal UI:** A rich, interactive terminal interface provides a powerful and efficient way to manage your places.
* **Data Enrichment:** placeli can enrich your data with additional details from the Google Maps API, such as photos, reviews, and opening hours.
* **Extensible:** The tool is designed to be extensible, with a flexible data model and a command-line interface that can be easily scripted and integrated with other tools.

## Core Features

* **Import:** Import your saved places from Google Takeout or by providing a link to your Google Maps lists.
* **Interactive Review:** A TUI for browsing, searching, and filtering your saved places.
* **Detailed View:** View detailed information for each place, including photos, reviews, and opening hours.
* **Editing:** Add notes and custom tags to your places.
* **Export:** Export your data to common formats like GeoJSON, CSV, and Markdown.
* **Web UI:** A simple web interface for viewing your places on a map.

## Data Model

The core data model is centered around the `Place` object. The schema should be flexible and allow for storing additional metadata as needed.

```go
type Place struct {
    // Core Fields
    ID           string
    PlaceID      string
    Name         string
    Address      string
    Coordinates  struct {
        Lat float64
        Lng float64
    }
    Categories   []string

    // Rich Media
    Photos       []Photo
    Reviews      []Review

    // Details
    Rating       float32
    UserRatings  int
    PriceLevel   int
    Hours        string // Flexible format
    Phone        string
    Website      string

    // User Data
    UserNotes    string
    UserTags     []string
}

type Photo struct {
    Reference    string
    LocalPath    string
}

type Review struct {
    Author       string
    Rating       int
    Text         string
}
```

## Commands

The command-line interface should be simple and intuitive.

```bash
# Import from Google Takeout or Maps Lists
placeli import <source>

# Interactive review mode
placeli review

# Browse local database
placeli browse

# Export data
placeli export <format> <output>

# Web interface
placeli web
```

## TUI Design

The TUI should be clean, responsive, and keyboard-driven. It should provide a list view of places, a detailed view for a selected place, and a search/filter interface. The layout should be inspired by popular TUI applications like `lazygit` and `htop`. Image support in the terminal is a plus, but not a requirement for the initial version.

## Database Schema

A simple SQLite database is sufficient for local storage. The schema should be designed to be easily migrated and extended in the future.

```sql
CREATE TABLE places (
    id TEXT PRIMARY KEY,
    place_id TEXT UNIQUE,
    name TEXT NOT NULL,
    address TEXT,
    lat REAL,
    lng REAL,
    categories TEXT, -- JSON array
    data TEXT -- JSON blob for all other fields
);

CREATE TABLE user_data (
    place_id TEXT PRIMARY KEY REFERENCES places(id),
    notes TEXT,
    tags TEXT -- JSON array
);
```

## License

MIT