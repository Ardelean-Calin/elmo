package textarea

import (
	"moe/pkg/buffer"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* These two will be displayed in the status bar */
type StatusMsg string
type ErrorMsg error

type EvtBufferSwitched string

// Return a message as an event
func Event(T any) tea.Cmd {
	return func() tea.Msg {
		return T
	}
}

// Return a status string. Will be displayed in the statusbar
func StatusCmd(status string) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(status)
	}
}

// Return an error. Will be displayed in the statusbar
func ErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg(err)
	}
}

type Model struct {
	bufList       buffer.LinkedList // Linked list containing all buffers as displayed in the bufferline
	CurBuf        *buffer.Buffer    // Currently active buffer
	Focused       bool              // If focused, we react to events
	Height, Width int               // Size of the textarea
	Viewport      viewport.Model    // Scrollable viewport
}

func New() Model {
	return Model{
		bufList: buffer.NewList(),
		CurBuf:  nil,
		Focused: false,
	}
}

// CurBufPath returns the path of the currently active buffer
func (m *Model) CurBufPath() string {
	if m.CurBuf != nil {
		return m.CurBuf.Path
	}
	return ""
}

// SwitchBuffer tries to switch to the given buffer. Returns false if buffer doesn't exist
func (m *Model) SwitchBuffer(path string) bool {
	iterator := m.bufList.Iter()
	for iterator.HasNext() {
		node := iterator.Next()
		if node.Buffer.Path == path {
			m.CurBuf = node.Buffer
			return true
		}
	}

	return false
}

// OpenBuffer opens a new buffer for editing. If the buffer is already
// opened in one of our tabs, we just switch to the tab.
func (m *Model) OpenBuffer(path string) tea.Cmd {
	// Check if buffer already is open and focus it
	if !m.SwitchBuffer(path) {
		// Buffer wasn't found. Create a new buffer
		b, err := buffer.NewBuffer(path)
		if err != nil {
			return ErrorCmd(err)
		}
		n := buffer.Node(b)
		m.bufList.AddNode(n)
		m.CurBuf = b
	}

	// We already had the buffer opened and focused it.
	return Event(EvtBufferSwitched(path))
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height - 3
	case tea.KeyMsg:
		if msg.String() == "l" {
			m.CurBuf.CursorRight()
		}
		if msg.String() == "h" {
			m.CurBuf.CursorLeft()
		}
	default:
		if !m.Focused {
			return m, nil
		}
	}

	// TODO. Handle any other events.

	return m, nil
}

func (m Model) View() string {
	bufferline := ""

	iterator := m.bufList.Iter()
	for iterator.HasNext() {
		node := iterator.Next()
		style := lipgloss.NewStyle().Padding(0, 1)
		if node.Buffer == m.CurBuf {
			style = style.Reverse(true)
		}
		bufferline += style.
			Render(node.Buffer.Name())
	}

	if m.CurBuf != nil {
		// Render the contents to screen, as well as the cursor
		var sb strings.Builder
		var runes []rune = []rune(m.CurBuf.String())
		for pos, r := range runes {
			if pos == m.CurBuf.Cursor.Pos {
				m.CurBuf.Cursor.Char = string(r)
				sb.WriteString(m.CurBuf.Cursor.View())
			} else {
				sb.WriteRune(r)
			}
		}
		m.Viewport.SetContent(sb.String())
	} else {
		m.Viewport.SetContent("")
	}
	return lipgloss.JoinVertical(lipgloss.Left, bufferline, m.Viewport.View())
}
