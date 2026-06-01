package nav

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/components"
)

// Component represents a navigable view/page.
// All views pushed to Pages must implement this interface.
//
// The rendering surface (Draw + SetRect + Blur + HasFocus) is a subset of
// core.Widget that all dado widgets and composite
// components can satisfy. Focus() is intentionally omitted: Pages routes
// key events via HandleKey type-assertions; focus ownership is managed by
// the FocusManager in layout.App.
type Component interface {
	Draw(screen tcell.Screen)
	SetRect(x, y, w, h int)
	Blur()
	HasFocus() bool

	// Name returns the display name for this component (used in breadcrumbs).
	Name() string

	// Start is called when the component becomes active (shown).
	Start()

	// Stop is called when the component becomes inactive (hidden).
	Stop()

	// Hints returns key binding hints for this component.
	Hints() []components.KeyHint
}
