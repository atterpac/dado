package core

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// TextView displays multi-line scrollable text with optional ANSI color tag parsing.
type TextView struct {
	Box
	text       string
	lines      []*Text // cached wrapped lines (styled spans)
	scrollY    int
	scrollX    int
	scrollable bool
	dynamic    bool        // parse [color] tags
	wrapWidth  int         // last width lines were computed for
	fgColor    tcell.Color // foreground override (ColorDefault = use style/default)
	textStyle  tcell.Style // base style for plain text
	textAlign  int         // AlignLeft/AlignCenter/AlignRight
}

// NewTextView returns a ready TextView.
func NewTextView() *TextView { return &TextView{textAlign: AlignLeft} }

// SetText sets the displayed text and resets scroll.
func (tv *TextView) SetText(s string) *TextView {
	tv.text = s
	tv.lines = nil // invalidate cache
	tv.scrollY = 0
	return tv
}

// SetScrollable enables or disables keyboard/programmatic scrolling.
func (tv *TextView) SetScrollable(on bool) *TextView { tv.scrollable = on; return tv }

// SetDynamicColors enables [color] tag parsing.
func (tv *TextView) SetDynamicColors(on bool) *TextView { tv.dynamic = on; return tv }

// GetText returns the stored text.
func (tv *TextView) GetText() string { return tv.text }

// SetTextColor sets the foreground color for plain text rendering.
func (tv *TextView) SetTextColor(c tcell.Color) *TextView { tv.fgColor = c; return tv }

// SetTextAlign sets the horizontal text alignment (AlignLeft, AlignCenter, AlignRight).
func (tv *TextView) SetTextAlign(align int) *TextView { tv.textAlign = align; return tv }

// SetTextStyle sets the base tcell style used for plain text segments.
func (tv *TextView) SetTextStyle(style tcell.Style) *TextView { tv.textStyle = style; return tv }

// SetWordWrap is a no-op stub; the view always wraps at the widget boundary.
func (tv *TextView) SetWordWrap(_ bool) *TextView { return tv }

// SetRegions is a no-op stub; region tags are not implemented.
func (tv *TextView) SetRegions(_ bool) *TextView { return tv }

// GetScrollOffset returns the current scroll position (row, col).
func (tv *TextView) GetScrollOffset() (row, col int) { return tv.scrollY, tv.scrollX }

// ScrollTo sets the scroll position.
func (tv *TextView) ScrollTo(row, col int) { tv.scrollY = row; tv.scrollX = col }

// Draw renders the text into the inner rect.
func (tv *TextView) Draw(screen tcell.Screen) {
	tv.Box.Draw(screen)
	vp := NewViewport(tv.InnerRect())
	if vp.Empty() {
		return
	}
	w, _ := vp.Size()
	if len(tv.lines) == 0 || tv.wrapWidth != w {
		tv.reflow(w)
	}
	vp.SetContentSize(w, len(tv.lines))
	vp.SetOffset(0, tv.scrollY)

	baseStyle := tv.textStyle
	if tv.fgColor != tcell.ColorDefault {
		baseStyle = baseStyle.Foreground(tv.fgColor)
	}
	clearStyle := tcell.StyleDefault.Background(tv.Box.bg())

	first, last := vp.VisibleRows()
	for lineIdx := first; lineIdx < last; lineIdx++ {
		// Compute line width for alignment
		lineRunes := tv.lines[lineIdx].Width()
		startCol := 0
		switch tv.textAlign {
		case AlignCenter:
			startCol = (w - lineRunes) / 2
		case AlignRight:
			startCol = w - lineRunes
		}
		if startCol < 0 {
			startCol = 0
		}
		col := 0
		// fill leading space for alignment
		for ; col < startCol; col++ {
			vp.SetContent(screen, col, lineIdx, ' ', clearStyle)
		}
		for _, seg := range tv.lines[lineIdx].Spans() {
			segStyle := seg.Style
			// Apply base fg if segment is using default foreground
			fg, bg, _ := segStyle.Decompose()
			if fg == tcell.ColorDefault {
				segStyle = baseStyle
			}
			// Any segment without an explicit background follows the live
			// box background so colored text tracks theme switches too.
			if bg == tcell.ColorDefault {
				bg = tv.Box.bg()
			}
			segStyle = segStyle.Background(bg)
			for _, r := range seg.Text {
				if col >= w {
					break
				}
				vp.SetContent(screen, col, lineIdx, r, segStyle)
				col++
			}
		}
		// fill rest of row with background
		for ; col < w; col++ {
			vp.SetContent(screen, col, lineIdx, ' ', clearStyle)
		}
	}
}

