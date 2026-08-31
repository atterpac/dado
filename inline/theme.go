package inline

import "github.com/gdamore/tcell/v2"

// BorderSet contains the glyphs used to draw a rectangular control.
type BorderSet struct{ TopLeft, TopRight, BottomLeft, BottomRight, Horizontal, Vertical string }

// FormGlyphs contains semantic form glyphs.
type FormGlyphs struct{ Focus, Required, Dropdown, Selected, Unselected, Checked, Unchecked, Disabled, Error, Success string }

// TreeGlyphs contains the guides used to show tree hierarchy.
type TreeGlyphs struct{ Branch, Last, Vertical, Space string }

// TableGlyphs contains the junctions used to connect table borders.
type TableGlyphs struct{ Top, Left, Middle, Right, Bottom string }

// InlineTheme styles renderer-native forms and carries the StatusTheme used by
// activity components.
type InlineTheme struct {
	Text, Muted, Label, Accent, Border, FocusedBorder, Error, Success tcell.Style
	Borders                                                           BorderSet
	Glyphs                                                            FormGlyphs
	Tree                                                              TreeGlyphs
	Table                                                             TableGlyphs
	Status                                                            StatusTheme
	FieldGap                                                          int
}

// RoundedInlineTheme is the colorful default playground preset.
func RoundedInlineTheme() InlineTheme {
	accent := tcell.NewHexColor(0x9b87f5)
	muted := tcell.NewHexColor(0x71717a)
	status := DefaultStatusTheme()
	status.ActiveStyle = tcell.StyleDefault.Foreground(accent)
	return InlineTheme{
		Text: tcell.StyleDefault, Muted: tcell.StyleDefault.Foreground(muted), Label: tcell.StyleDefault,
		Accent: tcell.StyleDefault.Foreground(accent), Border: tcell.StyleDefault.Foreground(muted), FocusedBorder: tcell.StyleDefault.Foreground(accent).Bold(true),
		Error: tcell.StyleDefault.Foreground(tcell.NewHexColor(0xfb7185)), Success: tcell.StyleDefault.Foreground(tcell.NewHexColor(0x4ade80)),
		Borders: BorderSet{"╭", "╮", "╰", "╯", "─", "│"},
		Glyphs:  FormGlyphs{Focus: "›", Required: "*", Dropdown: "▾", Checked: "✓", Unchecked: "○", Disabled: "–", Error: "!", Success: "✓"}, Status: status, FieldGap: 1,
		Tree:  TreeGlyphs{Branch: "├─", Last: "└─", Vertical: "│ ", Space: "  "},
		Table: TableGlyphs{Top: "┬", Left: "├", Middle: "┼", Right: "┤", Bottom: "┴"},
	}
}

// SquareInlineTheme uses crisp square borders and a cyan accent.
func SquareInlineTheme() InlineTheme {
	theme := RoundedInlineTheme()
	accent := tcell.ColorAqua
	theme.Accent = tcell.StyleDefault.Foreground(accent)
	theme.FocusedBorder = theme.Accent.Bold(true)
	theme.Status.ActiveStyle = theme.Accent
	theme.Borders = BorderSet{"┌", "┐", "└", "┘", "─", "│"}
	theme.Glyphs.Focus = ">"
	theme.Table = TableGlyphs{Top: "┬", Left: "├", Middle: "┼", Right: "┤", Bottom: "┴"}
	return theme
}

// ASCIIInlineTheme avoids Unicode glyphs while retaining semantic color.
func ASCIIInlineTheme() InlineTheme {
	theme := SquareInlineTheme()
	theme.Borders = BorderSet{"+", "+", "+", "+", "-", "|"}
	theme.Glyphs = FormGlyphs{Focus: ">", Required: "*", Dropdown: "v", Checked: "x", Unchecked: ".", Disabled: "-", Error: "!", Success: "+"}
	theme.Tree = TreeGlyphs{Branch: "|-", Last: "`-", Vertical: "| ", Space: "  "}
	theme.Table = TableGlyphs{Top: "+", Left: "+", Middle: "+", Right: "+", Bottom: "+"}
	theme.Status.PendingMarker = "o"
	theme.Status.ActiveMarker = ">"
	theme.Status.SuccessMarker = "+"
	theme.Status.FailureMarker = "x"
	theme.Status.SkippedMarker = "-"
	theme.Status.CancelledMarker = "!"
	theme.Status.BarFilled = "#"
	theme.Status.BarEmpty = "-"
	theme.Status.BarPartial = []string{"#"}
	theme.Status.SpinnerFrames = []string{"-", "\\", "|", "/"}
	return theme
}

func normalizedInlineTheme(theme InlineTheme) InlineTheme {
	defaults := RoundedInlineTheme()
	if theme.Borders.Horizontal == "" {
		theme.Borders = defaults.Borders
	}
	if theme.Glyphs.Focus == "" {
		theme.Glyphs = defaults.Glyphs
	}
	if theme.Tree.Branch == "" {
		theme.Tree = defaults.Tree
	}
	if theme.Table.Top == "" {
		theme.Table.Top = defaults.Table.Top
	}
	if theme.Table.Left == "" {
		theme.Table.Left = defaults.Table.Left
	}
	if theme.Table.Middle == "" {
		theme.Table.Middle = defaults.Table.Middle
	}
	if theme.Table.Right == "" {
		theme.Table.Right = defaults.Table.Right
	}
	if theme.Table.Bottom == "" {
		theme.Table.Bottom = defaults.Table.Bottom
	}
	if theme.FieldGap < 0 {
		theme.FieldGap = 0
	}
	theme.Status = normalizedStatusTheme(theme.Status)
	return theme
}
