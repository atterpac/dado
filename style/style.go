// Package style provides a fluent, value-based style builder inspired by
// Charmbracelet's Lip Gloss. It composes alongside dado's theme system rather
// than replacing it.
//
// A Style is an immutable value: every builder method returns a new Style,
// so styles compose by chaining without mutating shared state.
//
//	s := style.New().
//	    Bold().
//	    ForegroundFn(func(t theme.Theme) tcell.Color { return t.Accent() }).
//	    Border(style.RoundedBorder)
//
//	tv.SetText(s.Render("hello"))
//
// Two render modes are supported:
//
//   - Render(string) returns a string with color tags, suitable for
//     TextView content, table cells, and any widget that interprets
//     dynamic colors.
//   - TcellStyle() returns a tcell.Style for direct use in custom Draw().
//
// Apply(*core.Box) is a narrow integration point that copies border, border
// color, and background color settings onto a Box. Layout (width, height,
// alignment) is intentionally NOT modeled — Flex/Grid layouts remain the
// layout engines.
package style

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// Style is an immutable value describing visual decoration.
// Builder methods return new Style values; the zero value is a valid empty style.
type Style struct {
	bold      bool
	italic    bool
	underline bool
	reverse   bool
	blink     bool
	dim       bool

	fg  *colorRef
	bg  *colorRef
	bdr *colorRef // border color, if a border is set

	border *BorderSet // nil = no border

	padX int
	padY int
}

// colorRef holds either a static tcell.Color or a theme-reactive resolver.
type colorRef struct {
	static  tcell.Color
	dynamic func(theme.Theme) tcell.Color
}

func (c *colorRef) resolve() tcell.Color {
	if c == nil {
		return tcell.ColorDefault
	}
	if c.dynamic != nil {
		return c.dynamic(theme.Get())
	}
	return c.static
}

// New returns the zero Style (no decoration).
func New() Style { return Style{} }

// --- Attribute builders ---

// Bold returns a Style with bold text.
func (s Style) Bold() Style { s.bold = true; return s }

// Italic returns a Style with italic text.
func (s Style) Italic() Style { s.italic = true; return s }

// Underline returns a Style with underlined text.
func (s Style) Underline() Style { s.underline = true; return s }

// Reverse returns a Style with reversed fg/bg.
func (s Style) Reverse() Style { s.reverse = true; return s }

// Blink returns a Style with blinking text (terminal support varies).
func (s Style) Blink() Style { s.blink = true; return s }

// Dim returns a Style with dimmed text.
func (s Style) Dim() Style { s.dim = true; return s }

// --- Color builders ---

// Foreground returns a Style with a static foreground color.
func (s Style) Foreground(c tcell.Color) Style {
	s.fg = &colorRef{static: c}
	return s
}

// ForegroundFn returns a Style whose foreground resolves at render time
// from the active theme. The function may be called with a nil Theme;
// callers should handle that case.
func (s Style) ForegroundFn(fn func(theme.Theme) tcell.Color) Style {
	s.fg = &colorRef{dynamic: fn}
	return s
}

// Background returns a Style with a static background color.
func (s Style) Background(c tcell.Color) Style {
	s.bg = &colorRef{static: c}
	return s
}

// BackgroundFn returns a Style whose background resolves at render time
// from the active theme.
func (s Style) BackgroundFn(fn func(theme.Theme) tcell.Color) Style {
	s.bg = &colorRef{dynamic: fn}
	return s
}

// --- Border ---

// Border returns a Style with the given border set. Border colors default
// to the active theme's BorderColor unless overridden via BorderColor.
func (s Style) Border(b BorderSet) Style {
	s.border = &b
	return s
}

// BorderColor returns a Style with a static border color (only meaningful
// when a Border is also set).
func (s Style) BorderColor(c tcell.Color) Style {
	s.bdr = &colorRef{static: c}
	return s
}

// BorderColorFn returns a Style with a theme-reactive border color.
func (s Style) BorderColorFn(fn func(theme.Theme) tcell.Color) Style {
	s.bdr = &colorRef{dynamic: fn}
	return s
}

// --- Padding ---

