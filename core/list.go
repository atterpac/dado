package core

import "github.com/gdamore/tcell/v2"

type listItem struct {
	main      string
	secondary string
	shortcut  rune
	selected  func()
	cached    *Text // main parsed once into styled spans (see AddItem)
}

// List is a scrollable selectable list.
type List struct {
	Box
	items   []listItem
	current int
	offset  int

	wrapAround    bool
	showSecondary bool

	mainTextColor      tcell.Color
	secondaryTextColor tcell.Color
	selectedBgColor    tcell.Color
	selectedTextColor  tcell.Color
	shortcutColor      tcell.Color

	onChange func(cur, prev int, main, secondary string, shortcut rune)
	onSelect func(idx int, main, secondary string, shortcut rune)
}

// NewList returns an empty List.
func NewList() *List { return &List{} }

// AddItem appends an item. shortcut 0 = no shortcut. selected is called on Enter.
func (l *List) AddItem(main, secondary string, shortcut rune, selected func()) *List {
	l.items = append(l.items, listItem{
		main:      main,
		secondary: secondary,
		shortcut:  shortcut,
		selected:  selected,
		cached:    ParseTagged(main, tcell.StyleDefault),
	})
	return l
}

// GetItemCount returns the number of items.
func (l *List) GetItemCount() int { return len(l.items) }

// GetCurrentItem returns the index of the currently selected item.
func (l *List) GetCurrentItem() int { return l.current }

// SetCurrentItem sets the selected item by index (clamped to valid range).
func (l *List) SetCurrentItem(idx int) *List {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(l.items) && len(l.items) > 0 {
		idx = len(l.items) - 1
	}
	l.current = idx
	return l
}

// SetSelectedFunc sets the callback fired when an item is activated (Enter).
func (l *List) SetSelectedFunc(fn func(idx int, main, secondary string, shortcut rune)) *List {
	l.onSelect = fn
	return l
}

// SetChangedFunc sets the callback fired when the selection changes.
func (l *List) SetChangedFunc(fn func(cur, prev int, main, secondary string, shortcut rune)) *List {
	l.onChange = fn
	return l
}

// Draw renders visible items.
func (l *List) Draw(screen tcell.Screen) {
	l.Box.Draw(screen)
	vp := NewViewport(l.InnerRect())
	if vp.Empty() {
		return
	}
	w, _ := vp.Size()

	// Ensure current item is visible (scroll offset adjustment).
	vp.SetContentSize(0, len(l.items))
	vp.SetOffset(0, l.offset)
	vp.EnsureVisible(0, l.current)
	_, l.offset = vp.Offset() // keep field in sync for HandleKey/paging

	first, last := vp.VisibleRows()
	for idx := first; idx < last; idx++ {
		item := l.items[idx]
		selected := idx == l.current
		style := tcell.StyleDefault
		if selected {
			style = style.Reverse(true)
		}
		// Fill row background, then render the pre-parsed text. The selection
		// overlay applies reverse on top of each span's own style, so cached
		// spans (parsed once in AddItem) need no per-frame re-parsing.
		for col := 0; col < w; col++ {
			vp.SetContent(screen, col, idx, ' ', style)
		}
		var overlay func(tcell.Style) tcell.Style
		if selected {
			overlay = func(s tcell.Style) tcell.Style { return s.Reverse(true) }
		}
		rx, ry := vp.ScreenXY(0, idx)
		item.cached.DrawFunc(screen, rx, ry, w, overlay)
	}
}

// HandleKey handles Up/Down/Enter/Home/End.
func (l *List) HandleKey(ev *tcell.EventKey) bool {
	if len(l.items) == 0 {
		return false
	}
	prev := l.current
	switch ev.Key() {
	case tcell.KeyDown:
		if l.current < len(l.items)-1 {
			l.current++
			l.fireChange(prev)
		}
		return true
	case tcell.KeyUp:
		if l.current > 0 {
			l.current--
			l.fireChange(prev)
		}
		return true
	case tcell.KeyHome:
		l.current = 0
		l.fireChange(prev)
		return true
	case tcell.KeyEnd:
		l.current = len(l.items) - 1
		l.fireChange(prev)
		return true
	case tcell.KeyEnter:
		l.activate()
		return true
	}
	return false
}

func (l *List) fireChange(prev int) {
	if l.onChange != nil && l.current != prev {
		item := l.items[l.current]
		l.onChange(l.current, prev, item.main, item.secondary, item.shortcut)
	}
}

// Clear removes all items and resets the selection.
func (l *List) Clear() *List {
	l.items = nil
	l.current = 0
	l.offset = 0
	return l
}

// SetWrapAround enables or disables wrap-around navigation.
func (l *List) SetWrapAround(wrap bool) *List { l.wrapAround = wrap; return l }

// SetHighlightFullLine is a no-op (full-line highlight is always used). Kept for API compatibility.
func (l *List) SetHighlightFullLine(_ bool) *List { return l }

// ShowSecondaryText enables or disables secondary text display.
func (l *List) ShowSecondaryText(show bool) *List { l.showSecondary = show; return l }

// SetMainTextColor sets the color for main item text.
func (l *List) SetMainTextColor(c tcell.Color) *List { l.mainTextColor = c; return l }

// SetSecondaryTextColor sets the color for secondary item text.
func (l *List) SetSecondaryTextColor(c tcell.Color) *List { l.secondaryTextColor = c; return l }

// SetSelectedBackgroundColor sets the background color for the selected item.
func (l *List) SetSelectedBackgroundColor(c tcell.Color) *List { l.selectedBgColor = c; return l }

// SetSelectedTextColor sets the text color for the selected item.
func (l *List) SetSelectedTextColor(c tcell.Color) *List { l.selectedTextColor = c; return l }

// SetShortcutColor sets the color for shortcut characters.
func (l *List) SetShortcutColor(c tcell.Color) *List { l.shortcutColor = c; return l }

func (l *List) activate() {
	if l.current < 0 || l.current >= len(l.items) {
		return
	}
	item := l.items[l.current]
	if item.selected != nil {
		item.selected()
	}
	if l.onSelect != nil {
		l.onSelect(l.current, item.main, item.secondary, item.shortcut)
	}
}
