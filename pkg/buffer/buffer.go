package buffer

import (
	"io"
	"os"
	"path"

	"github.com/Ardelean-Calin/moe/pkg/gapbuffer"
	"github.com/Ardelean-Calin/moe/ui/components/cursor"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents an opened file.
type Model struct {
	Path     string                     // Absolute path on disk.
	fd       *os.File                   // File descriptor.
	Val      *gapbuffer.GapBuffer[rune] // Actual raw text data. Gap Buffer is a nice compromise between Piece Chain and buffer.
	Lines    *gapbuffer.GapBuffer[int]  // The line numbers are also stored in a Gap Buffer
	Focused  bool
	modified bool // Content was modified and not saved to disk
	// Used just once on load
	ready bool
	// TODO: Replace cursor with bubbletea cursor.
	//  Then, the cursor will be strictly for display only (see footer.go)
	Cursor   cursor.Model   // Cursor position inside this buffer.
	viewport viewport.Model // Scrollable viewport
}

func New() Model {
	return Model{
		Path:     "",
		fd:       nil,
		Val:      nil,
		Lines:    nil,
		Focused:  true,
		modified: false,
		ready:    false,
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
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.YPosition = 0
			if m.Val != nil {
				m.viewport.SetContent(m.Val.String())
			}
			m.ready = true
		} else {
			m.viewport.Height = msg.Height
			m.viewport.Width = msg.Width
		}

	case tea.KeyMsg:
		// TODO: Normal mode, insert mode, etc.
		if msg.String() == "j" {
			m.CursorDown()
			// TODO. I can move the viewport only here if I am in normal mode.
			// m.viewport.LineDown(1)
		}
		if msg.String() == "k" {
			m.CursorUp()
		}
		if msg.String() == "l" {
			m.CursorRight()
		}
		if msg.String() == "h" {
			m.CursorLeft()
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.viewport.View()
}

// OpenFile opens the given file inside the buffer
func (m *Model) OpenFile(path string) error {
	var bytes []byte

	fd, err := os.OpenFile(path, os.O_RDWR, 0664) // taken from helix
	if err == nil {
		// File exists
		bytes, err = io.ReadAll(fd)
		if err != nil {
			// Some weird error happened. Display it.
			return err
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
	m.Val = &buf
	m.Lines = &lineBuf
	m.modified = false
	m.Cursor = cursor.New(&buf)
	m.viewport.SetContent(buf.String())

	return nil
}

// Name returns the title of the buffer window to display
func (b Model) Name() string {
	_, name := path.Split(b.Path)
	return name
}

func (b *Model) CursorDown() {
	b.Lines.CursorRight()
	lineIndex := b.Lines.GetAbs(b.Lines.GapEnd)
	b.Cursor.Goto(lineIndex)
}

func (b *Model) CursorUp() {
	b.Lines.CursorLeft()
	lineIndex := b.Lines.GetAbs(b.Lines.GapEnd)
	b.Cursor.Goto(lineIndex)

}

func (b *Model) CursorLeft() {
	// curLineStart := b.Lines.Get()
	b.Cursor.Left()
}

func (b *Model) CursorRight() {
	b.Cursor.Right()
}
