package input

import "github.com/gdamore/tcell/v2"

// VimNavigator defines navigation callbacks for vim-style key bindings.
// Set any callback to nil to skip binding that action.
type VimNavigator struct {
	Up       func() // k, Up arrow
	Down     func() // j, Down arrow
	Left     func() // h, Left arrow
	Right    func() // l, Right arrow
	Top      func() // g (first press stores, second executes gg), Home
	Bottom   func() // G, End
	PageUp   func() // Ctrl+U, PgUp
	PageDown func() // Ctrl+D, PgDn
	Select   func() // Enter, Space
	Back     func() // Escape, q
}

// AddVimNavigation adds vim-style navigation bindings to the KeyBindings.
// Only non-nil callbacks in the VimNavigator will be bound.
func (kb *KeyBindings) AddVimNavigation(nav VimNavigator) *KeyBindings {
	// Basic movement - Up
	if nav.Up != nil {
		kb.OnRune('k', func(e *tcell.EventKey) bool {
			nav.Up()
			return true
		})
		kb.On(tcell.KeyUp, func(e *tcell.EventKey) bool {
			nav.Up()
			return true
		})
	}

	// Basic movement - Down
	if nav.Down != nil {
		kb.OnRune('j', func(e *tcell.EventKey) bool {
			nav.Down()
			return true
		})
		kb.On(tcell.KeyDown, func(e *tcell.EventKey) bool {
			nav.Down()
			return true
		})
	}

	// Basic movement - Left
	if nav.Left != nil {
		kb.OnRune('h', func(e *tcell.EventKey) bool {
			nav.Left()
			return true
		})
		kb.On(tcell.KeyLeft, func(e *tcell.EventKey) bool {
			nav.Left()
			return true
		})
	}

	// Basic movement - Right
	if nav.Right != nil {
		kb.OnRune('l', func(e *tcell.EventKey) bool {
			nav.Right()
			return true
		})
		kb.On(tcell.KeyRight, func(e *tcell.EventKey) bool {
			nav.Right()
			return true
		})
	}

	// Jump to top - Home key always works
	if nav.Top != nil {
		kb.On(tcell.KeyHome, func(e *tcell.EventKey) bool {
			nav.Top()
			return true
		})
		// 'g' for gg sequence is handled separately via GG helper if needed
	}

	// Jump to bottom
	if nav.Bottom != nil {
		kb.OnRune('G', func(e *tcell.EventKey) bool {
			nav.Bottom()
			return true
		})
		kb.On(tcell.KeyEnd, func(e *tcell.EventKey) bool {
			nav.Bottom()
			return true
		})
	}

	// Page navigation
	if nav.PageUp != nil {
		kb.OnCtrlRune('u', func(e *tcell.EventKey) bool {
			nav.PageUp()
			return true
		})
		kb.On(tcell.KeyPgUp, func(e *tcell.EventKey) bool {
			nav.PageUp()
			return true
		})
	}

	if nav.PageDown != nil {
		kb.OnCtrlRune('d', func(e *tcell.EventKey) bool {
			nav.PageDown()
			return true
		})
		kb.On(tcell.KeyPgDn, func(e *tcell.EventKey) bool {
			nav.PageDown()
			return true
		})
	}

	// Selection
	if nav.Select != nil {
		kb.On(tcell.KeyEnter, func(e *tcell.EventKey) bool {
			nav.Select()
			return true
		})
		kb.OnRune(' ', func(e *tcell.EventKey) bool {
			nav.Select()
			return true
		})
	}

	// Back/escape
	if nav.Back != nil {
		kb.On(tcell.KeyEscape, func(e *tcell.EventKey) bool {
			nav.Back()
			return true
		})
		kb.OnRune('q', func(e *tcell.EventKey) bool {
			nav.Back()
			return true
		})
	}

	return kb
}

// ListNavigator is an interface for components that support list-style navigation.
type ListNavigator interface {
	MoveUp()
	MoveDown()
	MoveToTop()
	MoveToBottom()
}

// VimListBindings returns standard vim bindings for list navigation.
// The returned KeyBindings can be extended with additional bindings.
func VimListBindings(list ListNavigator) *KeyBindings {
	return NewKeyBindings().AddVimNavigation(VimNavigator{
		Up:     list.MoveUp,
		Down:   list.MoveDown,
		Top:    list.MoveToTop,
		Bottom: list.MoveToBottom,
	})
}

// GGHandler creates a handler for the 'gg' vim sequence.
// It returns a handler that tracks 'g' presses and executes onGG on the second press.
// The returned reset function should be called when focus is lost or other keys are pressed.
func GGHandler(onGG func()) (handler Handler, reset func()) {
	gPressed := false

	reset = func() {
		gPressed = false
	}

	handler = func(e *tcell.EventKey) bool {
		if e.Key() == tcell.KeyRune && e.Rune() == 'g' {
			if gPressed {
				onGG()
				gPressed = false
				return true
			}
			gPressed = true
			return true
		}
		// Reset on any other key
		gPressed = false
		return false
	}

	return handler, reset
}

// AddGG adds 'gg' sequence handling for jumping to top.
// This properly handles the two-key sequence.
func (kb *KeyBindings) AddGG(onGG func()) *KeyBindings {
	var gPressed bool

	// Override 'g' to handle the sequence
	kb.OnRune('g', func(e *tcell.EventKey) bool {
		if gPressed {
			onGG()
			gPressed = false
			return true
		}
		gPressed = true
		return true
	})

	// Store original fallback
	originalFallback := kb.fallback

	// Set new fallback that resets g state
	kb.SetFallback(func(e *tcell.EventKey) bool {
		gPressed = false
		if originalFallback != nil {
			return originalFallback(e)
		}
		return false
	})

	return kb
}
