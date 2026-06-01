package theme

import (
	"github.com/gdamore/tcell/v2"
)

// ColorFunc is a function that returns a color dynamically from the current theme.
type ColorFunc func() tcell.Color

// Status represents a typed status with dynamic color and icon.
// Use DefineStatus to create instances, then call methods like .Color(), .ColorTag(), .Icon().
//
// Example:
//
//	var StatusRunning = theme.DefineStatus("Running", theme.Info, theme.IconRunning)
//	var StatusFailed = theme.DefineStatus("Failed", theme.Error, theme.IconFailed)
//
//	// Use the typed handle - compile-time safe!
//	color := StatusRunning.Color()       // Returns tcell.Color
//	tag := StatusRunning.ColorTag()      // Returns "#RRGGBB"
//	icon := StatusRunning.Icon()         // Returns icon string
type Status struct {
	name      string
	colorFunc ColorFunc
	icon      string
}

// DefineStatus creates a new typed status with a dynamic color function.
// The color function is called at lookup time, enabling live theme switching.
//
// Example:
//
//	var StatusRunning = theme.DefineStatus("Running", theme.Info, theme.IconRunning)
//	var StatusCompleted = theme.DefineStatus("Completed", theme.Success, theme.IconCompleted)
func DefineStatus(name string, colorFunc ColorFunc, icon string) *Status {
	return &Status{
		name:      name,
		colorFunc: colorFunc,
		icon:      icon,
	}
}

// DefineStatusStatic creates a status with a fixed color that doesn't change with theme.
// Use this for colors that should remain constant regardless of theme.
//
// Example:
//
//	var StatusCritical = theme.DefineStatusStatic("Critical", tcell.ColorRed, theme.IconError)
func DefineStatusStatic(name string, color tcell.Color, icon string) *Status {
	return &Status{
		name:      name,
		colorFunc: func() tcell.Color { return color },
		icon:      icon,
	}
}

// Name returns the display name of this status.
func (s *Status) Name() string {
	if s == nil {
		return ""
	}
	return s.name
}

// Color returns the tcell.Color for this status.
// The color is computed at call time, so it updates with theme changes.
func (s *Status) Color() tcell.Color {
	if s == nil || s.colorFunc == nil {
		return tcell.ColorGray
	}
	return s.colorFunc()
}

// ColorTag returns the hex color string for color tags (e.g., "#FF0000").
func (s *Status) ColorTag() string {
	return ColorToHex(s.Color())
}

// Icon returns the icon string for this status.
func (s *Status) Icon() string {
	if s == nil {
		return ""
	}
	return s.icon
}

// String returns a formatted string with icon and name (e.g., "󰄬 Running").
func (s *Status) String() string {
	if s == nil {
		return ""
	}
	if s.icon != "" {
		return s.icon + " " + s.name
	}
	return s.name
}

// Format returns a string with color tags.
// Example output: "[#00FF00]󰄬 Running[-]"
func (s *Status) Format() string {
	if s == nil {
		return ""
	}
	return "[" + s.ColorTag() + "]" + s.String() + "[-]"
}

// FormatName returns just the name with color tags.
// Example output: "[#00FF00]Running[-]"
func (s *Status) FormatName() string {
	if s == nil {
		return ""
	}
	return "[" + s.ColorTag() + "]" + s.name + "[-]"
}

// FormatIcon returns just the icon with color tags.
// Example output: "[#00FF00]󰄬[-]"
func (s *Status) FormatIcon() string {
	if s == nil {
		return ""
	}
	if s.icon == "" {
		return ""
	}
	return "[" + s.ColorTag() + "]" + s.icon + "[-]"
}

// Default status values for common use cases.
// Apps can use these directly or define their own domain-specific statuses.
var (
	StatusSuccess = DefineStatus("Success", Success, IconCheck)
	StatusError   = DefineStatus("Error", Error, IconError)
	StatusWarning = DefineStatus("Warning", Warning, IconWarning)
	StatusInfo    = DefineStatus("Info", Info, IconInfo)
	StatusPending = DefineStatus("Pending", FgDim, IconPending)
	StatusRunning = DefineStatus("Running", Info, IconRunning)
)
