package core_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

// Compile-time checks.
var (
	_ core.Widget    = (*core.Flex)(nil)
	_ core.Container = (*core.Flex)(nil)
	_ core.KeyHandler = (*core.Flex)(nil)
)

// --- Layout arithmetic (no screen needed) ---

func TestFlex_Column_FixedItems(t *testing.T) {
	f := core.NewFlex().SetDirection(core.Column)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	f.AddItem(a, 5, 0, false)
	f.AddItem(b, 3, 0, false)
	f.SetRect(0, 0, 20, 20)
	screen := coretest.NewTestScreen(20, 20)
	f.Draw(screen.SimulationScreen)

	ax, ay, aw, ah := a.Rect()
	bx, by, bw, bh := b.Rect()
	assert.Equal(t, 0, ax); assert.Equal(t, 0, ay)
	assert.Equal(t, 20, aw); assert.Equal(t, 5, ah)
	assert.Equal(t, 0, bx); assert.Equal(t, 5, by)
	assert.Equal(t, 20, bw); assert.Equal(t, 3, bh)
}

func TestFlex_Column_ProportionalItems(t *testing.T) {
	f := core.NewFlex().SetDirection(core.Column)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	f.AddItem(a, 0, 1, false)
	f.AddItem(b, 0, 1, false)
	f.SetRect(0, 0, 20, 20)
	coretest.DrawWidget(f, 20, 20)

	_, _, _, ah := a.Rect()
	_, _, _, bh := b.Rect()
	assert.Equal(t, 10, ah)
	assert.Equal(t, 10, bh)
}

func TestFlex_Column_Mixed(t *testing.T) {
	f := core.NewFlex().SetDirection(core.Column)
	top := coretest.NewMockWidget("top")
	mid := coretest.NewMockWidget("mid")
	bot := coretest.NewMockWidget("bot")
	f.AddItem(top, 3, 0, false)
	f.AddItem(mid, 0, 1, false)
	f.AddItem(bot, 2, 0, false)
	f.SetRect(0, 0, 20, 20)
	coretest.DrawWidget(f, 20, 20)

	_, _, _, th := top.Rect()
	_, _, _, mh := mid.Rect()
	_, _, _, bh := bot.Rect()
	assert.Equal(t, 3, th)
	assert.Equal(t, 15, mh) // 20 - 3 - 2
	assert.Equal(t, 2, bh)
}

func TestFlex_Row_Layout(t *testing.T) {
	f := core.NewFlex().SetDirection(core.Row)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	f.AddItem(a, 0, 1, false)
	f.AddItem(b, 0, 1, false)
	f.SetRect(0, 0, 20, 10)
	coretest.DrawWidget(f, 20, 10)

	_, _, aw, ah := a.Rect()
	_, _, bw, bh := b.Rect()
	assert.Equal(t, 10, aw)
	assert.Equal(t, 10, bw)
	assert.Equal(t, 10, ah) // full height
	assert.Equal(t, 10, bh)
}

func TestFlex_Column_LastItemGetsRemainder(t *testing.T) {
	// Odd total height to expose rounding drift
	f := core.NewFlex().SetDirection(core.Column)
	items := make([]*coretest.MockWidget, 3)
	for i := range items {
		items[i] = coretest.NewMockWidget("i")
		f.AddItem(items[i], 0, 1, false)
	}
	f.SetRect(0, 0, 10, 10) // 10 / 3 = 3,3,4 (remainder to last)
	coretest.DrawWidget(f, 10, 10)

	_, _, _, h0 := items[0].Rect()
	_, _, _, h1 := items[1].Rect()
	_, _, _, h2 := items[2].Rect()
	assert.Equal(t, 10, h0+h1+h2, "heights must sum to total")
}

func TestFlex_RemoveItem(t *testing.T) {
	f := core.NewFlex()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	f.AddItem(a, 0, 1, false)
	f.AddItem(b, 0, 1, false)
	f.RemoveItem(a)
	assert.Equal(t, 1, f.ItemCount())
	children := f.Children()
	require.Len(t, children, 1)
	assert.Equal(t, b, children[0])
}

