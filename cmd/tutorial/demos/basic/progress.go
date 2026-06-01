package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&ProgressDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Progress",
			DemoDescription: "Horizontal progress bar",
			DemoCategory:    demos.Basic,
			DemoCode:        progressCode,
		},
	})
}

// ProgressDemo demonstrates the ProgressBar component.
type ProgressDemo struct {
	demos.DemoBase
	progress    *components.ProgressBar
	showPercent bool
}

// Component returns the demo component.
func (d *ProgressDemo) Component() core.Widget {
	d.showPercent = true

	d.progress = components.NewProgressBar().
		SetProgress(0.65).
		SetShowPercentage(d.showPercent).
		SetLabel("Download Progress")

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showPercent", "Show percentage label",
			func() bool { return d.showPercent },
			func(v bool) { d.showPercent = v; d.progress.SetShowPercentage(v) },
			true,
		),
	}
	d.ResetFunc = func() {
		d.progress.SetProgress(0.65)
	}

	return d.progress
}

const progressCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create a progress bar
progress := components.NewProgress().
    SetLabel("Download Progress").
    SetValue(0.65).           // 0.0 to 1.0
    SetShowPercentage(true)

// Update progress
progress.SetValue(0.75)

// Set to complete
progress.SetValue(1.0)

// Show value instead of percent
progress.SetShowValue(true)
progress.SetMaxValue(100)
progress.SetValue(75)  // Will show "75/100"
`
