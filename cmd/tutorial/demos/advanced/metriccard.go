package advanced

import (
	"math"
	"math/rand"

	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
)

func init() {
	demos.Register(&MetricCardDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "MetricCard",
			DemoDescription: "Dashboard metric cards with sparklines",
			DemoCategory:    demos.Advanced,
			DemoCode:        metricCardCode,
		},
	})
}

// MetricCardDemo demonstrates the MetricCard component.
type MetricCardDemo struct {
	demos.DemoBase
	container *tview.Flex
	cards     []*components.MetricCard
	compact   bool
}

// Component returns the demo component.
func (d *MetricCardDemo) Component() tview.Primitive {
	d.compact = false

	// Generate sparkline data
	cpuData := make([]float64, 20)
	memData := make([]float64, 20)
	reqData := make([]float64, 20)

	for i := range cpuData {
		cpuData[i] = 30 + 40*math.Sin(float64(i)/4) + rand.Float64()*10
		memData[i] = 60 + rand.Float64()*15
		reqData[i] = 100 + rand.Float64()*50
	}

	// Create metric cards
	cpuCard := components.NewMetricCard().
		SetLabel("CPU Usage").
		SetValue("42%").
		SetTrend(components.TrendUp, "+5%", false).
		SetSparkline(cpuData).
		SetThresholds(70, 90, true)

	memCard := components.NewMetricCard().
		SetLabel("Memory").
		SetValue("6.2").
		SetUnit(" GB").
		SetTrend(components.TrendDown, "-0.3", true).
		SetSparkline(memData)

	reqCard := components.NewMetricCard().
		SetLabel("Requests/s").
		SetValue("1,234").
		SetTrend(components.TrendUp, "+12%", true).
		SetSparkline(reqData)

	errCard := components.NewMetricCard().
		SetLabel("Error Rate").
		SetValue("0.02%").
		SetTrend(components.TrendNeutral, "", false)

	d.cards = []*components.MetricCard{cpuCard, memCard, reqCard, errCard}

	// Layout
	d.container = tview.NewFlex().SetDirection(tview.FlexColumn)
	for _, card := range d.cards {
		d.container.AddItem(card, 0, 1, false)
	}

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("compact", "Compact single-line mode",
			func() bool { return d.compact },
			func(v bool) {
				d.compact = v
				for _, card := range d.cards {
					card.SetCompact(v)
				}
			},
			false,
		),
	}

	return d.container
}

const metricCardCode = `package main

import "github.com/atterpac/jig/components"

// Create a metric card
card := components.NewMetricCard().
    SetLabel("CPU Usage").
    SetValue("42%")

// With numeric value
card.SetNumericValue(42.5, "%.1f")
card.SetUnit("%")

// Trend indicator
card.SetTrend(
    components.TrendUp,    // TrendUp, TrendDown, TrendNeutral
    "+5%",                 // Trend value string
    false,                 // Is this trend good?
)

// Add sparkline
card.SetSparkline([]float64{10, 20, 15, 25, 30, 28, 35})

// Real-time update
card.AddSparkValue(newValue, 20)  // Rolling window of 20

// Thresholds (for value coloring)
card.SetThresholds(
    70,     // Warning threshold
    90,     // Error threshold
    true,   // Invert (higher is worse)
)

// Compact mode (single line)
card.SetCompact(true)

// Hide border
card.SetShowBorder(false)
`
