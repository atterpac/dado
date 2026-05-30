package components

import (
	"github.com/rivo/tview"
)

// GridMode represents the current interaction mode of the DataGrid.
type GridMode int

const (
	// GridModeNormal is the default navigation mode.
	GridModeNormal GridMode = iota
	// GridModeEdit is the inline cell editing mode.
	GridModeEdit
)

// editState holds the active edit buffer state.
type editState struct {
	originalValue string // Value before editing began
	buffer        []rune // Current edit buffer
	cursorPos     int    // Text cursor position within buffer
}

// DataGrid is an advanced virtualized table component with cell-level navigation,
// inline editing, and changeset tracking. Built on tview.Box with custom rendering,
// following the same approach as VirtualList.
type DataGrid struct {
	widgetBase

	// Data
	source   DataGridSource // Data provider
	viewport Viewport       // Virtual scroll state

	// Cursor & selection
	cursor       CellPosition // Current cell position
	mode         GridMode     // Current interaction mode
	selectedRows map[int]bool // Multi-select row tracking

	// Editing
	changeset *Changeset // Dirty cell tracking
	editState *editState // Active edit buffer (nil when not editing)

	// Options
	showRowNumbers bool // Show row number gutter
	showHeader     bool // Show column header row
	separator      rune // Column separator character
	overscan       int  // Rows to pre-fetch outside visible area

	// Computed layout (recalculated each Draw)
	colWidths []int // Rendered width of each column

	// Callbacks
	onCellSelect      func(pos CellPosition, cell GridCell)
	onCellEdit        func(pos CellPosition, oldValue, newValue string)
	onModeChange      func(mode GridMode)
	onCursorMove      func(pos CellPosition)
	onChangesetUpdate func(changeset *Changeset)
	onModalEdit       func(pos CellPosition, currentValue string, commit func(string))
	onSearch          func()
	onCopy            func(value string)
	onBack            func()
	onRowSelect       func(row int, data map[string]string)
	onSelectionChange func(rows []int)
	onSubmit          func(changeset *Changeset)

	// deferredCallback is set by input handlers for callbacks that must
	// run after mu is released (external callbacks that may re-enter public methods).
	deferredCallback func()
}

// NewDataGrid creates a new DataGrid component with default settings.
func NewDataGrid() *DataGrid {
	dg := &DataGrid{
		changeset:    NewChangeset(),
		selectedRows: make(map[int]bool),
		showHeader:   true,
		separator:    '│',
		overscan:     5,
	}
	dg.initWidget(tview.NewBox())
	return dg
}

// --- Fluent Configuration API ---

// SetSource sets the data source for the grid.
func (dg *DataGrid) SetSource(source DataGridSource) *DataGrid {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	dg.source = source
	dg.clampCursor()
	return dg
}

// SetShowRowNumbers enables or disables the row number gutter.
func (dg *DataGrid) SetShowRowNumbers(show bool) *DataGrid {
	dg.showRowNumbers = show
	return dg
}

// SetShowHeader enables or disables the column header row.
func (dg *DataGrid) SetShowHeader(show bool) *DataGrid {
	dg.showHeader = show
	return dg
}

// SetSeparator sets the column separator character.
func (dg *DataGrid) SetSeparator(sep rune) *DataGrid {
	dg.separator = sep
	return dg
}

// SetOverscan sets how many rows to pre-fetch outside the visible area.
func (dg *DataGrid) SetOverscan(count int) *DataGrid {
	dg.overscan = count
	return dg
}

// SetOnCellSelect sets the callback for when Enter is pressed on a cell.
func (dg *DataGrid) SetOnCellSelect(fn func(pos CellPosition, cell GridCell)) *DataGrid {
	dg.onCellSelect = fn
	return dg
}

// SetOnCellEdit sets the callback for when a cell edit is committed.
func (dg *DataGrid) SetOnCellEdit(fn func(pos CellPosition, oldValue, newValue string)) *DataGrid {
	dg.onCellEdit = fn
	return dg
}

// SetOnModeChange sets the callback for mode transitions.
func (dg *DataGrid) SetOnModeChange(fn func(mode GridMode)) *DataGrid {
	dg.onModeChange = fn
	return dg
}

