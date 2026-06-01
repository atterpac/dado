package core_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

var _ core.Widget     = (*core.List)(nil)
var _ core.KeyHandler = (*core.List)(nil)

func TestList_AddItem_Count(t *testing.T) {
	l := core.NewList()
	l.AddItem("Item A", "", 0, nil)
	l.AddItem("Item B", "", 0, nil)
	assert.Equal(t, 2, l.GetItemCount())
}

func TestList_Renders_Items(t *testing.T) {
	l := core.NewList()
	l.AddItem("Apple", "", 0, nil)
	l.AddItem("Banana", "", 0, nil)
	screen := coretest.DrawWidget(l, 30, 10)
	assert.True(t, screen.ContainsText("Apple"))
	assert.True(t, screen.ContainsText("Banana"))
}

func TestList_HandleKey_Down_MovesSelection(t *testing.T) {
	l := core.NewList()
	l.AddItem("A", "", 0, nil)
	l.AddItem("B", "", 0, nil)
	assert.Equal(t, 0, l.GetCurrentItem())
	consumed := coretest.SimulateKey(l, tcell.KeyDown)
	assert.True(t, consumed)
	assert.Equal(t, 1, l.GetCurrentItem())
}

func TestList_HandleKey_Up_MovesSelection(t *testing.T) {
	l := core.NewList()
	l.AddItem("A", "", 0, nil)
	l.AddItem("B", "", 0, nil)
	l.SetCurrentItem(1)
	coretest.SimulateKey(l, tcell.KeyUp)
	assert.Equal(t, 0, l.GetCurrentItem())
}

func TestList_HandleKey_Down_Clamps(t *testing.T) {
	l := core.NewList()
	l.AddItem("A", "", 0, nil)
	coretest.SimulateKey(l, tcell.KeyDown) // already at last
	assert.Equal(t, 0, l.GetCurrentItem())
}

func TestList_HandleKey_Enter_FiresCallback(t *testing.T) {
	l := core.NewList()
	fired := false
	l.AddItem("A", "", 0, func() { fired = true })
	coretest.SimulateKey(l, tcell.KeyEnter)
	assert.True(t, fired)
}

func TestList_SetCurrentItem(t *testing.T) {
	l := core.NewList()
	l.AddItem("A", "", 0, nil)
	l.AddItem("B", "", 0, nil)
	l.AddItem("C", "", 0, nil)
	l.SetCurrentItem(2)
	assert.Equal(t, 2, l.GetCurrentItem())
}

func TestList_SetSelectedFunc(t *testing.T) {
	l := core.NewList()
	var gotIdx int
	var gotMain string
	l.SetSelectedFunc(func(idx int, main, secondary string, shortcut rune) {
		gotIdx = idx
		gotMain = main
	})
	l.AddItem("Mango", "", 0, nil)
	coretest.SimulateKey(l, tcell.KeyEnter)
	assert.Equal(t, 0, gotIdx)
	assert.Equal(t, "Mango", gotMain)
}

func TestList_SetChangedFunc(t *testing.T) {
	l := core.NewList()
	var changed bool
	l.SetChangedFunc(func(_, _ int, _, _ string, _ rune) { changed = true })
	l.AddItem("A", "", 0, nil)
	l.AddItem("B", "", 0, nil)
	coretest.SimulateKey(l, tcell.KeyDown)
	assert.True(t, changed)
}
