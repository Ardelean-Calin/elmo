package themes

import "github.com/charmbracelet/lipgloss"

type Base16Theme struct {
	Base00 lipgloss.Color // Default Background
	Base01 lipgloss.Color // Lighter Background (Used for status bars, line number and folding marks)
	Base02 lipgloss.Color // Selection Background
	Base03 lipgloss.Color // Comments, Invisibles, Line Highlighting
	Base04 lipgloss.Color // Dark Foreground (Used for status bars)
	Base05 lipgloss.Color // Default Foreground, Caret, Delimiters, Operators
	Base06 lipgloss.Color // Light Foreground (Not often used)
	Base07 lipgloss.Color // Light Background (Not often used)
	Base08 lipgloss.Color // Variables, XML Tags, Markup Link Text, Markup Lists, Diff Deleted
	Base09 lipgloss.Color // Integers, Boolean, Constants, XML Attributes, Markup Link Url
	Base0A lipgloss.Color // Classes, Markup Bold, Search Text Background
	Base0B lipgloss.Color // Strings, Inherited Class, Markup Code, Diff Inserted
	Base0C lipgloss.Color // Support, Regular Expressions, Escape Characters, Markup Quotes
	Base0D lipgloss.Color // Functions, Methods, Attribute IDs, Headings
	Base0E lipgloss.Color // Keywords, Storage, Selector, Markup Italic, Diff Changed
	Base0F lipgloss.Color // Deprecated, Opening/Closing Embedded Language Tags, e.g. <?php ?>

}

func DefaultTheme() Base16Theme {
	return Base16Theme{
		Base00: lipgloss.Color("#1d2021"),
		Base01: lipgloss.Color("#3c3836"),
		Base02: lipgloss.Color("#504945"),
		Base03: lipgloss.Color("#665c54"),
		Base04: lipgloss.Color("#bdae93"),
		Base05: lipgloss.Color("#d5c4a1"),
		Base06: lipgloss.Color("#ebdbb2"),
		Base07: lipgloss.Color("#fbf1c7"),
		Base08: lipgloss.Color("#fb4934"),
		Base09: lipgloss.Color("#fe8019"),
		Base0A: lipgloss.Color("#fabd2f"),
		Base0B: lipgloss.Color("#b8bb26"),
		Base0C: lipgloss.Color("#8ec07c"),
		Base0D: lipgloss.Color("#83a598"),
		Base0E: lipgloss.Color("#d3869b"),
		Base0F: lipgloss.Color("#d65d0e"),
	}
}
