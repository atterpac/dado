package theme

import (
	"sync"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/bus"
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

// Refreshable is an optional interface for components that need custom
// refresh logic when the theme changes.
type Refreshable interface {
	RefreshTheme()
}

// themeHolder wraps Theme for atomic.Value storage so the concrete stored
// type stays consistent across different Theme implementations.
type themeHolder struct {
	theme Theme
}

// Provider owns a theme and the observers attached to it: primitives that
// receive automatic background updates, refreshables, and change callbacks.
//
// The package-level Default() provider backs the convenience functions
// (theme.Bg(), theme.Register(), theme.SetProvider(), etc.) — most apps
// use Default() implicitly. Construct a separate *Provider to scope a
// different theme to a subtree (preview pane, modal, test fixture) or to
// embed the library inside another themed app without colliding on the
// global.
type Provider struct {
	active atomic.Value // *themeHolder, lock-free reads for Draw paths

	primitivesMu sync.RWMutex
	primitives   []func(tcell.Color)

	refreshablesMu sync.RWMutex
	refreshables   []Refreshable

	callbacksMu sync.RWMutex
	callbacks   []callbackEntry
	callbackSeq uint64

	autoRefreshMu sync.RWMutex
	autoRefresh   bool
}

// NewProvider creates an empty Provider with auto-refresh enabled.
// Pass the result to ComponentBase.SetTheme (or equivalent) to scope it.
func NewProvider() *Provider {
	return &Provider{autoRefresh: true}
}

var defaultProvider = NewProvider()

// Default returns the package-level Provider used by the convenience
// functions. Mutating this provider affects all components that have not
// opted into a different one.
func Default() *Provider { return defaultProvider }

// --- Provider methods ---

// SetTheme sets the active theme on this provider, updates registered
// primitive backgrounds, calls Refreshable.RefreshTheme on registered
// components, and fires change callbacks.
func (p *Provider) SetTheme(t Theme) {
	var prev Theme
	if h := p.active.Load(); h != nil {
		prev = h.(*themeHolder).theme
	}
	p.active.Store(&themeHolder{theme: t})

	p.notifyThemeChange()

	if bus.Enabled() {
		bus.Publish(bus.Event{
			Kind:    bus.KindThemeSwitch,
			Source:  bus.SourceTheme,
			Payload: bus.ThemeSwitch{From: themeName(prev), To: themeName(t)},
		})
	}

	p.autoRefreshMu.RLock()
	auto := p.autoRefresh
	p.autoRefreshMu.RUnlock()

	if auto {
		// Off the caller goroutine so callers inside draw callbacks
		// don't deadlock against QueueUpdateDraw.
		go func() {
			QueueUpdateDraw(func() {
				p.updatePrimitives(t)
				p.refreshComponents()
			})
		}()
	} else {
		p.updatePrimitives(t)
	}
}

// Theme returns the active theme (lock-free). Returns nil if unset.
func (p *Provider) Theme() Theme {
	if h := p.active.Load(); h != nil {
		return h.(*themeHolder).theme
	}
	return nil
}

// SetAutoRefresh toggles automatic redraw on SetTheme. Default true.
func (p *Provider) SetAutoRefresh(enabled bool) {
	p.autoRefreshMu.Lock()
	p.autoRefresh = enabled
	p.autoRefreshMu.Unlock()
}

// RegisterFn registers a background-setter func for automatic updates on theme change.
// Returns an unregister function.
func (p *Provider) RegisterFn(fn func(tcell.Color)) func() {
	p.primitivesMu.Lock()
	idx := len(p.primitives)
	p.primitives = append(p.primitives, fn)
	p.primitivesMu.Unlock()
	return func() {
		p.primitivesMu.Lock()
		defer p.primitivesMu.Unlock()
		if idx < len(p.primitives) {
			p.primitives[idx] = nil
		}
	}
}

// RegisterRefreshable adds a component for RefreshTheme() on theme change.
// Returns an unregister function.
func (p *Provider) RegisterRefreshable(r Refreshable) func() {
	p.refreshablesMu.Lock()
	p.refreshables = append(p.refreshables, r)
	p.refreshablesMu.Unlock()
	return func() { p.UnregisterRefreshable(r) }
}

// UnregisterRefreshable removes a refreshable.
func (p *Provider) UnregisterRefreshable(r Refreshable) {
	p.refreshablesMu.Lock()
	defer p.refreshablesMu.Unlock()
	for i, reg := range p.refreshables {
		if reg == r {
			p.refreshables = append(p.refreshables[:i], p.refreshables[i+1:]...)
			return
		}
	}
}

type callbackEntry struct {
	id uint64
	fn func()
}

// OnThemeChange registers a callback fired on every SetTheme call.
// Returns an unregister function.
func (p *Provider) OnThemeChange(fn func()) func() {
	p.callbacksMu.Lock()
	p.callbackSeq++
	id := p.callbackSeq
	p.callbacks = append(p.callbacks, callbackEntry{id: id, fn: fn})
	p.callbacksMu.Unlock()
	return func() {
		p.callbacksMu.Lock()
		defer p.callbacksMu.Unlock()
		for i, c := range p.callbacks {
			if c.id == id {
				p.callbacks = append(p.callbacks[:i], p.callbacks[i+1:]...)
				return
			}
		}
	}
}

func (p *Provider) updatePrimitives(t Theme) {
	if t == nil {
		return
	}
	p.primitivesMu.RLock()
	defer p.primitivesMu.RUnlock()
	bg := t.Bg()
	for _, fn := range p.primitives {
		if fn != nil {
			fn(bg)
		}
	}
}

func (p *Provider) refreshComponents() {
	p.refreshablesMu.RLock()
	defer p.refreshablesMu.RUnlock()
	for _, r := range p.refreshables {
		r.RefreshTheme()
	}
}

func (p *Provider) notifyThemeChange() {
	p.callbacksMu.RLock()
	cbs := make([]func(), 0, len(p.callbacks))
	for _, c := range p.callbacks {
		cbs = append(cbs, c.fn)
	}
	p.callbacksMu.RUnlock()
	for _, fn := range cbs {
		fn()
	}
}

// --- Package-level forwarders (default provider) ---

// SetProvider sets the active theme on Default(). Kept for back-compat;
// prefer Default().SetTheme(t) or a scoped Provider.
func SetProvider(t Theme) { defaultProvider.SetTheme(t) }

// SetAutoRefresh forwards to Default().
func SetAutoRefresh(enabled bool) { defaultProvider.SetAutoRefresh(enabled) }

// Get returns the active theme on Default().
func Get() Theme { return defaultProvider.Theme() }

// RegisterFn registers a background-setter func for automatic background updates on theme change.
func RegisterFn(fn func(tcell.Color)) func() { return defaultProvider.RegisterFn(fn) }

// RegisterRefreshable forwards to Default().
func RegisterRefreshable(r Refreshable) func() { return defaultProvider.RegisterRefreshable(r) }

// UnregisterRefreshable forwards to Default().
func UnregisterRefreshable(r Refreshable) { defaultProvider.UnregisterRefreshable(r) }

// OnThemeChange forwards to Default().
func OnThemeChange(fn func()) func() { return defaultProvider.OnThemeChange(fn) }

// --- App registration (singleton, not provider-scoped) ---

var (
	queueFn func(func())
	queueMu sync.RWMutex
)

// SetQueue registers a function used by QueueUpdate and QueueUpdateDraw.
// Pass nil to clear and fall back to direct execution.
func SetQueue(fn func(func())) {
	queueMu.Lock()
	queueFn = fn
	queueMu.Unlock()
}

// QueueUpdate runs fn on the main UI thread. Falls back to immediate
// execution when no queue is registered (test mode).
func QueueUpdate(fn func()) {
	queueMu.RLock()
	q := queueFn
	queueMu.RUnlock()
	if q != nil {
		q(fn)
		return
	}
	fn()
}

// QueueUpdateDraw runs fn and triggers a redraw. Falls back to immediate
// execution when no queue is registered.
func QueueUpdateDraw(fn func()) {
	queueMu.RLock()
	q := queueFn
	queueMu.RUnlock()
	if q != nil {
		q(fn)
		return
	}
	fn()
}

// --- Internal helpers ---

func themeName(t Theme) string {
	if t == nil {
		return ""
	}
	return reflectTypeName(t)
}
