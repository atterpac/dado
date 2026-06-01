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
	_ core.Widget     = (*core.Grid)(nil)
	_ core.Container  = (*core.Grid)(nil)
	_ core.KeyHandler = (*core.Grid)(nil)
)

func TestGrid_FixedTracks(t *testing.T) {
	g := core.NewGrid().SetRows(5, 5).SetColumns(8, 12)
	a := coretest.NewMockWidget("a")
	g.AddItem(a, 1, 1, 1, 1, false)
	g.SetRect(0, 0, 20, 10)
	coretest.DrawWidget(g, 20, 10)

	x, y, w, h := a.Rect()
	assert.Equal(t, 8, x)
	assert.Equal(t, 5, y)
	assert.Equal(t, 12, w)
	assert.Equal(t, 5, h)
}

func TestGrid_FlexibleTracks(t *testing.T) {
	g := core.NewGrid().SetRows(0, 0).SetColumns(0, 0)
	a := coretest.NewMockWidget("a")
	g.AddItem(a, 0, 0, 1, 1, false)
	g.SetRect(0, 0, 20, 10)
	coretest.DrawWidget(g, 20, 10)

	_, _, w, h := a.Rect()
	assert.Equal(t, 10, w)
	assert.Equal(t, 5, h)
}

func TestGrid_WeightedColumns(t *testing.T) {
	// columns weighted 1 and 3 → 30/90 split of 120... here total 20 -> 5/15
	g := core.NewGrid().SetRows(10).SetColumns(0, -3)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	g.AddItem(a, 0, 0, 1, 1, false)
	g.AddItem(b, 0, 1, 1, 1, false)
	g.SetRect(0, 0, 20, 10)
	coretest.DrawWidget(g, 20, 10)

	_, _, aw, _ := a.Rect()
	_, _, bw, _ := b.Rect()
	assert.Equal(t, 5, aw)
	assert.Equal(t, 15, bw)
}

func TestGrid_Span(t *testing.T) {
	g := core.NewGrid().SetRows(5, 5).SetColumns(10, 10)
	header := coretest.NewMockWidget("header")
	g.AddItem(header, 0, 0, 1, 2, false) // spans both columns
	g.SetRect(0, 0, 20, 10)
	coretest.DrawWidget(g, 20, 10)

	x, y, w, h := header.Rect()
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
	assert.Equal(t, 20, w)
	assert.Equal(t, 5, h)
}

func TestGrid_Gap(t *testing.T) {
	g := core.NewGrid().SetRows(0, 0).SetColumns(10).SetGap(2, 0)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	g.AddItem(a, 0, 0, 1, 1, false)
	g.AddItem(b, 1, 0, 1, 1, false)
	g.SetRect(0, 0, 10, 10) // 10 - 2 gap = 8 -> 4/4
	coretest.DrawWidget(g, 10, 10)

	_, ay, _, ah := a.Rect()
	by, _ := mustPos(b)
	assert.Equal(t, 0, ay)
	assert.Equal(t, 4, ah)
	assert.Equal(t, 6, by) // 4 + gap 2
}

func mustPos(w *coretest.MockWidget) (y, h int) {
	_, y, _, h = w.Rect()
	return y, h
}

func TestGrid_OutOfRangeItemSkipped(t *testing.T) {
	g := core.NewGrid().SetRows(10).SetColumns(10)
	a := coretest.NewMockWidget("a")
	g.AddItem(a, 5, 5, 1, 1, false) // out of range
	g.SetRect(0, 0, 10, 10)
	coretest.DrawWidget(g, 10, 10)
	assert.Equal(t, 0, a.DrawCount)
}

func TestGrid_RemoveAndClear(t *testing.T) {
	g := core.NewGrid()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	g.AddItem(a, 0, 0, 1, 1, false)
	g.AddItem(b, 0, 1, 1, 1, false)
	g.RemoveItem(a)
	assert.Equal(t, 1, g.ItemCount())
	g.Clear()
	assert.Equal(t, 0, g.ItemCount())
	assert.Empty(t, g.Children())
}

func TestGrid_HandleKey_RoutesToFocused(t *testing.T) {
	g := core.NewGrid().SetRows(10).SetColumns(10, 10)
	a := coretest.NewMockKeyWidget()
	b := coretest.NewMockKeyWidget()
	g.AddItem(a, 0, 0, 1, 1, false)
	g.AddItem(b, 0, 1, 1, 1, false)
	a.Focus()

	consumed := coretest.SimulateKey(g, tcell.KeyEnter)
	assert.True(t, consumed)
	assert.Equal(t, tcell.KeyEnter, a.LastKey())
}

func TestGrid_DescendantsAt(t *testing.T) {
	g := core.NewGrid().SetRows(10).SetColumns(5, 5)
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	g.AddItem(a, 0, 0, 1, 1, false)
	g.AddItem(b, 0, 1, 1, 1, false)
	g.SetRect(0, 0, 10, 10)
	coretest.DrawWidget(g, 10, 10)

	hits := g.DescendantsAt(2, 3)
	require.Len(t, hits, 1)
	assert.Equal(t, a, hits[0])

	hits = g.DescendantsAt(7, 3)
	require.Len(t, hits, 1)
	assert.Equal(t, b, hits[0])
}
