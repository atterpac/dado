// Package async provides helpers for async data loading with automatic UI updates.
//
// The Loader type uses a builder pattern to configure async operations with:
//   - Type-safe success/error callbacks via generics
//   - Configurable timeout and context
//   - Pluggable loading indicators (toast, progress modal, status bar, or custom)
//   - Automatic QueueUpdateDraw for thread-safe UI updates
//
// Basic usage:
//
//	async.NewLoader[[]Workflow]().
//	    WithTimeout(10 * time.Second).
//	    OnSuccess(func(data []Workflow) { model.data = data }).
//	    OnError(func(err error) { showError(err) }).
//	    Run(func(ctx context.Context) ([]Workflow, error) {
//	        return provider.ListWorkflows(ctx)
//	    })
//
// With loading indicator:
//
//	async.NewLoader[[]Workflow]().
//	    WithIndicator(async.Toast("Loading workflows...")).
//	    OnSuccess(func(data []Workflow) { model.data = data }).
//	    Run(fetchWorkflows)
package async

import (
	"context"
	"sync"
	"time"

	"github.com/atterpac/jig/bus"
	"github.com/atterpac/jig/theme"
)

// publishLoader emits a loader lifecycle event when the bus is enabled.
func publishLoader(kind, phase string, err error) {
	if !bus.Enabled() {
		return
	}
	bus.Publish(bus.Event{
		Kind:    kind,
		Source:  bus.SourceAsync,
		Payload: bus.LoaderState{Phase: phase, Err: err},
	})
}

// LoadFunc is the function signature for async operations.
type LoadFunc[T any] func(ctx context.Context) (T, error)

// Loader is a builder for configuring and running async operations.
type Loader[T any] struct {
	timeout   time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
	indicator LoadingIndicator

	onSuccess func(T)
	onError   func(error)
	onFinally func()

	// State
	running bool
	mu      sync.Mutex
}

// NewLoader creates a new async loader builder.
func NewLoader[T any]() *Loader[T] {
	return &Loader[T]{
		timeout: 30 * time.Second, // Default timeout
	}
}

// WithTimeout sets the operation timeout.
// Default is 30 seconds. Use 0 for no timeout.
func (l *Loader[T]) WithTimeout(d time.Duration) *Loader[T] {
	l.timeout = d
	return l
}

// WithContext sets a custom context for the operation.
// If not set, a background context with timeout is used.
func (l *Loader[T]) WithContext(ctx context.Context) *Loader[T] {
	l.ctx = ctx
	return l
}

// WithIndicator sets the loading indicator to show during the operation.
// See Toast(), ProgressModal(), StatusBar(), or implement LoadingIndicator.
func (l *Loader[T]) WithIndicator(indicator LoadingIndicator) *Loader[T] {
	l.indicator = indicator
	return l
}

// OnSuccess sets the callback for successful completion.
// This is called on the UI thread via QueueUpdateDraw.
func (l *Loader[T]) OnSuccess(fn func(T)) *Loader[T] {
	l.onSuccess = fn
	return l
}

// OnError sets the callback for errors.
// This is called on the UI thread via QueueUpdateDraw.
func (l *Loader[T]) OnError(fn func(error)) *Loader[T] {
	l.onError = fn
	return l
}

// OnFinally sets a callback that runs after success or error.
// This is called on the UI thread via QueueUpdateDraw.
func (l *Loader[T]) OnFinally(fn func()) *Loader[T] {
	l.onFinally = fn
	return l
}

// Run executes the async operation.
// Returns the Loader for method chaining or later cancellation.
func (l *Loader[T]) Run(fn LoadFunc[T]) *Loader[T] {
	l.mu.Lock()
	if l.running {
		l.mu.Unlock()
		return l
	}
	l.running = true
	l.mu.Unlock()

	// Set up context
	ctx := l.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	if l.timeout > 0 {
		ctx, l.cancel = context.WithTimeout(ctx, l.timeout)
	} else {
		ctx, l.cancel = context.WithCancel(ctx)
	}

	// Capture indicator reference for goroutine
	indicator := l.indicator

	// Show indicator on UI thread to avoid race conditions
	if indicator != nil {
		theme.QueueUpdateDraw(func() {
			indicator.Show()
		})
	}

	publishLoader(bus.KindLoaderStart, "start", nil)

	// Run async
	go func() {
		defer func() {
			if l.cancel != nil {
				l.cancel()
			}
		}()

		result, err := fn(ctx)

		// Update UI on main thread
		theme.QueueUpdateDraw(func() {
			// Hide indicator
			if indicator != nil {
				if err != nil {
					indicator.Error(err)
				} else {
					indicator.Success()
				}
				indicator.Hide()
			}

			// Call callbacks
			if err != nil {
				if l.onError != nil {
					l.onError(err)
				}
				publishLoader(bus.KindLoaderError, "error", err)
			} else {
				if l.onSuccess != nil {
					l.onSuccess(result)
				}
				publishLoader(bus.KindLoaderSuccess, "success", nil)
			}

			if l.onFinally != nil {
				l.onFinally()
			}

			l.mu.Lock()
			l.running = false
			l.mu.Unlock()
		})
	}()

	return l
}

// Cancel cancels the running operation.
func (l *Loader[T]) Cancel() {
	if l.cancel != nil {
		l.cancel()
		publishLoader(bus.KindLoaderCancel, "cancel", nil)
	}
}

// IsRunning returns true if the operation is currently executing.
func (l *Loader[T]) IsRunning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}

// Load is a convenience function for simple one-shot async operations.
// For more control, use NewLoader[T]() builder.
//
// Example:
//
//	async.Load(
//	    func(ctx context.Context) ([]Workflow, error) {
//	        return provider.ListWorkflows(ctx)
//	    },
//	    func(data []Workflow) { model.data = data },
//	    func(err error) { showError(err) },
//	)
func Load[T any](fn LoadFunc[T], onSuccess func(T), onError func(error)) *Loader[T] {
	return NewLoader[T]().
		OnSuccess(onSuccess).
		OnError(onError).
		Run(fn)
}

// LoadWithIndicator is a convenience function with a loading message.
//
// Example:
//
//	async.LoadWithIndicator(
//	    "Loading workflows...",
//	    func(ctx context.Context) ([]Workflow, error) {
//	        return provider.ListWorkflows(ctx)
//	    },
//	    func(data []Workflow) { model.data = data },
//	    func(err error) { showError(err) },
//	)
func LoadWithIndicator[T any](message string, fn LoadFunc[T], onSuccess func(T), onError func(error)) *Loader[T] {
	return NewLoader[T]().
		WithIndicator(Toast(message)).
		OnSuccess(onSuccess).
		OnError(onError).
		Run(fn)
}
