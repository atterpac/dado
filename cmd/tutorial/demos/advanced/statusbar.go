package advanced

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&StatusBarDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "StatusBar",
			DemoDescription: "Application status bar with sections",
			DemoCategory:    demos.Advanced,
			DemoCode:        statusBarCode,
		},
	})
}

// StatusBarDemo demonstrates the StatusBar component.
type StatusBarDemo struct {
	demos.DemoBase
	statusbar  *components.StatusBar
	showBorder bool
}

// Component returns the demo component.
func (d *StatusBarDemo) Component() core.Widget {
	d.showBorder = true

	d.statusbar = components.NewStatusBar().
		SetLeft(
			components.StatusSection{Text: "main", Icon: "", Color: theme.Success()},
			components.StatusSection{Text: "src/app.go", Icon: ""},
		).
		SetCenter(
			components.StatusSection{Text: "Ln 42, Col 15"},
		).
		SetRight(
			components.StatusSection{Text: "Go", Icon: ""},
			components.StatusSection{Text: "UTF-8"},
			components.StatusSection{Text: "LF"},
		).
		SetShowBorder(true)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showBorder", "Show top border",
			func() bool { return d.showBorder },
			func(v bool) { d.showBorder = v; d.statusbar.SetShowBorder(v) },
			true,
		),
	}

	return d.statusbar
}

const statusBarCode = `package main

import (
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/theme"
)

// Create a status bar
status := components.NewStatusBar()

// Left-aligned sections
status.SetLeft(
    components.StatusSection{
        Text: "main",
        Icon: "",
        Color: theme.Success(),
    },
    components.StatusSection{
        Text: "src/app.go",
        Icon: "",
    },
)

// Center sections
status.SetCenter(
    components.StatusSection{Text: "Ln 42, Col 15"},
)

// Right-aligned sections
status.SetRight(
    components.StatusSection{Text: "Go", Icon: ""},
    components.StatusSection{Text: "UTF-8"},
)

// Add individual sections
status.AddLeft(components.StatusSection{Text: "Modified"})
status.AddRight(components.StatusSection{Text: "INS"})

// Update a section
status.UpdateSection("Ln", "Ln 100, Col 1")

// Show top border
status.SetShowBorder(true)

// Custom separator
status.SetSeparator('|')

// Clear all
status.Clear()
`
