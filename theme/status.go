package theme

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

// StatusStyle defines visual appearance for a named status.
type StatusStyle struct {
	Color tcell.Color
	Icon  string
}

var (
	statusRegistry = make(map[string]StatusStyle)
	statusMu       sync.RWMutex
)

// RegisterStatus adds a named status with styling.
// Apps should call this at init to register domain-specific statuses.
func RegisterStatus(name string, style StatusStyle) {
	statusMu.Lock()
	defer statusMu.Unlock()
	statusRegistry[name] = style
}

// StatusColor returns the color for a status name.
// Returns FgDim() if status not found.
func StatusColor(name string) tcell.Color {
	statusMu.RLock()
	defer statusMu.RUnlock()
	if style, ok := statusRegistry[name]; ok {
		return style.Color
	}
	return FgDim()
}

// StatusColorTag returns hex color for tview tags.
func StatusColorTag(name string) string {
	return ColorToHex(StatusColor(name))
}

// StatusIcon returns the icon for a status name.
// Returns empty string if status not found.
func StatusIcon(name string) string {
	statusMu.RLock()
	defer statusMu.RUnlock()
	if style, ok := statusRegistry[name]; ok {
		return style.Icon
	}
	return ""
}

// HasStatus checks if a status is registered.
func HasStatus(name string) bool {
	statusMu.RLock()
	defer statusMu.RUnlock()
	_, ok := statusRegistry[name]
	return ok
}

// ClearStatuses removes all registered statuses.
// Useful for testing or theme resets.
func ClearStatuses() {
	statusMu.Lock()
	defer statusMu.Unlock()
	statusRegistry = make(map[string]StatusStyle)
}

// RegisterDefaultStatuses registers common statuses.
// Apps can call this for sensible defaults, then override as needed.
func RegisterDefaultStatuses() {
	RegisterStatus("success", StatusStyle{Color: tcell.ColorGreen, Icon: IconCheck})
	RegisterStatus("error", StatusStyle{Color: tcell.ColorRed, Icon: IconError})
	RegisterStatus("warning", StatusStyle{Color: tcell.ColorYellow, Icon: IconWarning})
	RegisterStatus("info", StatusStyle{Color: tcell.ColorBlue, Icon: IconInfo})
	RegisterStatus("pending", StatusStyle{Color: tcell.ColorGray, Icon: IconPending})
	RegisterStatus("running", StatusStyle{Color: tcell.ColorBlue, Icon: IconRunning})
}
