# maps-list-manager (mlm) - Implementation Spec

## Project Overview

A rich terminal-based Google Maps list manager that creates a local clone of your Google Maps data with full media support, enabling git-diff-style interactive review with images, reviews, and detailed place information.

## Core Architecture

### Language: Go

- Efficient concurrent API fetching
- Single binary distribution
- Excellent terminal UI libraries (Charm stack)
- Good Google Maps API support

### Key Dependencies

```go
// go.mod
module github.com/user/maps-list-manager

require (
    github.com/spf13/cobra v1.8.0              // CLI framework
    github.com/charmbracelet/bubbletea v0.25.0 // TUI framework
    github.com/charmbracelet/lipgloss v0.9.1   // Terminal styling
    github.com/charmbracelet/bubblezone v0.7.0 // Mouse support
    googlemaps.github.io/maps v1.5.0           // Google Maps client
    github.com/paulmach/orb v0.10.0            // Geo operations
    github.com/mattn/go-sqlite3 v1.14.18       // Local database
    github.com/sahilm/fuzzy v0.1.0             // Fuzzy matching
    github.com/golang/freetype v0.0.0          // Image rendering
    github.com/nfnt/resize v0.0.0              // Image resizing
    golang.org/x/time v0.5.0                   // Rate limiting
)
```

## Data Models

```go
type Place struct {
    // Core Fields
    ID           string          `json:"id" db:"id"`
    PlaceID      string          `json:"place_id" db:"place_id"`
    Name         string          `json:"name" db:"name"`
    Address      string          `json:"address" db:"address"`
    Coordinates  LatLng          `json:"coordinates" db:"coordinates"`
    Categories   []Category      `json:"categories" db:"categories"`
    
    // Rich Media
    Photos       []Photo         `json:"photos" db:"photos"`
    Reviews      []Review        `json:"reviews" db:"reviews"`
    
    // Details
    Rating       float32         `json:"rating" db:"rating"`
    UserRatings  int             `json:"user_ratings_total" db:"user_ratings"`
    PriceLevel   int             `json:"price_level" db:"price_level"`
    Hours        OpeningHours    `json:"opening_hours" db:"hours"`
    Phone        string          `json:"phone" db:"phone"`
    Website      string          `json:"website" db:"website"`
    
    // Extended Data
    PopularTimes []PopularTime   `json:"popular_times" db:"popular_times"`
    MenuURL      string          `json:"menu_url" db:"menu_url"`
    Attributes   map[string]bool `json:"attributes" db:"attributes"`
    StreetViewID string          `json:"street_view_id" db:"street_view_id"`
    WikipediaURL string          `json:"wikipedia_url" db:"wikipedia_url"`
    
    // Metadata
    Status       BusinessStatus  `json:"status" db:"status"`
    LastUpdated  time.Time       `json:"last_updated" db:"last_updated"`
    LastVisited  *time.Time      `json:"last_visited" db:"last_visited"`
    UserNotes    string          `json:"user_notes" db:"user_notes"`
    UserPhotos   []string        `json:"user_photos" db:"user_photos"`
    ListNames    []string        `json:"list_names" db:"list_names"`
}

type Photo struct {
    Reference    string    `json:"reference"`
    Width        int       `json:"width"`
    Height       int       `json:"height"`
    Attribution  string    `json:"attribution"`
    LocalPath    string    `json:"local_path"`
    ThumbnailPath string   `json:"thumbnail_path"`
}

type Review struct {
    Author       string    `json:"author_name"`
    Rating       int       `json:"rating"`
    Text         string    `json:"text"`
    Time         time.Time `json:"time"`
    RelativeTime string    `json:"relative_time_description"`
    Photos       []Photo   `json:"photos"`
    OwnerReply   string    `json:"owner_reply"`
}

type BusinessStatus int
const (
    StatusOperational BusinessStatus = iota
    StatusClosedTemporarily
    StatusClosedPermanently
    StatusUnknown
)

type Category string
const (
    Restaurant  Category = "restaurant"
    Bar         Category = "bar"
    Cafe        Category = "cafe"
    Shop        Category = "shop"
    Grocery     Category = "grocery"
    Hotel       Category = "hotel"
    Attraction  Category = "attraction"
    Museum      Category = "museum"
    Park        Category = "park"
    Service     Category = "service"
    Transport   Category = "transport"
    Uncategorized Category = "uncategorized"
)
```

## File Structure

