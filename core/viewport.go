package core

import "github.com/gdamore/tcell/v2"

// Viewport maps a logical content area onto a fixed screen rectangle with a
// scroll offset. Components draw using content coordinates (row/column in the
// full, un-scrolled content); the viewport translates them to absolute screen
// coordinates and clips anything outside its rect.
//
// It dedupes the "draw rows within InnerRect at an offset" loop that TextView,
// List, Table, DataGrid and LogViewer each re-implement, and removes the
// recurring off-by-one overdraw bugs by making out-of-bounds writes a no-op
// rather than a manual guard at every call site.
//
// Typical use inside a Draw method:
//
//	vp := core.NewViewport(b.InnerRect())
//	vp.SetContentSize(maxLineWidth, len(lines))
//	vp.SetOffset(b.scrollX, b.scrollY) // clamps to valid range
//	vp.Fill(screen, ' ', clearStyle)
//	for row, line := range lines {
//	    vp.Print(screen, 0, row, line, style) // clipped automatically
//	}
type Viewport struct {
	x, y, w, h         int // screen rect (absolute coords)
	offsetX, offsetY   int // top-left content cell shown at the rect origin
	contentW, contentH int // total logical content size
}

// NewViewport returns a Viewport covering the given absolute screen rect. The
// argument order matches Box.InnerRect()'s return values, so callers can write
// NewViewport(b.InnerRect()).
func NewViewport(x, y, w, h int) Viewport {
	return Viewport{x: x, y: y, w: w, h: h}
}

// SetRect updates the screen rectangle and re-clamps the scroll offset.
func (v *Viewport) SetRect(x, y, w, h int) {
	v.x, v.y, v.w, v.h = x, y, w, h
	v.clamp()
}

// Rect returns the absolute screen rectangle (x, y, w, h).
func (v *Viewport) Rect() (int, int, int, int) { return v.x, v.y, v.w, v.h }

// Size returns the visible width and height of the viewport.
func (v *Viewport) Size() (int, int) { return v.w, v.h }

// Empty reports whether the viewport has no drawable area. Draw methods should
// return early when this is true.
func (v *Viewport) Empty() bool { return v.w <= 0 || v.h <= 0 }

// SetContentSize records the total logical content dimensions and re-clamps the
// scroll offset so it never points past the end of the content.
func (v *Viewport) SetContentSize(w, h int) {
	v.contentW, v.contentH = w, h
	v.clamp()
}

// Offset returns the current scroll offset (x, y) into the content.
func (v *Viewport) Offset() (int, int) { return v.offsetX, v.offsetY }

// SetOffset sets the scroll offset, clamped to the valid range.
func (v *Viewport) SetOffset(x, y int) {
	v.offsetX, v.offsetY = x, y
	v.clamp()
}

// ScrollBy adjusts the scroll offset by the given deltas, clamped.
func (v *Viewport) ScrollBy(dx, dy int) { v.SetOffset(v.offsetX+dx, v.offsetY+dy) }

// MaxOffset returns the largest valid scroll offset in each axis. It is never
// negative: when content fits within the viewport the max offset is 0.
func (v *Viewport) MaxOffset() (int, int) {
	return max(0, v.contentW-v.w), max(0, v.contentH-v.h)
}

// clamp keeps the offset within [0, MaxOffset].
func (v *Viewport) clamp() {
	mx, my := v.MaxOffset()
	v.offsetX = clampInt(v.offsetX, 0, mx)
	v.offsetY = clampInt(v.offsetY, 0, my)
}

// EnsureVisible scrolls the minimum amount needed to bring content cell
// (cx, cy) into view, then clamps. Use it to keep a selection on screen.
func (v *Viewport) EnsureVisible(cx, cy int) {
	if v.w > 0 {
		if cx < v.offsetX {
			v.offsetX = cx
		} else if cx >= v.offsetX+v.w {
			v.offsetX = cx - v.w + 1
		}
	}
	if v.h > 0 {
		if cy < v.offsetY {
			v.offsetY = cy
		} else if cy >= v.offsetY+v.h {
			v.offsetY = cy - v.h + 1
		}
	}
	v.clamp()
}

// Visible reports whether content cell (cx, cy) currently falls within the
// viewport given the scroll offset.
func (v *Viewport) Visible(cx, cy int) bool {
	return cx >= v.offsetX && cx < v.offsetX+v.w &&
		cy >= v.offsetY && cy < v.offsetY+v.h
}

// VisibleRows returns the half-open range of content row indices [first, last)
// that are on screen, clamped to the content height. Iterate as:
//
//	first, last := vp.VisibleRows()
//	for row := first; row < last; row++ { ... }
func (v *Viewport) VisibleRows() (int, int) {
	first := v.offsetY
	last := v.offsetY + v.h
	if v.contentH > 0 && last > v.contentH {
		last = v.contentH
	}
	if last < first {
		last = first
	}
	return first, last
}

// ScreenXY maps content cell (cx, cy) to its absolute screen coordinates,
// without clipping. Use it when delegating a row's render to a helper that
// writes directly to the screen (e.g. a cached draw func). Callers remain
// responsible for clipping width; pair with Size or VisibleRows.
func (v *Viewport) ScreenXY(cx, cy int) (int, int) {
	return v.x + (cx - v.offsetX), v.y + (cy - v.offsetY)
}

// SetContent writes a single rune at content cell (cx, cy), translating to
// screen coordinates and clipping. Writes outside the viewport are no-ops, so
// callers need no manual bounds checks.
func (v *Viewport) SetContent(screen tcell.Screen, cx, cy int, r rune, style tcell.Style) {
	if !v.Visible(cx, cy) {
		return
	}
	screen.SetContent(v.x+(cx-v.offsetX), v.y+(cy-v.offsetY), r, nil, style)
}

// Print writes text starting at content cell (cx, cy) along a single row, one
// rune per column. Runes scrolled off the left edge or clipped by the right
// edge are skipped.
func (v *Viewport) Print(screen tcell.Screen, cx, cy int, text string, style tcell.Style) {
	for i, r := range []rune(text) {
		v.SetContent(screen, cx+i, cy, r, style)
	}
}

// Fill paints every visible cell of the viewport with r and style. Use it to
// clear the background before drawing content.
func (v *Viewport) Fill(screen tcell.Screen, r rune, style tcell.Style) {
	if v.Empty() {
		return
	}
	FillRect(screen, v.x, v.y, v.w, v.h, r, style)
}

func clampInt(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}
