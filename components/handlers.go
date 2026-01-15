package components

import "sync"

// Handler types for each event kind
type (
	// ChangeHandler handles value changes
	ChangeHandler[T any] func(event *ChangeEvent[T])

	// SubmitHandler handles submissions
	SubmitHandler func(event *SubmitEvent)

	// CancelHandler handles cancellations
	CancelHandler func(event *CancelEvent)

	// FocusHandler handles focus changes
	FocusHandler func(event *FocusEvent)

	// SelectHandler handles selections
	SelectHandler[T any] func(event *SelectEvent[T])

	// ActivateHandler handles activations
	ActivateHandler func(event *ActivateEvent)

	// KeyHandler handles key events
	KeyHandler func(event *KeyEvent)

	// GenericHandler handles any event
	GenericHandler func(event Event)
)

// EventEmitter interface for components that emit events
type EventEmitter interface {
	// OnEvent registers a handler for all events
	OnEvent(handler GenericHandler)

	// EmitEvent dispatches an event to registered handlers
	EmitEvent(event Event)
}

// BaseEventEmitter provides thread-safe event emission
type BaseEventEmitter struct {
	mu       sync.RWMutex
	handlers []GenericHandler
}

// OnEvent registers a generic event handler
func (e *BaseEventEmitter) OnEvent(handler GenericHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers = append(e.handlers, handler)
}

// EmitEvent dispatches to all registered handlers (thread-safe)
func (e *BaseEventEmitter) EmitEvent(event Event) {
	e.mu.RLock()
	handlers := make([]GenericHandler, len(e.handlers))
	copy(handlers, e.handlers)
	e.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

// RemoveAllHandlers clears all event handlers
func (e *BaseEventEmitter) RemoveAllHandlers() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers = nil
}

// HandlerCount returns the number of registered handlers
func (e *BaseEventEmitter) HandlerCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.handlers)
}
