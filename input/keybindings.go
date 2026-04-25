package input

import (
	"maps"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Handler processes a key event and returns true if the event was handled.
type Handler func(event *tcell.EventKey) bool

// keyMod represents a key with modifiers for use as a map key.
type keyMod struct {
	key tcell.Key
	mod tcell.ModMask
}

// KeyBindings manages key bindings for a component.
// It provides a fluent API for building input handlers.
type KeyBindings struct {
	keyBindings  map[tcell.Key]Handler
	runeBindings map[rune]Handler
	modBindings  map[keyMod]Handler
	fallback     Handler
}

// NewKeyBindings creates a new key bindings builder.
func NewKeyBindings() *KeyBindings {
	return &KeyBindings{
		keyBindings:  make(map[tcell.Key]Handler),
		runeBindings: make(map[rune]Handler),
		modBindings:  make(map[keyMod]Handler),
	}
}

// On binds a handler to a specific key.
func (kb *KeyBindings) On(key tcell.Key, handler Handler) *KeyBindings {
	kb.keyBindings[key] = handler
	return kb
}

// OnRune binds a handler to a specific rune.
func (kb *KeyBindings) OnRune(r rune, handler Handler) *KeyBindings {
	kb.runeBindings[r] = handler
	return kb
}

// OnRunes binds a handler to multiple runes.
func (kb *KeyBindings) OnRunes(runes string, handler Handler) *KeyBindings {
	for _, r := range runes {
		kb.runeBindings[r] = handler
	}
	return kb
}

// OnMod binds a handler to a key with modifiers.
func (kb *KeyBindings) OnMod(key tcell.Key, mod tcell.ModMask, handler Handler) *KeyBindings {
	kb.modBindings[keyMod{key, mod}] = handler
	return kb
}

// OnCtrl binds a handler to Ctrl+key.
func (kb *KeyBindings) OnCtrl(key tcell.Key, handler Handler) *KeyBindings {
	return kb.OnMod(key, tcell.ModCtrl, handler)
}

// OnAlt binds a handler to Alt+key.
func (kb *KeyBindings) OnAlt(key tcell.Key, handler Handler) *KeyBindings {
	return kb.OnMod(key, tcell.ModAlt, handler)
}

// OnShift binds a handler to Shift+key.
func (kb *KeyBindings) OnShift(key tcell.Key, handler Handler) *KeyBindings {
	return kb.OnMod(key, tcell.ModShift, handler)
}

// OnCtrlRune binds a handler to Ctrl+letter (e.g., Ctrl+S).
// Note: Ctrl+letter combinations map to specific key codes in tcell.
func (kb *KeyBindings) OnCtrlRune(r rune, handler Handler) *KeyBindings {
	// Ctrl+letter maps to specific key codes
	// Ctrl+A = KeyCtrlA, Ctrl+B = KeyCtrlB, etc.
	if r >= 'a' && r <= 'z' {
		key := tcell.Key(int(tcell.KeyCtrlA) + int(r-'a'))
		kb.keyBindings[key] = handler
	} else if r >= 'A' && r <= 'Z' {
		key := tcell.Key(int(tcell.KeyCtrlA) + int(r-'A'))
		kb.keyBindings[key] = handler
	}
	return kb
}

// SetFallback sets a handler for unmatched keys.
// The fallback is called when no other handler matches.
func (kb *KeyBindings) SetFallback(handler Handler) *KeyBindings {
	kb.fallback = handler
	return kb
}

// Handle processes an event and returns true if handled.
func (kb *KeyBindings) Handle(event *tcell.EventKey) bool {
	// Check modifier combinations first (most specific)
	if event.Modifiers() != 0 {
		if handler, ok := kb.modBindings[keyMod{event.Key(), event.Modifiers()}]; ok {
			return handler(event)
		}
	}

	// Check specific keys
	if handler, ok := kb.keyBindings[event.Key()]; ok {
		return handler(event)
	}

	// Check runes
	if event.Key() == tcell.KeyRune {
		if handler, ok := kb.runeBindings[event.Rune()]; ok {
			return handler(event)
		}
	}

	// Fallback handler
	if kb.fallback != nil {
		return kb.fallback(event)
	}

	return false
}

// Build creates a tview-compatible input handler function.
// The returned function can be passed to SetInputHandler on tview primitives.
func (kb *KeyBindings) Build() func(*tcell.EventKey, func(tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
		if kb.Handle(event) {
			return nil
		}
		return event
	}
}

// BuildBool creates an input handler with the bool-return convention used by
// components.ComponentBase.SetInputHandler. Returns true when an event was consumed.
func (kb *KeyBindings) BuildBool() func(*tcell.EventKey, func(tview.Primitive)) bool {
	return func(event *tcell.EventKey, _ func(tview.Primitive)) bool {
		return kb.Handle(event)
	}
}

// BuildWithFocus creates a handler that provides access to the setFocus function.
// Use this when your handlers need to change focus to other primitives.
func (kb *KeyBindings) BuildWithFocus(focusHandler func(setFocus func(tview.Primitive))) func(*tcell.EventKey, func(tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
		if focusHandler != nil {
			focusHandler(setFocus)
		}
		if kb.Handle(event) {
			return nil
		}
		return event
	}
}

// Merge combines this KeyBindings with another, with the other taking precedence.
func (kb *KeyBindings) Merge(other *KeyBindings) *KeyBindings {
	if other == nil {
		return kb
	}

	maps.Copy(kb.keyBindings, other.keyBindings)
	maps.Copy(kb.runeBindings, other.runeBindings)
	maps.Copy(kb.modBindings, other.modBindings)

	if other.fallback != nil {
		kb.fallback = other.fallback
	}

	return kb
}

// Clone creates a copy of this KeyBindings.
func (kb *KeyBindings) Clone() *KeyBindings {
	clone := NewKeyBindings()

	maps.Copy(clone.keyBindings, kb.keyBindings)
	maps.Copy(clone.runeBindings, kb.runeBindings)
	maps.Copy(clone.modBindings, kb.modBindings)
	clone.fallback = kb.fallback

	return clone
}
