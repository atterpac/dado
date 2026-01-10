package nav

import "github.com/atterpac/jig/components"

// Modal is the unified interface for modal components.
// Implement this interface to have your component treated as a modal
// with automatic lifecycle management by Pages.
//
// The components.Modal type already implements this interface.
//
// Example for custom modals:
//
//	type ConfirmDialog struct {
//	    *components.Modal
//	}
//
//	// Inherit ModalBehavior() and OnDismiss() from components.Modal,
//	// or override them for custom behavior:
//
//	func (c *ConfirmDialog) OnDismiss() bool {
//	    // Custom dismiss logic
//	    return c.hasUnsavedChanges == false
//	}
type Modal interface {
	Component

	// ModalBehavior returns the modal's behavior configuration.
	// This controls how the modal handles input, dismissal, and focus.
	ModalBehavior() components.ModalBehavior

	// OnDismiss is called when the modal is about to be dismissed.
	// Return false to cancel the dismiss (e.g., to show unsaved changes warning).
	// Return true to allow the dismiss to proceed.
	OnDismiss() bool
}

// IsModal returns true if the component implements the Modal interface.
func IsModal(c Component) bool {
	_, ok := c.(Modal)
	return ok
}

// AsModal attempts to cast a Component to Modal.
// Returns nil if the component is not a modal.
func AsModal(c Component) Modal {
	if m, ok := c.(Modal); ok {
		return m
	}
	return nil
}

// GetModalBehavior returns the modal behavior if the component is a modal.
// Returns nil if the component is not a modal.
func GetModalBehavior(c Component) *components.ModalBehavior {
	if m, ok := c.(Modal); ok {
		b := m.ModalBehavior()
		return &b
	}
	return nil
}
