package basic

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
)

func init() {
	demos.Register(&SelectDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Select",
			DemoDescription: "Dropdown selection input",
			DemoCategory:    demos.Basic,
			DemoCode:        selectCode,
		},
	})
}

// SelectDemo demonstrates the Select component.
type SelectDemo struct {
	demos.DemoBase
	selectComp *components.Select
	label      string
}

// Component returns the demo component.
func (d *SelectDemo) Component() tview.Primitive {
	d.label = "Choose a color"

	d.selectComp = components.NewSelect("color").
		SetLabel(d.label).
		SetOptions([]string{"Red", "Green", "Blue", "Yellow", "Purple"}).
		SetPlaceholder("Select a color...")

	d.Props = []demos.PropertyDescriptor{
		demos.StringProp("label", "The field label",
			func() string { return d.label },
			func(v string) { d.label = v; d.selectComp.SetLabel(v) },
			"Choose a color",
		),
	}
	d.ResetFunc = func() {
		d.selectComp.Clear()
	}

	return d.selectComp
}

const selectCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create a select dropdown
sel := components.NewSelect("color").
    SetLabel("Choose a color").
    SetOptions([]string{"Red", "Green", "Blue"}).
    SetPlaceholder("Select a color...")

// With custom label/value pairs
sel.SetOptionsWithValues([]components.SelectOption{
    {Label: "Red", Value: "red"},
    {Label: "Green", Value: "green"},
    {Label: "Blue", Value: "blue"},
})

// Listen for changes
sel.SetOnChange(func(event *components.ChangeEvent[components.SelectOption]) {
    fmt.Printf("Selected: %s\n", event.NewValue.Label)
})

// Set default selection
sel.SetDefault("green")

// Get current value
value := sel.GetValue()
`
