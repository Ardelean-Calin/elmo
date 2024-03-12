package buffer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Ardelean-Calin/elmo/pkg/gapbuffer"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// Catppuccin Frappe
var theme = []string{
	"#303446", // base00
	"#292c3c", // base01
	"#414559", // base02
	"#51576d", // base03
	"#626880", // base04
	"#c6d0f5", // base05
	"#f2d5cf", // base06
	"#babbf1", // base07
	"#e78284", // base08
	"#ef9f76", // base09
	"#e5c890", // base0A
	"#a6d189", // base0B
	"#81c8be", // base0C
	"#8caaee", // base0D
	"#ca9ee6", // base0E
	"#eebebe", // base0F
}

// Contains the new cursor coordinates
type UpdateViewportMsg []byte
type LineChangedMsg int

// Model represents an opened file.
type Model struct {
	Path     string                    // Absolute path on disk.
	fd       *os.File                  // File descriptor.
	Buffer   []byte                    // Contains my file
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
	// TEMPORARY
	highlights []byte
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
		highlights: nil,
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
			// Clear the default keybinds
			m.viewport.KeyMap = New().viewport.KeyMap
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
		var sb strings.Builder
		for i, b := range msg {
			_ = i
			sb.WriteString(
				lipgloss.NewStyle().Foreground(lipgloss.Color(
					theme[m.highlights[i]],
				)).Render(string(b)))
		}
		content := sb.String()
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
	// var sb strings.Builder
	// for _, b := range m.Buffer {
	// 	sb.WriteByte(b)
	// }

	// return sb.String()
	return m.viewport.View()
}

type Hightlight struct {
	name      string
	startByte int
	endByte   int
}

type HightlightList []Hightlight

func (h *HightlightList) Find(index int) (int, bool) {
	for i, v := range *h {
		if index >= v.startByte && index < v.endByte {
			return i, true
		}
	}
	return -1, false
}

func loadTreesitter(m *Model, sourceCode []byte) {

	lang := golang.GetLanguage()
	tree, _ := sitter.ParseCtx(context.Background(), sourceCode, lang)

	highlights, _ := os.ReadFile("highlights.scm")
	q, err := sitter.NewQuery(highlights, lang)
	if err != nil {
		panic(err)
	}
	qc := sitter.NewQueryCursor()
	qc.Exec(q, tree)

	hi := make([]byte, len(sourceCode))
	for i := range hi {
		hi[i] = 0x05 // Base05 for the default text color
	}
	// Iterate over query results
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// log.Printf("M: %v", m)
		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		for _, c := range m.Captures {
			name := q.CaptureNameForId(c.Index)
			b, e := c.Node.StartByte(), c.Node.EndByte()
			length := int(e - b)

			// The most basic of syntax highlighting!
			var color uint8
			switch name {
			case "comment":
				color = 0x03
			case "constant.builtin":
				color = 0x09
			case "escape":
				color = 0x0C
			case "function":
				color = 0x0D
			case "function.method":
				color = 0x0D
			case "keyword":
				color = 0x0E
			case "number":
				color = 0x09
			case "operator":
				color = 0x0C
			case "package":
				color = 0x0D
			case "punctuation.bracket":
				color = 0x05
			case "string":
				color = 0x0B
			case "type":
				color = 0x0A
			case "variable.member":
				color = 0x0C
			default:
				color = 0x05
			}

			copy(hi[b:e], bytes.Repeat(
				[]byte{color},
				length,
			))

			// PoC, replace with some map or something
			// if strings.HasPrefix(name, "comment") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x03},
			// 		length,
			// 	))
			// } else if strings.HasPrefix(name, "constant") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x09},
			// 		length,
			// 	))
			// } else if name == "function" {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x0D},
			// 		length,
			// 	))
			// } else if name == "function.method" {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x0D},
			// 		length,
			// 	))
			// } else if strings.HasPrefix(name, "keyword") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x0E},
			// 		length,
			// 	))
			// 	// } else if strings.HasPrefix(name, "label") {
			// 	// 	copy(hi[b:e], bytes.Repeat(
			// 	// 		[]byte{0x0B},
			// 	// 		length,
			// 	// 	))
			// } else if strings.HasPrefix(name, "namespace") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x0D},
			// 		length,
			// 	))
			// } else if strings.HasPrefix(name, "operator") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x05},
			// 		length,
			// 	))
			// } else if strings.HasPrefix(name, "punctuation") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x05},
			// 		length,
			// 	))
			// } else if strings.HasPrefix(name, "string") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x0B},
			// 		length,
			// 	))
			// } else if strings.HasPrefix(name, "type") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x0A},
			// 		length,
			// 	))
			// } else if name == "variablem.parameter" || strings.HasPrefix(name, "varaible.other.member") {
			// 	copy(hi[b:e], bytes.Repeat(
			// 		[]byte{0x08},
			// 		length,
			// 	))
			// }
		}
	}
	m.highlights = hi
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

	m.Buffer = bytes
	// m.viewport.SetContent(string(bytes))
	loadTreesitter(m, bytes)

	m.Path = path
	m.fd = fd
	m.modified = false
	m.CursorPos = 0
	m.CursorPosH = 0
	m.Cursor = cursor.New()

	return func() tea.Msg { return UpdateViewportMsg(bytes) }, nil
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

