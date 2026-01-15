package components

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTabs_NewTabs tests Tabs creation.
func TestTabs_NewTabs(t *testing.T) {
	tabs := NewTabs()

	assert.NotNil(t, tabs)
	assert.Equal(t, 0, tabs.TabCount())
	assert.Equal(t, 0, tabs.GetActive())
}

// TestTabs_AddTab tests adding tabs.
func TestTabs_AddTab(t *testing.T) {
	tabs := NewTabs()
	content := tview.NewBox()

	result := tabs.AddTab("Tab 1", content)

	assert.Same(t, tabs, result) // Fluent API
	assert.Equal(t, 1, tabs.TabCount())
}

// TestTabs_AddTabWithIcon tests adding tabs with icons.
func TestTabs_AddTabWithIcon(t *testing.T) {
	tabs := NewTabs()
	content := tview.NewBox()

	result := tabs.AddTabWithIcon("Files", "📁", content)

	assert.Same(t, tabs, result)
	assert.Equal(t, 1, tabs.TabCount())

	tab := tabs.GetActiveTab()
	require.NotNil(t, tab)
	assert.Equal(t, "Files", tab.Name)
	assert.Equal(t, "📁", tab.Icon)
}

// TestTabs_RemoveTab tests removing tabs.
func TestTabs_RemoveTab(t *testing.T) {
	tabs := NewTabs()
	tabs.AddTab("Tab 1", nil)
	tabs.AddTab("Tab 2", nil)
	tabs.AddTab("Tab 3", nil)
	require.Equal(t, 3, tabs.TabCount())

	t.Run("remove middle tab", func(t *testing.T) {
		tabs := NewTabs()
		tabs.AddTab("Tab 1", nil)
		tabs.AddTab("Tab 2", nil)
		tabs.AddTab("Tab 3", nil)
		tabs.SetActive(1)

		result := tabs.RemoveTab(1)

		assert.Same(t, tabs, result)
		assert.Equal(t, 2, tabs.TabCount())
	})

	t.Run("remove last tab when active", func(t *testing.T) {
		tabs := NewTabs()
		tabs.AddTab("Tab 1", nil)
		tabs.AddTab("Tab 2", nil)
		tabs.SetActive(1)

		tabs.RemoveTab(1)

		assert.Equal(t, 0, tabs.GetActive()) // Should adjust to last available
	})

	t.Run("remove invalid index", func(t *testing.T) {
		tabs := NewTabs()
		tabs.AddTab("Tab 1", nil)

		tabs.RemoveTab(-1)
		tabs.RemoveTab(10)

		assert.Equal(t, 1, tabs.TabCount()) // Should remain unchanged
	})
}

// TestTabs_SetActive tests setting active tab by index.
func TestTabs_SetActive(t *testing.T) {
	tabs := NewTabs()
	tabs.AddTab("Tab 1", nil)
	tabs.AddTab("Tab 2", nil)
	tabs.AddTab("Tab 3", nil)

	t.Run("valid index", func(t *testing.T) {
		result := tabs.SetActive(1)

		assert.Same(t, tabs, result)
		assert.Equal(t, 1, tabs.GetActive())
	})

	t.Run("invalid index", func(t *testing.T) {
		tabs.SetActive(1)
		tabs.SetActive(-1)
		assert.Equal(t, 1, tabs.GetActive()) // Should not change

		tabs.SetActive(10)
		assert.Equal(t, 1, tabs.GetActive()) // Should not change
	})
}

// TestTabs_SetActiveByName tests setting active tab by name.
func TestTabs_SetActiveByName(t *testing.T) {
	tabs := NewTabs()
	tabs.AddTab("Files", nil)
	tabs.AddTab("Search", nil)
	tabs.AddTab("Settings", nil)

	t.Run("existing name", func(t *testing.T) {
		result := tabs.SetActiveByName("Search")

		assert.Same(t, tabs, result)
		assert.Equal(t, 1, tabs.GetActive())
	})

	t.Run("non-existing name", func(t *testing.T) {
		tabs.SetActive(0)
		tabs.SetActiveByName("NotFound")

		assert.Equal(t, 0, tabs.GetActive()) // Should not change
	})
}

