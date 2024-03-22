package buffer

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Ardelean-Calin/elmo/pkg/gapbuffer"
	"github.com/Ardelean-Calin/elmo/ui/components/footer"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/nix"
	"github.com/smacker/go-tree-sitter/rust"
)

// Catppuccin Mocha
var theme = []string{
	"#1e1e2e", // base
	"#181825", // mantle
	"#313244", // surface0
	"#45475a", // surface1
	"#585b70", // surface2
	"#cdd6f4", // text
	"#f5e0dc", // rosewater
	"#b4befe", // lavender
	"#f38ba8", // red
	"#fab387", // peach
	"#f9e2af", // yellow
	"#a6e3a1", // green
	"#94e2d5", // teal
	"#89b4fa", // blue
	"#cba6f7", // mauve
	"#f2cdcd", // flamingo
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
	data gapbuffer.GapBuffer[byte]
	// Contains a base16 color for each character
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
	start int
	end   int
}

// SetSource loads a file and computes the appropriate LineInfo's
func (s *SourceCode) SetSource(source []byte) {
	buf := gapbuffer.NewGapBuffer[byte]()
	buf.SetContent(source)

	s.data = buf
	s.colors = bytes.Repeat([]byte{0x05}, len(source))
	s.cursor = 0
	s.tree = nil
	s.RegenerateLines()
}

// RegenerateLines regenerates the line information
func (s *SourceCode) RegenerateLines() {
	lines := make(map[int]LineInfo)

	lineBreaks := s.data.FindAll('\n')
	lines[0] = LineInfo{0, lineBreaks[0]}

	for i, index := range lineBreaks {
		if i == len(lineBreaks)-1 {
			lines[i+1] = LineInfo{index + 1, s.data.Len()}
			break
		}
		// Since the range is [open, closed) we consider a line to be starting at the first
		// character after '\n' and ending at the last character before '\n'
		lines[i+1] = LineInfo{index + 1, lineBreaks[i+1]}
	}

	s.lines = lines
}

