package components

import (
	"github.com/gdamore/tcell/v2"
)

// HintGrid renders a collection of KeyHint items in a multi-row pill layout.
// Unlike KeyHintBar which only renders a single row, HintGrid wraps hints
// onto multiple rows when they don't fit the available width, and centers
// each row horizontally.
type HintGrid struct {
	widgetBase
	hints []KeyHint
}

// NewHintGrid creates a new hint grid.
func NewHintGrid() *HintGrid {
	g := &HintGrid{
		hints: make([]KeyHint, 0),
	}
	g.initWidget()
	return g
}

// SetHints sets the hints to display.
func (g *HintGrid) SetHints(hints []KeyHint) *HintGrid {
	g.hints = hints
	return g
}

// hintWidth returns the rendered width of a single hint pill: "[Key] Description"
func hintWidth(h KeyHint) int {
	return len("[") + len(h.Key) + len("]") + 1 + len(h.Description)
}

const hintSeparatorWidth = 3 // "   " between pills

// layoutRows calculates which hints go on each row given the available width.
// Returns a slice of slices, where each inner slice is the hints for one row.
func (g *HintGrid) layoutRows(width int) [][]KeyHint {
	if len(g.hints) == 0 || width < 1 {
		return nil
	}

	var rows [][]KeyHint
	var currentRow []KeyHint
	currentWidth := 0

	for _, h := range g.hints {
		hw := hintWidth(h)

		if len(currentRow) == 0 {
			// First hint on row always fits
			currentRow = append(currentRow, h)
			currentWidth = hw
			continue
		}

		// Check if hint fits on current row (with separator)
		needed := hintSeparatorWidth + hw
		if currentWidth+needed <= width {
			currentRow = append(currentRow, h)
			currentWidth += needed
		} else {
			// Wrap to next row
			rows = append(rows, currentRow)
			currentRow = []KeyHint{h}
			currentWidth = hw
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return rows
}

// GetPreferredHeight returns the total height needed to display all hints
// at the given width, including a blank-line gap between rows.
func (g *HintGrid) GetPreferredHeight(width int) int {
	rows := g.layoutRows(width)
	n := len(rows)
	if n == 0 {
		return 1
	}
	// Each row is 1 line, with a 1-line gap between rows
	return n + (n - 1)
}

// Draw renders the hint grid with multi-row pill layout.
func (g *HintGrid) Draw(screen tcell.Screen) {
	g.Box.DrawForSubclass(screen)

	x, y, width, height := g.GetInnerRect()
	if width < 1 || height < 1 || len(g.hints) == 0 {
		return
	}

	th := g.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	accentColor := th.Accent()

	bgStyle := tcell.StyleDefault.Background(bgColor)
	keyStyle := tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
	descStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)

	rows := g.layoutRows(width)

	for rowIdx, row := range rows {
		// Each row occupies 2 lines (content + gap), except the last row (no trailing gap)
		drawY := y + rowIdx*2
		if drawY >= y+height {
			break
		}

		// Clear the line
		fillLine(screen, x, drawY, width, bgStyle)

		// Calculate total width of this row for centering
		rowWidth := 0
		for i, h := range row {
			if i > 0 {
				rowWidth += hintSeparatorWidth
			}
			rowWidth += hintWidth(h)
		}

		// Center horizontally
		startX := x + (width-rowWidth)/2
		if startX < x {
			startX = x
		}

		currentX := startX
		for i, h := range row {
			if i > 0 {
				// Draw separator
				for s := 0; s < hintSeparatorWidth; s++ {
					if currentX < x+width {
						screen.SetContent(currentX, drawY, ' ', nil, bgStyle)
						currentX++
					}
				}
			}

			// Draw key pill: [Key]
			keyText := "[" + h.Key + "]"
			currentX = drawText(screen, currentX, drawY, x+width-currentX, keyText, keyStyle)

			// Draw space
			if currentX < x+width {
				screen.SetContent(currentX, drawY, ' ', nil, descStyle)
				currentX++
			}

			// Draw description
			currentX = drawText(screen, currentX, drawY, x+width-currentX, h.Description, descStyle)
		}
	}
}
