// Package testutil provides testing utilities for the jig TUI component library.
package testutil

import (
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/components"
)

// EventCollector captures events for testing with thread-safe access.
type EventCollector[T any] struct {
	mu     sync.Mutex
	events []T
	notify chan struct{}
}

// NewEventCollector creates a new EventCollector for capturing events.
func NewEventCollector[T any]() *EventCollector[T] {
	return &EventCollector[T]{
		events: make([]T, 0),
		notify: make(chan struct{}, 100),
	}
}

// Collect adds an event to the collector.
func (ec *EventCollector[T]) Collect(event T) {
	ec.mu.Lock()
	ec.events = append(ec.events, event)
	ec.mu.Unlock()
	select {
	case ec.notify <- struct{}{}:
	default:
	}
}

// Events returns a copy of all collected events.
func (ec *EventCollector[T]) Events() []T {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	result := make([]T, len(ec.events))
	copy(result, ec.events)
	return result
}

// Last returns the most recent event and true, or zero value and false if empty.
func (ec *EventCollector[T]) Last() (T, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	if len(ec.events) == 0 {
		var zero T
		return zero, false
	}
	return ec.events[len(ec.events)-1], true
}

// First returns the first event and true, or zero value and false if empty.
func (ec *EventCollector[T]) First() (T, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	if len(ec.events) == 0 {
		var zero T
		return zero, false
	}
	return ec.events[0], true
}

// Count returns the number of collected events.
func (ec *EventCollector[T]) Count() int {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return len(ec.events)
}

// Clear removes all collected events.
func (ec *EventCollector[T]) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.events = ec.events[:0]
}

// WaitForCount waits until n events are collected or timeout occurs.
func (ec *EventCollector[T]) WaitForCount(t *testing.T, n int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if ec.Count() >= n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %d events, got %d", n, ec.Count())
		}
		select {
		case <-ec.notify:
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// LifecycleRecorder tracks Start/Stop lifecycle calls.
type LifecycleRecorder struct {
	mu     sync.Mutex
	calls  []string
	notify chan struct{}
}

// NewLifecycleRecorder creates a new LifecycleRecorder.
func NewLifecycleRecorder() *LifecycleRecorder {
	return &LifecycleRecorder{
		calls:  make([]string, 0),
		notify: make(chan struct{}, 100),
	}
}

// RecordStart returns a function that records a "start" call.
func (lr *LifecycleRecorder) RecordStart() func() {
	return func() {
		lr.mu.Lock()
		lr.calls = append(lr.calls, "start")
		lr.mu.Unlock()
		select {
		case lr.notify <- struct{}{}:
		default:
		}
	}
}

// RecordStop returns a function that records a "stop" call.
func (lr *LifecycleRecorder) RecordStop() func() {
	return func() {
		lr.mu.Lock()
		lr.calls = append(lr.calls, "stop")
		lr.mu.Unlock()
		select {
		case lr.notify <- struct{}{}:
		default:
		}
	}
}

// Record returns a function that records a custom call.
func (lr *LifecycleRecorder) Record(name string) func() {
	return func() {
		lr.mu.Lock()
		lr.calls = append(lr.calls, name)
		lr.mu.Unlock()
		select {
		case lr.notify <- struct{}{}:
		default:
		}
	}
}

// Calls returns a copy of all recorded calls.
func (lr *LifecycleRecorder) Calls() []string {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	result := make([]string, len(lr.calls))
	copy(result, lr.calls)
	return result
}

// Clear removes all recorded calls.
func (lr *LifecycleRecorder) Clear() {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	lr.calls = lr.calls[:0]
}

// Count returns the number of recorded calls.
func (lr *LifecycleRecorder) Count() int {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	return len(lr.calls)
}

// WaitForCount waits until n calls are recorded or timeout occurs.
func (lr *LifecycleRecorder) WaitForCount(t *testing.T, n int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if lr.Count() >= n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %d lifecycle calls, got %d", n, lr.Count())
		}
		select {
		case <-lr.notify:
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// SimulateKey creates a tcell key event for testing.
func SimulateKey(key tcell.Key, r rune, mod tcell.ModMask) *tcell.EventKey {
	return tcell.NewEventKey(key, r, mod)
}

// SimulateRune creates a rune key event (no modifiers).
func SimulateRune(r rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
}

// SimulateEnter creates an Enter key event.
func SimulateEnter() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
}

// SimulateEscape creates an Escape key event.
func SimulateEscape() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
}