// SetOnCursorMove sets the callback for cursor position changes.
func (dg *DataGrid) SetOnCursorMove(fn func(pos CellPosition)) *DataGrid {
	dg.onCursorMove = fn
	return dg
}

// SetOnChangesetUpdate sets the callback for changeset modifications.
func (dg *DataGrid) SetOnChangesetUpdate(fn func(changeset *Changeset)) *DataGrid {
	dg.onChangesetUpdate = fn
	return dg
}

// SetOnModalEdit sets the callback for triggering modal (multi-line) editing.
// The consumer creates a Modal with TextArea and calls commit(newValue) when done.
func (dg *DataGrid) SetOnModalEdit(fn func(pos CellPosition, currentValue string, commit func(string))) *DataGrid {
	dg.onModalEdit = fn
	return dg
}

// SetOnSearch sets the callback for the search trigger (/ key).
func (dg *DataGrid) SetOnSearch(fn func()) *DataGrid {
	dg.onSearch = fn
	return dg
}

// SetOnCopy sets the callback for copying a cell value (y key).
func (dg *DataGrid) SetOnCopy(fn func(value string)) *DataGrid {
	dg.onCopy = fn
	return dg
}

// SetOnBack sets the callback for the back/escape action.
func (dg *DataGrid) SetOnBack(fn func()) *DataGrid {
	dg.onBack = fn
	return dg
}

// SetOnRowSelect sets the callback that fires when the cursor moves to a new row.
// Useful for updating preview panels with the current row data.
func (dg *DataGrid) SetOnRowSelect(fn func(row int, data map[string]string)) *DataGrid {
	dg.onRowSelect = fn
	return dg
}

// SetOnSelectionChange sets the callback that fires when multi-select changes.
func (dg *DataGrid) SetOnSelectionChange(fn func(rows []int)) *DataGrid {
	dg.onSelectionChange = fn
	return dg
}

// SetOnSubmit sets the callback for submitting pending changes.
// The callback receives the current changeset. After a successful apply,
// the consumer should call changeset.Clear() to reset dirty state.
func (dg *DataGrid) SetOnSubmit(fn func(changeset *Changeset)) *DataGrid {
	dg.onSubmit = fn
	return dg
}

// Submit triggers the submit callback if there are pending changes.
// Safe to call from external keybindings or buttons.
func (dg *DataGrid) Submit() {
	dg.mu.RLock()
	cb := dg.onSubmit
	cs := dg.changeset
	dg.mu.RUnlock()

	if cb != nil && cs.HasChanges() {
		cb(cs)
	}
}

// --- Public Accessors ---

// GetMode returns the current grid mode.
func (dg *DataGrid) GetMode() GridMode {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.mode
}

// GetCursor returns the current cursor position.
func (dg *DataGrid) GetCursor() CellPosition {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.cursor
}

// GetChangeset returns the changeset for external inspection.
func (dg *DataGrid) GetChangeset() *Changeset {
	return dg.changeset
}

// GetCellValue returns the display value for a cell, reading through the changeset overlay.
// If the cell is dirty, returns the modified value; otherwise returns the source value.
func (dg *DataGrid) GetCellValue(pos CellPosition) string {
	return dg.getCellValue(pos)
}

// getCellValue is the internal version — does not lock mu.
// Caller must hold mu (read or write) if accessing source.
func (dg *DataGrid) getCellValue(pos CellPosition) string {
	if change := dg.changeset.GetChange(pos); change != nil {
		return change.NewValue
	}
	if dg.source == nil {
		return ""
	}
	return dg.source.Cell(pos.Row, pos.Col).Value
}

// --- Row Data Access ---

// GetCursorRow returns the cursor row as a map of column name to display value (with changeset overlay).
func (dg *DataGrid) GetCursorRow() map[string]string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.getRowData(dg.cursor.Row)
}

// GetCursorRowRaw returns the cursor row as a map of column name to RawValue.
func (dg *DataGrid) GetCursorRowRaw() map[string]any {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.getRowDataRaw(dg.cursor.Row)
}

// GetCursorRowIndex returns the current cursor row index.
func (dg *DataGrid) GetCursorRowIndex() int {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.cursor.Row
}

// --- Multi-Select ---

