// Package bus provides an opt-in pub/sub event bus for observing internal
// state changes across jig: binding mutations, theme switches, page navigation,
// async loader transitions, input events, and effect dispatches.
//
// The bus is disabled by default and incurs only a single atomic.Bool load
// at each potential publish site. Enable it with SetEnabled(true) — typically
// from a debug build or behind a key binding — to start collecting events.
//
// Subscribers receive events on a background goroutine; they must not block.
// Handlers that touch UI state should marshal back via theme.QueueUpdateDraw.
package bus

import (
	"sync"
	"sync/atomic"
	"time"
)

// Event is the unit of observation published on the bus.
type Event struct {
	// Kind is a short stable identifier ("binding.set", "theme.switch", ...).
	Kind string
	// Source identifies the producing subsystem ("binding", "theme", "nav", "async", "input", "effect").
	Source string
	// Payload carries the kind-specific typed data (see events.go).
	Payload any
	// Time is the publish timestamp.
	Time time.Time
	// Seq is a monotonic sequence number assigned by the bus.
	Seq uint64
}

// Bus is the pub/sub interface.
type Bus interface {
	// Publish records an event and fans it out to subscribers.
	// Time and Seq are populated by the bus; callers leave them zero.
	Publish(Event)

	// Subscribe registers a handler invoked for every event matching filter.
	// A nil filter receives all events. Returns an unsubscribe function.
	Subscribe(filter func(Event) bool, handler func(Event)) func()

	// Recent returns up to n most recently published events (newest last).
	Recent(n int) []Event

	// Close stops the dispatch goroutine and releases resources.
	Close()
}

var (
	enabled atomic.Bool
	defMu   sync.RWMutex
	def     Bus
)

// Enabled reports whether the bus is currently accepting publishes.
// Producers should gate Publish calls on this to keep the off-path zero-cost.
func Enabled() bool {
	return enabled.Load()
}

// SetEnabled toggles the bus on or off. When disabled, Publish is a no-op
// and no event allocations occur at gated call sites.
func SetEnabled(on bool) {
	enabled.Store(on)
}

// Default returns the process-wide default bus, creating one on first use.
func Default() Bus {
	defMu.RLock()
	b := def
	defMu.RUnlock()
	if b != nil {
		return b
	}
	defMu.Lock()
	defer defMu.Unlock()
	if def == nil {
		def = New(DefaultRingSize, DefaultQueueSize)
	}
	return def
}

// SetDefault replaces the process-wide default bus. The previous default is
// closed. Pass nil to reset (a fresh default will be lazily created).
func SetDefault(b Bus) {
	defMu.Lock()
	prev := def
	def = b
	defMu.Unlock()
	if prev != nil && prev != b {
		prev.Close()
	}
}

// Publish is a convenience wrapper for Default().Publish. It is gated by
// Enabled() so a disabled bus avoids the interface call.
//
// Producers that need to construct expensive payloads should call
//
//	if bus.Enabled() { bus.Publish(bus.Event{...}) }
//
// directly to skip payload construction when the bus is off.
func Publish(e Event) {
	if !enabled.Load() {
		return
	}
	Default().Publish(e)
}

// DefaultRingSize is the default capacity of the in-memory event ring.
const DefaultRingSize = 1024

// DefaultQueueSize is the default capacity of the dispatch channel buffer.
// Publishes that overflow the buffer are dropped to keep producers non-blocking.
const DefaultQueueSize = 256

// New constructs a Bus with the given ring buffer capacity and dispatch queue size.
func New(ringSize, queueSize int) Bus {
	if ringSize <= 0 {
		ringSize = DefaultRingSize
	}
	if queueSize <= 0 {
		queueSize = DefaultQueueSize
	}
	b := &bus{
		ring:  newRing(ringSize),
		queue: make(chan Event, queueSize),
		done:  make(chan struct{}),
	}
	go b.dispatch()
	return b
}

type subscription struct {
	id      uint64
	filter  func(Event) bool
	handler func(Event)
}

type bus struct {
	mu     sync.RWMutex
	subs   []subscription
	nextID uint64
	seq    atomic.Uint64

	ring *ring

	queue   chan Event
	done    chan struct{}
	closed  atomic.Bool
	dropped atomic.Uint64
}

func (b *bus) Publish(e Event) {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	e.Seq = b.seq.Add(1)
	b.ring.push(e)

	if b.closed.Load() {
		return
	}
	select {
	case b.queue <- e:
	default:
		b.dropped.Add(1)
	}
}

func (b *bus) Subscribe(filter func(Event) bool, handler func(Event)) func() {
	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.subs = append(b.subs, subscription{id: id, filter: filter, handler: handler})
	b.mu.Unlock()

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		for i, s := range b.subs {
			if s.id == id {
				b.subs = append(b.subs[:i], b.subs[i+1:]...)
				return
			}
		}
	}
}

func (b *bus) Recent(n int) []Event {
	return b.ring.snapshot(n)
}

func (b *bus) Close() {
	if b.closed.Swap(true) {
		return
	}
	close(b.done)
}

// Dropped returns the number of publishes dropped due to a full dispatch queue.
// Exposed for diagnostics; not part of the Bus interface.
func (b *bus) Dropped() uint64 { return b.dropped.Load() }

func (b *bus) dispatch() {
	for {
		select {
		case <-b.done:
			return
		case e, ok := <-b.queue:
			if !ok {
				return
			}
			b.fanout(e)
		}
	}
}

func (b *bus) fanout(e Event) {
	b.mu.RLock()
	subs := make([]subscription, len(b.subs))
	copy(subs, b.subs)
	b.mu.RUnlock()

	for _, s := range subs {
		if s.filter != nil && !s.filter(e) {
			continue
		}
		// Each handler runs on the dispatch goroutine; handlers must not block.
		s.handler(e)
	}
}
