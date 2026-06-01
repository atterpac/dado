package components

import (
	"github.com/gdamore/tcell/v2"
)

// HeatMapCell represents a single cell in the heat map
type HeatMapCell struct {
	Value float64
	Label string // Optional tooltip/label
}

// ColorScale defines how values map to colors
type ColorScale int

const (
	ColorScaleGreen  ColorScale = iota // Low=dark, High=bright green
	ColorScaleRed                      // Low=dark, High=bright red
	ColorScaleBlue                     // Low=dark, High=bright blue
	ColorScaleHeat                     // Blue -> Green -> Yellow -> Red
	ColorScaleTheme                    // Bg (low) → Accent (high) from active theme
	ColorScaleCustom                   // Use custom color function
)

// HeatMap renders a grid with color intensity based on values
type HeatMap struct {
	widgetBase

	// Data (2D grid, row-major)
	cells [][]HeatMapCell
	rows  int
	cols  int

	// Labels
	rowLabels []string
	colLabels []string

	// Range
	minValue  float64
	maxValue  float64
	autoScale bool

	// Display options
	title       string
	cellWidth   int // Width per cell (0 = auto)
	cellHeight  int // Height per cell (default 1)
	colorScale  ColorScale
	colorFunc   func(normalized float64) tcell.Color // Custom color function
	showValues  bool                                 // Show values in cells
	valueFormat string

	// Characters
	cellChar rune // Character to fill cells (default '█')

	// Callbacks
	onSelect func(row, col int, cell HeatMapCell)
	onHover  func(row, col int, cell HeatMapCell)
}

// NewHeatMap creates a new heat map component
func NewHeatMap() *HeatMap {
	h := &HeatMap{
		autoScale:   true,
		cellWidth:   3,
		cellHeight:  1,
		colorScale:  ColorScaleHeat,
		valueFormat: "%.0f",
		cellChar:    '█',
	}
	h.initWidget()
	return h
}

// --- Configuration (Fluent API) ---

// SetTitle sets the chart title
func (h *HeatMap) SetTitle(title string) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.title = title
	return h
}

// SetData sets the 2D grid data
func (h *HeatMap) SetData(cells [][]HeatMapCell) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cells = cells
	h.rows = len(cells)
	if h.rows > 0 {
		h.cols = len(cells[0])
	} else {
		h.cols = 0
	}
	if h.autoScale {
		h.recalculateRange()
	}
	return h
}

// SetValues sets grid from 2D float array
func (h *HeatMap) SetValues(values [][]float64) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rows = len(values)
	if h.rows > 0 {
		h.cols = len(values[0])
	} else {
		h.cols = 0
	}
	h.cells = make([][]HeatMapCell, h.rows)
	for i, row := range values {
		h.cells[i] = make([]HeatMapCell, len(row))
		for j, v := range row {
			h.cells[i][j] = HeatMapCell{Value: v}
		}
	}
	if h.autoScale {
		h.recalculateRange()
	}
	return h
}

// SetRowLabels sets labels for rows
func (h *HeatMap) SetRowLabels(labels []string) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rowLabels = labels
	return h
}

// SetColLabels sets labels for columns
func (h *HeatMap) SetColLabels(labels []string) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.colLabels = labels
	return h
}

// SetRange pins the color scale to a fixed [min, max] range and disables
// auto-scaling. Use this when comparing multiple heatmaps side-by-side or
// when the range has domain meaning (e.g., 0–100 for percentages).
func (h *HeatMap) SetRange(min, max float64) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.minValue = min
	h.maxValue = max
	h.autoScale = false
	return h
}

// SetAutoScale enables/disables automatic range
func (h *HeatMap) SetAutoScale(enabled bool) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.autoScale = enabled
	if enabled {
		h.recalculateRange()
	}
	return h
}

// SetCellSize sets cell dimensions
func (h *HeatMap) SetCellSize(width, height int) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cellWidth = width
	h.cellHeight = height
	return h
}

// SetColorScale sets the color mapping strategy. ColorScaleTheme adapts to
// the active theme automatically; ColorScaleCustom requires a SetColorFunc.
func (h *HeatMap) SetColorScale(scale ColorScale) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.colorScale = scale
	return h
}

// SetColorFunc sets a custom color function and switches the scale to
// ColorScaleCustom. The argument is the normalized value in [0.0, 1.0].
func (h *HeatMap) SetColorFunc(fn func(normalized float64) tcell.Color) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.colorFunc = fn
	h.colorScale = ColorScaleCustom
	return h
}

