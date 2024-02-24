package cursor

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the Bubble Tea model for this cursor element.
type Model struct {
	// Style for styling the cursor block.
	Style lipgloss.Style
	// TextStyle is the style used for the cursor when it is hidden (when blinking).
	// I.e. displaying normal text.
	TextStyle lipgloss.Style

	// char is the character under the cursor
	Char string
	// Position of the cursor
	Row, Col int
}

func (m *Model) Up() {
	m.Row = max(0, m.Row-1)
}

func (m *Model) Down() {
	m.Row++
}

func (m *Model) Left() {
	m.Col = max(0, m.Col-1)
}

func (m *Model) Right() {
	m.Col++
}

// New creates a new model with default settings.
func New() Model {
	return Model{
		Row: 0,
		Col: 0,
	}
}

// Update updates the cursor.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	return m, nil
}

// View displays the cursor.
func (m Model) View() string {
	return m.Style.Inline(true).Reverse(true).Render(m.Char)
}
