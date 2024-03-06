package footer

import (
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	text          string // The command input is a simple line of text
	focused       bool
	error, status string
	errorStyle    lipgloss.Style
	cursor        cursor.Model
}

func New() Model {
	cursor := cursor.New()
	cursor.Focus()
	return Model{
		text:       "",
		focused:    false,
		error:      "",
		status:     "",
		errorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
		cursor:     cursor,
	}
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) ShowStatus(status string) {
	m.status = status
}

func (m *Model) ShowError(err string) {
	m.error = err
}

func (m *Model) Clear() {
	m.error = ""
	m.status = ""
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			cmd = Submit(m.text)
			m.text = ""
		case tea.KeyEscape:
			cmd = Cancel
			m.text = ""
		case tea.KeyBackspace:
			if len(m.text) > 0 {
				_, size := utf8.DecodeLastRuneInString(m.text)
				m.text = m.text[:len(m.text)-size]
			}
		case tea.KeySpace:
			m.text += " "
		case tea.KeyRunes:
			cmd = nil
			m.text += string(msg.Runes)
		}
	}
	return m, cmd
}

func (m Model) View() string {
	var s string
	if m.error != "" {
		s += m.errorStyle.Render(m.error)
	} else if m.status != "" {
		s += m.status
	} else if m.focused {
		m.cursor.SetChar(" ")
		s += ":" + m.text + m.cursor.View()
	}
	return s
}
