package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/theme"
)

// TableCell defines styling for a single cell.
type TableCell struct {
	Text       string
	Color      tcell.Color   // Explicit color (0 = use Status color or default Fg)
	Status     *theme.Status // Typed status for color/icon (takes precedence over Color if set)
	Align      int           // tview.AlignLeft, AlignCenter, AlignRight
	Expansion  int           // Column expansion factor
	MaxWidth   int           // Maximum width (0 = no limit)
	Selectable bool          // Whether this cell is selectable
}

// Table is an enhanced table wrapper with header support and selection.
type Table struct {
	*tview.Table
	headers           []string
	hasHeader         bool
	multiSelect       bool
	selectedRows      map[int]bool
	onSelect          func(row int)
	onSelectionChange func(rows []int)

	// Key-based selection for stability across row mutations
	selectedKeys  map[string]bool // Key -> selected state
	rowKeyToIndex map[string]int  // Key -> current row index
	rowIndexToKey map[int]string  // Row index -> key (reverse lookup)
}

// NewTable creates a new enhanced table.
func NewTable() *Table {
	t := &Table{
		Table:         tview.NewTable(),
		selectedRows:  make(map[int]bool),
		selectedKeys:  make(map[string]bool),
		rowKeyToIndex: make(map[string]int),
		rowIndexToKey: make(map[int]string),
	}

	t.Table.SetSelectable(true, false)
	t.Table.SetBorders(false)
	t.Table.SetSeparator(' ')
	t.Table.SetBackgroundColor(theme.Bg())

	// Register for automatic theme updates
	theme.Register(t.Table)

	return t
}

// SetHeaders sets column headers (row 0, fixed).
func (t *Table) SetHeaders(headers ...string) *Table {
	t.headers = headers
	t.hasHeader = true
	t.Table.SetFixed(1, 0)

	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetExpansion(1)
		t.Table.SetCell(0, col, cell)
	}

	return t
}

// AddRow adds a data row with default foreground color.
func (t *Table) AddRow(cells ...string) *Table {
	row := t.Table.GetRowCount()
	fg := theme.Fg()

	for col, text := range cells {
		cell := tview.NewTableCell(text).
			SetTextColor(fg).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetSelectable(true)
		t.Table.SetCell(row, col, cell)
	}

	return t
}

// AddColoredRow adds a row with explicit colors per cell.
func (t *Table) AddColoredRow(cells []string, colors []tcell.Color) *Table {
	row := t.Table.GetRowCount()

	for col, text := range cells {
		color := theme.Fg()
		if col < len(colors) && colors[col] != 0 {
			color = colors[col]
		}
		cell := tview.NewTableCell(text).
			SetTextColor(color).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetSelectable(true)
		t.Table.SetCell(row, col, cell)
	}

	return t
}

// AddStyledRow adds a row with full cell styling.
// If Status is set on a cell, its color takes precedence over the Color field.
func (t *Table) AddStyledRow(cells []TableCell) *Table {
	row := t.Table.GetRowCount()

	for col, tc := range cells {
		color := tc.Color
		text := tc.Text
		var cellRef interface{}

		// Status takes precedence over Color
		if tc.Status != nil {
			color = tc.Status.Color()
			cellRef = tc.Status // Store for theme updates
		}
		if color == 0 {
			color = theme.Fg()
		}

		cell := tview.NewTableCell(text).
			SetTextColor(color).
			SetAlign(tc.Align).
			SetExpansion(tc.Expansion).
			SetMaxWidth(tc.MaxWidth).
			SetSelectable(tc.Selectable).
			SetReference(cellRef)
		t.Table.SetCell(row, col, cell)
	}

	return t
}

// ClearRows removes all data rows (keeps headers).
func (t *Table) ClearRows() *Table {
	startRow := 0
	if t.hasHeader {
		startRow = 1
	}

	rowCount := t.Table.GetRowCount()
	for row := rowCount - 1; row >= startRow; row-- {
		t.Table.RemoveRow(row)
	}

	// Clear key mappings
	t.rowKeyToIndex = make(map[string]int)
	t.rowIndexToKey = make(map[int]string)

	t.ClearSelection()
	return t
}

// SetMultiSelect enables/disables multi-selection.
func (t *Table) SetMultiSelect(enabled bool) *Table {
	t.multiSelect = enabled
	return t
}

