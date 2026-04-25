package layout

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/effect"
	"github.com/atterpac/jig/nav"
	"github.com/atterpac/jig/theme"
)

// AppConfig configures the application layout.
type AppConfig struct {
	// TopBar is shown at the top (e.g., status bar). Can be nil.
	TopBar tview.Primitive

	// ShowCrumbs enables the breadcrumb bar below TopBar.
	ShowCrumbs bool

	// BottomBar is shown at the bottom (e.g., Menu). Can be nil.
	BottomBar tview.Primitive

	// TopBarHeight is the height of the top bar (default: 3).
	TopBarHeight int

	// BottomBarHeight is the height of the bottom bar (default: 1).
	BottomBarHeight int

	// OnComponentChange is called when the active component changes.
	// Useful for updating menu hints, crumbs, etc.
	OnComponentChange func(nav.Component)

	// Debug enables the bus debug overlay. When true, pressing DebugKey
	// pushes a DebugOverlay onto the page stack.
	Debug bool

	// DebugKey is the key that toggles the debug overlay. Defaults to
	// Ctrl+D when zero.
	DebugKey tcell.Key

	// EffectShutdownTimeout bounds how long Stop waits for in-flight
	// Effects to drain. Zero uses defaultEffectShutdownTimeout (2s).
	// Negative disables the wait (fire-and-forget Shutdown).
	EffectShutdownTimeout time.Duration
}

const defaultEffectShutdownTimeout = 2 * time.Second

// App is the application root that manages the overall layout.
type App struct {
	app              *tview.Application
	main             *tview.Flex
	topBar           tview.Primitive
	crumbs           *nav.Crumbs
	pages            *nav.Pages
	menu             *Menu
	config           AppConfig
	userInputCapture func(*tcell.EventKey) *tcell.EventKey // User's custom input capture
	effects          *effect.Dispatcher
	subs             components.Subscriptions
}

// NewApp creates a new application with the given configuration.
func NewApp(config AppConfig) *App {
	// Apply defaults
	if config.TopBarHeight == 0 {
		config.TopBarHeight = 3
	}
	if config.BottomBarHeight == 0 {
		config.BottomBarHeight = 1
	}
	if config.Debug && config.DebugKey == 0 {
		config.DebugKey = tcell.KeyCtrlD
	}

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	mainFlex.SetBackgroundColor(theme.Bg())

	a := &App{
		app:     tview.NewApplication(),
		main:    mainFlex,
		pages:   nav.NewPages(),
		config:  config,
		effects: effect.NewDispatcher(),
	}

	// Register main flex for automatic theme updates
	a.subs.Add(theme.Register(mainFlex))

	// Register with theme system
	theme.SetApp(a.app)

	// Set app reference in pages for focus management
	a.pages.SetApplication(a.app)

	// Build layout
	a.buildLayout()

	// Set up automatic modal input handling
	a.setupModalInputCapture()

	// Set up page change handler
	a.pages.SetOnChange(func(c nav.Component) {
		// Update menu hints
		if a.menu != nil && c != nil {
			a.menu.SetHints(c.Hints())
		}

		// Notify app
		if a.config.OnComponentChange != nil {
			a.config.OnComponentChange(c)
		}
	})

	return a
}

// buildLayout constructs the layout structure.
func (a *App) buildLayout() {
	// Top bar
	if a.config.TopBar != nil {
		a.topBar = a.config.TopBar
		a.main.AddItem(a.topBar, a.config.TopBarHeight, 0, false)
	}

	// Crumbs
	if a.config.ShowCrumbs {
		a.crumbs = nav.NewCrumbs()
		a.main.AddItem(a.crumbs, 1, 0, false)
		// Connect crumbs to pages for automatic updates
		a.pages.SetCrumbs(a.crumbs)
	}

	// Pages (main content area)
	a.main.AddItem(a.pages, 0, 1, true)

	// Bottom bar (menu)
	if a.config.BottomBar != nil {
		if menu, ok := a.config.BottomBar.(*Menu); ok {
			a.menu = menu
		}
		a.main.AddItem(a.config.BottomBar, a.config.BottomBarHeight, 0, false)
	}

	a.app.SetRoot(a.main, true)
}

// Run starts the application event loop.
func (a *App) Run() error {
	return a.app.Run()
}

