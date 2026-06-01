package nav

import (
	"testing"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

// TestPages_SetFocusManager_API verifies the FocusManager integration compiles and wires.
func TestPages_SetFocusManager_API(t *testing.T) {
	pages := NewPages()
	fm := core.NewFocusManager()
	pages.SetFocusManager(fm) // must compile and not panic
}

// TestPages_FocusManager_SavesRestoresFocus verifies modal push saves focus
// and pop restores it via the FocusManager.
func TestPages_FocusManager_SavesRestoresFocus(t *testing.T) {
	pages := NewPages()
	fm := core.NewFocusManager()
	pages.SetFocusManager(fm)

	// Use a mock widget as the "focused view widget"
	viewWidget := coretest.NewMockWidget("view")
	fm.Focus(viewWidget)

	view := newMockComponent("main")
	modal := newMockModal("confirm")
	modal.behavior.RestoreFocusOnDismiss = true

	pages.Push(view)
	pages.Push(modal) // should save viewWidget as focus

	// Change focus while modal is open
	otherWidget := coretest.NewMockWidget("other")
	fm.Focus(otherWidget)

	// Pop modal — should restore focus to viewWidget
	pages.Pop()

	if fm.Focused() != viewWidget {
		t.Fatalf("focus not restored: got %v, want viewWidget", fm.Focused())
	}
}

// TestPages_FocusManager_NoRestoreWhenDisabled verifies focus is NOT restored
// when RestoreFocusOnDismiss is false.
func TestPages_FocusManager_NoRestoreWhenDisabled(t *testing.T) {
	pages := NewPages()
	fm := core.NewFocusManager()
	pages.SetFocusManager(fm)

	viewWidget := coretest.NewMockWidget("view")
	fm.Focus(viewWidget)

	view := newMockComponent("main")
	modal := newMockModal("confirm")
	modal.behavior.RestoreFocusOnDismiss = false // disable restore

	pages.Push(view)
	pages.Push(modal)

	otherWidget := coretest.NewMockWidget("other")
	fm.Focus(otherWidget)

	pages.Pop()

	// Focus should NOT be restored — still otherWidget
	if fm.Focused() != otherWidget {
		t.Fatalf("focus should not be restored: got %v, want otherWidget", fm.Focused())
	}
}
