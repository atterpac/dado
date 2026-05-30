// Package effect provides an opt-in command/effect layer inspired by
// Bubble Tea's tea.Cmd model.
//
// An Effect is a function that performs work (typically async / side-effectful)
// and returns a Msg describing the result. A Dispatcher launches Effects on
// goroutines and delivers their resulting Msgs to subscribers on the UI thread.
//
// The canonical pattern is:
//
//	disp := effect.NewDispatcher()
//
//	// Subscribe a typed handler.
//	unsub := effect.On(disp, func(m UsersLoaded) {
//	    usersValue.SetAndDraw(m.Users)
//	})
//	defer unsub()
//
//	// Run an Effect that emits a Msg.
//	disp.Run(func(ctx context.Context) effect.Msg {
//	    users, _ := api.ListUsers(ctx)
//	    return UsersLoaded{Users: users}
//	})
//
// All handlers run on the UI thread via theme.QueueUpdateDraw, so handler
// code may freely mutate UI state without QueueUpdate.
//
// Cancellation: each Dispatcher owns a context derived from a parent. Calling
// Shutdown cancels in-flight Effects and waits for goroutines to drain.
package effect

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atterpac/dado/bus"
	"github.com/atterpac/dado/theme"
)

// Msg is the unit of communication from an Effect to its subscribers.
// Concrete Msg types are user-defined; subscribers filter by type via On.
type Msg any

// Effect performs work and returns a Msg. A nil Msg signals "no result";
// the dispatcher will not fan out a nil Msg.
type Effect func(ctx context.Context) Msg

// Handler receives Msgs delivered by a Dispatcher.
type Handler func(Msg)

// Dispatcher launches Effects, delivers resulting Msgs to subscribers, and
// owns the lifecycle (context cancellation + goroutine draining).
type Dispatcher struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu     sync.RWMutex
	subs   []subscription
	nextID uint64

	closed atomic.Bool
}

type subscription struct {
	id      uint64
	matcher func(Msg) bool
	handler Handler
}

// NewDispatcher creates a Dispatcher with a background-derived context.
// Use NewDispatcherWithContext to supply a parent context.
func NewDispatcher() *Dispatcher {
	return NewDispatcherWithContext(context.Background())
}

// NewDispatcherWithContext creates a Dispatcher whose Effects derive their
// context from parent. Cancelling parent cancels in-flight Effects.
func NewDispatcherWithContext(parent context.Context) *Dispatcher {
	ctx, cancel := context.WithCancel(parent)
	return &Dispatcher{ctx: ctx, cancel: cancel}
}

// Run launches the Effect on a goroutine. The resulting Msg is delivered on
// the UI thread via theme.QueueUpdateDraw. Nil and post-Shutdown Effects
// are dropped silently.
func (d *Dispatcher) Run(e Effect) {
	if e == nil || d.closed.Load() {
		return
	}
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		msg := e(d.ctx)
		if msg == nil || d.closed.Load() {
			return
		}
		theme.QueueUpdateDraw(func() {
			d.deliver(msg)
		})
	}()
}

// Subscribe registers a handler for Msgs matching matcher. A nil matcher
// receives all Msgs. Returns an unsubscribe function.
//
// Pair with components.Subscriptions for automatic teardown:
//
//	cb.Subs().Add(disp.Subscribe(matcher, handler))
func (d *Dispatcher) Subscribe(matcher func(Msg) bool, h Handler) func() {
	d.mu.Lock()
	id := d.nextID
	d.nextID++
	d.subs = append(d.subs, subscription{id: id, matcher: matcher, handler: h})
	d.mu.Unlock()

	return func() {
		d.mu.Lock()
		defer d.mu.Unlock()
		for i, s := range d.subs {
			if s.id == id {
				d.subs = append(d.subs[:i], d.subs[i+1:]...)
				return
			}
		}
	}
}

