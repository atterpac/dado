package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&PanelDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Panel",
			DemoDescription: "Bordered container with title",
			DemoCategory:    demos.Basic,
			DemoCode:        panelCode,
		},
	})
}

// PanelDemo demonstrates the Panel component.
type PanelDemo struct {
	demos.DemoBase
	panel      *components.Panel
	title      string
	titleAlign string
}

// Component returns the demo component.
func (d *PanelDemo) Component() core.Widget {
	d.title = "Panel Title"
	d.titleAlign = "center"

	content := core.NewTextView()
	content.SetText("This is the panel content.\nPanels provide a bordered container with an optional title.")
	content.SetBackgroundColor(theme.Bg())

	d.panel = components.NewPanel().
		SetTitle(d.title).
		SetTitleAlign(components.AlignCenter).
		SetContent(content)

	d.Props = []demos.PropertyDescriptor{
		demos.StringProp("title", "Panel title text",
			func() string { return d.title },
			func(v string) { d.title = v; d.panel.SetTitle(v) },
			"Panel Title",
		),
		demos.SelectProp("titleAlign", "Title alignment",
			[]string{"left", "center", "right"},
			func() string { return d.titleAlign },
			func(v string) {
				d.titleAlign = v
				switch v {
				case "left":
					d.panel.SetTitleAlign(components.AlignLeft)
				case "right":
					d.panel.SetTitleAlign(components.AlignRight)
				default:
					d.panel.SetTitleAlign(components.AlignCenter)
				}
			},
			"center",
		),
	}

	return d.panel
}

const panelCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create a panel
panel := components.NewPanel().
    SetTitle("My Panel").
    SetTitleAlign(components.AlignLeft).
    SetContent(myContent)

// Change title color
panel.SetTitleColor(tcell.ColorBlue)

// Set focus state (changes border color)
panel.SetFocused(true)

// Get the content back
content := panel.GetContent()
`
