package components

import (
	"math"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// Braille dot positions (2x4 grid per cell):
//   0 3
//   1 4
//   2 5
//   6 7
// Each dot corresponds to a bit: dot N = 1 << N
// Base braille character: U+2800

const brailleBase = 0x2800

// brailleDots maps (x, y) position within cell to bit value
// x: 0-1, y: 0-3
var brailleDots = [2][4]rune{
	{0x01, 0x02, 0x04, 0x40}, // left column (x=0): dots 1,2,3,7
	{0x08, 0x10, 0x20, 0x80}, // right column (x=1): dots 4,5,6,8
}

// DataSeries represents a single line in the graph
type DataSeries struct {
	Label  string
	Values []float64
	Color  tcell.Color // 0 = use theme accent
}

// LineGraphStyle configures the graph appearance
type LineGraphStyle int

const (
	LineGraphDots    LineGraphStyle = iota // Individual points only
	LineGraphSolid                         // Connected line
	LineGraphFilled                        // Fill area under line
)

// AxisConfig configures axis display
type AxisConfig struct {
	Show       bool
	LabelCount int    // Number of labels on Y axis (0 = auto)
	Format     string // Printf format for labels (default "%.1f")
}

// LineGraph renders line charts using braille characters
type LineGraph struct {
	*tview.Box

	mu sync.RWMutex

	// Data
	series []DataSeries

	// Range
	minValue  float64
	maxValue  float64
	autoScale bool // Auto-calculate min/max from data

	// Display options
	style     LineGraphStyle
	title     string
	showLegend bool
	yAxis     AxisConfig
	xAxis     AxisConfig

	// Grid
	showGrid   bool
	gridColor  tcell.Color // 0 = use theme FgDim

	// Callbacks
	onHover func(seriesIdx, pointIdx int, value float64)
}

// NewLineGraph creates a new line graph component
func NewLineGraph() *LineGraph {
	return &LineGraph{
		Box:       tview.NewBox(),
		autoScale: true,
		style:     LineGraphSolid,
		yAxis: AxisConfig{
			Show:       true,
			LabelCount: 5,
			Format:     "%.1f",
		},
	}
}

// --- Configuration (Fluent API) ---

// SetTitle sets the graph title
func (g *LineGraph) SetTitle(title string) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.title = title
	return g
}

// SetStyle sets the line rendering style
func (g *LineGraph) SetStyle(style LineGraphStyle) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.style = style
	return g
}

// SetSeries sets all data series
func (g *LineGraph) SetSeries(series ...DataSeries) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.series = series
	if g.autoScale {
		g.recalculateRange()
	}
	return g
}

// AddSeries appends a data series
func (g *LineGraph) AddSeries(s DataSeries) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.series = append(g.series, s)
	if g.autoScale {
		g.recalculateRange()
	}
	return g
}

// SetValues sets values for a single series (convenience for single-line graphs)
func (g *LineGraph) SetValues(values []float64) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.series) == 0 {
		g.series = []DataSeries{{Values: values}}
	} else {
		g.series[0].Values = values
	}
	if g.autoScale {
		g.recalculateRange()
	}
	return g
}

// AddValue appends a value to the first series with rolling window
func (g *LineGraph) AddValue(value float64, maxLen int) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.series) == 0 {
		g.series = []DataSeries{{}}
	}
	g.series[0].Values = append(g.series[0].Values, value)
	if len(g.series[0].Values) > maxLen {
		newValues := make([]float64, maxLen)
		copy(newValues, g.series[0].Values[len(g.series[0].Values)-maxLen:])
		g.series[0].Values = newValues
	}
	if g.autoScale {
		g.recalculateRange()
	}
	return g
}

// SetRange sets fixed Y-axis range (disables auto-scale)
func (g *LineGraph) SetRange(min, max float64) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.minValue = min
	g.maxValue = max
	g.autoScale = false
	return g
}

