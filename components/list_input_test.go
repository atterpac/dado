package components

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
)

// newTestList builds a 6-item list at rect (0,0,20,10). Inner rect equals the
// rect, so item i sits at screen Y = i.
func newTestList() *List {
	l := NewList()
	l.AddItems("a", "b", "c", "d", "e", "f")
	l.SetRect(0, 0, 20, 10)
	return l
}

func TestList_HandleMouse_ClickSelects(t *testing.T) {
	l := newTestList()
	changed := -1
	l.SetOnChange(func(index int, item ListItem) { changed = index })

	consumed, capture := l.HandleMouse(core.MouseLeftClick, mouseAt(1, 3))
	assert.True(t, consumed)
	assert.Nil(t, capture)
	idx, _, _ := l.GetSelected()
	assert.Equal(t, 3, idx)
	assert.Equal(t, 3, changed)
}

func TestList_HandleMouse_DoubleClickActivates(t *testing.T) {
	l := newTestList()
	activated := -1
	l.SetOnSelect(func(index int, item ListItem) { activated = index })

	l.HandleMouse(core.MouseLeftDoubleClick, mouseAt(1, 2))
	idx, _, _ := l.GetSelected()
	assert.Equal(t, 2, idx)
	assert.Equal(t, 2, activated)
}

func TestList_HandleMouse_Scroll(t *testing.T) {
	l := newTestList()
	l.SetSelected(0)

	l.HandleMouse(core.MouseScrollDown, mouseAt(0, 0))
	idx, _, _ := l.GetSelected()
	assert.Equal(t, 1, idx)

	l.HandleMouse(core.MouseScrollUp, mouseAt(0, 0))
	idx, _, _ = l.GetSelected()
	assert.Equal(t, 0, idx)
}

func TestList_HandleMouse_OutOfRect(t *testing.T) {
	l := newTestList()
	consumed, _ := l.HandleMouse(core.MouseLeftClick, mouseAt(100, 100))
	assert.False(t, consumed)
}
