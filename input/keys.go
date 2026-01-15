package input

import "github.com/gdamore/tcell/v2"

// Key represents a keyboard key.
// This wraps tcell.Key so users don't need to import tcell directly.
type Key = tcell.Key

// Common key constants - users can use these instead of tcell.Key* constants.
const (
	KeyEnter     Key = tcell.KeyEnter
	KeyTab       Key = tcell.KeyTab
	KeyBacktab   Key = tcell.KeyBacktab // Shift+Tab
	KeyEscape    Key = tcell.KeyEscape
	KeyEsc       Key = tcell.KeyEscape // Alias
	KeyBackspace Key = tcell.KeyBackspace2
	KeyDelete    Key = tcell.KeyDelete
	KeyInsert    Key = tcell.KeyInsert
	KeyHome      Key = tcell.KeyHome
	KeyEnd       Key = tcell.KeyEnd
	KeyPageUp    Key = tcell.KeyPgUp
	KeyPageDown  Key = tcell.KeyPgDn
	KeyUp        Key = tcell.KeyUp
	KeyDown      Key = tcell.KeyDown
	KeyLeft      Key = tcell.KeyLeft
	KeyRight     Key = tcell.KeyRight
	KeyF1        Key = tcell.KeyF1
	KeyF2        Key = tcell.KeyF2
	KeyF3        Key = tcell.KeyF3
	KeyF4        Key = tcell.KeyF4
	KeyF5        Key = tcell.KeyF5
	KeyF6        Key = tcell.KeyF6
	KeyF7        Key = tcell.KeyF7
	KeyF8        Key = tcell.KeyF8
	KeyF9        Key = tcell.KeyF9
	KeyF10       Key = tcell.KeyF10
	KeyF11       Key = tcell.KeyF11
	KeyF12       Key = tcell.KeyF12
	KeySpace     Key = tcell.KeyRune // Special handling needed

	// Ctrl key combinations
	KeyCtrlA     Key = tcell.KeyCtrlA
	KeyCtrlB     Key = tcell.KeyCtrlB
	KeyCtrlC     Key = tcell.KeyCtrlC
	KeyCtrlD     Key = tcell.KeyCtrlD
	KeyCtrlE     Key = tcell.KeyCtrlE
	KeyCtrlF     Key = tcell.KeyCtrlF
	KeyCtrlG     Key = tcell.KeyCtrlG
	KeyCtrlH     Key = tcell.KeyCtrlH
	KeyCtrlI     Key = tcell.KeyCtrlI
	KeyCtrlJ     Key = tcell.KeyCtrlJ
	KeyCtrlK     Key = tcell.KeyCtrlK
	KeyCtrlL     Key = tcell.KeyCtrlL
	KeyCtrlM     Key = tcell.KeyCtrlM
	KeyCtrlN     Key = tcell.KeyCtrlN
	KeyCtrlO     Key = tcell.KeyCtrlO
	KeyCtrlP     Key = tcell.KeyCtrlP
	KeyCtrlQ     Key = tcell.KeyCtrlQ
	KeyCtrlR     Key = tcell.KeyCtrlR
	KeyCtrlS     Key = tcell.KeyCtrlS
	KeyCtrlT     Key = tcell.KeyCtrlT
	KeyCtrlU     Key = tcell.KeyCtrlU
	KeyCtrlV     Key = tcell.KeyCtrlV
	KeyCtrlW     Key = tcell.KeyCtrlW
	KeyCtrlX     Key = tcell.KeyCtrlX
	KeyCtrlY     Key = tcell.KeyCtrlY
	KeyCtrlZ     Key = tcell.KeyCtrlZ
)

// Bind creates a simple key binding that calls the handler when the key is pressed.
// This is a convenience function for the most common use case.
//
// Example:
//
//	bindings := input.NewKeyBindings().
//	    Bind(input.KeyEnter, func() { submit() }).
//	    Bind(input.KeyEscape, func() { cancel() }).
//	    Bind('q', func() { quit() })
func (kb *KeyBindings) Bind(key any, handler func()) *KeyBindings {
	wrappedHandler := func(event *tcell.EventKey) bool {
		handler()
		return true
	}

	switch k := key.(type) {
	case Key: // Key is an alias for tcell.Key
		kb.keyBindings[k] = wrappedHandler
	case rune:
		kb.runeBindings[k] = wrappedHandler
	case byte:
		kb.runeBindings[rune(k)] = wrappedHandler
	}

	return kb
}

