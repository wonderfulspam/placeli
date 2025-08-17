# placeli

A powerful terminal-based Google Maps list manager that gives you complete
control over your saved places.

![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Coverage](https://img.shields.io/badge/coverage-90%25%2B-brightgreen)

## Features

### 🚀 Core Capabilities

- **Import** your saved places from Google Takeout or other sources
- **Browse** places with a rich, keyboard-driven terminal interface
- **Search & Filter** by name, tags, ratings, distance, and custom fields
- **Enrich** data with Google Maps API (photos, reviews, hours)
- **Export** to CSV, JSON, GeoJSON, or Markdown
- **Web Interface** for viewing places on an interactive map

### 🎯 Advanced Features

- **Smart Sync** - Merge new Takeout data without duplicates
- **Tag Management** - Batch operations for organizing places
- **Custom Fields** - Add your own metadata (visited dates, priority, etc.)
- **Multi-Source Import** - Support for Apple Maps, OpenStreetMap, Foursquare
- **Terminal Map View** - ASCII-art map visualization right in your terminal

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/wonderfulspam/placeli.git
cd placeli

# Build the binary
go build -o placeli cmd/placeli/main.go

# Optional: Install globally
sudo mv placeli /usr/local/bin/
```

### Requirements

- Go 1.23 or higher
- SQLite3
- Google Maps API key (optional, for enrichment features)

## Getting Started

### Download Your Google Maps Data

Before using placeli, you'll need to export your Google Maps data:

#### Option 1: Google Takeout (Recommended)

1. Go to [Google Takeout](https://takeout.google.com/)
2. Click "Deselect all"
3. Scroll down and select "Maps (your places)"
4. Choose your preferred export format and delivery method
5. Download the resulting ZIP file

#### Option 2: Export Individual Lists

1. Open [Google Maps](https://maps.google.com/)
2. Click the menu (☰) → "Your places" → "Lists"
3. Open a list and click the share icon → "Export as KML"
4. Save the KML file to your computer

### 1. Import Your Data

```bash
# Import from Google Takeout
placeli imports from ~/Downloads/takeout-*.zip

# Import from KML files
placeli imports from ~/Downloads/places.kml

# Import from other sources
placeli imports from ~/Downloads/checkins.json --source=foursquare
```

### 2. Browse Your Places

```bash
# Launch interactive TUI
placeli browse

# Quick list view
placeli list --limit 20

# Filter by tags
placeli list --tags "restaurant,favorite"

# Search by name
placeli list --search "coffee"
```

### 3. Enrich with Google Maps Data

```bash
# Set your API key
export GOOGLE_MAPS_API_KEY="your-api-key"

# Enrich all places
placeli enrich

# Enrich specific places
placeli enrich --limit 10 --filter "restaurant"
```

### 4. Export Your Data

```bash
# Export to various formats
placeli export csv -o places.csv
placeli export json -o places.json
placeli export geojson -o places.geojson
placeli export markdown -o places.md

# Export with filters
placeli export csv --tags "to-visit" -o wishlist.csv
```

## Terminal UI Controls

The interactive TUI (`placeli browse` or `placeli review`) provides powerful
keyboard navigation:

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate list |
| `Enter` | View place details |
| `Tab` | Switch panels |
| `/` | Search |
| `f` | Filter menu |
| `t` | Add/edit tags |
| `n` | Add/edit notes |
| `e` | Edit custom fields |
| `m` | Toggle map view |
| `x` | Export current view |
| `q` or `Esc` | Back/Quit |

## Tag Management

Efficiently organize places with the tag system:

```bash
# List all tags with counts
placeli tags list

# Rename tags across all places
placeli tags rename "food" "restaurant"

# Delete a tag from all places
placeli tags delete "temporary"

# Batch apply tags
placeli tags apply "to-visit" --filter "rating>4.5"
```

## Custom Fields

Add your own metadata to places:

```bash
# Define custom fields
placeli fields add visited_date --type date
placeli fields add priority --type number
placeli fields add cuisine --type list

# List defined fields
placeli fields list

# Use templates
placeli fields templates
```

## Sync & Updates

Keep your data current without duplicates:

```bash
# Smart sync with new Takeout data
placeli sync ~/Downloads/new-takeout.zip

# Merge strategies
placeli sync --merge    # Preserve local edits (default)
placeli sync --force    # Overwrite with new data
placeli sync --skip     # Keep existing, skip conflicts
```

## Terminal Map View

Visualize your places directly in the terminal:

```bash
# Interactive map mode
placeli map

# Show map in list output
placeli list --map

# Map controls:
# Arrow keys: Pan
# +/-: Zoom in/out
# Enter: Select place
# c: Center on selection
```

## Web Interface

Launch a local web server to view places on an interactive map:

```bash
# Start web server
placeli web

# Custom port
placeli web --port 3000

# Open browser automatically
placeli web --open
```

## Configuration

### Environment Variables

```bash
# Google Maps API key for enrichment
export GOOGLE_MAPS_API_KEY="your-key"

# Default export format
export PLACELI_EXPORT_FORMAT="json"

# Database location
export PLACELI_DB_PATH="~/places.db"
```

### Database Location

By default, placeli stores data in `~/.placeli/places.db`. You can override this:

```bash
placeli --db /path/to/custom.db browse
```

## Examples

### Find Unvisited Restaurants

```bash
placeli list --tags "restaurant" --custom "visited:false" --sort rating
```

### Export Weekend Trip Ideas

```bash
placeli export markdown \
  --tags "activity,museum,park" \
  --within "50km" \
  --output weekend-ideas.md
```

### Bulk Import from Multiple Sources

```bash
# Import everything at once
placeli imports from ~/takeout.zip
placeli imports from ~/apple-maps.kml
placeli imports from ~/checkins.json --source=foursquare
placeli sync  # Deduplicate
```

## Development

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Check code quality
.claude/check

# Build for release
go build -ldflags="-s -w" -o placeli cmd/placeli/main.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details.

## Acknowledgments

Built with:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI
- [SQLite](https://sqlite.org/) - Local database
- [Google Maps API](https://developers.google.com/maps) - Place enrichment
