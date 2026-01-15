package basic

import (
	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
)

func init() {
	demos.Register(&TextFieldDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "TextField",
			DemoDescription: "Single-line text input",
			DemoCategory:    demos.Basic,
			DemoCode:        textFieldCode,
		},
	})
}

// TextFieldDemo demonstrates the TextField component.
type TextFieldDemo struct {
	demos.DemoBase
	textField   *components.TextField
	label       string
	placeholder string
}

// Component returns the demo component.
func (d *TextFieldDemo) Component() tview.Primitive {
	d.label = "Username"
	d.placeholder = "Enter your username"

	d.textField = components.NewTextField("username").
		SetLabel(d.label).
		SetPlaceholder(d.placeholder).
		SetValue("")

	d.Props = []demos.PropertyDescriptor{
		demos.StringProp("label", "Field label",
			func() string { return d.label },
			func(v string) { d.label = v; d.textField.SetLabel(v) },
			"Username",
		),
		demos.StringProp("placeholder", "Placeholder text",
			func() string { return d.placeholder },
			func(v string) { d.placeholder = v; d.textField.SetPlaceholder(v) },
			"Enter your username",
		),
	}
	d.ResetFunc = func() {
		d.textField.SetValue("")
	}

	return d.textField
}

const textFieldCode = `package main

import (
    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/validators"
)

// Create a text field
textField := components.NewTextField("username").
    SetLabel("Username").
    SetPlaceholder("Enter your username").
    SetValue("")

// Add validation
textField.SetValidator(func(value string) error {
    return validators.MinLength(3)(value)
})

// Listen for changes
textField.SetOnChange(func(value string) {
    fmt.Printf("Current value: %s\n", value)
})

// Listen for submit (Enter key)
textField.SetOnSubmit(func(value string) {
    fmt.Printf("Submitted: %s\n", value)
})

// Get current value
value := textField.GetValue()
`