// ToggleRowSelection toggles selection of the current cursor row.
// Safe to call from outside; also callable internally (caller must hold mu).
func (dg *DataGrid) ToggleRowSelection() {
	row := dg.cursor.Row
	if dg.selectedRows[row] {
		delete(dg.selectedRows, row)
	} else {
		dg.selectedRows[row] = true
	}
}

// SelectAllRows selects all rows.
func (dg *DataGrid) SelectAllRows() {
	if dg.source != nil {
		for i := 0; i < dg.source.RowCount(); i++ {
			dg.selectedRows[i] = true
		}
	}
}

// ClearRowSelection clears all row selections.
func (dg *DataGrid) ClearRowSelection() {
	dg.selectedRows = make(map[int]bool)
}

// IsRowSelected returns whether the given row is selected.
func (dg *DataGrid) IsRowSelected(row int) bool {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.selectedRows[row]
}

// GetSelectedRowIndices returns all selected row indices sorted ascending.
func (dg *DataGrid) GetSelectedRowIndices() []int {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.selectedRowIndices()
}

// GetSelectedRows returns all selected rows as column name to value maps.
func (dg *DataGrid) GetSelectedRows() []map[string]string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	indices := dg.selectedRowIndices()
	result := make([]map[string]string, len(indices))
	for i, row := range indices {
		result[i] = dg.getRowData(row)
	}
	return result
}

// GetSelectedRowsRaw returns all selected rows as column name to RawValue maps.
func (dg *DataGrid) GetSelectedRowsRaw() []map[string]any {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	indices := dg.selectedRowIndices()
	result := make([]map[string]any, len(indices))
	for i, row := range indices {
		result[i] = dg.getRowDataRaw(row)
	}
	return result
}

// selectedRowIndices returns sorted selected row indices. Must be called with mu held.
func (dg *DataGrid) selectedRowIndices() []int {
	indices := make([]int, 0, len(dg.selectedRows))
	for row := range dg.selectedRows {
		indices = append(indices, row)
	}
	for i := 1; i < len(indices); i++ {
		key := indices[i]
		j := i - 1
		for j >= 0 && indices[j] > key {
			indices[j+1] = indices[j]
			j--
		}
		indices[j+1] = key
	}
	return indices
}

// --- Hints ---

// Hints returns mode-appropriate key hints.
func (dg *DataGrid) Hints() []KeyHint {
	dg.mu.RLock()
	mode := dg.mode
	dg.mu.RUnlock()

	if mode == GridModeEdit {
		return []KeyHint{
			{Key: "Esc", Description: "Cancel"},
			{Key: "Enter", Description: "Commit"},
			{Key: "Tab", Description: "Commit & Next"},
		}
	}
	return []KeyHint{
		{Key: "h/j/k/l", Description: "Navigate"},
		{Key: "i/Enter", Description: "Edit"},
		{Key: "e", Description: "Modal Edit"},
		{Key: "u", Description: "Revert Cell"},
		{Key: "y", Description: "Copy"},
		{Key: "/", Description: "Search"},
		{Key: "Space", Description: "Select Row"},
	}
}

// --- Internal Helpers ---

func (dg *DataGrid) getRowData(row int) map[string]string {
	if dg.source == nil {
		return nil
	}
	cols := dg.source.Columns()
	result := make(map[string]string, len(cols))
	for i, col := range cols {
		pos := CellPosition{Row: row, Col: i}
		if change := dg.changeset.GetChange(pos); change != nil {
			result[col.Name] = change.NewValue
		} else {
			result[col.Name] = dg.source.Cell(row, i).Value
		}
	}
	return result
}

func (dg *DataGrid) getRowDataRaw(row int) map[string]any {
	if dg.source == nil {
		return nil
	}
	cols := dg.source.Columns()
	result := make(map[string]any, len(cols))
	for i, col := range cols {
		result[col.Name] = dg.source.Cell(row, i).RawValue
	}
	return result
}

func (dg *DataGrid) clampCursor() {
	if dg.source == nil {
		dg.cursor = CellPosition{}
		return
	}
	rowCount := dg.source.RowCount()
	colCount := dg.source.ColCount()

	if dg.cursor.Row >= rowCount {
		dg.cursor.Row = rowCount - 1
	}
	if dg.cursor.Row < 0 {
		dg.cursor.Row = 0
	}
	if dg.cursor.Col >= colCount {
		dg.cursor.Col = colCount - 1
	}
	if dg.cursor.Col < 0 {
		dg.cursor.Col = 0
	}
}

