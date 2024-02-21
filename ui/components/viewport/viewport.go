package viewport

import (
	"io"
	"moe/pkg/themes"
	"os"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// buffer represents an opened file.
type buffer struct {
	parentNode *BufferNode

	cursor   Cursor    // Cursor position inside this buffer.
	path     string    // Absolute path on disk.
	fd       *os.File  // File descriptor.
	modified bool      // Content was modified and not saved to disk
	content  *[][]rune // Actual raw text data. TODO: Piece Chain.
}

// String returns the string contained in this buffer
func (b *buffer) String() string {
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

// Name returns the title of the buffer window to display
func (b buffer) Name() string {
	_, name := path.Split(b.path)
	return name
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

type Cursor struct {
	row, col int  // Location inside the text
	thin     bool // Wether to render a thin cursor or not
}

// ViewportStyle describes the style of each element in the viewport
type ViewportStyle struct {
	colorPalette         themes.Base16Theme
	cursorLine           lipgloss.Style
	bufferLineBackground lipgloss.Style
	bufferLineActive     lipgloss.Style
	bufferLineInactive   lipgloss.Style
	buffer               lipgloss.Style
}

// The bufferline is composed of a linked-list
type BufferNode struct {
	prev *BufferNode
	next *BufferNode
	val  *buffer
}

type LinkedList struct {
	head *BufferNode
	tail *BufferNode
}

func (l *LinkedList) AddNode(node *BufferNode) {
	InsertNode(l.tail, node)
}

// FocusBuffer finds the node with the given path and focuses it.
func (m *Model) FocusBuffer(path string) bool {
	list := m.bufl
	for node := list.head.next; node.next != nil; node = node.next {
		if node.val.path == path {
			m.focusedBuffer = node.val

			return true
		}
	}

	return false
}

type Model struct {
	bufl          LinkedList // Buffer Linked List
	focusedBuffer *buffer

	renderWindow              RenderWindow
	style                     ViewportStyle
	bufferWidth, bufferHeight int // Dimensions
}

func New() Model {

	// Loads the default theme. Note: in the future, default theme can be customizable
	theme := themes.DefaultTheme()

	defaultStyle := ViewportStyle{
		cursorLine:           lipgloss.NewStyle().Background(theme.Base03),
		bufferLineBackground: lipgloss.NewStyle().Background(theme.Base00),
		bufferLineActive:     lipgloss.NewStyle().Padding(0, 1).AlignHorizontal(lipgloss.Left).Background(theme.Base00).Foreground(theme.Base05).Reverse(true),
		bufferLineInactive:   lipgloss.NewStyle().Padding(0, 1).AlignHorizontal(lipgloss.Left).Background(theme.Base00).Foreground(theme.Base05).Reverse(false),
		buffer:               lipgloss.NewStyle().AlignHorizontal(lipgloss.Left).Background(theme.Base00),
	}

	nodeHead := &BufferNode{prev: nil, next: nil, val: nil}
	nodeTail := &BufferNode{prev: nil, next: nil, val: nil}
	nodeHead.next = nodeTail
	nodeTail.prev = nodeHead
	list := LinkedList{head: nodeHead, tail: nodeTail}

	return Model{
		focusedBuffer: nil,
		bufl:          list,
		style:         defaultStyle,
		renderWindow:  NewRenderWindow(),
	}
}

func newBuffer(path string) (*buffer, error) {
	fd, err := os.OpenFile(path, os.O_RDWR, 0664) // taken from helix
	if err != nil {
		return nil, err
	}

	// Note: I don't close the file descriptor, as I will also write to it...
	bytes, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	s := string(bytes)
	lines := strings.Split(s, "\n")
	var content [][]rune
	for _, line := range lines {
		content = append(content, []rune(line))
	}

	return &buffer{
		parentNode: nil,
		path:       path,
		fd:         fd,
		content:    &content,
		modified:   false,
		cursor: Cursor{
			row:  0,
			col:  0,
			thin: false,
		},
	}, nil
}

// InsertNode inserts node `n` before node `src`
func InsertNode(src *BufferNode, n *BufferNode) {
	n.prev = src.prev
	n.next = src
	src.prev.next = n
	src.prev = n
}

// ReplaceNode replaces node `old` with `new` in the Linked List
func ReplaceNode(old *BufferNode, new *BufferNode) {
	old.prev.next = new
	old.next.prev = new
	new.next = old.next
	new.prev = old.prev
}

// NewNode takes a *buffer and returns a *BufferNode
func NewNode(buf *buffer) *BufferNode {
	node := BufferNode{
		prev: nil,
		next: nil,
		val:  buf,
	}
	buf.parentNode = &node
	return &node
}

/* The three ELM functions: Init, Update and View */

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.bufferWidth = msg.Width
		m.bufferHeight = msg.Height - 3              // 2 lines for status bar and command bar and one for bufferline
		m.renderWindow.height = (m.bufferHeight - 1) // -1 line for bufferline
	case MsgOpenBuffer:
		path := string(msg)
		// Try and switch to that buffer if it is loaded already
		if m.FocusBuffer(path) {
			return m, nil
		}

		// If not, create a new buffer
		buf, err := newBuffer(path)
		if err != nil {
			return m, func() tea.Msg { return MsgError(err) }
		}

		node := NewNode(buf)
		m.bufl.AddNode(node)
		m.FocusBuffer(path) // Not that efficient, but fine unless you have 1M buffers open
	case MsgCloseBuffers:
		// bufs := []string(msg)
		// for _, buf := range bufs {
		// 	// cmd = m.closeBuf(buf, false)
		// }

	}
	return m, cmd
}

func (m Model) View() string {
	var bufferLine string
	var bufferText string

	// 1. Render the bufferline buffers
	for node := m.bufl.head.next; node.next != nil; node = node.next {
		if node == m.focusedBuffer.parentNode {
			bufferLine += m.style.bufferLineActive.Render(node.val.Name())
		} else {
			bufferLine += m.style.bufferLineInactive.Render(node.val.Name())
		}
	}

	// 1.1 Render the rest of the background TODO
	bufferLine += m.style.bufferLineBackground.Width(m.bufferWidth - lipgloss.Width(bufferLine)).Render()
	// 2. Render the text
	activeBuffer := m.focusedBuffer
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