// ToggleSelection toggles selection of current row.
func (t *Table) ToggleSelection() {
	row, _ := t.Table.GetSelection()
	if t.hasHeader && row == 0 {
		return
	}

	dataIdx := t.tableRowToDataIndex(row)
	key := t.rowIndexToKey[dataIdx]

	if t.selectedRows[row] {
		delete(t.selectedRows, row)
		if key != "" {
			delete(t.selectedKeys, key)
		}
	} else {
		t.selectedRows[row] = true
		if key != "" {
			t.selectedKeys[key] = true
		}
	}

	t.notifySelectionChange()
}

// SelectAll selects all data rows.
func (t *Table) SelectAll() {
	startRow := 0
	if t.hasHeader {
		startRow = 1
	}

	for row := startRow; row < t.Table.GetRowCount(); row++ {
		t.selectedRows[row] = true
		// Also select by key if available
		dataIdx := t.tableRowToDataIndex(row)
		if key := t.rowIndexToKey[dataIdx]; key != "" {
			t.selectedKeys[key] = true
		}
	}

	t.notifySelectionChange()
}

// ClearSelection deselects all rows.
func (t *Table) ClearSelection() {
	t.selectedRows = make(map[int]bool)
	t.selectedKeys = make(map[string]bool)
	t.notifySelectionChange()
}

// GetSelectedRows returns indices of selected rows.
func (t *Table) GetSelectedRows() []int {
	rows := make([]int, 0, len(t.selectedRows))
	for row := range t.selectedRows {
		rows = append(rows, row)
	}
	return rows
}

// IsRowSelected checks if a row is selected.
func (t *Table) IsRowSelected(row int) bool {
	return t.selectedRows[row]
}

// SetOnSelect sets callback for row selection (Enter key).
// The callback receives the 0-based data row index (excluding header).
func (t *Table) SetOnSelect(fn func(row int)) *Table {
	t.onSelect = fn
	t.Table.SetSelectedFunc(func(row, column int) {
		if fn != nil {
			dataIdx := t.tableRowToDataIndex(row)
			if dataIdx >= 0 {
				fn(dataIdx)
			}
		}
	})
	return t
}

// SetOnSelectionChange sets callback when multi-selection changes.
func (t *Table) SetOnSelectionChange(fn func(rows []int)) *Table {
	t.onSelectionChange = fn
	return t
}

// Draw renders the table with theme colors.
func (t *Table) Draw(screen tcell.Screen) {
	// Cache theme colors once at start to avoid lock contention in loops
	bg := theme.Bg()
	accent := theme.Accent()
	selectionStyle := theme.SelectionStyle()

	// Update table background color from theme
	t.Table.SetBackgroundColor(bg)

	// Update all cell backgrounds for theme changes
	rowCount := t.Table.GetRowCount()
	colCount := t.Table.GetColumnCount()
	startRow := 0
	if t.hasHeader {
		startRow = 1
	}

	// Refresh header colors from theme
	if t.hasHeader {
		for col := 0; col < colCount; col++ {
			cell := t.Table.GetCell(0, col)
			if cell != nil {
				cell.SetTextColor(accent)
				cell.SetBackgroundColor(bg)
			}
		}
	}

	// Update data cell backgrounds and refresh text colors for theme changes
	for row := startRow; row < rowCount; row++ {
		if t.multiSelect && t.selectedRows[row] {
			continue // Skip multi-selected rows, handled below
		}
		for col := 0; col < colCount; col++ {
			cell := t.Table.GetCell(row, col)
			if cell != nil {
				cell.SetBackgroundColor(bg)
				// Refresh color for cells with a typed Status reference
				if status, ok := cell.GetReference().(*theme.Status); ok && status != nil {
					cell.SetTextColor(status.Color())
				}
				// Note: We intentionally don't set fg for cells without a reference,
				// as that would overwrite custom colors set via AddRowWithColor().
			}
		}
	}

	// Update selected style with high-contrast colors
	t.Table.SetSelectedStyle(selectionStyle)

	// Mark multi-selected rows
	if t.multiSelect {
		for row := range t.selectedRows {
			for col := 0; col < colCount; col++ {
				cell := t.Table.GetCell(row, col)
				if cell != nil {
					// Visual indicator for multi-selected rows
					cell.SetBackgroundColor(accent)
					cell.SetTextColor(bg)
				}
			}
		}
	}

	t.Table.Draw(screen)
}

// InputHandler handles table input including multi-select.
func (t *Table) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// Handle multi-select keys
		if t.multiSelect {
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case ' ':
					t.ToggleSelection()
					return
				}
			case tcell.KeyCtrlA:
				t.SelectAll()
				return
			}
		}

		// Default table handling
		if handler := t.Table.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}


