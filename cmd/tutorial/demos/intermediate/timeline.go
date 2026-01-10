package intermediate

import (
	"time"

	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
)

func init() {
	demos.Register(&TimelineDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Timeline",
			DemoDescription: "Gantt-style timeline chart",
			DemoCategory:    demos.Intermediate,
			DemoCode:        timelineCode,
		},
	})
}

// TimelineDemo demonstrates the Timeline component.
type TimelineDemo struct {
	demos.DemoBase
	timeline   *components.Timeline
	showLegend bool
}

// Component returns the demo component.
func (d *TimelineDemo) Component() tview.Primitive {
	d.showLegend = true

	d.timeline = components.NewTimeline().
		SetShowLegend(d.showLegend).
		SetLabelWidth(20)

	// Create sample lanes
	now := time.Now()

	running := theme.StatusRunning
	success := theme.StatusSuccess
	failed := theme.StatusError
	pending := theme.StatusPending

	end1 := now.Add(-10 * time.Minute)
	end2 := now.Add(-5 * time.Minute)
	end3 := now.Add(-2 * time.Minute)

	lanes := []components.TimelineLane{
		{ID: "1", Name: "Build", Status: success, StartTime: now.Add(-30 * time.Minute), EndTime: &end1},
		{ID: "2", Name: "Test", Status: success, StartTime: now.Add(-20 * time.Minute), EndTime: &end2},
		{ID: "3", Name: "Lint", Status: failed, StartTime: now.Add(-15 * time.Minute), EndTime: &end3},
		{ID: "4", Name: "Deploy", Status: running, StartTime: now.Add(-5 * time.Minute), EndTime: nil},
		{ID: "5", Name: "Verify", Status: pending, StartTime: now, EndTime: nil},
	}

	d.timeline.SetLanes(lanes)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showLegend", "Show status legend",
			func() bool { return d.showLegend },
			func(v bool) { d.showLegend = v; d.timeline.SetShowLegend(v) },
			true,
		),
	}

	return d.timeline
}

const timelineCode = `package main

import (
    "time"
    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/theme"
)

// Create timeline
timeline := components.NewTimeline().
    SetShowLegend(true).
    SetLabelWidth(20)

// Add lanes with status-based colors
now := time.Now()
end := now.Add(-5 * time.Minute)

lanes := []components.TimelineLane{
    {
        ID:        "build",
        Name:      "Build",
        Status:    theme.StatusSuccess(),
        StartTime: now.Add(-30 * time.Minute),
        EndTime:   &end,
    },
    {
        ID:        "deploy",
        Name:      "Deploy",
        Status:    theme.StatusRunning(),
        StartTime: now.Add(-5 * time.Minute),
        EndTime:   nil, // Still running
    },
}

timeline.SetLanes(lanes)

// Or add lanes one by one
timeline.AddLane(components.TimelineLane{
    ID:     "test",
    Name:   "Test",
    Status: theme.StatusPending(),
})

// Set custom time range
timeline.SetTimeRange(startTime, endTime)

// Selection callback
timeline.SetOnSelect(func(lane *components.TimelineLane) {
    fmt.Printf("Selected: %s\n", lane.Name)
})
`
