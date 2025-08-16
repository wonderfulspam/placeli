package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/user/placeli/internal/database"
	"github.com/user/placeli/internal/models"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170")).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type BrowseModel struct {
	db       *database.DB
	places   []*models.Place
	cursor   int
	selected map[int]struct{}
	width    int
	height   int
	search   string
	message  string
}

func NewBrowseModel(db *database.DB) BrowseModel {
	return BrowseModel{
		db:       db,
		selected: make(map[int]struct{}),
	}
}

func (m BrowseModel) Init() tea.Cmd {
	return m.loadPlaces()
}

func (m BrowseModel) loadPlaces() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.search != "" {
			places, err := m.db.SearchPlaces(m.search)
			if err != nil {
				return errMsg{err}
			}
			return placesLoadedMsg{places}
		}
		places, err := m.db.ListPlaces(100, 0) // Load first 100 places
		if err != nil {
			return errMsg{err}
		}
		return placesLoadedMsg{places}
	})
}

type placesLoadedMsg struct {
	places []*models.Place
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

func (m BrowseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case placesLoadedMsg:
		m.places = msg.places
		if m.cursor >= len(m.places) {
			m.cursor = 0
		}

	case errMsg:
		m.message = fmt.Sprintf("Error: %v", msg.err)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.places)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.places) > 0 {
				_, ok := m.selected[m.cursor]
				if ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}

		case "g":
			m.cursor = 0

		case "G":
			if len(m.places) > 0 {
				m.cursor = len(m.places) - 1
			}

		case "/":
			// TODO: Implement search mode
			m.message = "Search mode not yet implemented"

		case "r":
			return m, m.loadPlaces()
		}
	}

	return m, nil
}

func (m BrowseModel) View() string {
	var b strings.Builder

	// Header
	title := titleStyle.Render(fmt.Sprintf("placeli browse (%d places)", len(m.places)))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Message
	if m.message != "" {
		b.WriteString(fmt.Sprintf("🔔 %s\n\n", m.message))
	}

	// Places list
	if len(m.places) == 0 {
		b.WriteString("No places found. Use 'placeli import' to add places.\n")
	} else {
		for i, place := range m.places {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.selected[i]; ok {
				checked = "✓"
			}

			stars := ""
			if place.Rating > 0 {
				starCount := int(place.Rating)
				stars = strings.Repeat("⭐", starCount)
			}

			line := fmt.Sprintf("%s [%s] %s", cursor, checked, place.Name)
			if place.Address != "" {
				line += fmt.Sprintf("\n     📍 %s", place.Address)
			}
			if stars != "" {
				line += fmt.Sprintf(" %s %.1f", stars, place.Rating)
			}
			if len(place.Categories) > 0 {
				line += fmt.Sprintf("\n     🏷️  %s", strings.Join(place.Categories, ", "))
			}

			if i == m.cursor {
				b.WriteString(selectedItemStyle.Render(line))
			} else {
				b.WriteString(itemStyle.Render(line))
			}
			b.WriteString("\n\n")
		}
	}

	// Help
	help := helpStyle.Render("↑/k up • ↓/j down • enter/space select • g top • G bottom • r refresh • / search • q quit")
	b.WriteString(fmt.Sprintf("\n%s", help))

	return b.String()
}

func RunBrowse(db *database.DB) error {
	m := NewBrowseModel(db)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
