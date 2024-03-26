package main

import (
	"fmt"
	"os"

	"github.com/Ardelean-Calin/elmo/ui/components/footer"
	"github.com/Ardelean-Calin/elmo/ui/components/statusbar"
	"github.com/Ardelean-Calin/elmo/ui/components/textarea"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Let's try to create an editor that is as simple as possible.
// Goals:
//
// - simple underlying buffer implementation      [x]
// - bufferline, statusbar, footer and main view
// - no helptext
// - Helix keybindings
// - no themes - actually, some themes maybe :D

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
	textarea  textarea.Model
	statusbar statusbar.Model
	footer    footer.Model // Command bar + error and status messages
	// Internal data
	currentMode Mode // Current editor mode
}

func initialModel() Model {
	return Model{
		textarea:    textarea.New(),
		statusbar:   statusbar.New(),
		footer:      footer.New(),
		currentMode: Normal,
	}
}

func OpenBufferCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return OpenBufferMsg(path)
	}
}

func (m Model) Init() tea.Cmd {
	if len(os.Args) > 1 {
		return OpenBufferCmd(os.Args[1])
	}
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// Window was resized
	case tea.WindowSizeMsg:
		m.statusbar.Width = msg.Width

	case tea.KeyMsg:
		m.footer.Clear()

		key := msg.String()
		if m.currentMode == Command { // Handle command mode
			// Exit command mode
			if key == "esc" {
				cmd = SwitchMode(Normal)
			}
		} else if m.currentMode == Insert { // Handle insert mode
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
				// m.viewport, cmd = m.viewport.Update(msg)
			}
		}
		// TODO. Just for now so I can quit. Remove
		if key == "ctrl+c" {
			return m, tea.Quit
		}

	// An action such as open, write, etc.
	// We process the action and switch the mode
	case footer.SubmitMsg:
		action := msg
		command, arguments := action.Decode()

		switch command {
		case "o", "open":
			if arguments == nil {
				cmd = footer.ShowError(fmt.Errorf("Please specify a path to open."))
			}
			cmd = func() tea.Msg { return OpenBufferMsg(arguments[0]) }
		case "q", "quit":
			if arguments != nil {
				cmd = footer.ShowError(fmt.Errorf("'quit' takes no arguments."))
			}
			cmd = tea.Quit
		case "bc", "buffer-close":
			// Can close multiple buffers by just specifying the buffer name
			cmd = CloseBuffers(arguments...)
		case "w", "write":
			cmd = m.textarea.Buffer.WriteToDisk()
		default:
			cmd = footer.ShowError(fmt.Errorf("Unrecognized command: '%s'", command))
		}

		cmds = append(cmds, cmd, SwitchMode(Normal))

	case footer.CancelMsg:
		cmd = SwitchMode(Normal)

	// Switched to a new buffer
	case textarea.BufSwitchedMsg:
		m.statusbar.SetOpenBuffer(m.textarea.CurBufPath())

	// An "open a new buffer" message was received
	case OpenBufferMsg:
		path := string(msg)
		cmd = m.textarea.OpenBuffer(path)

	// A mode switch was selected.
	case ModeSwitchMsg:
		m.currentMode = Mode(msg)

		// Depending on mode, we can do stuff, like a hook on mode change
		switch Mode(msg) {
		case Insert:
			// Do stuff, for example, enable absolute line mode in the editor
			m.textarea.Focused = true
			m.statusbar.InsertMode()
			m.footer.Blur()
		case Normal:
			m.footer.Blur()
			m.statusbar.NormalMode()
			m.textarea.Focused = false
		case Select:
			m.statusbar.SelectMode()
			m.footer.Blur()
		case Command:
			m.footer.Focus()
		}

	}
	cmds = append(cmds, cmd)

	// Send all events to each of the components. If they are focused they might react.
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	m.footer, cmd = m.footer.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.textarea.View(),
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
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
