package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/placeli/internal/database"
	"github.com/user/placeli/internal/models"
	"github.com/user/placeli/internal/tui/mapview"
)

var (
	mapTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	mapHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type MapModel struct {
	db      *database.DB
	places  []*models.Place
	mapView *mapview.MapView
	width   int
	height  int
	message string
	loading bool
}

func NewMapModel(db *database.DB) MapModel {
	return MapModel{
		db:      db,
		loading: true,
	}
}

func (m MapModel) Init() tea.Cmd {
	return m.loadPlaces()
}

func (m MapModel) loadPlaces() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		places, err := m.db.ListPlaces(1000, 0) // Get up to 1000 places
		if err != nil {
			return errMsg{err}
		}
		return placesLoadedMsg{places}
	})
}

func (m MapModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateMapView()

	case placesLoadedMsg:
		m.places = msg.places
		m.loading = false
		m.updateMapView()
		if len(m.places) == 0 {
			m.message = "No places found. Use 'placeli import' to add places."
		} else {
			m.message = fmt.Sprintf("Loaded %d places", len(m.places))
		}

	case errMsg:
		m.loading = false
		m.message = fmt.Sprintf("Error: %v", msg.err)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.mapView != nil {
				m.mapView.Pan(0.01, 0)
				m.message = "Panned north"
			}

		case "down", "j":
			if m.mapView != nil {
				m.mapView.Pan(-0.01, 0)
				m.message = "Panned south"
			}

		case "left", "h":
			if m.mapView != nil {
				m.mapView.Pan(0, -0.01)
				m.message = "Panned west"
			}

		case "right", "l":
			if m.mapView != nil {
				m.mapView.Pan(0, 0.01)
				m.message = "Panned east"
			}

		case "+", "=":
			if m.mapView != nil {
				m.mapView.ZoomIn()
				m.message = fmt.Sprintf("Zoomed in (level %d)", m.mapView.ZoomLevel)
			}

		case "-", "_":
			if m.mapView != nil {
				m.mapView.ZoomOut()
				m.message = fmt.Sprintf("Zoomed out (level %d)", m.mapView.ZoomLevel)
			}

		case "f":
			if m.mapView != nil {
				m.mapView.FitBounds()
				m.message = "Fit all places in view"
			}

		case "r":
			return m, m.loadPlaces()

		case "t":
			if m.mapView != nil {
				m.mapView.ShowLabels = !m.mapView.ShowLabels
				if m.mapView.ShowLabels {
					m.message = "Labels enabled"
				} else {
					m.message = "Labels disabled"
				}
			}
		}
	}

	return m, nil
}

func (m MapModel) View() string {
	var b strings.Builder

	// Header
	title := mapTitleStyle.Render(fmt.Sprintf("placeli map (%d places)", len(m.places)))
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("Loading places...\n")
	} else if m.mapView != nil {
		// Render the map
		b.WriteString(m.mapView.Render())
	} else {
		b.WriteString("Map not initialized\n")
	}

	// Message
	if m.message != "" {
		b.WriteString(fmt.Sprintf("\n🔔 %s\n", m.message))
	}

	// Help
	help := mapHelpStyle.Render("↑↓←→/hjkl move • +/- zoom • f fit • t toggle labels • r refresh • q quit")
	b.WriteString(fmt.Sprintf("\n%s", help))

	return b.String()
}

func (m *MapModel) updateMapView() {
	if m.width == 0 || m.height == 0 || len(m.places) == 0 {
		return
	}

	// Calculate map dimensions (leave space for UI elements)
	mapWidth := m.width - 4
	mapHeight := m.height - 10

	if mapWidth < 20 {
		mapWidth = 20
	}
	if mapHeight < 10 {
		mapHeight = 10
	}

	config := mapview.MapConfig{
		Width:      mapWidth,
		Height:     mapHeight,
		ZoomLevel:  10,
		ShowLabels: false,
	}

	if m.mapView == nil {
		m.mapView = mapview.NewMapView(m.places, config)
		m.mapView.FitBounds() // Fit all places initially
	} else {
		// Update existing mapview with new dimensions
		m.mapView.Width = mapWidth
		m.mapView.Height = mapHeight
		m.mapView.Places = m.places
	}
}

func RunMap(db *database.DB) error {
	m := NewMapModel(db)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}