package core

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

// Box-drawing runes for borders.
var (
	borderHoriz     = '─'
	borderVert      = '│'
	borderTopLeft   = '┌'
	borderTopRight  = '┐'
	borderBotLeft   = '└'
	borderBotRight  = '┘'
)

// FillRect fills a rectangular region with a rune and style.
func FillRect(screen tcell.Screen, x, y, w, h int, r rune, style tcell.Style) {
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			screen.SetContent(col, row, r, nil, style)
		}
	}
}

// DrawBorder draws a single-line box border at x,y,w,h with the given style.
func DrawBorder(screen tcell.Screen, x, y, w, h int, style tcell.Style) {
	if w < 2 || h < 2 {
		return
	}
	// Horizontal lines
	for col := x + 1; col < x+w-1; col++ {
		screen.SetContent(col, y, borderHoriz, nil, style)
		screen.SetContent(col, y+h-1, borderHoriz, nil, style)
	}
	// Vertical lines
	for row := y + 1; row < y+h-1; row++ {
		screen.SetContent(x, row, borderVert, nil, style)
		screen.SetContent(x+w-1, row, borderVert, nil, style)
	}
	// Corners
	screen.SetContent(x, y, borderTopLeft, nil, style)
	screen.SetContent(x+w-1, y, borderTopRight, nil, style)
	screen.SetContent(x, y+h-1, borderBotLeft, nil, style)
	screen.SetContent(x+w-1, y+h-1, borderBotRight, nil, style)
}

// DrawTitle renders a title string centered (or left/right aligned) inside the
// top border row. align: AlignLeft=-1, AlignCenter=0, AlignRight=1.
func DrawTitle(screen tcell.Screen, x, y, w int, title string, align int, style tcell.Style) {
	if w < 4 || title == "" {
		return
	}
	inner := w - 2 // space inside corners
	runes := []rune(title)
	if len(runes) > inner {
		runes = runes[:inner]
	}
	var startX int
	switch align {
	case AlignLeft:
		startX = x + 1
	case AlignRight:
		startX = x + w - 1 - len(runes)
	default: // center
		startX = x + (w-len(runes))/2
	}
	for i, r := range runes {
		screen.SetContent(startX+i, y, r, nil, style)
	}
}

// PrintAt writes a string at (x, y) with the given style.
func PrintAt(screen tcell.Screen, x, y int, text string, style tcell.Style) {
	for i, r := range []rune(text) {
		screen.SetContent(x+i, y, r, nil, style)
	}
}

// PrintClipped writes text at (x, y) clipped to maxWidth columns, left-aligned.
// Left-aligned within the clip width.
func PrintClipped(screen tcell.Screen, text string, x, y, maxWidth int, style tcell.Style) {
	if maxWidth <= 0 {
		return
	}
	for i, r := range []rune(text) {
		if i >= maxWidth {
			break
		}
		screen.SetContent(x+i, y, r, nil, style)
	}
}

// PrintTagged writes text at (x, y) clipped to maxWidth, parsing
// color tags ([#rrggbb], [-], [::b], etc.) as it renders.
// baseStyle is the fallback style for uncolored segments.
// Returns the number of visible columns written.
func PrintTagged(screen tcell.Screen, text string, x, y, maxWidth int, baseStyle tcell.Style) int {
	if maxWidth <= 0 {
		return 0
	}
	col := 0
	style := baseStyle
	for len(text) > 0 {
		if col >= maxWidth {
			break
		}
		if text[0] == '[' {
			end := strings.Index(text, "]")
			if end > 0 {
				tag := text[1:end]
				if newStyle, ok := parseTag(tag, style); ok {
					style = newStyle
					text = text[end+1:]
					continue
				}
			}
		}
		r, size := utf8.DecodeRuneInString(text)
		screen.SetContent(x+col, y, r, nil, style)
		col++
		text = text[size:]
	}
	return col
}
