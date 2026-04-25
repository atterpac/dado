package binding

import (
	"reflect"
	"sync"

	"github.com/atterpac/jig/theme"
)

// valueEqual compares two T values using reflect.DeepEqual.
// Used to skip redundant redraws when a value is set to its current value.
func valueEqual[T any](a, b T) bool {
	return reflect.DeepEqual(a, b)
}

// Value is an observable wrapper for a single value.
// Changes can be subscribed to and automatically trigger UI updates.
type Value[T any] struct {
	value     T
	listeners []func(old, new T)
	mu        sync.RWMutex
}

// NewValue creates a new observable value with an initial value.
func NewValue[T any](initial T) *Value[T] {
	return &Value[T]{
		value: initial,
	}
}

// Get returns the current value.
func (v *Value[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.value
}

// Set updates the value and notifies all listeners.
func (v *Value[T]) Set(newVal T) {
	v.mu.Lock()
	old := v.value
	v.value = newVal
	listeners := v.listeners
	v.mu.Unlock()

	for _, fn := range listeners {
		if fn != nil {
			fn(old, newVal)
		}
	}
}

// SetAndDraw updates the value and triggers a UI redraw.
// Skips the redraw when newVal equals the current value (no observable change).
func (v *Value[T]) SetAndDraw(newVal T) {
	v.mu.RLock()
	unchanged := valueEqual(v.value, newVal)
	v.mu.RUnlock()
	v.Set(newVal)
	if unchanged {
		return
	}
	theme.QueueUpdateDraw(func() {})
}

// Subscribe adds a listener for value changes.
// Returns an unsubscribe function to remove the listener.
//
// Example:
//
//	unsubscribe := value.Subscribe(func(old, new string) {
//	    log.Printf("Value changed from %s to %s", old, new)
//	})
//	defer unsubscribe()
func (v *Value[T]) Subscribe(fn func(old, new T)) func() {
	v.mu.Lock()
	v.listeners = append(v.listeners, fn)
	idx := len(v.listeners) - 1
	v.mu.Unlock()

	return func() {
		v.mu.Lock()
		defer v.mu.Unlock()
		// Mark as nil rather than shifting (avoids invalidating indices)
		if idx < len(v.listeners) {
			v.listeners[idx] = nil
		}
	}
}

// Update modifies the value using a function and notifies listeners.
func (v *Value[T]) Update(fn func(T) T) {
	v.mu.Lock()
	old := v.value
	v.value = fn(old)
	newVal := v.value
	listeners := v.listeners
	v.mu.Unlock()

	for _, listener := range listeners {
		if listener != nil {
			listener(old, newVal)
		}
	}
}

// UpdateAndDraw modifies the value and triggers a UI redraw.
// Skips the redraw when the updater returns a value equal to the prior value.
func (v *Value[T]) UpdateAndDraw(fn func(T) T) {
	v.mu.Lock()
	old := v.value
	v.value = fn(old)
	newVal := v.value
	listeners := v.listeners
	v.mu.Unlock()

	for _, listener := range listeners {
		if listener != nil {
			listener(old, newVal)
		}
	}

	if valueEqual(old, newVal) {
		return
	}
	theme.QueueUpdateDraw(func() {})
}

// BindTo creates a one-way binding to a setter function.
// The setter is called immediately with the current value,
// and again whenever the value changes.
// Returns an unsubscribe function.
//
// Example:
//
//	unsubscribe := connectionStatus.BindTo(func(status string) {
//	    statusText.SetText(status)
//	})
func (v *Value[T]) BindTo(setter func(T)) func() {
	// Set initial value
	setter(v.Get())

	// Subscribe to changes
	return v.Subscribe(func(_, new T) {
		theme.QueueUpdate(func() {
			setter(new)
		})
	})
}

// BindToWithDraw creates a one-way binding that triggers redraws.
func (v *Value[T]) BindToWithDraw(setter func(T)) func() {
	// Set initial value
	setter(v.Get())

	// Subscribe to changes
	return v.Subscribe(func(_, new T) {
		theme.QueueUpdateDraw(func() {
			setter(new)
		})
	})
}

// ListenerCount returns the number of active listeners.
func (v *Value[T]) ListenerCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()

	count := 0
	for _, l := range v.listeners {
		if l != nil {
			count++
		}
	}
	return count
}

// Computed creates a derived value that updates when this value changes.
// The compute function transforms the source value.
//
// Example:
//
//	count := binding.NewValue(0)
//	doubled := count.Computed(func(n int) int { return n * 2 })
func (v *Value[T]) Computed(compute func(T) T) *Value[T] {
	result := NewValue(compute(v.Get()))

	v.Subscribe(func(_, new T) {
		result.Set(compute(new))
	})

	return result
}

// ComputedTo creates a derived value of a different type.
//
// Example:
//
//	count := binding.NewValue(42)
//	text := binding.ComputedTo(count, func(n int) string {
//	    return fmt.Sprintf("Count: %d", n)
//	})
func ComputedTo[T, U any](source *Value[T], compute func(T) U) *Value[U] {
	result := NewValue(compute(source.Get()))

	source.Subscribe(func(_, new T) {
		result.Set(compute(new))
	})

	return result
}
