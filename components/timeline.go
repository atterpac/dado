package components

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

const (
	timelineLabelWidth = 25 // Width for lane labels on the left
	timelineMinWidth   = 40 // Minimum timeline bar area width
)

// TimelineLane represents a horizontal lane in the timeline.
type TimelineLane struct {
	ID        string     // Unique identifier
	Name      string     // Display name
	Status    string     // Status string (for color lookup via theme.StatusColor)
	StartTime time.Time  // When the lane starts
	EndTime   *time.Time // When the lane ends (nil = ongoing)
	Data      any        // Optional user data
}

// Timeline displays data as a horizontal Gantt-style timeline.
// Each lane represents a time-bounded item with status-based coloring.
type Timeline struct {
	*tview.Box
	lanes        []TimelineLane
	startTime    time.Time
	endTime      time.Time
	scrollX      int
	scrollY      int
	zoomLevel    float64
	selectedLane int

	// Configuration
	labelWidth int
	showLegend bool

	// Callbacks
	onSelect func(lane *TimelineLane)

	// Custom bar styling (optional)
	barStyleFn func(status string) (rune, tcell.Color)
}

// NewTimeline creates a new timeline/Gantt chart view.
func NewTimeline() *Timeline {
	t := &Timeline{
		Box:          tview.NewBox(),
		lanes:        []TimelineLane{},
		zoomLevel:    1.0,
		selectedLane: 0,
		labelWidth:   timelineLabelWidth,
		showLegend:   true,
	}

	t.SetBackgroundColor(theme.Bg())

	// Register for automatic theme updates
	theme.Register(t.Box)

	return t
}

// SetLanes sets the timeline lanes.
func (t *Timeline) SetLanes(lanes []TimelineLane) *Timeline {
	t.lanes = lanes
	t.selectedLane = 0
	t.calculateTimeRange()
	return t
}

// AddLane adds a single lane to the timeline.
func (t *Timeline) AddLane(lane TimelineLane) *Timeline {
	t.lanes = append(t.lanes, lane)
	t.calculateTimeRange()
	return t
}

// ClearLanes removes all lanes.
func (t *Timeline) ClearLanes() *Timeline {
	t.lanes = nil
	t.selectedLane = 0
	t.startTime = time.Time{}
	t.endTime = time.Time{}
	return t
}

// SetLabelWidth sets the width of the label column.
func (t *Timeline) SetLabelWidth(width int) *Timeline {
	t.labelWidth = width
	return t
}

// SetShowLegend enables/disables the legend display.
func (t *Timeline) SetShowLegend(show bool) *Timeline {
	t.showLegend = show
	return t
}

// SetOnSelect sets the callback for lane selection.
func (t *Timeline) SetOnSelect(fn func(lane *TimelineLane)) *Timeline {
	t.onSelect = fn
	return t
}

// SetBarStyleFn sets a custom function for determining bar character and color.
// If not set, uses theme.StatusColor for coloring.
func (t *Timeline) SetBarStyleFn(fn func(status string) (rune, tcell.Color)) *Timeline {
	t.barStyleFn = fn
	return t
}

// SetTimeRange manually sets the time range for the timeline.
// If not called, the range is calculated from lane data.
func (t *Timeline) SetTimeRange(start, end time.Time) *Timeline {
	t.startTime = start
	t.endTime = end
	return t
}

// GetSelectedLane returns the currently selected lane, or nil if none.
func (t *Timeline) GetSelectedLane() *TimelineLane {
	if t.selectedLane >= 0 && t.selectedLane < len(t.lanes) {
		return &t.lanes[t.selectedLane]
	}
	return nil
}

// GetLaneCount returns the number of lanes.
func (t *Timeline) GetLaneCount() int {
	return len(t.lanes)
}

// calculateTimeRange determines the time bounds from lane data.
func (t *Timeline) calculateTimeRange() {
	if len(t.lanes) == 0 {
		return
	}

	t.startTime = time.Now()
	t.endTime = time.Time{}

	for _, lane := range t.lanes {
		if lane.StartTime.Before(t.startTime) {
			t.startTime = lane.StartTime
		}
		if lane.EndTime != nil && lane.EndTime.After(t.endTime) {
			t.endTime = *lane.EndTime
		} else if lane.EndTime == nil && time.Now().After(t.endTime) {
			t.endTime = time.Now()
		}
	}

	// Ensure we have at least some time range
	if t.endTime.IsZero() || t.endTime.Before(t.startTime) {
		t.endTime = t.startTime.Add(time.Minute)
	}
}

