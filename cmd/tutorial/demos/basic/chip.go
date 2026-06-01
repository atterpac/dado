package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&ChipDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Chip",
			DemoDescription: "Removable tag elements",
			DemoCategory:    demos.Basic,
			DemoCode:        chipCode,
		},
	})
}

// ChipDemo demonstrates the Chip component.
type ChipDemo struct {
	demos.DemoBase
	container *core.Flex
	chips     []*components.Chip
	removable bool
}

// Component returns the demo component.
func (d *ChipDemo) Component() core.Widget {
	d.removable = true

	d.chips = []*components.Chip{
		components.NewChip("Go").SetIcon("").SetRemovable(true),
		components.NewChip("Rust").SetIcon("").SetRemovable(true),
		components.NewChip("Python").SetIcon("").SetRemovable(true).SetSelected(true),
		components.NewChip("TypeScript").SetRemovable(true),
	}

	d.container = core.NewFlex()

	row := core.NewFlex().SetDirection(core.Row)
	for _, chip := range d.chips {
		row.AddItem(chip, chip.Width()+2, 0, false)
	}

	d.container.AddItem(nil, 0, 1, false)
	d.container.AddItem(row, 1, 0, false)
	d.container.AddItem(nil, 0, 1, false)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("removable", "Show remove button",
			func() bool { return d.removable },
			func(v bool) {
				d.removable = v
				for _, chip := range d.chips {
					chip.SetRemovable(v)
				}
			},
			true,
		),
	}

	return d.container
}

const chipCode = `package main


// Create a basic chip
chip := components.NewChip("Go")

// Add an icon
chip.SetIcon("")

// Make it removable
chip.SetRemovable(true)

// Handle removal
chip.SetOnRemove(func() {
    // Remove chip from container
})

// Select/deselect
chip.SetSelected(true)

// Handle clicks
chip.SetOnClick(func() {
    chip.SetSelected(!chip.IsSelected())
})

// Disable the chip
chip.SetDisabled(true)
`
