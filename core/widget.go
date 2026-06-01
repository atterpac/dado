// Package core provides the foundational widget abstractions for dado.
// It imports only tcell — no dado sub-packages.
package core

import (
	"github.com/gdamore/tcell/v2"
)

// Widget is the minimum interface every drawable component implements.
// Compose with KeyHandler, MouseHandler, and Container for interactive widgets.
type Widget interface {
	Draw(screen tcell.Screen)
	Rect() (x, y, w, h int)
	SetRect(x, y, w, h int)
	Focus()
	Blur()
	HasFocus() bool
}

// KeyHandler is implemented by widgets that consume keyboard input.
// Return true to mark the event consumed and stop propagation up the tree.
type KeyHandler interface {
	HandleKey(ev *tcell.EventKey) bool
}

// MouseHandler is implemented by widgets that consume mouse input.
// capture: non-nil widget receives all subsequent mouse events (drag support).
type MouseHandler interface {
	HandleMouse(action MouseAction, ev *tcell.EventMouse) (consumed bool, capture Widget)
}

// PasteHandler is implemented by widgets that accept pasted text (e.g. inputs).
type PasteHandler interface {
	HandlePaste(text string) bool
}

// Container is implemented by layout widgets that own children.
// The App uses DescendantsAt for mouse hit-testing without each container
// reimplementing the tree walk.
type Container interface {
	Widget
	Children() []Widget
	// DescendantsAt returns widgets whose Rect contains (x, y), deepest first.
	DescendantsAt(x, y int) []Widget
}
