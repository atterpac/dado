package intermediate

import (
	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&ModalDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Modal",
			DemoDescription: "Centered dialog overlay",
			DemoCategory:    demos.Intermediate,
			DemoCode:        modalCode,
		},
	})
}

// ModalDemo demonstrates the Modal component.
type ModalDemo struct {
	demos.DemoBase
	container *core.Flex
	modal     *components.Modal
	backdrop  bool
}

// Component returns the demo component.
func (d *ModalDemo) Component() core.Widget {
	d.backdrop = false // Disabled for demo so background shows

	// Create a container to show the modal on
	d.container = core.NewFlex()
	d.container.SetDirection(core.Column)

	// Background content
	bg := core.NewTextView()
	bg.SetText("Background content behind the modal.\nThis simulates content that would be dimmed.")
	bg.SetBackgroundColor(theme.Bg())

	// Create modal
	d.modal = components.NewModal(components.ModalConfig{
		Title:    "Example Modal",
		Width:    40,
		Height:   12,
		Backdrop: d.backdrop,
	})

	content := core.NewTextView()
	content.SetText("This is modal content.\n\nModals are centered overlays\nfor dialogs and prompts.")
	content.SetBackgroundColor(theme.Bg())

	d.modal.SetContent(content)
	d.modal.SetHints([]components.KeyHint{
		{Key: "Esc", Description: "Close"},
		{Key: "Enter", Description: "Confirm"},
	})

	// Layer modal on top
	pages := core.NewPages()
	pages.AddPage("bg", bg, true, true)
	pages.AddPage("modal", d.modal, true, true)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("backdrop", "Show dimmed backdrop",
			func() bool { return d.backdrop },
			func(v bool) {
				// Note: backdrop requires recreating the modal
				// This is a limitation of the demo - in real usage
				// you'd set this at creation time
				d.backdrop = v
			},
			false,
		),
	}
	d.ResetFunc = func() {
		d.backdrop = false
	}

	return pages
}

const modalCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create a modal
modal := components.NewModal(components.ModalConfig{
    Title:     "Confirm Action",
    Width:     50,
    Height:    15,
    MinWidth:  30,
    MinHeight: 10,
    Backdrop:  true,
})

// Set content
modal.SetContent(myForm)

// Set key hints shown at bottom
modal.SetHints([]components.KeyHint{
    {Key: "Enter", Description: "Submit"},
    {Key: "Esc", Description: "Cancel"},
})

// Configure behavior
modal.SetDismissOnEsc(true)
modal.SetFocusOnShow(myForm)

// Callbacks
modal.SetOnClose(func() {
    // Modal was closed
})

modal.SetOnSubmit(func() {
    // Enter was pressed
})

// Use with nav.Pages for proper lifecycle
pages.Push(modal)
`
