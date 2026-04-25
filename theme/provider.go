package theme

import (
	"sync"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/bus"
)

// Theme defines the color contract all themes must implement.
// Components read these values at draw time for live theme switching.
type Theme interface {
	// Base colors
	Bg() tcell.Color
	BgLight() tcell.Color
	BgDark() tcell.Color
	Fg() tcell.Color
	FgDim() tcell.Color
	FgMuted() tcell.Color

	// Accent colors
	Accent() tcell.Color
	AccentDim() tcell.Color
	Highlight() tcell.Color

	// Semantic colors
	Success() tcell.Color
	Warning() tcell.Color
	Error() tcell.Color
	Info() tcell.Color

	// Border colors
	Border() tcell.Color
	BorderFocus() tcell.Color

	// UI element colors
	Header() tcell.Color
	Menu() tcell.Color
	TableHeader() tcell.Color
	Key() tcell.Color
	Crumb() tcell.Color
	PanelBorder() tcell.Color
	PanelTitle() tcell.Color
}

// Primitive is any tview primitive that can have its background color set.
type Primitive interface {
	SetBackgroundColor(tcell.Color) *tview.Box
}

// Refreshable is an optional interface for components that need custom
// refresh logic when the theme changes. Components implementing this
// interface will have RefreshTheme() called automatically after SetProvider().
//
// Use this for components that cache colors or need to update state beyond
// just background colors. For most components, the automatic background
// update via Register() is sufficient.
type Refreshable interface {
	RefreshTheme()
}

// themeHolder wraps Theme interface for atomic.Value storage.
// atomic.Value requires consistent concrete types - this wrapper ensures
// we always store *themeHolder regardless of the underlying Theme implementation.
type themeHolder struct {
	theme Theme
}

var (
	// Use atomic.Value for lock-free theme reads to prevent deadlocks
	// during Draw() cycles when SetProvider() is called from callbacks
	activeTheme atomic.Value // stores *themeHolder

	appInstance *tview.Application
	appMu       sync.RWMutex

	// Registry of primitives that need background updates on theme change
	registeredPrimitives []Primitive
	primitivesMu         sync.RWMutex

	// Registry of refreshable components for custom theme refresh logic
	registeredRefreshables []Refreshable
	refreshablesMu         sync.RWMutex

	// Callbacks to notify when theme changes
	themeChangeCallbacks []func()
	callbacksMu          sync.RWMutex

	// autoRefresh controls whether SetProvider automatically triggers a redraw.
	// Default is true for ergonomic theme switching.
	autoRefresh     = true
	autoRefreshOnce sync.Once
)

// SetProvider sets the active theme provider and updates tview global styles.
// Also updates all registered primitives' background colors, calls RefreshTheme()
// on registered Refreshable components, notifies callbacks, and automatically
// triggers a redraw (unless disabled via SetAutoRefresh(false)).
//
// This function is safe to call from any goroutine including during Draw() cycles.
func SetProvider(t Theme) {
	// Capture prior theme for bus instrumentation before atomic swap.
	var prev Theme
	if h := activeTheme.Load(); h != nil {
		prev = h.(*themeHolder).theme
	}

	// Store theme atomically - lock-free for readers
	// Wrap in themeHolder to ensure consistent concrete type for atomic.Value
	activeTheme.Store(&themeHolder{theme: t})

	// Update tview global styles for components using tcell.ColorDefault
	applyGlobalStyles(t)

	// Notify theme change callbacks (sync - these should be lightweight)
	notifyThemeChange()

	if bus.Enabled() {
		bus.Publish(bus.Event{
			Kind:    bus.KindThemeSwitch,
			Source:  bus.SourceTheme,
			Payload: bus.ThemeSwitch{From: themeName(prev), To: themeName(t)},
		})
	}

	// Auto-trigger redraw if enabled and app is registered
	if autoRefresh {
		// Use a goroutine to avoid any potential deadlock when called from
		// within tview callbacks. The QueueUpdateDraw will safely execute
		// on the main UI thread.
		go func() {
			QueueUpdateDraw(func() {
				// Update all registered primitives' backgrounds
				updateRegisteredPrimitives(t)

				// Call RefreshTheme() on all registered Refreshable components
				refreshRegisteredComponents()
			})
		}()
	} else {
		// If auto-refresh is disabled, update primitives synchronously
		// but skip the RefreshTheme calls (user handles manually)
		updateRegisteredPrimitives(t)
	}
}

// SetAutoRefresh controls whether SetProvider automatically triggers a redraw.
// Default is true. Set to false if you want to batch theme changes or handle
// redraws manually.
func SetAutoRefresh(enabled bool) {
	autoRefresh = enabled
}

// themeName returns a stable identifier for a theme — its concrete Go type.
// Returns an empty string for nil. Used only for bus event payloads.
func themeName(t Theme) string {
	if t == nil {
		return ""
	}
	return reflectTypeName(t)
}

// applyGlobalStyles updates tview.Styles from the theme.
func applyGlobalStyles(t Theme) {
	if t == nil {
		return
	}
	tview.Styles.PrimitiveBackgroundColor = t.Bg()
	tview.Styles.ContrastBackgroundColor = t.BgLight()
	tview.Styles.MoreContrastBackgroundColor = t.BgDark()
	tview.Styles.BorderColor = t.Border()
	tview.Styles.TitleColor = t.Accent()
	tview.Styles.PrimaryTextColor = t.Fg()
	tview.Styles.SecondaryTextColor = t.FgDim()
}

