package core_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

// cursorWidget is a widget that requests the terminal cursor during Draw.
type cursorWidget struct {
	coretest.MockWidget
	app  *core.App
	x, y int
	show bool
}

func (c *cursorWidget) Draw(screen tcell.Screen) {
	c.MockWidget.Draw(screen)
	if c.show {
		c.app.ShowCursor(c.x, c.y)
	}
}

func TestApp_ShowCursor_AppliedAfterDraw(t *testing.T) {
	app, screen := newTestApp(20, 10)
	w := &cursorWidget{app: app, x: 3, y: 4, show: true}
	app.SetRoot(w)
	app.Draw()

	cx, cy, visible := screen.SimulationScreen.GetCursor()
	assert.True(t, visible)
	assert.Equal(t, 3, cx)
	assert.Equal(t, 4, cy)
}

func TestApp_Cursor_HiddenByDefault(t *testing.T) {
	app, screen := newTestApp(20, 10)
	app.SetRoot(coretest.NewMockWidget("root"))
	app.Draw()

	_, _, visible := screen.SimulationScreen.GetCursor()
	assert.False(t, visible)
}

func TestApp_Cursor_ResetEachFrame(t *testing.T) {
	app, screen := newTestApp(20, 10)
	w := &cursorWidget{app: app, x: 5, y: 5, show: true}
	app.SetRoot(w)
	app.Draw()
	_, _, visible := screen.SimulationScreen.GetCursor()
	assert.True(t, visible)

	// Widget stops requesting the cursor; next frame it must disappear.
	w.show = false
	app.Draw()
	_, _, visible = screen.SimulationScreen.GetCursor()
	assert.False(t, visible)
}

func TestApp_HideCursor(t *testing.T) {
	app, screen := newTestApp(20, 10)
	app.ShowCursor(2, 2)
	app.HideCursor()
	app.SetRoot(coretest.NewMockWidget("root"))
	app.Draw()
	_, _, visible := screen.SimulationScreen.GetCursor()
	assert.False(t, visible)
}
