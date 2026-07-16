package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
)

// HandleKey processes a key event for the DataGrid.
func (dg *DataGrid) HandleKey(ev *tcell.EventKey) bool {
	dg.mu.Lock()

	if dg.source == nil || dg.source.RowCount() == 0 {
		dg.mu.Unlock()
		return false
	}

	if dg.mode == GridModeEdit {
		dg.gPressed = false
		dg.handleEditInput(ev)
		dg.mu.Unlock()
		return true
	}

	consumed := dg.handleNormalInput(ev, &dg.gPressed)

	// Collect deferred callback before releasing the lock.
	// External callbacks (onModalEdit, onBack, onSearch, onCopy)
	// may call public methods that take mu, so they must run unlocked.
	deferred := dg.deferredCallback
	dg.deferredCallback = nil
	dg.mu.Unlock()

	if deferred != nil {
		deferred()
	}
	return consumed
}

// handleNormalInput processes input in normal mode. It returns true when the
// grid acted on the event so the caller can report consumption — otherwise the
// app's dispatcher bubbles the key up to a parent container that re-routes it
// back to this (focused) grid, moving the cursor twice for a single press.
func (dg *DataGrid) handleNormalInput(event *tcell.EventKey, gPressed *bool) bool {
	key := event.Key()

	// Check Ctrl combos first
	switch key {
	case tcell.KeyCtrlD:
		*gPressed = false
		dg.halfPageDown()
		return true
	case tcell.KeyCtrlU:
		*gPressed = false
		dg.halfPageUp()
		return true
	case tcell.KeyCtrlZ:
		*gPressed = false
		dg.revertAllChanges()
		return true
	case tcell.KeyCtrlA:
		*gPressed = false
		dg.SelectAllRows()
		return true
	case tcell.KeyCtrlS:
		*gPressed = false
		if dg.onSubmit != nil && dg.changeset.HasChanges() {
			cb := dg.onSubmit
			cs := dg.changeset
			dg.deferredCallback = func() { cb(cs) }
		}
		return true
	case tcell.KeyUp:
		*gPressed = false
		dg.moveCursorUp()
		return true
	case tcell.KeyDown:
		*gPressed = false
		dg.moveCursorDown()
		return true
	case tcell.KeyLeft:
		*gPressed = false
		dg.moveCursorLeft()
		return true
	case tcell.KeyRight:
		*gPressed = false
		dg.moveCursorRight()
		return true
	case tcell.KeyHome:
		*gPressed = false
		dg.moveCursorToFirstRow()
		return true
	case tcell.KeyEnd:
		*gPressed = false
		dg.moveCursorToLastRow()
		return true
	case tcell.KeyPgUp:
		*gPressed = false
		dg.halfPageUp()
		return true
	case tcell.KeyPgDn:
		*gPressed = false
		dg.halfPageDown()
		return true
	case tcell.KeyEnter:
		*gPressed = false
		// Enter edits the current cell when it is editable; otherwise it falls
		// back to the cell-select callback (e.g. read-only grids used as pickers).
		if dg.source != nil && !dg.source.Cell(dg.cursor.Row, dg.cursor.Col).ReadOnly {
			dg.enterEdit()
			return true
		}
		if dg.onCellSelect != nil {
			pos := dg.cursor
			val := dg.getCellValue(pos)
			cb := dg.onCellSelect
			dg.deferredCallback = func() { cb(pos, GridCell{Value: val}) }
		}
		return true
	case tcell.KeyEscape:
		*gPressed = false
		if dg.onBack != nil {
			cb := dg.onBack
			dg.deferredCallback = func() { cb() }
			return true
		}
		return false
	}

	if key != tcell.KeyRune {
		*gPressed = false
		return false
	}

	r := event.Rune()

	// Handle gg sequence
	if r == 'g' {
		if *gPressed {
			*gPressed = false
			dg.moveCursorToFirstRow()
			return true
		}
		*gPressed = true
		return true
	}
	*gPressed = false

	switch r {
	case 'j':
		dg.moveCursorDown()
	case 'k':
		dg.moveCursorUp()
	case 'h':
		dg.moveCursorLeft()
	case 'l':
		dg.moveCursorRight()
	case 'G':
		dg.moveCursorToLastRow()
	case '0':
		dg.moveCursorToFirstCol()
	case '$':
		dg.moveCursorToLastCol()
	case 'i':
		dg.enterEdit()
	case 'c':
		dg.enterEditClear()
	case 'e':
		dg.triggerModalEdit()
	case 'u':
		dg.revertCurrentCell()
	case 'U':
		dg.revertAllChanges()
	case 'y':
		if dg.onCopy != nil {
			value := dg.getCellValue(dg.cursor)
			cb := dg.onCopy
			dg.deferredCallback = func() { cb(value) }
		} else {
			return false
		}
	case '/':
		if dg.onSearch != nil {
			cb := dg.onSearch
			dg.deferredCallback = func() { cb() }
		} else {
			return false
		}
	case ' ':
		dg.ToggleRowSelection()
	case 'V':
		dg.ClearRowSelection()
	case 'q':
		if dg.onBack != nil {
			cb := dg.onBack
			dg.deferredCallback = func() { cb() }
		} else {
			return false
		}
	default:
		// Unhandled rune — let the app bubble it to parent handlers
		// (global shortcuts, page router, etc.).
		return false
	}
	return true
}