// BindString binds a handler to a key described by a string.
// Supported formats:
//   - Single character: "q", "r", "1", etc.
//   - Special keys: "enter", "escape", "esc", "tab", "space", "backspace", "delete"
//   - Arrow keys: "up", "down", "left", "right"
//   - Function keys: "f1", "f2", ... "f12"
//   - Ctrl combinations: "ctrl+s", "ctrl+q", etc.
//
// Example:
//
//	bindings := input.NewKeyBindings().
//	    BindString("enter", func() { submit() }).
//	    BindString("ctrl+s", func() { save() }).
//	    BindString("q", func() { quit() })
func (kb *KeyBindings) BindString(keyStr string, handler func()) *KeyBindings {
	wrappedHandler := func(event *tcell.EventKey) bool {
		handler()
		return true
	}

	// Handle ctrl+ prefix
	if len(keyStr) > 5 && keyStr[:5] == "ctrl+" {
		letter := keyStr[5:]
		if len(letter) == 1 {
			r := rune(letter[0])
			if r >= 'a' && r <= 'z' {
				key := tcell.Key(int(tcell.KeyCtrlA) + int(r-'a'))
				kb.keyBindings[key] = wrappedHandler
				return kb
			}
			if r >= 'A' && r <= 'Z' {
				key := tcell.Key(int(tcell.KeyCtrlA) + int(r-'A'))
				kb.keyBindings[key] = wrappedHandler
				return kb
			}
		}
	}

	// Handle special keys
	switch keyStr {
	case "enter", "return":
		kb.keyBindings[tcell.KeyEnter] = wrappedHandler
	case "escape", "esc":
		kb.keyBindings[tcell.KeyEscape] = wrappedHandler
	case "tab":
		kb.keyBindings[tcell.KeyTab] = wrappedHandler
	case "backtab", "shift+tab":
		kb.keyBindings[tcell.KeyBacktab] = wrappedHandler
	case "space":
		kb.runeBindings[' '] = wrappedHandler
	case "backspace":
		kb.keyBindings[tcell.KeyBackspace2] = wrappedHandler
	case "delete", "del":
		kb.keyBindings[tcell.KeyDelete] = wrappedHandler
	case "insert", "ins":
		kb.keyBindings[tcell.KeyInsert] = wrappedHandler
	case "home":
		kb.keyBindings[tcell.KeyHome] = wrappedHandler
	case "end":
		kb.keyBindings[tcell.KeyEnd] = wrappedHandler
	case "pageup", "pgup":
		kb.keyBindings[tcell.KeyPgUp] = wrappedHandler
	case "pagedown", "pgdown", "pgdn":
		kb.keyBindings[tcell.KeyPgDn] = wrappedHandler
	case "up":
		kb.keyBindings[tcell.KeyUp] = wrappedHandler
	case "down":
		kb.keyBindings[tcell.KeyDown] = wrappedHandler
	case "left":
		kb.keyBindings[tcell.KeyLeft] = wrappedHandler
	case "right":
		kb.keyBindings[tcell.KeyRight] = wrappedHandler
	case "f1":
		kb.keyBindings[tcell.KeyF1] = wrappedHandler
	case "f2":
		kb.keyBindings[tcell.KeyF2] = wrappedHandler
	case "f3":
		kb.keyBindings[tcell.KeyF3] = wrappedHandler
	case "f4":
		kb.keyBindings[tcell.KeyF4] = wrappedHandler
	case "f5":
		kb.keyBindings[tcell.KeyF5] = wrappedHandler
	case "f6":
		kb.keyBindings[tcell.KeyF6] = wrappedHandler
	case "f7":
		kb.keyBindings[tcell.KeyF7] = wrappedHandler
	case "f8":
		kb.keyBindings[tcell.KeyF8] = wrappedHandler
	case "f9":
		kb.keyBindings[tcell.KeyF9] = wrappedHandler
	case "f10":
		kb.keyBindings[tcell.KeyF10] = wrappedHandler
	case "f11":
		kb.keyBindings[tcell.KeyF11] = wrappedHandler
	case "f12":
		kb.keyBindings[tcell.KeyF12] = wrappedHandler
	default:
		// Single character
		if len(keyStr) == 1 {
			kb.runeBindings[rune(keyStr[0])] = wrappedHandler
		}
	}

	return kb
}

// OnKey is an alias for On that accepts jig Key constants.
// Deprecated: Use Bind instead for simpler handlers.
func (kb *KeyBindings) OnKey(key Key, handler Handler) *KeyBindings {
	return kb.On(key, handler)
}

// BindCtrl binds a simple handler to Ctrl+letter.
// Example: BindCtrl('s', save) binds Ctrl+S to save().
func (kb *KeyBindings) BindCtrl(r rune, handler func()) *KeyBindings {
	wrappedHandler := func(event *tcell.EventKey) bool {
		handler()
		return true
	}

	if r >= 'a' && r <= 'z' {
		key := tcell.Key(int(tcell.KeyCtrlA) + int(r-'a'))
		kb.keyBindings[key] = wrappedHandler
	} else if r >= 'A' && r <= 'Z' {
		key := tcell.Key(int(tcell.KeyCtrlA) + int(r-'A'))
		kb.keyBindings[key] = wrappedHandler
	}
	return kb
}

// BindShift binds a simple handler to Shift+key.
// Note: For letters, just use the uppercase rune with Bind().
// This is mainly useful for Shift+special keys like Shift+Tab.
func (kb *KeyBindings) BindShift(key Key, handler func()) *KeyBindings {
	wrappedHandler := func(event *tcell.EventKey) bool {
		handler()
		return true
	}
	kb.modBindings[keyMod{key, tcell.ModShift}] = wrappedHandler
	return kb
}

// BindAlt binds a simple handler to Alt+key (for special keys like Alt+Enter).
func (kb *KeyBindings) BindAlt(key Key, handler func()) *KeyBindings {
	wrappedHandler := func(event *tcell.EventKey) bool {
		handler()
		return true
	}
	kb.modBindings[keyMod{key, tcell.ModAlt}] = wrappedHandler
	return kb
}
