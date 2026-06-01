package intermediate

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&SparklineDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Sparkline",
			DemoDescription: "Compact inline chart",
			DemoCategory:    demos.Intermediate,
			DemoCode:        sparklineCode,
		},
	})
}

// SparklineDemo demonstrates the Sparkline component.
type SparklineDemo struct {
	demos.DemoBase
	sparkline *components.Sparkline
}

// Component returns the demo component.
func (d *SparklineDemo) Component() core.Widget {
	d.sparkline = components.NewSparkline().
		SetLabel("CPU History").
		SetValues([]float64{20, 35, 45, 30, 55, 70, 65, 80, 75, 60, 45, 50, 65, 70, 85, 90, 75, 60}).
		SetMaxValue(100)

	// Wrap in panel
	panel := components.NewPanel()
	panel.SetTitle("Sparkline")
	panel.SetContent(d.sparkline)

	return panel
}

const sparklineCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create sparkline
sparkline := components.NewSparkline().
    SetLabel("Requests/s").
    SetValues([]float64{10, 25, 30, 45, 60, 55, 70, 65}).
    SetMaxValue(100)

// Add values incrementally (useful for streaming data)
sparkline.AddValue(75, 20)  // Add value, keep last 20 points

// Update all values
sparkline.SetValues(newDataPoints)

// Auto-scale (no max set)
sparkline := components.NewSparkline().
    SetValues(data)  // Will auto-scale to data range
`
