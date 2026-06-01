package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&BadgeDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Badge",
			DemoDescription: "Small label/count indicators",
			DemoCategory:    demos.Basic,
			DemoCode:        badgeCode,
		},
	})
}

// BadgeDemo demonstrates the Badge component.
type BadgeDemo struct {
	demos.DemoBase
	container *core.Flex
	badges    []*components.Badge
	pill      bool
}

// Component returns the demo component.
func (d *BadgeDemo) Component() core.Widget {
	d.pill = true

	d.badges = []*components.Badge{
		components.NewBadge("Default").SetVariant(components.BadgeDefault).SetPill(true),
		components.NewBadge("Primary").SetVariant(components.BadgePrimary).SetPill(true),
		components.NewBadge("Success").SetVariant(components.BadgeSuccess).SetPill(true),
		components.NewBadge("Warning").SetVariant(components.BadgeWarning).SetPill(true),
		components.NewBadge("Error").SetVariant(components.BadgeError).SetPill(true),
		components.NewBadge("99+").SetVariant(components.BadgeInfo).SetIcon("").SetPill(true),
	}

	d.container = core.NewFlex()

	row := core.NewFlex().SetDirection(core.Row)
	for _, badge := range d.badges {
		row.AddItem(badge, badge.Width()+2, 0, false)
	}

	d.container.AddItem(nil, 0, 1, false)
	d.container.AddItem(row, 1, 0, false)
	d.container.AddItem(nil, 0, 1, false)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("pill", "Use pill (rounded) style",
			func() bool { return d.pill },
			func(v bool) {
				d.pill = v
				for _, badge := range d.badges {
					badge.SetPill(v)
				}
			},
			true,
		),
	}

	return d.container
}

const badgeCode = `package main


// Create badges with different variants
defaultBadge := components.NewBadge("Default").
    SetVariant(components.BadgeDefault)

successBadge := components.NewBadge("Success").
    SetVariant(components.BadgeSuccess)

errorBadge := components.NewBadge("3").
    SetVariant(components.BadgeError).
    SetIcon("!")  // Optional icon

// Pill style (rounded)
badge.SetPill(true)

// Available variants:
// - BadgeDefault
// - BadgePrimary
// - BadgeSuccess
// - BadgeWarning
// - BadgeError
// - BadgeInfo
`