// SetShowValues shows/hides values in cells
func (h *HeatMap) SetShowValues(show bool) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.showValues = show
	return h
}

// SetValueFormat sets printf format for values
func (h *HeatMap) SetValueFormat(format string) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.valueFormat = format
	return h
}

// SetCellChar sets the character used to fill cells
func (h *HeatMap) SetCellChar(char rune) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cellChar = char
	return h
}

// SetOnSelect sets callback for cell selection
func (h *HeatMap) SetOnSelect(fn func(row, col int, cell HeatMapCell)) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onSelect = fn
	return h
}

// SetOnHover sets callback for cell hover
func (h *HeatMap) SetOnHover(fn func(row, col int, cell HeatMapCell)) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onHover = fn
	return h
}

// UpdateCell updates a single cell value
func (h *HeatMap) UpdateCell(row, col int, value float64) *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	if row >= 0 && row < h.rows && col >= 0 && col < h.cols {
		h.cells[row][col].Value = value
		if h.autoScale {
			h.recalculateRange()
		}
	}
	return h
}

// Clear removes all data
func (h *HeatMap) Clear() *HeatMap {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cells = nil
	h.rows = 0
	h.cols = 0
	h.minValue = 0
	h.maxValue = 0
	return h
}

// --- Internal ---

func (h *HeatMap) recalculateRange() {
	if len(h.cells) == 0 {
		h.minValue = 0
		h.maxValue = 1
		return
	}

	h.minValue = h.cells[0][0].Value
	h.maxValue = h.cells[0][0].Value

	for _, row := range h.cells {
		for _, cell := range row {
			if cell.Value < h.minValue {
				h.minValue = cell.Value
			}
			if cell.Value > h.maxValue {
				h.maxValue = cell.Value
			}
		}
	}

	if h.minValue == h.maxValue {
		h.maxValue = h.minValue + 1
	}
}

func (h *HeatMap) getColor(normalized float64) tcell.Color {
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	switch h.colorScale {
	case ColorScaleGreen:
		// Dark green to bright green
		g := int(normalized*200) + 55
		return tcell.NewRGBColor(0, int32(g), 0)

	case ColorScaleRed:
		// Dark red to bright red
		r := int(normalized*200) + 55
		return tcell.NewRGBColor(int32(r), 0, 0)

	case ColorScaleBlue:
		// Dark blue to bright blue
		b := int(normalized*200) + 55
		return tcell.NewRGBColor(0, 0, int32(b))

	case ColorScaleHeat:
		// Blue -> Cyan -> Green -> Yellow -> Red
		if normalized < 0.25 {
			// Blue to Cyan
			t := normalized * 4
			return tcell.NewRGBColor(0, int32(t*255), 255)
		} else if normalized < 0.5 {
			// Cyan to Green
			t := (normalized - 0.25) * 4
			return tcell.NewRGBColor(0, 255, int32((1-t)*255))
		} else if normalized < 0.75 {
			// Green to Yellow
			t := (normalized - 0.5) * 4
			return tcell.NewRGBColor(int32(t*255), 255, 0)
		} else {
			// Yellow to Red
			t := (normalized - 0.75) * 4
			return tcell.NewRGBColor(255, int32((1-t)*255), 0)
		}

	case ColorScaleTheme:
		// Interpolate from theme bg (low) to accent (high)
		th := h.th()
		lr, lg, lb := th.Bg().RGB()
		hr, hg, hb := th.Accent().RGB()
		r := int32(float64(lr) + normalized*float64(int32(hr)-int32(lr)))
		g := int32(float64(lg) + normalized*float64(int32(hg)-int32(lg)))
		b := int32(float64(lb) + normalized*float64(int32(hb)-int32(lb)))
		return tcell.NewRGBColor(r, g, b)

	case ColorScaleCustom:
		if h.colorFunc != nil {
			return h.colorFunc(normalized)
		}
		return h.th().Accent()

	default:
		return h.th().Accent()
	}
}