```
maps-list-manager/
├── cmd/
│   └── mlm/
│       └── main.go                 # Entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                # Root command
│   │   ├── import.go              # Import command
│   │   ├── review.go              # Review command
│   │   ├── export.go              # Export command
│   │   └── web.go                 # Web UI command
│   ├── tui/
│   │   ├── model.go               # Bubbletea model
│   │   ├── view.go                # Render functions
│   │   ├── update.go              # Event handlers
│   │   ├── components/
│   │   │   ├── place_view.go      # Full place display
│   │   │   ├── image_gallery.go   # Image carousel
│   │   │   ├── review_list.go     # Reviews component
│   │   │   └── comparison.go      # Side-by-side view
│   │   └── styles/
│   │       └── theme.go           # Lipgloss styles
│   ├── google/
│   │   ├── client.go              # API client wrapper
│   │   ├── places.go              # Places API
│   │   ├── photos.go              # Photo downloading
│   │   └── ratelimit.go           # Rate limiting
│   ├── storage/
│   │   ├── database.go            # SQLite interface
│   │   ├── cache.go               # Caching layer
│   │   ├── migrations.go          # Schema migrations
│   │   └── queries.go             # SQL queries
│   ├── analyzer/
│   │   ├── duplicates.go          # Duplicate detection
│   │   ├── categorizer.go         # Auto-categorization
│   │   ├── staleness.go           # Staleness scoring
│   │   └── closure.go             # Closure detection
│   ├── export/
│   │   ├── geojson.go             # GeoJSON export
│   │   ├── osm.go                 # OpenStreetMap export
│   │   ├── csv.go                 # CSV export
│   │   ├── html.go                # Static site generator
│   │   └── markdown.go            # Markdown export
│   ├── terminal/
│   │   ├── images.go              # Image protocol detection
│   │   ├── sixel.go               # Sixel rendering
│   │   ├── kitty.go               # Kitty protocol
│   │   ├── iterm2.go              # iTerm2 protocol
│   │   └── ascii.go               # ASCII fallback
│   └── web/
│       ├── server.go              # Web server
│       ├── handlers.go            # HTTP handlers
│       └── static/
│           ├── index.html         # Web UI
│           ├── map.js             # Mapbox integration
│           └── style.css          # Styles
├── configs/
│   └── default.yaml               # Default configuration
├── scripts/
│   ├── install.sh                 # Installation script
│   └── test_images.sh             # Terminal capability test
└── README.md
```

## Commands & Usage

### Primary Commands

```bash
# Import from Google Takeout or Maps Lists
mlm import <source>
  --fetch-all              # Fetch all available data (default: true)
  --photos                 # Download photos (default: true)
  --reviews                # Fetch reviews (default: true)
  --concurrent-workers=10  # Number of parallel fetchers

# Interactive review mode
mlm review
  --images                 # Show images in terminal (auto-detected)
  --auto-categorize        # Run categorization before review
  --filter="status:closed" # Pre-filter places
  --sort="rating:desc"     # Sort order

# Browse local database
mlm browse
  --search="pizza"         # Full-text search
  --category="restaurant"  # Filter by category
  --map                    # Open web map view

# Export data
mlm export <format> <output>
  formats: geojson, osm, csv, html, markdown
  --include-photos         # Include photo references
  --embed-images          # Embed images in HTML export

# Web interface
mlm web
  --port=8080             # Web server port
  --open                  # Auto-open browser

# Database management
mlm stats                 # Show database statistics
mlm cache clean          # Clean old cache entries
mlm cache rebuild        # Rebuild from API
```

### Configuration File

```yaml
# ~/.config/mlm/config.yaml
google:
  api_key: ${GOOGLE_MAPS_API_KEY}
  max_photos_per_place: 5
  max_reviews_per_place: 10
  rate_limit:
    requests_per_second: 10
    burst: 20

terminal:
  image_protocol: auto  # auto, sixel, kitty, iterm2, ascii, none
  image_width: 400
  thumbnail_size: 150
  
categories:
  enabled:
    - restaurant
    - bar
    - cafe
    - shop
    - grocery
    - hotel
    - attraction
    - museum
    - park
    - service
    - transport
  
  mapping:  # Map Google categories to our categories
    food: restaurant
    meal_takeaway: restaurant
    night_club: bar
    shopping_mall: shop

analyzer:
  duplicate_distance_meters: 50
  duplicate_name_similarity: 0.85
  stale_days: 730
  auto_delete_closed: false
  category_confidence_threshold: 0.9

storage:
  cache_dir: ~/.mlm
  database_file: places.db
  image_cache_days: 30
  api_cache_days: 7

export:
  include_closed: false
  include_uncategorized: true
  html_template: default

web:
  mapbox_token: ${MAPBOX_TOKEN}
  default_zoom: 13
  cluster_markers: true
```

