package theme

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Convenience color getters that wrap Get().Color()
// These read from active theme at call time for live switching.

// Base colors

func Bg() tcell.Color {
	if t := Get(); t != nil {
		return t.Bg()
	}
	return tcell.ColorDefault
}

func BgLight() tcell.Color {
	if t := Get(); t != nil {
		return t.BgLight()
	}
	return tcell.ColorDefault
}

func BgDark() tcell.Color {
	if t := Get(); t != nil {
		return t.BgDark()
	}
	return tcell.ColorDefault
}

func Fg() tcell.Color {
	if t := Get(); t != nil {
		return t.Fg()
	}
	return tcell.ColorDefault
}

func FgDim() tcell.Color {
	if t := Get(); t != nil {
		return t.FgDim()
	}
	return tcell.ColorGray
}

func FgMuted() tcell.Color {
	if t := Get(); t != nil {
		return t.FgMuted()
	}
	return tcell.ColorDarkGray
}

// Accent colors

func Accent() tcell.Color {
	if t := Get(); t != nil {
		return t.Accent()
	}
	return tcell.ColorBlue
}

func AccentDim() tcell.Color {
	if t := Get(); t != nil {
		return t.AccentDim()
	}
	return tcell.ColorBlue
}

func Highlight() tcell.Color {
	if t := Get(); t != nil {
		return t.Highlight()
	}
	return tcell.ColorYellow
}

// Semantic colors

func Success() tcell.Color {
	if t := Get(); t != nil {
		return t.Success()
	}
	return tcell.ColorGreen
}

func Warning() tcell.Color {
	if t := Get(); t != nil {
		return t.Warning()
	}
	return tcell.ColorYellow
}

func Error() tcell.Color {
	if t := Get(); t != nil {
		return t.Error()
	}
	return tcell.ColorRed
}

func Info() tcell.Color {
	if t := Get(); t != nil {
		return t.Info()
	}
	return tcell.ColorBlue
}

// Border colors

func Border() tcell.Color {
	if t := Get(); t != nil {
		return t.Border()
	}
	return tcell.ColorGray
}

func BorderFocus() tcell.Color {
	if t := Get(); t != nil {
		return t.BorderFocus()
	}
	return tcell.ColorWhite
}

// UI element colors

func Header() tcell.Color {
	if t := Get(); t != nil {
		return t.Header()
	}
	return tcell.ColorDefault
}

func Menu() tcell.Color {
	if t := Get(); t != nil {
		return t.Menu()
	}
	return tcell.ColorDefault
}

func TableHeader() tcell.Color {
	if t := Get(); t != nil {
		return t.TableHeader()
	}
	return tcell.ColorBlue
}

func Key() tcell.Color {
	if t := Get(); t != nil {
		return t.Key()
	}
	return tcell.ColorBlue
}

func Crumb() tcell.Color {
	if t := Get(); t != nil {
		return t.Crumb()
	}
	return tcell.ColorBlue
}

func PanelBorder() tcell.Color {
	if t := Get(); t != nil {
		return t.PanelBorder()
	}
	return tcell.ColorGray
}

func PanelTitle() tcell.Color {
	if t := Get(); t != nil {
		return t.PanelTitle()
	}
	return tcell.ColorBlue
}

// Tag color functions for tview color tags [#hexcolor]text[-]

// Base colors
func TagBg() string      { return ColorToHex(Bg()) }
func TagBgLight() string { return ColorToHex(BgLight()) }
func TagBgDark() string  { return ColorToHex(BgDark()) }
func TagFg() string      { return ColorToHex(Fg()) }
func TagFgDim() string   { return ColorToHex(FgDim()) }
func TagFgMuted() string { return ColorToHex(FgMuted()) }

// Accent colors
func TagAccent() string    { return ColorToHex(Accent()) }
func TagAccentDim() string { return ColorToHex(AccentDim()) }
func TagHighlight() string { return ColorToHex(Highlight()) }

// Semantic colors
func TagSuccess() string { return ColorToHex(Success()) }
func TagWarning() string { return ColorToHex(Warning()) }
func TagError() string   { return ColorToHex(Error()) }
func TagInfo() string    { return ColorToHex(Info()) }

// Border colors
func TagBorder() string      { return ColorToHex(Border()) }
func TagBorderFocus() string { return ColorToHex(BorderFocus()) }

// UI element colors
func TagHeader() string      { return ColorToHex(Header()) }
func TagMenu() string        { return ColorToHex(Menu()) }
func TagTableHeader() string { return ColorToHex(TableHeader()) }
func TagKey() string         { return ColorToHex(Key()) }
func TagCrumb() string       { return ColorToHex(Crumb()) }
func TagPanelBorder() string { return ColorToHex(PanelBorder()) }
func TagPanelTitle() string  { return ColorToHex(PanelTitle()) }

// ColorToHex converts tcell.Color to "#RRGGBB" string for tview tags.
func ColorToHex(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Selection colors - high contrast for readability

// SelectionBg returns the background color for selected items.
func SelectionBg() tcell.Color {
	return Accent()
}

// SelectionFg returns the foreground color for selected items.
// Uses black for high contrast against the accent background.
func SelectionFg() tcell.Color {
	return tcell.ColorBlack
}

// SelectionStyle returns a tcell style for selected items with high contrast.
func SelectionStyle() tcell.Style {
	return tcell.StyleDefault.Background(SelectionBg()).Foreground(SelectionFg()).Bold(true)
}

// InactiveSelectionStyle returns a style for selected items in inactive panes.
func InactiveSelectionStyle() tcell.Style {
	return tcell.StyleDefault.Background(Bg()).Foreground(Fg())
}

// ConfigureList sets up a tview.List with proper theme colors including
// high-contrast selection styling.
func ConfigureList(list *tview.List) *tview.List {
	bg := Bg()
	fg := Fg()
	fgDim := FgDim()

	list.SetBackgroundColor(bg)
	list.SetMainTextColor(fg)
	list.SetMainTextStyle(tcell.StyleDefault.Background(bg).Foreground(fg))
	list.SetSecondaryTextColor(fgDim)
	list.SetSecondaryTextStyle(tcell.StyleDefault.Background(bg).Foreground(fgDim))
	list.SetSelectedBackgroundColor(SelectionBg())
	list.SetSelectedTextColor(SelectionFg())
	list.SetSelectedStyle(SelectionStyle())
	list.SetHighlightFullLine(true)

	return list
}

// ConfigureListInactive sets a list to show as inactive (no selection highlight).
func ConfigureListInactive(list *tview.List) *tview.List {
	bg := Bg()
	fg := Fg()

	list.SetSelectedBackgroundColor(bg)
	list.SetSelectedTextColor(fg)
	list.SetSelectedStyle(InactiveSelectionStyle())

	return list
}

// ConfigureListActive sets a list to show as active (with selection highlight).
func ConfigureListActive(list *tview.List) *tview.List {
	list.SetSelectedBackgroundColor(SelectionBg())
	list.SetSelectedTextColor(SelectionFg())
	list.SetSelectedStyle(SelectionStyle())

	return list
}
