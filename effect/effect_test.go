package effect

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// theme.QueueUpdateDraw with no registered app falls back to immediate
// execution, so tests run synchronously without a core.App.

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

type msgString string
type msgInt int

func TestRun_DeliversMsg(t *testing.T) {
	d := NewDispatcher()
	t.Cleanup(func() { _ = d.Shutdown(context.Background()) })

	var got atomic.Value
	d.Subscribe(nil, func(m Msg) { got.Store(m) })

	d.Run(func(_ context.Context) Msg { return msgString("hi") })

	waitFor(t, func() bool { v := got.Load(); return v != nil && v.(Msg) == msgString("hi") }, time.Second)
}

func TestRun_NilMsgNotDelivered(t *testing.T) {
	d := NewDispatcher()
	t.Cleanup(func() { _ = d.Shutdown(context.Background()) })

	var hits atomic.Int32
	d.Subscribe(nil, func(Msg) { hits.Add(1) })

	d.Run(func(_ context.Context) Msg { return nil })
	time.Sleep(20 * time.Millisecond)

	if hits.Load() != 0 {
		t.Fatal("nil Msg fanned out to handler")
	}
}

func TestOn_TypeFiltered(t *testing.T) {
	d := NewDispatcher()
	t.Cleanup(func() { _ = d.Shutdown(context.Background()) })

	var seen atomic.Int32
	On(d, func(m msgInt) {
		if m == 42 {
			seen.Add(1)
		}
	})

	d.Run(func(_ context.Context) Msg { return msgString("ignored") })
	d.Run(func(_ context.Context) Msg { return msgInt(42) })

	waitFor(t, func() bool { return seen.Load() == 1 }, time.Second)
}

func TestUnsubscribe(t *testing.T) {
	d := NewDispatcher()
	t.Cleanup(func() { _ = d.Shutdown(context.Background()) })

	var hits atomic.Int32
	unsub := d.Subscribe(nil, func(Msg) { hits.Add(1) })

	d.Run(func(_ context.Context) Msg { return msgInt(1) })
	waitFor(t, func() bool { return hits.Load() == 1 }, time.Second)

	unsub()
	d.Run(func(_ context.Context) Msg { return msgInt(2) })
	time.Sleep(30 * time.Millisecond)

	if hits.Load() != 1 {
		t.Fatalf("handler fired after unsubscribe: hits=%d", hits.Load())
	}
}

func TestBatch_FansOutChildren(t *testing.T) {
	d := NewDispatcher()
	t.Cleanup(func() { _ = d.Shutdown(context.Background()) })

	var got atomic.Int32
	On(d, func(m msgInt) { got.Add(int32(m)) })

	d.Run(Batch(
		func(_ context.Context) Msg { return msgInt(1) },
		func(_ context.Context) Msg { return msgInt(2) },
		func(_ context.Context) Msg { return msgInt(4) },
	))

	waitFor(t, func() bool { return got.Load() == 7 }, time.Second)
}

func TestSequence_PreservesOrder(t *testing.T) {
	d := NewDispatcher()
	t.Cleanup(func() { _ = d.Shutdown(context.Background()) })

	var (
		mu  sync.Mutex
		out []int
	)
	On(d, func(m msgInt) {
		mu.Lock()
		out = append(out, int(m))
		mu.Unlock()
	})

	// First effect is intentionally slower; Sequence must still deliver in order.
	d.Run(Sequence(
		func(_ context.Context) Msg { time.Sleep(40 * time.Millisecond); return msgInt(1) },
		func(_ context.Context) Msg { return msgInt(2) },
		func(_ context.Context) Msg { return msgInt(3) },
	))

	waitFor(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(out) == 3
	}, 2*time.Second)

	mu.Lock()
	defer mu.Unlock()
	for i, want := range []int{1, 2, 3} {
		if out[i] != want {
			t.Fatalf("out[%d]=%d want %d (full=%v)", i, out[i], want, out)
		}
	}
}

func TestTick_FiresAfterDuration(t *testing.T) {
	d := NewDispatcher()
	t.Cleanup(func() { _ = d.Shutdown(context.Background()) })

	var fired atomic.Int32
	On(d, func(_ time.Time) { fired.Add(1) })

	d.Run(Tick(20*time.Millisecond, func(t time.Time) Msg { return t }))

	waitFor(t, func() bool { return fired.Load() == 1 }, time.Second)
}

func TestTick_CancelledByShutdown(t *testing.T) {
	d := NewDispatcher()

	var fired atomic.Int32
	On(d, func(_ time.Time) { fired.Add(1) })

	d.Run(Tick(500*time.Millisecond, func(t time.Time) Msg { return t }))

	// Shutdown before the tick fires.
	time.Sleep(20 * time.Millisecond)
	if err := d.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	time.Sleep(600 * time.Millisecond)
	if fired.Load() != 0 {
		t.Fatal("Tick fired after Shutdown")
	}
}

func TestShutdown_PostShutdownRunIsNoop(t *testing.T) {
	d := NewDispatcher()
	_ = d.Shutdown(context.Background())

	var hits atomic.Int32
	d.Subscribe(nil, func(Msg) { hits.Add(1) })
	d.Run(func(_ context.Context) Msg { return msgInt(1) })

	time.Sleep(30 * time.Millisecond)
	if hits.Load() != 0 {
		t.Fatalf("Run delivered after Shutdown: hits=%d", hits.Load())
	}
}

func TestShutdown_Idempotent(t *testing.T) {
	d := NewDispatcher()
	if err := d.Shutdown(context.Background()); err != nil {
		t.Fatalf("first Shutdown: %v", err)
	}
	if err := d.Shutdown(context.Background()); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
}

func TestNone_EmitsNothing(t *testing.T) {
	d := NewDispatcher()
	t.Cleanup(func() { _ = d.Shutdown(context.Background()) })

	var hits atomic.Int32
	d.Subscribe(nil, func(Msg) { hits.Add(1) })

	d.Run(None())
	time.Sleep(20 * time.Millisecond)

	if hits.Load() != 0 {
		t.Fatal("None() emitted a Msg")
	}
}
