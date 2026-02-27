package components

// enterEdit starts inline editing of the current cursor cell.
// The buffer is pre-populated with the cell's current display value.
func (dg *DataGrid) enterEdit() {
	if dg.source == nil {
		return
	}

	cell := dg.source.Cell(dg.cursor.Row, dg.cursor.Col)
	if cell.ReadOnly {
		return
	}

	currentValue := dg.getCellValue(dg.cursor)
	runes := []rune(currentValue)
	dg.editState = &editState{
		originalValue: currentValue,
		buffer:        runes,
		cursorPos:     len(runes),
	}
	dg.setMode(GridModeEdit)
}

// enterEditClear starts editing with an empty buffer (like vim 'c').
func (dg *DataGrid) enterEditClear() {
	if dg.source == nil {
		return
	}

	cell := dg.source.Cell(dg.cursor.Row, dg.cursor.Col)
	if cell.ReadOnly {
		return
	}

	currentValue := dg.getCellValue(dg.cursor)
	dg.editState = &editState{
		originalValue: currentValue,
		buffer:        nil,
		cursorPos:     0,
	}
	dg.setMode(GridModeEdit)
}

// commitEdit commits the current edit buffer and returns to normal mode.
// If the value changed, records it in the changeset and fires callbacks.
func (dg *DataGrid) commitEdit() {
	if dg.editState == nil {
		return
	}

	newValue := string(dg.editState.buffer)
	oldValue := dg.editState.originalValue

	if newValue != oldValue {
		dg.changeset.RecordChange(dg.cursor, oldValue, newValue)

		if dg.onCellEdit != nil {
			dg.onCellEdit(dg.cursor, oldValue, newValue)
		}
		if dg.onChangesetUpdate != nil {
			dg.onChangesetUpdate(dg.changeset)
		}
	}

	dg.editState = nil
	dg.setMode(GridModeNormal)
}

// cancelEdit discards the edit buffer and returns to normal mode.
func (dg *DataGrid) cancelEdit() {
	dg.editState = nil
	dg.setMode(GridModeNormal)
}

// commitAndMoveNext commits the edit and moves to the next cell (Tab behavior).
func (dg *DataGrid) commitAndMoveNext() {
	dg.commitEdit()
	dg.moveCursorRight()
}

// setMode transitions to a new mode and fires the mode change callback.
func (dg *DataGrid) setMode(mode GridMode) {
	if dg.mode == mode {
		return
	}
	dg.mode = mode
	if dg.onModeChange != nil {
		dg.onModeChange(mode)
	}
}

// --- Edit buffer operations ---

func (dg *DataGrid) editInsertRune(r rune) {
	if dg.editState == nil {
		return
	}
	buf := dg.editState.buffer
	pos := dg.editState.cursorPos

	// Insert rune at cursor position
	newBuf := make([]rune, len(buf)+1)
	copy(newBuf, buf[:pos])
	newBuf[pos] = r
	copy(newBuf[pos+1:], buf[pos:])

	dg.editState.buffer = newBuf
	dg.editState.cursorPos++
}

func (dg *DataGrid) editBackspace() {
	if dg.editState == nil || dg.editState.cursorPos == 0 {
		return
	}
	buf := dg.editState.buffer
	pos := dg.editState.cursorPos

	newBuf := make([]rune, len(buf)-1)
	copy(newBuf, buf[:pos-1])
	copy(newBuf[pos-1:], buf[pos:])

	dg.editState.buffer = newBuf
	dg.editState.cursorPos--
}

func (dg *DataGrid) editDelete() {
	if dg.editState == nil || dg.editState.cursorPos >= len(dg.editState.buffer) {
		return
	}
	buf := dg.editState.buffer
	pos := dg.editState.cursorPos

	newBuf := make([]rune, len(buf)-1)
	copy(newBuf, buf[:pos])
	copy(newBuf[pos:], buf[pos+1:])

	dg.editState.buffer = newBuf
}

func (dg *DataGrid) editMoveCursorLeft() {
	if dg.editState != nil && dg.editState.cursorPos > 0 {
		dg.editState.cursorPos--
	}
}

func (dg *DataGrid) editMoveCursorRight() {
	if dg.editState != nil && dg.editState.cursorPos < len(dg.editState.buffer) {
		dg.editState.cursorPos++
	}
}

// --- Revert operations ---

// revertCurrentCell reverts the cell at the cursor position.
func (dg *DataGrid) revertCurrentCell() {
	change := dg.changeset.RevertCell(dg.cursor)
	if change != nil && dg.onChangesetUpdate != nil {
		dg.onChangesetUpdate(dg.changeset)
	}
}

// revertAllChanges reverts all changes in the changeset.
func (dg *DataGrid) revertAllChanges() {
	if !dg.changeset.HasChanges() {
		return
	}
	dg.changeset.RevertAll()
	if dg.onChangesetUpdate != nil {
		dg.onChangesetUpdate(dg.changeset)
	}
}

// triggerModalEdit schedules the modal edit callback to run after the lock is released.
func (dg *DataGrid) triggerModalEdit() {
	if dg.onModalEdit == nil {
		return
	}
	pos := dg.cursor
	currentValue := dg.getCellValue(pos)
	cb := dg.onModalEdit

	commit := func(newValue string) {
		dg.mu.Lock()
		defer dg.mu.Unlock()

		oldValue := currentValue
		if newValue != oldValue {
			dg.changeset.RecordChange(pos, oldValue, newValue)
			if dg.onCellEdit != nil {
				dg.onCellEdit(pos, oldValue, newValue)
			}
			if dg.onChangesetUpdate != nil {
				dg.onChangesetUpdate(dg.changeset)
			}
		}
	}

	dg.deferredCallback = func() {
		cb(pos, currentValue, commit)
	}
}
