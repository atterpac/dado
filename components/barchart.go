package components

import (
	"github.com/gdamore/tcell/v2"
)

// BarOrientation controls whether bars are drawn horizontally (left-to-right)
// or vertically (bottom-to-top). Horizontal works better for long labels.
type BarOrientation int

const (
	BarHorizontal BarOrientation = iota
	BarVertical
)

// BarItem is a single bar entry. Color 0 falls back to the theme accent color.
type BarItem struct {
	Label string
	Value float64
	Color tcell.Color // 0 = use theme accent
}

// BarChart renders read-only horizontal or vertical bar charts. Update data
// with SetItems; the component redraws on the next draw cycle.
type BarChart struct {
	widgetBase

	// Data
	items []BarItem

	// Range
	minValue  float64
	maxValue  float64
	autoScale bool

	// Display options
	orientation BarOrientation
	title       string
	showValues  bool   // Show value next to/above bar
	showLabels  bool   // Show labels
	barWidth    int    // Width for vertical bars (0 = auto)
	barGap      int    // Gap between bars
	valueFormat string // Printf format for values
	filledChar  rune   // Character for filled portion
	emptyChar   rune   // Character for empty portion

	// Callbacks
	onSelect func(index int, item BarItem)
}

// NewBarChart creates a new bar chart component
func NewBarChart() *BarChart {
	c := &BarChart{
		autoScale:   true,
		orientation: BarHorizontal,
		showValues:  true,
		showLabels:  true,
		barGap:      1,
		valueFormat: "%.1f",
		filledChar:  '█',
		emptyChar:   '░',
	}
	c.initWidget()
	return c
}

// --- Configuration (Fluent API) ---

// SetTitle sets the chart title
func (c *BarChart) SetTitle(title string) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.title = title
	return c
}

// SetOrientation sets horizontal or vertical bars
func (c *BarChart) SetOrientation(o BarOrientation) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orientation = o
	return c
}

// SetItems sets all bar items
func (c *BarChart) SetItems(items ...BarItem) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = items
	if c.autoScale {
		c.recalculateRange()
	}
	return c
}

// AddItem appends a bar item
func (c *BarChart) AddItem(item BarItem) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = append(c.items, item)
	if c.autoScale {
		c.recalculateRange()
	}
	return c
}

// SetValues sets values with auto-generated labels
func (c *BarChart) SetValues(values []float64, labels []string) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make([]BarItem, len(values))
	for i, v := range values {
		label := ""
		if i < len(labels) {
			label = labels[i]
		}
		c.items[i] = BarItem{Label: label, Value: v}
	}
	if c.autoScale {
		c.recalculateRange()
	}
	return c
}

// SetRange sets fixed value range (disables auto-scale)
func (c *BarChart) SetRange(min, max float64) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.minValue = min
	c.maxValue = max
	c.autoScale = false
	return c
}

// SetAutoScale enables/disables automatic range
func (c *BarChart) SetAutoScale(enabled bool) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.autoScale = enabled
	if enabled {
		c.recalculateRange()
	}
	return c
}

// SetShowValues shows/hides values on bars
func (c *BarChart) SetShowValues(show bool) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.showValues = show
	return c
}

// SetShowLabels shows/hides bar labels
func (c *BarChart) SetShowLabels(show bool) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.showLabels = show
	return c
}

// SetBarWidth sets bar width (vertical orientation)
func (c *BarChart) SetBarWidth(width int) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.barWidth = width
	return c
}

// SetBarGap sets gap between bars
func (c *BarChart) SetBarGap(gap int) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.barGap = gap
	return c
}

// SetValueFormat sets printf format for values
func (c *BarChart) SetValueFormat(format string) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.valueFormat = format
	return c
}

// SetChars sets filled and empty bar characters
func (c *BarChart) SetChars(filled, empty rune) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.filledChar = filled
	c.emptyChar = empty
	return c
}

// SetOnSelect sets callback for bar selection
func (c *BarChart) SetOnSelect(fn func(index int, item BarItem)) *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onSelect = fn
	return c
}

// Clear removes all items
func (c *BarChart) Clear() *BarChart {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = nil
	c.minValue = 0
	c.maxValue = 0
	return c
}

// --- Internal ---

func (c *BarChart) recalculateRange() {
	if len(c.items) == 0 {
		c.minValue = 0
		c.maxValue = 1
		return
	}

	c.minValue = 0 // Bars typically start at 0
	c.maxValue = 0

	for _, item := range c.items {
		if item.Value > c.maxValue {
			c.maxValue = item.Value
		}
		if item.Value < c.minValue {
			c.minValue = item.Value
		}
	}

	if c.maxValue == c.minValue {
		c.maxValue = c.minValue + 1
	}
}

// Draw renders the bar chart
func (c *BarChart) Draw(screen tcell.Screen) {
	c.Box.DrawForSubclass(screen)
	x, y, width, height := c.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Get colors at draw time
	th := c.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	accentColor := th.Accent()

	bgStyle := tcell.StyleDefault.Background(bgColor)

	// Clear area
	fillRect(screen, x, y, width, height, bgStyle)

	if len(c.items) == 0 {
		return
	}

	chartX := x
	chartY := y
	chartWidth := width
	chartHeight := height

	// Draw title
	if c.title != "" {
		titleStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := chartX + (chartWidth-len(c.title))/2
		drawText(screen, col, chartY, chartX+chartWidth-col, c.title, titleStyle)
		chartY++
		chartHeight--
	}

	if c.orientation == BarHorizontal {
		c.drawHorizontal(screen, chartX, chartY, chartWidth, chartHeight, bgColor, fgColor, fgDimColor, accentColor)
	} else {
		c.drawVertical(screen, chartX, chartY, chartWidth, chartHeight, bgColor, fgColor, fgDimColor, accentColor)
	}
}

