package components

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// LoadState represents the loading state of a component.
type LoadState int

const (
	// LoadStateIdle means no data operation is in progress.
	LoadStateIdle LoadState = iota
	// LoadStateLoading means data is being fetched/processed.
	LoadStateLoading
	// LoadStateError means the last operation failed.
	LoadStateError
	// LoadStateSuccess means data was loaded successfully.
	LoadStateSuccess
)

// StatefulComponentBase wraps a tview.Primitive and provides state management.
// Use this for components that need to manage data state (loading, error, success).
//
// Example:
//
//	type UserList struct {
//	    base  *StatefulComponentBase[[]User]
//	    table *Table
//	}
//
//	func NewUserList() *UserList {
//	    table := NewTable()
//	    u := &UserList{table: table}
//	    u.base = NewStatefulComponentBase[[]User](table).
//	        SetName("user-list").
//	        SetOnStateChange(u.render)
//	    return u
//	}
//
//	func (u *UserList) LoadUsers() {
//	    u.base.SetLoadState(LoadStateLoading)
//	    go func() {
//	        users, err := fetchUsers()
//	        if err != nil {
//	            u.base.SetError(err)
//	            return
//	        }
//	        u.base.SetData(users)
//	    }()
//	}
type StatefulComponentBase[T any] struct {
	*ComponentBase
	mu sync.RWMutex

	data      T
	loadState LoadState
	err       error

	// Callbacks
	onStateChange func(state LoadState, data T, err error)
	onDataChange  func(data T)
}

// NewStatefulComponentBase creates a new stateful component base wrapping the given primitive.
func NewStatefulComponentBase[T any](p tview.Primitive) *StatefulComponentBase[T] {
	return &StatefulComponentBase[T]{
		ComponentBase: NewComponentBase(p),
		loadState:     LoadStateIdle,
	}
}

// --- Configuration Methods (Fluent API) ---

// SetName sets the component name (used for debugging and events).
func (scb *StatefulComponentBase[T]) SetName(name string) *StatefulComponentBase[T] {
	scb.ComponentBase.SetName(name)
	return scb
}

// SetHints sets the key hints displayed in the menu bar.
func (scb *StatefulComponentBase[T]) SetHints(hints []KeyHint) *StatefulComponentBase[T] {
	scb.ComponentBase.SetHints(hints)
	return scb
}

// AddHint adds a single key hint.
func (scb *StatefulComponentBase[T]) AddHint(key, description string) *StatefulComponentBase[T] {
	scb.ComponentBase.AddHint(key, description)
	return scb
}

// SetOnStart sets the callback invoked when component becomes active.
func (scb *StatefulComponentBase[T]) SetOnStart(fn func()) *StatefulComponentBase[T] {
	scb.ComponentBase.SetOnStart(fn)
	return scb
}

// SetOnStop sets the callback invoked when component becomes inactive.
func (scb *StatefulComponentBase[T]) SetOnStop(fn func()) *StatefulComponentBase[T] {
	scb.ComponentBase.SetOnStop(fn)
	return scb
}

// SetInputHandler sets a custom input handler.
func (scb *StatefulComponentBase[T]) SetInputHandler(fn func(*tcell.EventKey, func(tview.Primitive)) *tcell.EventKey) *StatefulComponentBase[T] {
	scb.ComponentBase.SetInputHandler(fn)
	return scb
}

// SetDrawOverlay sets a function called after the primitive draws.
func (scb *StatefulComponentBase[T]) SetDrawOverlay(fn func(screen tcell.Screen)) *StatefulComponentBase[T] {
	scb.ComponentBase.SetDrawOverlay(fn)
	return scb
}

// SetOnStateChange sets a callback for when the load state changes.
func (scb *StatefulComponentBase[T]) SetOnStateChange(fn func(state LoadState, data T, err error)) *StatefulComponentBase[T] {
	scb.mu.Lock()
	defer scb.mu.Unlock()
	scb.onStateChange = fn
	return scb
}

