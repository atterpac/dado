package core_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

// newTestApp builds an App backed by a SimulationScreen for unit tests.
func newTestApp(w, h int) (*core.App, *coretest.TestScreen) {
	screen := coretest.NewTestScreen(w, h)
	app := core.NewAppFromScreen(screen.SimulationScreen)
	return app, screen
}

// runApp starts app.Run in a goroutine and returns a stop func.
func runApp(app *core.App) func() {
	done := make(chan struct{})
	go func() {
		_ = app.Run()
		close(done)
	}()
	// Give the loop a moment to start
	time.Sleep(5 * time.Millisecond)
	return func() {
		app.Stop()
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
}

// --- Draw ---

func TestApp_Draw_CallsRoot(t *testing.T) {
	app, _ := newTestApp(20, 10)
	root := coretest.NewMockWidget("root")
	app.SetRoot(root)
	app.Draw()
	assert.Equal(t, 1, root.DrawCount)
}

func TestApp_Draw_SetsRootToScreenSize(t *testing.T) {
	app, _ := newTestApp(40, 20)
	root := coretest.NewMockWidget("root")
	app.SetRoot(root)
	app.Draw()
	x, y, w, h := root.Rect()
	assert.Equal(t, 0, x)
	assert.Equal(t, 0, y)
	assert.Equal(t, 40, w)
	assert.Equal(t, 20, h)
}

func TestApp_Draw_NilRoot_NoPanic(t *testing.T) {
	app, _ := newTestApp(20, 10)
	assert.NotPanics(t, func() { app.Draw() })
}

// --- Queue ---

func TestApp_QueueUpdate_RunsFunction(t *testing.T) {
	app, _ := newTestApp(20, 10)
	root := coretest.NewMockWidget("root")
	app.SetRoot(root)

	stop := runApp(app)
	defer stop()

	called := make(chan struct{})
	app.QueueUpdate(func() { close(called) })

	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("QueueUpdate function was not called")
	}
}

func TestApp_QueueUpdateDraw_RunsAndRedraws(t *testing.T) {
	app, _ := newTestApp(20, 10)
	root := coretest.NewMockWidget("root")
	app.SetRoot(root)

	stop := runApp(app)
	defer stop()

	// Capture drawsBefore and then trigger a QueueUpdateDraw, all on the app goroutine.
	type result struct{ before, after int }
	resCh := make(chan result, 1)

	var drawsBefore int
	app.QueueUpdate(func() {
		drawsBefore = root.DrawCount
	})

	done := make(chan struct{})
	app.QueueUpdateDraw(func() { close(done) })

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("QueueUpdateDraw function was not called")
	}

	// Read DrawCount back on the app goroutine to avoid the race.
	drawsAfter := make(chan int, 1)
	app.QueueUpdate(func() { drawsAfter <- root.DrawCount })
	select {
	case after := <-drawsAfter:
		resCh <- result{drawsBefore, after}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("could not read DrawCount")
	}

	r := <-resCh
	assert.Greater(t, r.after, r.before, "Draw should be called after QueueUpdateDraw")
}

func TestApp_QueueUpdateDraw_SafeFromGoroutine(t *testing.T) {
	app, _ := newTestApp(20, 10)
	app.SetRoot(coretest.NewMockWidget("root"))

	stop := runApp(app)
	defer stop()

	var counter atomic.Int32
	const N = 20
	done := make(chan struct{})
	for i := 0; i < N; i++ {
		go func() {
			app.QueueUpdateDraw(func() {
				if counter.Add(1) == N {
					close(done)
				}
			})
		}()
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("only %d/%d queue callbacks ran", counter.Load(), N)
	}
}

// --- Input dispatch ---

func TestApp_DispatchKey_ToFocused(t *testing.T) {
	app, screen := newTestApp(20, 10)

	root := core.NewFlex()
	focused := coretest.NewMockKeyWidget()
	other := coretest.NewMockKeyWidget()
	root.AddItem(focused, 0, 1, true)
	root.AddItem(other, 0, 1, false)
	app.SetRoot(root)
	app.SetFocus(focused)

	stop := runApp(app)
	defer stop()

	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, tcell.KeyEnter, focused.LastKey())
	assert.Equal(t, tcell.Key(0), other.LastKey(), "non-focused widget must not receive key")
}

func TestApp_DispatchKey_InputCapture_Consumes(t *testing.T) {
	app, screen := newTestApp(20, 10)
	root := coretest.NewMockKeyWidget()
	app.SetRoot(root)
	app.SetFocus(root)
	app.SetInputCapture(func(_ *tcell.EventKey) *tcell.EventKey { return nil })

	stop := runApp(app)
	defer stop()

	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, tcell.Key(0), root.LastKey(), "capture returning nil should stop dispatch")
}

func TestApp_DispatchKey_InputCapture_Transforms(t *testing.T) {
	app, screen := newTestApp(20, 10)
	root := coretest.NewMockKeyWidget()
	app.SetRoot(root)
	app.SetFocus(root)
	// Transform Enter → Down
	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	})

	stop := runApp(app)
	defer stop()

	screen.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)
	time.Sleep(20 * time.Millisecond)

	assert.Equal(t, tcell.KeyDown, root.LastKey())
}

// --- Stop ---

func TestApp_Stop_ExitsRun(t *testing.T) {
	app, _ := newTestApp(20, 10)
	app.SetRoot(coretest.NewMockWidget("root"))

	done := make(chan error)
	go func() { done <- app.Run() }()
	time.Sleep(5 * time.Millisecond)
	app.Stop()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("app.Run did not exit after Stop()")
	}
}

// --- Focus integration ---

func TestApp_SetFocus_GetFocus(t *testing.T) {
	app, _ := newTestApp(20, 10)
	w := coretest.NewMockWidget("w")
	app.SetFocus(w)
	require.Equal(t, w, app.GetFocus())
	assert.True(t, w.HasFocus())
}

func TestApp_FocusManager_Available(t *testing.T) {
	app, _ := newTestApp(20, 10)
	require.NotNil(t, app.Focus())
}
