package components

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
)

func newTestERD(t *testing.T) *ERDGraph {
	t.Helper()
	tables := []*ERDTable{
		{ID: "users", Name: "users", Columns: []ERDColumn{{Name: "id"}}},
		{ID: "orders", Name: "orders", Columns: []ERDColumn{{Name: "id"}}},
	}
	g := NewERDGraph()
	g.SetData(tables, nil)
	g.SetRect(0, 0, 60, 20)
	g.Draw(newTestScreen(60, 20)) // compute layout
	return g
}

func TestERDGraph_HandleMouse_DoubleClickActivatesFocused(t *testing.T) {
	g := newTestERD(t)
	g.SetFocusedTable("orders")

	var opened string
	g.SetOnSelect(func(tbl *ERDTable) { opened = tbl.ID })

	// A double-click activates the focused table (Enter-equivalent).
	consumed, capture := g.HandleMouse(core.MouseLeftDoubleClick, mouseAt(5, 5))
	assert.True(t, consumed)
	assert.Nil(t, capture)
	assert.Equal(t, "orders", opened)
}

func TestERDGraph_HandleMouse_ScrollPans(t *testing.T) {
	g := newTestERD(t)
	before := g.offsetY
	consumed, _ := g.HandleMouse(core.MouseScrollDown, mouseAt(5, 5))
	assert.True(t, consumed)
	assert.NotEqual(t, before, g.offsetY)
}

func TestERDGraph_HandleMouse_OutOfRect(t *testing.T) {
	g := newTestERD(t)
	consumed, _ := g.HandleMouse(core.MouseLeftClick, mouseAt(100, 100))
	assert.False(t, consumed)
}
