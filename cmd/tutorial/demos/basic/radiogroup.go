package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&RadioGroupDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "RadioGroup",
			DemoDescription: "Single choice option group",
			DemoCategory:    demos.Basic,
			DemoCode:        radioGroupCode,
		},
	})
}

// RadioGroupDemo demonstrates the RadioGroup component.
type RadioGroupDemo struct {
	demos.DemoBase
	radioGroup *components.RadioGroup
	label      string
}

// Component returns the demo component.
func (d *RadioGroupDemo) Component() core.Widget {
	d.label = "Select size"

	d.radioGroup = components.NewRadioGroup("size").
		SetLabel(d.label).
		SetOptions([]string{"Small", "Medium", "Large", "X-Large"}).
		SetSelected(1)

	d.Props = []demos.PropertyDescriptor{
		demos.StringProp("label", "The group label",
			func() string { return d.label },
			func(v string) { d.label = v; d.radioGroup.SetLabel(v) },
			"Select size",
		),
	}
	d.ResetFunc = func() {
		d.radioGroup.SetSelected(1)
	}

	return d.radioGroup
}

const radioGroupCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create a radio group
radio := components.NewRadioGroup("size").
    SetLabel("Select size").
    SetOptions([]string{"Small", "Medium", "Large"}).
    SetSelected(1) // Select "Medium" by default

// Listen for changes
radio.SetOnChange(func(event *components.ChangeEvent[string]) {
    fmt.Printf("Selected: %s\n", event.NewValue)
})

// Get selected value
value := radio.Value()

// Get selected index
index := radio.SelectedIndex()
`
