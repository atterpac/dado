package advanced

import (
	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&DrawerDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Drawer",
			DemoDescription: "Slide-out side panel",
			DemoCategory:    demos.Advanced,
			DemoCode:        drawerCode,
		},
	})
}

// DrawerDemo demonstrates the Drawer component.
type DrawerDemo struct {
	demos.DemoBase
	drawer   *components.Drawer
	position string
}

// Component returns the demo component.
func (d *DrawerDemo) Component() core.Widget {
	d.position = "right"

	d.drawer = components.NewDrawer(components.DrawerConfig{
		Title:    "Settings",
		Width:    35,
		Position: components.DrawerRight,
		Backdrop: false, // Disabled for demo so background shows
	})

	// Drawer content
	content := core.NewTextView()
	content.SetText("Drawer Content\n\nDrawers slide in from screen edges.\nUseful for settings, details, or navigation.")
	content.SetBackgroundColor(theme.Bg())

	d.drawer.SetContent(content)
	d.drawer.SetHints([]components.KeyHint{
		{Key: "Esc", Description: "Close"},
	})

	// Background
	bg := core.NewTextView()
	bg.SetText("Drawer Demo\n\nThe drawer slides in from the right edge")
	bg.SetBackgroundColor(theme.Bg())

	pages := core.NewPages()
	pages.AddPage("bg", bg, true, true)
	pages.AddPage("drawer", d.drawer, true, true)

	d.Props = []demos.PropertyDescriptor{
		demos.SelectProp("position", "Drawer position", []string{"right", "left"},
			func() string { return d.position },
			func(v string) {
				// Position requires recreating the drawer
				// This is a limitation - shown for demo purposes
			},
			"right",
		),
	}

	return pages
}

const drawerCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create drawer (right edge)
drawer := components.NewDrawer(components.DrawerConfig{
    Title:    "Settings",
    Width:    40,
    Position: components.DrawerRight,
    Backdrop: true,
})

// Or left edge
drawer := components.NewDrawer(components.DrawerConfig{
    Title:    "Navigation",
    Width:    30,
    Position: components.DrawerLeft,
    Backdrop: false,
})

// Set content
drawer.SetContent(settingsForm)

// Set key hints
drawer.SetHints([]components.KeyHint{
    {Key: "Esc", Description: "Close"},
    {Key: "Enter", Description: "Save"},
})

// Dismiss on Escape
drawer.SetDismissOnEsc(true)

// Focus specific element when opened
drawer.SetFocusOnShow(firstField)

// Callbacks
drawer.SetOnClose(func() {
    saveSettings()
})

drawer.SetOnDismiss(func() bool {
    // Return false to prevent closing
    return confirmDiscard()
})

// Show with nav.Pages
pages.Push(drawer)
`
