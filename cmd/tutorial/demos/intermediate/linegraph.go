package intermediate

import (
	"math"
	"math/rand"


	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&LineGraphDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "LineGraph",
			DemoDescription: "Braille-based line charts",
			DemoCategory:    demos.Intermediate,
			DemoCode:        lineGraphCode,
		},
	})
}

// LineGraphDemo demonstrates the LineGraph component.
type LineGraphDemo struct {
	demos.DemoBase
	graph    *components.LineGraph
	style    string
	showGrid bool
}

// Component returns the demo component.
func (d *LineGraphDemo) Component() core.Widget {
	d.style = "solid"
	d.showGrid = true

	// Generate sample data
	cpuData := make([]float64, 30)
	memData := make([]float64, 30)
	for i := range cpuData {
		cpuData[i] = 30 + 40*math.Sin(float64(i)/5) + rand.Float64()*10
		memData[i] = 50 + 20*math.Cos(float64(i)/4) + rand.Float64()*5
	}

	d.graph = components.NewLineGraph().
		SetTitle("System Metrics").
		SetSeries(
			components.DataSeries{Label: "CPU", Values: cpuData, Color: theme.Success()},
			components.DataSeries{Label: "Memory", Values: memData, Color: theme.Warning()},
		).
		SetStyle(components.LineGraphSolid).
		SetShowGrid(true).
		SetShowLegend(true).
		SetYAxis(components.AxisConfig{Show: true, LabelCount: 5})

	d.Props = []demos.PropertyDescriptor{
		demos.SelectProp("style", "Line rendering style",
			[]string{"solid", "dots", "filled"},
			func() string { return d.style },
			func(v string) {
				d.style = v
				switch v {
				case "solid":
					d.graph.SetStyle(components.LineGraphSolid)
				case "dots":
					d.graph.SetStyle(components.LineGraphDots)
				case "filled":
					d.graph.SetStyle(components.LineGraphFilled)
				}
			},
			"solid",
		),
		demos.BoolProp("showGrid", "Show background grid",
			func() bool { return d.showGrid },
			func(v bool) { d.showGrid = v; d.graph.SetShowGrid(v) },
			true,
		),
	}

	return d.graph
}

const lineGraphCode = `package main

import (
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/theme"
)

// Create a line graph
graph := components.NewLineGraph().
    SetTitle("CPU Usage").
    SetValues(cpuData).  // []float64
    SetStyle(components.LineGraphSolid)

// Multiple series
graph.SetSeries(
    components.DataSeries{
        Label: "CPU",
        Values: cpuData,
        Color: theme.Success(),
    },
    components.DataSeries{
        Label: "Memory",
        Values: memData,
        Color: theme.Warning(),
    },
)

// Options
graph.SetShowGrid(true)
graph.SetShowLegend(true)
graph.SetYAxis(components.AxisConfig{
    Show: true,
    LabelCount: 5,
})

// Real-time streaming
graph.AddValue(newValue, 100)  // Rolling window

// Line styles:
// - LineGraphSolid  (connected line)
// - LineGraphDots   (points only)
// - LineGraphFilled (area under curve)
`