// updateRegisteredPrimitives updates backgrounds of all registered primitives.
func updateRegisteredPrimitives(t Theme) {
	if t == nil {
		return
	}

	primitivesMu.RLock()
	defer primitivesMu.RUnlock()

	bg := t.Bg()
	for _, p := range registeredPrimitives {
		p.SetBackgroundColor(bg)
	}
}

// refreshRegisteredComponents calls RefreshTheme() on all registered Refreshable components.
func refreshRegisteredComponents() {
	refreshablesMu.RLock()
	defer refreshablesMu.RUnlock()

	for _, r := range registeredRefreshables {
		r.RefreshTheme()
	}
}

// triggerRedraw queues a redraw if an app is registered.
func triggerRedraw() {
	appMu.RLock()
	app := appInstance
	appMu.RUnlock()

	if app != nil {
		// Use QueueUpdateDraw to safely trigger redraw from any goroutine
		app.QueueUpdateDraw(func() {})
	}
}

// Register adds a primitive to receive automatic background updates on theme change.
// Call this when creating components that contain tview primitives.
// The primitive will have SetBackgroundColor called whenever SetProvider is called.
//
// Returns an unregister function. Call it (or pass it to a Subscriptions set)
// when the primitive's owner is torn down to prevent leaks.
func Register(p Primitive) func() {
	primitivesMu.Lock()
	registeredPrimitives = append(registeredPrimitives, p)
	primitivesMu.Unlock()
	return func() { Unregister(p) }
}

// Unregister removes a primitive from automatic background updates.
// Call this when a component is destroyed to prevent memory leaks.
func Unregister(p Primitive) {
	primitivesMu.Lock()
	defer primitivesMu.Unlock()

	for i, registered := range registeredPrimitives {
		if registered == p {
			registeredPrimitives = append(registeredPrimitives[:i], registeredPrimitives[i+1:]...)
			return
		}
	}
}

// RegisterRefreshable adds a component to receive RefreshTheme() calls on theme change.
// Use this for components that need custom refresh logic beyond background color updates.
// Components are called in registration order after primitives are updated.
//
// Returns an unregister function for convenience:
//
//	unregister := theme.RegisterRefreshable(myComponent)
//	defer unregister() // Clean up when component is destroyed
func RegisterRefreshable(r Refreshable) func() {
	refreshablesMu.Lock()
	defer refreshablesMu.Unlock()
	registeredRefreshables = append(registeredRefreshables, r)

	// Return unregister function
	return func() {
		UnregisterRefreshable(r)
	}
}

// UnregisterRefreshable removes a component from RefreshTheme() notifications.
// Call this when a component is destroyed to prevent memory leaks.
func UnregisterRefreshable(r Refreshable) {
	refreshablesMu.Lock()
	defer refreshablesMu.Unlock()

	for i, registered := range registeredRefreshables {
		if registered == r {
			registeredRefreshables = append(registeredRefreshables[:i], registeredRefreshables[i+1:]...)
			return
		}
	}
}

// Get returns the current theme (thread-safe, lock-free).
// Returns nil if no theme has been set.
func Get() Theme {
	if holder := activeTheme.Load(); holder != nil {
		return holder.(*themeHolder).theme
	}
	return nil
}

// SetApp registers the tview application for queue operations.
// Call this after creating your tview.Application.
func SetApp(app *tview.Application) {
	appMu.Lock()
	defer appMu.Unlock()
	appInstance = app
}

// GetApp returns the registered application.
func GetApp() *tview.Application {
	appMu.RLock()
	defer appMu.RUnlock()
	return appInstance
}

// QueueUpdate runs fn on main UI thread.
// Falls back to immediate execution if no app instance.
func QueueUpdate(fn func()) {
	appMu.RLock()
	app := appInstance
	appMu.RUnlock()

	if app != nil {
		app.QueueUpdate(fn)
	} else {
		fn()
	}
}

// QueueUpdateDraw runs fn and triggers redraw.
// Falls back to immediate execution if no app instance.
func QueueUpdateDraw(fn func()) {
	appMu.RLock()
	app := appInstance
	appMu.RUnlock()

	if app != nil {
		app.QueueUpdateDraw(fn)
	} else {
		fn()
	}
}

// OnThemeChange registers a callback to be called when the theme changes.
// Returns an unregister function to remove the callback.
func OnThemeChange(fn func()) func() {
	callbacksMu.Lock()
	defer callbacksMu.Unlock()
	themeChangeCallbacks = append(themeChangeCallbacks, fn)

	// Return unregister function
	return func() {
		callbacksMu.Lock()
		defer callbacksMu.Unlock()
		for i, cb := range themeChangeCallbacks {
			// Compare function pointers (this works for the same function reference)
			if &cb == &fn {
				themeChangeCallbacks = append(themeChangeCallbacks[:i], themeChangeCallbacks[i+1:]...)
				return
			}
		}
	}
}

// notifyThemeChange calls all registered theme change callbacks.
func notifyThemeChange() {
	callbacksMu.RLock()
	callbacks := make([]func(), len(themeChangeCallbacks))
	copy(callbacks, themeChangeCallbacks)
	callbacksMu.RUnlock()

	for _, fn := range callbacks {
		fn()
	}
}