func (c *BarChart) drawHorizontal(screen tcell.Screen, x, y, width, height int, bgColor, fgColor, fgDimColor, accentColor tcell.Color) {
	// Calculate label width
	labelWidth := 0
	if c.showLabels {
		for _, item := range c.items {
			if len(item.Label) > labelWidth {
				labelWidth = len(item.Label)
			}
		}
		labelWidth += 1 // Space after label
	}

	// Calculate bar area
	barWidth := width - labelWidth
	if c.showValues {
		barWidth -= 8 // Space for value display
	}
	if barWidth < 10 {
		barWidth = 10
	}

	// Calculate bar height
	barCount := len(c.items)
	totalHeight := barCount + (barCount-1)*c.barGap
	if totalHeight > height {
		// Can't fit all bars, reduce gap
		c.barGap = 0
		totalHeight = barCount
	}

	startY := y + (height-totalHeight)/2
	if startY < y {
		startY = y
	}

	// Draw each bar
	for i, item := range c.items {
		if startY+i*(1+c.barGap) >= y+height {
			break
		}

		barY := startY + i*(1+c.barGap)
		barX := x + labelWidth

		// Get bar color
		barColor := item.Color
		if barColor == 0 {
			barColor = accentColor
		}

		// Draw label
		if c.showLabels && item.Label != "" {
			labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			drawText(screen, x, barY, labelWidth-1, item.Label, labelStyle)
		}

		// Calculate filled width
		var filledWidth int
		if c.maxValue != c.minValue {
			normalized := (item.Value - c.minValue) / (c.maxValue - c.minValue)
			filledWidth = int(normalized * float64(barWidth))
		}
		if filledWidth < 0 {
			filledWidth = 0
		}
		if filledWidth > barWidth {
			filledWidth = barWidth
		}

		// Draw bar
		filledStyle := tcell.StyleDefault.Background(bgColor).Foreground(barColor)
		emptyStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

		for j := 0; j < barWidth; j++ {
			if barX+j >= x+width {
				break
			}
			if j < filledWidth {
				screen.SetContent(barX+j, barY, c.filledChar, nil, filledStyle)
			} else {
				screen.SetContent(barX+j, barY, c.emptyChar, nil, emptyStyle)
			}
		}

		// Draw value
		if c.showValues {
			valueStr := formatFloat(item.Value, c.valueFormat)
			valueStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
			valueX := barX + barWidth + 1
			drawText(screen, valueX, barY, x+width-valueX, valueStr, valueStyle)
		}
	}
}

func (c *BarChart) drawVertical(screen tcell.Screen, x, y, width, height int, bgColor, fgColor, fgDimColor, accentColor tcell.Color) {
	// Reserve space for labels at bottom
	chartHeight := height
	if c.showLabels {
		chartHeight -= 2 // Label row + gap
	}
	if c.showValues {
		chartHeight -= 1 // Value row at top
	}

	if chartHeight < 3 {
		chartHeight = 3
	}

	// Calculate bar width and positions
	barCount := len(c.items)
	barWidth := c.barWidth
	if barWidth <= 0 {
		// Auto-calculate
		totalGaps := (barCount - 1) * c.barGap
		barWidth = (width - totalGaps) / barCount
		if barWidth < 1 {
			barWidth = 1
		}
	}

	totalWidth := barCount*barWidth + (barCount-1)*c.barGap
	startX := x + (width-totalWidth)/2
	if startX < x {
		startX = x
	}

	barTop := y
	if c.showValues {
		barTop++
	}

	// Draw each bar
	for i, item := range c.items {
		barX := startX + i*(barWidth+c.barGap)
		if barX >= x+width {
			break
		}

		// Get bar color
		barColor := item.Color
		if barColor == 0 {
			barColor = accentColor
		}

		// Calculate filled height
		var filledHeight int
		if c.maxValue != c.minValue {
			normalized := (item.Value - c.minValue) / (c.maxValue - c.minValue)
			filledHeight = int(normalized * float64(chartHeight))
		}
		if filledHeight < 0 {
			filledHeight = 0
		}
		if filledHeight > chartHeight {
			filledHeight = chartHeight
		}

		// Draw bar (from bottom up)
		filledStyle := tcell.StyleDefault.Background(bgColor).Foreground(barColor)
		emptyStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

		for row := 0; row < chartHeight; row++ {
			char := c.emptyChar
			style := emptyStyle
			if row >= chartHeight-filledHeight {
				char = c.filledChar
				style = filledStyle
			}

			for col := 0; col < barWidth && barX+col < x+width; col++ {
				screen.SetContent(barX+col, barTop+row, char, nil, style)
			}
		}

		// Draw value above bar
		if c.showValues {
			valueStr := formatFloat(item.Value, c.valueFormat)
			valueStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
			valueX := barX + (barWidth-len(valueStr))/2
			for j, r := range valueStr {
				if valueX+j >= x && valueX+j < x+width {
					screen.SetContent(valueX+j, y, r, nil, valueStyle)
				}
			}
		}

		// Draw label below bar
		if c.showLabels && item.Label != "" {
			labelY := barTop + chartHeight + 1
			labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			label := item.Label
			if len(label) > barWidth {
				label = label[:barWidth]
			}
			labelX := barX + (barWidth-len(label))/2
			for j, r := range label {
				if labelX+j >= x && labelX+j < x+width {
					screen.SetContent(labelX+j, labelY, r, nil, labelStyle)
				}
			}
		}
	}
}

// GetFieldHeight returns preferred height
func (c *BarChart) GetFieldHeight() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.orientation == BarHorizontal {
		return len(c.items) + 2
	}
	return 10
}
