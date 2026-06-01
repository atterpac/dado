package core

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gdamore/tcell/v2"
)

// App owns the tcell screen, event loop, draw queue, and focus manager.
// QueueUpdateDraw schedules work onto the draw goroutine.
type App struct {
	screen tcell.Screen
	root   Widget
	focus  *FocusManager
	queue  chan func()
	done   chan struct{}
	once   sync.Once // guards done close

	mu           sync.Mutex
	mouseCapture Widget

	inputCapture func(*tcell.EventKey) *tcell.EventKey
	onResize     func(w, h int)
	afterDraw    func(screen tcell.Screen)
}

// NewApp returns a ready App. The tcell screen is initialised lazily in Run().
// This allows construction without a terminal (e.g., in layout.NewApp during tests).
func NewApp() *App {
	return &App{
		focus: NewFocusManager(),
		queue: make(chan func(), 64),
		done:  make(chan struct{}),
	}
}

// NewAppFromScreen constructs an App from an existing tcell.Screen.
// Used in tests with tcell.SimulationScreen.
func NewAppFromScreen(screen tcell.Screen) *App {
	return newApp(screen)
}

func newApp(screen tcell.Screen) *App {
	screen.EnableMouse(tcell.MouseMotionEvents)
	screen.EnablePaste()
	return &App{
		screen: screen,
		focus:  NewFocusManager(),
		queue:  make(chan func(), 64),
		done:   make(chan struct{}),
	}
}

// SetRoot sets the root widget drawn to fill the screen.
func (a *App) SetRoot(w Widget) *App { a.root = w; return a }

// Focus returns the FocusManager. Use to Push/Pop or register OnChange.
func (a *App) Focus() *FocusManager { return a.focus }

// Screen returns the underlying tcell.Screen (for tests and advanced use).
func (a *App) Screen() tcell.Screen { return a.screen }

// SetFocus is shorthand for a.Focus().Focus(w).
func (a *App) SetFocus(w Widget) *App { a.focus.Focus(w); return a }

// GetFocus returns the currently focused widget.
func (a *App) GetFocus() Widget { return a.focus.Focused() }

// QueueUpdate schedules fn on the draw goroutine without forcing a redraw.
// Falls back to direct execution when the event loop is not running (screen nil).
func (a *App) QueueUpdate(fn func()) *App {
	if a.screen == nil {
		fn()
		return a
	}
	a.queue <- fn
	return a
}

// QueueUpdateDraw schedules fn then triggers a redraw.
// Safe to call from any goroutine.
// Falls back to direct execution when the event loop is not running (screen nil).
func (a *App) QueueUpdateDraw(fn func()) *App {
	if a.screen == nil {
		fn()
		return a
	}
	a.queue <- func() {
		fn()
		a.draw()
	}
	return a
}

// Draw triggers a synchronous redraw. Call from the main goroutine only.
func (a *App) Draw() *App { a.draw(); return a }

// SetInputCapture installs a function that intercepts every key event.
// Return nil to consume the event; return a (possibly modified) event to pass it on.
func (a *App) SetInputCapture(fn func(*tcell.EventKey) *tcell.EventKey) *App {
	a.inputCapture = fn
	return a
}

// SetOnResize sets a callback invoked on terminal resize.
func (a *App) SetOnResize(fn func(w, h int)) *App { a.onResize = fn; return a }

// SetAfterDrawFunc installs a callback invoked after the root widget is drawn
// each frame, before the screen is shown. Use it to paint screen-wide overlays
// (e.g. toasts) on top of all content.
func (a *App) SetAfterDrawFunc(fn func(screen tcell.Screen)) *App { a.afterDraw = fn; return a }

// Stop signals the event loop to exit cleanly.
func (a *App) Stop() {
	a.once.Do(func() {
		close(a.done)
	})
}

// Suspend pauses the app, runs fn with the terminal restored, then resumes.
func (a *App) Suspend(fn func()) bool {
	if err := a.screen.Suspend(); err != nil {
		return false
	}
	fn()
	return a.screen.Resume() == nil
}