// SimulateTab creates a Tab key event.
func SimulateTab() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
}

// SimulateBacktab creates a Shift+Tab key event.
func SimulateBacktab() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModShift)
}

// SimulateUp creates an Up arrow key event.
func SimulateUp() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
}

// SimulateDown creates a Down arrow key event.
func SimulateDown() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
}

// SimulateLeft creates a Left arrow key event.
func SimulateLeft() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
}

// SimulateRight creates a Right arrow key event.
func SimulateRight() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
}

// SimulateSpace creates a Space key event.
func SimulateSpace() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
}

// SimulateBackspace creates a Backspace key event.
func SimulateBackspace() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone)
}

// SimulateDelete creates a Delete key event.
func SimulateDelete() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone)
}

// TypeString simulates typing a string by invoking the handler for each character.
func TypeString(handler func(*tcell.EventKey, func(tview.Primitive)), s string) {
	for _, r := range s {
		handler(SimulateRune(r), nil)
	}
}

// MockComponent implements nav.Component for testing.
type MockComponent struct {
	tview.Primitive
	name         string
	StartCalled  bool
	StopCalled   bool
	StartCount   int
	StopCount    int
	hints        []components.KeyHint
	onStart      func()
	onStop       func()
}

// NewMockComponent creates a new MockComponent with the given name.
func NewMockComponent(name string) *MockComponent {
	return &MockComponent{
		Primitive: tview.NewBox(),
		name:      name,
		hints:     make([]components.KeyHint, 0),
	}
}

// Start implements nav.Component.
func (m *MockComponent) Start() {
	m.StartCalled = true
	m.StartCount++
	if m.onStart != nil {
		m.onStart()
	}
}

// Stop implements nav.Component.
func (m *MockComponent) Stop() {
	m.StopCalled = true
	m.StopCount++
	if m.onStop != nil {
		m.onStop()
	}
}

// Hints implements nav.Component.
func (m *MockComponent) Hints() []components.KeyHint {
	return m.hints
}

// SetHints sets the key hints for the mock component.
func (m *MockComponent) SetHints(hints []components.KeyHint) *MockComponent {
	m.hints = hints
	return m
}

// SetOnStart sets a callback to invoke when Start is called.
func (m *MockComponent) SetOnStart(fn func()) *MockComponent {
	m.onStart = fn
	return m
}

// SetOnStop sets a callback to invoke when Stop is called.
func (m *MockComponent) SetOnStop(fn func()) *MockComponent {
	m.onStop = fn
	return m
}

// Reset clears the called flags and counts.
func (m *MockComponent) Reset() {
	m.StartCalled = false
	m.StopCalled = false
	m.StartCount = 0
	m.StopCount = 0
}

// Name returns the component name.
func (m *MockComponent) Name() string {
	return m.name
}

// MockModal implements nav.Modal for testing.
type MockModal struct {
	*MockComponent
	behavior        components.ModalBehavior
	onDismissReturn bool
	DismissCalled   bool
}

// NewMockModal creates a new MockModal with the given name.
func NewMockModal(name string) *MockModal {
	return &MockModal{
		MockComponent: NewMockComponent(name),
		behavior: components.ModalBehavior{
			CapturesAllInput:      true,
			DismissOnEsc:          true,
			RestoreFocusOnDismiss: true,
		},
		onDismissReturn: true,
	}
}

// ModalBehavior implements nav.Modal.
func (m *MockModal) ModalBehavior() components.ModalBehavior {
	return m.behavior
}

// OnDismiss implements nav.Modal.
func (m *MockModal) OnDismiss() bool {
	m.DismissCalled = true
	return m.onDismissReturn
}

// SetBehavior sets the modal behavior.
func (m *MockModal) SetBehavior(b components.ModalBehavior) *MockModal {
	m.behavior = b
	return m
}

// SetOnDismissReturn sets the return value for OnDismiss.
func (m *MockModal) SetOnDismissReturn(v bool) *MockModal {
	m.onDismissReturn = v
	return m
}

// SetBlockUntilDismissed sets the BlockUntilDismissed behavior.
func (m *MockModal) SetBlockUntilDismissed(v bool) *MockModal {
	m.behavior.BlockUntilDismissed = v
	return m
}

// Reset clears the mock state.
func (m *MockModal) Reset() {
	m.MockComponent.Reset()
	m.DismissCalled = false
}
