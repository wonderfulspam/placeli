# Placeli Improvement TODOs

This document outlines areas for improvement in the Placeli codebase, focusing on architectural design, code quality, and testing.

##  Architectural & Design Improvements

- **[ ] Refactor Database Layer:** Create a helper function in the `database` package to reduce code duplication in `GetPlace`, `ListPlaces`, and `SearchPlaces`. The current implementation has a lot of repeated code for scanning rows and unmarshaling JSON.
- **[ ] Decouple Database Initialization:** Move database initialization out of the `main` package to reduce tight coupling. A dependency injection container or a factory pattern could be used to manage the database instance.
- **[ ] Introduce a Service Layer:** Create a service layer to abstract business logic from the `cmd` package. This will make the code more modular, easier to test, and less dependent on the command-line interface.
- **[ ] Centralize Configuration Management:** Implement a dedicated configuration management solution for handling settings like the Google Maps API key and database path. This should support configuration from a file, environment variables, and command-line flags.
- **[ ] Consistent Error Handling:** Establish a consistent error handling strategy throughout the application. Avoid printing errors directly to the console in lower-level packages.

## Testing

- **[ ] Add Tests for `importer/takeout.go`:** Write tests for the `ImportFromTakeout` function to ensure the importer works as expected.
- **[ ] Add Tests for `tui/review.go`:** The interactive TUI in `review.go` is untested. Add tests to cover the UI logic and state management.
- **[ ] Add Tests for CLI Commands:** Implement tests for the Cobra commands in the `cmd` package to verify their behavior.
- **[ ] Improve `database` Tests:** Enhance the existing tests for the `database` package to cover more edge cases, such as empty results and handling of invalid data.

## Code Quality & General Improvements

- **[ ] Remove Hardcoded Values:** Replace hardcoded values, such as the `10000` limit in `export.go` and `enrich.go`, and the photo limit in `enrichment.go`, with configurable options.
- **[ ] Implement Structured Logging:** Replace `fmt.Printf` with a structured logging library to provide more informative and filterable log output.
- **[ ] Add Code Comments:** Add comments to complex sections of the code, particularly in the `tui` package, to improve readability and maintainability.
- **[ ] Implement `DeletePlace`:** The `deletePlace` function in `internal/tui/review.go` is a stub. Implement the actual database deletion logic.
- **[ ] Add Search to Review TUI:** The search functionality is missing in the review TUI. Implement it to be consistent with the browse TUI.

## Test Coverage Gaps

- **[ ] Write comprehensive tests for the `tui` package:** The TUI is a critical part of the application, but it has very little test coverage. Tests should cover user interactions, state changes, and rendering.
- **[ ] Implement tests for the `cmd` package:** The CLI commands are the main entry point for the application, but they are not tested. Tests should cover command-line arguments, flags, and error handling.
- **[ ] Improve tests for the `importer` package:** The existing tests for the importer are minimal. They should be expanded to cover edge cases, such as invalid or malformed Takeout data.

## Feature Implementation Gaps (from spec.md)

- **[ ] Implement `placeli sync` command:** The `sync` command is a core feature in the spec, but it is not implemented. This command should intelligently merge new data from Google Takeout without creating duplicates.
- **[ ] Implement `placeli web` command:** The spec mentions a web interface for viewing places on a map. This feature is not implemented.
- **[ ] Implement advanced filtering:** The `list` command and the TUI lack the advanced filtering capabilities described in the spec (e.g., by tags, categories, rating, proximity).
- **[ ] Rename `list` command to `query`:** The spec uses the name `query` for the command to query the local database. The current implementation uses `list`. Consider renaming the command for consistency.