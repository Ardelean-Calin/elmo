package footer

import (
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	text       string // The command input is a simple line of text
	visible    bool
	error      string
	errorStyle lipgloss.Style
}

func New() Model {
	return Model{
		text:       "",
		visible:    false,
		error:      "",
		errorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")),
	}
}

func (m *Model) SetVisible(v bool) {
	m.visible = v
}

func (m *Model) ShowError(err string) {
	m.error = err
}

func (m *Model) ClearError() {
	m.error = ""
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
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
	} else if m.visible {
		s += ":" + m.text
	}
	return s
}