// reflow wraps tv.text into lines of maxWidth, parsing color tags if dynamic.
func (tv *TextView) reflow(maxWidth int) {
	tv.wrapWidth = maxWidth
	tv.lines = nil

	rawLines := strings.Split(tv.text, "\n")
	for _, raw := range rawLines {
		if tv.dynamic {
			tv.lines = append(tv.lines, tv.parseColorLine(raw, maxWidth)...)
		} else {
			tv.lines = append(tv.lines, tv.wrapPlain(raw, maxWidth)...)
		}
	}
}

func (tv *TextView) wrapPlain(text string, width int) []*Text {
	return NewText().Append(text, tcell.StyleDefault).splitWidth(width)
}

// parseColorLine parses a line that may contain [color] tags into styled spans,
// then hard-wraps it to width. Supports: [#rrggbb], [red], [-], [fg:bg:attrs],
// [::b], [-:-:-], etc. Tag parsing and span merging are handled by ParseTagged.
func (tv *TextView) parseColorLine(text string, width int) []*Text {
	return ParseTagged(text, tcell.StyleDefault).splitWidth(width)
}

// parseTag parses a [color] tag and returns the updated style.
// Tag formats: [color], [fg:bg:attrs], [-], [-:-:-], [::b], [#rrggbb], [#rrggbb::b]
// Returns (newStyle, true) if the tag was recognized, (_, false) if not a color tag.
func parseTag(tag string, current, base tcell.Style) (tcell.Style, bool) {
	// Fast path: full reset shortcuts. Reset to the caller's base style (e.g. a
	// table row's selection background), not the terminal default, so a "[-]"
	// inside a styled region doesn't punch a hole in that styling.
	if tag == "-" || tag == "" {
		return base, true
	}
	// Full reset [-:-:-] or similar all-dash forms
	if tag == "-:-:-" || tag == "-:-" {
		return base, true
	}

	// Split into up to 3 parts: fg:bg:attrs
	parts := strings.SplitN(tag, ":", 3)
	fgPart := parts[0]
	bgPart := ""
	attrPart := ""
	if len(parts) >= 2 {
		bgPart = parts[1]
	}
	if len(parts) >= 3 {
		attrPart = parts[2]
	}

	// If all parts are empty or dash with no valid color, not a tag
	// (avoids matching things like [0] which are not color tags)
	fgColor, fgOK := resolveColor(fgPart)
	bgColor, bgOK := resolveColor(bgPart)

	// Must have at least one recognized part to be a color tag
	if !fgOK && !bgOK && attrPart == "" {
		return current, false
	}

	style := current
	// An empty part means "unspecified": keep the current color. Only "-" and a
	// real color change anything; without this guard a fg-only tag like
	// "[#88ccff]" would resolve its empty bg part to ColorDefault and wipe the
	// inherited background (e.g. a selected row's highlight).
	if fgPart == "-" || fgPart == "" {
		// keep current fg
	} else if fgOK {
		style = style.Foreground(fgColor)
	}
	if bgPart == "-" || bgPart == "" {
		// keep current bg
	} else if bgOK {
		style = style.Background(bgColor)
	}

	// Parse attributes: b=bold, u=underline, i=italic, s=strikethrough, d=dim, r=reverse, l=blink
	for _, a := range attrPart {
		switch a {
		case 'b':
			style = style.Bold(true)
		case 'u':
			style = style.Underline(true)
		case 'i':
			style = style.Italic(true)
		case 'r':
			style = style.Reverse(true)
		case 'd':
			style = style.Dim(true)
		case 'l':
			style = style.Blink(true)
		case '-':
			// reset attrs — but we can't reset individual attrs in tcell without rebuilding
			fg, bg, _ := style.Decompose()
			style = tcell.StyleDefault.Foreground(fg).Background(bg)
		}
	}

	return style, true
}

