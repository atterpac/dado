package components

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// Trend indicates value direction
type Trend int

const (
	TrendNeutral Trend = iota
	TrendUp
	TrendDown
)

// TrendIcon returns the icon for a trend
func (t Trend) Icon() string {
	switch t {
	case TrendUp:
		return "↑"
	case TrendDown:
		return "↓"
	default:
		return "→"
	}
}

// MetricCard displays a metric with optional sparkline and trend
type MetricCard struct {
	*tview.Box

	mu sync.RWMutex

	// Primary display
	label string
	value string
	unit  string

	// Trend
	trend       Trend
	trendValue  string // e.g., "+12%"
	trendGood   bool   // Is the trend direction good? (affects color)

	// Sparkline data
	sparkData   []float64
	sparkMax    float64
	showSpark   bool

	// Styling
	showBorder  bool
	compact     bool // Compact mode (single line)

	// Thresholds for value coloring
	warningThreshold float64
	errorThreshold   float64
	thresholdsSet    bool
	invertThresholds bool // True = higher is worse
}

// NewMetricCard creates a new metric card component
func NewMetricCard() *MetricCard {
	return &MetricCard{
		Box:        tview.NewBox(),
		showBorder: true,
		trend:      TrendNeutral,
	}
}

// --- Configuration (Fluent API) ---

// SetLabel sets the metric label
func (m *MetricCard) SetLabel(label string) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.label = label
	return m
}

// SetValue sets the displayed value
func (m *MetricCard) SetValue(value string) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value = value
	return m
}

// SetNumericValue sets value from a number with formatting
func (m *MetricCard) SetNumericValue(value float64, format string) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value = formatFloat(value, format)
	return m
}

// SetUnit sets the unit suffix
func (m *MetricCard) SetUnit(unit string) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.unit = unit
	return m
}

// SetTrend sets the trend indicator
func (m *MetricCard) SetTrend(trend Trend, value string, good bool) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trend = trend
	m.trendValue = value
	m.trendGood = good
	return m
}

// SetSparkline sets sparkline data
func (m *MetricCard) SetSparkline(data []float64) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sparkData = data
	m.showSpark = true
	// Auto-calculate max
	m.sparkMax = 0
	for _, v := range data {
		if v > m.sparkMax {
			m.sparkMax = v
		}
	}
	return m
}

// SetSparklineMax sets the max value for sparkline scaling
func (m *MetricCard) SetSparklineMax(max float64) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sparkMax = max
	return m
}

// AddSparkValue appends a value to sparkline with rolling window
func (m *MetricCard) AddSparkValue(value float64, maxLen int) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sparkData = append(m.sparkData, value)
	if len(m.sparkData) > maxLen {
		m.sparkData = m.sparkData[len(m.sparkData)-maxLen:]
	}
	m.showSpark = true
	// Update max
	if value > m.sparkMax {
		m.sparkMax = value
	}
	return m
}

// SetShowSpark enables/disables sparkline
func (m *MetricCard) SetShowSpark(show bool) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.showSpark = show
	return m
}

// SetShowBorder enables/disables the card border
func (m *MetricCard) SetShowBorder(show bool) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.showBorder = show
	return m
}

// SetCompact enables compact single-line mode
func (m *MetricCard) SetCompact(compact bool) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.compact = compact
	return m
}

// SetThresholds sets warning/error thresholds for value coloring
func (m *MetricCard) SetThresholds(warning, error float64, invert bool) *MetricCard {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warningThreshold = warning
	m.errorThreshold = error
	m.thresholdsSet = true
	m.invertThresholds = invert
	return m
}

// --- Data Access ---

// GetValue returns the current value string
func (m *MetricCard) GetValue() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}

// GetTrend returns the current trend
func (m *MetricCard) GetTrend() Trend {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.trend
}

