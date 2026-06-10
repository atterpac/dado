package core

import (
	"github.com/gdamore/tcell/v2"
)

// TableCell is a single cell in a Table.
type TableCell struct {
	Text         string
	Align        int
	Color        tcell.Color
	BgColor      tcell.Color
	MaxWidth     int
	Expansion    int
	Reference    any
	NotSelectable bool // true means the cell cannot be selected
}

// NewTableCell creates a cell with default settings (AlignLeft, default colors).
func NewTableCell(text string) *TableCell {
	return &TableCell{
		Text:  text,
		Align: AlignLeft,
		Color: tcell.ColorDefault,
	}
}

// SetText sets the cell text.
func (c *TableCell) SetText(text string) *TableCell { c.Text = text; return c }

// SetTextColor sets the cell foreground color.
func (c *TableCell) SetTextColor(color tcell.Color) *TableCell { c.Color = color; return c }

// SetBackgroundColor sets the cell background color.
func (c *TableCell) SetBackgroundColor(color tcell.Color) *TableCell { c.BgColor = color; return c }

// SetExpansion sets the cell expansion factor (larger = wider).
func (c *TableCell) SetExpansion(exp int) *TableCell { c.Expansion = exp; return c }

// SetAlign sets the cell text alignment.
func (c *TableCell) SetAlign(align int) *TableCell { c.Align = align; return c }

// SetMaxWidth sets the maximum column width for this cell (0 = no limit).
func (c *TableCell) SetMaxWidth(maxWidth int) *TableCell { c.MaxWidth = maxWidth; return c }

// SetReference stores an arbitrary reference value on the cell.
func (c *TableCell) SetReference(ref any) *TableCell { c.Reference = ref; return c }

// GetReference returns the reference value stored on the cell.
func (c *TableCell) GetReference() any { return c.Reference }

// SetSelectable sets whether this cell can be selected (default true).
// Pass false to make the cell non-selectable (e.g. header cells).
func (c *TableCell) SetSelectable(selectable bool) *TableCell {
	c.NotSelectable = !selectable
	return c
}

type tablePos struct{ row, col int }

// Table displays a grid of TableCells with optional row/column selection.
type Table struct {
	Box
	cells             map[tablePos]*TableCell
	rows              int
	cols              int
	selRow            int
	selCol            int
	selRows           bool
	selCols           bool
	onSelect          func(row, col int)
	onSelectionChange func(row, col int)
	selectedStyle     tcell.Style
	fixedRows         int
	fixedCols         int
	separator         rune
	rowOffset         int
}

// NewTable returns an empty Table.
func NewTable() *Table {
	return &Table{cells: make(map[tablePos]*TableCell)}
}

// SetCell stores a cell at (row, col). Updates row/col counts.
func (t *Table) SetCell(row, col int, cell *TableCell) *Table {
	t.cells[tablePos{row, col}] = cell
	if row+1 > t.rows {
		t.rows = row + 1
	}
	if col+1 > t.cols {
		t.cols = col + 1
	}
	return t
}

// GetCell returns the cell at (row, col). Returns an empty default cell if not set.
func (t *Table) GetCell(row, col int) *TableCell {
	if c, ok := t.cells[tablePos{row, col}]; ok {
		return c
	}
	return &TableCell{Align: AlignLeft, Color: tcell.ColorDefault}
}

// GetRowCount returns the number of rows (max row index + 1).
func (t *Table) GetRowCount() int { return t.rows }

// GetColumnCount returns the number of columns (max col index + 1).
func (t *Table) GetColumnCount() int { return t.cols }

// Clear removes all cells and resets counts.
func (t *Table) Clear() *Table {
	t.cells = make(map[tablePos]*TableCell)
	t.rows = 0
	t.cols = 0
	t.selRow = 0
	t.selCol = 0
	return t
}

// SetSelectable enables row/column selection.
func (t *Table) SetSelectable(rows, cols bool) *Table {
	t.selRows = rows
	t.selCols = cols
	return t
}

// GetSelection returns the currently selected (row, col).
func (t *Table) GetSelection() (row, col int) { return t.selRow, t.selCol }

// SetSelectedFunc sets the callback fired when a cell is activated (Enter).
func (t *Table) SetSelectedFunc(fn func(row, col int)) *Table {
	t.onSelect = fn
	return t
}

// SetSelectionChangedFunc sets a callback fired when the selection moves.
func (t *Table) SetSelectionChangedFunc(fn func(row, col int)) *Table {
	t.onSelectionChange = fn
	return t
}

