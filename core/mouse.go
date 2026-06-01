package core

// MouseAction identifies a mouse event type.
// Values mirror the underlying tcell mouse action semantics.
type MouseAction int16

const (
	MouseMove         MouseAction = 0
	MouseLeftDown     MouseAction = 1
	MouseMiddleDown   MouseAction = 2
	MouseRightDown    MouseAction = 3
	MouseLeftUp       MouseAction = 4
	MouseMiddleUp     MouseAction = 5
	MouseRightUp      MouseAction = 6
	MouseLeftClick    MouseAction = 7
	MouseMiddleClick  MouseAction = 8
	MouseRightClick   MouseAction = 9
	MouseLeftDoubleClick  MouseAction = 10
	MouseScrollUp     MouseAction = 11
	MouseScrollDown   MouseAction = 12
	MouseScrollLeft   MouseAction = 13
	MouseScrollRight  MouseAction = 14
)