// handleEditInput processes input in edit mode (raw event handling for full text input).
func (dg *DataGrid) handleEditInput(event *tcell.EventKey) {
	if dg.editState == nil {
		dg.setMode(GridModeNormal)
		return
	}

	switch event.Key() {
	case tcell.KeyEscape:
		dg.commitEdit()
	case tcell.KeyEnter:
		dg.commitEdit()
	case tcell.KeyTab:
		dg.commitAndMoveNext()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		dg.editBackspace()
	case tcell.KeyDelete:
		dg.editDelete()
	case tcell.KeyLeft:
		dg.editMoveCursorLeft()
	case tcell.KeyRight:
		dg.editMoveCursorRight()
	case tcell.KeyRune:
		dg.editInsertRune(event.Rune())
	}
}

// --- Cursor Navigation ---

func (dg *DataGrid) moveCursorUp() {
	if dg.cursor.Row > 0 {
		prevRow := dg.cursor.Row
		dg.cursor.Row--
		dg.ensureCursorVisible()
		if prevRow != dg.cursor.Row {
			dg.fireCursorMove()
		}
	}
}

func (dg *DataGrid) moveCursorDown() {
	if dg.source != nil && dg.cursor.Row < dg.source.RowCount()-1 {
		prevRow := dg.cursor.Row
		dg.cursor.Row++
		dg.ensureCursorVisible()
		if prevRow != dg.cursor.Row {
			dg.fireCursorMove()
		}
	}
}

func (dg *DataGrid) moveCursorLeft() {
	if dg.cursor.Col > 0 {
		dg.cursor.Col--
		dg.ensureCursorVisible()
	}
}

func (dg *DataGrid) moveCursorRight() {
	if dg.source != nil && dg.cursor.Col < dg.source.ColCount()-1 {
		dg.cursor.Col++
		dg.ensureCursorVisible()
	}
}

func (dg *DataGrid) moveCursorToFirstRow() {
	if dg.cursor.Row != 0 {
		dg.cursor.Row = 0
		dg.ensureCursorVisible()
		dg.fireCursorMove()
	}
}

func (dg *DataGrid) moveCursorToLastRow() {
	if dg.source == nil {
		return
	}
	lastRow := dg.source.RowCount() - 1
	if dg.cursor.Row != lastRow {
		dg.cursor.Row = lastRow
		dg.ensureCursorVisible()
		dg.fireCursorMove()
	}
}

func (dg *DataGrid) moveCursorToFirstCol() {
	if dg.cursor.Col != 0 {
		dg.cursor.Col = 0
		dg.ensureCursorVisible()
	}
}

func (dg *DataGrid) moveCursorToLastCol() {
	if dg.source == nil {
		return
	}
	lastCol := dg.source.ColCount() - 1
	if dg.cursor.Col != lastCol {
		dg.cursor.Col = lastCol
		dg.ensureCursorVisible()
	}
}

func (dg *DataGrid) halfPageDown() {
	if dg.source == nil {
		return
	}
	half := dg.viewport.VisRows / 2
	if half < 1 {
		half = 1
	}
	prevRow := dg.cursor.Row
	dg.cursor.Row += half
	maxRow := dg.source.RowCount() - 1
	if dg.cursor.Row > maxRow {
		dg.cursor.Row = maxRow
	}
	dg.ensureCursorVisible()
	if prevRow != dg.cursor.Row {
		dg.fireCursorMove()
	}
}

