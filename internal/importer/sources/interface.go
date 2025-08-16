package sources

import (
	"github.com/user/placeli/internal/models"
)

// ImportSource represents a source that can import places
type ImportSource interface {
	// Name returns the human-readable name of the source
	Name() string
	
	// SupportedFormats returns file extensions or format types this source can handle
	SupportedFormats() []string
	
	// ImportFromFile imports places from a file path
	ImportFromFile(filePath string) ([]*models.Place, error)
	
	// ImportFromData imports places from raw data
	ImportFromData(data []byte, format string) ([]*models.Place, error)
}

// SourceManager manages multiple import sources
type SourceManager struct {
	sources map[string]ImportSource
}

// NewSourceManager creates a new source manager with all available sources
func NewSourceManager() *SourceManager {
	sm := &SourceManager{
		sources: make(map[string]ImportSource),
	}
	
	// Register all available sources
	sm.RegisterSource("apple", &AppleImporter{})
	sm.RegisterSource("osm", &OSMImporter{})
	sm.RegisterSource("foursquare", &FoursquareImporter{})
	
	return sm
}

// RegisterSource registers a new import source
func (sm *SourceManager) RegisterSource(name string, source ImportSource) {
	sm.sources[name] = source
}

// GetSource returns a source by name
func (sm *SourceManager) GetSource(name string) ImportSource {
	return sm.sources[name]
}

// ListSources returns all available sources
func (sm *SourceManager) ListSources() map[string]ImportSource {
	return sm.sources
}

// DetectSource attempts to detect the appropriate source for a file
func (sm *SourceManager) DetectSource(filePath string) ImportSource {
	for _, source := range sm.sources {
		for _, format := range source.SupportedFormats() {
			if matchesFormat(filePath, format) {
				return source
			}
		}
	}
	return nil
}

// ImportFromFile imports places from a file using the appropriate source
func (sm *SourceManager) ImportFromFile(filePath string) ([]*models.Place, error) {
	source := sm.DetectSource(filePath)
	if source == nil {
		// Try to detect from file content
		for _, src := range sm.sources {
			places, err := src.ImportFromFile(filePath)
			if err == nil && len(places) > 0 {
				return places, nil
			}
		}
		return nil, &UnsupportedFormatError{FilePath: filePath}
	}
	
	return source.ImportFromFile(filePath)
}

// UnsupportedFormatError is returned when no source can handle a file
type UnsupportedFormatError struct {
	FilePath string
}

func (e *UnsupportedFormatError) Error() string {
	return "unsupported format for file: " + e.FilePath
}

// Helper functions

func matchesFormat(filePath, format string) bool {
	// Simple extension matching for now
	// Could be enhanced with more sophisticated detection
	switch format {
	case "kml", "kmz":
		return hasExtension(filePath, ".kml") || hasExtension(filePath, ".kmz")
	case "gpx":
		return hasExtension(filePath, ".gpx")
	case "json":
		return hasExtension(filePath, ".json")
	case "csv":
		return hasExtension(filePath, ".csv")
	}
	return false
}

func hasExtension(filePath, ext string) bool {
	return len(filePath) >= len(ext) && 
		   filePath[len(filePath)-len(ext):] == ext
}