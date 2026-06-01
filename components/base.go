package components

import (
	"sync"
	"sync/atomic"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

var componentIDCounter uint64

func nextComponentID() uint64 {
	return atomic.AddUint64(&componentIDCounter, 1)
}

// ComponentBase wraps a core.Widget and provides nav.Component implementation.
// Use as a field (composition), not embedded, for type-safe access to the underlying primitive.
//
// Example:
//
//	type MyView struct {
//	    base  *ComponentBase
//	    table *Table
//	}
//
//	func NewMyView() *MyView {
//	    table := NewTable()
//	    v := &MyView{table: table}
//	    v.base = NewComponentBase(table).
//	        SetName("my-view").
//	        SetHints([]KeyHint{{Key: "Enter", Description: "Select"}}).
//	        SetOnStart(v.loadData)
//	    return v
//	}
type ComponentBase struct {
	mu        sync.RWMutex
	primitive core.Widget
	name      string
	id        uint64
	hints     []KeyHint
	onStart   func()
	onStop    func()
	subs      Subscriptions

	// Optional overrides
	inputHandler func(*tcell.EventKey) bool
	drawOverlay  func(screen tcell.Screen)
	themeP       *theme.Provider
}

// Subscriptions aggregates unregister functions so component teardown
// releases all observer/refresh hooks registered during the component's lifetime.
//
// Zero value is ready to use. Methods are safe for concurrent use.
type Subscriptions struct {
	mu    sync.Mutex
	funcs []func()
}

// Add stores an unregister function. Nil is ignored.
func (s *Subscriptions) Add(unsub func()) {
	if unsub == nil {
		return
	}
	s.mu.Lock()
	s.funcs = append(s.funcs, unsub)
	s.mu.Unlock()
}

// Release invokes every registered unregister function in LIFO order and clears the list.
// Safe to call multiple times.
func (s *Subscriptions) Release() {
	s.mu.Lock()
	fns := s.funcs
	s.funcs = nil
	s.mu.Unlock()
	for i := len(fns) - 1; i >= 0; i-- {
		fns[i]()
	}
}

// Len returns the number of pending unregister functions. Intended for tests.
func (s *Subscriptions) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.funcs)
}

// NewComponentBase creates a new component base wrapping the given widget.
func NewComponentBase(p core.Widget) *ComponentBase {
	return &ComponentBase{
		primitive: p,
		id:        nextComponentID(),
		hints:     make([]KeyHint, 0),
	}
}

// --- Configuration Methods (Fluent API) ---

// SetName sets the component name (used for debugging and events).
func (cb *ComponentBase) SetName(name string) *ComponentBase {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.name = name
	return cb
}

// Name returns the component name.
func (cb *ComponentBase) Name() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.name
}

// ID returns the unique component ID.
func (cb *ComponentBase) ID() uint64 {
	return cb.id
}

// SetHints sets the key hints displayed in the menu bar.
func (cb *ComponentBase) SetHints(hints []KeyHint) *ComponentBase {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.hints = hints
	return cb
}

// AddHint adds a single key hint.
func (cb *ComponentBase) AddHint(key, description string) *ComponentBase {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.hints = append(cb.hints, KeyHint{Key: key, Description: description})
	return cb
}

// SetOnStart sets the callback invoked when component becomes active.
func (cb *ComponentBase) SetOnStart(fn func()) *ComponentBase {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStart = fn
	return cb
}

// SetOnStop sets the callback invoked when component becomes inactive.
func (cb *ComponentBase) SetOnStop(fn func()) *ComponentBase {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStop = fn
	return cb
}

// SetInputHandler sets a custom input handler.
// Return true to indicate the event was consumed; false to delegate to the wrapped primitive.
func (cb *ComponentBase) SetInputHandler(fn func(*tcell.EventKey) bool) *ComponentBase {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.inputHandler = fn
	return cb
}

// SetDrawOverlay sets a function called after the primitive draws.
func (cb *ComponentBase) SetDrawOverlay(fn func(screen tcell.Screen)) *ComponentBase {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.drawOverlay = fn
	return cb
}

// Primitive returns the underlying core.Widget.
func (cb *ComponentBase) Primitive() core.Widget {
	return cb.primitive
}

