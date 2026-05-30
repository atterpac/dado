package advanced

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/input"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&CommandBarDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "CommandBar",
			DemoDescription: "K9s-style command/filter input",
			DemoCategory:    demos.Advanced,
			DemoCode:        commandBarCode,
		},
	})
}

// CommandBarDemo demonstrates the CommandBar component.
type CommandBarDemo struct {
	demos.DemoBase
}

// Component returns the demo component.
func (d *CommandBarDemo) Component() tview.Primitive {
	// Create command bar
	cmdBar := input.NewCommandBar()
	cmdBar.Show(input.CommandTypeFilter)

	// Status text to show what's happening
	status := tview.NewTextView()
	status.SetDynamicColors(true)
	status.SetTextAlign(tview.AlignCenter)
	status.SetBackgroundColor(theme.Bg())
	status.SetText("[" + theme.TagFgDim() + "]Type to filter, Enter to submit, Esc to cancel[-]")

	cmdBar.SetOnChange(func(text string) {
		status.SetText("[" + theme.TagFg() + "]Filtering: [" + theme.TagAccent() + "]" + text + "[-]")
	})

	cmdBar.SetOnSubmit(func(cmdType input.CommandType, text string) {
		status.SetText("[" + theme.TagSuccess() + "]Submitted: " + text + "[-]")
	})

	// Instructions
	instructions := tview.NewTextView()
	instructions.SetDynamicColors(true)
	instructions.SetTextAlign(tview.AlignCenter)
	instructions.SetBackgroundColor(theme.Bg())
	instructions.SetText("CommandBar Demo\n\n" +
		"[" + theme.TagFgDim() + "]Supports different modes:[-]\n" +
		"[" + theme.TagAccent() + "]/[-] Filter mode\n" +
		"[" + theme.TagAccent() + "]:[-] Command mode\n" +
		"[" + theme.TagAccent() + "]?[-] Search mode")

	// Layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.SetBackgroundColor(theme.Bg())
	layout.AddItem(instructions, 0, 1, false)
	layout.AddItem(status, 3, 0, false)
	layout.AddItem(cmdBar, 1, 0, true)

	// Wrap in panel
	panel := components.NewPanel()
	panel.SetTitle("CommandBar")
	panel.SetContent(layout)

	return panel
}

const commandBarCode = `package main

import (
    "github.com/atterpac/dado/input"
)

// Create command bar
cmdBar := input.NewCommandBar()

// Configure command types (optional - defaults provided)
cmdBar.Configure(input.CommandTypeFilter, input.CommandTypeConfig{
    Prompt:      "/",
    Placeholder: "Filter...",
})
cmdBar.Configure(input.CommandTypeAction, input.CommandTypeConfig{
    Prompt:      ":",
    Placeholder: "Command...",
})
cmdBar.Configure(input.CommandTypeSearch, input.CommandTypeConfig{
    Prompt:      "?",
    Placeholder: "Search...",
})

// Show with a specific mode
cmdBar.Show(input.CommandTypeFilter)  // Shows "/"
cmdBar.Show(input.CommandTypeAction)  // Shows ":"
cmdBar.Show(input.CommandTypeSearch)  // Shows "?"

// Callbacks
cmdBar.SetOnSubmit(func(cmdType input.CommandType, text string) {
    switch cmdType {
    case input.CommandTypeFilter:
        applyFilter(text)
    case input.CommandTypeAction:
        executeCommand(text)
    case input.CommandTypeSearch:
        searchFor(text)
    }
})

cmdBar.SetOnCancel(func() {
    cmdBar.Hide()
})

// Filter-as-you-type
cmdBar.SetOnChange(func(text string) {
    liveFilter(text)
})

// Hide when done
cmdBar.Hide()
`
