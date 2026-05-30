package components

import "github.com/gdamore/tcell/v2"

// Shared screen-drawing helpers. Leaf widgets used to hand-roll background
// clears, run-by-rune emit loops, centering math, and bounds clamping in every
// Draw method; these collapse that boilerplate and centralize clipping so an
// off-by-one in one widget can't outlive a fix here.
//
// Columns advance one cell per rune, matching the library's existing draw
// behavior. Wide-rune (CJK/emoji) width is a separate, deliberate follow-up.

// fillLine fills the row [x, x+w) at y with spaces in style st.
func fillLine(screen tcell.Screen, x, y, w int, st tcell.Style) {
	for col := x; col < x+w; col++ {
		screen.SetContent(col, y, ' ', nil, st)
	}
}

// fillRect fills the rectangle [x, x+w) x [y, y+h) with spaces in style st.
func fillRect(screen tcell.Screen, x, y, w, h int, st tcell.Style) {
	for row := y; row < y+h; row++ {
		fillLine(screen, x, row, w, st)
	}
}

// drawText draws text at (x, y), clipped so it never writes at or past x+maxW.
// It returns the column immediately after the last rune drawn.
func drawText(screen tcell.Screen, x, y, maxW int, text string, st tcell.Style) int {
	end := x + maxW
	col := x
	for _, r := range text {
		if col >= end {
			break
		}
		screen.SetContent(col, y, r, nil, st)
		col++
	}
	return col
}

// drawCentered draws text horizontally centered within [x, x+w) at y, clipped
// to that span. Returns the starting column used.
func drawCentered(screen tcell.Screen, x, y, w int, text string, st tcell.Style) int {
	start := x + (w-runeLen(text))/2
	if start < x {
		start = x
	}
	drawText(screen, start, y, x+w-start, text, st)
	return start
}

// runeLen returns the rune count of s (one cell per rune, matching drawText).
func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
