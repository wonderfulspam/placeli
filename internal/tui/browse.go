package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	db          *database.DB
	places      []*models.Place
	cursor      int
	selected    map[int]struct{}
	width       int
	height      int
	search      string
	message     string
	searchMode  bool
	searchInput string
	tagMode     bool
	tagInput    string
	tagAction   string // "add" or "remove"
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
		places, err := m.db.ListPlaces(1000, 0) // Increased from hardcoded 100
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
		// Ensure cursor stays visible after resize
		return m, nil

	case placesLoadedMsg:
		m.places = msg.places
		if m.cursor >= len(m.places) {
			m.cursor = 0
		}
		return m, nil

	case errMsg:
		m.message = fmt.Sprintf("Error: %v", msg.err)
		return m, nil

	case tagAppliedMsg:
		action := "added to"
		if msg.action == "remove" {
			action = "removed from"
		}
		m.message = fmt.Sprintf("Tag '%s' %s %d places", msg.tag, action, msg.count)
		// Clear selection after applying tags
		m.selected = make(map[int]struct{})
		return m, nil

	case tea.KeyMsg:
		if m.searchMode {
			return m.updateSearch(msg)
		}
		if m.tagMode {
			return m.updateTag(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.places)-1 {
				m.cursor++
			}
			return m, nil

		case "enter", " ":
			if len(m.places) > 0 {
				_, ok := m.selected[m.cursor]
				if ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}
			return m, nil

		case "g":
			m.cursor = 0
			return m, nil

		case "G":
			if len(m.places) > 0 {
				m.cursor = len(m.places) - 1
			}
			return m, nil

		case "/":
			m.searchMode = true
			m.searchInput = ""
			m.message = ""
			return m, nil

		case "c":
			// Clear search
			m.search = ""
			m.searchInput = ""
			m.message = "Search cleared"
			return m, m.loadPlaces()

		case "r":
			return m, m.loadPlaces()

		case "t":
			// Add tag to selected places
			if len(m.selected) > 0 {
				m.tagMode = true
				m.tagAction = "add"
				m.tagInput = ""
				m.message = ""
			} else {
				m.message = "Select places first (space/enter to select)"
			}
			return m, nil

		case "T":
			// Remove tag from selected places
			if len(m.selected) > 0 {
				m.tagMode = true
				m.tagAction = "remove"
				m.tagInput = ""
				m.message = ""
			} else {
				m.message = "Select places first (space/enter to select)"
			}
			return m, nil
		}
	}

	return m, nil
}

func (m BrowseModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.searchMode = false
		m.searchInput = ""

	case "enter":
		m.searchMode = false
		m.search = m.searchInput
		m.cursor = 0
		if m.search == "" {
			m.message = "Search cleared"
		} else {
			m.message = fmt.Sprintf("Searching for: %s", m.search)
		}
		return m, m.loadPlaces()

	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}

	default:
		// Add printable characters to search input
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			m.searchInput += msg.String()
		}
	}

	return m, nil
}

