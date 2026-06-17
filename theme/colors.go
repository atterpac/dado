package theme

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Convenience color getters forward to the default provider's accessors
// (theme/provider_colors.go), which read the active theme at call time for
// live switching and apply the package fallbacks when no theme is set.

// Base colors

func Bg() tcell.Color      { return defaultProvider.Bg() }
func BgLight() tcell.Color { return defaultProvider.BgLight() }
func BgDark() tcell.Color  { return defaultProvider.BgDark() }
func Fg() tcell.Color      { return defaultProvider.Fg() }
func FgDim() tcell.Color   { return defaultProvider.FgDim() }
func FgMuted() tcell.Color { return defaultProvider.FgMuted() }

// Accent colors

func Accent() tcell.Color    { return defaultProvider.Accent() }
func AccentDim() tcell.Color { return defaultProvider.AccentDim() }
func Highlight() tcell.Color { return defaultProvider.Highlight() }

// Semantic colors

func Success() tcell.Color { return defaultProvider.Success() }
func Warning() tcell.Color { return defaultProvider.Warning() }
func Error() tcell.Color   { return defaultProvider.Error() }
func Info() tcell.Color    { return defaultProvider.Info() }

// Border colors

func Border() tcell.Color      { return defaultProvider.Border() }
func BorderFocus() tcell.Color { return defaultProvider.BorderFocus() }

// UI element colors

func Header() tcell.Color      { return defaultProvider.Header() }
func Menu() tcell.Color        { return defaultProvider.Menu() }
func TableHeader() tcell.Color { return defaultProvider.TableHeader() }
func Key() tcell.Color         { return defaultProvider.Key() }
func Crumb() tcell.Color       { return defaultProvider.Crumb() }
func PanelBorder() tcell.Color { return defaultProvider.PanelBorder() }
func PanelTitle() tcell.Color  { return defaultProvider.PanelTitle() }

// Tag color functions for color tags [#hexcolor]text[-]

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

// ColorToHex converts tcell.Color to "#RRGGBB" string for color tags.
func ColorToHex(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// Lerp linearly interpolates between two colors in RGB space. t is clamped to
// [0,1]: t=0 returns a, t=1 returns b. This is the building block for fades,
// gradients and glow/pulse animations.
//
// Interpolation is per-channel in sRGB, which is cheap and good enough for UI
// transitions; it is not perceptually uniform.
func Lerp(a, b tcell.Color, t float64) tcell.Color {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	ar, ag, ab := a.RGB()
	br, bg, bb := b.RGB()
	lerp := func(x, y int32) int32 { return x + int32(float64(y-x)*t) }
	return tcell.NewRGBColor(lerp(ar, br), lerp(ag, bg), lerp(ab, bb))
}

// Gradient interpolates across an ordered list of stops, mapping t in [0,1]
// onto the full sequence. With <2 stops it returns the single stop (or the
// fallback Fg when empty). Useful for multi-color shimmer and heat maps.
func Gradient(t float64, stops ...tcell.Color) tcell.Color {
	switch len(stops) {
	case 0:
		return Fg()
	case 1:
		return stops[0]
	}
	if t <= 0 {
		return stops[0]
	}
	if t >= 1 {
		return stops[len(stops)-1]
	}
	scaled := t * float64(len(stops)-1)
	i := int(scaled)
	return Lerp(stops[i], stops[i+1], scaled-float64(i))
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
