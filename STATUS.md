# placeli Development Status

## ✅ COMPLETED

**Core**: Database, Models, Import, CLI Foundation
**Commands**: `import`, `list`, `browse`, `review`, `export`, `enrich`, `web`
**Features**: TUI, Export (CSV/JSON/GeoJSON/Markdown), Maps API, Web Interface
**Quality**: >90% test coverage, linting, pre-commit hooks

## 🚧 ACTIVE DEVELOPMENT

### 1. Sync Command - Intelligent Takeout Merging

**Goal**: Allow users to periodically import new Takeout data without duplicates
**Implementation**:

- [ ] Create `cmd/placeli/sync.go` command
- [ ] Add `imported_at` and `source_hash` fields to database
- [ ] Implement duplicate detection based on place ID/coordinates
- [ ] Create merge strategy: preserve user edits, update API data
- [ ] Add conflict resolution options (--force, --merge, --skip)
- [ ] Test with multiple Takeout exports

### 2. Tag Management System

**Goal**: Efficient batch operations for organizing places with tags
**Implementation**:

- [ ] Create `cmd/placeli/tags.go` command with subcommands
- [ ] Add `tags list` - show all tags with counts
- [ ] Add `tags rename <old> <new>` - batch rename
- [ ] Add `tags delete <tag>` - remove tag from all places
- [ ] Add `tags apply <tag> --filter=<query>` - batch apply
- [ ] Update TUI to support multi-select for batch tagging
- [ ] Add tag autocomplete in TUI

### 3. Custom Fields

**Goal**: User-defined metadata fields for specialized use cases
**Implementation**:

- [ ] Add `custom_fields` JSON column to places table
- [ ] Create `cmd/placeli/fields.go` for field management
- [ ] Add field types: text, number, date, boolean, list
- [ ] Update TUI to display and edit custom fields
- [ ] Add custom field support to export formats
- [ ] Create field templates for common use cases

### 4. Multi-Source Import

**Goal**: Import from Apple Maps, OpenStreetMap, and other services
**Implementation**:

- [ ] Create `internal/importer/sources/` package
- [ ] Add Apple Maps KML/GPX parser
- [ ] Add OpenStreetMap Overpass API integration
- [ ] Add Foursquare/Swarm history import
- [ ] Create unified import interface
- [ ] Handle format-specific quirks and data mapping

### 5. Terminal Map View

**Goal**: ASCII-art map visualization in the terminal
**Implementation**:

- [ ] Create `internal/tui/mapview/` package
- [ ] Use Unicode box-drawing characters for map
- [ ] Implement zoom levels and pan controls
- [ ] Add place markers with density clustering
- [ ] Create mini-map for `list` command output
- [ ] Add interactive map mode in TUI

## 📊 PROJECT STATS

**Lines of Code**: ~4,000+
**Test Coverage**: >90%
**Stack**: Go 1.21+, SQLite, Cobra CLI, Bubbletea TUI

## 🔄 UPDATE LOG

- **2025-08-16**: Completed core features (import, TUI, export, enrichment, web)
- **2025-08-16**: Started advanced features implementation
