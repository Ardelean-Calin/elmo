package viewport

import (
	"io"
	"log"
	"os"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Buffer represents an opened file.
type Buffer struct {
	path     string    // Absolute path on disk.
	fd       *os.File  // File descriptor.
	focused  bool      // Bool indicating if we are currently editing this buffer.
	scratch  bool      // This is a scratch buffer
	modified bool      // Content was modified and not saved to disk
	content  *[][]rune // Actual raw text data. TODO: Piece Chain.
}

func (b *Buffer) Focus() {
	b.focused = true
}

func (b *Buffer) Unfocus() {
	b.focused = false
}

// String returns the string contained in this buffer
func (b *Buffer) String() string {
	content := b.content
	if content == nil {
		return ""
	}

	var sb strings.Builder
	for _, r := range *content {
		for _, v := range r {
			sb.WriteRune(v)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (b Buffer) GetName() string {
	if b.scratch {
		return "[scratch]"
	}

	_, name := path.Split(b.path)
	return name
}

func newScratchBuffer() *Buffer {
	tmp := [][]rune{{'a', 'b', 'c'}, {'\t', 'f', 'o', 'o'}}
	return &Buffer{
		path:    "",
		fd:      nil,
		focused: true,
		scratch: true,
		content: &tmp,
	}
}

// RenderWindow represents a rectangular window "sliding" over our text.
// Let's keep it as simple as possible, it is a window where only the height can be changed.
// Only the text that is inside the window gets rendered on screen.
type RenderWindow struct {
	startRow int
	height   int
}

func NewRenderWindow() RenderWindow {
	return RenderWindow{startRow: 0, height: 0}
}

func (w *RenderWindow) SetHeight(h int) {
	w.height = h
}

func (w *RenderWindow) SetStart(s int) {
	w.startRow = s
}

func (w *RenderWindow) GetHeight() int {
	return w.height
}

func (w *RenderWindow) GetStart() int {
	return w.startRow
}

func (w *RenderWindow) Apply(content [][]rune) string {
	if content == nil {
		return ""
	}

	var sb strings.Builder
	for i, r := range content[w.startRow:] {
		if i == w.height {
			break
		}

		for _, v := range r {
			sb.WriteRune(v)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

type cursor struct {
	row, col int  // Location inside the text
	thin     bool // Wether to render a thin cursor or not
}

type ViewportStyle struct {
	cursorLine           lipgloss.Style
	bufferLineBackground lipgloss.Style
	bufferLineActive     lipgloss.Style
	bufferLineInactive   lipgloss.Style
	buffer               lipgloss.Style
}

type Model struct {
	buffers                   []*Buffer
	Cursor                    cursor
	renderWindow              RenderWindow
	style                     ViewportStyle
	bufferWidth, bufferHeight int // Dimensions
}

func New() Model {
	// scratch := newScratchBuffer()
	test, _ := newBuffer("foo.go")
	test.Focus()

	// bgColor := lipgloss.Color("#1d2021")
	// bgColorBufferline := lipgloss.Color("#928374")
	// bgColorBufferline := lipgloss.Color("#1d2021")
	base00 := lipgloss.Color("#1d2021") // Default Background
	_ = base00
	base01 := lipgloss.Color("#3c3836") // Lighter Background (Used for status bars, line number and folding marks)
	_ = base01
	base02 := lipgloss.Color("#504945") // Selection Background
	_ = base02
	base03 := lipgloss.Color("#665c54") // Comments, Invisibles, Line Highlighting
	_ = base03
	base04 := lipgloss.Color("#bdae93") // Dark Foreground (Used for status bars)
	_ = base04
	base05 := lipgloss.Color("#d5c4a1") // Default Foreground, Caret, Delimiters, Operators
	_ = base05
	base06 := lipgloss.Color("#ebdbb2") // Light Foreground (Not often used)
	_ = base06
	base07 := lipgloss.Color("#fbf1c7") // Light Background (Not often used)
	_ = base07
	base08 := lipgloss.Color("#fb4934") // Variables, XML Tags, Markup Link Text, Markup Lists, Diff Deleted
	_ = base08
	base09 := lipgloss.Color("#fe8019") // Integers, Boolean, Constants, XML Attributes, Markup Link Url
	_ = base09
	base0A := lipgloss.Color("#fabd2f") // Classes, Markup Bold, Search Text Background
	_ = base0A
	base0B := lipgloss.Color("#b8bb26") // Strings, Inherited Class, Markup Code, Diff Inserted
	_ = base0B
	base0C := lipgloss.Color("#8ec07c") // Support, Regular Expressions, Escape Characters, Markup Quotes
	_ = base0C
	base0D := lipgloss.Color("#83a598") // Functions, Methods, Attribute IDs, Headings
	_ = base0D
	base0E := lipgloss.Color("#d3869b") // Keywords, Storage, Selector, Markup Italic, Diff Changed
	_ = base0E
	base0F := lipgloss.Color("#d65d0e") // Deprecated, Opening/Closing Embedded Language Tags, e.g. <?php ?>
	_ = base0F

	defaultStyle := ViewportStyle{
		cursorLine:           lipgloss.NewStyle().Background(base03),
		bufferLineBackground: lipgloss.NewStyle().Background(base00),
		bufferLineActive:     lipgloss.NewStyle().Padding(0, 1).AlignHorizontal(lipgloss.Left).Background(base00).Foreground(base05).Reverse(true),
		bufferLineInactive:   lipgloss.NewStyle().Padding(0, 1).AlignHorizontal(lipgloss.Left).Background(base00).Foreground(base05).Reverse(false),
		buffer:               lipgloss.NewStyle().AlignHorizontal(lipgloss.Left).Background(base00),
	}

	// bLineBackground := lipgloss.NewStyle().Background(bgColorBufferline)
	// bLineInactive := lipgloss.NewStyle().Padding(0, 1).AlignHorizontal(lipgloss.Left)
	// bLineActive := bLineInactive.Copy().Reverse(true)
	// bStyle := lipgloss.NewStyle().
	// 	AlignHorizontal(lipgloss.Left)
	// Background(bgColor)

	return Model{
		buffers:      []*Buffer{test},
		style:        defaultStyle,
		Cursor:       cursor{0, 0, false}, // By default, we have a normal-mode thick cursor on the first character
		renderWindow: NewRenderWindow(),
	}
}

// GetActiveBuffer returns a pointer to the currently active buffer.
func (m *Model) GetActiveBuffer() *Buffer {
	for _, b := range m.buffers {
		if b.focused {
			return b
		}
	}
	return nil
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.bufferWidth = msg.Width
		m.bufferHeight = msg.Height - 3              // 2 lines for status bar and command bar and one for bufferline
		m.renderWindow.SetHeight(m.bufferHeight - 1) // -1 line for bufferline
	}
	return m, nil
}

func newBuffer(path string) (*Buffer, error) {
	fd, err := os.OpenFile(path, os.O_RDWR, 0664) // taken from helix
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	// Note: I don't close the file descriptor, as I will also write to it...

	s := string(bytes)
	lines := strings.Split(s, "\n")
	var content [][]rune
	for _, line := range lines {
		content = append(content, []rune(line))
		// content[i] = []rune(line)
	}

	return &Buffer{
		path:    path,
		fd:      fd,
		focused: false,
		scratch: false,
		content: &content,
	}, nil
}

func (m *Model) deactivateAll() {
	for _, b := range m.buffers {
		b.Unfocus()
	}
}

func (m Model) OpenBuffer(path string) (Model, tea.Cmd) {
	for _, b := range m.buffers {
		if b.path == path {
			m.deactivateAll()
			b.Focus()

			return m, nil
		}
	}

	buf, err := newBuffer(path)
	if err != nil {
		log.Fatalf("Error opening %s: %v\n", path, err)
		// TODO: Launch an error message
	}
	buf.Focus()
	m.deactivateAll()
	m.buffers = append(m.buffers, buf)

	return m, nil
}

// Render the fg string over the bg string
// func Overlay(bg string, fg string) string {
// 	lipgloss.JoinHorizontal(lipgloss.Left, fg, bg)
// 	// maxSize := max(len(bg), len(fg))

// 	// var s string = make(string, maxSize)
// 	return ""
// }

func (m Model) View() string {
	var bufferLine string
	var bufferText string
	// 1. Render the bufferline buffers
	for _, b := range m.buffers {
		name := b.GetName()
		if b.focused {
			bufferLine += m.style.bufferLineActive.Render(name)
		} else {
			bufferLine += m.style.bufferLineInactive.Render(name)
		}
	}
	// 1.1 Render the rest of the background TODO
	bufferLine += m.style.bufferLineBackground.Width(m.bufferWidth - lipgloss.Width(bufferLine)).Render()
	// 2. Render the text
	activeBuffer := m.GetActiveBuffer()
	if activeBuffer != nil {
		content := m.renderWindow.Apply(*activeBuffer.content) // Only display what we can see through the window
		bufferText = m.style.buffer.Width(m.bufferWidth).Height(m.bufferHeight).Render(content)
	}

	// 3. Render the cursor... How?
	// lipgloss.Place(1, 1, )
	bufferText = m.style.buffer.Height(m.bufferHeight).Render(bufferText)

	// 4. Sum everything up
	s := lipgloss.JoinVertical(lipgloss.Left, bufferLine, bufferText)
	return s
}
