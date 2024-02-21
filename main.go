package main

import (
	"fmt"
	"moe/ui/components/footer"
	"moe/ui/components/statusbar"
	"moe/ui/components/viewport"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Let's try to create an editor that is as simple as possible.
// Goals:
//
// - simple underlying buffer implementation
// - bufferline, statusbar, footer and main view
// - no helptext
// - Helix keybindings
// - no themes

// Editor mode
type Mode int

const (
	Normal Mode = iota
	Insert
	Select
	Command
)

// Supported messages
type OpenBufferMsg string
type ModeSwitchMsg Mode
type DisplayErrorMsg string
type CloseBufferMsg string

// SwitchMode is a bubbletea command that handles mode switching
func SwitchMode(mode Mode) tea.Cmd {
	return func() tea.Msg {
		return ModeSwitchMsg(mode)
	}
}

// Model of Moe
type Model struct {
	// UI elements
	viewport  viewport.Model
	statusbar statusbar.Model
	footer    footer.Model
	// Internal data
	currentMode Mode // Current editor mode
}

func initialModel() Model {
	return Model{
		viewport:  viewport.New(),
		statusbar: statusbar.New(),
		footer:    footer.New(),
	}
}

func ErrorCmd(errorMessage string) tea.Cmd {
	return func() tea.Msg {
		return DisplayErrorMsg(errorMessage)
	}
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// Window was resized
	case tea.WindowSizeMsg:
		m.viewport, cmd = m.viewport.Update(msg)
		// m.statusbar, cmd = m.statusbar.Update(msg)
		// m.footer, cmd = m.footer.Update(msg)
		// TODO: Move in ResizeHandler() function

	case tea.KeyMsg:
		m.footer.ClearError()

		key := msg.String()
		if m.currentMode == Command { // Handle command mode
			// Pass the input to the command window
			m.footer, cmd = m.footer.Update(msg)

			// Exit command mode
			if key == "esc" {
				cmd = SwitchMode(Normal)
			}
		} else if m.currentMode == Insert { // Handle insert mode
			// Pass the input to the viewport
			m.viewport, cmd = m.viewport.Update(msg)

			// Exit insert mode
			if key == "esc" {
				cmd = SwitchMode(Normal)
			}
		} else if m.currentMode == Select { // Handle select mode
			// Exit select mode
			if key == "esc" {
				cmd = SwitchMode(Normal)
			}
		} else if m.currentMode == Normal { // Handle normal mode
			if key == ":" {
				cmd = SwitchMode(Command)
			} else if key == "i" {
				cmd = SwitchMode(Insert)
			} else if key == "v" {
				cmd = SwitchMode(Select)
			} else {
				m.viewport, cmd = m.viewport.Update(msg)
			}
		}
		// TODO. Just for now so I can quit. Remove
		if key == "ctrl+c" {
			return m, tea.Quit
		}

	// An action such as open, write, etc.
	case footer.SubmitMsg:
		cmd = tea.Batch(
			ParseAction(msg),
			SwitchMode(Normal),
		)

	case footer.CancelMsg:
		cmd = SwitchMode(Normal)

	// The command line has the following messages which need to be handled.
	// It also has some internal messages which do not need to be handled by the main loop.
	// An "open a new buffer" message was received
	case OpenBufferMsg:
		path := string(msg)
		m.viewport, cmd = m.viewport.Update(viewport.MsgOpenBuffer(path))
	// path := string(msg)
	// buffer := newBuffer(path)
	// // Set all other buffers as inactive
	// for _, b := range m.buffers {
	// 	b.active = false
	// }
	// m.buffers = append(m.buffers, buffer)
	// m.header.SetBuffers(m.buffers...)
	// m.header, _ = m.header.Update(msg)

	// case footer.CloseBufferMsg:
	// bufName := string(msg)
	// for i, b := range m.buffers {
	// 	if b.GetName() == bufName {
	// 		m.buffers = slices.Delete(m.buffers, i, i+1)
	// 	}
	// }

	// A mode switch was selected.
	case ModeSwitchMsg:
		m.currentMode = Mode(msg)

		// Depending on mode, we can do stuff, like a hook on mode change
		switch Mode(msg) {
		case Insert:
			// Do stuff, for example, enable absolute line mode in the editor
		case Normal:
			m.footer.SetVisible(false)
		case Select:
		case Command:
			m.footer.SetVisible(true)
		}

	// Handle errors
	case DisplayErrorMsg:
		m.footer.ShowError(string(msg))
	}

	return m, cmd
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewport.View(),
		m.statusbar.View(),
		m.footer.View())
	// TODO: I can enhance the experience with pop-ups which render **over** the text I got above.
}

// Tries to close all the buffers received. Called when running "bc", for example
func CloseBuffers(buffers ...string) tea.Cmd {
	var msgs tea.BatchMsg
	for _, b := range buffers {
		msgs = append(msgs, func() tea.Msg { return CloseBufferMsg(b) })
	}
	return tea.Batch(msgs...)
}

type ActionInterface interface {
	Decode() (string, []string)
}

// ParseAction parses the given user command and arguments and does stuff.
func ParseAction(action ActionInterface) tea.Cmd {
	command, arguments := action.Decode()

	switch command {
	case "o", "open":
		if arguments == nil {
			return ErrorCmd("Please specify a path to open.")
		}
		return func() tea.Msg { return OpenBufferMsg(arguments[0]) }
	case "q", "quit":
		if arguments != nil {
			return ErrorCmd(":quit takes no arguments.")
		}
		return tea.Quit
	case "bc", "buffer-close":
		// Can close multiple buffers by just specifying the buffer name
		return CloseBuffers(arguments...)
	}

	return nil
}

func main() {
	var debugFile string
	if len(os.Getenv("DEBUG")) > 0 {
		debugFile = "debug.log"
	} else {
		debugFile = "/dev/null"
	}
	// Log to file. Either /dev/null or a real file
	f, err := tea.LogToFile(debugFile, "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	// Start Bubbletea
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
