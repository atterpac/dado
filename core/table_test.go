package core_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

var _ core.Widget = (*core.Table)(nil)

func TestTable_SetCell_GetCell(t *testing.T) {
	tbl := core.NewTable()
	cell := core.NewTableCell("hello")
	tbl.SetCell(0, 0, cell)
	got := tbl.GetCell(0, 0)
	require.NotNil(t, got)
	assert.Equal(t, "hello", got.Text)
}

func TestTable_GetCell_Empty_ReturnsDefault(t *testing.T) {
	tbl := core.NewTable()
	got := tbl.GetCell(5, 5)
	require.NotNil(t, got)
	assert.Equal(t, "", got.Text)
}

func TestTable_Draw_RendersContent(t *testing.T) {
	tbl := core.NewTable()
	tbl.SetCell(0, 0, core.NewTableCell("Name"))
	tbl.SetCell(0, 1, core.NewTableCell("Age"))
	tbl.SetCell(1, 0, core.NewTableCell("Alice"))
	tbl.SetCell(1, 1, core.NewTableCell("30"))
	screen := coretest.DrawWidget(tbl, 40, 10)
	assert.True(t, screen.ContainsText("Name"))
	assert.True(t, screen.ContainsText("Alice"))
}

func TestTable_GetRowCount(t *testing.T) {
	tbl := core.NewTable()
	tbl.SetCell(0, 0, core.NewTableCell("r0"))
	tbl.SetCell(2, 0, core.NewTableCell("r2"))
	assert.Equal(t, 3, tbl.GetRowCount())
}

func TestTable_GetColumnCount(t *testing.T) {
	tbl := core.NewTable()
	tbl.SetCell(0, 0, core.NewTableCell("c0"))
	tbl.SetCell(0, 3, core.NewTableCell("c3"))
	assert.Equal(t, 4, tbl.GetColumnCount())
}

func TestTable_Clear_EmptiesCells(t *testing.T) {
	tbl := core.NewTable()
	tbl.SetCell(0, 0, core.NewTableCell("hello"))
	tbl.Clear()
	assert.Equal(t, 0, tbl.GetRowCount())
	assert.Equal(t, 0, tbl.GetColumnCount())
}

func TestTable_Selectable_HandleKey_Down(t *testing.T) {
	tbl := core.NewTable()
	tbl.SetCell(0, 0, core.NewTableCell("r0"))
	tbl.SetCell(1, 0, core.NewTableCell("r1"))
	tbl.SetSelectable(true, false)
	consumed := coretest.SimulateKey(tbl, tcell.KeyDown)
	assert.True(t, consumed)
	row, _ := tbl.GetSelection()
	assert.Equal(t, 1, row)
}

func TestTable_SetSelectedFunc(t *testing.T) {
	tbl := core.NewTable()
	tbl.SetCell(0, 0, core.NewTableCell("x"))
	tbl.SetSelectable(true, false)
	var gotRow, gotCol int
	tbl.SetSelectedFunc(func(row, col int) { gotRow = row; gotCol = col })
	coretest.SimulateKey(tbl, tcell.KeyEnter)
	assert.Equal(t, 0, gotRow)
	assert.Equal(t, 0, gotCol)
}

func TestTable_NewTableCell_Defaults(t *testing.T) {
	cell := core.NewTableCell("text")
	assert.Equal(t, "text", cell.Text)
	assert.Equal(t, core.AlignLeft, cell.Align)
	assert.Equal(t, tcell.ColorDefault, cell.Color)
}
