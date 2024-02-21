package textarea

import (
	"moe/pkg/buffer"

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
	buffers       buffer.LinkedList  // Linked list containing all buffers as displayed in the bufferline
	activeNode    *buffer.BufferNode // Currently active buffer
	Focused       bool               // If focused, we react to events
	Height, Width int                // Size of the textarea
	Viewport      viewport.Model     // Scrollable viewport
}

func New() Model {
	return Model{
		buffers:    buffer.NewList(),
		activeNode: nil,
		Focused:    false,
	}
}

// CurrentBuffer returns the path of the currently active buffer
func (m *Model) CurrentBuffer() string {
	if m.activeNode != nil {
		return m.activeNode.Buffer.Path
	}
	return ""
}

// SwitchBuffer tries to switch to the given buffer. Returns false if buffer doesn't exist
func (m *Model) SwitchBuffer(path string) bool {
	for node := range m.buffers.Iter() {
		if node.Buffer.Path == path {
			m.activeNode = node
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
		m.buffers.AddNode(n)
		m.activeNode = n
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
	// NOTE: needs GOEXPERIMENT=rangefunc as of Go 1.22
	for node := range m.buffers.Iter() {
		style := lipgloss.NewStyle().Padding(0, 1)
		if node == m.activeNode {
			style = style.Reverse(true)
		}
		bufferline += style.
			Render(node.Buffer.Name())
	}

	if m.activeNode != nil {
		m.Viewport.SetContent(m.activeNode.Buffer.String())
	} else {
		m.Viewport.SetContent("")
	}
	return lipgloss.JoinVertical(lipgloss.Left, bufferline, m.Viewport.View())
}
