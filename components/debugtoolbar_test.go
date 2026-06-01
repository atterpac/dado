package components

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
)

func newToolbarHarness() (*core.App, *DebugToolbar, tcell.SimulationScreen) {
	sim := tcell.NewSimulationScreen("UTF-8")
	_ = sim.Init()
	sim.SetSize(80, 24)
	app := core.NewAppFromScreen(sim)
	root := core.NewFlex().SetDirection(core.Column)
	root.AddItem(core.NewTextView(), 0, 1, false)
	app.SetRoot(root)
	root.SetRect(0, 0, 80, 24)
	return app, NewDebugToolbar(app), sim
}

func screenText(sim tcell.SimulationScreen) string {
	cells, w, h := sim.GetContents()
	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			if len(c.Runes) > 0 && c.Runes[0] != 0 {
				b.WriteRune(c.Runes[0])
			} else {
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestDebugToolbar_HiddenByDefault(t *testing.T) {
	app, tb, sim := newToolbarHarness()
	_ = app
	assert.False(t, tb.Visible())
	tb.Draw(sim)
	sim.Show()
	// Nothing drawn by the toolbar when hidden.
	assert.NotContains(t, screenText(sim), "Events")
}

func TestDebugToolbar_ToggleShowsStrip(t *testing.T) {
	_, tb, sim := newToolbarHarness()
	tb.Toggle()
	assert.True(t, tb.Visible())
	tb.Draw(sim)
	sim.Show()
	out := screenText(sim)
	assert.Contains(t, out, "Events")
	assert.Contains(t, out, "Tree")
	assert.Contains(t, out, "Probe")
}

func TestDebugToolbar_TabCyclesTools(t *testing.T) {
	_, tb, _ := newToolbarHarness()
	tb.Toggle() // active = Events (panel)
	assert.Equal(t, 0, tb.active)

	consumed := tb.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	assert.True(t, consumed)
	assert.Equal(t, 1, tb.active) // Tree

	tb.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	assert.Equal(t, 2, tb.active) // Probe (inline)

	tb.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	assert.Equal(t, 0, tb.active) // wraps to Events
}

func TestDebugToolbar_EscHides(t *testing.T) {
	_, tb, _ := newToolbarHarness()
	tb.Toggle()
	assert.True(t, tb.Visible())
	consumed := tb.HandleKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	assert.True(t, consumed)
	assert.False(t, tb.Visible())
}

func TestDebugToolbar_PanelCapturesInlinePassesThrough(t *testing.T) {
	_, tb, _ := newToolbarHarness()
	tb.Toggle() // Events (panel) active

	// A stray rune on a panel tool is swallowed (capture).
	got := tb.HandleKey(tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone))
	assert.True(t, got, "panel tool should capture keys")

	// Switch to the inline probe; keys should pass through to the app.
	tb.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)) // Tree
	tb.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)) // Probe
	got = tb.HandleKey(tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone))
	assert.False(t, got, "inline tool should pass keys through")
}

func TestDebugToolbar_DrawPanelDoesNotPanic(t *testing.T) {
	_, tb, sim := newToolbarHarness()
	tb.Toggle()
	tb.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)) // Tree panel
	assert.NotPanics(t, func() {
		tb.Draw(sim)
		sim.Show()
	})
	assert.Contains(t, screenText(sim), "Widget Tree")
}