// Run starts the event loop. Blocks until Stop is called.
// Returns nil on clean shutdown. Initialises the tcell screen on first call.
//
// PollEvent is run in a dedicated goroutine so the main select can
// simultaneously watch the queue channel and the done signal.
func (a *App) Run() error {
	if a.screen == nil {
		screen, err := tcell.NewScreen()
		if err != nil {
			return err
		}
		if err := screen.Init(); err != nil {
			return err
		}
		a.screen = screen
		a.screen.EnableMouse(tcell.MouseMotionEvents)
		a.screen.EnablePaste()
	}

	eventCh := make(chan tcell.Event, 4)
	go func() {
		for {
			ev := a.screen.PollEvent()
			if ev == nil {
				close(eventCh)
				return
			}
			select {
			case eventCh <- ev:
			case <-a.done:
				close(eventCh)
				return
			}
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case <-sigCh:
			a.Stop()
		case <-a.done:
		}
	}()

	a.draw()
	for {
		select {
		case <-a.done:
			a.screen.Fini()
			return nil
		case fn := <-a.queue:
			fn()
			// Drain remaining queue items before blocking.
			for len(a.queue) > 0 {
				(<-a.queue)()
			}
		case ev, ok := <-eventCh:
			if !ok {
				return nil
			}
			switch ev := ev.(type) {
			case *tcell.EventResize:
				a.screen.Sync()
				if a.onResize != nil {
					w, h := ev.Size()
					a.onResize(w, h)
				}
				a.draw()
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlC {
					a.Stop()
					continue
				}
				processed := ev
				if a.inputCapture != nil {
					processed = a.inputCapture(ev)
				}
				if processed != nil {
					a.dispatchKey(processed)
				}
				a.draw()
			case *tcell.EventMouse:
				a.dispatchMouse(ev)
			case *tcell.EventPaste:
				// tcell uses EventPaste as a start/end bracket marker.
				_ = ev
			}
		}
	}
}

func (a *App) draw() {
	if a.root == nil || a.screen == nil {
		return
	}
	w, h := a.screen.Size()
	a.root.SetRect(0, 0, w, h)
	a.screen.Clear()
	a.root.Draw(a.screen)
	if a.afterDraw != nil {
		a.afterDraw(a.screen)
	}
	a.screen.Show()
}

// dispatchKey walks from the focused widget up through Container parents
// until the event is consumed or the root is reached.
func (a *App) dispatchKey(ev *tcell.EventKey) {
	w := a.focus.Focused()
	if w == nil {
		w = a.root
	}
	for w != nil {
		if kh, ok := w.(KeyHandler); ok && kh.HandleKey(ev) {
			return
		}
		w = findParent(a.root, w)
	}
}

func (a *App) dispatchMouse(ev *tcell.EventMouse) {
	a.mu.Lock()
	cap := a.mouseCapture
	a.mu.Unlock()

	action := tcellMouseAction(ev)

	if cap != nil {
		if mh, ok := cap.(MouseHandler); ok {
			consumed, next := mh.HandleMouse(action, ev)
			if !consumed || next == nil {
				a.mu.Lock()
				a.mouseCapture = nil
				a.mu.Unlock()
			}
		}
		return
	}

	if a.root == nil {
		return
	}
	mx, my := ev.Position()
	if c, ok := a.root.(Container); ok {
		for _, w := range c.DescendantsAt(mx, my) {
			if mh, ok := w.(MouseHandler); ok {
				consumed, capture := mh.HandleMouse(action, ev)
				if capture != nil {
					a.mu.Lock()
					a.mouseCapture = capture
					a.mu.Unlock()
				}
				if consumed {
					return
				}
			}
		}
	}
}

func (a *App) dispatchPaste(text string) {
	w := a.focus.Focused()
	if ph, ok := w.(PasteHandler); ok {
		ph.HandlePaste(text)
	}
}

func findParent(root, target Widget) Widget {
	if root == target {
		return nil
	}
	if c, ok := root.(Container); ok {
		for _, child := range c.Children() {
			if child == target {
				return root
			}
			if p := findParent(child, target); p != nil {
				return p
			}
		}
	}
	return nil
}

// tcellMouseAction maps a tcell.EventMouse to a core.MouseAction.
func tcellMouseAction(ev *tcell.EventMouse) MouseAction {
	switch ev.Buttons() {
	case tcell.ButtonNone:
		return MouseMove
	case tcell.Button1:
		return MouseLeftDown
	case tcell.Button2:
		return MouseMiddleDown
	case tcell.Button3:
		return MouseRightDown
	case tcell.WheelUp:
		return MouseScrollUp
	case tcell.WheelDown:
		return MouseScrollDown
	case tcell.WheelLeft:
		return MouseScrollLeft
	case tcell.WheelRight:
		return MouseScrollRight
	}
	return MouseMove
}