func (dg *DataGrid) halfPageUp() {
	half := dg.viewport.VisRows / 2
	if half < 1 {
		half = 1
	}
	prevRow := dg.cursor.Row
	dg.cursor.Row -= half
	if dg.cursor.Row < 0 {
		dg.cursor.Row = 0
	}
	dg.ensureCursorVisible()
	if prevRow != dg.cursor.Row {
		dg.fireCursorMove()
	}
}

// --- Mouse Handler ---

// HandleMouse processes mouse input for the DataGrid. A left press moves the
// cell cursor to the clicked cell (clicks in the row-number gutter keep the
// current column) and arms a drag; dragging the pointer across rows selects
// that row range (into the multi-select set). A double click activates the
// cell — entering inline edit on an editable cell, else firing the cell-select
// callback — and the scroll wheel moves the cursor (and thus the viewport). It
// mirrors the keyboard handler's locking and deferred-callback conventions.
func (dg *DataGrid) HandleMouse(action core.MouseAction, ev *tcell.EventMouse) (bool, core.Widget) {
	mx, my := ev.Position()
	if !dg.InRect(mx, my) {
		if action == core.MouseMove { // a release outside ends any armed drag
			dg.mu.Lock()
			dg.dragArmed, dg.dragging = false, false
			dg.mu.Unlock()
		}
		return false, nil
	}

	switch action {
	case core.MouseLeftDown, core.MouseMove, core.MouseLeftClick, core.MouseLeftUp,
		core.MouseLeftDoubleClick, core.MouseScrollUp, core.MouseScrollDown:
	default:
		return false, nil
	}

	dg.mu.Lock()
	if dg.source == nil || dg.source.RowCount() == 0 {
		dg.mu.Unlock()
		return true, nil
	}

	switch action {
	case core.MouseScrollUp:
		dg.scrollCursor(-mouseScrollRows)
		dg.mu.Unlock()
		return true, nil
	case core.MouseScrollDown:
		dg.scrollCursor(mouseScrollRows)
		dg.mu.Unlock()
		return true, nil

	case core.MouseLeftDown:
		// Commit any active edit, move the cursor to the pressed cell, and arm
		// a drag anchored there. Selection is left untouched until the drag
		// actually engages (see MouseMove).
		if dg.mode == GridModeEdit {
			dg.commitEdit()
		}
		if row, col, ok := dg.cellAtLocked(mx, my); ok {
			prev := dg.cursor
			dg.cursor = CellPosition{Row: row, Col: col}
			dg.clampCursor()
			dg.ensureCursorVisible()
			if prev != dg.cursor {
				dg.fireCursorMove()
			}
			dg.dragArmed = true
			dg.dragging = false
			dg.dragAnchorRow = dg.cursor.Row
		}
		dg.mu.Unlock()
		return true, nil

	case core.MouseMove:
		changed := dg.dragMove(my, ev.Buttons()&tcell.Button1 != 0)
		dg.mu.Unlock()
		return changed, nil
	}

	// Release actions (LeftUp / LeftClick / LeftDoubleClick).
	dragged := dg.dragging
	dg.dragArmed = false
	dg.dragging = false

	// Position the cursor at the released cell (a no-op after the press-time
	// move in the normal down→click sequence, but keeps a bare click correct).
	if action == core.MouseLeftClick || action == core.MouseLeftDoubleClick {
		if row, col, ok := dg.cellAtLocked(mx, my); ok {
			prev := dg.cursor
			dg.cursor = CellPosition{Row: row, Col: col}
			dg.clampCursor()
			dg.ensureCursorVisible()
			if prev != dg.cursor {
				dg.fireCursorMove()
			}
		}
	}

	if action == core.MouseLeftDoubleClick {
		if !dg.source.Cell(dg.cursor.Row, dg.cursor.Col).ReadOnly {
			dg.enterEdit()
		} else if dg.onCellSelect != nil {
			pos := dg.cursor
			val := dg.getCellValue(pos)
			cb := dg.onCellSelect
			dg.deferredCallback = func() { cb(pos, GridCell{Value: val}) }
		}
	}

	// A completed drag notifies the selection-change callback once, after the
	// lock is released (the callback may re-enter public methods).
	if dragged && dg.onSelectionChange != nil {
		cb := dg.onSelectionChange
		idx := dg.selectedRowIndices()
		prev := dg.deferredCallback
		dg.deferredCallback = func() {
			if prev != nil {
				prev()
			}
			cb(idx)
		}
	}

	deferred := dg.deferredCallback
	dg.deferredCallback = nil
	dg.mu.Unlock()

	if deferred != nil {
		deferred()
	}
	return true, nil
}