// notifySelectionChange calls the selection change callback.
func (t *Table) notifySelectionChange() {
	if t.onSelectionChange != nil {
		t.onSelectionChange(t.GetSelectedRows())
	}
}

// GetDataRowCount returns the number of data rows (excluding header).
func (t *Table) GetDataRowCount() int {
	count := t.Table.GetRowCount()
	if t.hasHeader {
		count--
	}
	if count < 0 {
		count = 0
	}
	return count
}

// ScrollToRow scrolls to make the specified row visible.
func (t *Table) ScrollToRow(row int) {
	t.Table.Select(row, 0)
}

// =============================================================================
// Index Conversion Helpers
// =============================================================================

// dataIndexToTableRow converts a 0-based data row index to tview table row.
// Data index 0 = first data row (after header if present).
func (t *Table) dataIndexToTableRow(dataIndex int) int {
	if t.hasHeader {
		return dataIndex + 1
	}
	return dataIndex
}

// tableRowToDataIndex converts a tview table row to 0-based data index.
func (t *Table) tableRowToDataIndex(tableRow int) int {
	if t.hasHeader {
		return tableRow - 1
	}
	return tableRow
}

// =============================================================================
// Row Update Methods
// =============================================================================

// UpdateRow updates an existing row's cells in-place.
// Index is 0-based for data rows (after header).
// Returns error if index is out of bounds.
func (t *Table) UpdateRow(index int, cells ...string) error {
	tableRow := t.dataIndexToTableRow(index)
	if tableRow < 0 || tableRow >= t.Table.GetRowCount() {
		return fmt.Errorf("row index %d out of bounds (have %d data rows)", index, t.GetDataRowCount())
	}
	if t.hasHeader && tableRow == 0 {
		return fmt.Errorf("cannot update header row via UpdateRow")
	}

	fg := theme.Fg()
	for col, text := range cells {
		cell := tview.NewTableCell(text).
			SetTextColor(fg).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetSelectable(true)
		t.Table.SetCell(tableRow, col, cell)
	}

	return nil
}

// UpdateColoredRow updates a row with explicit colors per cell.
// Index is 0-based for data rows (after header).
func (t *Table) UpdateColoredRow(index int, cells []string, colors []tcell.Color) error {
	tableRow := t.dataIndexToTableRow(index)
	if tableRow < 0 || tableRow >= t.Table.GetRowCount() {
		return fmt.Errorf("row index %d out of bounds (have %d data rows)", index, t.GetDataRowCount())
	}
	if t.hasHeader && tableRow == 0 {
		return fmt.Errorf("cannot update header row via UpdateColoredRow")
	}

	for col, text := range cells {
		color := theme.Fg()
		if col < len(colors) && colors[col] != 0 {
			color = colors[col]
		}
		cell := tview.NewTableCell(text).
			SetTextColor(color).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetSelectable(true)
		t.Table.SetCell(tableRow, col, cell)
	}

	return nil
}

// UpdateStyledRow updates a row with full cell styling.
// Index is 0-based for data rows (after header).
func (t *Table) UpdateStyledRow(index int, cells []TableCell) error {
	tableRow := t.dataIndexToTableRow(index)
	if tableRow < 0 || tableRow >= t.Table.GetRowCount() {
		return fmt.Errorf("row index %d out of bounds (have %d data rows)", index, t.GetDataRowCount())
	}
	if t.hasHeader && tableRow == 0 {
		return fmt.Errorf("cannot update header row via UpdateStyledRow")
	}

	for col, tc := range cells {
		color := tc.Color
		text := tc.Text
		var cellRef interface{}

		// Status takes precedence over Color
		if tc.Status != nil {
			color = tc.Status.Color()
			cellRef = tc.Status // Store for theme updates
		}
		if color == 0 {
			color = theme.Fg()
		}

		cell := tview.NewTableCell(text).
			SetTextColor(color).
			SetAlign(tc.Align).
			SetExpansion(tc.Expansion).
			SetMaxWidth(tc.MaxWidth).
			SetSelectable(tc.Selectable).
			SetReference(cellRef)
		t.Table.SetCell(tableRow, col, cell)
	}

	return nil
}