// SetAutoScale enables/disables automatic range calculation
func (g *LineGraph) SetAutoScale(enabled bool) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.autoScale = enabled
	if enabled {
		g.recalculateRange()
	}
	return g
}

// SetShowLegend enables/disables the legend
func (g *LineGraph) SetShowLegend(show bool) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.showLegend = show
	return g
}

// SetYAxis configures the Y axis
func (g *LineGraph) SetYAxis(config AxisConfig) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.yAxis = config
	return g
}

// SetXAxis configures the X axis
func (g *LineGraph) SetXAxis(config AxisConfig) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.xAxis = config
	return g
}

// SetShowGrid enables/disables the background grid
func (g *LineGraph) SetShowGrid(show bool) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.showGrid = show
	return g
}

// SetGridColor sets the grid line color
func (g *LineGraph) SetGridColor(color tcell.Color) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.gridColor = color
	return g
}

// SetOnHover sets the hover callback
func (g *LineGraph) SetOnHover(fn func(seriesIdx, pointIdx int, value float64)) *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onHover = fn
	return g
}

// --- Data Access ---

// GetRange returns the current Y-axis range
func (g *LineGraph) GetRange() (min, max float64) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.minValue, g.maxValue
}

// GetSeries returns all series
func (g *LineGraph) GetSeries() []DataSeries {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]DataSeries, len(g.series))
	copy(result, g.series)
	return result
}

// Clear removes all data
func (g *LineGraph) Clear() *LineGraph {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.series = nil
	g.minValue = 0
	g.maxValue = 0
	return g
}

// --- Internal ---

func (g *LineGraph) recalculateRange() {
	if len(g.series) == 0 {
		g.minValue = 0
		g.maxValue = 1
		return
	}

	g.minValue = math.MaxFloat64
	g.maxValue = -math.MaxFloat64

	for _, s := range g.series {
		for _, v := range s.Values {
			if v < g.minValue {
				g.minValue = v
			}
			if v > g.maxValue {
				g.maxValue = v
			}
		}
	}

	// Ensure we have a valid range
	if g.minValue == g.maxValue {
		g.minValue -= 1
		g.maxValue += 1
	}

	// Add 10% padding
	padding := (g.maxValue - g.minValue) * 0.1
	g.minValue -= padding
	g.maxValue += padding
}

// mapValueToY maps a data value to braille Y coordinate (0 = top)
func (g *LineGraph) mapValueToY(value float64, height int) int {
	if g.maxValue == g.minValue {
		return height / 2
	}
	// Braille has 4 dots per cell vertically
	brailleHeight := height * 4
	normalized := (value - g.minValue) / (g.maxValue - g.minValue)
	// Invert because Y=0 is top
	y := int(float64(brailleHeight-1) * (1 - normalized))
	if y < 0 {
		y = 0
	}
	if y >= brailleHeight {
		y = brailleHeight - 1
	}
	return y
}

// mapValueToX maps a data index to braille X coordinate
func (g *LineGraph) mapValueToX(index, dataLen, width int) int {
	if dataLen <= 1 {
		return 0
	}
	// Braille has 2 dots per cell horizontally
	brailleWidth := width * 2
	x := int(float64(index) / float64(dataLen-1) * float64(brailleWidth-1))
	return x
}

