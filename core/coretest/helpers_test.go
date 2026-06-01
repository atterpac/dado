package coretest_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core/coretest"
)

// --- MockWidget compile-time interface checks ---
// These fail until Widget interface and MockWidget exist.

func TestMockWidget_ImplementsWidget(t *testing.T) {
	w := coretest.NewMockWidget("test")
	assert.NotNil(t, w)
	assert.Equal(t, "test", w.Name)
}

func TestMockWidget_DrawIncrementsCount(t *testing.T) {
	w := coretest.NewMockWidget("test")
	screen := coretest.NewTestScreen(20, 10)
	w.SetRect(0, 0, 20, 10)
	w.Draw(screen.SimulationScreen)
	assert.Equal(t, 1, w.DrawCount)
	w.Draw(screen.SimulationScreen)
	assert.Equal(t, 2, w.DrawCount)
}

func TestMockWidget_FocusBlur(t *testing.T) {
	w := coretest.NewMockWidget("test")
	assert.False(t, w.HasFocus())
	w.Focus()
	assert.True(t, w.HasFocus())
	w.Blur()
	assert.False(t, w.HasFocus())
}

func TestMockWidget_RectRoundtrip(t *testing.T) {
	w := coretest.NewMockWidget("test")
	w.SetRect(5, 10, 20, 8)
	x, y, ww, h := w.Rect()
	assert.Equal(t, 5, x)
	assert.Equal(t, 10, y)
	assert.Equal(t, 20, ww)
	assert.Equal(t, 8, h)
}

func TestDrawWidget_RendersAtFullSize(t *testing.T) {
	w := coretest.NewMockWidget("draw-test")
	screen := coretest.DrawWidget(w, 40, 20)
	assert.NotNil(t, screen)
	assert.Equal(t, 1, w.DrawCount)
	sw, sh := screen.Size()
	assert.Equal(t, 40, sw)
	assert.Equal(t, 20, sh)
}

func TestSimulateKey_ReturnsFalseByDefault(t *testing.T) {
	w := coretest.NewMockWidget("test")
	// MockWidget doesn't implement KeyHandler, SimulateKey returns false
	consumed := coretest.SimulateKey(w, tcell.KeyEnter)
	assert.False(t, consumed)
}

func TestSimulateKey_OnKeyHandlerWidget(t *testing.T) {
	w := coretest.NewMockKeyWidget()
	consumed := coretest.SimulateKey(w, tcell.KeyEnter)
	assert.True(t, consumed)
	assert.Equal(t, tcell.KeyEnter, w.LastKey())
}

func TestSimulateRune_OnKeyHandlerWidget(t *testing.T) {
	w := coretest.NewMockKeyWidget()
	consumed := coretest.SimulateRune(w, 'a')
	assert.True(t, consumed)
	assert.Equal(t, 'a', w.LastRune())
}