// Typed returns the wrapped widget cast to P. Panics if the stored
// widget is not a P — the cast mirrors the type the caller used at
// NewComponentBase, so a mismatch is a programmer error, not a runtime
// condition. Use when a caller needs typed access without going through
// a component-specific accessor.
//
//	tbl := components.Typed[*Table](cb)
//	tbl.SetCell(0, 0, ...)
func Typed[P core.Widget](cb *ComponentBase) P {
	return cb.primitive.(P)
}

// Subs returns the component's subscription set. Register theme/binding
// unsubscribers here; they fire automatically on Stop().
func (cb *ComponentBase) Subs() *Subscriptions {
	return &cb.subs
}

// SetTheme scopes a theme.Provider to this component. When set, code that
// reads via cb.Theme() uses this provider instead of theme.Default().
// Pass nil to clear the override and fall back to the package default.
func (cb *ComponentBase) SetTheme(p *theme.Provider) *ComponentBase {
	cb.mu.Lock()
	cb.themeP = p
	cb.mu.Unlock()
	return cb
}

// Theme returns the scoped Provider if SetTheme was called, otherwise
// theme.Default(). New code reads colors via cb.Theme().Bg() to honor
// per-subtree theme overrides; legacy theme.Bg() always reads Default().
func (cb *ComponentBase) Theme() *theme.Provider {
	cb.mu.RLock()
	p := cb.themeP
	cb.mu.RUnlock()
	if p != nil {
		return p
	}
	return theme.Default()
}

// --- nav.Component Implementation ---

// Start is called when the component becomes the active view.
func (cb *ComponentBase) Start() {
	cb.mu.RLock()
	fn := cb.onStart
	cb.mu.RUnlock()
	if fn != nil {
		fn()
	}
}

// Stop is called when the component is no longer the active view.
// Releases all registered Subscriptions after the user-provided onStop
// runs. If the wrapped widget exposes a Subs() *Subscriptions
// accessor (the convention used by dado's leaf widgets), its
// subscriptions are released as well so theme/binding hooks attached
// inside the widget's own constructor are torn down with the component.
func (cb *ComponentBase) Stop() {
	cb.mu.RLock()
	fn := cb.onStop
	cb.mu.RUnlock()
	if fn != nil {
		fn()
	}
	if w, ok := cb.primitive.(interface{ Subs() *Subscriptions }); ok {
		if s := w.Subs(); s != nil {
			s.Release()
		}
	}
	cb.subs.Release()
}

// Hints returns the key hints for this component.
func (cb *ComponentBase) Hints() []KeyHint {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	result := make([]KeyHint, len(cb.hints))
	copy(result, cb.hints)
	return result
}

// --- core.Widget Implementation (Delegation) ---

// Draw delegates to the wrapped widget.
func (cb *ComponentBase) Draw(screen tcell.Screen) {
	cb.primitive.Draw(screen)
	cb.mu.RLock()
	overlay := cb.drawOverlay
	cb.mu.RUnlock()
	if overlay != nil {
		overlay(screen)
	}
}

// GetRect delegates to the wrapped widget.
func (cb *ComponentBase) GetRect() (int, int, int, int) {
	return cb.primitive.Rect()
}

// SetRect delegates to the wrapped widget.
func (cb *ComponentBase) SetRect(x, y, width, height int) {
	cb.primitive.SetRect(x, y, width, height)
}

// HandleKey implements core.KeyHandler — custom handler first, then primitive.
func (cb *ComponentBase) HandleKey(ev *tcell.EventKey) bool {
	cb.mu.RLock()
	custom := cb.inputHandler
	cb.mu.RUnlock()
	if custom != nil && custom(ev) {
		return true
	}
	if kh, ok := cb.primitive.(interface{ HandleKey(*tcell.EventKey) bool }); ok {
		return kh.HandleKey(ev)
	}
	return false
}

// Focus delegates to the wrapped widget.
func (cb *ComponentBase) Focus() {
	cb.primitive.Focus()
}

// Blur delegates to the wrapped widget.
func (cb *ComponentBase) Blur() {
	cb.primitive.Blur()
}

// HasFocus delegates to the wrapped widget.
func (cb *ComponentBase) HasFocus() bool {
	return cb.primitive.HasFocus()
}
