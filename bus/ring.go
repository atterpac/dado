package bus

import "sync"

// ring is a fixed-capacity circular buffer of Events with newest-last snapshot semantics.
type ring struct {
	mu    sync.Mutex
	buf   []Event
	head  int  // next write index
	full  bool // true once the buffer has wrapped at least once
}

func newRing(cap int) *ring {
	return &ring{buf: make([]Event, cap)}
}

func (r *ring) push(e Event) {
	r.mu.Lock()
	r.buf[r.head] = e
	r.head++
	if r.head == len(r.buf) {
		r.head = 0
		r.full = true
	}
	r.mu.Unlock()
}

// snapshot returns up to n most-recent events in chronological order (oldest first).
// If n <= 0 or n exceeds the stored count, all stored events are returned.
func (r *ring) snapshot(n int) []Event {
	r.mu.Lock()
	defer r.mu.Unlock()

	size := r.head
	if r.full {
		size = len(r.buf)
	}
	if size == 0 {
		return nil
	}
	if n <= 0 || n > size {
		n = size
	}

	out := make([]Event, n)
	// Walk backwards from the most recent event.
	idx := r.head - 1
	if idx < 0 {
		idx = len(r.buf) - 1
	}
	for i := n - 1; i >= 0; i-- {
		out[i] = r.buf[idx]
		idx--
		if idx < 0 {
			idx = len(r.buf) - 1
		}
	}
	return out
}