func (m BrowseModel) updateTag(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.tagMode = false
		m.tagInput = ""

	case "enter":
		m.tagMode = false
		if m.tagInput != "" {
			return m, m.applyTagToSelected(m.tagInput, m.tagAction)
		}

	case "backspace":
		if len(m.tagInput) > 0 {
			m.tagInput = m.tagInput[:len(m.tagInput)-1]
		}

	default:
		// Add printable characters to tag input
		if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
			m.tagInput += msg.String()
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

	// Search interface
	if m.searchMode {
		searchPrompt := fmt.Sprintf("Search: %s_", m.searchInput)
		b.WriteString(fmt.Sprintf("🔍 %s\n", searchPrompt))
		b.WriteString("(enter to search, esc to cancel)\n\n")
	} else if m.search != "" {
		b.WriteString(fmt.Sprintf("🔍 Active search: %s (press 'c' to clear)\n\n", m.search))
	}

	// Tag interface
	if m.tagMode {
		action := "Add"
		if m.tagAction == "remove" {
			action = "Remove"
		}
		tagPrompt := fmt.Sprintf("%s tag: %s_", action, m.tagInput)
		b.WriteString(fmt.Sprintf("🏷️  %s\n", tagPrompt))
		b.WriteString(fmt.Sprintf("(enter to %s tag to %d places, esc to cancel)\n\n", strings.ToLower(action), len(m.selected)))
	}

	// Message
	if m.message != "" {
		b.WriteString(fmt.Sprintf("🔔 %s\n\n", m.message))
	}

	// Places list with responsive pagination
	if len(m.places) == 0 {
		b.WriteString("No places found. Use 'placeli import' to add places.\n")
	} else {
		start, end := m.calculateViewport()

		// Show pagination info with terminal size
		b.WriteString(fmt.Sprintf("Showing %d-%d of %d places (cursor at %d) [%dx%d]\n\n",
			start+1, end, len(m.places), m.cursor+1, m.width, m.height))

		// Render visible items
		for i := start; i < end; i++ {
			place := m.places[i]
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
			if len(place.UserTags) > 0 {
				line += fmt.Sprintf("\n     🔖 %s", strings.Join(place.UserTags, ", "))
			}
			if len(place.CustomFields) > 0 {
				customFieldsDisplay := m.formatCustomFields(place.CustomFields)
				if customFieldsDisplay != "" {
					line += fmt.Sprintf("\n     ⚙️  %s", customFieldsDisplay)
				}
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
	if m.searchMode {
		help := helpStyle.Render("Type to search • enter confirm • esc cancel")
		b.WriteString(fmt.Sprintf("\n%s", help))
	} else if m.tagMode {
		help := helpStyle.Render("Type tag name • enter confirm • esc cancel")
		b.WriteString(fmt.Sprintf("\n%s", help))
	} else {
		help := helpStyle.Render("↑/k up • ↓/j down • enter/space select • t add tag • T remove tag • g top • G bottom • / search • c clear • r refresh • q quit")
		b.WriteString(fmt.Sprintf("\n%s", help))
	}

	return b.String()
}

func (m BrowseModel) applyTagToSelected(tag, action string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		count := 0
		for i := range m.selected {
			if i < len(m.places) {
				place := m.places[i]

				if action == "add" {
					if !place.HasTag(tag) {
						place.AddTag(tag)
						if err := m.db.SavePlace(place); err != nil {
							return errMsg{err}
						}
						count++
					}
				} else if action == "remove" {
					if place.HasTag(tag) {
						place.RemoveTag(tag)
						if err := m.db.SavePlace(place); err != nil {
							return errMsg{err}
						}
						count++
					}
				}
			}
		}

		return tagAppliedMsg{tag: tag, action: action, count: count}
	})
}

// calculateViewport returns the start and end indices for visible items
func (m BrowseModel) calculateViewport() (start, end int) {
	if len(m.places) == 0 {
		return 0, 0
	}

	// Calculate available space for items based on terminal height
	usedLines := 2 // title
	if m.searchMode || m.search != "" {
		usedLines += 3 // search interface
	}
	if m.tagMode {
		usedLines += 3 // tag interface
	}
	if m.message != "" {
		usedLines += 2 // message
	}
	usedLines += 3 // pagination info + spacing
	usedLines += 2 // help text

	// Each place takes roughly 4-6 lines (name + address + categories + tags + custom fields + spacing)
	// Be conservative and assume 5 lines per item on average
	linesPerItem := 5
	availableHeight := m.height - usedLines
	if availableHeight < 10 {
		availableHeight = 10 // minimum reasonable height
	}

	itemsPerPage := availableHeight / linesPerItem
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}
	if itemsPerPage > len(m.places) {
		itemsPerPage = len(m.places)
	}

	// Calculate viewport to keep cursor visible and centered when possible
	start = m.cursor - itemsPerPage/2
	if start < 0 {
		start = 0
	}
	end = start + itemsPerPage
	if end > len(m.places) {
		end = len(m.places)
		start = end - itemsPerPage
		if start < 0 {
			start = 0
		}
	}

	// Ensure cursor is always visible
	if m.cursor < start {
		start = m.cursor
		end = start + itemsPerPage
		if end > len(m.places) {
			end = len(m.places)
		}
	} else if m.cursor >= end {
		end = m.cursor + 1
		start = end - itemsPerPage
		if start < 0 {
			start = 0
		}
	}

	return start, end
}

func (m BrowseModel) formatCustomFields(fields map[string]interface{}) string {
	var displayFields []string
	systemFields := map[string]bool{
		"google_maps_url": true,
		"imported_from":   true,
		"import_date":     true,
		"last_sync":       true,
	}

	// Get keys and sort them for deterministic output
	var keys []string
	for key := range fields {
		// Skip system fields
		if !systemFields[key] {
			keys = append(keys, key)
		}
	}

	// Sort keys alphabetically for consistent ordering
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for _, key := range keys {
		value := fields[key]

		// Format the value for display
		var valueStr string
		switch v := value.(type) {
		case string:
			if v != "" {
				valueStr = v
			}
		case float64:
			if v != 0 {
				valueStr = fmt.Sprintf("%.1f", v)
			}
		case bool:
			if v {
				valueStr = "true"
			}
		case []interface{}:
			if len(v) > 0 {
				var items []string
				for _, item := range v {
					items = append(items, fmt.Sprintf("%v", item))
				}
				valueStr = strings.Join(items, ",")
			}
		default:
			if value != nil {
				valueStr = fmt.Sprintf("%v", value)
			}
		}

		// Only show fields with non-empty values
		if valueStr != "" {
			displayFields = append(displayFields, fmt.Sprintf("%s:%s", key, valueStr))
		}
	}

	return strings.Join(displayFields, " ")
}

type tagAppliedMsg struct {
	tag    string
	action string
	count  int
}

func RunBrowse(db *database.DB) error {
	m := NewBrowseModel(db)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
