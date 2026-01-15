package components

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChangeEvent_Creation tests ChangeEvent creation and properties.
func TestChangeEvent_Creation(t *testing.T) {
	event := NewChangeEvent("field-name", "old", "new")

	assert.Equal(t, EventChange, event.Type())
	assert.Equal(t, "field-name", event.Source())
	assert.Equal(t, "old", event.OldValue)
	assert.Equal(t, "new", event.NewValue)
	assert.Equal(t, -1, event.Index)
	assert.NotZero(t, event.ID())
}

// TestChangeEvent_WithIndex tests ChangeEvent index chaining.
func TestChangeEvent_WithIndex(t *testing.T) {
	event := NewChangeEvent("list", "", "item").WithIndex(5)

	assert.Equal(t, 5, event.Index)
}

// TestSubmitEvent_Creation tests SubmitEvent creation.
func TestSubmitEvent_Creation(t *testing.T) {
	event := NewSubmitEvent("form", "submitted-value")

	assert.Equal(t, EventSubmit, event.Type())
	assert.Equal(t, "form", event.Source())
	assert.Equal(t, "submitted-value", event.Value)
	assert.Nil(t, event.Values)
}

// TestSubmitEvent_WithFormValues tests SubmitEvent with form values.
func TestSubmitEvent_WithFormValues(t *testing.T) {
	values := map[string]any{
		"name":  "John",
		"email": "john@example.com",
	}
	event := NewSubmitEvent("form", nil).WithFormValues(values)

	assert.Equal(t, values, event.Values)
}

// TestCancelEvent_Creation tests CancelEvent creation.
func TestCancelEvent_Creation(t *testing.T) {
	event := NewCancelEvent("modal")

	assert.Equal(t, EventCancel, event.Type())
	assert.Equal(t, "modal", event.Source())
}

// TestFocusEvent_Creation tests FocusEvent creation.
func TestFocusEvent_Creation(t *testing.T) {
	focusEvent := NewFocusEvent("input", true)
	assert.Equal(t, EventFocus, focusEvent.Type())
	assert.True(t, focusEvent.Focused)

	blurEvent := NewFocusEvent("input", false)
	assert.Equal(t, EventBlur, blurEvent.Type())
	assert.False(t, blurEvent.Focused)
}

// TestSelectEvent_Creation tests SelectEvent creation.
func TestSelectEvent_Creation(t *testing.T) {
	event := NewSelectEvent("list", 3, "item-data")

	assert.Equal(t, EventSelect, event.Type())
	assert.Equal(t, "list", event.Source())
	assert.Equal(t, 3, event.Index)
	assert.Equal(t, "item-data", event.Item)
}

// TestActivateEvent_Creation tests ActivateEvent creation.
func TestActivateEvent_Creation(t *testing.T) {
	event := NewActivateEvent("button")

	assert.Equal(t, EventActivate, event.Type())
	assert.Equal(t, "button", event.Source())
}

// TestKeyEvent_Creation tests KeyEvent creation.
func TestKeyEvent_Creation(t *testing.T) {
	tcellEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModCtrl)
	event := NewKeyEvent("form", tcellEvent)

	assert.Equal(t, EventKey, event.Type())
	assert.Equal(t, "form", event.Source())
	assert.Equal(t, tcell.KeyEnter, event.Key)
	assert.Equal(t, tcell.ModCtrl, event.Modifiers)
}

// TestEventID_Uniqueness tests that event IDs are unique.
func TestEventID_Uniqueness(t *testing.T) {
	ids := make(map[uint64]bool)
	const count = 1000

	for i := 0; i < count; i++ {
		event := NewChangeEvent("test", "", "")
		if ids[event.ID()] {
			t.Fatalf("duplicate event ID: %d at iteration %d", event.ID(), i)
		}
		ids[event.ID()] = true
	}

	assert.Len(t, ids, count)
}

// TestTextField_TypedChangeHandler tests TextField typed change handler.
func TestTextField_TypedChangeHandler(t *testing.T) {
	field := NewTextField("test").SetValue("initial")

	var received *ChangeEvent[string]
	field.SetOnChange(func(e *ChangeEvent[string]) {
		received = e
	})

	// Simulate typing by calling the input handler
	handler := field.InputHandler()
	require.NotNil(t, handler)

	// Type 'x' - this should trigger change event
	event := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	handler(event, nil)

	require.NotNil(t, received)
	assert.Equal(t, EventChange, received.Type())
	assert.Equal(t, "test", received.Source())
	assert.Equal(t, "initial", received.OldValue)
	assert.Equal(t, "initialx", received.NewValue)
}

// TestTextField_SubmitHandler tests TextField submit handler.
func TestTextField_SubmitHandler(t *testing.T) {
	field := NewTextField("test").SetValue("hello")

	var received *SubmitEvent
	field.SetOnSubmit(func(e *SubmitEvent) {
		received = e
	})

	handler := field.InputHandler()
	enterEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	handler(enterEvent, nil)

	require.NotNil(t, received)
	assert.Equal(t, EventSubmit, received.Type())
	assert.Equal(t, "hello", received.Value)
}

