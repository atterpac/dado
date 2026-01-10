package nav

import (
	"github.com/rivo/tview"

	"github.com/atterpac/jig/components"
)

// Component represents a navigable view/page.
// All views pushed to Pages must implement this interface.
type Component interface {
	tview.Primitive

	// Start is called when the component becomes active (shown).
	Start()

	// Stop is called when the component becomes inactive (hidden).
	Stop()

	// Hints returns key binding hints for this component.
	Hints() []components.KeyHint
}
