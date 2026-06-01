package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&DividerDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Divider",
			DemoDescription: "Horizontal/vertical separators",
			DemoCategory:    demos.Basic,
			DemoCode:        dividerCode,
		},
	})
}

// DividerDemo demonstrates the Divider component.
type DividerDemo struct {
	demos.DemoBase
	container *core.Flex
	divider   *components.Divider
	showLabel bool
}

// Component returns the demo component.
func (d *DividerDemo) Component() core.Widget {
	d.showLabel = true

	d.container = core.NewFlex()

	d.divider = components.NewDivider().SetLabel("Section")
	simple := components.NewDivider()

	text1 := core.NewTextView().SetText("Content above divider")
	text2 := core.NewTextView().SetText("Content between dividers")
	text3 := core.NewTextView().SetText("Content below divider")

	d.container.AddItem(text1, 2, 0, false)
	d.container.AddItem(d.divider, 1, 0, false)
	d.container.AddItem(text2, 2, 0, false)
	d.container.AddItem(simple, 1, 0, false)
	d.container.AddItem(text3, 2, 0, false)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showLabel", "Show label in divider",
			func() bool { return d.showLabel },
			func(v bool) {
				d.showLabel = v
				if v {
					d.divider.SetLabel("Section")
				} else {
					d.divider.SetLabel("")
				}
			},
			true,
		),
	}

	return d.container
}

const dividerCode = `package main


// Simple horizontal divider
divider := components.NewDivider()

// Divider with centered label
divider.SetLabel("Section Title")

// Custom character
divider.SetStyle('═')

// Vertical divider
vertical := components.NewVerticalDivider()

// Or change orientation
divider.SetOrientation(components.DividerVertical)
`