## TUI Interface Design

### Main Review Screen

```
┌─────────────────────────────────────────────────────────────────────┐
│ mlm review                                           [1/234] 0.4%   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  Joe's Pizza 🍕                                    ⭐ 4.2 (523)    │
│  Italian Restaurant • Pizza • $$ • Brooklyn                         │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────┐    │
│  │                                                              │    │
│  │                    [Terminal Image Display]                  │    │
│  │                         3 photos available                   │    │
│  │                       Use ← → to browse                      │    │
│  │                                                              │    │
│  └────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  📍 123 Main St, Brooklyn, NY 11201                                 │
│  📞 (718) 555-0123                                                  │
│  🌐 joespizzabrooklyn.com                                          │
│  🕒 Mon-Thu 11AM-10PM, Fri-Sat 11AM-11PM, Sun 12PM-10PM           │
│     Currently: CLOSED (Opens at 11 AM)                             │
│                                                                      │
│  📊 Popular Times:                                                  │
│     12PM ▂▃▄▅▆▇█▇▆▅▄▃▂ 11PM                                      │
│           ↑ Usually busy                                           │
│                                                                      │
│  Recent Reviews ────────────────────────────────────────────       │
│                                                                      │
│  ⭐⭐⭐⭐⭐ Sarah M. • 2 days ago                                    │
│  "Absolutely the best pizza in Brooklyn! The crust is perfect"     │
│                                                                      │
│  ⭐⭐⭐⭐ John D. • 1 week ago                                       │
│  "Good pizza but a bit overpriced. Service was excellent though"   │
│  ↳ Owner response: "Thanks for the feedback John!"                 │
│                                                                      │
│  Issues Detected ───────────────────────────────────────────       │
│                                                                      │
│  ⚠️  Permanently closed as of 2024-03-15                           │
│  🔄 Potential duplicate: "Joe's Pizza & Pasta" (500m away)         │
│  📝 No category assigned (suggested: Restaurant)                    │
│                                                                      │
├─────────────────────────────────────────────────────────────────────┤
│ [d]elete [k]eep [c]ategory [m]erge [n]ote [e]dit [w]eb [s]kip [q]uit│
│ [/]search [f]ilter [?]help                         [u]ndo [r]edo    │
└─────────────────────────────────────────────────────────────────────┘
```

### Comparison View (for merges)

```
┌─────────────────────────────────────────────────────────────────────┐
│ Duplicate Detection - Choose primary                                │
├──────────────────────────────────┬──────────────────────────────────┤
│        Joe's Pizza               │      Joe's Pizza & Pasta         │
├──────────────────────────────────┼──────────────────────────────────┤
│ ⭐ 4.2 (523 reviews)             │ ⭐ 4.5 (892 reviews)            │
│ 💰 $$                            │ 💰 $$$                          │
│ 📍 123 Main St                   │ 📍 456 Court St                 │
│ 📏 Distance: 500m                │                                  │
│                                  │                                  │
│ ⚠️  CLOSED (March 2024)          │ ✅ Currently OPEN                │
│ Last visited: Jan 2024           │ Last visited: Never              │
│ In lists: Favorites, Brooklyn    │ In lists: Want to Try           │
│                                  │                                  │
│ Photos: 12                       │ Photos: 34                       │
│ Menu: Not available              │ Menu: Available                  │
│ Delivery: No                     │ Delivery: Yes (Uber, DoorDash)  │
├──────────────────────────────────┴──────────────────────────────────┤
│ [1] Keep left  [2] Keep right  [m] Merge both  [s] Skip  [?] Help  │
└─────────────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: Core Import & Storage (Week 1)

- [ ] Project setup with Go modules
- [ ] Google Takeout parser
- [ ] Google Maps Lists importer
- [ ] SQLite schema and migrations
- [ ] Basic Place model and storage

### Phase 2: API Integration (Week 1-2)

- [ ] Google Maps/Places API client
- [ ] Rate limiting and quota management
- [ ] Photo downloader with local caching
- [ ] Review fetcher
- [ ] Popular times and extended attributes

### Phase 3: TUI Development (Week 2-3)

- [ ] Basic Bubbletea app structure
- [ ] Place detail view component
- [ ] Image display with protocol detection
- [ ] Review list component
- [ ] Keyboard navigation

### Phase 4: Analysis & Intelligence (Week 3)

- [ ] Duplicate detection algorithm
- [ ] Auto-categorization
- [ ] Closure detection
- [ ] Staleness scoring
- [ ] Fuzzy search implementation

### Phase 5: Export & Web UI (Week 4)

- [ ] GeoJSON exporter
- [ ] CSV exporter
- [ ] HTML static site generator
- [ ] Web server with API
- [ ] Mapbox integration

### Phase 6: Polish & Testing (Week 4)

- [ ] Comprehensive error handling
- [ ] Progress bars for long operations
- [ ] Undo/redo functionality
- [ ] Configuration management
- [ ] Installation script

## API Usage & Quotas

### Required Google APIs

1. **Places API (New)**
   - Place Details: $17/1000 requests
   - Nearby Search: $32/1000 requests
   - Photos: $7/1000 requests

2. **Maps Static API** (optional)
   - For terminal map previews: $2/1000 requests

3. **Street View API** (optional)
   - Metadata: $7/1000 requests

### Optimization Strategies

- Cache all API responses for 7-30 days
- Batch requests where possible
- Progressive enhancement (basic data first)
- User-configurable fetch levels
- Implement exponential backoff

## Terminal Image Protocols

### Detection Order

1. Check `$TERM` and `$TERM_PROGRAM`
2. Query terminal capabilities
3. Test protocols in order: Sixel → Kitty → iTerm2
4. Fall back to Unicode blocks or ASCII art

### Implementation

```go
func DetectImageSupport() ImageProtocol {
    if os.Getenv("TERM_PROGRAM") == "iTerm.app" {
        return ITerm2
    }
    if os.Getenv("TERM") == "xterm-kitty" {
        return Kitty
    }
    if CheckSixelSupport() {
        return Sixel
    }
    return ASCII
}
```

## Database Schema

```sql
CREATE TABLE places (
    id TEXT PRIMARY KEY,
    place_id TEXT UNIQUE,
    name TEXT NOT NULL,
    address TEXT,
    lat REAL,
    lng REAL,
    categories TEXT, -- JSON array
    rating REAL,
    user_ratings INTEGER,
    price_level INTEGER,
    phone TEXT,
    website TEXT,
    hours TEXT, -- JSON
    status INTEGER,
    last_updated TIMESTAMP,
    data TEXT -- JSON for all fields
);

