package viewport

import (
	"fmt"
	"io"
	"moe/pkg/themes"
	"moe/ui/components/viewport/render"
	"os"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* These two will be displayed in the status bar */
type StatusMsg string
type ErrorMsg error

// Return a status string. Will be displayed in the statusbar
func StatusCmd(status string) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg(status)
	}
}

// Return an error. Will be displayed in the statusbar
func ErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg(err)
	}
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

type LinkedList struct {
	head *BufferNode
	tail *BufferNode
}

func (l *LinkedList) AddNode(node *BufferNode) *BufferNode {
	InsertNode(l.tail, node)
	return node
}

// FocusBuffer finds the node with the given path and focuses it.
func (m *Model) FocusBuffer(path string) bool {
	list := m.blist
	for node := list.head.next; node.next != nil; node = node.next {
		if node.buffer.path == path {
			m.focusedNode = node

			return true
		}
	}

	return false
}

type Model struct {
	blist       LinkedList  // Buffer Linked List
	focusedNode *BufferNode // The currently active buffer

	Style                     ViewportStyle
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

	// By default, we have an empty screen. TODO: add a scratch buffer
	nodeHead := &BufferNode{prev: nil, next: nil, buffer: nil}
	nodeTail := &BufferNode{prev: nil, next: nil, buffer: nil}
	nodeHead.next = nodeTail
	nodeTail.prev = nodeHead
	list := LinkedList{head: nodeHead, tail: nodeTail}

	return Model{
		focusedNode: nil,
		blist:       list,
		Style:       defaultStyle,
	}
}

// WriteBuffer saves the currently focused buffer to disk.
func (m *Model) WriteBuffer() tea.Cmd {
	node := m.focusedNode
	if node == nil {
		return ErrorCmd(fmt.Errorf("Unknown error"))
	}

	if node.buffer.path == "" {
		return ErrorCmd(fmt.Errorf("No write path specified"))
	}

	if node.buffer.fd == nil {
		return ErrorCmd(fmt.Errorf("Unimplemented"))
	}

	_, err := node.buffer.fd.WriteString(node.buffer.String())
	if err != nil {
		return ErrorCmd(err)
	}
	node.buffer.modified = false

	return nil
}

// findBuffer searches the linked list and returns the BufferNode
func (m *Model) findBuffer(bPath string) (node *BufferNode) {
	for node := m.blist.head.next; node.next != nil; node = node.next {
		if node.buffer.path == bPath {
			return node
		}
	}

	return nil
}

// OpenBuffer takes in a list of paths and opens the files for editing.
func (m *Model) OpenBuffer(paths ...string) tea.Cmd {
	var cmds []tea.Cmd
	for _, p := range paths {
		cmd := m.openBuffer(p)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// openBuffer is the internal function of viewport.Model that only handles one buffer.
func (m *Model) openBuffer(bname string) tea.Cmd {
	var bytes []byte

	// Check if buffer already is open and focus it
	if m.FocusBuffer(bname) {
		// We already had the buffer opened and focused it.
		return nil
	}

	// Else create a new buffer
	fd, err := os.OpenFile(bname, os.O_RDWR, 0664) // taken from helix
	if err == nil {
		// File doesn't exist
		bytes, err = io.ReadAll(fd)
		if err != nil {
			// Some weird error happened. Display it.
			return ErrorCmd(err)
		}
	}

	// Ok by this point I either have a fd with some bytes or a nil fd and nil bytes
	s := string(bytes)
	var content [][]rune
	for _, line := range strings.Split(s, "\n") {
		content = append(content, []rune(line))
	}

	// Create a new buffer and add it to our linked-list
	b := buffer{
		parentNode: nil,
		cursor:     Cursor{0, 0, false},
		path:       bname,
		fd:         fd,
		modified:   false,
		content:    &content,
		window:     render.NewRenderWindow(),
	}

	node := m.blist.AddNode(Node(&b))
	m.focusedNode = node

	return nil
}

// func (m * Model) NewBuffer(bname string) tea{}
// func (m * Model) NewBuffer(bname string) {}

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
		m.bufferHeight = msg.Height - 3 // 2 lines for status bar and command bar and one for bufferline
		for node := m.blist.head.next; node.next != nil; node = node.next {
			node.buffer.window.Height = (m.bufferHeight - 1)
		}
	}
	return m, cmd
}

func (m Model) View() string {
	var bufferLine string
	var bufferText string

	// 1. Render the bufferline buffers
	for node := m.blist.head.next; node.next != nil; node = node.next {
		if node == m.focusedNode {
			bufferLine += m.Style.bufferLineActive.Render(node.buffer.Name())
		} else {
			bufferLine += m.Style.bufferLineInactive.Render(node.buffer.Name())
		}
	}

	// 1.1 Render the rest of the background TODO
	bufferLine += m.Style.bufferLineBackground.Width(m.bufferWidth - lipgloss.Width(bufferLine)).Render()
	// 2. Render the text
	if m.focusedNode != nil {
		content := m.focusedNode.buffer.String()
		m.focusedNode.buffer.window.SetContent(content)
		bufferText = m.Style.buffer.Width(m.bufferWidth).Height(m.bufferHeight).Render(m.focusedNode.buffer.window.View())
	}

	// 3. Render the cursor... How?
	// lipgloss.Place(1, 1, )
	bufferText = m.Style.buffer.Height(m.bufferHeight).Render(bufferText)

	// 4. Sum everything up
	s := lipgloss.JoinVertical(lipgloss.Left, bufferLine, bufferText)
	return s
}
