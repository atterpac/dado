package components

import "github.com/atterpac/dado/theme"

// GridColumn describes a single column in the DataGrid.
type GridColumn struct {
	Name     string // Display name (header text)
	Width    int    // Fixed width (0 = auto-size)
	MinWidth int    // Minimum width for auto-sizing
	MaxWidth int    // Maximum width for auto-sizing (0 = unlimited)
	Align    Align  // Horizontal alignment (AlignLeft, AlignRight, AlignCenter)
	Frozen   bool   // Pinned to left edge regardless of horizontal scroll
}

// GridCell represents a single cell value in the grid.
type GridCell struct {
	Value    string        // Display string
	RawValue any           // Underlying typed value
	Status   *theme.Status // Optional status coloring (takes precedence)
	ReadOnly bool          // Prevents editing
}

// DataGridSource provides data to a DataGrid.
// Implementations must be safe for concurrent access from Draw() and input handlers.
type DataGridSource interface {
	// RowCount returns the total number of data rows.
	RowCount() int
	// ColCount returns the total number of columns.
	ColCount() int
	// Columns returns column metadata.
	Columns() []GridColumn
	// Cell returns the cell at the given position.
	Cell(row, col int) GridCell
	// FetchRange is a pre-fetch hint for lazy loading sources.
	// In-memory sources may ignore this.
	FetchRange(startRow, endRow int)
}

// SliceSource is an in-memory DataGridSource backed by a 2D slice.
type SliceSource struct {
	columns  []GridColumn
	rows     [][]GridCell
	readOnly bool
}

// NewSliceSource creates a SliceSource from column definitions and cell data.
func NewSliceSource(columns []GridColumn, rows [][]GridCell) *SliceSource {
	return &SliceSource{
		columns: columns,
		rows:    rows,
	}
}

// SetSliceData is a convenience method to populate a SliceSource from string data.
// Column widths are auto-sized.
func (s *SliceSource) SetSliceData(columnNames []string, rows [][]string) {
	s.columns = make([]GridColumn, len(columnNames))
	for i, name := range columnNames {
		s.columns[i] = GridColumn{Name: name}
	}

	s.rows = make([][]GridCell, len(rows))
	for i, row := range rows {
		s.rows[i] = make([]GridCell, len(row))
		for j, val := range row {
			s.rows[i][j] = GridCell{Value: val, RawValue: val}
		}
	}
}

func (s *SliceSource) RowCount() int         { return len(s.rows) }
func (s *SliceSource) ColCount() int         { return len(s.columns) }
func (s *SliceSource) Columns() []GridColumn { return s.columns }
func (s *SliceSource) FetchRange(_, _ int)   {} // No-op for in-memory

// SetReadOnly marks all cells in this source as read-only, preventing edits.
func (s *SliceSource) SetReadOnly(readOnly bool) *SliceSource {
	s.readOnly = readOnly
	return s
}

func (s *SliceSource) Cell(row, col int) GridCell {
	if row < 0 || row >= len(s.rows) || col < 0 || col >= len(s.columns) {
		return GridCell{}
	}
	r := s.rows[row]
	if col >= len(r) {
		return GridCell{}
	}
	cell := r[col]
	if s.readOnly {
		cell.ReadOnly = true
	}
	return cell
}

// Viewport tracks the visible window into the data grid.
type Viewport struct {
	RowOffset int // First visible data row
	ColOffset int // First non-frozen visible column index
	VisRows   int // Number of visible rows (from widget height minus header)
	VisCols   int // Number of visible columns (computed from width and column widths)
}

// computeColumnWidths calculates the rendered width of each column.
// It auto-sizes columns based on header names and visible cell content.
func computeColumnWidths(source DataGridSource, vp *Viewport, availWidth int, showRowNumbers bool) []int {
	cols := source.Columns()
	if len(cols) == 0 {
		return nil
	}

	widths := make([]int, len(cols))

	// Start with header name widths
	for i, col := range cols {
		w := len(col.Name)
		if col.Width > 0 {
			widths[i] = col.Width
			continue
		}
		if w < 3 {
			w = 3 // Minimum default
		}
		widths[i] = w
	}

	// Scan visible rows to expand auto-sized columns
	startRow := vp.RowOffset
	endRow := vp.RowOffset + vp.VisRows
	if endRow > source.RowCount() {
		endRow = source.RowCount()
	}

	for row := startRow; row < endRow; row++ {
		for col := range cols {
			if cols[col].Width > 0 {
				continue // Fixed width, skip
			}
			cell := source.Cell(row, col)
			cellWidth := len(cell.Value)
			if cellWidth > widths[col] {
				widths[col] = cellWidth
			}
		}
	}

	// Apply min/max constraints
	for i, col := range cols {
		if col.Width > 0 {
			continue
		}
		if col.MinWidth > 0 && widths[i] < col.MinWidth {
			widths[i] = col.MinWidth
		}
		if col.MaxWidth > 0 && widths[i] > col.MaxWidth {
			widths[i] = col.MaxWidth
		}
	}

	// Add 2 for padding (1 char each side)
	for i := range widths {
		widths[i] += 2
	}

	return widths
}

// computeVisibleCols determines how many non-frozen columns fit in the available width
// starting from ColOffset, and returns that count.
func computeVisibleCols(colWidths []int, cols []GridColumn, colOffset, availWidth int) int {
	// First subtract frozen column widths
	remaining := availWidth
	for i, col := range cols {
		if col.Frozen {
			remaining -= colWidths[i]
		}
	}
	if remaining <= 0 {
		return 0
	}

	// Count how many non-frozen columns fit starting at colOffset
	nonFrozenIdx := 0
	count := 0
	for i, col := range cols {
		if col.Frozen {
			continue
		}
		if nonFrozenIdx < colOffset {
			nonFrozenIdx++
			continue
		}
		if remaining < colWidths[i] {
			// Partial column still counts
			count++
			break
		}
		remaining -= colWidths[i]
		count++
		_ = i
	}
	return count
}
