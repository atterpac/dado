package core

import "github.com/gdamore/tcell/v2"

type listItem struct {
	main      string
	secondary string
	shortcut  rune
	selected  func()
}

// List is a scrollable selectable list.
type List struct {
	Box
	items    []listItem
	current  int
	offset   int

	wrapAround    bool
	showSecondary bool

	mainTextColor         tcell.Color
	secondaryTextColor    tcell.Color
	selectedBgColor       tcell.Color
	selectedTextColor     tcell.Color
	shortcutColor         tcell.Color

	onChange  func(cur, prev int, main, secondary string, shortcut rune)
	onSelect  func(idx int, main, secondary string, shortcut rune)
}

// NewList returns an empty List.
func NewList() *List { return &List{} }

// AddItem appends an item. shortcut 0 = no shortcut. selected is called on Enter.
func (l *List) AddItem(main, secondary string, shortcut rune, selected func()) *List {
	l.items = append(l.items, listItem{main, secondary, shortcut, selected})
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
	x, y, w, h := l.InnerRect()
	if w <= 0 || h <= 0 {
		return
	}

	// Ensure current item is visible (scroll offset adjustment)
	if l.current < l.offset {
		l.offset = l.current
	}
	if l.current >= l.offset+h {
		l.offset = l.current - h + 1
	}

	for row := 0; row < h; row++ {
		idx := l.offset + row
		if idx >= len(l.items) {
			break
		}
		item := l.items[idx]
		style := tcell.StyleDefault
		if idx == l.current {
			style = style.Reverse(true)
		}
		// Fill row background, then render text with tag parsing
		for col := 0; col < w; col++ {
			screen.SetContent(x+col, y+row, ' ', nil, style)
		}
		PrintTagged(screen, item.main, x, y+row, w, style)
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
