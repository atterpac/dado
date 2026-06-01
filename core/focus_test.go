package core_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

func TestFocusManager_FocusedNilInitially(t *testing.T) {
	fm := core.NewFocusManager()
	assert.Nil(t, fm.Focused())
}

func TestFocusManager_Focus_CallsWidgetFocus(t *testing.T) {
	fm := core.NewFocusManager()
	w := coretest.NewMockWidget("a")
	fm.Focus(w)
	assert.True(t, w.HasFocus())
	assert.Equal(t, w, fm.Focused())
}

func TestFocusManager_Focus_BlursPrevious(t *testing.T) {
	fm := core.NewFocusManager()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	fm.Focus(a)
	fm.Focus(b)
	assert.False(t, a.HasFocus(), "previous widget should be blurred")
	assert.True(t, b.HasFocus())
}

func TestFocusManager_Focus_SameWidget_NoBlur(t *testing.T) {
	fm := core.NewFocusManager()
	w := coretest.NewMockWidget("a")
	fm.Focus(w)
	fm.Focus(w) // same widget again
	assert.True(t, w.HasFocus(), "widget should stay focused when re-focused")
}

func TestFocusManager_Focus_Nil_ClearsFocus(t *testing.T) {
	fm := core.NewFocusManager()
	w := coretest.NewMockWidget("a")
	fm.Focus(w)
	fm.Focus(nil)
	assert.Nil(t, fm.Focused())
	assert.False(t, w.HasFocus())
}

func TestFocusManager_Push_SavesCurrent(t *testing.T) {
	fm := core.NewFocusManager()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	fm.Focus(a)
	fm.Push(b)
	assert.Equal(t, b, fm.Focused())
}

func TestFocusManager_Pop_RestoresPrevious(t *testing.T) {
	fm := core.NewFocusManager()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	fm.Focus(a)
	fm.Push(b)
	restored := fm.Pop()
	require.Equal(t, a, restored)
	assert.Equal(t, a, fm.Focused())
	assert.True(t, a.HasFocus())
	assert.False(t, b.HasFocus())
}

func TestFocusManager_Pop_EmptyStack_ReturnsNil(t *testing.T) {
	fm := core.NewFocusManager()
	result := fm.Pop()
	assert.Nil(t, result)
}

func TestFocusManager_OnChange_FiredOnFocus(t *testing.T) {
	fm := core.NewFocusManager()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")

	var gotPrev, gotNext core.Widget
	fm.OnChange(func(prev, next core.Widget) {
		gotPrev = prev
		gotNext = next
	})

	fm.Focus(a)
	assert.Nil(t, gotPrev)
	assert.Equal(t, a, gotNext)

	fm.Focus(b)
	assert.Equal(t, a, gotPrev)
	assert.Equal(t, b, gotNext)
}

func TestFocusManager_OnChange_Unregister(t *testing.T) {
	fm := core.NewFocusManager()
	w := coretest.NewMockWidget("a")

	callCount := 0
	unsub := fm.OnChange(func(_, _ core.Widget) { callCount++ })

	fm.Focus(w)
	assert.Equal(t, 1, callCount)

	unsub()
	fm.Focus(nil)
	assert.Equal(t, 1, callCount, "callback should not fire after unregister")
}

func TestFocusManager_ConcurrentAccess(t *testing.T) {
	fm := core.NewFocusManager()
	widgets := make([]*coretest.MockWidget, 10)
	for i := range widgets {
		widgets[i] = coretest.NewMockWidget("w")
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			fm.Focus(widgets[idx%len(widgets)])
			_ = fm.Focused()
		}(i)
	}
	wg.Wait()
}
