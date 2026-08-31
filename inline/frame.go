package inline

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Frame is a rectangular, styled cell buffer rendered by Renderer. It uses
// tcell's grapheme-aware CellBuffer for text measurement and placement, but it
// does not initialize or take ownership of a tcell terminal screen.
//
// Frame is not safe for concurrent use. Finish drawing a frame before passing
// it to Renderer.Render.
type Frame struct {
	cells  tcell.CellBuffer
	width  int
	height int
	cursor *Cursor
}

// Cursor describes a visible terminal cursor relative to the top-left corner
// of a Frame.
type Cursor struct {
	X int
	Y int
}

// FrameProvider builds a renderer-native frame at a requested width.
type FrameProvider interface {
	Frame(width int) *Frame
}

// NewFrame creates an empty frame. Negative dimensions are treated as zero.
func NewFrame(width, height int) *Frame {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	f := &Frame{width: width, height: height}
	f.cells.Resize(width, height)
	f.Clear(tcell.StyleDefault)
	return f
}

// Width returns the frame width in terminal cells.
func (f *Frame) Width() int { return f.width }

// Height returns the frame height in terminal cells.
func (f *Frame) Height() int { return f.height }

// Size returns the frame dimensions in terminal cells.
func (f *Frame) Size() (width, height int) { return f.width, f.height }

// Resize changes the frame dimensions while preserving overlapping content.
func (f *Frame) Resize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	f.width, f.height = width, height
	f.cells.Resize(width, height)
	if f.cursor != nil && !f.contains(f.cursor.X, f.cursor.Y) {
		f.cursor = nil
	}
}

// Clear fills the frame with styled spaces and hides its cursor.
func (f *Frame) Clear(style tcell.Style) {
	f.cells.Fill(' ', style)
	f.cursor = nil
}

// Put places the first grapheme cluster in text at x, y. It returns the
// unconsumed text and the number of terminal cells occupied by the grapheme.
// Out-of-bounds writes are ignored.
func (f *Frame) Put(x, y int, text string, style tcell.Style) (rest string, width int) {
	return f.cells.Put(x, y, text, style)
}

// SetContent is a compatibility helper for code that already draws with
// tcell's legacy primary-rune and combining-rune representation.
func (f *Frame) SetContent(x, y int, primary rune, combining []rune, style tcell.Style) {
	f.cells.SetContent(x, y, primary, combining, style)
}

// DrawString draws text starting at x, y. Newlines advance to the next row;
// other text is clipped at the right and bottom edges rather than wrapped.
func (f *Frame) DrawString(x, y int, text string, style tcell.Style) {
	for _, line := range strings.Split(text, "\n") {
		if y >= f.height {
			return
		}
		col := x
		for line != "" && col < f.width {
			rest, width := f.cells.Put(col, y, line, style)
			if width <= 0 || rest == line {
				break
			}
			line = rest
			col += width
		}
		y++
	}
}

// Cell returns the grapheme, style, and display width stored at x, y.
func (f *Frame) Cell(x, y int) (text string, style tcell.Style, width int) {
	return f.cells.Get(x, y)
}

// ShowCursor requests a visible cursor at x, y. An out-of-bounds position
// hides the cursor.
func (f *Frame) ShowCursor(x, y int) {
	if !f.contains(x, y) {
		f.cursor = nil
		return
	}
	f.cursor = &Cursor{X: x, Y: y}
}

// HideCursor requests that the terminal cursor remain hidden after rendering.
func (f *Frame) HideCursor() { f.cursor = nil }

func (f *Frame) contains(x, y int) bool {
	return x >= 0 && y >= 0 && x < f.width && y < f.height
}

// StackFrames composes frames vertically with blank rows between them. The
// cursor from the last frame that requests one is preserved in the result.
func StackFrames(gap int, frames ...*Frame) *Frame {
	gap = max(gap, 0)
	width, height, count := 0, 0, 0
	for _, frame := range frames {
		if frame == nil {
			continue
		}
		width = max(width, frame.Width())
		height += frame.Height()
		count++
	}
	if count > 1 {
		height += gap * (count - 1)
	}
	result := NewFrame(width, height)
	offset := 0
	seen := 0
	for _, frame := range frames {
		if frame == nil {
			continue
		}
		if seen > 0 {
			offset += gap
		}
		copyFrame(result, frame, offset)
		offset += frame.Height()
		seen++
	}
	return result
}

func copyFrame(target, source *Frame, yOffset int) {
	for y := 0; y < source.Height(); y++ {
		for x := 0; x < source.Width(); {
			text, style, width := source.Cell(x, y)
			if width < 1 {
				x++
				continue
			}
			_, _ = target.Put(x, yOffset+y, text, style)
			x += width
		}
	}
	if source.cursor != nil {
		target.ShowCursor(source.cursor.X, yOffset+source.cursor.Y)
	}
}