// SetSelectedStyle sets the style used for the selected row/column.
func (t *Table) SetSelectedStyle(s tcell.Style) *Table {
	t.selectedStyle = s
	return t
}

// Select moves the selection to (row, col) and fires the change callback.
func (t *Table) Select(row, col int) *Table {
	t.selRow = row
	t.selCol = col
	if t.onSelectionChange != nil {
		t.onSelectionChange(row, col)
	}
	return t
}

// SetFixed fixes the first rows/cols so they do not scroll.
func (t *Table) SetFixed(rows, cols int) *Table {
	t.fixedRows = rows
	t.fixedCols = cols
	return t
}

// SetBorders is a no-op kept for API compatibility (core.Table never draws borders).
func (t *Table) SetBorders(borders bool) *Table { return t }

// SetSeparator sets the rune drawn between columns.
func (t *Table) SetSeparator(sep rune) *Table { t.separator = sep; return t }

// RemoveRow removes all cells in the given row and shifts subsequent rows up.
func (t *Table) RemoveRow(row int) *Table {
	newCells := make(map[tablePos]*TableCell, len(t.cells))
	for pos, cell := range t.cells {
		if pos.row == row {
			continue
		}
		if pos.row > row {
			newCells[tablePos{pos.row - 1, pos.col}] = cell
		} else {
			newCells[pos] = cell
		}
	}
	t.cells = newCells
	if t.rows > 0 {
		t.rows--
	}
	// Adjust selection
	if t.selRow > row && t.selRow > 0 {
		t.selRow--
	} else if t.selRow == row && t.selRow >= t.rows {
		if t.rows > 0 {
			t.selRow = t.rows - 1
		} else {
			t.selRow = 0
		}
	}
	return t
}

// InsertRow inserts a blank row at the given index, shifting subsequent rows down.
func (t *Table) InsertRow(row int) *Table {
	newCells := make(map[tablePos]*TableCell, len(t.cells))
	for pos, cell := range t.cells {
		if pos.row >= row {
			newCells[tablePos{pos.row + 1, pos.col}] = cell
		} else {
			newCells[pos] = cell
		}
	}
	t.cells = newCells
	t.rows++
	return t
}

// DrawForSubclass draws the table background box (for embedding).
func (t *Table) DrawForSubclass(screen tcell.Screen, _ interface{}) {
	t.Box.Draw(screen)
}

// GetInnerRect returns the inner drawing area (same as Box.InnerRect).
func (t *Table) GetInnerRect() (x, y, width, height int) {
	return t.InnerRect()
}

// Draw renders visible rows and columns.
func (t *Table) Draw(screen tcell.Screen) {
	t.Box.Draw(screen)
	vp := NewViewport(t.InnerRect())
	if vp.Empty() || t.rows == 0 {
		return
	}
	x, _, w, h := vp.Rect()

	// Compute per-column widths from content, honoring MaxWidth and Expansion.
	colWidths := t.computeColumnWidths(w)

	// Auto-scroll to keep selection visible, centered when possible. The
	// viewport clamps the offset to the valid range when applied.
	if t.selRows {
		if t.selRow < t.rowOffset || t.selRow >= t.rowOffset+h {
			t.rowOffset = t.selRow - h/2
		}
	}
	vp.SetContentSize(w, t.rows)
	vp.SetOffset(0, t.rowOffset)
	_, t.rowOffset = vp.Offset() // keep field in sync for HandleKey/paging

	first, last := vp.VisibleRows()
	for row := first; row < last; row++ {
		_, sy := vp.ScreenXY(0, row)
		rowSelected := t.selRows && row == t.selRow

		// Base row style and selection style. Selected rows are filled across
		// the full width (including inter-column gaps and trailing space) so the
		// highlight covers the whole row, not just the cell contents.
		rowStyle := tcell.StyleDefault
		if bg := t.bg(); bg != tcell.ColorDefault {
			rowStyle = rowStyle.Background(bg)
		}
		selStyle := rowStyle
		if rowSelected {
			if t.selectedStyle != (tcell.Style{}) {
				selStyle = t.selectedStyle
			} else {
				selStyle = rowStyle.Reverse(true)
			}
			for c := 0; c < w; c++ {
				screen.SetContent(x+c, sy, ' ', nil, selStyle)
			}
		}

		colX := x
		for col := 0; col < t.cols; col++ {
			cell := t.GetCell(row, col)
			style := selStyle
			if !rowSelected {
				style = rowStyle
				if cell.BgColor != tcell.ColorDefault {
					style = style.Background(cell.BgColor)
				}
				if cell.Color != tcell.ColorDefault {
					style = style.Foreground(cell.Color)
				}
				// Fill the cell background for unselected rows.
				for c := 0; c < colWidths[col]; c++ {
					screen.SetContent(colX+c, sy, ' ', nil, style)
				}
			}
			cw := colWidths[col]
			// Offset the text within the cell to honor alignment.
			textX := colX
			if pad := cw - TaggedWidth(cell.Text); pad > 0 {
				switch cell.Align {
				case AlignRight:
					textX += pad
				case AlignCenter:
					textX += pad / 2
				}
			}
			// Render text with tag parsing (handles [#color], [-], etc.). On a
			// selected row, lock the colors to the selection style so per-cell
			// tag colors can't clash with — or disappear into — the highlight.
			if rowSelected {
				PrintTaggedLockColors(screen, cell.Text, textX, sy, colX+cw-textX, style)
			} else {
				PrintTagged(screen, cell.Text, textX, sy, colX+cw-textX, style)
			}
			colX += cw + 1 // one column gap between cells
		}
	}
}

