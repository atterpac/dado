package core

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

// press/release helpers build the raw button events tcell would deliver.
func press(x, y int) *tcell.EventMouse {
	return tcell.NewEventMouse(x, y, tcell.Button1, tcell.ModNone)
}
func release(x, y int) *tcell.EventMouse {
	return tcell.NewEventMouse(x, y, tcell.ButtonNone, tcell.ModNone)
}

func TestResolveMouseAction_ClickSynthesis(t *testing.T) {
	a := &App{}
	assert.Equal(t, MouseLeftDown, a.resolveMouseAction(press(3, 4)))
	assert.Equal(t, MouseLeftClick, a.resolveMouseAction(release(3, 4)))
}

func TestResolveMouseAction_DoubleClick(t *testing.T) {
	a := &App{}
	a.resolveMouseAction(press(3, 4))
	assert.Equal(t, MouseLeftClick, a.resolveMouseAction(release(3, 4)))
	a.resolveMouseAction(press(3, 4))
	// Second release at the same cell within the interval is a double-click.
	assert.Equal(t, MouseLeftDoubleClick, a.resolveMouseAction(release(3, 4)))
}

func TestResolveMouseAction_ReleaseAwayIsNotClick(t *testing.T) {
	a := &App{}
	a.resolveMouseAction(press(3, 4))
	assert.Equal(t, MouseLeftUp, a.resolveMouseAction(release(20, 20)))
}

func TestResolveMouseAction_Wheel(t *testing.T) {
	a := &App{}
	up := tcell.NewEventMouse(0, 0, tcell.WheelUp, tcell.ModNone)
	down := tcell.NewEventMouse(0, 0, tcell.WheelDown, tcell.ModNone)
	assert.Equal(t, MouseScrollUp, a.resolveMouseAction(up))
	assert.Equal(t, MouseScrollDown, a.resolveMouseAction(down))
}

// clickWidget is a minimal MouseHandler+Container root for dispatch tests.
type clickRoot struct {
	Box
	child *clickWidget
}

func (r *clickRoot) Draw(tcell.Screen)              {}
func (r *clickRoot) Children() []Widget             { return []Widget{r.child} }
func (r *clickRoot) DescendantsAt(x, y int) []Widget { return []Widget{r.child} }

type clickWidget struct {
	Box
	consumed bool
}

func (w *clickWidget) Draw(tcell.Screen) {}
func (w *clickWidget) HandleMouse(action MouseAction, ev *tcell.EventMouse) (bool, Widget) {
	return w.consumed, nil
}

func TestDispatchMouse_ConsumedRequestsRedraw(t *testing.T) {
	child := &clickWidget{consumed: true}
	child.SetRect(0, 0, 10, 10)
	root := &clickRoot{child: child}
	a := &App{root: root}

	// A consumed click must request a redraw so selection/focus changes repaint.
	assert.True(t, a.dispatchMouse(press(2, 2)))

	// When the widget does not consume, no redraw is forced.
	child.consumed = false
	a2 := &App{root: &clickRoot{child: &clickWidget{consumed: false}}}
	assert.False(t, a2.dispatchMouse(press(2, 2)))
}

func TestTable_RowAt(t *testing.T) {
	tbl := NewTable()
	for r := 0; r < 5; r++ {
		tbl.SetCell(r, 0, NewTableCell("row"))
	}
	tbl.SetRect(0, 0, 20, 10) // no border/pad: inner == rect, offset 0

	row, ok := tbl.RowAt(0)
	assert.True(t, ok)
	assert.Equal(t, 0, row)

	row, ok = tbl.RowAt(3)
	assert.True(t, ok)
	assert.Equal(t, 3, row)

	_, ok = tbl.RowAt(9) // within rect but beyond 5 rows
	assert.False(t, ok)

	_, ok = tbl.RowAt(-1)
	assert.False(t, ok)
}

func TestList_ItemAt(t *testing.T) {
	l := NewList()
	for i := 0; i < 4; i++ {
		l.AddItem("item", "", 0, nil)
	}
	l.SetRect(0, 0, 20, 10)

	idx, ok := l.ItemAt(2)
	assert.True(t, ok)
	assert.Equal(t, 2, idx)

	_, ok = l.ItemAt(8) // beyond item count
	assert.False(t, ok)
}
