package buffer

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Ardelean-Calin/elmo/pkg/gapbuffer"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Contains the new cursor coordinates
type UpdateViewportMsg string
type LineChangedMsg int

// Model represents an opened file.
type Model struct {
	Path     string                    // Absolute path on disk.
	fd       *os.File                  // File descriptor.
	GapBuf   gapbuffer.GapBuffer[rune] // Actual raw text data. Gap Buffer is a nice compromise between Piece Chain and buffer.
	Lines    gapbuffer.GapBuffer[int]  // The line numbers are also stored in a Gap Buffer
	Focused  bool
	modified bool // Content was modified and not saved to disk
	// Used just once on load
	ready bool
	//  Then, the cursor will be strictly for display only (see footer.go)
	Cursor    cursor.Model // Cursor model
	CursorPos int          // Current cursor position
	// Horizontal position of the cursor within the line
	// A move down or up will try to keep this position
	CursorPosH int
	viewport   viewport.Model // Scrollable viewport
}

func New() Model {
	return Model{
		Path:       "",
		fd:         nil,
		GapBuf:     gapbuffer.NewGapBuffer[rune](),
		Lines:      gapbuffer.NewGapBuffer[int](),
		Focused:    true,
		modified:   false,
		ready:      false,
		Cursor:     cursor.New(),
		CursorPos:  0,
		CursorPosH: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// On Resize, re-render the viewport
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-2)
			m.viewport.YPosition = 0
			m.ready = true
			// Clear the default keybinds for up/down using k/j
			m.viewport.KeyMap.Down = key.NewBinding()
			m.viewport.KeyMap.Up = key.NewBinding()
		} else {
			m.viewport.Height = msg.Height - 2
			m.viewport.Width = msg.Width
		}
		// Render the buffer content
		cmds = append(cmds, Render(&m))

	case tea.KeyMsg:
		// TODO: Normal mode, insert mode, etc.
		if msg.String() == "j" {
			cmd = CursorDown(&m, 1)
		}
		if msg.String() == "k" {
			cmd = CursorUp(&m, 1)
		}
		if msg.String() == "l" {
			cmd = CursorRight(&m, 1)
		}
		if msg.String() == "h" {
			cmd = CursorLeft(&m, 1)
		}
		cmds = append(cmds, cmd)

	case tea.MouseMsg:
		// TODO react to mouse clicks by moving the cursor

	// Update the viewport content
	case UpdateViewportMsg:
		content := string(msg)
		m.viewport.SetContent(content)

	// The current line has changed
	case LineChangedMsg:
		// Calculate viewport boundaries
		viewStart := m.viewport.YOffset + 5
		viewEnd := m.viewport.YOffset + m.viewport.VisibleLineCount() - 5

		if m.Lines.Cursor() < viewStart {
			m.viewport.SetYOffset(m.Lines.Cursor() - 5 + 1)
		} else if m.Lines.Cursor() > viewEnd {
			m.viewport.SetYOffset(m.Lines.Cursor() - m.viewport.VisibleLineCount() + 5)
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Show the cursor
	m.Cursor, cmd = m.Cursor.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.viewport.View()
}

// OpenFile opens the given file inside the buffer
func (m *Model) OpenFile(path string) (tea.Cmd, error) {
	var bytes []byte

	fd, err := os.OpenFile(path, os.O_RDWR, 0664) // taken from helix
	if err == nil {
		// File exists
		bytes, err = io.ReadAll(fd)
		if err != nil {
			// Some weird error happened. Display it.
			return nil, err
		}
	}

	// TODO
	// Temporary. Until I fix rendering tabs and horizontal positions
	text := string(bytes)
	text = strings.ReplaceAll(text, "\t", "    ")

	// Ok by this point I either have a fd with some bytes or a nil fd and nil bytes
	// Create a gap buffer with the contents of the file
	content := []rune(text)
	buf := gapbuffer.NewGapBuffer[rune]()
	buf.SetContent(content)
	// And create a gap buffer with all the newline indices. This way I can simply
	// index the line as line[n] and get the index inside the gap buffer where the n-th line
	// starts.
	indices := []int{-1}
	indices = append(indices, buf.FindAll('\n')...)
	// Increment the indices so that they point to the line starts
	for i := range indices {
		indices[i]++
	}
	lineBuf := gapbuffer.NewGapBuffer[int]()
	lineBuf.SetContent(indices)

	m.Path = path
	m.fd = fd
	m.GapBuf = buf
	m.Lines = lineBuf
	m.modified = false

	return Render(m), nil
}

