package advanced

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&ContextMenuDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "ContextMenu",
			DemoDescription: "Right-click popup menu",
			DemoCategory:    demos.Advanced,
			DemoCode:        contextMenuCode,
		},
	})
}

// ContextMenuDemo demonstrates the ContextMenu component.
type ContextMenuDemo struct {
	demos.DemoBase
	container *tview.Flex
	menu      *components.ContextMenu
}

// Component returns the demo component.
func (d *ContextMenuDemo) Component() tview.Primitive {
	d.container = tview.NewFlex()
	d.container.SetDirection(tview.FlexRow)

	// Create sample menu
	d.menu = components.NewContextMenu()

	d.menu.AddItemWithIcon("new", "New File", theme.IconFile, func() {})
	d.menu.AddItemWithIcon("open", "Open...", theme.IconFolder, func() {})
	d.menu.AddDivider()
	d.menu.AddItemWithShortcut("cut", "Cut", "Ctrl+X", func() {})
	d.menu.AddItemWithShortcut("copy", "Copy", "Ctrl+C", func() {})
	d.menu.AddItemWithShortcut("paste", "Paste", "Ctrl+V", func() {})
	d.menu.AddDivider()
	d.menu.AddItem("delete", "Delete", func() {})
	d.menu.SetDanger("delete", true)

	// Show the menu
	d.menu.ShowAt(5, 2)

	// Background with instructions
	bg := tview.NewTextView()
	bg.SetText("Context Menu Demo\n\nUse j/k to navigate, Enter to select")
	bg.SetTextAlign(tview.AlignCenter)
	bg.SetBackgroundColor(theme.Bg())
	bg.SetTextColor(theme.FgDim())

	// Layer menu on top
	pages := tview.NewPages()
	pages.AddPage("bg", bg, true, true)
	pages.AddPage("menu", d.menu, true, true)

	return pages
}

const contextMenuCode = `package main

import (
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/theme"
)

// Create context menu
menu := components.NewContextMenu()

// Simple items
menu.AddItem("save", "Save", func() {
    saveFile()
})

// With keyboard shortcut display
menu.AddItemWithShortcut("copy", "Copy", "Ctrl+C", func() {
    copyToClipboard()
})

// With icon
menu.AddItemWithIcon("new", "New File", theme.IconFile, func() {
    createNewFile()
})

// Divider
menu.AddDivider()

// Dangerous action (red text)
menu.AddDangerItem("delete", "Delete", func() {
    deleteFile()
})

// Toggle/checkbox item
menu.AddToggleItem("wrap", "Word Wrap", isWrapped, func() {
    toggleWrap()
})

// Disabled item
menu.AddDisabledItem("locked", "Locked Feature")

// Show at position
menu.Show(x, y)

// Hide menu
menu.Hide()

// Callbacks
menu.SetOnSelect(func(item components.MenuItem) {
    fmt.Printf("Selected: %s\n", item.Label)
})

menu.SetOnClose(func() {
    // Menu was closed
})
`
