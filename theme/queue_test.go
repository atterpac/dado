package theme_test

import (
	"testing"

	"github.com/atterpac/dado/theme"
)

// TestSetQueue_OverridesQueueMechanism verifies that SetQueue replaces the
// QueueUpdateDraw mechanism without requiring a core.App.
func TestSetQueue_QueueUpdateDraw_UsesFn(t *testing.T) {
	var called int
	theme.SetQueue(func(fn func()) {
		called++
		fn()
	})
	defer theme.SetQueue(nil) // reset after test

	executed := false
	theme.QueueUpdateDraw(func() { executed = true })

	if called != 1 {
		t.Fatalf("SetQueue fn called %d times, want 1", called)
	}
	if !executed {
		t.Fatal("queued function was not executed")
	}
}

func TestSetQueue_Nil_ExecutesDirectly(t *testing.T) {
	theme.SetQueue(nil)

	executed := false
	theme.QueueUpdateDraw(func() { executed = true })

	if !executed {
		t.Fatal("with nil SetQueue, function should execute directly")
	}
}
