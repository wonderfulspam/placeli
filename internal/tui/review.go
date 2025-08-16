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
	detailTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#F25D94")).
				Padding(0, 1)

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2)

	fieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))
)

type ReviewMode int

const (
	ReviewModeList ReviewMode = iota
	ReviewModeDetail
	ReviewModeEdit
)

type ReviewModel struct {
	db        *database.DB
	places    []*models.Place
	cursor    int
	current   *models.Place
	mode      ReviewMode
	width     int
	height    int
	message   string
	editField string
	editValue string
}

func NewReviewModel(db *database.DB) ReviewModel {
	return ReviewModel{
		db:   db,
		mode: ReviewModeList,
	}
}

func (m ReviewModel) Init() tea.Cmd {
	return m.loadPlaces()
}

func (m ReviewModel) loadPlaces() tea.Cmd {
	return m.loadPlacesWithLimit(1000) // Increased from hardcoded 100
}

func (m ReviewModel) loadPlacesWithLimit(limit int) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		places, err := m.db.ListPlaces(limit, 0)
		if err != nil {
			return errMsg{err}
		}
		return placesLoadedMsg{places}
	})
}

func (m ReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case placesLoadedMsg:
		m.places = msg.places
		if m.cursor >= len(m.places) {
			m.cursor = 0
		}
		if len(m.places) > 0 {
			m.current = m.places[m.cursor]
		}

	case errMsg:
		m.message = fmt.Sprintf("Error: %v", msg.err)

	case saveSuccessMsg:
		m.message = "Saved successfully"

	case deleteSuccessMsg:
		m.message = "Place deleted"
		m.mode = ReviewModeList
		return m, m.loadPlaces()

	case tea.KeyMsg:
		switch m.mode {
		case ReviewModeList:
			return m.updateList(msg)
		case ReviewModeDetail:
			return m.updateDetail(msg)
		case ReviewModeEdit:
			return m.updateEdit(msg)
		}
	}

	return m, nil
}

func (m ReviewModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			if len(m.places) > 0 {
				m.current = m.places[m.cursor]
			}
		}

	case "down", "j":
		if m.cursor < len(m.places)-1 {
			m.cursor++
			if len(m.places) > 0 {
				m.current = m.places[m.cursor]
			}
		}

	case "enter", " ":
		if len(m.places) > 0 {
			m.mode = ReviewModeDetail
		}

	case "/":
		// TODO: Add search mode to review
		m.message = "Search functionality available in browse mode"

	case "r":
		return m, m.loadPlaces()
	}

	return m, nil
}

func (m ReviewModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "esc", "backspace":
		m.mode = ReviewModeList

	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
			m.current = m.places[m.cursor]
		}

	case "right", "l":
		if m.cursor < len(m.places)-1 {
			m.cursor++
			m.current = m.places[m.cursor]
		}

	case "n":
		m.mode = ReviewModeEdit
		m.editField = "notes"
		m.editValue = m.current.UserNotes

	case "t":
		m.mode = ReviewModeEdit
		m.editField = "tags"
		m.editValue = strings.Join(m.current.UserTags, ", ")

	case "o":
		m.mode = ReviewModeEdit
		m.editField = "hours"
		m.editValue = m.current.Hours

	case "p":
		m.mode = ReviewModeEdit
		m.editField = "phone"
		m.editValue = m.current.Phone

	case "w":
		m.mode = ReviewModeEdit
		m.editField = "website"
		m.editValue = m.current.Website

	case "d":
		return m, m.deletePlace()
	}

	return m, nil
}

func (m ReviewModel) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.mode = ReviewModeDetail
		m.editField = ""
		m.editValue = ""

	case "enter":
		return m, m.saveEdit()

	case "backspace":
		if len(m.editValue) > 0 {
			m.editValue = m.editValue[:len(m.editValue)-1]
		}

	default:
		m.editValue += msg.String()
	}

	return m, nil
}

func (m ReviewModel) saveEdit() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		switch m.editField {
		case "notes":
			m.current.UserNotes = m.editValue
		case "tags":
			if m.editValue == "" {
				m.current.UserTags = []string{}
			} else {
				tags := strings.Split(m.editValue, ",")
				for i, tag := range tags {
					tags[i] = strings.TrimSpace(tag)
				}
				m.current.UserTags = tags
			}
		case "hours":
			m.current.Hours = m.editValue
		case "phone":
			m.current.Phone = m.editValue
		case "website":
			m.current.Website = m.editValue
		}

		err := m.db.SavePlace(m.current)
		if err != nil {
			return errMsg{err}
		}

		m.mode = ReviewModeDetail
		m.editField = ""
		m.editValue = ""
		return saveSuccessMsg{}
	})
}

func (m ReviewModel) deletePlace() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.current == nil {
			return errMsg{fmt.Errorf("no place selected")}
		}

		err := m.db.DeletePlace(m.current.ID)
		if err != nil {
			return errMsg{err}
		}

		return deleteSuccessMsg{}
	})
}

type saveSuccessMsg struct{}
type deleteSuccessMsg struct{}

func (m ReviewModel) View() string {
	switch m.mode {
	case ReviewModeList:
		return m.viewList()
	case ReviewModeDetail:
		return m.viewDetail()
	case ReviewModeEdit:
		return m.viewEdit()
	default:
		return "Unknown mode"
	}
}

