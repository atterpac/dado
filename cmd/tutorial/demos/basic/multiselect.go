package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&MultiSelectDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "MultiSelect",
			DemoDescription: "Multiple choice selection",
			DemoCategory:    demos.Basic,
			DemoCode:        multiSelectCode,
		},
	})
}

// MultiSelectDemo demonstrates the MultiSelect component.
type MultiSelectDemo struct {
	demos.DemoBase
	multiSelect *components.MultiSelect
	label       string
}

// Component returns the demo component.
func (d *MultiSelectDemo) Component() core.Widget {
	d.label = "Select toppings"

	d.multiSelect = components.NewMultiSelect("toppings").
		SetLabel(d.label).
		SetOptions([]string{"Cheese", "Pepperoni", "Mushrooms", "Onions", "Peppers"}).
		SetSelected([]int{0, 1})

	d.Props = []demos.PropertyDescriptor{
		demos.StringProp("label", "The field label",
			func() string { return d.label },
			func(v string) { d.label = v; d.multiSelect.SetLabel(v) },
			"Select toppings",
		),
	}
	d.ResetFunc = func() {
		d.multiSelect.SetSelected([]int{0, 1})
	}

	return d.multiSelect
}

const multiSelectCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create a multi-select
ms := components.NewMultiSelect("toppings").
    SetLabel("Select toppings").
    SetOptions([]string{"Cheese", "Pepperoni", "Mushrooms"})

// Pre-select some options by index
ms.SetSelected([]int{0, 1})

// Or by value
ms.SetSelectedValues([]string{"Cheese", "Pepperoni"})

// Listen for changes
ms.SetOnChange(func(event *components.ChangeEvent[[]components.SelectOption]) {
    for _, opt := range event.NewValue {
        fmt.Printf("Selected: %s\n", opt.Label)
    }
})

// Get selected values
values := ms.Values()

// Get selected indices
indices := ms.SelectedIndices()
`
