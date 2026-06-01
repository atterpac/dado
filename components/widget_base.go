package components

import (
	"sync"
	"sync/atomic"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// widgetBase is the shared embeddable base for leaf widgets. It bundles the
// four things every widget previously declared and wired by hand:
//
//   - the embedded core.Box (value type, no heap alloc)
//   - a sync.RWMutex guarding widget state
//   - a Subscriptions set released by ComponentBase.Stop via Subs()
//   - a theme.Provider hook so reads route through th() (base-routed theming)
//
// Embed it by value and call initWidget in the constructor:
//
//	type Badge struct {
//	    widgetBase
//	    text string
//	}
//
//	func NewBadge(text string) *Badge {
//	    b := &Badge{text: text}
//	    b.initWidget()
//	    return b
//	}
//
// Never copy a widgetBase after initWidget (it holds a mutex); widgets are
// always used as heap pointers, so this is not a concern in practice.
type widgetBase struct {
	core.Box // embedded value — promotes Draw, Rect, SetRect, InnerRect, GetInnerRect, GetRect, Focus, Blur, HasFocus, DrawForSubclass, SetBorder, SetTitle, SetPadding, SetBackgroundColor, etc.

	mu   sync.RWMutex
	subs Subscriptions
	// themeP is the scoped theme provider, read on the Draw hot path via th().
	themeP atomic.Pointer[theme.Provider]
}

// initWidget wires the box for theme-aware background (via core.SetDefaultBackgroundFunc).
func (w *widgetBase) initWidget() {}

// Subs returns the widget's subscription set; released by ComponentBase.Stop.
func (w *widgetBase) Subs() *Subscriptions { return &w.subs }

// SetTheme scopes a theme.Provider to this widget.
func (w *widgetBase) SetTheme(p *theme.Provider) {
	w.themeP.Store(p)
}

// th returns the scoped Provider if SetTheme was called, otherwise theme.Default().
func (w *widgetBase) th() *theme.Provider {
	if p := w.themeP.Load(); p != nil {
		return p
	}
	return theme.Default()
}