CREATE TABLE photos (
    id TEXT PRIMARY KEY,
    place_id TEXT REFERENCES places(id),
    reference TEXT,
    width INTEGER,
    height INTEGER,
    local_path TEXT,
    thumbnail_path TEXT
);

CREATE TABLE reviews (
    id TEXT PRIMARY KEY,
    place_id TEXT REFERENCES places(id),
    author TEXT,
    rating INTEGER,
    text TEXT,
    time TIMESTAMP,
    data TEXT -- JSON
);

CREATE TABLE user_data (
    place_id TEXT PRIMARY KEY REFERENCES places(id),
    notes TEXT,
    last_visited TIMESTAMP,
    lists TEXT, -- JSON array
    custom_category TEXT
);

-- Full-text search
CREATE VIRTUAL TABLE places_fts USING fts5(
    name, address, categories, content=places
);
```

## Performance Targets

- Import 1000 places: < 2 minutes (with API fetching)
- Load place details in TUI: < 100ms
- Image rendering: < 200ms
- Search 10,000 places: < 50ms
- Export 1000 places to GeoJSON: < 1 second

## Error Handling

- All API errors should be retried with exponential backoff
- Network failures should not lose progress (checkpoint regularly)
- Malformed data should be logged but not stop processing
- Terminal capability detection failures should gracefully degrade
- Database corruption should trigger automatic backup restore

## Testing Strategy

- Unit tests for all analyzers and exporters
- Integration tests for Google API client
- Terminal emulator tests for image protocols
- Benchmark tests for large datasets
- Manual testing on: macOS (iTerm2), Linux (Kitty), Windows (Windows Terminal)

## Distribution

```bash
# Build for all platforms
make build-all

# Outputs
dist/
├── mlm-darwin-amd64
├── mlm-darwin-arm64
├── mlm-linux-amd64
├── mlm-linux-arm64
└── mlm-windows-amd64.exe

# Installation
curl -L https://github.com/user/mlm/releases/latest/download/mlm-$(uname -s)-$(uname -m) -o mlm
chmod +x mlm
sudo mv mlm /usr/local/bin/
```

## License
MIT

## Resources

- [Google Places API Documentation](https://developers.google.com/maps/documentation/places/web-service)
- [Bubbletea TUI Framework](https://github.com/charmbracelet/bubbletea)
- [Sixel Graphics Format](https://en.wikipedia.org/wiki/Sixel)
- [Terminal Image Protocol Comparison](https://github.com/trashhalo/imgcat)

---

*This specification is designed for implementation by Claude Code in YOLO mode with zero manual intervention required.*