func (m ReviewModel) viewList() string {
	var b strings.Builder

	title := detailTitleStyle.Render(fmt.Sprintf("placeli review (%d places)", len(m.places)))
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.message != "" {
		b.WriteString(fmt.Sprintf("🔔 %s\n\n", m.message))
	}

	if len(m.places) == 0 {
		b.WriteString("No places found. Use 'placeli import' to add places.\n")
	} else {
		for i, place := range m.places {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			line := fmt.Sprintf("%s %s", cursor, place.Name)
			if place.Address != "" {
				line += fmt.Sprintf("\n   📍 %s", place.Address)
			}

			if i == m.cursor {
				b.WriteString(selectedItemStyle.Render(line))
			} else {
				b.WriteString(itemStyle.Render(line))
			}
			b.WriteString("\n")
		}
	}

	help := helpStyle.Render("↑/k up • ↓/j down • enter view details • r refresh • q quit")
	b.WriteString(fmt.Sprintf("\n%s", help))

	return b.String()
}

func (m ReviewModel) viewDetail() string {
	if m.current == nil {
		return "No place selected"
	}

	var b strings.Builder

	title := detailTitleStyle.Render(fmt.Sprintf("Review: %s", m.current.Name))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Main details
	content := fmt.Sprintf("%s %s\n", fieldStyle.Render("Name:"), valueStyle.Render(m.current.Name))

	if m.current.Address != "" {
		content += fmt.Sprintf("%s %s\n", fieldStyle.Render("Address:"), valueStyle.Render(m.current.Address))
	}

	if m.current.Rating > 0 {
		stars := strings.Repeat("⭐", int(m.current.Rating))
		content += fmt.Sprintf("%s %s %.1f (%d reviews)\n",
			fieldStyle.Render("Rating:"), stars, m.current.Rating, m.current.UserRatings)
	}

	if len(m.current.Categories) > 0 {
		content += fmt.Sprintf("%s %s\n",
			fieldStyle.Render("Categories:"), valueStyle.Render(strings.Join(m.current.Categories, ", ")))
	}

	if m.current.Phone != "" {
		content += fmt.Sprintf("%s %s\n", fieldStyle.Render("Phone:"), valueStyle.Render(m.current.Phone))
	}

	if m.current.Website != "" {
		content += fmt.Sprintf("%s %s\n", fieldStyle.Render("Website:"), valueStyle.Render(m.current.Website))
	}

	if m.current.Hours != "" {
		content += fmt.Sprintf("%s %s\n", fieldStyle.Render("Hours:"), valueStyle.Render(m.current.Hours))
	}

	// Photos
	if len(m.current.Photos) > 0 {
		content += fmt.Sprintf("%s %d photos available\n",
			fieldStyle.Render("Photos:"), len(m.current.Photos))
	}

	// Reviews
	if len(m.current.Reviews) > 0 {
		content += fmt.Sprintf("%s %d reviews\n",
			fieldStyle.Render("Reviews:"), len(m.current.Reviews))
		// Show first review if available
		if len(m.current.Reviews) > 0 {
			review := m.current.Reviews[0]
			reviewStars := strings.Repeat("⭐", review.Rating)
			reviewText := review.Text
			if len(reviewText) > 80 {
				reviewText = reviewText[:77] + "..."
			}
			content += fmt.Sprintf("  %s %s: \"%s\"\n",
				reviewStars, review.Author, reviewText)
		}
	}

	// User data
	content += "\n" + fieldStyle.Render("USER DATA") + "\n"

	if m.current.UserNotes != "" {
		content += fmt.Sprintf("%s %s\n", fieldStyle.Render("Notes:"), valueStyle.Render(m.current.UserNotes))
	} else {
		content += fmt.Sprintf("%s %s\n", fieldStyle.Render("Notes:"), valueStyle.Render("(none)"))
	}

	if len(m.current.UserTags) > 0 {
		content += fmt.Sprintf("%s %s\n",
			fieldStyle.Render("Tags:"), valueStyle.Render(strings.Join(m.current.UserTags, ", ")))
	} else {
		content += fmt.Sprintf("%s %s\n", fieldStyle.Render("Tags:"), valueStyle.Render("(none)"))
	}

	box := detailBoxStyle.Render(content)
	b.WriteString(box)

	b.WriteString("\n\n")
	help := helpStyle.Render("← → navigate • [n]otes • [t]ags • h[o]urs • [p]hone • [w]ebsite • [d]elete • esc back • q quit")
	b.WriteString(help)

	return b.String()
}

func (m ReviewModel) viewEdit() string {
	var b strings.Builder

	title := detailTitleStyle.Render(fmt.Sprintf("Edit %s: %s", m.editField, m.current.Name))
	b.WriteString(title)
	b.WriteString("\n\n")

	prompt := fmt.Sprintf("Enter %s:", m.editField)
	if m.editField == "tags" {
		prompt += " (comma-separated)"
	}

	content := fmt.Sprintf("%s\n\n%s", prompt, m.editValue+"_")
	box := detailBoxStyle.Render(content)
	b.WriteString(box)

	b.WriteString("\n\n")
	help := helpStyle.Render("enter save • esc cancel")
	b.WriteString(help)

	return b.String()
}

func RunReview(db *database.DB) error {
	m := NewReviewModel(db)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
