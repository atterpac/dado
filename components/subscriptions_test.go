package components

import (
	"testing"

	"github.com/atterpac/dado/core"
)

func TestSubscriptions_ReleaseInvokesAllLIFO(t *testing.T) {
	var s Subscriptions
	var order []int
	s.Add(func() { order = append(order, 1) })
	s.Add(func() { order = append(order, 2) })
	s.Add(func() { order = append(order, 3) })

	if s.Len() != 3 {
		t.Fatalf("Len=%d want 3", s.Len())
	}

	s.Release()

	if got, want := order, []int{3, 2, 1}; !equalInts(got, want) {
		t.Fatalf("order=%v want %v", got, want)
	}
	if s.Len() != 0 {
		t.Fatalf("Len=%d want 0 after Release", s.Len())
	}
}

func TestSubscriptions_ReleaseIdempotent(t *testing.T) {
	var s Subscriptions
	calls := 0
	s.Add(func() { calls++ })
	s.Release()
	s.Release()
	if calls != 1 {
		t.Fatalf("calls=%d want 1", calls)
	}
}

func TestSubscriptions_NilIgnored(t *testing.T) {
	var s Subscriptions
	s.Add(nil)
	if s.Len() != 0 {
		t.Fatalf("nil unsub stored")
	}
	s.Release() // must not panic
}

func TestComponentBase_StopReleasesSubs(t *testing.T) {
	cb := NewComponentBase(new(core.Box))
	released := false
	cb.Subs().Add(func() { released = true })

	cb.Stop()

	if !released {
		t.Fatal("Stop did not release subscriptions")
	}
	if cb.Subs().Len() != 0 {
		t.Fatal("subs not cleared after Stop")
	}
}

func TestComponentBase_StopRunsOnStopBeforeRelease(t *testing.T) {
	cb := NewComponentBase(new(core.Box))
	var order []string
	cb.SetOnStop(func() { order = append(order, "onStop") })
	cb.Subs().Add(func() { order = append(order, "unsub") })

	cb.Stop()

	if len(order) != 2 || order[0] != "onStop" || order[1] != "unsub" {
		t.Fatalf("order=%v want [onStop unsub]", order)
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
