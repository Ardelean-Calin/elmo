package header

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BufferInterface describes the required functions for displaying
// the active open files
type BufferInterface interface {
	GetName() string
	IsActive() bool
}

type buffer struct {
	BufferInterface
}

type Model struct {
	openedBuffers []buffer
	activeStyle   lipgloss.Style
	inactiveStyle lipgloss.Style
}

func New() Model {
	inactiveStyle := lipgloss.NewStyle().Padding(0, 1).AlignHorizontal(lipgloss.Left)
	activeStyle := inactiveStyle.Copy().Reverse(true)

	return Model{
		openedBuffers: []buffer{},
		activeStyle:   activeStyle,
		inactiveStyle: inactiveStyle,
	}
}

func (m *Model) SetBuffers(bufs ...BufferInterface) {
	var bs []buffer
	for _, b := range bufs {
		bs = append(bs, buffer{b})
	}
	m.openedBuffers = bs
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.inactiveStyle = m.inactiveStyle.MaxWidth(msg.Width)
		m.activeStyle = m.activeStyle.MaxWidth(msg.Width)
	}
	return m, nil
}

func (m Model) View() string {
	var bufferNames []string
	if len(m.openedBuffers) == 0 {
		// Display scratch buffer
		return m.activeStyle.Render("[scratch]")
	}
	for _, b := range m.openedBuffers {
		var s string
		name := b.GetName()
		if b.IsActive() {
			s = m.activeStyle.Render(name)
		} else {
			s = m.inactiveStyle.Render(name)
		}
		bufferNames = append(bufferNames, s)
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, bufferNames...)
}
