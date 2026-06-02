package core

import "github.com/gdamore/tcell/v2"

// gridItem is a child placed on the grid with its span.
type gridItem struct {
	widget                 Widget
	row, col               int
	rowSpan, colSpan       int
}

// Grid is a two-dimensional constraint layout. Rows and columns are sized with
// the same fixed/proportional scheme as Flex: a size > 0 is an exact number of
// cells, a size of 0 shares leftover space proportionally (weight 1 each), and a
// negative size is a proportional weight of -size.
//
// Items are placed at a (row, col) origin and may span multiple rows/columns.
//
//	grid := core.NewGrid().
//	    SetRows(3, 0, 1).     // header: 3 cells, body: flexible, footer: 1 cell
//	    SetColumns(20, 0).    // sidebar: 20 cells, content: flexible
//	    AddItem(header,  0, 0, 1, 2, false). // row 0, spans both columns
//	    AddItem(sidebar, 1, 0, 1, 1, false).
//	    AddItem(content, 1, 1, 1, 1, true).
//	    AddItem(footer,  2, 0, 1, 2, false)
type Grid struct {
	Box
	rows, columns []int
	gapRow, gapCol int
	items          []gridItem
}

// NewGrid returns an empty Grid.
func NewGrid() *Grid { return &Grid{} }

// SetRows defines row sizes. >0 fixed cells, 0 flexible (weight 1), <0 weight -n.
func (g *Grid) SetRows(rows ...int) *Grid { g.rows = rows; return g }

// SetColumns defines column sizes with the same convention as SetRows.
func (g *Grid) SetColumns(columns ...int) *Grid { g.columns = columns; return g }

// SetGap sets the spacing between rows and columns.
func (g *Grid) SetGap(row, col int) *Grid { g.gapRow, g.gapCol = row, col; return g }

// AddItem places w at (row, col) spanning rowSpan rows and colSpan columns.
// A span of 0 or less is treated as 1. The final bool (focus hint) is ignored,
// matching Flex.AddItem so call sites are grep-replaceable.
func (g *Grid) AddItem(w Widget, row, col, rowSpan, colSpan int, _ bool) *Grid {
	if rowSpan < 1 {
		rowSpan = 1
	}
	if colSpan < 1 {
		colSpan = 1
	}
	g.items = append(g.items, gridItem{w, row, col, rowSpan, colSpan})
	return g
}

// RemoveItem removes the first occurrence of w.
func (g *Grid) RemoveItem(w Widget) *Grid {
	out := g.items[:0]
	for _, it := range g.items {
		if it.widget != w {
			out = append(out, it)
		}
	}
	g.items = out
	return g
}

// Clear removes all children.
func (g *Grid) Clear() *Grid { g.items = g.items[:0]; return g }

// ItemCount returns the number of children.
func (g *Grid) ItemCount() int { return len(g.items) }

// Draw lays out children on the grid and draws them.
func (g *Grid) Draw(screen tcell.Screen) {
	g.Box.Draw(screen)
	x, y, w, h := g.InnerRect()
	if w <= 0 || h <= 0 || len(g.rows) == 0 || len(g.columns) == 0 {
		return
	}

	rowPos, rowSize := distribute(g.rows, h, g.gapRow)
	colPos, colSize := distribute(g.columns, w, g.gapCol)

	for _, it := range g.items {
		r, c := it.row, it.col
		if r < 0 || c < 0 || r >= len(g.rows) || c >= len(g.columns) {
			continue
		}
		rEnd := min(r+it.rowSpan, len(g.rows)) - 1
		cEnd := min(c+it.colSpan, len(g.columns)) - 1

		ix := x + colPos[c]
		iy := y + rowPos[r]
		iw := colPos[cEnd] + colSize[cEnd] - colPos[c]
		ih := rowPos[rEnd] + rowSize[rEnd] - rowPos[r]
		if iw <= 0 || ih <= 0 {
			continue
		}
		it.widget.SetRect(ix, iy, iw, ih)
		it.widget.Draw(screen)
	}
}

// distribute computes the start offset and size of each track given the total
// available space and the inter-track gap. Sizes follow the fixed/proportional
// convention; the last proportional track absorbs rounding remainder.
func distribute(tracks []int, available, gap int) (pos, size []int) {
	n := len(tracks)
	pos = make([]int, n)
	size = make([]int, n)

	remaining := max(available-gap*(n-1), 0)

	totalWeight := 0
	for _, t := range tracks {
		if t > 0 {
			remaining -= t
		} else {
			totalWeight += weight(t)
		}
	}
	remaining = max(remaining, 0)
	flexRemaining := remaining
	for i, t := range tracks {
		if t > 0 {
			size[i] = t
			continue
		}
		wgt := weight(t)
		if totalWeight > 0 {
			if isLastFlex(tracks[i:]) {
				size[i] = flexRemaining
			} else {
				size[i] = remaining * wgt / totalWeight
			}
			flexRemaining -= size[i]
		}
	}

	offset := 0
	for i := range tracks {
		pos[i] = offset
		offset += size[i] + gap
	}
	return pos, size
}

// weight returns the proportional weight of a track size: 0 → 1, -n → n.
func weight(t int) int {
	if t == 0 {
		return 1
	}
	return -t
}

// isLastFlex reports whether tracks[0] is the last proportional track.
func isLastFlex(tracks []int) bool {
	for _, t := range tracks[1:] {
		if t <= 0 {
			return false
		}
	}
	return true
}

// HandleKey routes the event to the focused child.
func (g *Grid) HandleKey(ev *tcell.EventKey) bool {
	for _, it := range g.items {
		if it.widget.HasFocus() {
			if kh, ok := it.widget.(KeyHandler); ok {
				return kh.HandleKey(ev)
			}
		}
	}
	return false
}

// HandleMouse routes to the child at the event position.
func (g *Grid) HandleMouse(action MouseAction, ev *tcell.EventMouse) (bool, Widget) {
	mx, my := ev.Position()
	for _, it := range g.items {
		wx, wy, ww, wh := it.widget.Rect()
		if mx >= wx && mx < wx+ww && my >= wy && my < wy+wh {
			if mh, ok := it.widget.(MouseHandler); ok {
				return mh.HandleMouse(action, ev)
			}
		}
	}
	return false, nil
}

// Children returns the child widgets in placement order (implements Container).
func (g *Grid) Children() []Widget {
	out := make([]Widget, len(g.items))
	for i, it := range g.items {
		out[i] = it.widget
	}
	return out
}

// DescendantsAt returns widgets whose Rect contains (x, y), deepest first
// (implements Container).
func (g *Grid) DescendantsAt(x, y int) []Widget {
	var result []Widget
	for _, it := range g.items {
		wx, wy, ww, wh := it.widget.Rect()
		if x >= wx && x < wx+ww && y >= wy && y < wy+wh {
			if c, ok := it.widget.(Container); ok {
				result = append(c.DescendantsAt(x, y), it.widget)
			} else {
				result = append(result, it.widget)
			}
		}
	}
	return result
}
