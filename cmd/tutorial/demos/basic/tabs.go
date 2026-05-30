package basic

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&TabsDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Tabs",
			DemoDescription: "Tabbed container with icons and badges",
			DemoCategory:    demos.Basic,
			DemoCode:        tabsCode,
		},
	})
}

// TabsDemo demonstrates the Tabs component.
type TabsDemo struct {
	demos.DemoBase
	tabs      *components.Tabs
	showIcons bool
	closable  bool
}

// Component returns the demo component.
func (d *TabsDemo) Component() tview.Primitive {
	d.showIcons = true
	d.closable = false

	d.tabs = components.NewTabs().
		SetShowIcons(d.showIcons).
		SetClosable(d.closable)

	tab1 := d.createContent("Welcome to the first tab!")
	tab2 := d.createContent("This is the second tab content.")
	tab3 := d.createContent("And here's the third tab.")

	d.tabs.AddTabWithIcon("Home", theme.IconHome, tab1)
	d.tabs.AddTabWithIcon("Files", theme.IconFolder, tab2)
	d.tabs.AddTabWithIcon("Settings", theme.IconSettings, tab3)
	d.tabs.SetBadge("Files", 3)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showIcons", "Show tab icons",
			func() bool { return d.showIcons },
			func(v bool) { d.showIcons = v; d.tabs.SetShowIcons(v) },
			true,
		),
		demos.BoolProp("closable", "Allow closing tabs",
			func() bool { return d.closable },
			func(v bool) { d.closable = v; d.tabs.SetClosable(v) },
			false,
		),
	}
	d.ResetFunc = func() {
		d.tabs.SetActive(0)
	}

	return d.tabs
}

func (d *TabsDemo) createContent(text string) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetText(text)
	tv.SetTextAlign(tview.AlignCenter)
	tv.SetBackgroundColor(theme.Bg())
	tv.SetTextColor(theme.Fg())
	return tv
}

const tabsCode = `package main

import (
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/theme"
)

// Create tabs container
tabs := components.NewTabs().
    SetShowIcons(true).
    SetShowBadges(true).
    SetClosable(false)

// Add tabs with icons
tabs.AddTabWithIcon("Home", theme.IconHome, homeContent)
tabs.AddTabWithIcon("Files", theme.IconFolder, filesContent)
tabs.AddTab("Settings", settingsContent)

// Set badge on a tab
tabs.SetBadge("Files", 5)

// Switch tabs programmatically
tabs.SetActive(0)
tabs.SetActiveByName("Files")

// Listen for tab changes
tabs.SetOnChange(func(index int, name string) {
    fmt.Printf("Switched to tab: %s\n", name)
})

// Handle tab close (return false to prevent)
tabs.SetOnClose(func(index int) bool {
    return true // Allow close
})
`
