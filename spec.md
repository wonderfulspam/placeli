# placeli - A Terminal-Based Google Maps List Manager

## Project Overview

placeli is a terminal-based tool for managing your Google Maps lists. It
provides a local, offline-first copy of your saved places, allowing you to
interact with your data in a fast, keyboard-driven interface. It's designed for
users who want to efficiently manage their saved places, add notes, and export
their data to various formats.

## Core Concepts

* **Local First:** Your Google Maps data is synced to a local database, allowing
  for fast access and offline use.
* **Terminal UI:** A rich, interactive terminal interface provides a powerful
  and efficient way to manage your places.
* **Data Enrichment:** placeli can enrich your data with additional details from
  the Google Maps API, such as photos, reviews, and opening hours.
* **Extensible:** The tool is designed to be extensible, with a flexible data
  model and a command-line interface that can be easily scripted and integrated
  with other tools.

## Core Features

* **Import:** Import your saved places from Google Takeout.
* **Synchronization:** Periodically re-import from Google Takeout to keep the
  local database up-to-date. The tool should be able to merge changes
  intelligently, preserving user-added data.
* **Interactive Review:** A TUI for browsing, searching, and filtering your
  saved places. Supports advanced filtering by tags, categories, rating, and
  proximity.
* **Detailed View:** View detailed information for each place, including photos,
  reviews, and opening hours.
* **Editing:** Add notes, custom tags, and user-defined custom fields (e.g.,
  priority, visited_date) to your places.
* **Export:** Export your data to common formats like GeoJSON, CSV, and
  Markdown.
* **Web UI:** A simple web interface for viewing your places on a map.
* **Terminal Map View:** (Optional) Render a simple map in the terminal using
  ASCII or other character-based representations.

## Data Model

The core data model is centered around the `Place` object. The schema should be
flexible and allow for storing additional metadata as needed.

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
    CustomFields map[string]interface{}
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
# Import from Google Takeout
placeli import <source>

# Sync with Google Takeout
placeli sync <source>

# Interactive review mode
placeli review

# Query the local database
placeli query --tag "favorite" --format json

# Export data
placeli export <format> <output>

# Web interface
placeli web
```

## TUI Design

The TUI should be clean, responsive, and keyboard-driven. It should provide a
list view of places, a detailed view for a selected place, and a search/filter
interface. The layout should be inspired by popular TUI applications like
`lazygit` and `htop`. Image support in the terminal is a plus, but not a
requirement for the initial version.

```terminal
┌──────────────────────────────────────────────────────────────────────────────┐
│ placeli review                                                  [1/234] 0.4% │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Joe's Pizza 🍕                                            ⭐ 4.2 (523)      │
│  Italian Restaurant • Pizza • $ • Brooklyn                                   │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────────┐  │
│  │                                                                        │  │
│  │                    [Terminal Image Display]                            │  │
│  │                                                                        │  │
│  │                  A vibrant photo of a pizza slice                      │  │
│  │              with pepperoni and melted cheese.                         │  │
│  │                                                                        │  │
│  └────────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  📍 123 Main St, Brooklyn, NY 11201                                          │
│  📞 (718) 555-0123                                                           │
│  🌐 joespizzabrooklyn.com                                                    │
│  🕒 Mon-Thu 11AM-10PM, Fri-Sat 11AM-11PM, Sun 12PM-10PM                      │
│                                                                              │
│  Recent Reviews ───────────────────────────────────────────────────────────  │
│                                                                              │
│  ⭐⭐⭐⭐⭐ Sarah M. • 2 days ago                                            │
│  "Absolutely the best pizza in Brooklyn! The crust is perfect"               │
│                                                                              │
│  ⭐⭐⭐⭐ John D. • 1 week ago                                               │
│  "Good pizza but a bit overpriced. Service was excellent though"             │
│                                                                              │
├──────────────────────────────────────────────────────────────────────────────┤
│ [d]elete [k]eep [t]ag [n]ote [e]dit [w]eb [s]kip [q]uit | [/]search [f]ilter │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Database Schema

A simple SQLite database is sufficient for local storage. The schema should be
designed to be easily migrated and extended in the future.

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
