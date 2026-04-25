package style

import "github.com/rivo/tview"

// BorderSet describes the runes used to draw a border. Layout is unaffected;
// only the glyphs change.
type BorderSet struct {
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	Horizontal  rune
	Vertical    rune
}

// Common border presets.
var (
	NormalBorder = BorderSet{
		TopLeft: '┌', TopRight: '┐',
		BottomLeft: '└', BottomRight: '┘',
		Horizontal: '─', Vertical: '│',
	}
	RoundedBorder = BorderSet{
		TopLeft: '╭', TopRight: '╮',
		BottomLeft: '╰', BottomRight: '╯',
		Horizontal: '─', Vertical: '│',
	}
	ThickBorder = BorderSet{
		TopLeft: '┏', TopRight: '┓',
		BottomLeft: '┗', BottomRight: '┛',
		Horizontal: '━', Vertical: '┃',
	}
	DoubleBorder = BorderSet{
		TopLeft: '╔', TopRight: '╗',
		BottomLeft: '╚', BottomRight: '╝',
		Horizontal: '═', Vertical: '║',
	}
	BlockBorder = BorderSet{
		TopLeft: '█', TopRight: '█',
		BottomLeft: '█', BottomRight: '█',
		Horizontal: '█', Vertical: '█',
	}
)

// applyBorderRunes is currently a no-op: tview.Box draws borders using the
// global tview.Borders rune set, which is not per-Box customizable. BorderSet
// values remain exposed so custom Draw() implementations can render their
// own borders with the chosen runes.
func applyBorderRunes(_ *tview.Box, _ BorderSet) {}