// TestTabs_GetActiveTab tests getting the active tab.
func TestTabs_GetActiveTab(t *testing.T) {
	t.Run("with tabs", func(t *testing.T) {
		tabs := NewTabs()
		tabs.AddTab("Tab 1", nil)
		tabs.AddTab("Tab 2", nil)
		tabs.SetActive(1)

		tab := tabs.GetActiveTab()

		require.NotNil(t, tab)
		assert.Equal(t, "Tab 2", tab.Name)
	})

	t.Run("empty tabs", func(t *testing.T) {
		tabs := NewTabs()

		tab := tabs.GetActiveTab()

		assert.Nil(t, tab)
	})
}

// TestTabs_SetBadge tests setting badges on tabs.
func TestTabs_SetBadge(t *testing.T) {
	tabs := NewTabs()
	tabs.AddTab("Inbox", nil)
	tabs.AddTab("Sent", nil)

	result := tabs.SetBadge("Inbox", 5)

	assert.Same(t, tabs, result)

	tab := tabs.GetActiveTab()
	require.NotNil(t, tab)
	assert.Equal(t, 5, tab.Badge)
}

// TestTabs_ClearBadge tests clearing badges.
func TestTabs_ClearBadge(t *testing.T) {
	tabs := NewTabs()
	tabs.AddTab("Inbox", nil)
	tabs.SetBadge("Inbox", 5)

	result := tabs.ClearBadge("Inbox")

	assert.Same(t, tabs, result)

	tab := tabs.GetActiveTab()
	require.NotNil(t, tab)
	assert.Equal(t, 0, tab.Badge)
}

// TestTabs_SetShowIcons tests icon display toggle.
func TestTabs_SetShowIcons(t *testing.T) {
	tabs := NewTabs()

	result := tabs.SetShowIcons(false)
	assert.Same(t, tabs, result)

	result = tabs.SetShowIcons(true)
	assert.Same(t, tabs, result)
}

// TestTabs_SetShowBadges tests badge display toggle.
func TestTabs_SetShowBadges(t *testing.T) {
	tabs := NewTabs()

	result := tabs.SetShowBadges(false)
	assert.Same(t, tabs, result)
}

// TestTabs_SetClosable tests closable tabs setting.
func TestTabs_SetClosable(t *testing.T) {
	tabs := NewTabs()

	result := tabs.SetClosable(true)
	assert.Same(t, tabs, result)
}

// TestTabs_SetOnChange tests change callback.
func TestTabs_SetOnChange(t *testing.T) {
	tabs := NewTabs()

	var calledIndex int
	var calledName string
	result := tabs.SetOnChange(func(index int, name string) {
		calledIndex = index
		calledName = name
	})

	assert.Same(t, tabs, result)

	tabs.AddTab("Tab 1", nil)
	tabs.AddTab("Tab 2", nil)
	tabs.SetActive(1)

	assert.Equal(t, 1, calledIndex)
	assert.Equal(t, "Tab 2", calledName)
}

// TestTabs_SetOnClose tests close callback.
func TestTabs_SetOnClose(t *testing.T) {
	tabs := NewTabs()

	result := tabs.SetOnClose(func(index int) bool {
		return true
	})

	assert.Same(t, tabs, result)
}

// TestTabs_TabNavigation tests keyboard navigation.
func TestTabs_TabNavigation(t *testing.T) {
	tabs := NewTabs()
	tabs.AddTab("Tab 1", nil)
	tabs.AddTab("Tab 2", nil)
	tabs.AddTab("Tab 3", nil)

	handler := tabs.InputHandler()

	t.Run("Tab key cycles forward", func(t *testing.T) {
		tabs.SetActive(0)
		handler(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), nil)
		assert.Equal(t, 1, tabs.GetActive())

		handler(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), nil)
		assert.Equal(t, 2, tabs.GetActive())

		handler(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), nil)
		assert.Equal(t, 0, tabs.GetActive()) // Wraps around
	})

	t.Run("BackTab cycles backward", func(t *testing.T) {
		tabs.SetActive(0)
		handler(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone), nil)
		assert.Equal(t, 2, tabs.GetActive()) // Wraps to end
	})

	t.Run("number keys select tabs directly", func(t *testing.T) {
		tabs.SetActive(0)
		handler(tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone), nil)
		assert.Equal(t, 1, tabs.GetActive()) // 0-indexed

		handler(tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone), nil)
		assert.Equal(t, 0, tabs.GetActive())
	})

	t.Run("H and L for navigation", func(t *testing.T) {
		tabs.SetActive(1)
		handler(tcell.NewEventKey(tcell.KeyRune, 'H', tcell.ModNone), nil)
		assert.Equal(t, 0, tabs.GetActive())

		handler(tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone), nil)
		assert.Equal(t, 1, tabs.GetActive())
	})
}

