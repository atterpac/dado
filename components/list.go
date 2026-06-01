package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// ListItem represents an item in a List.
type ListItem struct {
	Text      string
	Secondary string // Optional secondary text
	Reference any    // Optional reference data
}

// List is a simple list component with selection support.
// It wraps core.List with themed defaults and a cleaner API.
type List struct {
	*core.List

	items    []ListItem
	onSelect func(index int, item ListItem)
	onChange func(index int, item ListItem)
	subs     Subscriptions
}

// Subs returns the widget's subscription set. ComponentBase.Stop releases
// it automatically when the List is wrapped; standalone callers should
// call Subs().Release() at teardown to drop the theme registration.
func (l *List) Subs() *Subscriptions { return &l.subs }

// NewList creates a new List.
func NewList() *List {
	list := core.NewList()
	list.SetBackgroundColor(theme.Bg())
	list.SetMainTextColor(theme.Fg())
	list.SetSecondaryTextColor(theme.FgDim())
	list.SetSelectedBackgroundColor(theme.Accent())
	list.SetSelectedTextColor(theme.Bg())
	list.SetShortcutColor(theme.AccentDim())
	list.ShowSecondaryText(false)

	l := &List{
		List:  list,
		items: make([]ListItem, 0),
	}

	// Wire up selection handlers
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if l.onSelect != nil && index >= 0 && index < len(l.items) {
			l.onSelect(index, l.items[index])
		}
	})

	list.SetChangedFunc(func(cur, prev int, mainText, secondaryText string, shortcut rune) {
		if l.onChange != nil && cur >= 0 && cur < len(l.items) {
			l.onChange(cur, l.items[cur])
		}
	})

	l.subs.Add(theme.RegisterFn(func(c tcell.Color) { list.SetBackgroundColor(c) }))

	return l
}

// AddItem adds an item to the list.
func (l *List) AddItem(text string) *List {
	item := ListItem{Text: text}
	l.items = append(l.items, item)
	l.List.AddItem(text, "", 0, nil)
	return l
}

// AddItemWithSecondary adds an item with secondary text.
func (l *List) AddItemWithSecondary(text, secondary string) *List {
	item := ListItem{Text: text, Secondary: secondary}
	l.items = append(l.items, item)
	l.List.AddItem(text, secondary, 0, nil)
	l.List.ShowSecondaryText(true)
	return l
}

// AddItemWithRef adds an item with a reference object.
func (l *List) AddItemWithRef(text string, ref any) *List {
	item := ListItem{Text: text, Reference: ref}
	l.items = append(l.items, item)
	l.List.AddItem(text, "", 0, nil)
	return l
}

// AddItems adds multiple items at once.
func (l *List) AddItems(texts ...string) *List {
	for _, text := range texts {
		l.AddItem(text)
	}
	return l
}

// SetItems replaces all items.
func (l *List) SetItems(items []ListItem) *List {
	l.Clear()
	hasSecondary := false
	for _, item := range items {
		l.items = append(l.items, item)
		l.List.AddItem(item.Text, item.Secondary, 0, nil)
		if item.Secondary != "" {
			hasSecondary = true
		}
	}
	l.List.ShowSecondaryText(hasSecondary)
	return l
}

// Clear removes all items.
func (l *List) Clear() *List {
	l.items = make([]ListItem, 0)
	l.List.Clear()
	return l
}

// GetItem returns the item at the given index.
func (l *List) GetItem(index int) (ListItem, bool) {
	if index >= 0 && index < len(l.items) {
		return l.items[index], true
	}
	return ListItem{}, false
}

// GetItems returns all items.
func (l *List) GetItems() []ListItem {
	return l.items
}

// GetItemCount returns the number of items.
func (l *List) GetItemCount() int {
	return len(l.items)
}

// GetSelected returns the currently selected item.
func (l *List) GetSelected() (int, ListItem, bool) {
	index := l.List.GetCurrentItem()
	if index >= 0 && index < len(l.items) {
		return index, l.items[index], true
	}
	return -1, ListItem{}, false
}

// SetSelected sets the selected item by index.
func (l *List) SetSelected(index int) *List {
	l.List.SetCurrentItem(index)
	return l
}

// SetOnSelect sets the handler called when an item is selected (Enter pressed).
func (l *List) SetOnSelect(handler func(index int, item ListItem)) *List {
	l.onSelect = handler
	return l
}

// SetOnChange sets the handler called when the selection changes (navigation).
func (l *List) SetOnChange(handler func(index int, item ListItem)) *List {
	l.onChange = handler
	return l
}

// SetShowSecondary enables or disables secondary text display.
func (l *List) SetShowSecondary(show bool) *List {
	l.List.ShowSecondaryText(show)
	return l
}

// SetWrapAround enables or disables wrap-around navigation.
func (l *List) SetWrapAround(wrap bool) *List {
	l.List.SetWrapAround(wrap)
	return l
}

// SetHighlightFullLine enables or disables full-line highlighting.
func (l *List) SetHighlightFullLine(full bool) *List {
	l.List.SetHighlightFullLine(full)
	return l
}

// MoveUp moves the selection up.
func (l *List) MoveUp() {
	current := l.List.GetCurrentItem()
	if current > 0 {
		l.List.SetCurrentItem(current - 1)
	}
}

// MoveDown moves the selection down.
func (l *List) MoveDown() {
	current := l.List.GetCurrentItem()
	if current < l.List.GetItemCount()-1 {
		l.List.SetCurrentItem(current + 1)
	}
}

// MoveToTop moves the selection to the first item.
func (l *List) MoveToTop() {
	l.List.SetCurrentItem(0)
}

// MoveToBottom moves the selection to the last item.
func (l *List) MoveToBottom() {
	count := l.List.GetItemCount()
	if count > 0 {
		l.List.SetCurrentItem(count - 1)
	}
}

// HandleKey processes a key event for the List, supporting vim-style navigation.
func (l *List) HandleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j':
			l.MoveDown()
			return true
		case 'k':
			l.MoveUp()
			return true
		case 'g':
			l.MoveToTop()
			return true
		case 'G':
			l.MoveToBottom()
			return true
		}
	}
	// Delegate everything else to core.List's built-in handler (Enter, arrows, etc.)
	return l.List.HandleKey(ev)
}
