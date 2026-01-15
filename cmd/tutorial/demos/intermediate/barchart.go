package intermediate

import (
	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
)

func init() {
	demos.Register(&BarChartDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "BarChart",
			DemoDescription: "Horizontal/vertical bar charts",
			DemoCategory:    demos.Intermediate,
			DemoCode:        barChartCode,
		},
	})
}

// BarChartDemo demonstrates the BarChart component.
type BarChartDemo struct {
	demos.DemoBase
	chart       *components.BarChart
	orientation string
	showValues  bool
}

// Component returns the demo component.
func (d *BarChartDemo) Component() tview.Primitive {
	d.orientation = "horizontal"
	d.showValues = true

	d.chart = components.NewBarChart().
		SetTitle("Language Popularity").
		SetItems(
			components.BarItem{Label: "Go", Value: 85, Color: theme.Info()},
			components.BarItem{Label: "Rust", Value: 72, Color: theme.Warning()},
			components.BarItem{Label: "Python", Value: 95, Color: theme.Success()},
			components.BarItem{Label: "TypeScript", Value: 78, Color: theme.Accent()},
			components.BarItem{Label: "C++", Value: 65, Color: theme.Error()},
		).
		SetOrientation(components.BarHorizontal).
		SetShowValues(true).
		SetShowLabels(true)

	d.Props = []demos.PropertyDescriptor{
		demos.SelectProp("orientation", "Bar orientation",
			[]string{"horizontal", "vertical"},
			func() string { return d.orientation },
			func(v string) {
				d.orientation = v
				if v == "horizontal" {
					d.chart.SetOrientation(components.BarHorizontal)
				} else {
					d.chart.SetOrientation(components.BarVertical)
				}
			},
			"horizontal",
		),
		demos.BoolProp("showValues", "Show values on bars",
			func() bool { return d.showValues },
			func(v bool) { d.showValues = v; d.chart.SetShowValues(v) },
			true,
		),
	}

	return d.chart
}

const barChartCode = `package main

import (
    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/theme"
)

// Create a bar chart
chart := components.NewBarChart().
    SetTitle("Sales by Region")

// Add items with labels and colors
chart.SetItems(
    components.BarItem{Label: "North", Value: 85, Color: theme.Info()},
    components.BarItem{Label: "South", Value: 72, Color: theme.Warning()},
    components.BarItem{Label: "East", Value: 95, Color: theme.Success()},
    components.BarItem{Label: "West", Value: 78, Color: theme.Accent()},
)

// Or set values with auto-labels
chart.SetValues(
    []float64{85, 72, 95, 78},
    []string{"North", "South", "East", "West"},
)

// Orientation
chart.SetOrientation(components.BarHorizontal)
chart.SetOrientation(components.BarVertical)

// Display options
chart.SetShowValues(true)   // Show values on bars
chart.SetShowLabels(true)   // Show labels
chart.SetBarGap(1)          // Gap between bars
chart.SetValueFormat("%.0f") // Format string
`
