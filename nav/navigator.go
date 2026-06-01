package nav

import (
	"github.com/gdamore/tcell/v2"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

// ListNavigator provides standard list navigation behavior.
// Implement this interface for components that support j/k navigation.
type ListNavigator interface {
	// MoveUp moves selection up one item
	MoveUp()

	// MoveDown moves selection down one item
	MoveDown()

	// MoveToTop moves selection to first item
	MoveToTop()

	// MoveToBottom moves selection to last item
	MoveToBottom()

	// GetSelectedIndex returns current selection index
	GetSelectedIndex() int

	// SetSelectedIndex sets selection to specific index
	SetSelectedIndex(index int)

	// GetItemCount returns total number of items
	GetItemCount() int
}

// TableNavigator implements ListNavigator for Table components.
type TableNavigator struct {
	table     *components.Table
	hasHeader bool // Skip header row in navigation
}

// NewTableNavigator creates a navigator for a Table.
func NewTableNavigator(table *components.Table, hasHeader bool) *TableNavigator {
	return &TableNavigator{
		table:     table,
		hasHeader: hasHeader,
	}
}

func (n *TableNavigator) MoveUp() {
	current := n.GetSelectedIndex()
	minRow := 0
	if n.hasHeader {
		minRow = 1
	}
	if current > minRow {
		n.SetSelectedIndex(current - 1)
	}
}

func (n *TableNavigator) MoveDown() {
	current := n.GetSelectedIndex()
	maxRow := n.GetItemCount() - 1
	if current < maxRow {
		n.SetSelectedIndex(current + 1)
	}
}

func (n *TableNavigator) MoveToTop() {
	minRow := 0
	if n.hasHeader {
		minRow = 1
	}
	n.SetSelectedIndex(minRow)
}

func (n *TableNavigator) MoveToBottom() {
	maxRow := n.GetItemCount() - 1
	if maxRow >= 0 {
		n.SetSelectedIndex(maxRow)
	}
}

func (n *TableNavigator) GetSelectedIndex() int {
	row, _ := n.table.GetSelection()
	return row
}

func (n *TableNavigator) SetSelectedIndex(index int) {
	n.table.Select(index, 0)
}

func (n *TableNavigator) GetItemCount() int {
	return n.table.GetRowCount()
}

// TextViewNavigator implements ListNavigator for scrollable TextView.
type TextViewNavigator struct {
	textView *core.TextView
}

// NewTextViewNavigator creates a navigator for a TextView.
func NewTextViewNavigator(tv *core.TextView) *TextViewNavigator {
	return &TextViewNavigator{
		textView: tv,
	}
}

func (n *TextViewNavigator) MoveUp() {
	row, col := n.textView.GetScrollOffset()
	if row > 0 {
		n.textView.ScrollTo(row-1, col)
	}
}

func (n *TextViewNavigator) MoveDown() {
	row, col := n.textView.GetScrollOffset()
	n.textView.ScrollTo(row+1, col)
}

func (n *TextViewNavigator) MoveToTop() {
	_, col := n.textView.GetScrollOffset()
	n.textView.ScrollTo(0, col)
}

func (n *TextViewNavigator) MoveToBottom() {
	_, col := n.textView.GetScrollOffset()
	n.textView.ScrollTo(999999, col)
}

func (n *TextViewNavigator) GetSelectedIndex() int {
	row, _ := n.textView.GetScrollOffset()
	return row
}

func (n *TextViewNavigator) SetSelectedIndex(index int) {
	_, col := n.textView.GetScrollOffset()
	n.textView.ScrollTo(index, col)
}

func (n *TextViewNavigator) GetItemCount() int {
	return 999999
}

// NavigationInputHandler returns an input handler for standard navigation keys.
// Returns true if the key was handled.
func NavigationInputHandler(nav ListNavigator) func(*tcell.EventKey) bool {
	return func(event *tcell.EventKey) bool {
		switch event.Key() {
		case tcell.KeyUp:
			nav.MoveUp()
			return true
		case tcell.KeyDown:
			nav.MoveDown()
			return true
		case tcell.KeyHome:
			nav.MoveToTop()
			return true
		case tcell.KeyEnd:
			nav.MoveToBottom()
			return true
		case tcell.KeyPgUp:
			for i := 0; i < 10; i++ {
				nav.MoveUp()
			}
			return true
		case tcell.KeyPgDn:
			for i := 0; i < 10; i++ {
				nav.MoveDown()
			}
			return true
		}

		switch event.Rune() {
		case 'j':
			nav.MoveDown()
			return true
		case 'k':
			nav.MoveUp()
			return true
		case 'g':
			nav.MoveToTop()
			return true
		case 'G':
			nav.MoveToBottom()
			return true
		}

		return false
	}
}