// SetOnDataChange sets a callback for when data changes (regardless of load state).
func (scb *StatefulComponentBase[T]) SetOnDataChange(fn func(data T)) *StatefulComponentBase[T] {
	scb.mu.Lock()
	defer scb.mu.Unlock()
	scb.onDataChange = fn
	return scb
}

// --- State Management ---

// SetData sets the component data and updates state to Success.
func (scb *StatefulComponentBase[T]) SetData(data T) *StatefulComponentBase[T] {
	scb.mu.Lock()
	scb.data = data
	scb.loadState = LoadStateSuccess
	scb.err = nil
	onStateChange := scb.onStateChange
	onDataChange := scb.onDataChange
	scb.mu.Unlock()

	if onDataChange != nil {
		onDataChange(data)
	}
	if onStateChange != nil {
		onStateChange(LoadStateSuccess, data, nil)
	}
	return scb
}

// Data returns the current data.
func (scb *StatefulComponentBase[T]) Data() T {
	scb.mu.RLock()
	defer scb.mu.RUnlock()
	return scb.data
}

// SetLoadState sets the loading state.
func (scb *StatefulComponentBase[T]) SetLoadState(state LoadState) *StatefulComponentBase[T] {
	scb.mu.Lock()
	scb.loadState = state
	if state != LoadStateError {
		scb.err = nil
	}
	onStateChange := scb.onStateChange
	data := scb.data
	err := scb.err
	scb.mu.Unlock()

	if onStateChange != nil {
		onStateChange(state, data, err)
	}
	return scb
}

// LoadState returns the current load state.
func (scb *StatefulComponentBase[T]) LoadState() LoadState {
	scb.mu.RLock()
	defer scb.mu.RUnlock()
	return scb.loadState
}

// SetError sets an error and updates state to Error.
func (scb *StatefulComponentBase[T]) SetError(err error) *StatefulComponentBase[T] {
	scb.mu.Lock()
	scb.err = err
	scb.loadState = LoadStateError
	onStateChange := scb.onStateChange
	data := scb.data
	scb.mu.Unlock()

	if onStateChange != nil {
		onStateChange(LoadStateError, data, err)
	}
	return scb
}

// Error returns the current error, if any.
func (scb *StatefulComponentBase[T]) Error() error {
	scb.mu.RLock()
	defer scb.mu.RUnlock()
	return scb.err
}

// IsLoading returns true if the component is currently loading.
func (scb *StatefulComponentBase[T]) IsLoading() bool {
	scb.mu.RLock()
	defer scb.mu.RUnlock()
	return scb.loadState == LoadStateLoading
}

// HasError returns true if the component has an error.
func (scb *StatefulComponentBase[T]) HasError() bool {
	scb.mu.RLock()
	defer scb.mu.RUnlock()
	return scb.loadState == LoadStateError
}

// IsReady returns true if data was loaded successfully.
func (scb *StatefulComponentBase[T]) IsReady() bool {
	scb.mu.RLock()
	defer scb.mu.RUnlock()
	return scb.loadState == LoadStateSuccess
}

// Reset clears the data and resets to idle state.
func (scb *StatefulComponentBase[T]) Reset() *StatefulComponentBase[T] {
	scb.mu.Lock()
	var zero T
	scb.data = zero
	scb.loadState = LoadStateIdle
	scb.err = nil
	onStateChange := scb.onStateChange
	scb.mu.Unlock()

	if onStateChange != nil {
		onStateChange(LoadStateIdle, zero, nil)
	}
	return scb
}

// --- Convenience Methods ---

// UpdateData updates the data using a function, maintaining thread safety.
func (scb *StatefulComponentBase[T]) UpdateData(fn func(T) T) *StatefulComponentBase[T] {
	scb.mu.Lock()
	scb.data = fn(scb.data)
	data := scb.data
	onDataChange := scb.onDataChange
	onStateChange := scb.onStateChange
	scb.mu.Unlock()

	if onDataChange != nil {
		onDataChange(data)
	}
	if onStateChange != nil && scb.loadState == LoadStateSuccess {
		onStateChange(LoadStateSuccess, data, nil)
	}
	return scb
}