// GetSlice returns the slice between start and end
func (s *SourceCode) GetSlice(start, end int) []byte {
	return s.data.Collect()[start:end]
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

// Stores editor mode
type Mode int

const (
	Normal Mode = iota
	Insert
	Select
)

// Model represents an opened file.
type Model struct {
	Path     string   // Absolute path on disk.
	fd       *os.File // File descriptor.
	Focused  bool
	modified bool // Content was modified and not saved to disk
	// Used just once on load
	ready bool
	//  Then, the cursor will be strictly for display only (see footer.go)
	// TEMPORARY
	source    *SourceCode // This replaces everything below
	viewport  Viewport    // Scrollable viewport
	selection [2]int      // 2 indices for the currently selected text
	Mode      Mode        // Current buffer mode
}

func New() Model {
	return Model{
		Path:      "",
		fd:        nil,
		Focused:   true,
		modified:  false,
		ready:     false,
		Mode:      Normal,
		source:    nil,
		selection: [2]int{0, 0},
	}
}

func (m Model) Init() tea.Cmd {
	log.Printf("buffer.go: Init() called")
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
		// Normal mode keybindings
		if m.Mode == Normal {
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
				m.source.cursorDown(1)
			}
			if msg.String() == "k" {
				m.source.cursorUp(1)
			}
			if msg.String() == "l" {
				m.source.cursorRight(1)
			}
			if msg.String() == "h" {
				m.source.cursorLeft(1)
			}
			if msg.String() == "i" {
				m.Mode = Insert
				m.source.data.CursorGoto(m.source.cursor)
			}
		} else if m.Mode == Insert && msg.Alt == false {
			if msg.String() == "esc" {
				m.Mode = Normal
				cmd = nil
			}

			if msg.Type == tea.KeyRunes {
				chars := []byte(msg.String())
				m.source.data.InsertSlice(chars)
				m.source.cursor += len(chars)
				m.source.RegenerateLines()
				cmd = GenerateSyntaxTree(m.source, ".go")
				// TODO Invalidate treesitter & Schedule a treesitter regeneration 100-200ms into the future
			}

			if msg.Type == tea.KeySpace {
				m.source.data.Insert(' ')
				m.source.cursor++
				m.source.RegenerateLines()
				cmd = GenerateSyntaxTree(m.source, ".go")
			}

			if msg.Type == tea.KeyEnter {
				m.source.data.Insert('\n')
				m.source.cursor++
				m.source.RegenerateLines()
				cmd = GenerateSyntaxTree(m.source, ".go")
			}

			if msg.Type == tea.KeyBackspace {
				m.source.data.Backspace()
				m.source.cursor--
				m.source.RegenerateLines()
				cmd = GenerateSyntaxTree(m.source, ".go")
			}

		}
		cmds = append(cmds, cmd)

	case tea.MouseMsg:
		evt, action := msg.Button, msg.Action
		switch evt {
		// Scroll the viewport with the mouse wheel
		case tea.MouseButtonWheelUp:
			m.viewport.offset = clamp(m.viewport.offset-3, 0, len(m.source.lines)-m.viewport.height+2)
		case tea.MouseButtonWheelDown:
			m.viewport.offset = clamp(m.viewport.offset+3, 0, len(m.source.lines)-m.viewport.height+2)
		case tea.MouseButtonLeft:
			x, y := msg.X-7, msg.Y // Allocate 7 for the line numbers + gutter
			row := m.viewport.offset + y
			line := m.source.lines[row]

			// Handle line indentation when rendering by mapping the
			// x coordinate of the mouse click to a line offset iteratively
			pos := 0
			for i := line.start; i < line.end; i++ {
				c := m.source.data.GetAbs(i)
				if c == '\t' {
					x -= 4
				} else {
					x -= 1
				}

				if x < 0 {
					break
				}
				pos += 1
			}

			m.source.cursor = clamp(line.start+pos, line.start, line.end+1)
			if action == tea.MouseActionPress {
				// Start selection => save selection start to a variable
				m.selection[0] = m.source.cursor
			}
			m.selection[1] = m.source.cursor
		}

	// A new syntax tree has been generated
	case TreeReloadMsg:
		m.source.colors = msg.colors
		m.source.tree = msg.tree
		// Issue a viewport update
		cmds = append(cmds, UpdateViewport, footer.ShowStatus("Treesitter loaded successfully"))
	}

	return m, tea.Batch(cmds...)
}

// View renders the Buffer content to screen
func (m Model) View() string {
	if m.source == nil {
		return lipgloss.NewStyle().
			Width(m.viewport.width).
			Height(m.viewport.height).
			Render("")
	}

	var sb strings.Builder
	start := clamp(m.viewport.offset, 0, len(m.source.lines))
	end := clamp(m.viewport.offset+m.viewport.height, 0, len(m.source.lines))
	for i := start; i < end; i++ {
		var lb strings.Builder
		var fg, bg lipgloss.Color

		lineinfo := m.source.lines[i]
		line := m.source.GetSlice(lineinfo.start, lineinfo.end)
		colors := m.source.GetColors(lineinfo.start, lineinfo.end)

		// Write line numbers TODO I could maybe move this inside another component?
		numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme[0x03])).Background(lipgloss.Color(theme[0x00]))
		lb.WriteString(numberStyle.Render(fmt.Sprintf("%5d  ", i+1)))
		// TODO: Also render the Git Gutter here using these: ▔ ▍

		// Render the cursor and the selection
		for j, b := range line {
			absolutePos := lineinfo.start + j
			if m.source.cursor == absolutePos {
				lb.WriteString(lipgloss.NewStyle().Reverse(true).Render(string(b)))
			} else {
				fg = lipgloss.Color(theme[colors[j]])
				// Normal render. All characters are rendered one-by-one
				// with their appropriate color
				selStart := min(m.selection[0], m.selection[1])
				selEnd := max(m.selection[0], m.selection[1])
				if absolutePos < selEnd && absolutePos >= selStart {
					bg = lipgloss.Color(theme[0x02])
				} else {
					bg = lipgloss.Color(theme[0x00])
				}

				lb.WriteString(
					lipgloss.NewStyle().
						Foreground(fg).
						Background(bg).
						Render(string(b)))
			}
		}

		// If the cursor is on a line end (aka \n), render a whitespace
		if m.source.cursor == lineinfo.end {
			lb.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
		}

		// Render the background
		// bg = lipgloss.Color(theme[0x00])
		// textLen := lipgloss.Width(lb.String())
		// lb.WriteString(lipgloss.NewStyle().Background(bg).Width(m.viewport.width - textLen).Render(" "))

		// Last character in the viewport needs not be a newline, or
		// I will get a weird empty line at the end
		if i < end-1 {
			lb.WriteByte('\n')
		}

		// Finally write the line to the framebuffer
		sb.WriteString(lb.String())
	}
	return sb.String()
}

