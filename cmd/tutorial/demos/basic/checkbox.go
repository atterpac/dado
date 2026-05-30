package basic

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
)

func init() {
	demos.Register(&CheckboxDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Checkbox",
			DemoDescription: "Boolean toggle input",
			DemoCategory:    demos.Basic,
			DemoCode:        checkboxCode,
		},
	})
}

// CheckboxDemo demonstrates the Checkbox component.
type CheckboxDemo struct {
	demos.DemoBase
	checkbox *components.Checkbox
	label    string
	checked  bool
}

// Component returns the demo component.
func (d *CheckboxDemo) Component() tview.Primitive {
	d.label = "Enable notifications"
	d.checked = false

	d.checkbox = components.NewCheckbox("demo").
		SetLabel(d.label).
		SetChecked(d.checked)

	// Setup property descriptors (must be done after component exists)
	d.Props = []demos.PropertyDescriptor{
		demos.StringProp("label", "The checkbox label text",
			func() string { return d.label },
			func(v string) { d.label = v; d.checkbox.SetLabel(v) },
			"Enable notifications",
		),
		demos.BoolProp("checked", "Initial checked state",
			func() bool { return d.checked },
			func(v bool) { d.checked = v; d.checkbox.SetChecked(v) },
			false,
		),
	}

	return d.checkbox
}

const checkboxCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create a checkbox
checkbox := components.NewCheckbox("notifications").
    SetLabel("Enable notifications").
    SetChecked(false)

// Listen for changes
checkbox.SetOnChange(func(checked bool) {
    if checked {
        fmt.Println("Notifications enabled")
    } else {
        fmt.Println("Notifications disabled")
    }
})

// Toggle programmatically
checkbox.Toggle()

// Get current state
isEnabled := checkbox.Checked()
`
