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
	// pos is the absolute position of the cursor in the string
	Pos int
}

// New creates a new model with default settings.
func New() Model {
	return Model{
		Pos: 0,
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
