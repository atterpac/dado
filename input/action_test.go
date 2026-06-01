package input

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestCtrlActionDoesNotMatchPlainRune guards against the regression where a
// Ctrl+<rune> action (e.g. Ctrl+I) matched a plain <rune> press (plain 'i'),
// because the plain-rune match ran before the Ctrl-modifier check.
func TestCtrlActionDoesNotMatchPlainRune(t *testing.T) {
	fired := false
	r := NewActionRegistry().
		AddCtrl("jump-forward", 'i', "Jump forward", func() { fired = true })

	// Plain 'i' must NOT trigger the Ctrl+I action.
	if r.Handle(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone)) {
		t.Fatal("plain 'i' should not be handled by a Ctrl+I action")
	}
	if fired {
		t.Fatal("Ctrl+I handler fired on a plain 'i' press")
	}

	// Ctrl+I (delivered by tcell as KeyTab) must trigger it.
	if !r.Handle(tcell.NewEventKey(tcell.KeyCtrlI, 0, tcell.ModCtrl)) {
		t.Fatal("Ctrl+I should be handled by the Ctrl+I action")
	}
	if !fired {
		t.Fatal("Ctrl+I handler did not fire on a Ctrl+I press")
	}
}

// TestSimpleRuneActionMatches confirms plain-rune actions still work.
func TestSimpleRuneActionMatches(t *testing.T) {
	count := 0
	r := NewActionRegistry().
		AddSimple("theme", 'T', "Theme", func() { count++ })

	if !r.Handle(tcell.NewEventKey(tcell.KeyRune, 'T', tcell.ModNone)) {
		t.Fatal("plain 'T' should trigger the simple action")
	}
	if count != 1 {
		t.Fatalf("expected handler to fire once, got %d", count)
	}
}
