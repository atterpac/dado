// Package anim provides a shared frame clock for time-based widgets.
//
// Spinners, pulses, marquees and similar animations register a callback via
// Subscribe and are ticked on a single shared goroutine. Each tick runs on the
// UI thread (via theme.QueueUpdateDraw), so subscriber callbacks need no
// locking of their own frame state and every tick coalesces into one redraw.
//
// The goroutine is created lazily on the first subscription and exits when the
// last subscription is cancelled — an idle app spawns nothing and stays purely
// event-driven. Subscribers may use different intervals; the loop sleeps until
// the earliest due deadline rather than polling at a fixed base rate.
package anim

import (
	"sync"
	"time"

	"github.com/atterpac/dado/theme"
)

const defaultFrameInterval = 100 * time.Millisecond

type frameSub struct {
	interval time.Duration
	fn       func()
	next     time.Time
}

var (
	tickMu     sync.Mutex
	tickSubs   map[int]*frameSub
	tickNextID int
	tickWake   chan struct{}
	tickRun    bool
)

// Subscribe registers fn to be invoked every interval on the UI thread, each
// invocation triggering a redraw. The returned cancel func removes the
// subscription; it is safe to call multiple times and from any goroutine.
//
// A non-positive interval falls back to defaultFrameInterval.
func Subscribe(interval time.Duration, fn func()) (cancel func()) {
	if interval <= 0 {
		interval = defaultFrameInterval
	}

	tickMu.Lock()
	if tickSubs == nil {
		tickSubs = make(map[int]*frameSub)
	}
	id := tickNextID
	tickNextID++
	tickSubs[id] = &frameSub{interval: interval, fn: fn, next: time.Now().Add(interval)}
	if !tickRun {
		tickRun = true
		tickWake = make(chan struct{}, 1)
		go runTicker()
	} else {
		wakeTicker()
	}
	tickMu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			tickMu.Lock()
			delete(tickSubs, id)
			wakeTicker()
			tickMu.Unlock()
		})
	}
}

// wakeTicker nudges the loop to recompute its deadline. Caller holds tickMu.
func wakeTicker() {
	if tickWake == nil {
		return
	}
	select {
	case tickWake <- struct{}{}:
	default:
	}
}

func runTicker() {
	for {
		tickMu.Lock()
		if len(tickSubs) == 0 {
			tickRun = false
			tickMu.Unlock()
			return
		}
		var earliest time.Time
		first := true
		for _, s := range tickSubs {
			if first || s.next.Before(earliest) {
				earliest, first = s.next, false
			}
		}
		tickMu.Unlock()

		if d := time.Until(earliest); d > 0 {
			t := time.NewTimer(d)
			select {
			case <-t.C:
			case <-tickWake:
				t.Stop()
				continue // subscription set changed; recompute deadline
			}
		}

		now := time.Now()
		var due []func()
		tickMu.Lock()
		for _, s := range tickSubs {
			if s.next.After(now) {
				continue
			}
			due = append(due, s.fn)
			// Advance one interval, snapping forward past any missed frames so
			// a slow UI thread never builds a backlog of catch-up ticks.
			s.next = s.next.Add(s.interval)
			if s.next.Before(now) {
				s.next = now.Add(s.interval)
			}
		}
		tickMu.Unlock()

		for _, fn := range due {
			theme.QueueUpdateDraw(fn)
		}
	}
}