// Draw renders the timeline view.
func (t *Timeline) Draw(screen tcell.Screen) {
	bgColor := theme.Bg()
	t.SetBackgroundColor(bgColor)

	t.Box.DrawForSubclass(screen, t)

	x, y, width, height := t.GetInnerRect()
	if width < t.labelWidth+10 || height < 3 {
		return
	}

	// Draw header with time scale
	t.drawHeader(screen, x, y, width)

	// Draw lanes starting from y+2 (after header)
	barAreaWidth := width - t.labelWidth - 1
	if barAreaWidth < timelineMinWidth {
		barAreaWidth = timelineMinWidth
	}

	timeRange := t.endTime.Sub(t.startTime)
	if timeRange <= 0 {
		timeRange = time.Minute
	}

	visibleLanes := height - 3 // Subtract header rows
	if t.showLegend {
		visibleLanes-- // Reserve space for legend
	}

	startLane := t.scrollY
	endLane := startLane + visibleLanes
	if endLane > len(t.lanes) {
		endLane = len(t.lanes)
	}

	for i := startLane; i < endLane; i++ {
		lane := t.lanes[i]
		laneY := y + 2 + (i - startLane)

		// Draw lane label
		t.drawLaneLabel(screen, x, laneY, lane, i == t.selectedLane)

		// Draw lane bar
		t.drawLaneBar(screen, x+t.labelWidth+1, laneY, barAreaWidth, lane, timeRange, i == t.selectedLane)
	}

	// Draw legend at bottom if enabled and space permits
	if t.showLegend && height > len(t.lanes)+4 {
		t.drawLegend(screen, x, y+height-1, width)
	}
}

// drawHeader draws the time scale header.
func (t *Timeline) drawHeader(screen tcell.Screen, x, y, width int) {
	barAreaWidth := width - t.labelWidth - 1
	timeRange := t.endTime.Sub(t.startTime)

	// Draw label column header
	labelStyle := tcell.StyleDefault.Foreground(theme.PanelTitle()).Background(theme.Bg())
	tview.Print(screen, "Event", x, y, t.labelWidth, tview.AlignLeft, theme.PanelTitle())

	// Draw time markers
	markerCount := 5
	if barAreaWidth < 60 {
		markerCount = 3
	}

	for i := 0; i <= markerCount; i++ {
		pos := x + t.labelWidth + 1 + (barAreaWidth * i / markerCount)
		if pos >= x+width {
			break
		}

		// Calculate time at this position
		offset := time.Duration(float64(timeRange) * float64(i) / float64(markerCount))
		tm := t.startTime.Add(offset)

		// Format time marker based on range
		var marker string
		if timeRange < time.Minute {
			marker = tm.Format("04:05.0")
		} else if timeRange < time.Hour {
			marker = tm.Format("04:05")
		} else if timeRange < 24*time.Hour {
			marker = tm.Format("15:04")
		} else {
			marker = tm.Format("01/02 15:04")
		}

		// Draw marker
		tview.Print(screen, marker, pos, y, 12, tview.AlignLeft, theme.FgDim())

		// Draw tick mark
		screen.SetContent(pos, y+1, '│', nil, labelStyle)
	}

	// Draw horizontal line under header
	lineStyle := tcell.StyleDefault.Foreground(theme.Border()).Background(theme.Bg())
	for i := x + t.labelWidth + 1; i < x+width; i++ {
		screen.SetContent(i, y+1, '─', nil, lineStyle)
	}
}

// drawLaneLabel draws the label for a lane.
func (t *Timeline) drawLaneLabel(screen tcell.Screen, x, y int, lane TimelineLane, selected bool) {
	// Truncate name if needed
	name := lane.Name
	maxLen := t.labelWidth - 2
	if len(name) > maxLen {
		name = name[:maxLen-1] + "…"
	}

	// Choose style based on selection
	var style tcell.Style
	if selected {
		style = tcell.StyleDefault.Foreground(theme.SelectionFg()).Background(theme.SelectionBg()).Bold(true)
	} else {
		color := t.statusColor(lane.Status)
		style = tcell.StyleDefault.Foreground(color).Background(theme.Bg())
	}

	// Clear label area
	for i := 0; i < t.labelWidth; i++ {
		screen.SetContent(x+i, y, ' ', nil, style)
	}

	// Draw name
	for i, r := range name {
		if x+i >= x+t.labelWidth {
			break
		}
		screen.SetContent(x+i, y, r, nil, style)
	}

	// Draw separator
	sepStyle := tcell.StyleDefault.Foreground(theme.Border()).Background(theme.Bg())
	screen.SetContent(x+t.labelWidth, y, '│', nil, sepStyle)
}