func TestFlex_Clear(t *testing.T) {
	f := core.NewFlex()
	f.AddItem(coretest.NewMockWidget("a"), 0, 1, false)
	f.AddItem(coretest.NewMockWidget("b"), 0, 1, false)
	f.Clear()
	assert.Equal(t, 0, f.ItemCount())
	assert.Empty(t, f.Children())
}

// --- Drawing ---

func TestFlex_Draw_CallsAllChildren(t *testing.T) {
	f := core.NewFlex().SetDirection(core.Column)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	f.AddItem(a, 0, 1, false)
	f.AddItem(b, 0, 1, false)
	coretest.DrawWidget(f, 20, 10)
	assert.Equal(t, 1, a.DrawCount)
	assert.Equal(t, 1, b.DrawCount)
}

func TestFlex_Draw_ZeroSizeItem_NotDrawn(t *testing.T) {
	f := core.NewFlex().SetDirection(core.Column)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	f.AddItem(a, 5, 0, false)
	f.AddItem(b, 0, 0, false) // proportion 0 and fixed 0 → zero size
	f.SetRect(0, 0, 10, 10)
	screen := coretest.NewTestScreen(10, 10)
	f.Draw(screen.SimulationScreen)
	assert.Equal(t, 1, a.DrawCount)
	assert.Equal(t, 0, b.DrawCount)
}

// --- Input routing ---

func TestFlex_HandleKey_RoutesToFocused(t *testing.T) {
	f := core.NewFlex()
	a := coretest.NewMockKeyWidget()
	b := coretest.NewMockKeyWidget()
	f.AddItem(a, 0, 1, true)
	f.AddItem(b, 0, 1, false)
	a.Focus()

	consumed := coretest.SimulateKey(f, tcell.KeyEnter)
	assert.True(t, consumed)
	assert.Equal(t, tcell.KeyEnter, a.LastKey())
	assert.Equal(t, tcell.Key(0), b.LastKey(), "non-focused widget should not receive key")
}

func TestFlex_HandleKey_NoFocused_ReturnsFalse(t *testing.T) {
	f := core.NewFlex()
	f.AddItem(coretest.NewMockWidget("a"), 0, 1, false)
	consumed := coretest.SimulateKey(f, tcell.KeyEnter)
	assert.False(t, consumed)
}

// --- Container interface ---

func TestFlex_Children_ReturnsAll(t *testing.T) {
	f := core.NewFlex()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	f.AddItem(a, 0, 1, false)
	f.AddItem(b, 0, 1, false)
	children := f.Children()
	require.Len(t, children, 2)
	assert.Equal(t, a, children[0])
	assert.Equal(t, b, children[1])
}

func TestFlex_DescendantsAt_HitTest(t *testing.T) {
	f := core.NewFlex().SetDirection(core.Column)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	f.AddItem(a, 5, 0, false)
	f.AddItem(b, 5, 0, false)
	f.SetRect(0, 0, 10, 10)
	coretest.DrawWidget(f, 10, 10)

	// a occupies rows 0-4, b rows 5-9
	hits := f.DescendantsAt(3, 2) // inside a
	require.Len(t, hits, 1)
	assert.Equal(t, a, hits[0])

	hits = f.DescendantsAt(3, 7) // inside b
	require.Len(t, hits, 1)
	assert.Equal(t, b, hits[0])
}

func TestFlex_DescendantsAt_Miss(t *testing.T) {
	f := core.NewFlex().SetDirection(core.Column)
	f.AddItem(coretest.NewMockWidget("a"), 5, 0, false)
	f.SetRect(0, 0, 10, 10)
	coretest.DrawWidget(f, 10, 10)

	hits := f.DescendantsAt(50, 50) // outside
	assert.Empty(t, hits)
}

func TestFlex_DescendantsAt_Nested(t *testing.T) {
	outer := core.NewFlex().SetDirection(core.Column)
	inner := core.NewFlex().SetDirection(core.Row)
	leaf := coretest.NewMockWidget("leaf")
	inner.AddItem(leaf, 0, 1, false)
	outer.AddItem(inner, 0, 1, false)
	outer.SetRect(0, 0, 10, 10)
	coretest.DrawWidget(outer, 10, 10)

	hits := outer.DescendantsAt(3, 3)
	require.GreaterOrEqual(t, len(hits), 2, "should include both inner flex and leaf")
	assert.Equal(t, leaf, hits[0], "deepest widget should be first")
}