// PaddingX adds n spaces of horizontal padding inside Render(string) output.
// Has no effect on Apply(*core.Box) — use the Box's own padding for layout.
func (s Style) PaddingX(n int) Style {
	if n < 0 {
		n = 0
	}
	s.padX = n
	return s
}

// PaddingY adds n blank lines above and below Render(string) output.
// Has no effect on Apply(*core.Box).
func (s Style) PaddingY(n int) Style {
	if n < 0 {
		n = 0
	}
	s.padY = n
	return s
}

// --- Render: color tags ---

// Render produces a string wrapped in color tags reflecting this Style.
// User-supplied content is escaped so embedded '[' characters do not open
// unintended tags. PaddingX/PaddingY are applied to the rendered text.
func (s Style) Render(text string) string {
	text = strings.ReplaceAll(text, "[", "[[]")

	if s.padX > 0 {
		pad := strings.Repeat(" ", s.padX)
		text = pad + text + pad
	}
	if s.padY > 0 {
		blank := strings.Repeat("\n", s.padY)
		text = blank + text + blank
	}

	open, close := s.tagPair()
	if open == "" {
		return text
	}
	return open + text + close
}

// tagPair builds the [fg:bg:attrs] open and [-:-:-] close tags.
// Returns ("", "") when the style is empty.
func (s Style) tagPair() (string, string) {
	if s.isEmpty() {
		return "", ""
	}

	var fg, bg, attrs string
	if s.fg != nil {
		fg = colorTag(s.fg.resolve())
	}
	if s.bg != nil {
		bg = colorTag(s.bg.resolve())
	}
	attrs = s.attrTag()

	return "[" + fg + ":" + bg + ":" + attrs + "]", "[-:-:-]"
}

func (s Style) isEmpty() bool {
	return s.fg == nil && s.bg == nil &&
		!s.bold && !s.italic && !s.underline &&
		!s.reverse && !s.blink && !s.dim
}

func (s Style) attrTag() string {
	var b strings.Builder
	if s.bold {
		b.WriteByte('b')
	}
	if s.italic {
		b.WriteByte('i')
	}
	if s.underline {
		b.WriteByte('u')
	}
	if s.reverse {
		b.WriteByte('r')
	}
	if s.blink {
		b.WriteByte('l')
	}
	if s.dim {
		b.WriteByte('d')
	}
	return b.String()
}

// colorTag formats a tcell.Color for use inside a color tag.
// Returns an empty string for ColorDefault so the slot is left blank.
func colorTag(c tcell.Color) string {
	if c == tcell.ColorDefault {
		return ""
	}
	if c.IsRGB() {
		return c.String() // produces "#rrggbb"
	}
	return c.Name()
}

// --- Render: tcell.Style for custom drawers ---

// TcellStyle returns a tcell.Style with this Style's attributes and resolved
// colors applied. Borders and padding are ignored — those only meaningful
// for Render and Apply.
func (s Style) TcellStyle() tcell.Style {
	st := tcell.StyleDefault
	if s.fg != nil {
		st = st.Foreground(s.fg.resolve())
	}
	if s.bg != nil {
		st = st.Background(s.bg.resolve())
	}
	st = st.
		Bold(s.bold).
		Italic(s.italic).
		Underline(s.underline).
		Reverse(s.reverse).
		Blink(s.blink).
		Dim(s.dim)
	return st
}

// --- Apply: integrate with core.Box ---

// Apply copies this Style's border + colors onto a core.Box.
// Padding is intentionally not applied (use Box.SetBorderPadding directly).
// Returns box for chaining.
func (s Style) Apply(box *core.Box) *core.Box {
	if s.bg != nil {
		box.SetBackgroundColor(s.bg.resolve())
	}
	if s.border != nil {
		box.SetBorder(true)
		// Apply rune set via Box.SetBorderStyle when available.
		applyBorderRunes(box, *s.border)
		bdr := s.bdr
		if bdr == nil {
			bdr = &colorRef{static: theme.Border()}
		}
		box.SetBorderStyle(tcell.StyleDefault.Foreground(bdr.resolve()))
	}
	return box
}
