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
	x, y, w, h := t.InnerRect()
	if w <= 0 || h <= 0 || t.rows == 0 {
		return
	}

	// Compute column widths (simple: divide equally)
	colW := w
	if t.cols > 0 {
		colW = w / t.cols
		if colW < 1 {
			colW = 1
		}
	}

	// Auto-scroll to keep selection visible, centered when possible.
	if t.selRows && h > 0 {
		if t.selRow < t.rowOffset {
			t.rowOffset = t.selRow - h/2
		} else if t.selRow >= t.rowOffset+h {
			t.rowOffset = t.selRow - h/2
		}
		if t.rowOffset > t.rows-h {
			t.rowOffset = t.rows - h
		}
		if t.rowOffset < 0 {
			t.rowOffset = 0
		}
	}

	for screenRow := 0; screenRow < h; screenRow++ {
		row := t.rowOffset + screenRow
		if row >= t.rows {
			break
		}
		colX := x
		for col := 0; col < t.cols; col++ {
			cell := t.GetCell(row, col)
			// Build base style: table background → cell background → cell foreground
			style := tcell.StyleDefault
			if t.backgroundColor != tcell.ColorDefault {
				style = style.Background(t.backgroundColor)
			}
			if cell.BgColor != tcell.ColorDefault {
				style = style.Background(cell.BgColor)
			}
			if cell.Color != tcell.ColorDefault {
				style = style.Foreground(cell.Color)
			}
			if t.selRows && row == t.selRow {
				if t.selectedStyle != (tcell.Style{}) {
					style = t.selectedStyle
				} else {
					style = style.Reverse(true)
				}
			}
			cw := colW
			if col == t.cols-1 {
				cw = x + w - colX // last col gets remainder
			}
			// Fill cell background first
			for c := 0; c < cw; c++ {
				screen.SetContent(colX+c, y+screenRow, ' ', nil, style)
			}
			// Render text with tag parsing (handles [#color], [-], etc.)
			PrintTagged(screen, cell.Text, colX, y+screenRow, cw, style)
			colX += colW
		}
	}
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
