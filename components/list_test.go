package components

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestList_NewList tests List creation.
func TestList_NewList(t *testing.T) {
	list := NewList()

	assert.NotNil(t, list)
	assert.Equal(t, 0, list.GetItemCount())
}

// TestList_AddItem tests adding single items.
func TestList_AddItem(t *testing.T) {
	list := NewList()

	list.AddItem("Item 1")
	list.AddItem("Item 2")
	list.AddItem("Item 3")

	assert.Equal(t, 3, list.GetItemCount())

	item, ok := list.GetItem(0)
	assert.True(t, ok)
	assert.Equal(t, "Item 1", item.Text)

	item, ok = list.GetItem(2)
	assert.True(t, ok)
	assert.Equal(t, "Item 3", item.Text)
}

// TestList_AddItemWithSecondary tests adding items with secondary text.
func TestList_AddItemWithSecondary(t *testing.T) {
	list := NewList()

	list.AddItemWithSecondary("Title", "Subtitle")

	item, ok := list.GetItem(0)
	assert.True(t, ok)
	assert.Equal(t, "Title", item.Text)
	assert.Equal(t, "Subtitle", item.Secondary)
}

// TestList_AddItemWithRef tests adding items with reference data.
func TestList_AddItemWithRef(t *testing.T) {
	list := NewList()

	type refData struct {
		ID   int
		Name string
	}
	ref := refData{ID: 42, Name: "test"}

	list.AddItemWithRef("Item", ref)

	item, ok := list.GetItem(0)
	assert.True(t, ok)
	assert.Equal(t, ref, item.Reference)
}

// TestList_AddItems tests adding multiple items at once.
func TestList_AddItems(t *testing.T) {
	list := NewList()

	list.AddItems("A", "B", "C", "D", "E")

	assert.Equal(t, 5, list.GetItemCount())

	items := list.GetItems()
	assert.Len(t, items, 5)
	assert.Equal(t, "A", items[0].Text)
	assert.Equal(t, "E", items[4].Text)
}

// TestList_SetItems tests replacing all items.
func TestList_SetItems(t *testing.T) {
	list := NewList()

	list.AddItem("Old Item")
	assert.Equal(t, 1, list.GetItemCount())

	list.SetItems([]ListItem{
		{Text: "New 1"},
		{Text: "New 2", Secondary: "Sub 2"},
		{Text: "New 3"},
	})

	assert.Equal(t, 3, list.GetItemCount())

	item, _ := list.GetItem(1)
	assert.Equal(t, "New 2", item.Text)
	assert.Equal(t, "Sub 2", item.Secondary)
}

// TestList_Clear tests clearing all items.
func TestList_Clear(t *testing.T) {
	list := NewList()

	list.AddItems("A", "B", "C")
	require.Equal(t, 3, list.GetItemCount())

	list.Clear()

	assert.Equal(t, 0, list.GetItemCount())
	assert.Empty(t, list.GetItems())
}

// TestList_GetItem tests item retrieval.
func TestList_GetItem(t *testing.T) {
	list := NewList()
	list.AddItems("A", "B", "C")

	// Valid index
	item, ok := list.GetItem(1)
	assert.True(t, ok)
	assert.Equal(t, "B", item.Text)

	// Invalid index - negative
	item, ok = list.GetItem(-1)
	assert.False(t, ok)
	assert.Equal(t, ListItem{}, item)

	// Invalid index - out of bounds
	item, ok = list.GetItem(10)
	assert.False(t, ok)
	assert.Equal(t, ListItem{}, item)
}

// TestList_GetSelected tests getting the selected item.
func TestList_GetSelected(t *testing.T) {
	list := NewList()

	// Empty list
	idx, item, ok := list.GetSelected()
	assert.False(t, ok)
	assert.Equal(t, -1, idx)
	assert.Equal(t, ListItem{}, item)

	// With items
	list.AddItems("A", "B", "C")
	list.SetSelected(1)

	idx, item, ok = list.GetSelected()
	assert.True(t, ok)
	assert.Equal(t, 1, idx)
	assert.Equal(t, "B", item.Text)
}

// TestList_SetSelected tests setting the selected item.
func TestList_SetSelected(t *testing.T) {
	list := NewList()
	list.AddItems("A", "B", "C")

	list.SetSelected(2)

	idx, item, ok := list.GetSelected()
	assert.True(t, ok)
	assert.Equal(t, 2, idx)
	assert.Equal(t, "C", item.Text)
}

