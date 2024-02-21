package footer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SubmitMsg is a string such as "w foo.txt" containing an action
// and arguments
type SubmitMsg string

// Decode the text, returning the action and its arguments (if any)
func (m SubmitMsg) Decode() (string, []string) {
	slices := strings.Fields(string(m))
	switch len(slices) {
	case 0:
		return "", nil
	case 1:
		return slices[0], nil
	default:
		return slices[0], slices[1:]
	}
}

// CancelMsg cancels the current action.
type CancelMsg struct{}

func Submit(action string) tea.Cmd {
	return func() tea.Msg { return SubmitMsg(action) }
}

func Cancel() tea.Msg {
	return CancelMsg{}
}

type OpenBufferMsg string
