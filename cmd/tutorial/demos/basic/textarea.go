package basic

import (
	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
)

func init() {
	demos.Register(&TextAreaDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "TextArea",
			DemoDescription: "Multi-line text input",
			DemoCategory:    demos.Basic,
			DemoCode:        textAreaCode,
		},
	})
}

// TextAreaDemo demonstrates the TextArea component.
type TextAreaDemo struct {
	demos.DemoBase
	textarea *components.TextArea
}

// Component returns the demo component.
func (d *TextAreaDemo) Component() tview.Primitive {
	d.textarea = components.NewTextArea("description")
	d.textarea.SetLabel("Description")
	d.textarea.SetPlaceholder("Enter a detailed description...")
	d.textarea.SetValue("This is a multi-line text area.\n\nYou can type multiple lines of text here.\n\nUse arrow keys to navigate.")

	panel := components.NewPanel()
	panel.SetTitle("TextArea")
	panel.SetContent(d.textarea)

	d.ResetFunc = func() {
		d.textarea.SetValue("")
	}

	return panel
}

const textAreaCode = `package main

import (
    "github.com/atterpac/jig/components"
)

// Create text area
textarea := components.NewTextArea()

// Configure
textarea.SetLabel("Description")
textarea.SetPlaceholder("Enter description...")
textarea.SetText("Initial content\nwith multiple lines")

// Set dimensions
textarea.SetRows(5)  // Visible rows

// Get content
text := textarea.GetText()

// Callbacks
textarea.SetOnChange(func(text string) {
    updatePreview(text)
})

// Validation
textarea.SetMaxLength(500)
`
