package buffer

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Ardelean-Calin/elmo/commands"
	"github.com/Ardelean-Calin/elmo/pkg/gapbuffer"

	"github.com/charmbracelet/bubbles/cursor"
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
type UpdateViewportMsg int
type LineChangedMsg int
type TreeReloadMsg struct {
	tree   *sitter.Node
	colors []byte
}

// SourceCode is the main container for the opened files. TODO name to something more generic, like Buffer?
type SourceCode struct {
	// Stores the raw data bytes. To be replaced with gapbuffer
	data []byte
	// Stores the colors for each byte. To be replaced with gapbuffer
	colors []byte
	// Cursor index
	cursor int
	// Treesitter representation
	tree *sitter.Node
	//
	lines map[int]LineInfo
}

// LineInfo describes a line. Using this I can easily index lines and get their length and indentation
type LineInfo struct {
	start       int
	end         int
	indentation int
}

// SetSource loads a file and computes the appropriate LineInfo's
func (s *SourceCode) SetSource(source []byte) {
	lines := make(map[int]LineInfo)
	i := 0
	prevLine := -1
	currentLine := 0
	foundChar := false
	lineInfo := LineInfo{
		start:       0,
		end:         0,
		indentation: 0,
	}
	for _, b := range source {
		if prevLine != currentLine {
			foundChar = false
			lineInfo.start = i
			lineInfo.end = i
			lineInfo.indentation = 0
			prevLine = currentLine
		}

		// Calculate the indentation by counting the number of space characters
		if b == '\t' && !foundChar {
			lineInfo.indentation += 4
		} else if b == ' ' && !foundChar {
			lineInfo.indentation++
		} else {
			foundChar = true
		}

		if b == '\n' {
			lineInfo.end = i
			lines[currentLine] = lineInfo
			currentLine++
		}
		i++
	}

	s.data = source
	s.colors = bytes.Repeat([]byte{0x05}, len(source))
	s.cursor = 0
	s.tree = nil
	s.lines = lines
}

// GetSlice returns the slice between start and end
func (s *SourceCode) GetSlice(start, end int) []byte {
	return s.data[start:end]
}

func (s *SourceCode) GetColors(start, end int) []byte {
	return s.colors[start:end]
}

// Returns a map of type lineIndex: {start in buffer, end in buffer}
func (s *SourceCode) Lines() map[int]LineInfo {
	return s.lines
}

type Viewport struct {
	offset        int
	width, height int
}

// Model represents an opened file.
type Model struct {
	Path     string                    // Absolute path on disk.
	fd       *os.File                  // File descriptor.
	GapBuf   gapbuffer.GapBuffer[rune] // Actual raw text data. Gap Buffer is a nice compromise between Piece Chain and buffer.
	Lines    gapbuffer.GapBuffer[int]  // The line numbers are also stored in a Gap Buffer
	Focused  bool
	modified bool // Content was modified and not saved to disk
	// Used just once on load
	ready         bool
	Width, Height int
	//  Then, the cursor will be strictly for display only (see footer.go)
	Cursor     cursor.Model // Cursor model
	CursorPos  int          // Current cursor position inside the current row
	CurrentRow int          // Current row index
	viewport   Viewport     // Scrollable viewport
	// TEMPORARY
	source     *SourceCode  // This replaces everything below
	Buffer     []byte       // Contains my file
	highlights []byte       // Contains a base16 color for each character
	tree       *sitter.Node // The current source's syntax tree
	// Viewport-related
	yOffset int
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
		CurrentRow: 0,
		tree:       nil,
		Buffer:     nil,
		highlights: nil,
		source:     nil,
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
			m.viewport = Viewport{offset: 0, height: msg.Height - 2, width: msg.Width}
			m.ready = true
			// Clear the default keybinds
		} else {
			m.viewport.height = msg.Height - 2
			m.viewport.width = msg.Width
		}

	case tea.KeyMsg:
		// Half page up
		if msg.String() == "ctrl+u" {
			m.viewport.offset = clamp(m.viewport.offset-m.viewport.height/2, 0, len(m.source.lines)-m.viewport.height+2)
		}
		// Half page down
		if msg.String() == "ctrl+d" {
			m.viewport.offset = clamp(m.viewport.offset+m.viewport.height/2, 0, len(m.source.lines)-m.viewport.height+2)
		}
		// TODO: Normal mode, insert mode, etc.
		if msg.String() == "j" {
			// m.viewport.LineDown(1)
			// cmd = CursorDown(&m, 1)
		}
		if msg.String() == "k" {
			// m.viewport.LineUp(1)
			// cmd = CursorUp(&m, 1)
		}
		if msg.String() == "l" {
			m.source.cursor += 1
		}
		if msg.String() == "h" {
			m.source.cursor -= 1
		}
		cmds = append(cmds, cmd)

	case tea.MouseMsg:
		switch msg.Button {
		// Scroll the viewport with the mouse wheel
		case tea.MouseButtonWheelUp:
			m.viewport.offset = clamp(m.viewport.offset-3, 0, len(m.source.lines)-m.viewport.height+2)
		case tea.MouseButtonWheelDown:
			m.viewport.offset = clamp(m.viewport.offset+3, 0, len(m.source.lines)-m.viewport.height+2)
		case tea.MouseButtonLeft:
			x, y := msg.X-6, msg.Y
			row := m.viewport.offset + y
			line := m.source.lines[row]
			var pos int
			if line.indentation > 0 {
				pos = min(x, line.indentation)/4 + clamp(x-line.indentation, 0, line.end)
			} else {
				pos = x
			}
			m.source.cursor = clamp(line.start+pos, line.start, line.end)
		}

	// Update the viewport content.
	// This content doesn't change when the cursor moves, only
	// when I enter a new character or stuff like that
	case UpdateViewportMsg:
		// var sb strings.Builder
		// for _, line := range m.Buffer {
		// 	sb.WriteString(string(line))
		// }

		// m.viewport.SetContent(sb.String())

	// A new syntax tree has been generated
	case TreeReloadMsg:
		m.source.colors = msg.colors
		m.source.tree = msg.tree
		// Issue a viewport update
		cmds = append(cmds, UpdateViewport, commands.ShowStatus("Treesitter loaded successfully"))
	}

	// Show the cursor
	m.Cursor, cmd = m.Cursor.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the Buffer content to screen
