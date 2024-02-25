package statusbar

import (
	"github.com/Ardelean-Calin/moe/pkg/common"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode string

const (
	Normal Mode = "NOR"
	Insert      = "INS"
	Select      = "SEL"
)

type Model struct {
	mode       Mode
	bufferPath string
	Width      int
}

func New() Model {
	return Model{
		mode:       Normal,
		bufferPath: "",
	}
}

func (m *Model) NormalMode() {
	m.mode = Normal
}

func (m *Model) InsertMode() {
	m.mode = Insert
}

func (m *Model) SelectMode() {
	m.mode = Select
}

func (m *Model) SetOpenBuffer(path string) {
	m.bufferPath = path
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	modeString := lipgloss.NewStyle().
		Padding(0, 1).
		Render(string(m.mode))
	infoString := lipgloss.NewStyle().
		Padding(0, 1).
		Render("LF  go")

	// Center the buffer string.
	bufferString := lipgloss.PlaceHorizontal(m.Width, lipgloss.Center, m.bufferPath)
	start := common.Clamp(lipgloss.Width(modeString), 0, m.Width)
	bufferString = bufferString[start:]
	stop := common.Clamp(len(bufferString)-lipgloss.Width(infoString), 0, m.Width)
	bufferString = bufferString[:stop]

	s := modeString + bufferString + infoString
	return s

	// return lipgloss.JoinHorizontal(lipgloss.Left, modeString, bufferString, infoString)

	// return lipgloss.NewStyle().Width(m.Width).Reverse(true).Render("")
	// return lipgloss.NewStyle().Width(m.Width).Render(lipgloss.JoinHorizontal(lipgloss.Left,
	// 	lipgloss.NewStyle().
	// 		Padding(0, 1).
	// 		Render(string(m.mode)),
	// 	lipgloss.NewStyle().
	// 		AlignHorizontal(lipgloss.Center).
	// 		Render(name),
	// ),
	// )
}
