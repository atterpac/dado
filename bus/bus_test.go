package bus

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func waitFor(t *testing.T, cond func() bool, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("condition not met within %v", timeout)
}

func TestEnabledGate_OffDoesNotPublish(t *testing.T) {
	SetEnabled(false)
	t.Cleanup(func() { SetEnabled(false) })

	b := New(16, 16)
	t.Cleanup(b.Close)
	SetDefault(b)
	t.Cleanup(func() { SetDefault(nil) })

	var seen atomic.Int32
	b.Subscribe(nil, func(Event) { seen.Add(1) })

	Publish(Event{Kind: "x"})
	time.Sleep(10 * time.Millisecond)

	if got := seen.Load(); got != 0 {
		t.Fatalf("publish leaked while disabled: got %d", got)
	}
}

func TestPublishAndSubscribe(t *testing.T) {
	SetEnabled(true)
	t.Cleanup(func() { SetEnabled(false) })

	b := New(16, 16)
	t.Cleanup(b.Close)

	var (
		mu  sync.Mutex
		got []Event
	)
	b.Subscribe(nil, func(e Event) {
		mu.Lock()
		got = append(got, e)
		mu.Unlock()
	})

	for i := 0; i < 3; i++ {
		b.Publish(Event{Kind: "k", Source: "s", Payload: i})
	}

	waitFor(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(got) == 3
	}, time.Second)

	mu.Lock()
	defer mu.Unlock()
	for i, e := range got {
		if e.Seq != uint64(i+1) {
			t.Fatalf("event %d: seq=%d want %d", i, e.Seq, i+1)
		}
		if e.Time.IsZero() {
			t.Fatalf("event %d: time not set", i)
		}
	}
}

func TestSubscribeFilter(t *testing.T) {
	SetEnabled(true)
	t.Cleanup(func() { SetEnabled(false) })
	b := New(16, 16)
	t.Cleanup(b.Close)

	var hits atomic.Int32
	b.Subscribe(func(e Event) bool { return e.Kind == "want" }, func(Event) { hits.Add(1) })

	b.Publish(Event{Kind: "skip"})
	b.Publish(Event{Kind: "want"})
	b.Publish(Event{Kind: "want"})
	b.Publish(Event{Kind: "skip"})

	waitFor(t, func() bool { return hits.Load() == 2 }, time.Second)
}

func TestUnsubscribe(t *testing.T) {
	SetEnabled(true)
	t.Cleanup(func() { SetEnabled(false) })
	b := New(16, 16)
	t.Cleanup(b.Close)

	var hits atomic.Int32
	unsub := b.Subscribe(nil, func(Event) { hits.Add(1) })

	b.Publish(Event{Kind: "1"})
	waitFor(t, func() bool { return hits.Load() == 1 }, time.Second)

	unsub()
	b.Publish(Event{Kind: "2"})
	time.Sleep(20 * time.Millisecond)

	if got := hits.Load(); got != 1 {
		t.Fatalf("handler fired after unsubscribe: hits=%d", got)
	}
}

func TestRecentRingChronological(t *testing.T) {
	SetEnabled(true)
	t.Cleanup(func() { SetEnabled(false) })
	b := New(4, 16)
	t.Cleanup(b.Close)

	for i := 0; i < 6; i++ {
		b.Publish(Event{Kind: "k", Payload: i})
	}

	got := b.Recent(0)
	if len(got) != 4 {
		t.Fatalf("ring size=%d want 4", len(got))
	}
	for i, e := range got {
		want := i + 2 // dropped first two
		if e.Payload.(int) != want {
			t.Fatalf("ring[%d].Payload=%v want %d", i, e.Payload, want)
		}
	}

	last3 := b.Recent(3)
	if len(last3) != 3 || last3[2].Payload.(int) != 5 {
		t.Fatalf("Recent(3) = %v want last=5", last3)
	}
}

func TestPublishNonBlockingOnFullQueue(t *testing.T) {
	SetEnabled(true)
	t.Cleanup(func() { SetEnabled(false) })

	// Tiny queue, slow handler — Publish must not block.
	b := New(64, 1)
	t.Cleanup(b.Close)

	block := make(chan struct{})
	b.Subscribe(nil, func(Event) { <-block })

	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			b.Publish(Event{Kind: "k"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked when queue was full")
	}
	close(block)
}
