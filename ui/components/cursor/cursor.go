package cursor

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Gap Buffer interface needed for cursor navigation
type GapInterface interface {
	GetAbs(pos int) rune
	Count() int
}

// Model is the Bubble Tea model for this cursor element.
type Model struct {
	// Style for styling the cursor block.
	Style lipgloss.Style
	// TextStyle is the style used for the cursor when it is hidden (when blinking).
	// I.e. displaying normal text.
	TextStyle lipgloss.Style

	// Underlying text buffer on which we navigate
	buf GapInterface
	// char is the character under the cursor
	Char string
	// Position of the cursor inside the raw text
	Pos int
}

func (m *Model) Goto(pos int) {
	m.Pos = pos
	m.Char = string(m.buf.GetAbs(m.Pos))
}

func (m *Model) Left() {
	m.Pos = max(0, m.Pos-1)
	m.Char = string(m.buf.GetAbs(m.Pos))
}

func (m *Model) Right() {
	m.Pos = min(m.buf.Count(), m.Pos+1)
	m.Char = string(m.buf.GetAbs(m.Pos))
}

// New creates a new cursor bound to the given Gap Buffer
func New(bufPtr GapInterface) Model {
	return Model{
		buf:  bufPtr,
		Char: string(bufPtr.GetAbs(0)),
		Pos:  0,
	}
}

// Update updates the cursor.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View displays the cursor.
func (m Model) View() string {
	style := m.Style.Inline(false).Reverse(true)
	if m.Char == "\n" {
		return style.Width(1).Render("")
	} else {
		return style.Render(m.Char)
	}

}
