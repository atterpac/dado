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

// ModalBehavior configures how a modal handles input and lifecycle.
type ModalBehavior struct {
	// CapturesAllInput prevents input from reaching underlying views.
	// When true, only the modal receives keyboard input.
	// Default: true
	CapturesAllInput bool

	// DismissOnEsc automatically dismisses the modal when Escape is pressed.
	// Default: true
	DismissOnEsc bool

	// RestoreFocusOnDismiss returns focus to the previous component when dismissed.
	// Default: true
	RestoreFocusOnDismiss bool

	// Backdrop draws a semi-transparent overlay behind the modal.
	// Default: true
	Backdrop bool

	// BlockUntilDismissed prevents other stack operations until this modal is dismissed.
	// Use for critical confirmations that must be addressed.
	// Default: false
	BlockUntilDismissed bool
}

// DefaultModalBehavior returns the standard modal behavior settings.
// All boolean fields except BlockUntilDismissed default to true.
func DefaultModalBehavior() ModalBehavior {
	return ModalBehavior{
		CapturesAllInput:      true,
		DismissOnEsc:          true,
		RestoreFocusOnDismiss: true,
		Backdrop:              true,
		BlockUntilDismissed:   false,
	}
}

// ModalComponent is implemented by components that behave as modals.
// When pushed to Pages, components implementing this interface receive
// automatic modal handling based on their ModalBehavior() configuration.
//
// Example implementation:
//
//	type MyModal struct {
//	    *components.Modal
//	}
//
//	func (m *MyModal) ModalBehavior() nav.ModalBehavior {
//	    return nav.DefaultModalBehavior()
//	}
//
//	func (m *MyModal) OnDismiss() bool {
//	    // Return false to cancel dismiss (e.g., unsaved changes)
//	    return true
//	}
type ModalComponent interface {
	Component

	// ModalBehavior returns the modal's behavior configuration.
	ModalBehavior() ModalBehavior

	// OnDismiss is called when the modal is about to be dismissed.
	// Return false to cancel the dismiss (e.g., to show unsaved changes warning).
	// Return true to allow the dismiss to proceed.
	OnDismiss() bool
}

// ModalMarker is implemented by components that identify as modals.
// This is used for quick modal detection without full ModalComponent implementation.
type ModalMarker interface {
	IsModal() bool
}

// IsModal returns true if the component implements ModalComponent or ModalMarker.
func IsModal(c Component) bool {
	if _, ok := c.(ModalComponent); ok {
		return true
	}
	if marker, ok := c.(ModalMarker); ok {
		return marker.IsModal()
	}
	return false
}

// ModalBehaviorProvider is implemented by components that provide modal behavior
// but cannot import the nav package (to avoid import cycles).
type ModalBehaviorProvider interface {
	GetModalBehavior() (capturesAllInput, dismissOnEsc, restoreFocusOnDismiss, backdrop, blockUntilDismissed bool)
}

// ModalDismissHandler is implemented by components that need to handle dismiss events.
type ModalDismissHandler interface {
	OnDismissNav() bool
}

// GetModalBehavior returns the modal behavior if the component is a modal.
// Returns nil if the component is not a modal.
func GetModalBehavior(c Component) *ModalBehavior {
	// Check for full ModalComponent implementation
	if m, ok := c.(ModalComponent); ok {
		b := m.ModalBehavior()
		return &b
	}

	// Check for ModalBehaviorProvider (used by components.Modal)
	if bp, ok := c.(ModalBehaviorProvider); ok {
		captures, dismiss, restore, backdrop, block := bp.GetModalBehavior()
		return &ModalBehavior{
			CapturesAllInput:      captures,
			DismissOnEsc:          dismiss,
			RestoreFocusOnDismiss: restore,
			Backdrop:              backdrop,
			BlockUntilDismissed:   block,
		}
	}

	// If it's a modal marker but doesn't provide behavior, use defaults
	if IsModal(c) {
		b := DefaultModalBehavior()
		return &b
	}

	return nil
}

// AsModal attempts to cast a Component to ModalComponent.
// Returns nil if the component is not a modal.
func AsModal(c Component) ModalComponent {
	if m, ok := c.(ModalComponent); ok {
		return m
	}
	return nil
}

// modalComponentAdapter adapts components with modal-like methods to ModalComponent.
// This is used internally when components.Modal is pushed to Pages.
type modalComponentAdapter struct {
	Component
	getBehavior func() ModalBehavior
	onDismiss   func() bool
}

func (a *modalComponentAdapter) ModalBehavior() ModalBehavior {
	if a.getBehavior != nil {
		return a.getBehavior()
	}
	return DefaultModalBehavior()
}

func (a *modalComponentAdapter) OnDismiss() bool {
	if a.onDismiss != nil {
		return a.onDismiss()
	}
	return true
}
