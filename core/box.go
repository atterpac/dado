package core

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

var (
	defaultBgMu sync.RWMutex
	defaultBgFn func() tcell.Color
)

// SetDefaultBackgroundFunc registers a global fallback called by every Box
// whose bgFn is nil. Call this once at app startup (e.g. theme.Bg) so all
// core primitives (Flex, TextView, Table, …) track the active theme without
// per-instance wiring.
func SetDefaultBackgroundFunc(fn func() tcell.Color) {
	defaultBgMu.Lock()
	defaultBgFn = fn
	defaultBgMu.Unlock()
}

// Box is a value-type rect/focus/border manager.
// Embed in leaf widgets via widgetBase — zero value is ready, no constructor needed.
//
//	type Badge struct {
//	    core.Box           // embeds by value — no heap alloc
//	    text string
//	}
type Box struct {
	x, y, w, h       int
	padTop, padBot    int
	padLeft, padRight int

	border           bool
	borderStyle      tcell.Style
	borderFocusStyle tcell.Style
	title            string
	titleAlign       int

	hasFocus        bool
	backgroundColor tcell.Color
	bgFn            func() tcell.Color // if set, overrides backgroundColor in drawBackground
}

// --- Widget interface ---

// Rect returns the box's position and size.
func (b *Box) Rect() (int, int, int, int) { return b.x, b.y, b.w, b.h }

// SetRect sets the box's position and size.
func (b *Box) SetRect(x, y, w, h int) { b.x, b.y, b.w, b.h = x, y, w, h }

// Focus marks the box as focused.
func (b *Box) Focus() { b.hasFocus = true }

// Blur removes focus.
func (b *Box) Blur() { b.hasFocus = false }

// HasFocus reports whether the box is focused.
func (b *Box) HasFocus() bool { return b.hasFocus }

// Draw renders background and border. Leaf widgets call DrawForSubclass instead.
func (b *Box) Draw(screen tcell.Screen) {
	b.drawBackground(screen)
	if b.border {
		b.drawBorder(screen)
	}
}

// GetInnerRect is an alias for InnerRect.
func (b *Box) GetInnerRect() (x, y, w, h int) { return b.InnerRect() }

// GetRect is an alias for Rect.
func (b *Box) GetRect() (int, int, int, int) { return b.Rect() }

// InRect reports whether (x, y) falls within the box's current rect.
func (b *Box) InRect(x, y int) bool {
	return x >= b.x && x < b.x+b.w && y >= b.y && y < b.y+b.h
}

// InnerRect returns the drawable area after border and padding are applied.
func (b *Box) InnerRect() (x, y, w, h int) {
	x, y, w, h = b.x, b.y, b.w, b.h
	if b.border {
		x++
		y++
		w -= 2
		h -= 2
	}
	x += b.padLeft
	y += b.padTop
	w -= b.padLeft + b.padRight
	h -= b.padTop + b.padBot
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return
}

// DrawForSubclass draws box chrome (background, border, title).
// The subclass calls this at the start of its Draw method to render the frame,
// then draws its own content into InnerRect().
//
//	func (b *Badge) Draw(screen tcell.Screen) {
//	    b.Box.DrawForSubclass(screen)      // draws bg + border only
//	    x, y, w, h := b.GetInnerRect()    // get inner area
//	    // draw badge content at x,y,w,h
//	}
func (b *Box) DrawForSubclass(screen tcell.Screen) {
	b.drawBackground(screen)
	if b.border {
		b.drawBorder(screen)
	}
}

// --- Configuration (fluent) ---

// SetBorder enables or disables the border.
func (b *Box) SetBorder(on bool) *Box { b.border = on; return b }

// SetTitle sets the title shown in the top border row.
func (b *Box) SetTitle(t string) *Box { b.title = t; return b }

// SetTitleAlign sets title alignment: AlignLeft, AlignCenter, AlignRight.
func (b *Box) SetTitleAlign(a int) *Box { b.titleAlign = a; return b }

// SetPadding sets inner padding (top, bottom, left, right).
func (b *Box) SetPadding(top, bot, left, right int) *Box {
	b.padTop, b.padBot, b.padLeft, b.padRight = top, bot, left, right
	return b
}

// SetBackgroundColor sets the fill color. Satisfies theme.Backgroundable.
func (b *Box) SetBackgroundColor(c tcell.Color) *Box { b.backgroundColor = c; return b }

// SetBackgroundFunc sets a live getter called on every draw instead of the cached backgroundColor.
// Pass theme.Bg (or any func() tcell.Color) to track the active theme automatically.
func (b *Box) SetBackgroundFunc(fn func() tcell.Color) *Box { b.bgFn = fn; return b }

// bg returns the effective background color: bgFn → defaultBgFn → backgroundColor.
func (b *Box) bg() tcell.Color {
	if b.bgFn != nil {
		return b.bgFn()
	}
	defaultBgMu.RLock()
	fn := defaultBgFn
	defaultBgMu.RUnlock()
	if fn != nil {
		return fn()
	}
	return b.backgroundColor
}

// SetBorderStyle sets the style used for the border when unfocused.
func (b *Box) SetBorderStyle(s tcell.Style) *Box { b.borderStyle = s; return b }

// SetBorderFocusStyle sets the style used for the border when focused.
func (b *Box) SetBorderFocusStyle(s tcell.Style) *Box { b.borderFocusStyle = s; return b }

// --- internal ---

func (b *Box) drawBackground(screen tcell.Screen) {
	style := tcell.StyleDefault.Background(b.bg())
	FillRect(screen, b.x, b.y, b.w, b.h, ' ', style)
}

func (b *Box) drawBorder(screen tcell.Screen) {
	style := b.borderStyle
	if b.hasFocus && b.borderFocusStyle != (tcell.Style{}) {
		style = b.borderFocusStyle
	}
	DrawBorder(screen, b.x, b.y, b.w, b.h, style)
	if b.title != "" {
		DrawTitle(screen, b.x, b.y, b.w, b.title, b.titleAlign, style)
	}
}