//go:embed syntax/go/highlights.scm
var highlightsGo []byte

//go:embed syntax/rust/highlights.scm
var highlightsRust []byte

//go:embed syntax/nix/highlights.scm
var highlightsNix []byte

// GenerateSyntaxTree parses the source code using treesitter and generates
// a syntax tree for it.
func GenerateSyntaxTree(sourceCode *SourceCode, ext string) tea.Cmd {
	return func() tea.Msg {
		// Prepare the colors. One for each byte!
		colors := bytes.Repeat([]byte{0x05}, sourceCode.data.Len())

		var lang *sitter.Language
		var highlights []byte
		if ext == ".go" {
			lang = golang.GetLanguage()
			highlights = highlightsGo
		} else if ext == ".rs" {
			lang = rust.GetLanguage()
			highlights = highlightsRust
		} else if ext == ".nix" {
			lang = nix.GetLanguage()
			highlights = highlightsNix
		} else {
			log.Printf("[Treesitter] Unsupported language: %s", ext)
			return nil
		}
		tree, _ := sitter.ParseCtx(context.Background(), sourceCode.data.Collect(), lang)

		q, err := sitter.NewQuery(highlights, lang)
		if err != nil {
			log.Panic(err)
		}
		qc := sitter.NewQueryCursor()
		qc.Exec(q, tree)

		// Iterate over query results
		for {
			m, ok := qc.NextMatch()
			if !ok {
				break
			}
			// Apply predicates filtering
			m = qc.FilterPredicates(m, sourceCode.data.Collect())
			for _, c := range m.Captures {
				name := q.CaptureNameForId(c.Index)

				// The most basic of syntax highlighting!
				// TODO. Load these associations from a file
				var color uint8
				switch name {
				case "attribute":
					color = 0x0E
				case "comment":
					color = 0x04
				case "constant.builtin":
					color = 0x09
				case "escape":
					color = 0x0C
				case "function", "function.builtin", "function.method", "function.macro":
					color = 0x0D
				case "keyword":
					color = 0x0E
				case "label":
					color = 0x0C
				case "number":
					color = 0x09
				case "operator":
					color = 0x0C
				case "package":
					color = 0x0D
				case "property":
					color = 0x0D
				case "punctuation.bracket":
					color = 0x05
				case "string", "string.special.path", "string.special.uri":
					color = 0x0B
				case "type", "type.builtin":
					color = 0x0A
				case "variable.member":
					color = 0x0C
				case "variable.parameter":
					color = 0x08
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
		return footer.ShowError(err)
	}

	content, err := io.ReadAll(fd)
	if err != nil {
		return footer.ShowError(err)
	}
	extension := filepath.Ext(path)

	source := SourceCode{}
	source.SetSource(content)

	m.source = &source
	m.viewport.offset = 0
	m.Path = path
	m.fd = fd
	m.modified = false

	return tea.Batch(
		UpdateViewport,
		GenerateSyntaxTree(&source, extension))

}

// Name returns the title of the buffer window to display
func (b Model) Name() string {
	_, name := path.Split(b.Path)
	return name
}

func (source *SourceCode) cursorDown(n int) {
}

func (source *SourceCode) cursorUp(n int) {
}

func (source *SourceCode) cursorLeft(n int) {
	source.cursor = clamp(source.cursor-n, 0, source.data.Len())
}

func (source *SourceCode) cursorRight(n int) {
	source.cursor = clamp(source.cursor+n, 0, source.data.Len())
}

func UpdateViewport() tea.Msg {
	return UpdateViewportMsg(0)
}

// clamp limits the value of val between [low, high)
func clamp(val, low, high int) int {
	return max(low, min(val, high-1))
}
