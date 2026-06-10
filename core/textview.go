package core

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

// TextView displays multi-line scrollable text with optional ANSI color tag parsing.
// lineSpan indexes one wrapped line's spans within TextView.spanArena by integer
// range (not pointer/sub-slice), so the arena can grow mid-reflow without
// invalidating earlier lines.
type lineSpan struct {
	start, end int // half-open range into spanArena
	width      int // visible columns (rune count) for alignment
}

type TextView struct {
	Box
	text string
	// Wrapped-line cache, rebuilt by reflow: spanArena holds all spans
	// contiguously, lineSpans indexes it. Both reused (reset to [:0]) across
	// reflows, so a resize allocates nothing once capacity fits.
	spanArena []Span
	lineSpans []lineSpan
	scrollY   int
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
	tv.lineSpans = tv.lineSpans[:0] // invalidate cache; len 0 forces reflow
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
	if len(tv.lineSpans) == 0 || tv.wrapWidth != w {
		tv.reflow(w)
	}
	vp.SetContentSize(w, len(tv.lineSpans))
	vp.SetOffset(0, tv.scrollY)
	// Write the clamped offset back so scrollY never lands past the last line.
	_, tv.scrollY = vp.Offset()

	baseStyle := tv.textStyle
	if tv.fgColor != tcell.ColorDefault {
		baseStyle = baseStyle.Foreground(tv.fgColor)
	}
	clearStyle := tcell.StyleDefault.Background(tv.Box.bg())

	first, last := vp.VisibleRows()
	for lineIdx := first; lineIdx < last; lineIdx++ {
		ls := tv.lineSpans[lineIdx]
		// Compute line width for alignment
		lineRunes := ls.width
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
		for _, seg := range tv.spanArena[ls.start:ls.end] {
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

// reflow wraps tv.text into lines of maxWidth (parsing color tags if dynamic),
// writing into the reused spanArena/lineSpans buffers. IndexByte splits on '\n'
// to avoid strings.Split's per-call []string allocation.
func (tv *TextView) reflow(maxWidth int) {
	tv.wrapWidth = maxWidth
	tv.spanArena = tv.spanArena[:0]
	tv.lineSpans = tv.lineSpans[:0]

	text := tv.text
	for {
		nl := strings.IndexByte(text, '\n')
		if nl < 0 {
			tv.wrapLineInto(text, maxWidth)
			return
		}
		tv.wrapLineInto(text[:nl], maxWidth)
		text = text[nl+1:]
	}
}

// wrapLineInto parses one raw line (color tags if tv.dynamic) and hard-wraps it
// to width, appending spans to spanArena and one lineSpan per output line. Each
// span's Text slices raw (no copy). Matches ParseTagged+splitWidth semantics:
// strict column breaks, no word wrap.
func (tv *TextView) wrapLineInto(raw string, width int) {
	// Rare "[[" escape (literal "[" != source bytes): defer this one line to the
	// allocating ParseTagged path and copy its spans in.
	if tv.dynamic && strings.Contains(raw, "[[") {
		for _, ln := range ParseTagged(raw, tcell.StyleDefault).splitWidth(width) {
			start := len(tv.spanArena)
			tv.spanArena = append(tv.spanArena, ln.Spans()...)
			tv.lineSpans = append(tv.lineSpans, lineSpan{start, len(tv.spanArena), ln.Width()})
		}
		return
	}

	nowrap := width <= 0
	style := tcell.StyleDefault
	spanStart := len(tv.spanArena) // first span of the current output line
	segStart := 0                  // byte offset in raw where the pending span begins
	col := 0                       // visible columns in the current output line
	emitted := false               // whether this raw produced any output line yet

	flush := func(end int) {
		if end > segStart {
			tv.spanArena = append(tv.spanArena, Span{Text: raw[segStart:end], Style: style})
		}
	}
	endLine := func(end int) {
		flush(end)
		tv.lineSpans = append(tv.lineSpans, lineSpan{spanStart, len(tv.spanArena), col})
		spanStart = len(tv.spanArena)
		segStart = end
		col = 0
		emitted = true
	}

	i := 0
	for i < len(raw) {
		if tv.dynamic && raw[i] == '[' {
			if end := strings.IndexByte(raw[i:], ']'); end > 0 {
				end += i
				if newStyle, ok := parseTag(raw[i+1:end], style, tcell.StyleDefault); ok {
					flush(i)
					style = newStyle
					i = end + 1
					segStart = i
					continue
				}
			}
		}
		_, size := utf8.DecodeRuneInString(raw[i:])
		i += size
		col++
		if !nowrap && col >= width {
			endLine(i)
		}
	}
	// Trailing partial line, or a single (possibly empty) line that never wrapped.
	if col > 0 || !emitted {
		flush(i)
		tv.lineSpans = append(tv.lineSpans, lineSpan{spanStart, len(tv.spanArena), col})
	}
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

	// Split into fg:bg:attrs manually — strings.SplitN would allocate a slice
	// per tag, per frame.
	fgPart := tag
	bgPart := ""
	attrPart := ""
	bgProvided := false
	if i := strings.IndexByte(tag, ':'); i >= 0 {
		fgPart = tag[:i]
		bgProvided = true
		rest := tag[i+1:]
		if j := strings.IndexByte(rest, ':'); j >= 0 {
			bgPart = rest[:j]
			attrPart = rest[j+1:]
		} else {
			bgPart = rest
		}
	}

	// If all parts are empty or dash with no valid color, not a tag
	fgColor, fgOK := resolveColor(fgPart)
	bgColor, bgOK := resolveColor(bgPart)

	// Must have at least one recognized part to be a color tag
	if !fgOK && !(bgProvided && bgOK) && attrPart == "" {
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
	_, _, w, h := tv.InnerRect()
	if len(tv.lineSpans) == 0 || tv.wrapWidth != w {
		tv.reflow(w)
	}
	switch ev.Key() {
	case tcell.KeyDown:
		tv.scrollY++
	case tcell.KeyUp:
		tv.scrollY--
	case tcell.KeyPgDn:
		tv.scrollY += h
	case tcell.KeyPgUp:
		tv.scrollY -= h
	case tcell.KeyHome:
		tv.scrollY = 0
	case tcell.KeyEnd:
		tv.scrollY = tv.maxScrollY(h)
	default:
		return false
	}
	tv.scrollY = clampInt(tv.scrollY, 0, tv.maxScrollY(h))
	return true
}

// maxScrollY is the largest valid top row for a viewport of height h: never
// negative, and zero when the content fits.
func (tv *TextView) maxScrollY(h int) int {
	return max(0, len(tv.lineSpans)-h)
}
