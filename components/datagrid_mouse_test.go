package components

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atterpac/dado/core"
)

// newTestDataGrid builds a 2-column, 8-row grid drawn at (0,0,40,12) so the
// layout geometry (gutter, header, viewport) is established for hit-testing.
func newTestDataGrid(t *testing.T) *DataGrid {
	t.Helper()
	cols := []GridColumn{{Name: "A"}, {Name: "B"}}
	rows := make([][]GridCell, 8)
	for r := range rows {
		rows[r] = []GridCell{
			{Value: "a", RawValue: "a"},
			{Value: "b", RawValue: "b"},
		}
	}
	dg := NewDataGrid()
	dg.SetShowRowNumbers(true)
	dg.SetSource(NewSliceSource(cols, rows))
	dg.SetRect(0, 0, 40, 12)

	screen := newTestScreen(40, 12)
	dg.Draw(screen)
	return dg
}

func TestDataGrid_HandleMouse_ClickMovesCursor(t *testing.T) {
	dg := newTestDataGrid(t)
	moved := -1
	dg.SetOnRowSelect(func(row int, data map[string]string) { moved = row })

	// Header at inner y=0, data row 0 at y=1. Click y=3 → data row 2. Click in
	// the content area (past the gutter) on column B.
	x, y, _, _ := dg.GetInnerRect()
	consumed, capture := dg.HandleMouse(core.MouseLeftClick, mouseAt(x+dg.gutterWidth+1, y+3))
	assert.True(t, consumed)
	assert.Nil(t, capture)
	assert.Equal(t, 2, dg.GetCursor().Row)
	assert.Equal(t, 2, moved)
}

func TestDataGrid_HandleMouse_GutterClickKeepsColumn(t *testing.T) {
	dg := newTestDataGrid(t)
	// Move cursor to column 1 first via a content click.
	x, y, _, _ := dg.GetInnerRect()
	// find an X inside the second column
	col1X := x + dg.gutterWidth + dg.colWidths[0] + 1
	dg.HandleMouse(core.MouseLeftClick, mouseAt(col1X, y+1))
	require.Equal(t, 1, dg.GetCursor().Col)

	// Click the gutter on row 3: column stays 1, row updates.
	dg.HandleMouse(core.MouseLeftClick, mouseAt(x, y+4))
	assert.Equal(t, 3, dg.GetCursor().Row)
	assert.Equal(t, 1, dg.GetCursor().Col)
}

func TestDataGrid_HandleMouse_DoubleClickEdits(t *testing.T) {
	dg := newTestDataGrid(t)
	x, y, _, _ := dg.GetInnerRect()
	dg.HandleMouse(core.MouseLeftDoubleClick, mouseAt(x+dg.gutterWidth+1, y+2))
	assert.Equal(t, GridModeEdit, dg.GetMode())
	assert.Equal(t, 1, dg.GetCursor().Row)
}

func TestDataGrid_HandleMouse_DoubleClickReadOnlySelects(t *testing.T) {
	cols := []GridColumn{{Name: "A"}}
	rows := [][]GridCell{
		{{Value: "x", ReadOnly: true}},
		{{Value: "y", ReadOnly: true}},
	}
	dg := NewDataGrid()
	dg.SetSource(NewSliceSource(cols, rows))
	dg.SetRect(0, 0, 30, 8)
	dg.Draw(newTestScreen(30, 8))

	var selected string
	dg.SetOnCellSelect(func(pos CellPosition, cell GridCell) { selected = cell.Value })

	x, y, _, _ := dg.GetInnerRect()
	dg.HandleMouse(core.MouseLeftDoubleClick, mouseAt(x+dg.gutterWidth+1, y+2)) // row 1
	assert.Equal(t, GridModeNormal, dg.GetMode())
	assert.Equal(t, "y", selected)
}

func TestDataGrid_HandleMouse_Scroll(t *testing.T) {
	dg := newTestDataGrid(t)
	require.Equal(t, 0, dg.GetCursor().Row)

	dg.HandleMouse(core.MouseScrollDown, mouseAt(1, 1))
	assert.Equal(t, mouseScrollRows, dg.GetCursor().Row)

	dg.HandleMouse(core.MouseScrollUp, mouseAt(1, 1))
	assert.Equal(t, 0, dg.GetCursor().Row)
}

func TestDataGrid_HandleMouse_DragSelectsRowRange(t *testing.T) {
	dg := newTestDataGrid(t)
	x, y, _, _ := dg.GetInnerRect()
	cx := x + dg.gutterWidth + 1
	held := func(sy int) *tcell.EventMouse { return tcell.NewEventMouse(cx, sy, tcell.Button1, tcell.ModNone) }

	// Header at y; data row i at y+1+i. Press row 1, drag through rows 2 and 3.
	dg.HandleMouse(core.MouseLeftDown, held(y+2))
	dg.HandleMouse(core.MouseMove, held(y+3))
	changed, _ := dg.HandleMouse(core.MouseMove, held(y+4))
	assert.True(t, changed) // selection extended → repaint warranted
	dg.HandleMouse(core.MouseLeftUp, mouseAt(cx, y+4))

	assert.Equal(t, []int{1, 2, 3}, dg.GetSelectedRowIndices())
	assert.Equal(t, 3, dg.GetCursor().Row)
}

func TestDataGrid_HandleMouse_PlainClickKeepsSelection(t *testing.T) {
	dg := newTestDataGrid(t)
	x, y, _, _ := dg.GetInnerRect()
	cx := x + dg.gutterWidth + 1
	held := func(sy int) *tcell.EventMouse { return tcell.NewEventMouse(cx, sy, tcell.Button1, tcell.ModNone) }

	// Establish a multi-row selection via drag.
	dg.HandleMouse(core.MouseLeftDown, held(y+2))
	dg.HandleMouse(core.MouseMove, held(y+4))
	dg.HandleMouse(core.MouseLeftUp, mouseAt(cx, y+4))
	require.Equal(t, []int{1, 2, 3}, dg.GetSelectedRowIndices())

	// A plain click (press + release, no cross-row move) moves the cursor but
	// leaves the existing selection untouched.
	dg.HandleMouse(core.MouseLeftDown, held(y+6)) // row 5
	dg.HandleMouse(core.MouseLeftClick, mouseAt(cx, y+6))
	assert.Equal(t, 5, dg.GetCursor().Row)
	assert.Equal(t, []int{1, 2, 3}, dg.GetSelectedRowIndices())
}

func TestDataGrid_HandleMouse_HoverDoesNotConsume(t *testing.T) {
	dg := newTestDataGrid(t)
	x, y, _, _ := dg.GetInnerRect()
	// Motion with no button held (not armed) must not consume / force redraws.
	consumed, _ := dg.HandleMouse(core.MouseMove, mouseAt(x+dg.gutterWidth+1, y+3))
	assert.False(t, consumed)
}

func TestDataGrid_HandleMouse_OutOfRect(t *testing.T) {
	dg := newTestDataGrid(t)
	consumed, _ := dg.HandleMouse(core.MouseLeftClick, mouseAt(100, 100))
	assert.False(t, consumed)
}