// Draw renders the heat map
func (h *HeatMap) Draw(screen tcell.Screen) {
	h.Box.DrawForSubclass(screen)
	x, y, width, height := h.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get colors at draw time
	th := h.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()

	bgStyle := tcell.StyleDefault.Background(bgColor)

	// Clear area
	fillRect(screen, x, y, width, height, bgStyle)

	if len(h.cells) == 0 {
		return
	}

	chartX := x
	chartY := y
	chartWidth := width
	chartHeight := height

	// Draw title
	if h.title != "" {
		titleStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := chartX + (chartWidth-len(h.title))/2
		for i, r := range h.title {
			if col+i < chartX+chartWidth {
				screen.SetContent(col+i, chartY, r, nil, titleStyle)
			}
		}
		chartY++
		chartHeight--
	}

	// Calculate row label width
	rowLabelWidth := 0
	if len(h.rowLabels) > 0 {
		for _, label := range h.rowLabels {
			if len(label) > rowLabelWidth {
				rowLabelWidth = len(label)
			}
		}
		rowLabelWidth++ // Space after label
	}

	// Reserve space for column labels
	colLabelHeight := 0
	if len(h.colLabels) > 0 {
		colLabelHeight = 1
		chartHeight--
	}

	// Calculate cell size
	cellWidth := h.cellWidth
	cellHeight := h.cellHeight
	if cellWidth <= 0 {
		cellWidth = (chartWidth - rowLabelWidth) / h.cols
		if cellWidth < 1 {
			cellWidth = 1
		}
	}
	if cellHeight <= 0 {
		cellHeight = 1
	}

	// Center the grid horizontally within the available chart width so a
	// fixed cell size doesn't hug the left edge of an oversized box.
	gridWidth := rowLabelWidth + h.cols*cellWidth
	leftPad := 0
	if gridWidth < chartWidth {
		leftPad = (chartWidth - gridWidth) / 2
	}

	gridX := chartX + leftPad + rowLabelWidth
	gridY := chartY + colLabelHeight

	// Draw column labels
	if len(h.colLabels) > 0 {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		for col := 0; col < h.cols && col < len(h.colLabels); col++ {
			label := h.colLabels[col]
			if len(label) > cellWidth {
				label = label[:cellWidth]
			}
			labelX := gridX + col*cellWidth + (cellWidth-len(label))/2
			for i, r := range label {
				if labelX+i < chartX+chartWidth {
					screen.SetContent(labelX+i, chartY, r, nil, labelStyle)
				}
			}
		}
	}

	// Draw row labels
	if len(h.rowLabels) > 0 {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		for row := 0; row < h.rows && row < len(h.rowLabels); row++ {
			label := h.rowLabels[row]
			if len(label) > rowLabelWidth-1 {
				label = label[:rowLabelWidth-1]
			}
			labelY := gridY + row*cellHeight + cellHeight/2
			if labelY < gridY+chartHeight {
				for i, r := range label {
					screen.SetContent(chartX+i, labelY, r, nil, labelStyle)
				}
			}
		}
	}

	// Draw cells
	for row := 0; row < h.rows; row++ {
		for col := 0; col < h.cols; col++ {
			cellX := gridX + col*cellWidth
			cellY := gridY + row*cellHeight

			if cellX >= chartX+chartWidth || cellY >= chartY+chartHeight {
				continue
			}

			cell := h.cells[row][col]

			// Calculate normalized value
			var normalized float64
			if h.maxValue != h.minValue {
				normalized = (cell.Value - h.minValue) / (h.maxValue - h.minValue)
			}

			color := h.getColor(normalized)
			cellStyle := tcell.StyleDefault.Background(bgColor).Foreground(color)

			// Draw cell
			for dy := 0; dy < cellHeight && cellY+dy < chartY+chartHeight; dy++ {
				for dx := 0; dx < cellWidth && cellX+dx < chartX+chartWidth; dx++ {
					screen.SetContent(cellX+dx, cellY+dy, h.cellChar, nil, cellStyle)
				}
			}

			// Draw value overlay
			if h.showValues && cellWidth >= 2 {
				valueStr := formatFloat(cell.Value, h.valueFormat)
				if len(valueStr) > cellWidth {
					valueStr = valueStr[:cellWidth]
				}
				// Use contrasting color for text
				textColor := bgColor
				valueStyle := tcell.StyleDefault.Background(color).Foreground(textColor)
				valueX := cellX + (cellWidth-len(valueStr))/2
				valueY := cellY + cellHeight/2
				for i, r := range valueStr {
					if valueX+i < cellX+cellWidth && valueX+i < chartX+chartWidth {
						screen.SetContent(valueX+i, valueY, r, nil, valueStyle)
					}
				}
			}
		}
	}
}

// GetFieldHeight returns preferred height
func (h *HeatMap) GetFieldHeight() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	height := h.rows * h.cellHeight
	if h.title != "" {
		height++
	}
	if len(h.colLabels) > 0 {
		height++
	}
	if height < 5 {
		height = 5
	}
	return height
}