// Draw renders the line graph
func (g *LineGraph) Draw(screen tcell.Screen) {
	g.Box.DrawForSubclass(screen, g)
	x, y, width, height := g.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()

	bgStyle := tcell.StyleDefault.Background(bgColor)

	// Clear area
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Calculate layout
	chartX := x
	chartY := y
	chartWidth := width
	chartHeight := height

	// Reserve space for title
	if g.title != "" {
		titleStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := chartX + (chartWidth-len(g.title))/2
		for i, r := range g.title {
			if col+i < chartX+chartWidth {
				screen.SetContent(col+i, chartY, r, nil, titleStyle)
			}
		}
		chartY++
		chartHeight--
	}

	// Reserve space for Y axis
	yAxisWidth := 0
	if g.yAxis.Show {
		yAxisWidth = 6 // Space for Y-axis labels
		chartX += yAxisWidth
		chartWidth -= yAxisWidth
	}

	// Reserve space for legend
	legendHeight := 0
	if g.showLegend && len(g.series) > 0 {
		legendHeight = 1
		chartHeight -= legendHeight
	}

	if chartWidth <= 0 || chartHeight <= 0 {
		return
	}

	// Compute Y-axis tick values
	labelCount := g.yAxis.LabelCount
	if labelCount <= 0 {
		labelCount = 5
	}
	format := g.yAxis.Format
	if format == "" {
		format = "%.1f"
	}

	// tickValues holds the Y values where grid lines and labels are drawn.
	// For integer format, snap to nice integer boundaries to avoid duplicate labels.
	var tickValues []float64
	if format == "%.0f" {
		intMin := int(math.Floor(g.minValue))
		intMax := int(math.Ceil(g.maxValue))
		if intMax <= intMin {
			intMax = intMin + 1
		}
		intRange := intMax - intMin
		step := 1
		if intRange > labelCount {
			step = (intRange + labelCount - 1) / labelCount
		}
		for v := intMax; v >= intMin; v -= step {
			tickValues = append(tickValues, float64(v))
		}
	} else {
		for i := 0; i <= labelCount; i++ {
			value := g.maxValue - (float64(i)/float64(labelCount))*(g.maxValue-g.minValue)
			tickValues = append(tickValues, value)
		}
	}

	// Helper to map a tick value to a screen row
	tickRow := func(value float64) int {
		if g.maxValue == g.minValue {
			return chartY + chartHeight/2
		}
		frac := (g.maxValue - value) / (g.maxValue - g.minValue)
		return chartY + int(frac*float64(chartHeight-1))
	}

	// Draw grid
	if g.showGrid {
		gridColor := g.gridColor
		if gridColor == 0 {
			gridColor = fgDimColor
		}
		gridStyle := tcell.StyleDefault.Background(bgColor).Foreground(gridColor)

		for _, tv := range tickValues {
			row := tickRow(tv)
			if row >= chartY && row < chartY+chartHeight {
				for col := chartX; col < chartX+chartWidth; col++ {
					screen.SetContent(col, row, '·', nil, gridStyle)
				}
			}
		}
	}

	// Draw Y axis labels
	if g.yAxis.Show {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

		for _, tv := range tickValues {
			row := tickRow(tv)
			if row < chartY || row >= chartY+chartHeight {
				continue
			}
			label := formatFloat(tv, format)

			// Right-align label
			labelX := x + yAxisWidth - len(label) - 1
			for j, r := range label {
				if labelX+j >= x && labelX+j < chartX {
					screen.SetContent(labelX+j, row, r, nil, labelStyle)
				}
			}
		}
	}

	// Create braille canvas
	brailleWidth := chartWidth * 2
	brailleHeight := chartHeight * 4
	canvas := make([][]rune, chartHeight)
	for i := range canvas {
		canvas[i] = make([]rune, chartWidth)
	}

	// Draw each series
	for sIdx, series := range g.series {
		if len(series.Values) == 0 {
			continue
		}

		// Determine series color
		seriesColor := series.Color
		if seriesColor == 0 {
			// Cycle through theme colors
			colors := []tcell.Color{accentColor, theme.Success(), theme.Warning(), theme.Info()}
			seriesColor = colors[sIdx%len(colors)]
		}

		// Plot points
		prevBX, prevBY := -1, -1
		for i, value := range series.Values {
			bx := g.mapValueToX(i, len(series.Values), chartWidth)
			by := g.mapValueToY(value, chartHeight)

			// Clamp to canvas
			if bx >= brailleWidth {
				bx = brailleWidth - 1
			}
			if by >= brailleHeight {
				by = brailleHeight - 1
			}

			// Set point
			cellX := bx / 2
			cellY := by / 4
			dotX := bx % 2
			dotY := by % 4

			if cellX >= 0 && cellX < chartWidth && cellY >= 0 && cellY < chartHeight {
				canvas[cellY][cellX] |= brailleDots[dotX][dotY]
			}

			// Connect to previous point if solid line style
			if g.style == LineGraphSolid && prevBX >= 0 {
				g.drawLine(canvas, prevBX, prevBY, bx, by, chartWidth, chartHeight)
			}

			// Fill area under line if filled style
			if g.style == LineGraphFilled {
				g.fillToBottom(canvas, bx, by, chartWidth, chartHeight)
			}

			prevBX, prevBY = bx, by
		}

		// Render braille characters for this series
		lineStyle := tcell.StyleDefault.Background(bgColor).Foreground(seriesColor)
		for row := 0; row < chartHeight; row++ {
			for col := 0; col < chartWidth; col++ {
				if canvas[row][col] != 0 {
					char := rune(brailleBase) + canvas[row][col]
					screen.SetContent(chartX+col, chartY+row, char, nil, lineStyle)
				}
			}
		}

		// Clear canvas for next series (so colors don't overlap)
		if sIdx < len(g.series)-1 {
			for row := range canvas {
				for col := range canvas[row] {
					canvas[row][col] = 0
				}
			}
		}
	}

	// Draw legend
	if g.showLegend && len(g.series) > 0 {
		legendY := chartY + chartHeight
		legendX := chartX
		for sIdx, series := range g.series {
			if series.Label == "" {
				continue
			}

			seriesColor := series.Color
			if seriesColor == 0 {
				colors := []tcell.Color{accentColor, theme.Success(), theme.Warning(), theme.Info()}
				seriesColor = colors[sIdx%len(colors)]
			}

			// Draw color indicator
			indicatorStyle := tcell.StyleDefault.Background(bgColor).Foreground(seriesColor)
			screen.SetContent(legendX, legendY, '●', nil, indicatorStyle)
			legendX++

			// Draw label
			labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
			for _, r := range series.Label {
				if legendX < x+width {
					screen.SetContent(legendX, legendY, r, nil, labelStyle)
					legendX++
				}
			}
			legendX += 2 // Space between legend items
		}
	}
}