// GetRowData returns the text content of all cells in a row.
// Index is 0-based for data rows (after header).
// Returns nil if index is out of bounds.
func (t *Table) GetRowData(index int) []string {
	tableRow := t.dataIndexToTableRow(index)
	if tableRow < 0 || tableRow >= t.Table.GetRowCount() {
		return nil
	}

	colCount := t.Table.GetColumnCount()
	data := make([]string, colCount)
	for col := 0; col < colCount; col++ {
		cell := t.Table.GetCell(tableRow, col)
		if cell != nil {
			data[col] = cell.Text
		}
	}
	return data
}

// GetRowCells returns TableCell structs for a row (read-only snapshot).
// Index is 0-based for data rows (after header).
// Returns nil if index is out of bounds.
func (t *Table) GetRowCells(index int) []TableCell {
	tableRow := t.dataIndexToTableRow(index)
	if tableRow < 0 || tableRow >= t.Table.GetRowCount() {
		return nil
	}

	colCount := t.Table.GetColumnCount()
	cells := make([]TableCell, colCount)
	for col := 0; col < colCount; col++ {
		cell := t.Table.GetCell(tableRow, col)
		if cell != nil {
			cells[col] = TableCell{
				Text:       cell.Text,
				Color:      cell.Color,
				Align:      cell.Align,
				Expansion:  cell.Expansion,
				MaxWidth:   cell.MaxWidth,
				Selectable: cell.NotSelectable == false,
			}
		}
	}
	return cells
}

// InsertRowAt inserts a new row at the specified index.
// Existing rows at and after index are shifted down.
// Index is 0-based for data rows (after header).
func (t *Table) InsertRowAt(index int, cells ...string) error {
	tableRow := t.dataIndexToTableRow(index)
	maxRow := t.Table.GetRowCount()

	// Allow inserting at end (append)
	if tableRow < 0 || tableRow > maxRow {
		return fmt.Errorf("row index %d out of bounds for insert", index)
	}
	if t.hasHeader && tableRow == 0 {
		return fmt.Errorf("cannot insert before header row")
	}

	// Use tview's InsertRow to shift existing rows
	t.Table.InsertRow(tableRow)

	// Populate the new row
	fg := theme.Fg()
	for col, text := range cells {
		cell := tview.NewTableCell(text).
			SetTextColor(fg).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetSelectable(true)
		t.Table.SetCell(tableRow, col, cell)
	}

	// Update key mappings for shifted rows
	t.adjustKeyMappingsForInsert(index)

	// Update selection indices
	t.adjustSelectionForInsert(index)

	return nil
}

// InsertColoredRowAt inserts a row with explicit colors at the specified index.
func (t *Table) InsertColoredRowAt(index int, cells []string, colors []tcell.Color) error {
	tableRow := t.dataIndexToTableRow(index)
	maxRow := t.Table.GetRowCount()

	if tableRow < 0 || tableRow > maxRow {
		return fmt.Errorf("row index %d out of bounds for insert", index)
	}
	if t.hasHeader && tableRow == 0 {
		return fmt.Errorf("cannot insert before header row")
	}

	t.Table.InsertRow(tableRow)

	for col, text := range cells {
		color := theme.Fg()
		if col < len(colors) && colors[col] != 0 {
			color = colors[col]
		}
		cell := tview.NewTableCell(text).
			SetTextColor(color).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetSelectable(true)
		t.Table.SetCell(tableRow, col, cell)
	}

	t.adjustKeyMappingsForInsert(index)
	t.adjustSelectionForInsert(index)

	return nil
}

// RemoveRowAt removes the row at the specified index.
// Index is 0-based for data rows (after header).
func (t *Table) RemoveRowAt(index int) error {
	tableRow := t.dataIndexToTableRow(index)
	if tableRow < 0 || tableRow >= t.Table.GetRowCount() {
		return fmt.Errorf("row index %d out of bounds", index)
	}
	if t.hasHeader && tableRow == 0 {
		return fmt.Errorf("cannot remove header row")
	}

	// Remove key mapping for this row
	if key := t.rowIndexToKey[index]; key != "" {
		delete(t.rowKeyToIndex, key)
		delete(t.selectedKeys, key)
	}
	delete(t.rowIndexToKey, index)

	// Remove from tview
	t.Table.RemoveRow(tableRow)

	// Update key mappings for shifted rows
	t.adjustKeyMappingsForRemove(index)

	// Update selection indices
	t.adjustSelectionForRemove(index)

	return nil
}

// =============================================================================
// Key-Based Selection
// =============================================================================