// drawLaneBar draws the timeline bar for a lane.
func (t *Timeline) drawLaneBar(screen tcell.Screen, x, y, width int, lane TimelineLane, timeRange time.Duration, selected bool) {
	// Calculate bar position and width
	startOffset := lane.StartTime.Sub(t.startTime)
	barStart := int(float64(width) * float64(startOffset) / float64(timeRange))

	var barEnd int
	if lane.EndTime != nil {
		endOffset := lane.EndTime.Sub(t.startTime)
		barEnd = int(float64(width) * float64(endOffset) / float64(timeRange))
	} else {
		// Running - extend to current time or end of view
		barEnd = width
	}

	// Ensure minimum bar width
	if barEnd <= barStart {
		barEnd = barStart + 1
	}

	// Apply zoom and scroll
	barStart = int(float64(barStart)*t.zoomLevel) - t.scrollX
	barEnd = int(float64(barEnd)*t.zoomLevel) - t.scrollX

	// Clamp to visible area
	if barStart < 0 {
		barStart = 0
	}
	if barEnd > width {
		barEnd = width
	}

	// Get bar style
	barChar, barColor := t.barStyle(lane.Status)
	barStyle := tcell.StyleDefault.Foreground(barColor).Background(theme.Bg())

	if selected {
		barStyle = barStyle.Bold(true)
	}

	// Draw empty space before bar
	emptyStyle := tcell.StyleDefault.Foreground(theme.BgLight()).Background(theme.Bg())
	for i := 0; i < barStart && i < width; i++ {
		screen.SetContent(x+i, y, '·', nil, emptyStyle)
	}

	// Draw the bar
	for i := barStart; i < barEnd && i < width; i++ {
		screen.SetContent(x+i, y, barChar, nil, barStyle)
	}

	// Draw empty space after bar
	for i := barEnd; i < width; i++ {
		screen.SetContent(x+i, y, '·', nil, emptyStyle)
	}
}

// drawLegend draws the status legend at the bottom.
func (t *Timeline) drawLegend(screen tcell.Screen, x, y, width int) {
	// Default legend items - can be customized via SetBarStyleFn
	legend := []struct {
		char   rune
		status string
		color  tcell.Color
	}{
		{'█', "Completed", theme.Success()},
		{'▓', "Running", theme.Info()},
		{'░', "Failed", theme.Error()},
		{'▒', "Pending", theme.FgDim()},
	}

	pos := x
	for _, item := range legend {
		if pos+15 > x+width {
			break
		}

		style := tcell.StyleDefault.Foreground(item.color).Background(theme.Bg())
		screen.SetContent(pos, y, item.char, nil, style)
		pos++

		labelStyle := tcell.StyleDefault.Foreground(theme.FgDim()).Background(theme.Bg())
		for _, r := range item.status {
			screen.SetContent(pos, y, r, nil, labelStyle)
			pos++
		}
		pos += 2 // spacing
	}
}

// barStyle returns the bar character and color for a status.
func (t *Timeline) barStyle(status string) (rune, tcell.Color) {
	// Use custom function if provided
	if t.barStyleFn != nil {
		return t.barStyleFn(status)
	}

	// Default styling based on common status names
	switch status {
	case "Running", "Active", "InProgress":
		return '▓', theme.Info()
	case "Completed", "Success", "Done", "Fired":
		return '█', theme.Success()
	case "Failed", "Error", "TimedOut":
		return '░', theme.Error()
	case "Canceled", "Terminated", "Aborted":
		return '▒', theme.Warning()
	case "Scheduled", "Initiated", "Pending", "Waiting":
		return '▒', theme.FgDim()
	default:
		// Try theme status registry
		if theme.HasStatus(status) {
			return '█', theme.StatusColor(status)
		}
		return '▒', theme.Fg()
	}
}

// statusColor returns the color for a status string.
func (t *Timeline) statusColor(status string) tcell.Color {
	// Use custom function if provided
	if t.barStyleFn != nil {
		_, color := t.barStyleFn(status)
		return color
	}

	// Default coloring
	switch status {
	case "Running", "Active", "InProgress":
		return theme.Info()
	case "Completed", "Success", "Done", "Fired":
		return theme.Success()
	case "Failed", "Error", "TimedOut":
		return theme.Error()
	case "Canceled", "Terminated", "Aborted":
		return theme.Warning()
	default:
		if theme.HasStatus(status) {
			return theme.StatusColor(status)
		}
		return theme.FgDim()
	}
}

