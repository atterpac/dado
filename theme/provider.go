package theme

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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

var (
	activeTheme Theme
	themeMu     sync.RWMutex

	appInstance *tview.Application
	appMu       sync.RWMutex

	// Registry of primitives that need background updates on theme change
	registeredPrimitives []Primitive
	primitivesMu         sync.RWMutex
)

// SetProvider sets the active theme provider and updates tview global styles.
// Also updates all registered primitives' background colors.
func SetProvider(t Theme) {
	themeMu.Lock()
	activeTheme = t
	themeMu.Unlock()

	// Update tview global styles for components using tcell.ColorDefault
	applyGlobalStyles(t)

	// Update all registered primitives' backgrounds
	updateRegisteredPrimitives(t)
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

// Register adds a primitive to receive automatic background updates on theme change.
// Call this when creating components that contain tview primitives.
// The primitive will have SetBackgroundColor called whenever SetProvider is called.
func Register(p Primitive) {
	primitivesMu.Lock()
	defer primitivesMu.Unlock()
	registeredPrimitives = append(registeredPrimitives, p)
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

// Get returns the current theme (thread-safe).
// Returns nil if no theme has been set.
func Get() Theme {
	themeMu.RLock()
	defer themeMu.RUnlock()
	return activeTheme
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
