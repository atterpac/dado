package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
)

// mouseScrollRows is how many rows a single scroll-wheel notch moves the
// selection in list-like components.
const mouseScrollRows = 3

// HandleMouse processes mouse input for the Table: single-click selects the row
// under the cursor (moving the selection and firing the change callback),
// double-click activates it (Enter-equivalent), and the scroll wheel moves the
// selection. Clicks on the header row or empty space are consumed but ignored.
func (t *Table) HandleMouse(action core.MouseAction, ev *tcell.EventMouse) (bool, core.Widget) {
	mx, my := ev.Position()
	if !t.Table.InRect(mx, my) {
		return false, nil
	}

	switch action {
	case core.MouseLeftClick:
		if idx, ok := t.dataRowAt(my); ok {
			t.SelectRow(idx)
		}
		return true, nil

	case core.MouseLeftDoubleClick:
		if idx, ok := t.dataRowAt(my); ok {
			t.SelectRow(idx)
			t.Table.Activate()
		}
		return true, nil

	case core.MouseScrollUp:
		t.scrollSelection(-mouseScrollRows)
		return true, nil

	case core.MouseScrollDown:
		t.scrollSelection(mouseScrollRows)
		return true, nil
	}

	return false, nil
}

// dataRowAt returns the 0-based data-row index under screen Y, or ok=false when
// the coordinate maps to the header row or outside the populated rows.
func (t *Table) dataRowAt(screenY int) (int, bool) {
	row, ok := t.Table.RowAt(screenY)
	if !ok {
		return 0, false
	}
	dataIdx := t.tableRowToDataIndex(row)
	if dataIdx < 0 || dataIdx >= t.GetDataRowCount() {
		return 0, false
	}
	return dataIdx, true
}

// scrollSelection moves the selection by delta data rows, clamped to range.
func (t *Table) scrollSelection(delta int) {
	n := t.GetDataRowCount()
	if n == 0 {
		return
	}
	target := t.SelectedRow() + delta
	if target < 0 {
		target = 0
	}
	if target >= n {
		target = n - 1
	}
	t.SelectRow(target)
}