// TestCheckbox_TypedChangeHandler tests Checkbox typed change handler.
func TestCheckbox_TypedChangeHandler(t *testing.T) {
	cb := NewCheckbox("test").SetChecked(false)

	var received *ChangeEvent[bool]
	cb.SetOnChange(func(e *ChangeEvent[bool]) {
		received = e
	})

	// Toggle the checkbox
	cb.Toggle()

	require.NotNil(t, received)
	assert.Equal(t, false, received.OldValue)
	assert.Equal(t, true, received.NewValue)
}

// TestSelect_TypedChangeHandler tests Select typed change handler.
func TestSelect_TypedChangeHandler(t *testing.T) {
	sel := NewSelect("test").
		SetOptions([]string{"a", "b", "c"}).
		SetSelected(0)

	var received *ChangeEvent[SelectOption]
	sel.SetOnChange(func(e *ChangeEvent[SelectOption]) {
		received = e
	})

	handler := sel.InputHandler()

	// Expand the dropdown (space opens it)
	handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone), nil)
	// Navigate down to second option
	handler(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), nil)
	// Close dropdown with space (this commits selection and emits change)
	handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone), nil)

	require.NotNil(t, received, "change event should be emitted when dropdown closes")
	assert.Equal(t, 1, received.Index)
	assert.Equal(t, "b", received.NewValue.Value)
}

// TestBaseEventEmitter_OnEvent tests generic event registration.
func TestBaseEventEmitter_OnEvent(t *testing.T) {
	field := NewTextField("test")

	var events []Event
	field.OnEvent(func(e Event) {
		events = append(events, e)
	})

	// Type and submit
	handler := field.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	require.Len(t, events, 2)
	assert.Equal(t, EventChange, events[0].Type())
	assert.Equal(t, EventSubmit, events[1].Type())
}

// TestBaseEventEmitter_MultipleHandlers tests multiple handler registration.
func TestBaseEventEmitter_MultipleHandlers(t *testing.T) {
	emitter := &BaseEventEmitter{}

	var count1, count2 int
	emitter.OnEvent(func(e Event) { count1++ })
	emitter.OnEvent(func(e Event) { count2++ })

	emitter.EmitEvent(NewActivateEvent("test"))

	assert.Equal(t, 1, count1)
	assert.Equal(t, 1, count2)
}

// TestBaseEventEmitter_HandlerCount tests handler count tracking.
func TestBaseEventEmitter_HandlerCount(t *testing.T) {
	emitter := &BaseEventEmitter{}

	assert.Equal(t, 0, emitter.HandlerCount())

	emitter.OnEvent(func(e Event) {})
	assert.Equal(t, 1, emitter.HandlerCount())

	emitter.OnEvent(func(e Event) {})
	assert.Equal(t, 2, emitter.HandlerCount())
}

// TestBaseEventEmitter_RemoveAllHandlers tests handler removal.
func TestBaseEventEmitter_RemoveAllHandlers(t *testing.T) {
	emitter := &BaseEventEmitter{}

	emitter.OnEvent(func(e Event) {})
	emitter.OnEvent(func(e Event) {})
	require.Equal(t, 2, emitter.HandlerCount())

	emitter.RemoveAllHandlers()
	assert.Equal(t, 0, emitter.HandlerCount())

	// Emit should not panic with no handlers
	emitter.EmitEvent(NewActivateEvent("test"))
}

// TestBaseEventEmitter_ThreadSafety tests concurrent access to event emitter.
func TestBaseEventEmitter_ThreadSafety(t *testing.T) {
	emitter := &BaseEventEmitter{}

	var wg sync.WaitGroup
	const goroutines = 50
	const eventsPerGoroutine = 100

	// Add handlers concurrently
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			emitter.OnEvent(func(e Event) {})
		}()
	}
	wg.Wait()

	assert.Equal(t, goroutines, emitter.HandlerCount())

	// Emit events concurrently
	var received int64
	emitter.RemoveAllHandlers()
	emitter.OnEvent(func(e Event) {
		atomic.AddInt64(&received, 1)
	})

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				emitter.EmitEvent(NewActivateEvent("test"))
			}
		}()
	}
	wg.Wait()

	expected := int64(goroutines * eventsPerGoroutine)
	assert.Equal(t, expected, atomic.LoadInt64(&received))
}

// TestBaseEventEmitter_CopyOnEmit tests that handlers are copied before emission.
func TestBaseEventEmitter_CopyOnEmit(t *testing.T) {
	emitter := &BaseEventEmitter{}

	// Handler that adds more handlers during emission
	emitter.OnEvent(func(e Event) {
		emitter.OnEvent(func(e Event) {})
	})

	// This should not cause infinite loop or panic
	emitter.EmitEvent(NewActivateEvent("test"))

	// Should have 2 handlers now (original + one added during emission)
	assert.Equal(t, 2, emitter.HandlerCount())
}