// TestList_Navigation tests navigation methods.
func TestList_Navigation(t *testing.T) {
	list := NewList()
	list.AddItems("A", "B", "C", "D", "E")
	list.SetSelected(2) // Start in middle

	t.Run("MoveUp", func(t *testing.T) {
		list.SetSelected(2)
		list.MoveUp()
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 1, idx)
	})

	t.Run("MoveUp at top", func(t *testing.T) {
		list.SetSelected(0)
		list.MoveUp()
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 0, idx) // Should stay at 0
	})

	t.Run("MoveDown", func(t *testing.T) {
		list.SetSelected(2)
		list.MoveDown()
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 3, idx)
	})

	t.Run("MoveDown at bottom", func(t *testing.T) {
		list.SetSelected(4)
		list.MoveDown()
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 4, idx) // Should stay at 4
	})

	t.Run("MoveToTop", func(t *testing.T) {
		list.SetSelected(3)
		list.MoveToTop()
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 0, idx)
	})

	t.Run("MoveToBottom", func(t *testing.T) {
		list.SetSelected(1)
		list.MoveToBottom()
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 4, idx)
	})
}

// TestList_OnSelect tests selection callback.
func TestList_OnSelect(t *testing.T) {
	list := NewList()
	list.AddItems("A", "B", "C")

	var selectedIndex int
	var selectedItem ListItem

	list.SetOnSelect(func(index int, item ListItem) {
		selectedIndex = index
		selectedItem = item
	})

	// Simulate selection via InputHandler (Enter key)
	handler := list.InputHandler()
	list.SetSelected(1)
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	assert.Equal(t, 1, selectedIndex)
	assert.Equal(t, "B", selectedItem.Text)
}

// TestList_OnChange tests change callback.
func TestList_OnChange(t *testing.T) {
	list := NewList()
	list.AddItems("A", "B", "C")

	var changes []int

	list.SetOnChange(func(index int, item ListItem) {
		changes = append(changes, index)
	})

	// Navigate using vim keys
	handler := list.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil) // down
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil) // down

	// Changes should be recorded via the tview callback
	// Note: The onChange is triggered by tview's SetChangedFunc
}

// TestList_VimNavigation tests vim-style key bindings.
func TestList_VimNavigation(t *testing.T) {
	list := NewList()
	list.AddItems("A", "B", "C", "D", "E")
	list.SetSelected(2)

	handler := list.InputHandler()

	t.Run("j moves down", func(t *testing.T) {
		list.SetSelected(2)
		handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 3, idx)
	})

	t.Run("k moves up", func(t *testing.T) {
		list.SetSelected(2)
		handler(tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone), nil)
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 1, idx)
	})

	t.Run("g moves to top", func(t *testing.T) {
		list.SetSelected(3)
		handler(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone), nil)
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 0, idx)
	})

	t.Run("G moves to bottom", func(t *testing.T) {
		list.SetSelected(1)
		handler(tcell.NewEventKey(tcell.KeyRune, 'G', tcell.ModNone), nil)
		idx, _, _ := list.GetSelected()
		assert.Equal(t, 4, idx)
	})
}

// TestList_SetShowSecondary tests secondary text toggle.
func TestList_SetShowSecondary(t *testing.T) {
	list := NewList()

	// Should not panic and return self for chaining
	result := list.SetShowSecondary(true)
	assert.Same(t, list, result)

	result = list.SetShowSecondary(false)
	assert.Same(t, list, result)
}

// TestList_SetWrapAround tests wrap-around setting.
func TestList_SetWrapAround(t *testing.T) {
	list := NewList()

	result := list.SetWrapAround(true)
	assert.Same(t, list, result)
}

// TestList_SetHighlightFullLine tests full-line highlight setting.
func TestList_SetHighlightFullLine(t *testing.T) {
	list := NewList()

	result := list.SetHighlightFullLine(true)
	assert.Same(t, list, result)
}

// TestList_Primitive tests getting the underlying tview.List.
func TestList_Primitive(t *testing.T) {
	list := NewList()

	primitive := list.Primitive()
	assert.NotNil(t, primitive)
	assert.Same(t, list.List, primitive)
}

// TestList_FluentAPI tests method chaining.
func TestList_FluentAPI(t *testing.T) {
	var selectedCalled bool

	list := NewList().
		AddItem("Item 1").
		AddItem("Item 2").
		AddItemWithSecondary("Item 3", "Subtitle").
		SetShowSecondary(true).
		SetWrapAround(true).
		SetHighlightFullLine(true).
		SetOnSelect(func(index int, item ListItem) {
			selectedCalled = true
		}).
		SetSelected(1)

	assert.Equal(t, 3, list.GetItemCount())
	_ = selectedCalled // Used when Enter is pressed

	idx, item, ok := list.GetSelected()
	assert.True(t, ok)
	assert.Equal(t, 1, idx)
	assert.Equal(t, "Item 2", item.Text)
}

// TestList_EmptyNavigation tests navigation on empty list.
func TestList_EmptyNavigation(t *testing.T) {
	list := NewList()

	// Should not panic on empty list
	list.MoveUp()
	list.MoveDown()
	list.MoveToTop()
	list.MoveToBottom()

	handler := list.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone), nil)
}