// computeColumnWidths returns the rendered width of each column. Natural widths
// come from cell content (capped by MaxWidth); any leftover horizontal space is
// distributed to columns whose cells request Expansion, and overflow is taken
// back from the widest columns so the row never exceeds the available width.
func (t *Table) computeColumnWidths(avail int) []int {
	widths := make([]int, t.cols)
	expansion := make([]int, t.cols)
	if t.cols == 0 {
		return widths
	}

	for col := 0; col < t.cols; col++ {
		for row := 0; row < t.rows; row++ {
			cell := t.GetCell(row, col)
			cw := TaggedWidth(cell.Text)
			if cell.MaxWidth > 0 && cw > cell.MaxWidth {
				cw = cell.MaxWidth
			}
			if cw > widths[col] {
				widths[col] = cw
			}
			if cell.Expansion > expansion[col] {
				expansion[col] = cell.Expansion
			}
		}
	}

	gaps := t.cols - 1 // single-space gap between columns
	usable := avail - gaps
	if usable < t.cols {
		usable = t.cols // guarantee at least 1 cell per column
	}

	total := 0
	for _, cw := range widths {
		total += cw
	}

	switch {
	case total < usable:
		// Distribute slack to expansion columns (or the last column if none).
		slack := usable - total
		totalExp := 0
		for _, e := range expansion {
			totalExp += e
		}
		if totalExp == 0 {
			widths[t.cols-1] += slack
			break
		}
		given := 0
		last := -1
		for col, e := range expansion {
			if e == 0 {
				continue
			}
			last = col
			add := slack * e / totalExp
			widths[col] += add
			given += add
		}
		if last >= 0 {
			widths[last] += slack - given // remainder from integer division
		}
	case total > usable:
		// Shrink the widest columns until the row fits.
		over := total - usable
		for over > 0 {
			widest := 0
			for col := 1; col < t.cols; col++ {
				if widths[col] > widths[widest] {
					widest = col
				}
			}
			if widths[widest] <= 1 {
				break
			}
			widths[widest]--
			over--
		}
	}

	return widths
}

// HandleKey handles selection movement and activation.
func (t *Table) HandleKey(ev *tcell.EventKey) bool {
	if !t.selRows && !t.selCols {
		return false
	}
	switch ev.Key() {
	case tcell.KeyDown:
		if t.selRows && t.selRow < t.rows-1 {
			t.selRow++
			if t.onSelectionChange != nil {
				t.onSelectionChange(t.selRow, t.selCol)
			}
		}
		return true
	case tcell.KeyUp:
		if t.selRows && t.selRow > 0 {
			t.selRow--
			if t.onSelectionChange != nil {
				t.onSelectionChange(t.selRow, t.selCol)
			}
		}
		return true
	case tcell.KeyRight:
		if t.selCols && t.selCol < t.cols-1 {
			t.selCol++
			if t.onSelectionChange != nil {
				t.onSelectionChange(t.selRow, t.selCol)
			}
		}
		return true
	case tcell.KeyLeft:
		if t.selCols && t.selCol > 0 {
			t.selCol--
			if t.onSelectionChange != nil {
				t.onSelectionChange(t.selRow, t.selCol)
			}
		}
		return true
	case tcell.KeyEnter:
		if t.onSelect != nil {
			t.onSelect(t.selRow, t.selCol)
		}
		return true
	}
	return false
}