// drawLine draws a line between two braille coordinates using Bresenham's algorithm
func (g *LineGraph) drawLine(canvas [][]rune, x0, y0, x1, y1, width, height int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	brailleWidth := width * 2
	brailleHeight := height * 4

	for {
		// Plot current point
		if x0 >= 0 && x0 < brailleWidth && y0 >= 0 && y0 < brailleHeight {
			cellX := x0 / 2
			cellY := y0 / 4
			dotX := x0 % 2
			dotY := y0 % 4
			if cellX < width && cellY < height {
				canvas[cellY][cellX] |= brailleDots[dotX][dotY]
			}
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// fillToBottom fills from a point down to the bottom of the chart
func (g *LineGraph) fillToBottom(canvas [][]rune, bx, by, width, height int) {
	brailleHeight := height * 4
	for y := by; y < brailleHeight; y++ {
		cellX := bx / 2
		cellY := y / 4
		dotX := bx % 2
		dotY := y % 4
		if cellX >= 0 && cellX < width && cellY >= 0 && cellY < height {
			canvas[cellY][cellX] |= brailleDots[dotX][dotY]
		}
	}
}

// GetFieldHeight returns preferred height
func (g *LineGraph) GetFieldHeight() int {
	return 10
}

// Helper functions

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func formatFloat(v float64, format string) string {
	// Simple formatting without fmt to avoid import
	// For production, use fmt.Sprintf
	if format == "%.0f" {
		return itoa(int(v))
	}
	// Default: one decimal place
	intPart := int(v)
	fracPart := int((v - float64(intPart)) * 10)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	if v < 0 && intPart == 0 {
		return "-0." + itoa(fracPart)
	}
	return itoa(intPart) + "." + itoa(fracPart)
}