// Render is the command which builds the text to be rendered.
// This should be separate from the cursor and line number rendering so that it only updates when needed.
// Only events such as adding a new character should trigger
// a re-render
func Render(m *Model) tea.Cmd {
	return nil
	// Render will create a Frame Buffer with all my data
	var sb strings.Builder
	// TODO. Change to a byte-buffer based approach which
	// I then send in the UpdateViewportMsg.
	// var x bytes.Buffer
	// Should I use bufio? I can then ReadFrom in the View method
	// var y bufio.Writer

	// for _, r := range m.GapBuf.Collect() {
	// 	y.WriteRune(r)
	// }

	// // Costly operation, only do this once
	// bufContent := []byte(m.GapBuf.String())
	// // bytes := []byte(bufContent)
	// sb.Write(bufContent)
	// return func() tea.Msg {
	// 	return UpdateViewportMsg(sb.String())
	// }

	for lineNo, bufIndex := range m.Lines.Collect() {
		var lineBuilder strings.Builder

		// Highlight the current line
		lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3c3d4f"))
		if lineNo == m.Lines.GapStart {
			lineStyle = lineStyle.Foreground(lipgloss.Color("#878ebf")).Background(lipgloss.Color("#2a2b3c"))
		}
		textStyle := lineStyle.Copy().UnsetForeground()

		// Write line numbers
		lineBuilder.WriteString(lineStyle.Render(fmt.Sprintf("%4d  ", lineNo+1)))

		// Now we render the text, then the remaining space
		for {
			done := false
			r := m.GapBuf.GetAbs(bufIndex)
			if r == '\n' {
				// Always render a space instead of a newline
				r = ' '
				done = true
			}

			// textStyle.UnsetForeground()
			// index, ok := m.highlights.Find(bufIndex)
			// if ok {
			// 	h := m.highlights[index]
			// 	// _ = h
			// 	if strings.HasPrefix(h.name, "string") {
			// 		textStyle = textStyle.Foreground(lipgloss.Color("#a6d189"))
			// 	} else if strings.HasPrefix(h.name, "comment") {
			// 		textStyle = textStyle.Foreground(lipgloss.Color("#626880"))
			// 	} else if strings.HasPrefix(h.name, "keyword") {
			// 		textStyle = textStyle.Foreground(lipgloss.Color("#ca9ee6"))
			// 	} else if strings.HasPrefix(h.name, "variable") {
			// 		textStyle = textStyle.Foreground(lipgloss.Color("#81c8be"))
			// 	} else if strings.HasPrefix(h.name, "operator") || strings.HasPrefix(h.name, "punctuation") {
			// 		textStyle = textStyle.Foreground(lipgloss.Color("#c6d0f5"))
			// 	} else if strings.HasPrefix(h.name, "constant") {
			// 		textStyle = textStyle.Foreground(lipgloss.Color("#ef9f76"))
			// 	} else if strings.HasPrefix(h.name, "type") {
			// 		textStyle = textStyle.Foreground(lipgloss.Color("#e5c890"))
			// 	}
			// }

			// Render cursor
			if bufIndex == m.CursorPos {
				m.Cursor.Focus()
				m.Cursor.SetChar(string(r))

				lineBuilder.WriteString(m.Cursor.View())
			} else {
				lineBuilder.WriteString(textStyle.Render(string(r)))
			}
			bufIndex++
			if done {
				break
			}
		}

		// Limit the width, so no wrapping (for now)
		sb.WriteString(
			lipgloss.NewStyle().
				Width(m.viewport.Width).
				Render(lineBuilder.String()))
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