func (m Model) View() string {
	var sb strings.Builder
	start := m.viewport.offset
	end := m.viewport.offset + m.viewport.height
	for i := start; i < end; i++ {
		lineinfo := m.source.lines[i]
		line := m.source.GetSlice(lineinfo.start, lineinfo.end)
		colors := m.source.GetColors(lineinfo.start, lineinfo.end)

		// Write line numbers
		numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme[0x02]))
		sb.WriteString(numberStyle.Render(fmt.Sprintf("%4d  ", i+1)))

		for j, b := range line {
			if lineinfo.start+j == m.source.cursor {
				sb.WriteString(lipgloss.NewStyle().Reverse(true).Render(string(b)))
			} else {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme[colors[j]])).Render(string(b)))
			}
		}
		// shlLine := lipgloss.NewStyle().Foreground()

		// sb.Write(line)
		if i < end-1 {
			sb.WriteByte('\n')
		}
		// colors := m.source.GetColors(lineinfo.start, lineinfo.end)

	}
	// var sb strings.Builder
	return sb.String()
}

//go:embed lang/go/highlights.scm
var highlightsGo []byte

// GenerateSyntaxTree parses the source code using treesitter and generates
// a syntax tree for it.
func GenerateSyntaxTree(sourceCode *SourceCode) tea.Cmd {
	return func() tea.Msg {
		// Prepare the colors. One for each byte!
		colors := bytes.Repeat([]byte{0x05}, len(sourceCode.data))

		lang := golang.GetLanguage()
		tree, _ := sitter.ParseCtx(context.Background(), sourceCode.data, lang)

		q, err := sitter.NewQuery(highlightsGo, lang)
		if err != nil {
			panic(err)
		}
		qc := sitter.NewQueryCursor()
		qc.Exec(q, tree)

		// Iterate over query results
		for {
			m, ok := qc.NextMatch()
			if !ok {
				break
			}
			// log.Printf("M: %v", m)
			// Apply predicates filtering
			m = qc.FilterPredicates(m, sourceCode.data)
			for _, c := range m.Captures {
				name := q.CaptureNameForId(c.Index)
				// noRunes := utf8.RuneCount(sourceCode[b:e])

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

				for index := c.Node.StartByte(); index < c.Node.EndByte(); index++ {
					colors[index] = color
				}

			}
		}
		// Save the current tree and syntax highlighting
		return TreeReloadMsg{
			tree:   tree,
			colors: colors,
		}
	}
}

