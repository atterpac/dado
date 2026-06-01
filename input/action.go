package input

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/dado/components"
)

// KeyAction represents a single key binding with its handler.
type KeyAction struct {
	Key         tcell.Key     // Special key (e.g., tcell.KeyEnter), or tcell.KeyRune for character keys
	Rune        rune          // Character key (e.g., 'd'), used when Key == tcell.KeyRune
	Modifiers   tcell.ModMask // Ctrl, Alt, Shift modifiers
	Description string        // Human-readable description for hints
	Handler     func()        // Action to execute when triggered
}

// ActionRegistry manages a collection of key bindings.
type ActionRegistry struct {
	actions map[string]KeyAction
	order   []string // Maintains insertion order for Hints()
}

// NewActionRegistry creates a new action registry.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]KeyAction),
		order:   make([]string, 0),
	}
}

// Add registers a key action with a unique name.
func (r *ActionRegistry) Add(name string, action KeyAction) *ActionRegistry {
	if _, exists := r.actions[name]; !exists {
		r.order = append(r.order, name)
	}
	r.actions[name] = action
	return r
}

// AddSimple registers a simple rune-based action without modifiers.
func (r *ActionRegistry) AddSimple(name string, key rune, description string, handler func()) *ActionRegistry {
	return r.Add(name, KeyAction{
		Key:         tcell.KeyRune,
		Rune:        key,
		Description: description,
		Handler:     handler,
	})
}

// AddKey registers a special key action (Enter, Esc, etc).
func (r *ActionRegistry) AddKey(name string, key tcell.Key, description string, handler func()) *ActionRegistry {
	return r.Add(name, KeyAction{
		Key:         key,
		Description: description,
		Handler:     handler,
	})
}

// AddCtrl registers a Ctrl+key action.
func (r *ActionRegistry) AddCtrl(name string, key rune, description string, handler func()) *ActionRegistry {
	return r.Add(name, KeyAction{
		Key:         tcell.KeyRune,
		Rune:        key,
		Modifiers:   tcell.ModCtrl,
		Description: description,
		Handler:     handler,
	})
}

// Remove removes an action by name.
func (r *ActionRegistry) Remove(name string) *ActionRegistry {
	delete(r.actions, name)
	// Remove from order
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	return r
}

// Handle processes a key event and returns true if handled.
func (r *ActionRegistry) Handle(event *tcell.EventKey) bool {
	for _, action := range r.actions {
		if r.matchesEvent(action, event) {
			if action.Handler != nil {
				action.Handler()
			}
			return true
		}
	}
	return false
}

// matchesEvent checks if an action matches the given event.
func (r *ActionRegistry) matchesEvent(action KeyAction, event *tcell.EventKey) bool {
	// Check modifiers (Ctrl handled specially by tcell)
	if action.Modifiers&tcell.ModAlt != 0 {
		if event.Modifiers()&tcell.ModAlt == 0 {
			return false
		}
	}

	// Match special keys (non-rune)
	if action.Key != tcell.KeyRune {
		return action.Key == event.Key()
	}

	// Match Ctrl+key FIRST (tcell delivers Ctrl+letter as KeyCtrl*, not as a
	// modified rune). This must precede the plain-rune match below, otherwise a
	// Ctrl+<rune> action (e.g. Ctrl+I) would incorrectly match a plain <rune>
	// press (plain 'i'), since both share Rune=='i'.
	if action.Modifiers&tcell.ModCtrl != 0 {
		// Ctrl+A through Ctrl+Z map to KeyCtrlA through KeyCtrlZ
		ctrlKey := tcell.Key(int(tcell.KeyCtrlA) + int(action.Rune) - int('a'))
		return event.Key() == ctrlKey
	}

	// Match plain rune keys (no Ctrl modifier).
	if action.Rune != 0 && event.Key() == tcell.KeyRune {
		return action.Rune == event.Rune()
	}

	return false
}

// Hints returns key hints for display in insertion order.
func (r *ActionRegistry) Hints() []components.KeyHint {
	hints := make([]components.KeyHint, 0, len(r.order))
	for _, name := range r.order {
		action := r.actions[name]
		hints = append(hints, components.KeyHint{
			Key:         FormatKey(action),
			Description: action.Description,
		})
	}
	return hints
}

// Clear removes all actions.
func (r *ActionRegistry) Clear() *ActionRegistry {
	r.actions = make(map[string]KeyAction)
	r.order = make([]string, 0)
	return r
}

// Merge combines another registry's actions into this one.
// Existing actions with same name are overwritten.
func (r *ActionRegistry) Merge(other *ActionRegistry) *ActionRegistry {
	for _, name := range other.order {
		r.Add(name, other.actions[name])
	}
	return r
}

// Has checks if an action with the given name exists.
func (r *ActionRegistry) Has(name string) bool {
	_, exists := r.actions[name]
	return exists
}

// Get returns an action by name, or empty action if not found.
func (r *ActionRegistry) Get(name string) (KeyAction, bool) {
	action, exists := r.actions[name]
	return action, exists
}

// FormatKey returns a human-readable key string for an action.
func FormatKey(action KeyAction) string {
	var parts []string

	if action.Modifiers&tcell.ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if action.Modifiers&tcell.ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if action.Modifiers&tcell.ModShift != 0 {
		parts = append(parts, "Shift")
	}

	switch action.Key {
	case tcell.KeyEnter:
		parts = append(parts, "Enter")
	case tcell.KeyEscape:
		parts = append(parts, "Esc")
	case tcell.KeyTab:
		parts = append(parts, "Tab")
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		parts = append(parts, "Backspace")
	case tcell.KeyDelete:
		parts = append(parts, "Delete")
	case tcell.KeyUp:
		parts = append(parts, "Up")
	case tcell.KeyDown:
		parts = append(parts, "Down")
	case tcell.KeyLeft:
		parts = append(parts, "Left")
	case tcell.KeyRight:
		parts = append(parts, "Right")
	case tcell.KeyHome:
		parts = append(parts, "Home")
	case tcell.KeyEnd:
		parts = append(parts, "End")
	case tcell.KeyPgUp:
		parts = append(parts, "PgUp")
	case tcell.KeyPgDn:
		parts = append(parts, "PgDn")
	case tcell.KeyF1:
		parts = append(parts, "F1")
	case tcell.KeyF2:
		parts = append(parts, "F2")
	case tcell.KeyF3:
		parts = append(parts, "F3")
	case tcell.KeyF4:
		parts = append(parts, "F4")
	case tcell.KeyF5:
		parts = append(parts, "F5")
	case tcell.KeyF6:
		parts = append(parts, "F6")
	case tcell.KeyF7:
		parts = append(parts, "F7")
	case tcell.KeyF8:
		parts = append(parts, "F8")
	case tcell.KeyF9:
		parts = append(parts, "F9")
	case tcell.KeyF10:
		parts = append(parts, "F10")
	case tcell.KeyF11:
		parts = append(parts, "F11")
	case tcell.KeyF12:
		parts = append(parts, "F12")
	case tcell.KeyRune:
		if action.Rune == ' ' {
			parts = append(parts, "Space")
		} else if action.Rune != 0 {
			parts = append(parts, string(action.Rune))
		}
	default:
		if action.Rune != 0 {
			parts = append(parts, string(action.Rune))
		}
	}

	if len(parts) > 1 {
		return strings.Join(parts, "+")
	} else if len(parts) == 1 {
		return parts[0]
	}
	return ""
}
