package commands

import (
	"github.com/Ardelean-Calin/elmo/messages"
	tea "github.com/charmbracelet/bubbletea"
)

func ShowError(err error) tea.Cmd {
	return func() tea.Msg {
		return messages.ShowErrorMsg(err.Error())
	}
}

func ShowStatus(status string) tea.Cmd {
	return func() tea.Msg {
		return messages.ShowStatusMsg(status)
	}
}
