package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&SpinnerDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Spinner",
			DemoDescription: "Animated loading indicator",
			DemoCategory:    demos.Basic,
			DemoCode:        spinnerCode,
		},
	})
}

// SpinnerDemo demonstrates the Spinner component.
type SpinnerDemo struct {
	demos.DemoBase
	spinner *components.Spinner
}

// Component returns the demo component.
func (d *SpinnerDemo) Component() core.Widget {
	d.spinner = components.NewSpinner().
		SetLabel("Loading...").
		SetStyle(components.SpinnerDots)
	d.spinner.Start()

	// Layout with multiple spinners showing different styles
	layout := core.NewFlex()
	layout.SetBackgroundColor(theme.Bg())

	// Add spinner
	layout.AddItem(d.spinner, 1, 0, false)

	// Show different styles
	styles := core.NewTextView()
	styles.SetDynamicColors(true)
	styles.SetBackgroundColor(theme.Bg())
	styles.SetText("\n[" + theme.TagFgDim() + "]Spinner Styles: Dots, Line, Circle, Bounce[-]")

	layout.AddItem(styles, 3, 0, false)

	// Wrap in panel
	panel := components.NewPanel()
	panel.SetTitle("Spinner")
	panel.SetContent(layout)

	d.ResetFunc = func() {
		d.spinner.Stop()
		d.spinner.Start()
	}

	return panel
}

const spinnerCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create spinner
spinner := components.NewSpinner().
    SetLabel("Loading...").
    SetStyle(components.SpinnerDots)

// Available styles
spinner.SetStyle(components.SpinnerDots)    // ⣾ ⣽ ⣻ ...
spinner.SetStyle(components.SpinnerLine)    // | / - \
spinner.SetStyle(components.SpinnerCircle)  // ◐ ◓ ◑ ◒
spinner.SetStyle(components.SpinnerBounce)  // ⠁ ⠂ ⠄ ...

// Control animation
spinner.Start()
spinner.Stop()

// Check state
if spinner.IsRunning() {
    // ...
}

// Set animation speed
spinner.SetInterval(100 * time.Millisecond)
`