// resolveColor parses a color string: "#rrggbb", named color, "-" (keep), "" (default).
// Returns (color, true) if recognized as a color token, (_, false) if not.
func resolveColor(s string) (tcell.Color, bool) {
	switch strings.ToLower(s) {
	case "":
		return tcell.ColorDefault, true
	case "-":
		return tcell.ColorDefault, false // "-" means keep — caller handles it
	case "default":
		return tcell.ColorDefault, true
	case "red":
		return tcell.ColorRed, true
	case "green":
		return tcell.ColorGreen, true
	case "blue":
		return tcell.ColorBlue, true
	case "yellow":
		return tcell.ColorYellow, true
	case "white":
		return tcell.ColorWhite, true
	case "black":
		return tcell.ColorBlack, true
	case "cyan":
		return tcell.ColorTeal, true
	case "magenta":
		return tcell.ColorFuchsia, true
	case "orange":
		return tcell.ColorOrange, true
	case "purple":
		return tcell.ColorPurple, true
	case "gray", "grey":
		return tcell.ColorGray, true
	case "darkgray", "darkgrey":
		return tcell.ColorDarkGray, true
	}
	// Hex: #rrggbb or #rgb
	if len(s) == 7 && s[0] == '#' {
		r, g, b, ok := parseHex6(s[1:])
		if ok {
			return tcell.NewRGBColor(int32(r), int32(g), int32(b)), true
		}
	}
	if len(s) == 4 && s[0] == '#' {
		r, g, b, ok := parseHex3(s[1:])
		if ok {
			return tcell.NewRGBColor(int32(r), int32(g), int32(b)), true
		}
	}
	return tcell.ColorDefault, false
}

func parseHex6(s string) (r, g, b uint8, ok bool) {
	ri, e1 := hexByte(s[0], s[1])
	gi, e2 := hexByte(s[2], s[3])
	bi, e3 := hexByte(s[4], s[5])
	if e1 || e2 || e3 {
		return 0, 0, 0, false
	}
	return ri, gi, bi, true
}

func parseHex3(s string) (r, g, b uint8, ok bool) {
	rv, e1 := hexNibble(s[0])
	gv, e2 := hexNibble(s[1])
	bv, e3 := hexNibble(s[2])
	if e1 || e2 || e3 {
		return 0, 0, 0, false
	}
	return rv * 17, gv * 17, bv * 17, true
}

func hexByte(hi, lo byte) (uint8, bool) {
	h, e1 := hexNibble(hi)
	l, e2 := hexNibble(lo)
	return h<<4 | l, e1 || e2
}

func hexNibble(c byte) (uint8, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', false
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, false
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, false
	}
	return 0, true
}

// HandleKey implements KeyHandler for scrolling.
func (tv *TextView) HandleKey(ev *tcell.EventKey) bool {
	if !tv.scrollable {
		return false
	}
	_, _, _, h := tv.InnerRect()
	switch ev.Key() {
	case tcell.KeyDown:
		tv.scrollY++
		return true
	case tcell.KeyUp:
		if tv.scrollY > 0 {
			tv.scrollY--
		}
		return true
	case tcell.KeyPgDn:
		tv.scrollY += h
		return true
	case tcell.KeyPgUp:
		tv.scrollY -= h
		if tv.scrollY < 0 {
			tv.scrollY = 0
		}
		return true
	case tcell.KeyHome:
		tv.scrollY = 0
		return true
	case tcell.KeyEnd:
		tv.scrollY = len(tv.lines) - h
		if tv.scrollY < 0 {
			tv.scrollY = 0
		}
		return true
	}
	return false
}
