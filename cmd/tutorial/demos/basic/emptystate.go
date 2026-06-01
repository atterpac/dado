package basic

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&EmptyStateDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "EmptyState",
			DemoDescription: "Centered placeholder display",
			DemoCategory:    demos.Basic,
			DemoCode:        emptyStateCode,
		},
	})
}

// EmptyStateDemo demonstrates the EmptyState component.
type EmptyStateDemo struct {
	demos.DemoBase
	empty   *components.EmptyState
	title   string
	message string
}

// Component returns the demo component.
func (d *EmptyStateDemo) Component() core.Widget {
	d.title = "No Items Found"
	d.message = "Try adjusting your search or filters"

	d.empty = components.NewEmptyState().
		SetIcon(theme.IconList).
		SetTitle(d.title).
		SetMessage(d.message)

	d.Props = []demos.PropertyDescriptor{
		demos.StringProp("title", "Main title text",
			func() string { return d.title },
			func(v string) { d.title = v; d.empty.SetTitle(v) },
			"No Items Found",
		),
		demos.StringProp("message", "Secondary message",
			func() string { return d.message },
			func(v string) { d.message = v; d.empty.SetMessage(v) },
			"Try adjusting your search or filters",
		),
	}

	return d.empty
}

const emptyStateCode = `package main

import (
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/theme"
)

// Create an empty state display
empty := components.NewEmptyState().
    SetIcon(theme.IconList).
    SetTitle("No Results").
    SetMessage("Try a different search term")

// Or configure all at once
empty.Configure(
    theme.IconError,
    "Something went wrong",
    "Please try again later",
)
`