// Stop stops the application and shuts down the Effect dispatcher,
// cancelling any in-flight Effects.
func (a *App) Stop() {
	a.subs.Release()
	if a.menu != nil {
		a.menu.Subs().Release()
	}
	if a.pages != nil {
		a.pages.Subs().Release()
	}
	if a.crumbs != nil {
		a.crumbs.Subs().Release()
	}
	if a.effects != nil {
		timeout := a.config.EffectShutdownTimeout
		switch {
		case timeout < 0:
			go a.effects.Shutdown(context.Background())
		default:
			if timeout == 0 {
				timeout = defaultEffectShutdownTimeout
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			_ = a.effects.Shutdown(ctx)
			cancel()
		}
	}
	a.app.Stop()
}

// Effects returns the application's default Effect dispatcher.
// Effects are an opt-in command/effect layer; see package effect.
func (a *App) Effects() *effect.Dispatcher {
	return a.effects
}

// Pages returns the page manager.
func (a *App) Pages() *nav.Pages {
	return a.pages
}

// Crumbs returns the breadcrumb component (nil if disabled).
func (a *App) Crumbs() *nav.Crumbs {
	return a.crumbs
}

// Menu returns the menu component (nil if no Menu in BottomBar).
func (a *App) Menu() *Menu {
	return a.menu
}

// TopBar returns the top bar primitive.
func (a *App) TopBar() tview.Primitive {
	return a.topBar
}

// SetTopBar replaces the top bar.
func (a *App) SetTopBar(bar tview.Primitive) *App {
	// Remove old top bar
	if a.topBar != nil {
		a.main.RemoveItem(a.topBar)
	}

	a.topBar = bar
	a.config.TopBar = bar

	// Rebuild layout
	a.main.Clear()
	a.buildLayout()

	return a
}

// SetBottomBar replaces the bottom bar.
func (a *App) SetBottomBar(bar tview.Primitive) *App {
	// Remove old bottom bar
	if a.config.BottomBar != nil {
		a.main.RemoveItem(a.config.BottomBar)
	}

	a.config.BottomBar = bar
	if menu, ok := bar.(*Menu); ok {
		a.menu = menu
	} else {
		a.menu = nil
	}

	// Rebuild layout
	a.main.Clear()
	a.buildLayout()

	return a
}

// GetApplication returns the underlying tview.Application.
func (a *App) GetApplication() *tview.Application {
	return a.app
}

// SetFocus sets focus to a specific primitive.
func (a *App) SetFocus(p tview.Primitive) *App {
	a.app.SetFocus(p)
	return a
}

// QueueUpdate queues a function to run on the main thread.
func (a *App) QueueUpdate(fn func()) *App {
	a.app.QueueUpdate(fn)
	return a
}

// QueueUpdateDraw queues a function to run and then redraws.
func (a *App) QueueUpdateDraw(fn func()) *App {
	a.app.QueueUpdateDraw(fn)
	return a
}

// Draw triggers a redraw.
func (a *App) Draw() *App {
	a.app.Draw()
	return a
}

// SetInputCapture sets a function to capture input before it reaches focused primitive.
// Note: Modal auto-dismiss handling runs before this capture function.
func (a *App) SetInputCapture(capture func(*tcell.EventKey) *tcell.EventKey) *App {
	a.userInputCapture = capture
	return a
}

// setupModalInputCapture configures automatic modal input handling.
func (a *App) setupModalInputCapture() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Debug overlay toggle, when enabled.
		if a.config.Debug && event.Key() == a.config.DebugKey {
			go func() {
				a.app.QueueUpdateDraw(func() {
					overlay := components.NewDebugOverlay(0)
					overlay.SetOnClose(func() { a.pages.Pop() })
					a.pages.Push(overlay)
				})
			}()
			return nil
		}

		// Check if current page is a modal with auto-dismiss
		if behavior := a.pages.CurrentModalBehavior(); behavior != nil {
			// Handle auto-dismiss on Escape
			if behavior.DismissOnEsc && event.Key() == tcell.KeyEscape {
				// Use a goroutine to completely defer the dismiss operation
				// outside of tview's event handling to avoid deadlocks.
				go func() {
					a.app.QueueUpdateDraw(func() {
						a.pages.DismissModal()
					})
				}()
				return nil // Event consumed
			}
		}

		// Pass to user's custom capture if set
		if a.userInputCapture != nil {
			return a.userInputCapture(event)
		}

		return event
	})
}

// UpdateMenuHints updates the menu with hints from a component.
func (a *App) UpdateMenuHints(hints []components.KeyHint) {
	if a.menu != nil {
		a.menu.SetHints(hints)
	}
}

// ShowModal displays a modal over the current content.
func (a *App) ShowModal(modal *components.Modal) {
	a.pages.Push(&modalWrapper{modal: modal, app: a})
}

// RefreshTheme forces a theme refresh and redraw.
//
// Note: As of v0.0.6, calling SetProvider() automatically triggers a redraw,
// so explicit RefreshTheme() calls are typically unnecessary. This method
// remains available for cases where you need to force a refresh without
// changing the theme, or when auto-refresh is disabled via SetAutoRefresh(false).
func (a *App) RefreshTheme() {
	a.main.SetBackgroundColor(theme.Bg())
	a.app.QueueUpdateDraw(func() {})
}

// Suspend temporarily suspends the application, executes the given function,
// and then resumes the application. This is useful for running external commands
// that need direct terminal access.
func (a *App) Suspend(fn func()) bool {
	return a.app.Suspend(fn)
}

// modalWrapper wraps a Modal to implement nav.Component and nav.ModalComponent.
type modalWrapper struct {
	modal *components.Modal
	app   *App
}

func (m *modalWrapper) Name() string { return "Modal" }

func (m *modalWrapper) Start() {
	// Set up close handler to pop
	m.modal.SetOnClose(func() {
		m.app.pages.Pop()
	})
}

func (m *modalWrapper) Stop() {}

func (m *modalWrapper) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "Esc", Description: "Close"},
	}
}

// ModalBehavior implements nav.Modal.
func (m *modalWrapper) ModalBehavior() components.ModalBehavior {
	return components.DefaultModalBehavior()
}

// OnDismiss implements nav.Modal.
func (m *modalWrapper) OnDismiss() bool {
	return true // Always allow dismiss
}

func (m *modalWrapper) Draw(screen tcell.Screen)       { m.modal.Draw(screen) }
func (m *modalWrapper) GetRect() (int, int, int, int)  { return m.modal.GetRect() }
func (m *modalWrapper) SetRect(x, y, w, h int)         { m.modal.SetRect(x, y, w, h) }
func (m *modalWrapper) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return m.modal.InputHandler()
}
func (m *modalWrapper) Focus(delegate func(tview.Primitive)) { m.modal.Focus(delegate) }
func (m *modalWrapper) Blur()                                { m.modal.Blur() }
func (m *modalWrapper) HasFocus() bool                       { return m.modal.HasFocus() }
func (m *modalWrapper) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return m.modal.MouseHandler()
}
func (m *modalWrapper) PasteHandler() func(string, func(tview.Primitive)) {
	return m.modal.PasteHandler()
}
