package intermediate

import (
	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
)

func init() {
	demos.Register(&GaugeDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Gauge",
			DemoDescription: "Boxed progress with label",
			DemoCategory:    demos.Intermediate,
			DemoCode:        gaugeCode,
		},
	})
}

// GaugeDemo demonstrates the Gauge component.
type GaugeDemo struct {
	demos.DemoBase
	gauge *components.Gauge
	value float64
}

// Component returns the demo component.
func (d *GaugeDemo) Component() tview.Primitive {
	d.value = 0.75

	d.gauge = components.NewGauge().
		SetValue(d.value).
		SetLabel("CPU Usage").
		SetUnit("%").
		SetMaxValue(100)

	// Wrap in panel
	panel := components.NewPanel()
	panel.SetTitle("Gauge")
	panel.SetContent(d.gauge)

	d.ResetFunc = func() {
		d.value = 0.75
		d.gauge.SetValue(d.value)
	}

	return panel
}

const gaugeCode = `package main

import (
    "github.com/atterpac/jig/components"
)

// Create gauge - a boxed progress bar with centered label
// Renders as:
// +------------+
// |  ######..  |
// |    75%     |
// |    CPU     |
// +------------+

gauge := components.NewGauge().
    SetValue(0.75).      // 0.0 to 1.0
    SetLabel("CPU").
    SetUnit("%").
    SetMaxValue(100)

// Update value (color changes based on threshold)
// <50% = green, 50-70% = accent, 70-90% = warning, >90% = error
gauge.SetValue(0.85)
`
