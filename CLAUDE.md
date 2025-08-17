# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Project Overview

placeli is a terminal-based Google Maps list manager written in Go. It provides
local, offline-first access to saved places with a rich TUI interface for
browsing, searching, and managing location data.

**Current Status**: Check [STATUS.md](STATUS.md) for detailed implementation
progress, completed features, and next priorities. Always review this file
before starting new work to understand the current state.

## Development Workflow

### Test-Driven Development

1. Write tests first in `*_test.go` files
2. Run `go test ./...` to see tests fail
3. Implement code until tests pass
4. Leverage Go's test caching - always run all tests

### Quality Checks

```bash
.claude/check  # Runs fmt, vet, test, mod tidy, golangci-lint, git checks and more
```

The `.claude/check` script runs automatically on pre-commit.

### Git Commit Practices

- Never commit test output files or temporary files
- Write clear commit messages describing the change
- Commit incrementally as features are completed

### Documentation Updates

- Update docs immediately after code changes
- Remove outdated information when adding new content
- Keep documentation concise - delete as much as you add
- Update this CLAUDE.md if development practices change

## Architecture

### Project Structure

- `cmd/placeli/` - Main entry point and CLI commands
- `internal/` - Private application code
  - `internal/tui/` - Terminal UI components (bubbletea)
  - `internal/database/` - SQLite database layer
  - `internal/maps/` - Google Maps API integration
  - `internal/export/` - Export functionality
  - `internal/models/` - Core data models

### Key Technologies

- **Language**: Go
- **Database**: SQLite with JSON fields
- **TUI**: bubbletea/bubbles or tview
- **CLI**: cobra for command parsing

### Core Data Model

The `Place` struct with enriched media (photos, reviews) and user data (notes,
tags). SQLite storage with flexible JSON fields.

### Commands

- `import` - Import from Google Takeout/Maps Lists
- `review` - Interactive TUI for reviewing places
- `browse` - Browse local database
- `export` - Export to various formats
- `web` - Simple web interface

## Build Commands

```bash
go build -o placeli ./cmd/placeli/  # Build binary
go test -race ./...                      # Test with race detector
go test -cover ./...                     # Test with coverage
```

## Testing Strategy

- Unit tests for all business logic
- Integration tests for database operations
- Mock external APIs (Google Maps)
- Use testify for assertions and mocks

## Development Guidelines

- Keep TUI responsive with async data loading
- Cache API responses to minimize calls
- Use context for cancellation and timeouts
- Wrap errors with meaningful context
- Store API keys in environment variables
