package core_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

func TestViewport_NewMatchesInnerRectOrder(t *testing.T) {
	vp := core.NewViewport(2, 3, 10, 5)
	x, y, w, h := vp.Rect()
	assert.Equal(t, []int{2, 3, 10, 5}, []int{x, y, w, h})
}

func TestViewport_Empty(t *testing.T) {
	noWidth := core.NewViewport(0, 0, 0, 5)
	noHeight := core.NewViewport(0, 0, 5, 0)
	ok := core.NewViewport(0, 0, 5, 5)
	assert.True(t, noWidth.Empty())
	assert.True(t, noHeight.Empty())
	assert.False(t, ok.Empty())
}

func TestViewport_OffsetClampedToContent(t *testing.T) {
	vp := core.NewViewport(0, 0, 10, 4)
	vp.SetContentSize(10, 100)
	vp.SetOffset(0, 500) // way past end
	_, oy := vp.Offset()
	assert.Equal(t, 96, oy, "offset clamped to contentH - h")

	vp.SetOffset(0, -5)
	_, oy = vp.Offset()
	assert.Equal(t, 0, oy, "negative offset clamped to 0")
}

func TestViewport_OffsetZeroWhenContentFits(t *testing.T) {
	vp := core.NewViewport(0, 0, 10, 20)
	vp.SetContentSize(10, 5) // smaller than view
	vp.SetOffset(0, 3)
	_, oy := vp.Offset()
	assert.Equal(t, 0, oy)
	mx, my := vp.MaxOffset()
	assert.Equal(t, 0, mx)
	assert.Equal(t, 0, my)
}

func TestViewport_SetContentSizeReclampsOffset(t *testing.T) {
	vp := core.NewViewport(0, 0, 10, 10)
	vp.SetContentSize(10, 100)
	vp.SetOffset(0, 90)
	vp.SetContentSize(10, 20) // shrink content; max offset now 10
	_, oy := vp.Offset()
	assert.Equal(t, 10, oy)
}

func TestViewport_EnsureVisible(t *testing.T) {
	vp := core.NewViewport(0, 0, 10, 5)
	vp.SetContentSize(10, 100)

	// Below the view scrolls down minimally.
	vp.EnsureVisible(0, 20)
	_, oy := vp.Offset()
	assert.Equal(t, 16, oy, "row 20 visible at bottom of height-5 view")

	// Above the view scrolls up to the cell.
	vp.EnsureVisible(0, 3)
	_, oy = vp.Offset()
	assert.Equal(t, 3, oy)

	// Already visible: no change.
	vp.EnsureVisible(0, 5)
	_, oy = vp.Offset()
	assert.Equal(t, 3, oy)
}

func TestViewport_VisibleRowsClampedToContent(t *testing.T) {
	vp := core.NewViewport(0, 0, 10, 10)
	vp.SetContentSize(10, 3) // fewer rows than height
	first, last := vp.VisibleRows()
	assert.Equal(t, 0, first)
	assert.Equal(t, 3, last, "last clamped to content height")

	vp.SetContentSize(10, 100)
	vp.SetOffset(0, 5)
	first, last = vp.VisibleRows()
	assert.Equal(t, 5, first)
	assert.Equal(t, 15, last)
}

func TestViewport_Visible(t *testing.T) {
	vp := core.NewViewport(0, 0, 4, 4)
	vp.SetContentSize(100, 100)
	vp.SetOffset(10, 10)
	assert.True(t, vp.Visible(10, 10))
	assert.True(t, vp.Visible(13, 13))
	assert.False(t, vp.Visible(14, 10), "past right edge")
	assert.False(t, vp.Visible(9, 10), "scrolled off left")
}

// drawToScreen runs fn against a TestScreen-backed tcell.Screen.
func drawToScreen(width, height int, fn func(s tcell.Screen)) *coretest.TestScreen {
	ts := coretest.NewTestScreen(width, height)
	fn(ts.SimulationScreen)
	ts.SimulationScreen.Show()
	return ts
}

func TestViewport_SetContentTranslatesAndClips(t *testing.T) {
	ts := drawToScreen(10, 10, func(s tcell.Screen) {
		vp := core.NewViewport(2, 2, 3, 3) // screen rect offset from origin
		vp.SetContentSize(100, 100)
		vp.SetOffset(5, 5)
		// Content cell (5,5) maps to screen (2,2).
		vp.SetContent(s, 5, 5, 'A', tcell.StyleDefault)
		// Content cell (7,7) maps to screen (4,4) — last visible cell.
		vp.SetContent(s, 7, 7, 'B', tcell.StyleDefault)
		// Out of bounds — must be a no-op.
		vp.SetContent(s, 8, 8, 'X', tcell.StyleDefault)
		vp.SetContent(s, 4, 5, 'Y', tcell.StyleDefault) // scrolled off left
	})
	assert.Equal(t, 'A', runeAt(ts, 2, 2))
	assert.Equal(t, 'B', runeAt(ts, 4, 4))
	// X and Y should not appear anywhere.
	assert.False(t, ts.ContainsText("X"))
	assert.False(t, ts.ContainsText("Y"))
}

func TestViewport_PrintClipsRow(t *testing.T) {
	ts := drawToScreen(20, 5, func(s tcell.Screen) {
		vp := core.NewViewport(0, 0, 5, 1) // only 5 columns wide
		vp.SetContentSize(100, 100)
		vp.Print(s, 0, 0, "HELLOWORLD", tcell.StyleDefault)
	})
	assert.Equal(t, "HELLO", ts.GetRow(0)[:5])
}

func TestViewport_PrintRespectsHorizontalOffset(t *testing.T) {
	ts := drawToScreen(20, 5, func(s tcell.Screen) {
		vp := core.NewViewport(0, 0, 5, 1)
		vp.SetContentSize(100, 100)
		vp.SetOffset(3, 0) // skip first 3 runes
		vp.Print(s, 0, 0, "HELLOWORLD", tcell.StyleDefault)
	})
	assert.Equal(t, "LOWOR", ts.GetRow(0)[:5])
}

func runeAt(ts *coretest.TestScreen, x, y int) rune {
	row := []rune(ts.GetRow(y))
	if x >= len(row) {
		return ' '
	}
	return row[x]
}
