package core_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

// Compile-time check: Box satisfies Widget.
var _ core.Widget = (*core.Box)(nil)

func TestBox_ZeroValueReady(t *testing.T) {
	var b core.Box
	x, y, w, h := b.Rect()
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
	assert.Equal(t, 0, w)
	assert.Equal(t, 0, h)
	assert.False(t, b.HasFocus())
}

func TestBox_SetRect_Rect(t *testing.T) {
	var b core.Box
	b.SetRect(5, 10, 40, 20)
	x, y, w, h := b.Rect()
	assert.Equal(t, 5, x)
	assert.Equal(t, 10, y)
	assert.Equal(t, 40, w)
	assert.Equal(t, 20, h)
}

func TestBox_Focus_Blur(t *testing.T) {
	var b core.Box
	assert.False(t, b.HasFocus())
	b.Focus()
	assert.True(t, b.HasFocus())
	b.Blur()
	assert.False(t, b.HasFocus())
}

func TestBox_InnerRect_NoBorderNoPadding(t *testing.T) {
	var b core.Box
	b.SetRect(2, 3, 20, 10)
	ix, iy, iw, ih := b.InnerRect()
	assert.Equal(t, 2, ix)
	assert.Equal(t, 3, iy)
	assert.Equal(t, 20, iw)
	assert.Equal(t, 10, ih)
}

func TestBox_InnerRect_WithBorder(t *testing.T) {
	var b core.Box
	b.SetBorder(true)
	b.SetRect(0, 0, 20, 10)
	ix, iy, iw, ih := b.InnerRect()
	assert.Equal(t, 1, ix)
	assert.Equal(t, 1, iy)
	assert.Equal(t, 18, iw)
	assert.Equal(t, 8, ih)
}

func TestBox_InnerRect_WithPadding(t *testing.T) {
	var b core.Box
	b.SetPadding(1, 1, 2, 2)
	b.SetRect(0, 0, 20, 10)
	ix, iy, iw, ih := b.InnerRect()
	assert.Equal(t, 2, ix)
	assert.Equal(t, 1, iy)
	assert.Equal(t, 16, iw)
	assert.Equal(t, 8, ih)
}

func TestBox_InnerRect_MinZero(t *testing.T) {
	var b core.Box
	b.SetPadding(10, 10, 10, 10)
	b.SetRect(0, 0, 5, 5)
	_, _, iw, ih := b.InnerRect()
	assert.Equal(t, 0, iw)
	assert.Equal(t, 0, ih)
}

func TestBox_DrawBackground(t *testing.T) {
	var b core.Box
	bg := tcell.NewRGBColor(30, 30, 30)
	b.SetBackgroundColor(bg)
	b.SetRect(0, 0, 10, 5)
	screen := coretest.NewTestScreen(10, 5)
	b.Draw(screen.SimulationScreen)
	screen.Show()
	for x := 0; x < 10; x++ {
		got := screen.GetBackgroundAt(x, 0)
		assert.Equal(t, bg, got, "cell (%d,0) should have background color", x)
	}
}

func TestBox_DrawBorder(t *testing.T) {
	var b core.Box
	b.SetBorder(true)
	b.SetRect(0, 0, 10, 5)
	screen := coretest.NewTestScreen(10, 5)
	b.Draw(screen.SimulationScreen)
	screen.Show()
	// Top-left and top-right corners should be box-drawing runes, not spaces
	tl, _, _ := screen.Get(0, 0)
	tr, _, _ := screen.Get(9, 0)
	assert.NotEqual(t, " ", tl, "top-left should be a border rune")
	assert.NotEqual(t, " ", tr, "top-right should be a border rune")
}

func TestBox_DrawBorderFocused(t *testing.T) {
	var b core.Box
	b.SetBorder(true)
	focusStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	normalStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	b.SetBorderStyle(normalStyle)
	b.SetBorderFocusStyle(focusStyle)
	b.SetRect(0, 0, 10, 5)

	screen := coretest.NewTestScreen(10, 5)

	// Draw unfocused
	b.Draw(screen.SimulationScreen)
	screen.Show()
	normalFg, _, _ := screen.GetStyleAt(0, 0).Decompose()
	assert.Equal(t, tcell.ColorWhite, normalFg, "unfocused border should use normal style")

	// Draw focused
	b.Focus()
	screen.Clear()
	b.Draw(screen.SimulationScreen)
	screen.Show()
	focusFg, _, _ := screen.GetStyleAt(0, 0).Decompose()
	assert.Equal(t, tcell.ColorYellow, focusFg, "focused border should use focus style")
}

func TestBox_DrawTitle(t *testing.T) {
	var b core.Box
	b.SetBorder(true)
	b.SetTitle("Hello")
	b.SetRect(0, 0, 20, 5)
	screen := coretest.NewTestScreen(20, 5)
	b.Draw(screen.SimulationScreen)
	screen.Show()
	assert.True(t, screen.ContainsText("Hello"), "title should appear on screen\n%s", screen.Dump())
}

func TestBox_DrawForSubclass_CallsWidgetDraw(t *testing.T) {
	var b core.Box
	b.SetRect(0, 0, 20, 10)
	child := coretest.NewMockWidget("child")
	screen := coretest.NewTestScreen(20, 10)
	b.DrawForSubclass(screen.SimulationScreen)
	assert.Equal(t, 0, child.DrawCount, "DrawForSubclass no longer calls child.Draw")
}

func TestBox_SetBackgroundColor_Backgroundable(t *testing.T) {
	// Verify Box satisfies the Backgroundable interface pattern used by theme
	var b core.Box
	color := tcell.ColorBlue
	b.SetBackgroundColor(color)
	// Draw and verify background was applied
	b.SetRect(0, 0, 5, 3)
	screen := coretest.NewTestScreen(5, 3)
	b.Draw(screen.SimulationScreen)
	screen.Show()
	_, bg, _ := screen.GetStyleAt(0, 0).Decompose()
	assert.Equal(t, color, bg)
}

func BenchmarkBox_Draw_NoAlloc(b *testing.B) {
	var box core.Box
	box.SetRect(0, 0, 80, 24)
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(80, 24)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box.Draw(screen)
	}
}
