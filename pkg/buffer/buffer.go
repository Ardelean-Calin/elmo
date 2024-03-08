package buffer

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Ardelean-Calin/moe/pkg/gapbuffer"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Contains the new cursor coordinates
type CursorMoveMsg int
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
	// TODO: Replace cursor with bubbletea cursor.
	//  Then, the cursor will be strictly for display only (see footer.go)
	Cursor           cursor.Model // Cursor model
	CursorPos        int          // Current cursor position
	CurrentLineIndex int
	viewport         viewport.Model // Scrollable viewport
}

func New() Model {
	return Model{
		Path:             "",
		fd:               nil,
		GapBuf:           gapbuffer.NewGapBuffer[rune](),
		Lines:            gapbuffer.NewGapBuffer[int](),
		Focused:          true,
		modified:         false,
		ready:            false,
		Cursor:           cursor.New(),
		CursorPos:        0,
		CurrentLineIndex: 0,
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
			// Render the buffer content
			cmds = append(cmds, Render(&m))
		} else {
			m.viewport.Height = msg.Height - 2
			m.viewport.Width = msg.Width
		}

	case tea.KeyMsg:
		// TODO: Normal mode, insert mode, etc.
		if msg.String() == "j" {
			cmd = CursorDown(&m, 1)
			// TODO. I can move the viewport only here if I am in normal mode.
			// m.viewport.LineDown(1)
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

	// Cursor has moved. We regenerate the content of the viewport
	case CursorMoveMsg:
		pos := int(msg)
		m.CursorPos = pos
		cmds = append(cmds, Render(&m))

	// The current line has changed
	case LineChangedMsg:
		m.CurrentLineIndex = int(msg)
		// Calculate viewport boundaries
		viewStart := m.viewport.YOffset + 5
		viewEnd := m.viewport.YOffset + m.viewport.Height - 5

		if m.CurrentLineIndex < viewStart {
			m.viewport.SetYOffset(m.CurrentLineIndex - 5 + 1)
		} else if m.CurrentLineIndex > viewEnd {
			m.viewport.SetYOffset(m.CurrentLineIndex - m.viewport.Height + 5)
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
	// The buffer View method also needs to render the cursor
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

	// Ok by this point I either have a fd with some bytes or a nil fd and nil bytes
	// Create a gap buffer with the contents of the file
	content := []rune(string(bytes))
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

func CursorDown(m *Model, n int) tea.Cmd {
	m.Lines.CursorRight()
	line := m.Lines.Pos()

	index := m.Lines.GetAbs(line)
	return tea.Batch(CursorGoto(index), ChangeLine(line))
}

func CursorUp(m *Model, n int) tea.Cmd {
	m.Lines.CursorLeft()
	line := m.Lines.Pos()

	index := m.Lines.GetAbs(line)
	return tea.Batch(CursorGoto(index), ChangeLine(line))
}

func CursorLeft(m *Model, n int) tea.Cmd {
	if m.CursorPos == 0 {
		return nil
	}

	return CursorGoto(m.CursorPos - n)
}

func CursorRight(m *Model, n int) tea.Cmd {
	// TODO check bounds
	return CursorGoto(m.CursorPos + n)
}

func CursorGoto(pos int) tea.Cmd {
	return func() tea.Msg { return CursorMoveMsg(pos) }
}

// Render is the command which renders our viewpoint content to screen
func Render(m *Model) tea.Cmd {
	var sb strings.Builder
	runes := m.GapBuf.Collect()

	for i, r := range runes {
		// Render the line number if the current character is a newline
		lineNo, ok := m.Lines.Find(i)
		if ok == true {
			sb.WriteString(fmt.Sprintf("%4d  ", lineNo+1))
		}

		if i == m.CursorPos {
			m.Cursor.Focus()
			// Newline, just render an empty cursor
			if r == 10 {
				m.Cursor.SetChar(" ")
				sb.WriteString(m.Cursor.View())
				sb.WriteRune(r)
			} else {
				m.Cursor.SetChar(string(r))
				sb.WriteString(m.Cursor.View())
			}
		} else {
			sb.WriteRune(r)
		}
	}
	content := sb.String()

	return func() tea.Msg {
		return UpdateViewportMsg(content)
	}
}

func ChangeLine(i int) tea.Cmd {
	return func() tea.Msg {
		return LineChangedMsg(i)
	}
}
