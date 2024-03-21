package textarea

import (
	"github.com/Ardelean-Calin/elmo/pkg/buffer"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* These two will be displayed in the status bar */
type StatusMsg string
type ErrorMsg error

type BufSwitchedMsg string

// Return a message as an event
func Event(T any) tea.Cmd {
	return func() tea.Msg {
		return T
	}
}

type Model struct {
	Height, Width int          // Size of the textarea
	Focused       bool         // If focused, we react to events
	Buffer        buffer.Model // Currently displayed buffer
}

func New() Model {
	return Model{
		Buffer:  buffer.New(),
		Focused: false,
	}
}

// CurBufPath returns the path of the currently active buffer
func (m *Model) CurBufPath() string {
	return m.Buffer.Path
}

// OpenBuffer opens a new buffer for editing. If the buffer is already
// opened in one of our tabs, we just switch to the tab.
func (m *Model) OpenBuffer(path string) tea.Cmd {
	cmd := m.Buffer.OpenFile(path)

	// Notify that a new buffer has been opened.
	return tea.Batch(cmd, Event(BufSwitchedMsg(path)))
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height-2
	}
	m.Buffer, cmd = m.Buffer.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	var bufferContent string

	bufferContent = lipgloss.NewStyle().Height(m.Height).Render(m.Buffer.View())

	return bufferContent
}