// OpenFile opens the given file inside the buffer
func (m *Model) OpenFile(path string) tea.Cmd {
	var err error
	// taken from helix
	fd, err := os.OpenFile(path, os.O_RDWR, 0664)
	if err != nil {
		return commands.ShowError(err)
	}

	content, err := io.ReadAll(fd)
	if err != nil {
		return commands.ShowError(err)
	}

	source := SourceCode{}
	source.SetSource(content)

	m.source = &source
	m.Path = path
	m.fd = fd
	m.modified = false
	m.CursorPos = 0
	m.CurrentRow = 0
	m.highlights = nil
	cursor := cursor.New()
	// cursor.SetChar(string(m.Buffer[0][0]))
	cursor.Focus()
	m.Cursor = cursor

	return tea.Batch(
		UpdateViewport,
		GenerateSyntaxTree(&source))

}

// Name returns the title of the buffer window to display
func (b Model) Name() string {
	_, name := path.Split(b.Path)
	return name
}

func (m *Model) updateCursor() {
	// line := m.lines[m.CursorPosV]
	// if m.CursorPosH <= line.indentation {
	// 	m.CursorPos = line.start + m.CursorPosH/4
	// } else {
	// 	m.CursorPos = line.start + m.CursorPosH - m.CursorPosH/4
	// }
	// hPos := m.CursorPosH
	// if hPos <= line.indentation && line.indentation != 0 {
	// 	m.CursorPosH = hPos / line.indentation
	// }
	// m.CursorPos = line.start + m.CursorPosH
	// char := m.Buffer.Bytes()[m.CursorPos]
	// if char == '\n' {
	// 	m.Cursor.SetChar(" ")
	// } else {
	// 	m.Cursor.SetChar(string(char))
	// }
}

func CursorDown(m *Model, n int) tea.Cmd {
	// m.CurrentRow = clamp(m.CurrentRow+1, 0, len(m.lines))
	// m.updateCursor()

	// return UpdateViewport
	return nil
}

func CursorUp(m *Model, n int) tea.Cmd {
	// m.CurrentRow = clamp(m.CurrentRow-1, 0, len(m.lines))
	// m.updateCursor()

	// return Render(m)
	// m.Lines.CursorLeft()
	// m.CursorPos = m.Lines.Current()
	// m.CursorPos = min(m.CursorPos+m.CursorPosH, m.Lines.Next()-1)

	// return tea.Batch(LineChanged(), Render(m))
	// return UpdateViewport
	return nil
}

func CursorLeft(m *Model, n int) tea.Cmd {
	// m.CursorPos = clamp(m.CursorPos-1, 0, m.lines[m.CurrentRow].length)
	return nil
	// var cmds []tea.Cmd

	// m.CursorPos = clamp(m.CursorPos-1, 0, m.GapBuf.Count())
	// // Going left got us on a new line
	// if m.CursorPos < m.Lines.Current() {
	// 	m.Lines.CursorLeft()
	// 	m.CursorPosH = m.CursorPos - m.Lines.Current()
	// 	cmds = append(cmds, LineChanged())
	// } else {
	// 	m.CursorPosH = m.CursorPos - m.Lines.Current()
	// }

	// cmds = append(cmds, Render(m))
	// return tea.Batch(cmds...)
}

func CursorRight(m *Model, n int) tea.Cmd {
	// m.CursorPos = clamp(m.CursorPos+1, 0, m.lines[m.CurrentRow].length)
	// m.CursorPosH += 1
	// line := m.lines[m.CursorPosV]
	// if m.CursorPosH >= line.length {
	// 	m.CursorPosH = 0
	// 	m.CursorPosV += 1
	// }
	return nil
	// m.updateCursor()
	// return UpdateViewport

	// var cmds []tea.Cmd

	// m.CursorPos = clamp(m.CursorPos+1, 0, m.GapBuf.Count())
	// // Going right got us to a new line
	// if m.CursorPos >= m.Lines.Next() {
	// 	m.Lines.CursorRight()
	// 	m.CursorPosH = 0
	// 	cmds = append(cmds, LineChanged())
	// } else {
	// 	m.CursorPosH = m.CursorPos - m.Lines.Current()
	// }

	// cmds = append(cmds, Render(m))
	// return tea.Batch(cmds...)
}

func UpdateViewport() tea.Msg {
	return UpdateViewportMsg(0)
}

// clamp limits the value of val between [low, high)
func clamp(val, low, high int) int {
	return max(low, min(val, high-1))
}