// TestTextField_BackspaceEmitsChange tests that backspace emits change events.
func TestTextField_BackspaceEmitsChange(t *testing.T) {
	field := NewTextField("test").SetValue("hello")

	var events []*ChangeEvent[string]
	field.SetOnChange(func(e *ChangeEvent[string]) {
		events = append(events, e)
	})

	handler := field.InputHandler()

	// Backspace
	handler(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone), nil)

	require.Len(t, events, 1)
	assert.Equal(t, "hello", events[0].OldValue)
	assert.Equal(t, "hell", events[0].NewValue)
}

// TestTextField_DeleteEmitsChange tests that delete emits change events.
func TestTextField_DeleteEmitsChange(t *testing.T) {
	field := NewTextField("test").SetValue("hello")

	var events []*ChangeEvent[string]
	field.SetOnChange(func(e *ChangeEvent[string]) {
		events = append(events, e)
	})

	// Move cursor to beginning first
	handler := field.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone), nil)

	// Delete
	handler(tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone), nil)

	require.Len(t, events, 1)
	assert.Equal(t, "hello", events[0].OldValue)
	assert.Equal(t, "ello", events[0].NewValue)
}

// TestTextField_CtrlU_EmitsChange tests Ctrl+U (delete to beginning) emits change.
func TestTextField_CtrlU_EmitsChange(t *testing.T) {
	field := NewTextField("test").SetValue("hello world")

	var events []*ChangeEvent[string]
	field.SetOnChange(func(e *ChangeEvent[string]) {
		events = append(events, e)
	})

	handler := field.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyCtrlU, 0, tcell.ModNone), nil)

	require.Len(t, events, 1)
	assert.Equal(t, "hello world", events[0].OldValue)
	assert.Equal(t, "", events[0].NewValue)
}

// TestTextField_CtrlK_EmitsChange tests Ctrl+K (delete to end) emits change.
func TestTextField_CtrlK_EmitsChange(t *testing.T) {
	field := NewTextField("test").SetValue("hello world")

	var events []*ChangeEvent[string]
	field.SetOnChange(func(e *ChangeEvent[string]) {
		events = append(events, e)
	})

	// Move to middle
	handler := field.InputHandler()
	for i := 0; i < 5; i++ {
		handler(tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone), nil)
	}

	handler(tcell.NewEventKey(tcell.KeyCtrlK, 0, tcell.ModNone), nil)

	require.Len(t, events, 1)
	assert.Equal(t, "hello world", events[0].OldValue)
	assert.Equal(t, "hello ", events[0].NewValue)
}

// TestRadioGroup_ChangeOnSelect tests RadioGroup emits change on selection.
func TestRadioGroup_ChangeOnSelect(t *testing.T) {
	rg := NewRadioGroup("test").SetOptions([]string{"opt1", "opt2", "opt3"})

	var events []*ChangeEvent[string]
	rg.SetOnChange(func(e *ChangeEvent[string]) {
		events = append(events, e)
	})

	handler := rg.InputHandler()

	// Press space to select first option
	handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone), nil)

	require.Len(t, events, 1)
	assert.Equal(t, "", events[0].OldValue)
	assert.Equal(t, "opt1", events[0].NewValue)
	assert.Equal(t, 0, events[0].Index)
}

// TestMultiSelect_ChangeOnToggle tests MultiSelect emits change on toggle.
func TestMultiSelect_ChangeOnToggle(t *testing.T) {
	ms := NewMultiSelect("test").SetOptions([]string{"a", "b", "c"})

	var events []*ChangeEvent[[]SelectOption]
	ms.SetOnChange(func(e *ChangeEvent[[]SelectOption]) {
		events = append(events, e)
	})

	handler := ms.InputHandler()

	// Toggle first option
	handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone), nil)

	require.Len(t, events, 1)
	assert.Len(t, events[0].OldValue, 0)
	assert.Len(t, events[0].NewValue, 1)
	assert.Equal(t, "a", events[0].NewValue[0].Value)
}

// TestEventOrdering tests that events are received in order.
func TestEventOrdering(t *testing.T) {
	field := NewTextField("test")

	var order []EventType
	field.OnEvent(func(e Event) {
		order = append(order, e.Type())
	})

	handler := field.InputHandler()

	// Type multiple characters then submit
	handler(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	expected := []EventType{EventChange, EventChange, EventChange, EventSubmit}
	assert.Equal(t, expected, order)
}

// TestEventTiming tests async event handling with timeout.
func TestEventTiming(t *testing.T) {
	field := NewTextField("test")

	eventCh := make(chan Event, 10)
	field.OnEvent(func(e Event) {
		eventCh <- e
	})

	go func() {
		handler := field.InputHandler()
		handler(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), nil)
	}()

	select {
	case e := <-eventCh:
		assert.Equal(t, EventChange, e.Type())
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}