// dragMove extends a drag selection to the row under screen Y. It engages the
// drag (clearing the prior selection) only once the pointer reaches a row other
// than the anchor, so a plain click never disturbs the selection. Returns
// whether anything changed (i.e. whether a repaint is warranted). buttonHeld
// reports whether the left button is still down; a released button ends the
// drag. Must be called with dg.mu held.
func (dg *DataGrid) dragMove(my int, buttonHeld bool) bool {
	if !dg.dragArmed {
		return false
	}
	if !buttonHeld {
		dg.dragArmed, dg.dragging = false, false
		return false
	}
	r := dg.dragRow(my)
	if !dg.dragging && r == dg.dragAnchorRow {
		return false // still on the anchor row: not yet a drag
	}
	dg.dragging = true
	dg.selectRowRange(dg.dragAnchorRow, r)
	if dg.cursor.Row != r {
		dg.cursor.Row = r
		dg.ensureCursorVisible()
		dg.fireCursorMove()
	}
	return true
}

// dragRow maps a screen Y to a row index clamped to the valid range, so a drag
// that runs past the top or bottom edge selects up to the first/last row. Must
// be called with dg.mu held.
func (dg *DataGrid) dragRow(my int) int {
	_, y, _, _ := dg.GetInnerRect()
	headerHeight := 0
	if dg.showHeader {
		headerHeight = 1
	}
	r := dg.viewport.RowOffset + (my - (y + headerHeight))
	if r < 0 {
		r = 0
	}
	if max := dg.source.RowCount() - 1; r > max {
		r = max
	}
	return r
}

// selectRowRange replaces the selection with the inclusive row range [a, b].
// Must be called with dg.mu held.
func (dg *DataGrid) selectRowRange(a, b int) {
	if a > b {
		a, b = b, a
	}
	dg.selectedRows = make(map[int]bool, b-a+1)
	for r := a; r <= b; r++ {
		dg.selectedRows[r] = true
	}
}

// scrollCursor moves the cursor by delta rows (clamped), which pulls the
// viewport with it via ensureCursorVisible. Must be called with dg.mu held.
func (dg *DataGrid) scrollCursor(delta int) {
	if dg.source == nil {
		return
	}
	prev := dg.cursor.Row
	dg.cursor.Row += delta
	if dg.cursor.Row < 0 {
		dg.cursor.Row = 0
	}
	if max := dg.source.RowCount() - 1; dg.cursor.Row > max {
		dg.cursor.Row = max
	}
	dg.ensureCursorVisible()
	if prev != dg.cursor.Row {
		dg.fireCursorMove()
	}
}

// cellAtLocked maps screen coordinates to a (row, col) cell using the geometry
// retained from the last Draw. A click in the row-number gutter (or past the
// last column) keeps the current column so the row still selects. Returns
// ok=false for the header row or coordinates outside the populated rows. Must
// be called with dg.mu held.
func (dg *DataGrid) cellAtLocked(mx, my int) (row, col int, ok bool) {
	x, y, _, _ := dg.GetInnerRect()
	headerHeight := 0
	if dg.showHeader {
		headerHeight = 1
	}

	relRow := my - (y + headerHeight)
	if relRow < 0 {
		return 0, 0, false // header row or above
	}
	row = dg.viewport.RowOffset + relRow
	if row < 0 || row >= dg.source.RowCount() {
		return 0, 0, false
	}

	col = dg.mouseXToCol(mx - (x + dg.gutterWidth))
	if col < 0 {
		col = dg.cursor.Col // gutter / trailing space: keep column, select row
	}
	return row, col, true
}

// mouseXToCol converts a mouse X position (relative to content area) to a column index.
func (dg *DataGrid) mouseXToCol(relX int) int {
	if relX < 0 || len(dg.colWidths) == 0 {
		return -1
	}

	cols := dg.source.Columns()
	x := 0

	// Check frozen columns first
	for i, col := range cols {
		if !col.Frozen || i >= len(dg.colWidths) {
			continue
		}
		if relX >= x && relX < x+dg.colWidths[i] {
			return i
		}
		x += dg.colWidths[i]
	}

	// Check non-frozen columns
	nonFrozenIdx := 0
	for i, col := range cols {
		if col.Frozen || i >= len(dg.colWidths) {
			continue
		}
		if nonFrozenIdx < dg.viewport.ColOffset {
			nonFrozenIdx++
			continue
		}
		if relX >= x && relX < x+dg.colWidths[i] {
			return i
		}
		x += dg.colWidths[i]
		nonFrozenIdx++
	}

	return -1
}
