package mapview

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/user/placeli/internal/models"
)

// MapView represents an ASCII map display
type MapView struct {
	Width      int
	Height     int
	CenterLat  float64
	CenterLng  float64
	ZoomLevel  int
	Places     []*models.Place
	ShowLabels bool
}

// MapConfig holds configuration for map rendering
type MapConfig struct {
	Width      int
	Height     int
	ZoomLevel  int
	ShowLabels bool
}

// BoundingBox represents geographical bounds
type BoundingBox struct {
	MinLat, MaxLat float64
	MinLng, MaxLng float64
}

// Point represents a screen coordinate
type Point struct {
	X, Y int
}

// PlaceMarker represents a place marker on the map
type PlaceMarker struct {
	Place  *models.Place
	Point  Point
	Symbol string
	Label  string
}

// NewMapView creates a new map view
func NewMapView(places []*models.Place, config MapConfig) *MapView {
	mv := &MapView{
		Width:      config.Width,
		Height:     config.Height,
		ZoomLevel:  config.ZoomLevel,
		Places:     places,
		ShowLabels: config.ShowLabels,
	}

	// Calculate center from places
	if len(places) > 0 {
		mv.CenterLat, mv.CenterLng = mv.calculateCenter()
	}

	return mv
}

// Render generates the ASCII map
func (mv *MapView) Render() string {
	if len(mv.Places) == 0 {
		return mv.renderEmpty()
	}

	// Create map grid
	grid := make([][]rune, mv.Height)
	for i := range grid {
		grid[i] = make([]rune, mv.Width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Draw border
	mv.drawBorder(grid)

	// Draw grid lines
	mv.drawGrid(grid)

	// Place markers
	markers := mv.createMarkers()
	mv.placeMarkers(grid, markers)

	// Convert grid to string
	result := mv.gridToString(grid)

	// Add labels if enabled
	if mv.ShowLabels {
		result += mv.renderLabels(markers)
	}

	// Add info panel
	result += mv.renderInfo()

	return result
}

// calculateCenter finds the geographic center of all places
func (mv *MapView) calculateCenter() (lat, lng float64) {
	if len(mv.Places) == 0 {
		return 0, 0
	}

	var sumLat, sumLng float64
	for _, place := range mv.Places {
		sumLat += place.Coordinates.Lat
		sumLng += place.Coordinates.Lng
	}

	return sumLat / float64(len(mv.Places)), sumLng / float64(len(mv.Places))
}

// getBounds calculates the bounding box for all places
func (mv *MapView) getBounds() BoundingBox {
	if len(mv.Places) == 0 {
		return BoundingBox{}
	}

	bounds := BoundingBox{
		MinLat: mv.Places[0].Coordinates.Lat,
		MaxLat: mv.Places[0].Coordinates.Lat,
		MinLng: mv.Places[0].Coordinates.Lng,
		MaxLng: mv.Places[0].Coordinates.Lng,
	}

	for _, place := range mv.Places {
		if place.Coordinates.Lat < bounds.MinLat {
			bounds.MinLat = place.Coordinates.Lat
		}
		if place.Coordinates.Lat > bounds.MaxLat {
			bounds.MaxLat = place.Coordinates.Lat
		}
		if place.Coordinates.Lng < bounds.MinLng {
			bounds.MinLng = place.Coordinates.Lng
		}
		if place.Coordinates.Lng > bounds.MaxLng {
			bounds.MaxLng = place.Coordinates.Lng
		}
	}

	return bounds
}

// coordToScreen converts geographic coordinates to screen coordinates
func (mv *MapView) coordToScreen(lat, lng float64) Point {
	// Calculate the scale based on zoom level
	scale := math.Pow(2, float64(mv.ZoomLevel-10))

	// Mercator projection
	x := (lng - mv.CenterLng) * scale
	y := (mv.CenterLat - lat) * scale * 0.5 // Adjust for aspect ratio

	// Convert to screen coordinates
	screenX := int(x*float64(mv.Width)/4) + mv.Width/2
	screenY := int(y*float64(mv.Height)/2) + mv.Height/2

	return Point{X: screenX, Y: screenY}
}

// createMarkers creates markers for all places
func (mv *MapView) createMarkers() []PlaceMarker {
	var markers []PlaceMarker

	// Group places by screen position for clustering
	positionMap := make(map[Point][]*models.Place)

	for _, place := range mv.Places {
		point := mv.coordToScreen(place.Coordinates.Lat, place.Coordinates.Lng)

		// Only include points within bounds
		if point.X >= 0 && point.X < mv.Width && point.Y >= 0 && point.Y < mv.Height {
			positionMap[point] = append(positionMap[point], place)
		}
	}

	// Create markers for each position
	for point, places := range positionMap {
		marker := mv.createMarkerForPosition(point, places)
		markers = append(markers, marker)
	}

	return markers
}

// createMarkerForPosition creates a marker for places at the same position
func (mv *MapView) createMarkerForPosition(point Point, places []*models.Place) PlaceMarker {
	if len(places) == 1 {
		// Single place
		place := places[0]
		symbol := mv.getMarkerSymbol(place)
		label := mv.getMarkerLabel(place)

		return PlaceMarker{
			Place:  place,
			Point:  point,
			Symbol: symbol,
			Label:  label,
		}
	}

	// Multiple places - use cluster marker
	symbol := mv.getClusterSymbol(len(places))
	label := fmt.Sprintf("%d places", len(places))

	// Use the first place as representative
	return PlaceMarker{
		Place:  places[0],
		Point:  point,
		Symbol: symbol,
		Label:  label,
	}
}

// getMarkerSymbol returns the appropriate symbol for a place
func (mv *MapView) getMarkerSymbol(place *models.Place) string {
	// Choose symbol based on categories
	for _, category := range place.Categories {
		switch strings.ToLower(category) {
		case "restaurant", "food":
			return "🍽"
		case "cafe", "coffee":
			return "☕"
		case "bar", "pub":
			return "🍺"
		case "hotel", "lodging":
			return "🏨"
		case "gas station", "fuel":
			return "⛽"
		case "hospital", "medical":
			return "🏥"
		case "school", "education":
			return "🏫"
		case "bank", "finance":
			return "🏦"
		case "shopping", "store":
			return "🛍"
		case "park", "outdoor":
			return "🌳"
		}
	}

	// Default marker
	return "📍"
}

// getClusterSymbol returns symbol for clustered places
func (mv *MapView) getClusterSymbol(count int) string {
	if count < 10 {
		return fmt.Sprintf("%d", count)
	} else if count < 100 {
		return "+"
	} else {
		return "*"
	}
}

// getMarkerLabel creates a label for a place
func (mv *MapView) getMarkerLabel(place *models.Place) string {
	name := place.Name
	if len(name) > 20 {
		name = name[:17] + "..."
	}
	return name
}

// drawBorder draws the map border using box-drawing characters
func (mv *MapView) drawBorder(grid [][]rune) {
	// Top and bottom borders
	for x := 0; x < mv.Width; x++ {
		if x == 0 {
			grid[0][x] = '┌'
			grid[mv.Height-1][x] = '└'
		} else if x == mv.Width-1 {
			grid[0][x] = '┐'
			grid[mv.Height-1][x] = '┘'
		} else {
			grid[0][x] = '─'
			grid[mv.Height-1][x] = '─'
		}
	}

	// Left and right borders
	for y := 1; y < mv.Height-1; y++ {
		grid[y][0] = '│'
		grid[y][mv.Width-1] = '│'
	}
}

// drawGrid draws grid lines for reference
func (mv *MapView) drawGrid(grid [][]rune) {
	// Draw sparse grid lines
	gridSpacing := 8

	// Vertical grid lines
	for x := gridSpacing; x < mv.Width-1; x += gridSpacing {
		for y := 1; y < mv.Height-1; y++ {
			if grid[y][x] == ' ' {
				grid[y][x] = '┊'
			}
		}
	}

	// Horizontal grid lines
	for y := gridSpacing; y < mv.Height-1; y += gridSpacing {
		for x := 1; x < mv.Width-1; x++ {
			if grid[y][x] == ' ' {
				grid[y][x] = '┈'
			} else if grid[y][x] == '┊' {
				grid[y][x] = '┼'
			}
		}
	}
}

// placeMarkers places all markers on the grid
func (mv *MapView) placeMarkers(grid [][]rune, markers []PlaceMarker) {
	for _, marker := range markers {
		if marker.Point.X > 0 && marker.Point.X < mv.Width-1 &&
			marker.Point.Y > 0 && marker.Point.Y < mv.Height-1 {

			// Use first character of symbol
			symbol := []rune(marker.Symbol)
			if len(symbol) > 0 {
				grid[marker.Point.Y][marker.Point.X] = symbol[0]
			} else {
				grid[marker.Point.Y][marker.Point.X] = '*'
			}
		}
	}
}

// gridToString converts the grid to a string
func (mv *MapView) gridToString(grid [][]rune) string {
	var builder strings.Builder

	for _, row := range grid {
		for _, char := range row {
			builder.WriteRune(char)
		}
		builder.WriteRune('\n')
	}

	return builder.String()
}

// renderLabels generates labels for markers
func (mv *MapView) renderLabels(markers []PlaceMarker) string {
	if len(markers) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("\nPlaces shown:\n")

	// Sort markers by Y position for consistent display
	sort.Slice(markers, func(i, j int) bool {
		if markers[i].Point.Y == markers[j].Point.Y {
			return markers[i].Point.X < markers[j].Point.X
		}
		return markers[i].Point.Y < markers[j].Point.Y
	})

	for i, marker := range markers {
		if i >= 10 { // Limit to 10 labels
			builder.WriteString(fmt.Sprintf("... and %d more\n", len(markers)-i))
			break
		}
		builder.WriteString(fmt.Sprintf("%s %s\n", marker.Symbol, marker.Label))
	}

	return builder.String()
}

// renderInfo generates info panel
func (mv *MapView) renderInfo() string {
	bounds := mv.getBounds()

	return fmt.Sprintf(`
Map Info:
  Center: %.4f, %.4f
  Zoom: %d
  Places: %d
  Bounds: %.4f,%.4f to %.4f,%.4f
`,
		mv.CenterLat, mv.CenterLng,
		mv.ZoomLevel,
		len(mv.Places),
		bounds.MinLat, bounds.MinLng,
		bounds.MaxLat, bounds.MaxLng,
	)
}

// renderEmpty renders an empty map message
func (mv *MapView) renderEmpty() string {
	return fmt.Sprintf(`
╭%s╮
│%s│
│%s│
│%s│
╰%s╯

No places to display on map.
`,
		strings.Repeat("─", mv.Width-2),
		mv.centerText("MAP VIEW", mv.Width-2),
		mv.centerText("", mv.Width-2),
		mv.centerText("No places found", mv.Width-2),
		strings.Repeat("─", mv.Width-2),
	)
}

// centerText centers text within a given width
func (mv *MapView) centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}

	padding := width - len(text)
	leftPad := padding / 2
	rightPad := padding - leftPad

	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

// SetCenter updates the map center
func (mv *MapView) SetCenter(lat, lng float64) {
	mv.CenterLat = lat
	mv.CenterLng = lng
}

// SetZoom updates the zoom level
func (mv *MapView) SetZoom(zoom int) {
	if zoom < 1 {
		zoom = 1
	}
	if zoom > 20 {
		zoom = 20
	}
	mv.ZoomLevel = zoom
}

// Pan moves the map center
func (mv *MapView) Pan(deltaLat, deltaLng float64) {
	mv.CenterLat += deltaLat
	mv.CenterLng += deltaLng
}

// ZoomIn increases zoom level
func (mv *MapView) ZoomIn() {
	mv.SetZoom(mv.ZoomLevel + 1)
}

// ZoomOut decreases zoom level
func (mv *MapView) ZoomOut() {
	mv.SetZoom(mv.ZoomLevel - 1)
}

// FitBounds adjusts zoom and center to fit all places
func (mv *MapView) FitBounds() {
	if len(mv.Places) == 0 {
		return
	}

	bounds := mv.getBounds()

	// Set center to bounds center
	mv.CenterLat = (bounds.MinLat + bounds.MaxLat) / 2
	mv.CenterLng = (bounds.MinLng + bounds.MaxLng) / 2

	// Calculate appropriate zoom level
	latSpan := bounds.MaxLat - bounds.MinLat
	lngSpan := bounds.MaxLng - bounds.MinLng

	// Use larger span to determine zoom
	maxSpan := math.Max(latSpan, lngSpan)

	// Calculate zoom level (rough approximation)
	if maxSpan > 10 {
		mv.ZoomLevel = 6
	} else if maxSpan > 5 {
		mv.ZoomLevel = 8
	} else if maxSpan > 1 {
		mv.ZoomLevel = 10
	} else if maxSpan > 0.5 {
		mv.ZoomLevel = 12
	} else if maxSpan > 0.1 {
		mv.ZoomLevel = 14
	} else {
		mv.ZoomLevel = 16
	}
}
