package layout

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/effect"
	"github.com/atterpac/dado/nav"
	"github.com/atterpac/dado/theme"
)

// AppConfig configures the application layout.
type AppConfig struct {
	// TopBar is shown at the top (e.g., status bar). Can be nil.
	TopBar core.Widget

	// ShowCrumbs enables the breadcrumb bar below TopBar.
	ShowCrumbs bool

	// BottomBar is shown at the bottom (e.g., Menu). Can be nil.
	BottomBar core.Widget

	// TopBarHeight is the height of the top bar (default: 3).
	TopBarHeight int

	// BottomBarHeight is the height of the bottom bar (default: 1).
	BottomBarHeight int

	// OnComponentChange is called when the active component changes.
	// Useful for updating menu hints, crumbs, etc.
	OnComponentChange func(nav.Component)

	// Debug enables the floating debug toolbar. When true, pressing DebugKey
	// toggles a DebugToolbar drawn on top of the live app (event log, widget-tree
	// inspector, cell probe). Tab cycles tools; Esc closes.
	Debug bool

	// DebugKey is the key that toggles the debug toolbar. Defaults to
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
	app              *core.App
	main             *core.Flex
	topBar           core.Widget
	crumbs           *nav.Crumbs
	pages            *nav.Pages
	menu             *Menu
	config           AppConfig
	userInputCapture func(*tcell.EventKey) *tcell.EventKey
	effects          *effect.Dispatcher
	subs             components.Subscriptions
	themeState       *themeState
	debug            *components.DebugToolbar
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

	a := &App{
		app:     core.NewApp(),
		main:    core.NewFlex().SetDirection(core.Column),
		pages:   nav.NewPages(),
		config:  config,
		effects: effect.NewDispatcher(),
	}

	// Wire theme queue to core.App so theme changes trigger redraws.
	theme.SetQueue(func(fn func()) { a.app.QueueUpdateDraw(fn) })

	// Wire theme.Bg as the global default background for all core primitives
	// (Flex, TextView, Table, etc.) so they track theme changes automatically.
	core.SetDefaultBackgroundFunc(theme.Bg)

	// Wire focus manager into pages for modal focus save/restore.
	a.pages.SetFocusManager(a.app.Focus())

	// Build layout
	a.buildLayout()

	// Debug toolbar: a float drawn on top of the app via the after-draw hook,
	// summoned by DebugKey. Created before input capture so it can intercept keys.
	if a.config.Debug {
		a.debug = components.NewDebugToolbar(a.app)
		a.app.SetAfterDrawFunc(a.debug.Draw)
	}

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
		a.pages.SetCrumbs(a.crumbs)
	}

	// Pages (main content area) — implements core.Widget directly
	a.main.AddItem(a.pages, 0, 1, true)

	// Bottom bar
	if a.config.BottomBar != nil {
		if menu, ok := a.config.BottomBar.(*Menu); ok {
			a.menu = menu
		}
		a.main.AddItem(a.config.BottomBar, a.config.BottomBarHeight, 0, false)
	}

	a.app.SetRoot(a.main)

	// Give focus to pages (the main content area).
	a.app.SetFocus(a.pages)
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

// TopBar returns the top bar widget.
func (a *App) TopBar() core.Widget {
	return a.topBar
}

// SetTopBar replaces the top bar.
func (a *App) SetTopBar(bar core.Widget) *App {
	a.topBar = bar
	a.config.TopBar = bar
	a.main.Clear()
	a.buildLayout()
	return a
}

// SetBottomBar replaces the bottom bar.
func (a *App) SetBottomBar(bar core.Widget) *App {
	a.config.BottomBar = bar
	if menu, ok := bar.(*Menu); ok {
		a.menu = menu
	} else {
		a.menu = nil
	}
	a.main.Clear()
	a.buildLayout()
	return a
}

// GetApp returns the underlying core.App.
func (a *App) GetApp() *core.App {
	return a.app
}

// SetFocus sets focus to a specific widget.
func (a *App) SetFocus(w core.Widget) *App {
	a.app.SetFocus(w)
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
		// Debug toolbar, when enabled. The toolbar is a float drawn over the
		// live app (not a page), driven by keys routed here.
		if a.debug != nil {
			if event.Key() == a.config.DebugKey {
				a.debug.Toggle()
				return nil
			}
			if a.debug.Visible() {
				if a.debug.HandleKey(event) {
					return nil // consumed by toolbar / panel tool
				}
				return event // inline tool passthrough — app still receives it
			}
		}

		// Theme selector toggle, when enabled via EnableThemes.
		if a.themeState != nil && event.Key() == a.themeState.key {
			go func() {
				a.app.QueueUpdateDraw(a.openThemeSelector)
			}()
			return nil
		}

		// Check if current page is a modal with auto-dismiss
		if behavior := a.pages.CurrentModalBehavior(); behavior != nil {
			// Handle auto-dismiss on Escape
			if behavior.DismissOnEsc && event.Key() == tcell.KeyEscape {
				// Use a goroutine to completely defer the dismiss operation
				// outside of the event-handling goroutine to avoid deadlocks.
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

func (m *modalWrapper) Draw(screen tcell.Screen) { m.modal.Draw(screen) }
func (m *modalWrapper) SetRect(x, y, w, h int)   { m.modal.SetRect(x, y, w, h) }
func (m *modalWrapper) Blur()                    { m.modal.Blur() }
func (m *modalWrapper) HasFocus() bool           { return m.modal.HasFocus() }
