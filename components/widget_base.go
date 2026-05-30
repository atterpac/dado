package components

import (
	"sync"
	"sync/atomic"

	"github.com/rivo/tview"

	"github.com/atterpac/dado/theme"
)

// widgetBase is the shared embeddable base for leaf widgets. It bundles the
// four things every widget previously declared and wired by hand:
//
//   - the embedded *tview.Box (primitive plumbing)
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
//	    b.initWidget(tview.NewBox())
//	    return b
//	}
//
// Never copy a widgetBase after initWidget (it holds a mutex); widgets are
// always used as heap pointers, so this is not a concern in practice.
type widgetBase struct {
	*tview.Box

	mu   sync.RWMutex
	subs Subscriptions
	// themeP is the scoped theme provider, read on the Draw hot path via th().
	// It is stored atomically rather than under mu because Draw methods routinely
	// hold mu (read or write) for their own state and then call th(); routing the
	// theme read through mu would self-deadlock a writer-holding Draw and risk a
	// recursive-RLock stall under a concurrent SetTheme. Independent of mu, the
	// theme pointer needs no such coupling.
	themeP atomic.Pointer[theme.Provider]
}

// initWidget wires the box: sets its background to the active theme and
// registers it for automatic background updates, recording the unregister in
// the widget's Subscriptions so teardown releases it.
func (w *widgetBase) initWidget(box *tview.Box) {
	w.Box = box
	box.SetBackgroundColor(theme.Bg())
	w.subs.Add(theme.Register(box))
}

// Subs returns the widget's subscription set; released by ComponentBase.Stop.
func (w *widgetBase) Subs() *Subscriptions { return &w.subs }

// SetTheme scopes a theme.Provider to this widget. Pass nil to fall back to
// the package default. Reads via th() honor the override.
func (w *widgetBase) SetTheme(p *theme.Provider) {
	w.themeP.Store(p)
}

// th returns the scoped Provider if SetTheme was called, otherwise
// theme.Default(). Widget Draw methods read colors via w.th().Bg() so a
// per-subtree theme override takes effect without threading state through the
// fixed tview.Primitive.Draw signature. Lock-free: see themeP.
func (w *widgetBase) th() *theme.Provider {
	if p := w.themeP.Load(); p != nil {
		return p
	}
	return theme.Default()
}
