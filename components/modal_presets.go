package components

import (
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// NewConfirmModal creates a modal optimized for yes/no confirmations.
// The modal dismisses on Escape (treated as cancel) and captures all input.
func NewConfirmModal(title, message string) *Modal {
	m := NewModal(ModalConfig{
		Title:     title,
		Width:     50,
		Height:    10,
		MinWidth:  40,
		MinHeight: 8,
		Backdrop:  true,
	})

	m.SetBehavior(ModalBehavior{
		CapturesAllInput:      true,
		DismissOnEsc:          true,
		RestoreFocusOnDismiss: true,
		Backdrop:              true,
		BlockUntilDismissed:   false,
	})

	// Create message text view
	messageView := core.NewTextView()
	messageView.SetText(message)
	messageView.Box.SetBackgroundColor(theme.Bg())
	messageView.SetDynamicColors(true)

	m.SetContent(messageView)
	m.SetHints([]KeyHint{
		{Key: "Enter", Description: "Confirm"},
		{Key: "Esc", Description: "Cancel"},
	})

	return m
}

// NewFormModal creates a modal optimized for form input.
// Does NOT dismiss on Escape to prevent accidental data loss.
// Users must explicitly cancel or submit the form.
func NewFormModal(title string, width, height int) *Modal {
	if width == 0 {
		width = 60
	}
	if height == 0 {
		height = 20
	}

	m := NewModal(ModalConfig{
		Title:    title,
		Width:    width,
		Height:   height,
		Backdrop: true,
	})

	m.SetBehavior(ModalBehavior{
		CapturesAllInput:      true,
		DismissOnEsc:          false, // Don't dismiss on Esc to prevent data loss
		RestoreFocusOnDismiss: true,
		Backdrop:              true,
		BlockUntilDismissed:   false,
	})

	m.SetHints([]KeyHint{
		{Key: "Tab", Description: "Next field"},
		{Key: "Enter", Description: "Submit"},
	})

	return m
}

// NewAlertModal creates a simple alert/info modal.
// Dismisses on any key press (Enter or Escape).
func NewAlertModal(title, message string) *Modal {
	m := NewModal(ModalConfig{
		Title:     title,
		Width:     50,
		Height:    8,
		MinWidth:  30,
		MinHeight: 6,
		Backdrop:  true,
	})

	m.SetBehavior(ModalBehavior{
		CapturesAllInput:      true,
		DismissOnEsc:          true,
		RestoreFocusOnDismiss: true,
		Backdrop:              true,
		BlockUntilDismissed:   false,
	})

	messageView := core.NewTextView()
	messageView.SetText(message)
	messageView.Box.SetBackgroundColor(theme.Bg())
	messageView.SetDynamicColors(true)

	m.SetContent(messageView)
	m.SetHints([]KeyHint{
		{Key: "Enter/Esc", Description: "Close"},
	})

	return m
}

// NewBlockingModal creates a modal that blocks all other navigation.
// Use for critical confirmations that must be addressed before continuing.
// Example: "Unsaved changes will be lost. Continue?"
func NewBlockingModal(title, message string) *Modal {
	m := NewModal(ModalConfig{
		Title:     title,
		Width:     50,
		Height:    10,
		MinWidth:  40,
		MinHeight: 8,
		Backdrop:  true,
	})

	m.SetBehavior(ModalBehavior{
		CapturesAllInput:      true,
		DismissOnEsc:          true,
		RestoreFocusOnDismiss: true,
		Backdrop:              true,
		BlockUntilDismissed:   true, // Blocks other navigation
	})

	messageView := core.NewTextView()
	messageView.SetText(message)
	messageView.Box.SetBackgroundColor(theme.Bg())
	messageView.SetDynamicColors(true)

	m.SetContent(messageView)
	m.SetHints([]KeyHint{
		{Key: "Enter", Description: "Confirm"},
		{Key: "Esc", Description: "Cancel"},
	})

	return m
}

// NewFloatingPanel creates a non-modal overlay panel.
// Unlike modals, floating panels don't block input to underlying views
// and don't draw a backdrop. Use for tooltips, autocomplete dropdowns, etc.
func NewFloatingPanel(width, height int) *Modal {
	m := NewModal(ModalConfig{
		Width:    width,
		Height:   height,
		Backdrop: false,
	})

	m.SetBehavior(ModalBehavior{
		CapturesAllInput:      false, // Let events pass through
		DismissOnEsc:          true,
		RestoreFocusOnDismiss: true,
		Backdrop:              false,
		BlockUntilDismissed:   false,
	})

	return m
}

// NewLoadingModal creates a modal for displaying loading/progress state.
// Cannot be dismissed by user (no Escape handling).
// Call Close() programmatically when loading is complete.
func NewLoadingModal(title, message string) *Modal {
	m := NewModal(ModalConfig{
		Title:     title,
		Width:     40,
		Height:    6,
		MinWidth:  30,
		MinHeight: 5,
		Backdrop:  true,
	})

	m.SetBehavior(ModalBehavior{
		CapturesAllInput:      true,
		DismissOnEsc:          false, // Cannot dismiss loading modal
		RestoreFocusOnDismiss: true,
		Backdrop:              true,
		BlockUntilDismissed:   true, // Block navigation while loading
	})

	messageView := core.NewTextView()
	messageView.SetText(message)
	messageView.Box.SetBackgroundColor(theme.Bg())
	messageView.SetDynamicColors(true)

	m.SetContent(messageView)
	// No hints - user cannot interact

	return m
}
