package anim

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func tickerRunning() bool {
	tickMu.Lock()
	defer tickMu.Unlock()
	return tickRun
}

// No goroutine exists until something subscribes, and it exits once the last
// subscription is cancelled — the idle app stays purely event-driven.
func TestSubscribe_GoroutineLifecycle(t *testing.T) {
	if tickerRunning() {
		t.Fatal("ticker running before any subscription")
	}

	var ticks int64
	cancel := Subscribe(5*time.Millisecond, func() { atomic.AddInt64(&ticks, 1) })

	if !tickerRunning() {
		t.Fatal("ticker not running after subscribe")
	}

	time.Sleep(40 * time.Millisecond)
	if atomic.LoadInt64(&ticks) == 0 {
		t.Fatal("callback never fired")
	}

	cancel()
	// Give the loop a moment to observe the empty set and exit.
	deadline := time.Now().Add(200 * time.Millisecond)
	for tickerRunning() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if tickerRunning() {
		t.Fatal("ticker still running after last cancel")
	}
}

func TestSubscribe_CancelIdempotent(t *testing.T) {
	cancel := Subscribe(time.Hour, func() {})
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); cancel() }()
	}
	wg.Wait()
}
