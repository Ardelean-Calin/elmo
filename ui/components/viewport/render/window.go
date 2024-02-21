package render

import (
	"log"
	"moe/pkg/common"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents a rectangular window "sliding" over our text.
// Let's keep it as simple as possible, it is a window where only the height can be changed.
// Only the text that is inside the window gets rendered on screen.
type Model struct {
	StartRow int
	Height   int
	Content  string
}

func NewRenderWindow() Model {
	return Model{StartRow: 0, Height: 0, Content: ""}
}

func (m *Model) SetContent(content string) {
	m.Content = content
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update() (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	lines := strings.Split(m.Content, "\n")
	start := common.Clamp(m.StartRow, 0, len(lines))
	stop := common.Clamp(m.StartRow+m.Height, 0, len(lines))
	log.Printf("start: %d\tstop: %d", start, stop)
	linesFiltered := lines[start:stop]

	var sb strings.Builder
	for _, l := range linesFiltered {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}

	return sb.String()
}
