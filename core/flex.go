package core

import "github.com/gdamore/tcell/v2"

// Direction controls whether Flex lays out children horizontally or vertically.
type Direction int

const (
	Column Direction = iota // vertical — children stacked top to bottom
	Row                     // horizontal — children side by side left to right
)

type flexItem struct {
	widget     Widget
	fixedSize  int
	proportion int
	focus      bool
}

// Flex is a linear layout that divides space between children using fixed sizes
// and proportions, mirroring the long-standing Flex layout API so call sites are
// grep-replaceable.
//
//	layout := core.NewFlex().
//	    SetDirection(core.Column).
//	    AddItem(topBar,  3, 0, false).
//	    AddItem(content, 0, 1, true).
//	    AddItem(status,  1, 0, false)
type Flex struct {
	Box
	direction Direction
	items     []flexItem
}

// NewFlex returns a Flex defaulting to Column direction.
func NewFlex() *Flex { return &Flex{direction: Column} }

// SetDirection sets layout direction.
func (f *Flex) SetDirection(d Direction) *Flex { f.direction = d; return f }

// AddItem appends a child. fixedSize > 0 gives exact cells; proportion > 0
// shares remaining space proportionally.
func (f *Flex) AddItem(w Widget, fixedSize, proportion int, focus bool) *Flex {
	f.items = append(f.items, flexItem{w, fixedSize, proportion, focus})
	return f
}

// Focus directs focus to the child flagged with focus=true in AddItem (the
// last such child wins). Without this, a parent focusing the Flex would leave
// focus on the Flex's own Box and key events would never reach any child.
func (f *Flex) Focus() {
	if t := f.focusTarget(); t != nil {
		t.Focus()
		return
	}
	f.Box.Focus()
}

// HasFocus reports whether any descendant currently holds focus.
func (f *Flex) HasFocus() bool {
	for _, it := range f.items {
		if it.widget.HasFocus() {
			return true
		}
	}
	return f.Box.HasFocus()
}

// focusTarget returns the child that should receive focus, or nil if none is
// flagged. The last item added with focus=true takes precedence.
func (f *Flex) focusTarget() Widget {
	var target Widget
	for _, it := range f.items {
		if it.focus {
			target = it.widget
		}
	}
	return target
}

// RemoveItem removes the first occurrence of w from the item list.
func (f *Flex) RemoveItem(w Widget) *Flex {
	out := f.items[:0]
	for _, it := range f.items {
		if it.widget != w {
			out = append(out, it)
		}
	}
	f.items = out
	return f
}

// Clear removes all children.
func (f *Flex) Clear() *Flex { f.items = f.items[:0]; return f }

// ItemCount returns the number of children.
func (f *Flex) ItemCount() int { return len(f.items) }

// Draw lays out children and draws them.
func (f *Flex) Draw(screen tcell.Screen) {
	f.Box.Draw(screen)
	x, y, w, h := f.InnerRect()
	if w <= 0 || h <= 0 {
		return
	}

	available := h
	if f.direction == Row {
		available = w
	}

	// Sum fixed sizes and total proportions.
	remaining := available
	totalProp := 0
	for _, it := range f.items {
		if it.fixedSize > 0 {
			remaining -= it.fixedSize
		} else {
			totalProp += it.proportion
		}
	}

	pos := 0
	for i, it := range f.items {
		var size int
		if it.fixedSize > 0 {
			size = it.fixedSize
		} else if totalProp > 0 && it.proportion > 0 {
			if i == len(f.items)-1 || isLastProp(f.items[i:]) {
				size = remaining
			} else {
				size = remaining * it.proportion / totalProp
			}
			remaining -= size
			totalProp -= it.proportion
		}

		if size <= 0 {
			continue
		}

		if f.direction == Row {
			it.widget.SetRect(x+pos, y, size, h)
		} else {
			it.widget.SetRect(x, y+pos, w, size)
		}
		it.widget.Draw(screen)
		pos += size
	}
}

// isLastProp reports whether items[0] is the last item with proportion > 0.
func isLastProp(items []flexItem) bool {
	for _, it := range items[1:] {
		if it.proportion > 0 {
			return false
		}
	}
	return true
}

// HandleKey routes the event to the focused child, falling back to the
// focus-flagged child when no descendant holds tcell focus.
//
// In manually-dispatched key trees (e.g. nav.Pages → component → Split →
// Panel → Flex) tcell focus rests on an ancestor and never reaches a leaf,
// so HasFocus() is false for every child. Without the fallback the event
// would be dropped here and navigation keys would never reach the child that
// was added with focus=true.
func (f *Flex) HandleKey(ev *tcell.EventKey) bool {
	for _, it := range f.items {
		if it.widget.HasFocus() {
			if kh, ok := it.widget.(KeyHandler); ok {
				return kh.HandleKey(ev)
			}
			return false
		}
	}
	if t := f.focusTarget(); t != nil {
		if kh, ok := t.(KeyHandler); ok {
			return kh.HandleKey(ev)
		}
	}
	return false
}

// HandleMouse routes to the child at the event position.
func (f *Flex) HandleMouse(action MouseAction, ev *tcell.EventMouse) (bool, Widget) {
	mx, my := ev.Position()
	for _, it := range f.items {
		wx, wy, ww, wh := it.widget.Rect()
		if mx >= wx && mx < wx+ww && my >= wy && my < wy+wh {
			if mh, ok := it.widget.(MouseHandler); ok {
				return mh.HandleMouse(action, ev)
			}
		}
	}
	return false, nil
}

// Children returns the child widgets in order (implements Container).
func (f *Flex) Children() []Widget {
	out := make([]Widget, len(f.items))
	for i, it := range f.items {
		out[i] = it.widget
	}
	return out
}

// DescendantsAt returns widgets whose Rect contains (x, y), deepest first
// (implements Container).
func (f *Flex) DescendantsAt(x, y int) []Widget {
	var result []Widget
	for _, it := range f.items {
		wx, wy, ww, wh := it.widget.Rect()
		if x >= wx && x < wx+ww && y >= wy && y < wy+wh {
			if c, ok := it.widget.(Container); ok {
				result = append(c.DescendantsAt(x, y), it.widget)
			} else {
				result = append(result, it.widget)
			}
		}
	}
	return result
}
