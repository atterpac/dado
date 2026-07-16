package components

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
)

func mouseAt(x, y int) *tcell.EventMouse {
	return tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone)
}

// newTestTable builds a header + 5 data rows at rect (0,0,40,10). With no
// border/padding the inner rect equals the rect and the scroll offset is 0, so
// data row i sits at screen Y = i+1 (row 0 is the header).
func newTestTable() *Table {
	tbl := NewTable()
	tbl.SetHeaders("Name")
	for i := 0; i < 5; i++ {
		tbl.AddRow("row")
	}
	tbl.SetRect(0, 0, 40, 10)
	return tbl
}

func TestTable_HandleMouse_ClickSelects(t *testing.T) {
	tbl := newTestTable()
	var changed int
	tbl.SetSelectionChangedFunc(func(row, col int) { changed = row })

	consumed, capture := tbl.HandleMouse(core.MouseLeftClick, mouseAt(2, 3)) // data row 2
	assert.True(t, consumed)
	assert.Nil(t, capture)
	assert.Equal(t, 2, tbl.SelectedRow())
	assert.Equal(t, 3, changed) // table row (header + data idx 2)
}

func TestTable_HandleMouse_ClickHeaderIgnored(t *testing.T) {
	tbl := newTestTable()
	tbl.SelectRow(1)
	consumed, _ := tbl.HandleMouse(core.MouseLeftClick, mouseAt(2, 0)) // header row
	assert.True(t, consumed)          // consumed (in-rect) ...
	assert.Equal(t, 1, tbl.SelectedRow()) // ... but selection unchanged
}

func TestTable_HandleMouse_DoubleClickActivates(t *testing.T) {
	tbl := newTestTable()
	activated := -1
	tbl.SetOnSelect(func(row int) { activated = row })

	tbl.HandleMouse(core.MouseLeftDoubleClick, mouseAt(2, 2)) // data row 1
	assert.Equal(t, 1, tbl.SelectedRow())
	assert.Equal(t, 1, activated)
}

func TestTable_HandleMouse_ScrollMovesSelection(t *testing.T) {
	tbl := newTestTable()
	tbl.SelectRow(0)

	tbl.HandleMouse(core.MouseScrollDown, mouseAt(1, 1))
	assert.Equal(t, mouseScrollRows, tbl.SelectedRow()) // clamped move down

	tbl.HandleMouse(core.MouseScrollUp, mouseAt(1, 1))
	assert.Equal(t, 0, tbl.SelectedRow())
}

func TestTable_HandleMouse_OutOfRect(t *testing.T) {
	tbl := newTestTable()
	consumed, _ := tbl.HandleMouse(core.MouseLeftClick, mouseAt(100, 100))
	assert.False(t, consumed)
}
