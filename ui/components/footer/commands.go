package footer

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Submit sumbits the currently written command
func Submit(action string) tea.Cmd {
	return func() tea.Msg { return SubmitMsg(action) }
}

// Cancel a command operation.
func Cancel() tea.Msg {
	return CancelMsg{}
}

// Show the status on the footer
func ShowStatus(msg string) tea.Cmd {
	return func() tea.Msg { return StatusMsg(msg) }
}

// Show an error on the footer
func ShowError(err error) tea.Cmd {
	return func() tea.Msg { return ErrorMsg(err.Error()) }
}