func (dg *DataGrid) fireSelectionChange() {
	if dg.onSelectionChange != nil {
		dg.mu.RLock()
		indices := dg.selectedRowIndices()
		dg.mu.RUnlock()
		dg.onSelectionChange(indices)
	}
}

func (dg *DataGrid) fireCursorMove() {
	if dg.onCursorMove != nil {
		dg.onCursorMove(dg.cursor)
	}
	if dg.onRowSelect != nil {
		dg.onRowSelect(dg.cursor.Row, dg.getRowData(dg.cursor.Row))
	}
}

// ensureCursorVisible adjusts the viewport so the cursor is visible.
func (dg *DataGrid) ensureCursorVisible() {
	// Vertical
	if dg.cursor.Row < dg.viewport.RowOffset {
		dg.viewport.RowOffset = dg.cursor.Row
	}
	if dg.viewport.VisRows > 0 && dg.cursor.Row >= dg.viewport.RowOffset+dg.viewport.VisRows {
		dg.viewport.RowOffset = dg.cursor.Row - dg.viewport.VisRows + 1
	}

	// Horizontal — adjust ColOffset so cursor column is visible.
	// Left edge is handled here; right edge is handled in ensureCursorColVisible
	// (called from Draw after column widths are computed).
	if dg.source == nil {
		return
	}
	cols := dg.source.Columns()
	if dg.cursor.Col < len(cols) && cols[dg.cursor.Col].Frozen {
		return // Frozen columns are always visible
	}

	nonFrozenIdx := dg.cursorNonFrozenIdx()
	if nonFrozenIdx < dg.viewport.ColOffset {
		dg.viewport.ColOffset = nonFrozenIdx
	}
}

// ensureCursorColVisible adjusts ColOffset so the cursor column fits within
// the visible content area. Must be called after colWidths are computed.
func (dg *DataGrid) ensureCursorColVisible(contentWidth int) {
	if dg.source == nil || len(dg.colWidths) == 0 {
		return
	}
	cols := dg.source.Columns()
	if dg.cursor.Col < len(cols) && cols[dg.cursor.Col].Frozen {
		return
	}

	// Calculate width consumed by frozen columns
	frozenWidth := 0
	for i, col := range cols {
		if col.Frozen && i < len(dg.colWidths) {
			frozenWidth += dg.colWidths[i]
		}
	}
	availForNonFrozen := contentWidth - frozenWidth
	if availForNonFrozen <= 0 {
		return
	}

	cursorNF := dg.cursorNonFrozenIdx()

	// Left edge: already handled by ensureCursorVisible, but re-check
	if cursorNF < dg.viewport.ColOffset {
		dg.viewport.ColOffset = cursorNF
	}

	// Right edge: accumulate widths from ColOffset; if cursor column
	// doesn't fit, increase ColOffset until it does.
	for {
		usedWidth := 0
		cursorVisible := false
		nfIdx := 0
		for i, col := range cols {
			if col.Frozen || i >= len(dg.colWidths) {
				continue
			}
			if nfIdx < dg.viewport.ColOffset {
				nfIdx++
				continue
			}
			usedWidth += dg.colWidths[i]
			if nfIdx == cursorNF {
				if usedWidth <= availForNonFrozen {
					cursorVisible = true
				}
				break
			}
			nfIdx++
		}

		if cursorVisible || dg.viewport.ColOffset >= cursorNF {
			break
		}
		dg.viewport.ColOffset++
	}
}

// cursorNonFrozenIdx returns the non-frozen column index of the cursor column.
func (dg *DataGrid) cursorNonFrozenIdx() int {
	if dg.source == nil {
		return 0
	}
	cols := dg.source.Columns()
	nfIdx := 0
	for i := 0; i < dg.cursor.Col; i++ {
		if i < len(cols) && !cols[i].Frozen {
			nfIdx++
		}
	}
	return nfIdx
}

// Focus handles focus.
func (dg *DataGrid) Focus(delegate func(tview.Primitive)) {
	dg.Box.Focus(delegate)
}

// HasFocus returns whether the component has focus.
func (dg *DataGrid) HasFocus() bool {
	return dg.Box.HasFocus()
}
