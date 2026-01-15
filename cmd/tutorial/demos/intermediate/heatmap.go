package intermediate

import (
	"math/rand"

	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
)

func init() {
	demos.Register(&HeatMapDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "HeatMap",
			DemoDescription: "Grid with color intensity",
			DemoCategory:    demos.Intermediate,
			DemoCode:        heatMapCode,
		},
	})
}

// HeatMapDemo demonstrates the HeatMap component.
type HeatMapDemo struct {
	demos.DemoBase
	heatmap    *components.HeatMap
	colorScale string
	showValues bool
}

// Component returns the demo component.
func (d *HeatMapDemo) Component() tview.Primitive {
	d.colorScale = "heat"
	d.showValues = false

	// Generate sample data (7 days x 24 hours)
	data := make([][]float64, 7)
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	hours := []string{"0", "4", "8", "12", "16", "20"}

	for i := range data {
		data[i] = make([]float64, 6)
		for j := range data[i] {
			// Simulate activity patterns
			if j >= 2 && j <= 4 { // Daytime
				data[i][j] = 50 + rand.Float64()*50
			} else {
				data[i][j] = rand.Float64() * 30
			}
		}
	}

	d.heatmap = components.NewHeatMap().
		SetTitle("Weekly Activity").
		SetValues(data).
		SetRowLabels(days).
		SetColLabels(hours).
		SetColorScale(components.ColorScaleGreen).
		SetCellSize(4, 1).
		SetShowValues(false)

	d.Props = []demos.PropertyDescriptor{
		demos.SelectProp("colorScale", "Color scale",
			[]string{"heat", "green", "red", "blue"},
			func() string { return d.colorScale },
			func(v string) {
				d.colorScale = v
				switch v {
				case "heat":
					d.heatmap.SetColorScale(components.ColorScaleHeat)
				case "green":
					d.heatmap.SetColorScale(components.ColorScaleGreen)
				case "red":
					d.heatmap.SetColorScale(components.ColorScaleRed)
				case "blue":
					d.heatmap.SetColorScale(components.ColorScaleBlue)
				}
			},
			"heat",
		),
		demos.BoolProp("showValues", "Show values in cells",
			func() bool { return d.showValues },
			func(v bool) { d.showValues = v; d.heatmap.SetShowValues(v) },
			false,
		),
	}

	return d.heatmap
}

const heatMapCode = `package main

import "github.com/atterpac/jig/components"

// Create a heat map
heatmap := components.NewHeatMap().
    SetTitle("Activity")

// Set 2D data grid
data := [][]float64{
    {10, 20, 30, 40},
    {15, 25, 35, 45},
    {20, 30, 40, 50},
}
heatmap.SetValues(data)

// Add labels
heatmap.SetRowLabels([]string{"Mon", "Tue", "Wed"})
heatmap.SetColLabels([]string{"Q1", "Q2", "Q3", "Q4"})

// Color scales
heatmap.SetColorScale(components.ColorScaleHeat)   // Blue->Red
heatmap.SetColorScale(components.ColorScaleGreen)  // Dark->Light green
heatmap.SetColorScale(components.ColorScaleRed)    // Dark->Light red
heatmap.SetColorScale(components.ColorScaleBlue)   // Dark->Light blue

// Custom color function
heatmap.SetColorFunc(func(normalized float64) tcell.Color {
    return tcell.NewRGBColor(int32(normalized*255), 0, 0)
})

// Display options
heatmap.SetCellSize(4, 1)    // Width, height
heatmap.SetShowValues(true)  // Show numbers in cells
heatmap.SetCellChar('#')     // Fill character
`