// Shutdown cancels in-flight Effects and waits for goroutines to drain.
// After Shutdown, Run is a no-op. Safe to call multiple times.
//
// If ctx expires before goroutines drain, Shutdown returns ctx.Err().
func (d *Dispatcher) Shutdown(ctx context.Context) error {
	if d.closed.Swap(true) {
		return nil
	}
	d.cancel()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// deliver fans out msg to subscribers. Recognizes batchMsg / sequenceMsg as
// internal control messages and expands them instead of fanning out.
//
// Must be called on the UI thread (Run uses theme.QueueUpdateDraw).
func (d *Dispatcher) deliver(msg Msg) {
	switch m := msg.(type) {
	case batchMsg:
		for _, e := range m.cmds {
			d.Run(e)
		}
		return
	case sequenceMsg:
		d.runSequence(m.cmds)
		return
	}

	d.mu.RLock()
	subs := make([]subscription, len(d.subs))
	copy(subs, d.subs)
	d.mu.RUnlock()

	for _, s := range subs {
		if s.matcher != nil && !s.matcher(msg) {
			continue
		}
		s.handler(msg)
	}

	if bus.Enabled() {
		bus.Publish(bus.Event{
			Kind:    bus.KindEffectMsg,
			Source:  bus.SourceEffect,
			Payload: bus.EffectMsg{Msg: msg},
		})
	}
}

// On is a generic helper that subscribes to Msgs of concrete type T.
// The handler receives the typed value directly.
//
//	unsub := effect.On(disp, func(m UsersLoaded) { ... })
func On[T any](d *Dispatcher, h func(T)) func() {
	return d.Subscribe(
		func(m Msg) bool { _, ok := m.(T); return ok },
		func(m Msg) { h(m.(T)) },
	)
}

// --- Combinators ---

// batchMsg is an internal sentinel that the Dispatcher recognizes to fan out
// child Effects in parallel.
type batchMsg struct{ cmds []Effect }

// sequenceMsg is an internal sentinel that the Dispatcher recognizes to run
// child Effects in order, awaiting each Msg's delivery before the next.
type sequenceMsg struct{ cmds []Effect }

// Batch returns an Effect that runs all child Effects in parallel. Each
// child's Msg is delivered to subscribers as it arrives (no aggregation).
// Nil children are skipped.
func Batch(cmds ...Effect) Effect {
	filtered := nonNil(cmds)
	if len(filtered) == 0 {
		return None()
	}
	return func(_ context.Context) Msg {
		return batchMsg{cmds: filtered}
	}
}

// Sequence returns an Effect that runs child Effects serially: each next
// Effect starts only after the previous Effect's Msg has been delivered.
// Nil children are skipped.
func Sequence(cmds ...Effect) Effect {
	filtered := nonNil(cmds)
	if len(filtered) == 0 {
		return None()
	}
	return func(_ context.Context) Msg {
		return sequenceMsg{cmds: filtered}
	}
}

// Tick returns a one-shot Effect that emits fn(now) after duration d.
// Returns nil if the dispatcher's context is cancelled before d elapses.
// For recurring timers, re-arm by returning a new Tick from the handler.
func Tick(dur time.Duration, fn func(time.Time) Msg) Effect {
	return func(ctx context.Context) Msg {
		select {
		case t := <-time.After(dur):
			return fn(t)
		case <-ctx.Done():
			return nil
		}
	}
}

// None returns an Effect that emits no Msg. Useful as a placeholder.
func None() Effect {
	return func(_ context.Context) Msg { return nil }
}

func nonNil(cmds []Effect) []Effect {
	out := cmds[:0:0]
	for _, c := range cmds {
		if c != nil {
			out = append(out, c)
		}
	}
	return out
}

// runSequence is the serial driver for sequenceMsg. Runs on a goroutine to
// avoid blocking the UI thread while awaiting delivery between steps.
func (d *Dispatcher) runSequence(cmds []Effect) {
	if d.closed.Load() {
		return
	}
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for _, e := range cmds {
			if d.closed.Load() {
				return
			}
			msg := e(d.ctx)
			if msg == nil {
				continue
			}
			done := make(chan struct{})
			theme.QueueUpdateDraw(func() {
				d.deliver(msg)
				close(done)
			})
			select {
			case <-done:
			case <-d.ctx.Done():
				return
			}
		}
	}()
}