// Draw renders the metric card
func (m *MetricCard) Draw(screen tcell.Screen) {
	m.Box.DrawForSubclass(screen, m)
	x, y, width, height := m.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	successColor := theme.Success()
	warningColor := theme.Warning()
	errorColor := theme.Error()

	bgStyle := tcell.StyleDefault.Background(bgColor)
	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	// Clear area
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	contentX := x
	contentY := y
	contentWidth := width
	contentHeight := height

	// Draw border
	if m.showBorder {
		// Corners
		screen.SetContent(x, y, '╭', nil, borderStyle)
		screen.SetContent(x+width-1, y, '╮', nil, borderStyle)
		screen.SetContent(x, y+height-1, '╰', nil, borderStyle)
		screen.SetContent(x+width-1, y+height-1, '╯', nil, borderStyle)

		// Top and bottom edges
		for col := x + 1; col < x+width-1; col++ {
			screen.SetContent(col, y, '─', nil, borderStyle)
			screen.SetContent(col, y+height-1, '─', nil, borderStyle)
		}

		// Left and right edges
		for row := y + 1; row < y+height-1; row++ {
			screen.SetContent(x, row, '│', nil, borderStyle)
			screen.SetContent(x+width-1, row, '│', nil, borderStyle)
		}

		contentX++
		contentY++
		contentWidth -= 2
		contentHeight -= 2
	}

	if contentWidth <= 0 || contentHeight <= 0 {
		return
	}

	if m.compact {
		m.drawCompact(screen, contentX, contentY, contentWidth, bgColor, fgColor, fgDimColor, accentColor, successColor, warningColor, errorColor)
	} else {
		m.drawFull(screen, contentX, contentY, contentWidth, contentHeight, bgColor, fgColor, fgDimColor, accentColor, successColor, warningColor, errorColor)
	}
}

func (m *MetricCard) drawCompact(screen tcell.Screen, x, y, width int, bgColor, fgColor, fgDimColor, accentColor, successColor, warningColor, errorColor tcell.Color) {
	col := x

	// Label
	if m.label != "" {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		for _, r := range m.label {
			if col < x+width {
				screen.SetContent(col, y, r, nil, labelStyle)
				col++
			}
		}
		screen.SetContent(col, y, ':', nil, labelStyle)
		col++
		screen.SetContent(col, y, ' ', nil, labelStyle)
		col++
	}

	// Value
	valueColor := m.getValueColor(fgColor, successColor, warningColor, errorColor)
	valueStyle := tcell.StyleDefault.Background(bgColor).Foreground(valueColor)
	for _, r := range m.value {
		if col < x+width {
			screen.SetContent(col, y, r, nil, valueStyle)
			col++
		}
	}

	// Unit
	if m.unit != "" {
		unitStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		for _, r := range m.unit {
			if col < x+width {
				screen.SetContent(col, y, r, nil, unitStyle)
				col++
			}
		}
	}

	// Trend (always render to keep layout stable)
	{
		col++ // space
		trendColor := fgDimColor
		if m.trend == TrendUp {
			if m.trendGood {
				trendColor = successColor
			} else {
				trendColor = errorColor
			}
		} else if m.trend == TrendDown {
			if m.trendGood {
				trendColor = errorColor
			} else {
				trendColor = successColor
			}
		}
		trendStyle := tcell.StyleDefault.Background(bgColor).Foreground(trendColor)

		icon := m.trend.Icon()
		for _, r := range icon {
			if col < x+width {
				screen.SetContent(col, y, r, nil, trendStyle)
				col++
			}
		}
		for _, r := range m.trendValue {
			if col < x+width {
				screen.SetContent(col, y, r, nil, trendStyle)
				col++
			}
		}
	}
}

