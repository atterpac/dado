package intermediate

import (
	"fmt"


	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&VirtualListDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "VirtualList",
			DemoDescription: "Efficient virtualized list",
			DemoCategory:    demos.Intermediate,
			DemoCode:        virtualListCode,
		},
	})
}

// VirtualListDemo demonstrates the VirtualList component.
type VirtualListDemo struct {
	demos.DemoBase
	list          *components.VirtualList
	showScrollbar bool
	showIndex     bool
}

// Component returns the demo component.
func (d *VirtualListDemo) Component() core.Widget {
	d.showScrollbar = true
	d.showIndex = true

	d.list = components.NewVirtualList().
		SetShowScrollbar(d.showScrollbar).
		SetShowIndex(d.showIndex)

	// Create sample items
	items := make([]components.VirtualListItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = components.VirtualListItem{
			ID:   fmt.Sprintf("item-%d", i),
			Data: fmt.Sprintf("List item %d - This is sample content", i+1),
		}
	}

	d.list.SetItems(items)

	// Custom render function
	d.list.SetRenderFunc(func(index int, item components.VirtualListItem, width int, selected bool) string {
		text := item.Data.(string)
		if selected {
			return fmt.Sprintf("> %s", text)
		}
		return fmt.Sprintf("  %s", text)
	})

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showScrollbar", "Show scrollbar",
			func() bool { return d.showScrollbar },
			func(v bool) { d.showScrollbar = v; d.list.SetShowScrollbar(v) },
			true,
		),
		demos.BoolProp("showIndex", "Show item index",
			func() bool { return d.showIndex },
			func(v bool) { d.showIndex = v; d.list.SetShowIndex(v) },
			true,
		),
	}

	return d.list
}

const virtualListCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create virtual list
list := components.NewVirtualList().
    SetShowScrollbar(true).
    SetShowIndex(true).
    SetDefaultItemHeight(1)

// Option 1: Set all items directly (for smaller lists)
items := []components.VirtualListItem{
    {ID: "1", Data: "First item"},
    {ID: "2", Data: "Second item"},
}
list.SetItems(items)

// Option 2: Lazy loading for large datasets
list.SetTotalCount(10000)
list.SetFetchFunc(func(start, count int) ([]components.VirtualListItem, int) {
    items := fetchFromDatabase(start, count)
    return items, totalCount
})

// Custom rendering
list.SetRenderFunc(func(index int, item components.VirtualListItem, width int, selected bool) string {
    prefix := "  "
    if selected {
        prefix = "> "
    }
    return prefix + item.Data.(string)
})

// Callbacks
list.SetOnSelect(func(index int, item components.VirtualListItem) {
    fmt.Printf("Selected: %v\n", item.Data)
})

list.SetOnChange(func(index int, item components.VirtualListItem) {
    // Highlight changed
})

// Navigation
list.Select(5)     // Select item at index 5
list.ScrollTo(100) // Scroll to item 100
`
