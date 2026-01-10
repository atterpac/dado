package components

import (
	"sync/atomic"

	"github.com/gdamore/tcell/v2"
)

// eventIDCounter generates unique event IDs
var eventIDCounter uint64

func nextEventID() uint64 {
	return atomic.AddUint64(&eventIDCounter, 1)
}

// EventType represents the type of component event
type EventType string

const (
	EventChange   EventType = "change"
	EventSubmit   EventType = "submit"
	EventCancel   EventType = "cancel"
	EventFocus    EventType = "focus"
	EventBlur     EventType = "blur"
	EventSelect   EventType = "select"
	EventActivate EventType = "activate"
	EventKey      EventType = "key"
)

// Event is the base event interface all events implement
type Event interface {
	Type() EventType
	Source() string // Human-readable component name
	ID() uint64     // Unique event ID for lookup
}

// BaseEvent provides common event fields
type BaseEvent struct {
	eventType EventType
	source    string
	id        uint64
}

// NewBaseEvent creates a new base event with the given type and source
func NewBaseEvent(eventType EventType, source string) BaseEvent {
	return BaseEvent{
		eventType: eventType,
		source:    source,
		id:        nextEventID(),
	}
}

func (e *BaseEvent) Type() EventType { return e.eventType }
func (e *BaseEvent) Source() string  { return e.source }
func (e *BaseEvent) ID() uint64      { return e.id }

// ChangeEvent is emitted when a component's value changes
type ChangeEvent[T any] struct {
	BaseEvent
	OldValue T
	NewValue T
	Index    int // -1 if not applicable
}

// NewChangeEvent creates a typed change event
func NewChangeEvent[T any](source string, old, new T) *ChangeEvent[T] {
	return &ChangeEvent[T]{
		BaseEvent: NewBaseEvent(EventChange, source),
		OldValue:  old,
		NewValue:  new,
		Index:     -1,
	}
}

// WithIndex sets the index for list-based components
func (e *ChangeEvent[T]) WithIndex(idx int) *ChangeEvent[T] {
	e.Index = idx
	return e
}

// SubmitEvent is emitted when a form or input is submitted
type SubmitEvent struct {
	BaseEvent
	Values map[string]any // All form values (for forms)
	Value  any            // Single value (for inputs)
}

// NewSubmitEvent creates a submit event
func NewSubmitEvent(source string, value any) *SubmitEvent {
	return &SubmitEvent{
		BaseEvent: NewBaseEvent(EventSubmit, source),
		Value:     value,
	}
}

// WithFormValues attaches form values to the submit event
func (e *SubmitEvent) WithFormValues(values map[string]any) *SubmitEvent {
	e.Values = values
	return e
}

// CancelEvent is emitted when an operation is cancelled
type CancelEvent struct {
	BaseEvent
}

// NewCancelEvent creates a cancel event
func NewCancelEvent(source string) *CancelEvent {
	return &CancelEvent{
		BaseEvent: NewBaseEvent(EventCancel, source),
	}
}

// FocusEvent is emitted when focus changes
type FocusEvent struct {
	BaseEvent
	Focused bool
}

// NewFocusEvent creates a focus event
func NewFocusEvent(source string, focused bool) *FocusEvent {
	eventType := EventFocus
	if !focused {
		eventType = EventBlur
	}
	return &FocusEvent{
		BaseEvent: NewBaseEvent(eventType, source),
		Focused:   focused,
	}
}

// SelectEvent is emitted when an item is selected (tables, lists, trees)
type SelectEvent[T any] struct {
	BaseEvent
	Index int
	Item  T
}

// NewSelectEvent creates a selection event
func NewSelectEvent[T any](source string, index int, item T) *SelectEvent[T] {
	return &SelectEvent[T]{
		BaseEvent: NewBaseEvent(EventSelect, source),
		Index:     index,
		Item:      item,
	}
}

// ActivateEvent is emitted when a component is activated (button press, menu item)
type ActivateEvent struct {
	BaseEvent
}

// NewActivateEvent creates an activate event
func NewActivateEvent(source string) *ActivateEvent {
	return &ActivateEvent{
		BaseEvent: NewBaseEvent(EventActivate, source),
	}
}

// KeyEvent wraps tcell key events with component context
type KeyEvent struct {
	BaseEvent
	Key       tcell.Key
	Rune      rune
	Modifiers tcell.ModMask
}

// NewKeyEvent creates a key event
func NewKeyEvent(source string, event *tcell.EventKey) *KeyEvent {
	return &KeyEvent{
		BaseEvent: NewBaseEvent(EventKey, source),
		Key:       event.Key(),
		Rune:      event.Rune(),
		Modifiers: event.Modifiers(),
	}
}