func (m *MetricCard) drawFull(screen tcell.Screen, x, y, width, height int, bgColor, fgColor, fgDimColor, accentColor, successColor, warningColor, errorColor tcell.Color) {
	row := y

	// Label (top)
	if m.label != "" {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		labelX := x + (width-len(m.label))/2
		for i, r := range m.label {
			if labelX+i >= x && labelX+i < x+width {
				screen.SetContent(labelX+i, row, r, nil, labelStyle)
			}
		}
		row++
	}

	// Value (large, centered)
	if m.value != "" && row < y+height {
		valueColor := m.getValueColor(accentColor, successColor, warningColor, errorColor)
		valueStyle := tcell.StyleDefault.Background(bgColor).Foreground(valueColor)

		valueStr := m.value + m.unit
		valueX := x + (width-len(valueStr))/2
		for i, r := range valueStr {
			if valueX+i >= x && valueX+i < x+width {
				screen.SetContent(valueX+i, row, r, nil, valueStyle)
			}
		}
		row++
	}

	// Trend indicator (always render to keep layout stable)
	if row < y+height {
		trendColor := fgDimColor
		if m.trend == TrendUp {
			if m.trendGood {
				trendColor = successColor
			} else {
				trendColor = errorColor
			}
		} else if m.trend == TrendDown {
			if m.trendGood {
				trendColor = errorColor
			} else {
				trendColor = successColor
			}
		}
		trendStyle := tcell.StyleDefault.Background(bgColor).Foreground(trendColor)

		trendStr := m.trend.Icon() + " " + m.trendValue
		trendX := x + (width-len(trendStr))/2
		for i, r := range trendStr {
			if trendX+i >= x && trendX+i < x+width {
				screen.SetContent(trendX+i, row, r, nil, trendStyle)
			}
		}
		row++
	}

	// Sparkline (bottom)
	if m.showSpark && len(m.sparkData) > 0 && row < y+height {
		sparkChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
		sparkStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)

		// Calculate how many points we can show
		sparkWidth := width
		if sparkWidth > len(m.sparkData) {
			sparkWidth = len(m.sparkData)
		}

		startIdx := len(m.sparkData) - sparkWidth
		sparkX := x + (width-sparkWidth)/2

		maxVal := m.sparkMax
		if maxVal == 0 {
			maxVal = 1
		}

		for i := 0; i < sparkWidth; i++ {
			value := m.sparkData[startIdx+i]
			normalized := value / maxVal
			if normalized > 1 {
				normalized = 1
			}
			if normalized < 0 {
				normalized = 0
			}
			charIdx := int(normalized * 7)
			screen.SetContent(sparkX+i, row, sparkChars[charIdx], nil, sparkStyle)
		}
	}
}

func (m *MetricCard) getValueColor(defaultColor, successColor, warningColor, errorColor tcell.Color) tcell.Color {
	if !m.thresholdsSet {
		return defaultColor
	}

	// Try to parse value as float for threshold comparison
	var numValue float64
	// Simple parsing - assumes value is a number string
	for i, r := range m.value {
		if r >= '0' && r <= '9' {
			numValue = numValue*10 + float64(r-'0')
		} else if r == '.' {
			// Handle decimal
			decimal := 0.1
			for j := i + 1; j < len(m.value); j++ {
				if m.value[j] >= '0' && m.value[j] <= '9' {
					numValue += float64(m.value[j]-'0') * decimal
					decimal /= 10
				} else {
					break
				}
			}
			break
		} else if r == '-' && i == 0 {
			continue
		} else {
			break
		}
	}
	if len(m.value) > 0 && m.value[0] == '-' {
		numValue = -numValue
	}

	if m.invertThresholds {
		// Higher is worse
		if numValue >= m.errorThreshold {
			return errorColor
		}
		if numValue >= m.warningThreshold {
			return warningColor
		}
		return successColor
	} else {
		// Lower is worse
		if numValue <= m.errorThreshold {
			return errorColor
		}
		if numValue <= m.warningThreshold {
			return warningColor
		}
		return successColor
	}
}

// GetFieldHeight returns preferred height
func (m *MetricCard) GetFieldHeight() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.compact {
		return 1
	}
	height := 2 // Value is required
	if m.label != "" {
		height++
	}
	height++ // Trend row always present to keep layout stable
	if m.showSpark {
		height++
	}
	if m.showBorder {
		height += 2
	}
	return height
}
