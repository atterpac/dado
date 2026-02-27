package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InputHandler handles keyboard input for the DataGrid.
func (dg *DataGrid) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	// gg sequence state
	gPressed := false

	return dg.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		dg.mu.Lock()

		if dg.source == nil || dg.source.RowCount() == 0 {
			dg.mu.Unlock()
			return
		}

		if dg.mode == GridModeEdit {
			gPressed = false
			dg.handleEditInput(event)
			dg.mu.Unlock()
			return
		}

		dg.handleNormalInput(event, &gPressed)

		// Collect deferred callback before releasing the lock.
		// External callbacks (onModalEdit, onBack, onSearch, onCopy)
		// may call public methods that take mu, so they must run unlocked.
		deferred := dg.deferredCallback
		dg.deferredCallback = nil
		dg.mu.Unlock()

		if deferred != nil {
			deferred()
		}
	})
}

// handleNormalInput processes input in normal mode.
func (dg *DataGrid) handleNormalInput(event *tcell.EventKey, gPressed *bool) {
	key := event.Key()

	// Check Ctrl combos first
	switch key {
	case tcell.KeyCtrlD:
		*gPressed = false
		dg.halfPageDown()
		return
	case tcell.KeyCtrlU:
		*gPressed = false
		dg.halfPageUp()
		return
	case tcell.KeyCtrlZ:
		*gPressed = false
		dg.revertAllChanges()
		return
	case tcell.KeyCtrlA:
		*gPressed = false
		dg.SelectAllRows()
		return
	case tcell.KeyCtrlS:
		*gPressed = false
		if dg.onSubmit != nil && dg.changeset.HasChanges() {
			cb := dg.onSubmit
			cs := dg.changeset
			dg.deferredCallback = func() { cb(cs) }
		}
		return
	case tcell.KeyUp:
		*gPressed = false
		dg.moveCursorUp()
		return
	case tcell.KeyDown:
		*gPressed = false
		dg.moveCursorDown()
		return
	case tcell.KeyLeft:
		*gPressed = false
		dg.moveCursorLeft()
		return
	case tcell.KeyRight:
		*gPressed = false
		dg.moveCursorRight()
		return
	case tcell.KeyHome:
		*gPressed = false
		dg.moveCursorToFirstRow()
		return
	case tcell.KeyEnd:
		*gPressed = false
		dg.moveCursorToLastRow()
		return
	case tcell.KeyPgUp:
		*gPressed = false
		dg.halfPageUp()
		return
	case tcell.KeyPgDn:
		*gPressed = false
		dg.halfPageDown()
		return
	case tcell.KeyEnter:
		*gPressed = false
		if dg.onCellSelect != nil {
			pos := dg.cursor
			val := dg.getCellValue(pos)
			cb := dg.onCellSelect
			dg.deferredCallback = func() { cb(pos, GridCell{Value: val}) }
		}
		return
	case tcell.KeyEscape:
		*gPressed = false
		if dg.onBack != nil {
			cb := dg.onBack
			dg.deferredCallback = func() { cb() }
		}
		return
	}

	if key != tcell.KeyRune {
		*gPressed = false
		return
	}

	r := event.Rune()

	// Handle gg sequence
	if r == 'g' {
		if *gPressed {
			*gPressed = false
			dg.moveCursorToFirstRow()
			return
		}
		*gPressed = true
		return
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
		}
	case '/':
		if dg.onSearch != nil {
			cb := dg.onSearch
			dg.deferredCallback = func() { cb() }
		}
	case ' ':
		dg.ToggleRowSelection()
	case 'V':
		dg.ClearRowSelection()
	case 'q':
		if dg.onBack != nil {
			cb := dg.onBack
			dg.deferredCallback = func() { cb() }
		}
	}
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

// MouseHandler handles mouse input.
func (dg *DataGrid) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return dg.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()
		if !dg.InRect(mx, my) {
			return false, nil
		}

		dg.mu.Lock()
		defer dg.mu.Unlock()

		x, y, _, _ := dg.GetInnerRect()

		headerHeight := 0
		if dg.showHeader {
			headerHeight = 1
		}
		gutterWidth := 0
		if dg.showRowNumbers {
			rowCount := 0
			if dg.source != nil {
				rowCount = dg.source.RowCount()
			}
			gutterWidth = len(itoa(rowCount)) + 2
			if gutterWidth < 4 {
				gutterWidth = 4
			}
		}

		switch action {
		case tview.MouseLeftClick:
			setFocus(dg)

			clickRow := (my - y - headerHeight) + dg.viewport.RowOffset
			if my-y < headerHeight || dg.source == nil {
				return true, dg
			}

			clickCol := dg.mouseXToCol(mx - x - gutterWidth)
			if clickCol < 0 {
				return true, dg
			}

			if clickRow >= 0 && clickRow < dg.source.RowCount() && clickCol < dg.source.ColCount() {
				prevRow := dg.cursor.Row
				dg.cursor.Row = clickRow
				dg.cursor.Col = clickCol
				if prevRow != dg.cursor.Row {
					dg.fireCursorMove()
				}
			}
			return true, dg

		case tview.MouseLeftDoubleClick:
			if dg.source != nil && dg.mode == GridModeNormal {
				dg.enterEdit()
			}
			return true, dg

		case tview.MouseScrollUp:
			if dg.viewport.RowOffset > 0 {
				dg.viewport.RowOffset--
			}
			return true, dg

		case tview.MouseScrollDown:
			if dg.source != nil && dg.viewport.RowOffset < dg.source.RowCount()-dg.viewport.VisRows {
				dg.viewport.RowOffset++
			}
			return true, dg
		}

		return false, nil
	})
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
