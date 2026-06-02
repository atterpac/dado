package core

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

// Span is a run of text sharing a single style.
type Span struct {
	Text  string
	Style tcell.Style
}

// Text is a styled string built from one or more spans. Build it once (directly
// or via ParseTagged) and draw it many times — Draw does no parsing and Width is
// O(1), unlike PrintTagged which re-scans markup on every frame.
//
// Widths are measured in runes (one column per rune), matching the convention of
// the other screen-drawing helpers (PrintAt, PrintClipped, PrintTagged).
type Text struct {
	spans []Span
	width int // cached visible width in columns (runes)
}

// NewText returns an empty Text ready for Append.
func NewText() *Text { return &Text{} }

// Append adds a styled span and returns the Text for chaining.
// Empty strings are ignored so callers can append conditionally without guards.
func (t *Text) Append(s string, style tcell.Style) *Text {
	if s == "" {
		return t
	}
	t.spans = append(t.spans, Span{Text: s, Style: style})
	t.width += utf8.RuneCountInString(s)
	return t
}

// Spans returns the underlying spans. The slice is shared; do not mutate.
func (t *Text) Spans() []Span { return t.spans }

// Width returns the visible width in columns. O(1).
func (t *Text) Width() int { return t.width }

// Empty reports whether the Text has no visible content.
func (t *Text) Empty() bool { return t.width == 0 }

// String returns the concatenated text with styling stripped.
func (t *Text) String() string {
	var b strings.Builder
	for _, sp := range t.spans {
		b.WriteString(sp.Text)
	}
	return b.String()
}

// Draw renders the Text at (x, y), clipped to maxWidth columns, left-aligned.
// Returns the number of columns written. No allocation, no parsing.
func (t *Text) Draw(screen tcell.Screen, x, y, maxWidth int) int {
	return t.DrawFunc(screen, x, y, maxWidth, nil)
}

// DrawFunc is Draw with a per-span style transform applied at draw time. overlay
// receives each span's own style and returns the style to render with; a nil
// overlay draws spans as-is. This lets a cached Text be drawn under varying
// per-draw state (selection reverse, hover background, theme tweaks) without
// re-parsing. The transform composes on top of each span's style, so additive
// changes (e.g. s.Reverse(true)) preserve a span's own colors; setting a color
// the span may have set itself will override it.
func (t *Text) DrawFunc(screen tcell.Screen, x, y, maxWidth int, overlay func(tcell.Style) tcell.Style) int {
	if maxWidth <= 0 {
		return 0
	}
	col := 0
	for _, sp := range t.spans {
		style := sp.Style
		if overlay != nil {
			style = overlay(style)
		}
		for _, r := range sp.Text {
			if col >= maxWidth {
				return col
			}
			screen.SetContent(x+col, y, r, nil, style)
			col++
		}
	}
	return col
}

// Wrap splits the Text into lines no wider than width columns, preserving styles.
// Existing '\n' runes are treated as hard breaks. Lines are broken at spaces when
// possible; a single word longer than width is hard-broken. A width <= 0 returns
// the Text as a single line.
func (t *Text) Wrap(width int) []*Text {
	if width <= 0 || t.width <= width && !strings.Contains(t.String(), "\n") {
		return []*Text{t}
	}

	// Flatten to styled runes so wrapping can break anywhere regardless of spans.
	runes := make([]styledRune, 0, t.width)
	for _, sp := range t.spans {
		for _, r := range sp.Text {
			runes = append(runes, styledRune{r, sp.Style})
		}
	}

	var lines []*Text
	col := 0
	lastSpace := -1 // index into pending where the last space sits

	// pending holds the runes of the current line so we can re-split at lastSpace.
	pending := make([]styledRune, 0, width)

	emit := func(runs []styledRune) {
		line := NewText()
		appendRuns(line, runs)
		lines = append(lines, line)
	}

	for _, c := range runes {
		if c.r == '\n' {
			emit(pending)
			pending = pending[:0]
			col = 0
			lastSpace = -1
			continue
		}
		pending = append(pending, c)
		if c.r == ' ' {
			lastSpace = len(pending) - 1
		}
		col++
		if col > width {
			if c.r == ' ' {
				// The overflowing rune is itself a space: drop it as the break.
				emit(pending[:len(pending)-1])
				pending = pending[:0]
				col = 0
				lastSpace = -1
			} else if lastSpace >= 0 && lastSpace < len(pending)-1 {
				// Break at the last space: emit up to it, carry the rest over.
				carry := append([]styledRune(nil), pending[lastSpace+1:]...)
				emit(pending[:lastSpace]) // drop the trailing space too
				pending = carry
				col = len(pending)
				lastSpace = -1
			} else {
				// No usable break point — hard break before the overflowing rune.
				overflow := pending[len(pending)-1]
				emit(pending[:len(pending)-1])
				pending = append(pending[:0], overflow)
				col = 1
				lastSpace = -1
			}
		}
	}
	emit(pending)
	return lines
}

// splitWidth hard-wraps the Text into lines of at most width columns, breaking
// strictly at the column boundary (no word awareness), preserving styles. Used
// for fixed-column rendering where word wrap is not wanted. Does not interpret
// '\n' — split on newlines before calling.
func (t *Text) splitWidth(width int) []*Text {
	if width <= 0 || t.width <= width {
		return []*Text{t}
	}
	runes := make([]styledRune, 0, t.width)
	for _, sp := range t.spans {
		for _, r := range sp.Text {
			runes = append(runes, styledRune{r, sp.Style})
		}
	}
	var lines []*Text
	for len(runes) > 0 {
		chunk := runes
		if len(chunk) > width {
			chunk = runes[:width]
		}
		runes = runes[len(chunk):]
		line := NewText()
		appendRuns(line, chunk)
		lines = append(lines, line)
	}
	return lines
}

// styledRune is a single rune with its style, used during wrapping.
type styledRune struct {
	r     rune
	style tcell.Style
}

// appendRuns coalesces consecutive same-style runes into spans on t.
func appendRuns(t *Text, runes []styledRune) {
	if len(runes) == 0 {
		return
	}
	var b strings.Builder
	cur := runes[0].style
	for _, c := range runes {
		if c.style != cur {
			t.Append(b.String(), cur)
			b.Reset()
			cur = c.style
		}
		b.WriteRune(c.r)
	}
	t.Append(b.String(), cur)
}

// ParseTagged compiles a markup string ([#rrggbb], [-], [::b], named colors, …)
// into a Text once, so it can be drawn repeatedly without re-parsing. baseStyle
// is the style for text before any tag. This is the parse-once counterpart to
// PrintTagged — prefer it for content that is drawn every frame.
func ParseTagged(markup string, baseStyle tcell.Style) *Text {
	t := NewText()
	style := baseStyle
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			t.Append(b.String(), style)
			b.Reset()
		}
	}
	rest := markup
	for len(rest) > 0 {
		if rest[0] == '[' {
			if end := strings.Index(rest, "]"); end > 0 {
				if newStyle, ok := parseTag(rest[1:end], style, baseStyle); ok {
					flush()
					style = newStyle
					rest = rest[end+1:]
					continue
				}
			}
		}
		r, size := utf8.DecodeRuneInString(rest)
		b.WriteRune(r)
		rest = rest[size:]
	}
	flush()
	return t
}
