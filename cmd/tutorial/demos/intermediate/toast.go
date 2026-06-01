package intermediate

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&ToastDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Toast",
			DemoDescription: "Notification popups",
			DemoCategory:    demos.Intermediate,
			DemoCode:        toastCode,
		},
	})
}

// ToastDemo demonstrates the Toast component.
type ToastDemo struct {
	demos.DemoBase
	container *core.Flex
	level     string
}

// Component returns the demo component.
func (d *ToastDemo) Component() core.Widget {
	d.level = "info"

	// Create a visual representation of toasts
	d.container = core.NewFlex()
	d.container.SetDirection(core.Column)

	// Show sample toasts
	levels := []struct {
		name    string
		icon    string
		message string
	}{
		{"Info", "i", "This is an info notification"},
		{"Success", "v", "Operation completed successfully"},
		{"Warning", "!", "Please review before continuing"},
		{"Error", "x", "An error occurred"},
	}

	for _, l := range levels {
		toast := d.createToastPreview(l.icon, l.name, l.message)
		d.container.AddItem(toast, 3, 0, false)
	}

	note := core.NewTextView()
	note.SetText("\nToasts require a ToastManager connected to the app.\nSee code example for usage.")
	note.SetBackgroundColor(theme.Bg())
	d.container.AddItem(note, 0, 1, false)

	d.Props = []demos.PropertyDescriptor{
		demos.SelectProp("level", "Toast severity level",
			[]string{"info", "success", "warning", "error"},
			func() string { return d.level },
			func(v string) { d.level = v },
			"info",
		),
	}

	return d.container
}

func (d *ToastDemo) createToastPreview(icon, level, message string) *core.Flex {
	row := core.NewFlex()
	row.SetDirection(core.Row)

	// Icon
	iconView := core.NewTextView()
	iconView.SetText(" " + icon + " ")
	iconView.SetBackgroundColor(theme.Bg())

	// Message
	msgView := core.NewTextView()
	msgView.SetText(message)
	msgView.SetBackgroundColor(theme.Bg())

	row.AddItem(iconView, 4, 0, false)
	row.AddItem(msgView, 0, 1, false)

	return row
}

func (d *ToastDemo) getLevelColor(level string) tcell.Color {
	switch level {
	case "Success":
		return theme.Success()
	case "Warning":
		return theme.Warning()
	case "Error":
		return theme.Error()
	default:
		return theme.Accent()
	}
}

const toastCode = `package main

import (
    "time"
    "github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

// Create toast manager (usually done once in app setup)
toastMgr := components.NewToastManager(app)

// Configure position and behavior
toastMgr.SetPosition(components.ToastTopRight)
toastMgr.SetMaxVisible(5)
toastMgr.SetDefaultDuration(3 * time.Second)

// Show simple toasts
toastMgr.Show("File saved", components.ToastSuccess)
toastMgr.Show("Connection lost", components.ToastError)
toastMgr.Show("Low disk space", components.ToastWarning)
toastMgr.Show("New message", components.ToastInfo)

// Show with custom duration
toastMgr.ShowWithDuration("Important!", components.ToastWarning, 10*time.Second)

// Show persistent toast (won't auto-dismiss)
toastMgr.ShowWithDuration("Action required", components.ToastError, 0)

// Show with actions
toastMgr.ShowWithActions("Update available", components.ToastInfo,
    []components.ToastAction{
        {Label: "Install", Handler: func() { /* install */ }},
        {Label: "Later", Handler: func() { /* dismiss */ }},
    },
)

// Dismiss programmatically
toast := toastMgr.Show("Message", components.ToastInfo)
toastMgr.Dismiss(toast.ID)
toastMgr.DismissAll()

// Callbacks
toastMgr.SetOnShow(func(t *components.Toast) {
    log.Printf("Showing: %s", t.Message)
})
`
