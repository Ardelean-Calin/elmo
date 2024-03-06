package textarea

import (
	"strings"
	// "strings"

	"github.com/Ardelean-Calin/moe/pkg/buffer"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* These two will be displayed in the status bar */
type StatusMsg string
type ErrorMsg error

type BufSwitchedMsg string
type LineChangedMsg int

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
	Buffers       buffer.LinkedList // Linked list containing all buffers as displayed in the bufferline
	CurBuf        *buffer.Buffer    // Currently active buffer
	Focused       bool              // If focused, we react to events
	Height, Width int               // Size of the textarea
	Viewport      viewport.Model    // Scrollable viewport
}

func New() Model {
	return Model{
		Buffers: buffer.NewList(),
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
	iterator := m.Buffers.Iter()
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
		m.Buffers.AddNode(n)
		m.CurBuf = b
	}

	// We already had the buffer opened and focused it.
	return Event(BufSwitchedMsg(path))
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height - 3
	case tea.KeyMsg:
		if msg.String() == "j" {
			m.CurBuf.CursorDown()
		}
		if msg.String() == "k" {
			m.CurBuf.CursorUp()
		}
		if msg.String() == "l" {
			m.CurBuf.CursorRight()
		}
		if msg.String() == "h" {
			m.CurBuf.CursorLeft()
		}
	// case buffer.BufferUpdateMsg:
	// A buffer's content has been updated...
	//     m.Viewport.SetContent(buf.View())
	default:
		if !m.Focused {
			return m, nil
		}
	}

	// TODO. Handle any other events.
	// Handle keyboard and mouse events in the viewport
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// func cursorToAbs(c cursor.Model) int {
// 	c.Row
// }

func (m Model) View() string {
	// Render the buffer line
	bufferline := ""

	iterator := m.Buffers.Iter()
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

		iterator := m.CurBuf.Val.Iter()
		for iterator.HasNext() {
			index, r := iterator.Next()
			if index == m.CurBuf.Cursor.Pos {
				sb.WriteString(m.CurBuf.Cursor.View())
				if r == '\n' {
					sb.WriteRune('\n')
				}
			} else {
				sb.WriteRune(r)
			}
		}
		// I don't like this approach, since I always render a full screen
		// and have to keep it into memory. I could use the newly defined
		// lineBuf
		m.Viewport.SetContent(sb.String())
	} else {
		m.Viewport.SetContent("")
	}
	return lipgloss.JoinVertical(lipgloss.Left, bufferline, m.Viewport.View())
}
