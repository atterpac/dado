package advanced

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&ProgressModalDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "ProgressModal",
			DemoDescription: "Progress indicator dialog",
			DemoCategory:    demos.Advanced,
			DemoCode:        progressModalCode,
		},
	})
}

// ProgressModalDemo demonstrates the ProgressModal component.
type ProgressModalDemo struct {
	demos.DemoBase
	modal      *components.ProgressModal
	cancelable bool
}

// Component returns the demo component.
func (d *ProgressModalDemo) Component() tview.Primitive {
	d.cancelable = true

	d.modal = components.NewProgressModal().
		SetTitle("Processing").
		SetMessage("Please wait...").
		SetCancelable(d.cancelable).
		SetWidth(50).
		SetProgress(0.6).
		SetShowBackdrop(false) // Disable for demo so background shows

	// Background
	bg := tview.NewTextView()
	bg.SetText("Progress Modal Demo\n\nShows a progress indicator for long operations")
	bg.SetTextAlign(tview.AlignCenter)
	bg.SetBackgroundColor(theme.Bg())
	bg.SetTextColor(theme.FgDim())

	pages := tview.NewPages()
	pages.AddPage("bg", bg, true, true)
	pages.AddPage("modal", d.modal, true, true)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("cancelable", "Allow cancellation",
			func() bool { return d.cancelable },
			func(v bool) { d.cancelable = v; d.modal.SetCancelable(v) },
			true,
		),
	}
	d.ResetFunc = func() {
		d.modal.SetProgress(0.6)
	}

	return pages
}

const progressModalCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create progress modal
modal := components.NewProgressModal().
    SetTitle("Downloading").
    SetMessage("Fetching data...").
    SetWidth(50).
    SetCancelable(true).
    SetShowBackdrop(true)

// Determinate progress (0.0 to 1.0)
modal.SetProgress(0.5) // 50%

// Indeterminate mode (spinner)
modal.SetIndeterminate(true)

// Update message during operation
modal.SetMessage("Processing file 1 of 10...")
modal.SetSubMessage("file.txt")

// Mark complete
modal.Complete("Download finished!")

// Mark failed
modal.Fail(errors.New("connection lost"))

// Callbacks
modal.SetOnCancel(func() {
    // User pressed Escape
    cancelOperation()
})

modal.SetOnComplete(func() {
    // Operation completed successfully
    closeModal()
})

// Show with app pages
pages.Push(modal)

// Start spinner animation (requires app reference)
modal.StartSpinner(app)
`
