package buffer

import tea "github.com/charmbracelet/bubbletea"

func ClearSelection() tea.Msg {
	return ClearSelectionMsg(true)
}
