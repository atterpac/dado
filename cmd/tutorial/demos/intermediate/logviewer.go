package intermediate

import (
	"time"


	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&LogViewerDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "LogViewer",
			DemoDescription: "Streaming logs with filtering",
			DemoCategory:    demos.Intermediate,
			DemoCode:        logViewerCode,
		},
	})
}

// LogViewerDemo demonstrates the LogViewer component.
type LogViewerDemo struct {
	demos.DemoBase
	viewer    *components.LogViewer
	follow    bool
	showLevel bool
}

// Component returns the demo component.
func (d *LogViewerDemo) Component() core.Widget {
	d.follow = true
	d.showLevel = true

	d.viewer = components.NewLogViewer().
		SetShowTimestamp(true).
		SetShowLevel(true).
		SetFollow(true).
		SetMaxEntries(1000)

	// Add sample log entries
	now := time.Now()
	entries := []struct {
		offset time.Duration
		level  components.LogLevel
		msg    string
	}{
		{-5 * time.Minute, components.LogLevelInfo, "Application starting..."},
		{-4 * time.Minute, components.LogLevelDebug, "Loading configuration from /etc/app/config.yaml"},
		{-4 * time.Minute, components.LogLevelInfo, "Configuration loaded successfully"},
		{-3 * time.Minute, components.LogLevelInfo, "Connecting to database..."},
		{-3 * time.Minute, components.LogLevelDebug, "Using connection pool size: 10"},
		{-2 * time.Minute, components.LogLevelInfo, "Database connection established"},
		{-2 * time.Minute, components.LogLevelInfo, "Starting HTTP server on :8080"},
		{-1 * time.Minute, components.LogLevelWarn, "Rate limit threshold approaching (80%)"},
		{-30 * time.Second, components.LogLevelError, "Failed to process request: timeout after 30s"},
		{-15 * time.Second, components.LogLevelInfo, "Request recovered, retrying..."},
		{-10 * time.Second, components.LogLevelDebug, "Cache hit ratio: 0.85"},
		{-5 * time.Second, components.LogLevelInfo, "Health check passed"},
		{0, components.LogLevelInfo, "System ready"},
	}

	for _, e := range entries {
		d.viewer.AddEntry(components.LogEntry{
			Timestamp: now.Add(e.offset),
			Level:     e.level,
			Message:   e.msg,
		})
	}

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("follow", "Auto-scroll to bottom",
			func() bool { return d.follow },
			func(v bool) { d.follow = v; d.viewer.SetFollow(v) },
			true,
		),
		demos.BoolProp("showLevel", "Show log level",
			func() bool { return d.showLevel },
			func(v bool) { d.showLevel = v; d.viewer.SetShowLevel(v) },
			true,
		),
	}

	return d.viewer
}

const logViewerCode = `package main


// Create a log viewer
viewer := components.NewLogViewer().
    SetMaxEntries(10000)

// Add log entries
viewer.AddEntry(components.LogEntry{
    Timestamp: time.Now(),
    Level:     components.LogLevelInfo,
    Message:   "Application started",
    Source:    "main",  // Optional
})

// Convenience methods
viewer.Debug("Debug message")
viewer.Info("Info message")
viewer.Warn("Warning message")
viewer.Error("Error message")

// Display options
viewer.SetShowTimestamp(true)
viewer.SetShowLevel(true)
viewer.SetShowSource(true)
viewer.SetTimestampFormat("15:04:05")

// Auto-scroll
viewer.SetFollow(true)

// Filtering
viewer.SetMinLevel(components.LogLevelInfo)  // Hide DEBUG
viewer.SetSearch("error")                     // Text search

// Clear logs
viewer.Clear()

// Keyboard:
// j/k - scroll
// g/G - top/bottom
// f   - toggle follow
// c   - clear
`