// TestTabs_CloseTab tests closing tabs with 'x' key.
func TestTabs_CloseTab(t *testing.T) {
	t.Run("closable tabs", func(t *testing.T) {
		tabs := NewTabs()
		tabs.SetClosable(true)
		tabs.AddTab("Tab 1", nil)
		tabs.AddTab("Tab 2", nil)
		tabs.SetActive(0)

		handler := tabs.InputHandler()
		handler(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), nil)

		assert.Equal(t, 1, tabs.TabCount())
	})

	t.Run("non-closable tabs ignore x", func(t *testing.T) {
		tabs := NewTabs()
		tabs.SetClosable(false)
		tabs.AddTab("Tab 1", nil)
		tabs.AddTab("Tab 2", nil)

		handler := tabs.InputHandler()
		handler(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), nil)

		assert.Equal(t, 2, tabs.TabCount()) // No change
	})

	t.Run("onClose can prevent closing", func(t *testing.T) {
		tabs := NewTabs()
		tabs.SetClosable(true)
		tabs.SetOnClose(func(index int) bool {
			return false // Prevent closing
		})
		tabs.AddTab("Tab 1", nil)

		handler := tabs.InputHandler()
		handler(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), nil)

		assert.Equal(t, 1, tabs.TabCount()) // Should not be closed
	})
}

// TestTabs_FluentAPI tests method chaining.
func TestTabs_FluentAPI(t *testing.T) {
	var changeCalled bool

	tabs := NewTabs().
		AddTab("Files", nil).
		AddTabWithIcon("Search", "🔍", nil).
		SetShowIcons(true).
		SetShowBadges(true).
		SetClosable(true).
		SetOnChange(func(index int, name string) {
			changeCalled = true
		}).
		SetOnClose(func(index int) bool {
			return true
		}).
		SetBadge("Files", 3).
		SetActive(1)

	assert.Equal(t, 2, tabs.TabCount())
	assert.Equal(t, 1, tabs.GetActive())
	assert.True(t, changeCalled)
}

// TestTabs_EmptyTabs tests operations on empty tabs.
func TestTabs_EmptyTabs(t *testing.T) {
	tabs := NewTabs()

	// Should not panic
	tabs.SetActive(0)
	tabs.SetActiveByName("NotFound")
	tabs.RemoveTab(0)
	tabs.SetBadge("NotFound", 5)
	tabs.ClearBadge("NotFound")

	handler := tabs.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), nil)
	handler(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone), nil)

	assert.Equal(t, 0, tabs.TabCount())
	assert.Nil(t, tabs.GetActiveTab())
}

// TestTabs_CalcTabWidth tests tab width calculation.
func TestTabs_CalcTabWidth(t *testing.T) {
	tabs := NewTabs()

	t.Run("basic tab", func(t *testing.T) {
		tabs.SetShowIcons(false)
		tabs.SetShowBadges(false)
		tabs.SetClosable(false)

		tab := &Tab{Name: "Test"}
		width := tabs.calcTabWidth(tab)

		// Should be name length + padding (2)
		assert.Equal(t, len("Test")+2, width)
	})

	t.Run("tab with icon", func(t *testing.T) {
		tabs.SetShowIcons(true)
		tabs.SetShowBadges(false)
		tabs.SetClosable(false)

		tab := &Tab{Name: "Test", Icon: "📁"}
		width := tabs.calcTabWidth(tab)

		// Should include icon + space
		assert.Greater(t, width, len("Test")+2)
	})

	t.Run("tab with badge", func(t *testing.T) {
		tabs.SetShowIcons(false)
		tabs.SetShowBadges(true)
		tabs.SetClosable(false)

		tab := &Tab{Name: "Test", Badge: 5}
		width := tabs.calcTabWidth(tab)

		// Should include badge
		assert.Greater(t, width, len("Test")+2)
	})
}