// SetRowKey associates a unique key with a row for stable selection.
// Index is 0-based for data rows (after header).
func (t *Table) SetRowKey(index int, key string) {
	// Remove old key for this index if exists
	if oldKey := t.rowIndexToKey[index]; oldKey != "" {
		delete(t.rowKeyToIndex, oldKey)
		// Transfer selection state to new key
		if t.selectedKeys[oldKey] {
			delete(t.selectedKeys, oldKey)
			t.selectedKeys[key] = true
		}
	}

	// Remove old index for this key if exists (key reassignment)
	if oldIdx, ok := t.rowKeyToIndex[key]; ok {
		delete(t.rowIndexToKey, oldIdx)
	}

	// Set new mapping
	t.rowKeyToIndex[key] = index
	t.rowIndexToKey[index] = key

	// If row is currently selected by index, add to selectedKeys
	if t.selectedRows[t.dataIndexToTableRow(index)] {
		t.selectedKeys[key] = true
	}
}

// GetRowByKey returns the current data index for a row key, or -1 if not found.
func (t *Table) GetRowByKey(key string) int {
	if idx, ok := t.rowKeyToIndex[key]; ok {
		return idx
	}
	return -1
}

// GetRowKey returns the key for a data row index, or empty string if not set.
func (t *Table) GetRowKey(index int) string {
	return t.rowIndexToKey[index]
}

// GetSelectedKeys returns keys of all selected rows.
func (t *Table) GetSelectedKeys() []string {
	keys := make([]string, 0, len(t.selectedKeys))
	for key := range t.selectedKeys {
		keys = append(keys, key)
	}
	return keys
}

// SelectByKey selects a row by its key.
func (t *Table) SelectByKey(key string) {
	if idx, ok := t.rowKeyToIndex[key]; ok {
		tableRow := t.dataIndexToTableRow(idx)
		t.selectedRows[tableRow] = true
		t.selectedKeys[key] = true
		t.notifySelectionChange()
	}
}

// DeselectByKey deselects a row by its key.
func (t *Table) DeselectByKey(key string) {
	if idx, ok := t.rowKeyToIndex[key]; ok {
		tableRow := t.dataIndexToTableRow(idx)
		delete(t.selectedRows, tableRow)
		delete(t.selectedKeys, key)
		t.notifySelectionChange()
	}
}

// ToggleSelectionByKey toggles selection of a row by its key.
func (t *Table) ToggleSelectionByKey(key string) {
	if t.selectedKeys[key] {
		t.DeselectByKey(key)
	} else {
		t.SelectByKey(key)
	}
}

// IsKeySelected checks if a row with the given key is selected.
func (t *Table) IsKeySelected(key string) bool {
	return t.selectedKeys[key]
}

// =============================================================================
// Selection Adjustment Helpers
// =============================================================================

// adjustKeyMappingsForInsert updates key mappings after a row insert.
func (t *Table) adjustKeyMappingsForInsert(insertedIndex int) {
	// Rebuild rowKeyToIndex by shifting indices >= insertedIndex
	newKeyToIndex := make(map[string]int)
	for key, idx := range t.rowKeyToIndex {
		if idx >= insertedIndex {
			newKeyToIndex[key] = idx + 1
		} else {
			newKeyToIndex[key] = idx
		}
	}
	t.rowKeyToIndex = newKeyToIndex

	// Rebuild reverse mapping
	t.rowIndexToKey = make(map[int]string)
	for key, idx := range t.rowKeyToIndex {
		t.rowIndexToKey[idx] = key
	}
}

// adjustKeyMappingsForRemove updates key mappings after a row removal.
func (t *Table) adjustKeyMappingsForRemove(removedIndex int) {
	// Shift indices > removedIndex down by 1
	newKeyToIndex := make(map[string]int)
	for key, idx := range t.rowKeyToIndex {
		if idx > removedIndex {
			newKeyToIndex[key] = idx - 1
		} else if idx < removedIndex {
			newKeyToIndex[key] = idx
		}
		// idx == removedIndex: already deleted in RemoveRowAt
	}
	t.rowKeyToIndex = newKeyToIndex

	// Rebuild reverse mapping
	t.rowIndexToKey = make(map[int]string)
	for key, idx := range t.rowKeyToIndex {
		t.rowIndexToKey[idx] = key
	}
}