// Name returns the title of the buffer window to display
func (b Model) Name() string {
	_, name := path.Split(b.Path)
	return name
}

func LineChanged() tea.Cmd {
	return func() tea.Msg {
		return LineChangedMsg(0)
	}
}

func CursorDown(m *Model, n int) tea.Cmd {
	// Going down will move the cursor until the next line *plus* the horizontal cursor position
	// Cursor pos needs to be min(pos + hpos, lineLength)
	m.Lines.CursorRight()
	m.CursorPos = m.Lines.Current()
	m.CursorPos = min(m.CursorPos+m.CursorPosH, m.Lines.Next()-1)

	return tea.Batch(LineChanged(), Render(m))
}

func CursorUp(m *Model, n int) tea.Cmd {
	m.Lines.CursorLeft()
	m.CursorPos = m.Lines.Current()
	m.CursorPos = min(m.CursorPos+m.CursorPosH, m.Lines.Next()-1)

	return tea.Batch(LineChanged(), Render(m))
}

func CursorLeft(m *Model, n int) tea.Cmd {
	var cmds []tea.Cmd

	m.CursorPos = clamp(m.CursorPos-1, 0, m.GapBuf.Count())
	// Going left got us on a new line
	if m.CursorPos < m.Lines.Current() {
		m.Lines.CursorLeft()
		m.CursorPosH = m.CursorPos - m.Lines.Current()
		cmds = append(cmds, LineChanged())
	} else {
		m.CursorPosH = m.CursorPos - m.Lines.Current()
	}

	cmds = append(cmds, Render(m))
	return tea.Batch(cmds...)
}

func CursorRight(m *Model, n int) tea.Cmd {
	var cmds []tea.Cmd

	m.CursorPos = clamp(m.CursorPos+1, 0, m.GapBuf.Count())
	// Going right got us to a new line
	if m.CursorPos >= m.Lines.Next() {
		m.Lines.CursorRight()
		m.CursorPosH = 0
		cmds = append(cmds, LineChanged())
	} else {
		m.CursorPosH = m.CursorPos - m.Lines.Current()
	}

	cmds = append(cmds, Render(m))
	return tea.Batch(cmds...)
}

// Render is the command which renders our viewpoint content to screen
func Render(m *Model) tea.Cmd {
	var sb strings.Builder
	sty := lipgloss.NewStyle().Width(m.viewport.Width)

	for lineNo, bufIndex := range m.Lines.Collect() {
		var lineBuilder strings.Builder

		// Highlight the current line
		sty.UnsetBackground()
		lineNoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3c3d4f"))
		if lineNo == m.Lines.GapStart {
			sty = sty.Background(lipgloss.Color("#2a2b3c"))
			lineNoStyle = lineNoStyle.Foreground(lipgloss.Color("#878ebf"))
		}

		// Write line numbers
		lineBuilder.WriteString(lineNoStyle.Render(fmt.Sprintf("%4d  ", lineNo+1)))

		// Now we render the text, then the remaining
		for {
			done := false
			r := m.GapBuf.GetAbs(bufIndex)
			if r == '\n' {
				// Always render a space instead of a newline
				r = ' '
				done = true
			}

			// Render cursor
			if bufIndex == m.CursorPos {
				m.Cursor.Focus()
				m.Cursor.SetChar(string(r))

				lineBuilder.WriteString(m.Cursor.View())
			} else {
				textStyle := sty.Copy().UnsetWidth()
				lineBuilder.WriteString(textStyle.Render(string(r)))
			}
			bufIndex++
			if done {
				break
			}
		}

		sb.WriteString(sty.Render(lineBuilder.String()))
		sb.WriteRune('\n')
	}

	return func() tea.Msg {
		return UpdateViewportMsg(sb.String())
	}
}

// clamp limits the value of val between [low, high)
func clamp(val, low, high int) int {
	return max(low, min(val, high-1))
}
