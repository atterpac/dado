package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

var _ core.Widget = (*core.Pages)(nil)

func TestPages_AddPage_ShowsPage(t *testing.T) {
	p := core.NewPages()
	w := coretest.NewMockWidget("page1")
	p.AddPage("page1", w, true, true)
	name, front := p.GetFrontPage()
	assert.Equal(t, "page1", name)
	assert.Equal(t, w, front)
}

func TestPages_RemovePage_HidesPage(t *testing.T) {
	p := core.NewPages()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	p.AddPage("a", a, true, true)
	p.AddPage("b", b, true, true)
	p.RemovePage("b")
	name, _ := p.GetFrontPage()
	assert.Equal(t, "a", name)
}

func TestPages_GetFrontPage_EmptyReturnsEmpty(t *testing.T) {
	p := core.NewPages()
	name, front := p.GetFrontPage()
	assert.Equal(t, "", name)
	assert.Nil(t, front)
}

func TestPages_ShowPage_SwitchesFront(t *testing.T) {
	p := core.NewPages()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	p.AddPage("a", a, true, true)
	p.AddPage("b", b, true, false) // not shown initially
	p.ShowPage("b")
	name, _ := p.GetFrontPage()
	assert.Equal(t, "b", name)
}

func TestPages_Draw_OnlyFrontPageDrawn(t *testing.T) {
	p := core.NewPages()
	a := coretest.NewMockWidget("a")
	b := coretest.NewMockWidget("b")
	p.AddPage("a", a, true, true)
	p.AddPage("b", b, true, false) // b is not the front
	coretest.DrawWidget(p, 20, 10)
	assert.Equal(t, 1, a.DrawCount)
	assert.Equal(t, 0, b.DrawCount, "non-front page must not be drawn")
}

func TestPages_Draw_FrontPageGetsFullRect(t *testing.T) {
	p := core.NewPages()
	w := coretest.NewMockWidget("w")
	p.AddPage("w", w, true, true)
	coretest.DrawWidget(p, 30, 15)
	x, y, pw, ph := w.Rect()
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
	assert.Equal(t, 30, pw)
	assert.Equal(t, 15, ph)
}

func TestPages_HandleKey_RoutesToFront(t *testing.T) {
	p := core.NewPages()
	front := coretest.NewMockKeyWidget()
	back := coretest.NewMockKeyWidget()
	p.AddPage("front", front, true, true)
	p.AddPage("back", back, true, false)

	consumed := coretest.SimulateRune(p, 'j')
	assert.True(t, consumed)
	assert.Equal(t, 'j', front.LastRune())
	assert.Equal(t, rune(0), back.LastRune(), "non-front page must not receive key")
}

func TestPages_GetPageNames(t *testing.T) {
	p := core.NewPages()
	p.AddPage("a", coretest.NewMockWidget("a"), true, true)
	p.AddPage("b", coretest.NewMockWidget("b"), true, false)
	names := p.GetPageNames()
	require.Len(t, names, 2)
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
}