// adjustSelectionForInsert updates selection indices after an insert.
func (t *Table) adjustSelectionForInsert(insertedIndex int) {
	if len(t.selectedKeys) > 0 {
		// Rebuild from keys (stable)
		t.rebuildSelectionFromKeys()
	} else {
		// Shift indices (fallback for non-keyed tables)
		insertedTableRow := t.dataIndexToTableRow(insertedIndex)
		newSelected := make(map[int]bool)
		for row := range t.selectedRows {
			if row >= insertedTableRow {
				newSelected[row+1] = true
			} else {
				newSelected[row] = true
			}
		}
		t.selectedRows = newSelected
	}
}

// adjustSelectionForRemove updates selection indices after a removal.
func (t *Table) adjustSelectionForRemove(removedIndex int) {
	if len(t.selectedKeys) > 0 {
		// Rebuild from keys (stable)
		t.rebuildSelectionFromKeys()
	} else {
		// Shift indices
		removedTableRow := t.dataIndexToTableRow(removedIndex)
		newSelected := make(map[int]bool)
		for row := range t.selectedRows {
			if row == removedTableRow {
				continue // Removed row loses selection
			}
			if row > removedTableRow {
				newSelected[row-1] = true
			} else {
				newSelected[row] = true
			}
		}
		t.selectedRows = newSelected
	}
}

// rebuildSelectionFromKeys rebuilds selectedRows from selectedKeys.
func (t *Table) rebuildSelectionFromKeys() {
	t.selectedRows = make(map[int]bool)
	for key := range t.selectedKeys {
		if idx, ok := t.rowKeyToIndex[key]; ok {
			tableRow := t.dataIndexToTableRow(idx)
			t.selectedRows[tableRow] = true
		}
	}
}

// =============================================================================
// Convenience Methods (backwards compatibility helpers)
// =============================================================================

// SelectedRow returns the currently selected data row index (0-based).
// Returns -1 if no row is selected or only header is selected.
func (t *Table) SelectedRow() int {
	row, _ := t.Table.GetSelection()
	return t.tableRowToDataIndex(row)
}

// RowCount returns the number of data rows (excluding header).
// Alias for GetDataRowCount().
func (t *Table) RowCount() int {
	return t.GetDataRowCount()
}

// SelectRow selects a data row by index (0-based).
// This sets the table cursor to the specified row.
func (t *Table) SelectRow(index int) {
	tableRow := t.dataIndexToTableRow(index)
	t.Table.Select(tableRow, 0)
}

// AddRowWithColor adds a row where all cells have the same color.
// Convenience method for adding styled rows with uniform color.
func (t *Table) AddRowWithColor(color tcell.Color, cells ...string) int {
	row := t.Table.GetRowCount()

	for col, text := range cells {
		cell := tview.NewTableCell(text).
			SetTextColor(color).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetSelectable(true)
		t.Table.SetCell(row, col, cell)
	}

	return t.tableRowToDataIndex(row)
}

// AddRowWithStatus adds a row with a typed status for a specific column.
// The status column will display with the status color and icon.
// Other cells use the default foreground color.
// The status is stored as a reference for dynamic theme updates.
func (t *Table) AddRowWithStatus(status *theme.Status, statusCol int, cells ...string) int {
	row := t.Table.GetRowCount()
	fg := theme.Fg()

	for col, text := range cells {
		displayText := text
		cellColor := fg
		var cellRef interface{}

		// Apply status color and icon to the designated column
		if col == statusCol && status != nil {
			icon := status.Icon()
			if icon != "" {
				displayText = icon + " " + text
			}
			cellColor = status.Color()
			// Store status for dynamic theme updates in Draw()
			cellRef = status
		}

		cell := tview.NewTableCell(displayText).
			SetTextColor(cellColor).
			SetAlign(tview.AlignLeft).
			SetExpansion(1).
			SetSelectable(true).
			SetReference(cellRef)
		t.Table.SetCell(row, col, cell)
	}

	return t.tableRowToDataIndex(row)
}

// GetCell returns the cell at the specified table row and column.
// Note: row is the table row (including header), not data index.
func (t *Table) GetCell(row, col int) *tview.TableCell {
	return t.Table.GetCell(row, col)
}

// SetSelectionChangedFunc sets the callback for when selection changes.
func (t *Table) SetSelectionChangedFunc(fn func(row, col int)) *Table {
	t.Table.SetSelectionChangedFunc(fn)
	return t
}

// SetSelectedFunc sets the callback when a row is selected with Enter.
func (t *Table) SetSelectedFunc(fn func(row, col int)) *Table {
	t.Table.SetSelectedFunc(fn)
	return t
}