// InputHandler handles keyboard input.
func (t *Timeline) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			t.moveSelection(-1)
		case tcell.KeyDown:
			t.moveSelection(1)
		case tcell.KeyLeft:
			t.scroll(-5)
		case tcell.KeyRight:
			t.scroll(5)
		case tcell.KeyHome:
			t.selectedLane = 0
			t.scrollY = 0
		case tcell.KeyEnd:
			if len(t.lanes) > 0 {
				t.selectedLane = len(t.lanes) - 1
			}
		case tcell.KeyEnter:
			if t.onSelect != nil && t.selectedLane >= 0 && t.selectedLane < len(t.lanes) {
				t.onSelect(&t.lanes[t.selectedLane])
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'k':
				t.moveSelection(-1)
			case 'j':
				t.moveSelection(1)
			case 'h':
				t.scroll(-5)
			case 'l':
				t.scroll(5)
			case '+', '=':
				t.zoom(1.2)
			case '-':
				t.zoom(0.8)
			case '0':
				t.resetView()
			case 'g':
				t.selectedLane = 0
				t.scrollY = 0
			case 'G':
				if len(t.lanes) > 0 {
					t.selectedLane = len(t.lanes) - 1
				}
			}
		case tcell.KeyCtrlD:
			_, _, _, height := t.GetInnerRect()
			t.moveSelection(height / 2)
		case tcell.KeyCtrlU:
			_, _, _, height := t.GetInnerRect()
			t.moveSelection(-height / 2)
		}
	})
}

// moveSelection moves the lane selection up or down.
func (t *Timeline) moveSelection(delta int) {
	if len(t.lanes) == 0 {
		return
	}

	t.selectedLane += delta
	if t.selectedLane < 0 {
		t.selectedLane = 0
	}
	if t.selectedLane >= len(t.lanes) {
		t.selectedLane = len(t.lanes) - 1
	}

	// Adjust scroll to keep selection visible
	_, _, _, height := t.GetInnerRect()
	visibleLanes := height - 3
	if t.showLegend {
		visibleLanes--
	}

	if t.selectedLane < t.scrollY {
		t.scrollY = t.selectedLane
	}
	if t.selectedLane >= t.scrollY+visibleLanes {
		t.scrollY = t.selectedLane - visibleLanes + 1
	}
}

// scroll horizontally scrolls the timeline.
func (t *Timeline) scroll(delta int) {
	t.scrollX += delta
	if t.scrollX < 0 {
		t.scrollX = 0
	}
}

// zoom adjusts the zoom level.
func (t *Timeline) zoom(factor float64) {
	t.zoomLevel *= factor
	if t.zoomLevel < 0.5 {
		t.zoomLevel = 0.5
	}
	if t.zoomLevel > 5.0 {
		t.zoomLevel = 5.0
	}
}

// resetView resets zoom and scroll to defaults.
func (t *Timeline) resetView() {
	t.zoomLevel = 1.0
	t.scrollX = 0
	t.scrollY = 0
}

// Zoom returns the current zoom level.
func (t *Timeline) Zoom() float64 {
	return t.zoomLevel
}

// SetZoom sets the zoom level.
func (t *Timeline) SetZoom(level float64) *Timeline {
	t.zoomLevel = level
	if t.zoomLevel < 0.5 {
		t.zoomLevel = 0.5
	}
	if t.zoomLevel > 5.0 {
		t.zoomLevel = 5.0
	}
	return t
}

// Focus implements tview.Primitive.
func (t *Timeline) Focus(delegate func(p tview.Primitive)) {
	t.Box.Focus(delegate)
}

// HasFocus implements tview.Primitive.
func (t *Timeline) HasFocus() bool {
	return t.Box.HasFocus()
}

// MouseHandler handles mouse input.
func (t *Timeline) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return t.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		x, y, _, _ := t.GetInnerRect()
		mx, my := event.Position()

		if !t.InRect(mx, my) {
			return false, nil
		}

		switch action {
		case tview.MouseLeftClick:
			setFocus(t)
			// Calculate which lane was clicked
			clickedLane := t.scrollY + (my - y - 2) // -2 for header rows
			if clickedLane >= 0 && clickedLane < len(t.lanes) {
				t.selectedLane = clickedLane
				return true, t
			}

		case tview.MouseLeftDoubleClick:
			clickedLane := t.scrollY + (my - y - 2)
			if clickedLane >= 0 && clickedLane < len(t.lanes) {
				t.selectedLane = clickedLane
				if t.onSelect != nil {
					t.onSelect(&t.lanes[t.selectedLane])
				}
				return true, t
			}

		case tview.MouseScrollUp:
			if mx > x+t.labelWidth {
				// Horizontal scroll in bar area
				t.scroll(-3)
			} else {
				// Vertical scroll in label area
				t.moveSelection(-1)
			}
			return true, t

		case tview.MouseScrollDown:
			if mx > x+t.labelWidth {
				t.scroll(3)
			} else {
				t.moveSelection(1)
			}
			return true, t
		}

		return false, nil
	})
}
