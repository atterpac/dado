package components

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// ProgressBar is a horizontal progress bar component.
type ProgressBar struct {
	*tview.Box

	progress       float64 // 0.0 to 1.0
	label          string
	showPercentage bool
	showValue      bool
	maxValue       float64
	currentValue   float64

	// Style
	filledChar rune
	emptyChar  rune
}

// NewProgressBar creates a new ProgressBar.
func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		Box:            tview.NewBox(),
		showPercentage: true,
		filledChar:     '█',
		emptyChar:      '░',
	}
}

// SetProgress sets the progress (0.0 to 1.0).
func (p *ProgressBar) SetProgress(progress float64) *ProgressBar {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	p.progress = progress
	return p
}

// GetProgress returns the current progress.
func (p *ProgressBar) GetProgress() float64 {
	return p.progress
}

// SetLabel sets the label displayed above the bar.
func (p *ProgressBar) SetLabel(label string) *ProgressBar {
	p.label = label
	return p
}

// SetShowPercentage enables/disables percentage display.
func (p *ProgressBar) SetShowPercentage(show bool) *ProgressBar {
	p.showPercentage = show
	return p
}

// SetShowValue enables value display (current/max).
func (p *ProgressBar) SetShowValue(show bool, current, max float64) *ProgressBar {
	p.showValue = show
	p.currentValue = current
	p.maxValue = max
	return p
}

// SetChars sets the filled and empty characters.
func (p *ProgressBar) SetChars(filled, empty rune) *ProgressBar {
	p.filledChar = filled
	p.emptyChar = empty
	return p
}

// Draw renders the progress bar.
func (p *ProgressBar) Draw(screen tcell.Screen) {
	p.Box.DrawForSubclass(screen, p)
	x, y, width, height := p.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	successColor := theme.Success()

	row := y

	// Draw label if present
	if p.label != "" && height > 1 {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := x
		for _, r := range p.label {
			if col < x+width {
				screen.SetContent(col, row, r, nil, labelStyle)
				col++
			}
		}
		row++
	}

	// Calculate bar width
	barWidth := width
	percentWidth := 0
	if p.showPercentage {
		percentWidth = 5 // " 100%"
		barWidth -= percentWidth
	}

	// Determine bar color based on progress
	barColor := accentColor
	if p.progress >= 1.0 {
		barColor = successColor
	}

	// Draw bar
	filledWidth := int(float64(barWidth) * p.progress)
	filledStyle := tcell.StyleDefault.Background(bgColor).Foreground(barColor)
	emptyStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	col := x
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			screen.SetContent(col, row, p.filledChar, nil, filledStyle)
		} else {
			screen.SetContent(col, row, p.emptyChar, nil, emptyStyle)
		}
		col++
	}

	// Draw percentage
	if p.showPercentage {
		percentStr := itoa(int(p.progress*100)) + "%"
		percentStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col++ // space
		for _, r := range percentStr {
			if col < x+width {
				screen.SetContent(col, row, r, nil, percentStyle)
				col++
			}
		}
	}
}

// GetFieldHeight returns the preferred height.
func (p *ProgressBar) GetFieldHeight() int {
	if p.label != "" {
		return 2
	}
	return 1
}

// SpinnerStyle defines the spinner animation style.
type SpinnerStyle int

const (
	SpinnerDots SpinnerStyle = iota
	SpinnerLine
	SpinnerBraille
	SpinnerCircle
	SpinnerArrow
)

var spinnerFrames = map[SpinnerStyle][]string{
	SpinnerDots:    {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	SpinnerLine:    {"-", "\\", "|", "/"},
	SpinnerBraille: {"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
	SpinnerCircle:  {"◐", "◓", "◑", "◒"},
	SpinnerArrow:   {"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
}

// Spinner is an animated loading indicator.
type Spinner struct {
	*tview.Box

	style    SpinnerStyle
	label    string
	frame    int
	running  bool
	interval time.Duration

	mu     sync.Mutex
	stopCh chan struct{}
}

// NewSpinner creates a new Spinner.
func NewSpinner() *Spinner {
	return &Spinner{
		Box:      tview.NewBox(),
		style:    SpinnerDots,
		interval: 100 * time.Millisecond,
	}
}

// SetStyle sets the spinner style.
func (s *Spinner) SetStyle(style SpinnerStyle) *Spinner {
	s.style = style
	return s
}

// SetLabel sets the label displayed next to the spinner.
func (s *Spinner) SetLabel(label string) *Spinner {
	s.label = label
	return s
}

// SetInterval sets the animation interval.
func (s *Spinner) SetInterval(interval time.Duration) *Spinner {
	s.interval = interval
	return s
}

// Start begins the animation.
func (s *Spinner) Start() *Spinner {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return s
	}

	s.running = true
	s.stopCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.mu.Lock()
				frames := spinnerFrames[s.style]
				s.frame = (s.frame + 1) % len(frames)
				s.mu.Unlock()

				// Request redraw
				theme.QueueUpdateDraw(func() {})
			}
		}
	}()

	return s
}

// Stop ends the animation.
func (s *Spinner) Stop() *Spinner {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return s
	}

	s.running = false
	close(s.stopCh)
	return s
}

// IsRunning returns whether the spinner is animating.
func (s *Spinner) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Draw renders the spinner.
func (s *Spinner) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen, s)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	accentColor := theme.Accent()

	s.mu.Lock()
	frames := spinnerFrames[s.style]
	currentFrame := frames[s.frame]
	s.mu.Unlock()

	col := x

	// Draw spinner
	spinnerStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)
	for _, r := range currentFrame {
		if col < x+width {
			screen.SetContent(col, y, r, nil, spinnerStyle)
			col++
		}
	}

	// Draw label
	if s.label != "" {
		col++ // space
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		for _, r := range s.label {
			if col < x+width {
				screen.SetContent(col, y, r, nil, labelStyle)
				col++
			}
		}
	}
}

// GetFieldHeight returns the preferred height.
func (s *Spinner) GetFieldHeight() int {
	return 1
}

// Gauge is a circular/arc style progress indicator.
type Gauge struct {
	*tview.Box

	value    float64 // 0.0 to 1.0
	label    string
	unit     string
	maxValue float64
}

// NewGauge creates a new Gauge.
func NewGauge() *Gauge {
	return &Gauge{
		Box:      tview.NewBox(),
		maxValue: 100,
	}
}

// SetValue sets the gauge value (0.0 to 1.0).
func (g *Gauge) SetValue(value float64) *Gauge {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	g.value = value
	return g
}

// SetLabel sets the label displayed below the gauge.
func (g *Gauge) SetLabel(label string) *Gauge {
	g.label = label
	return g
}

// SetUnit sets the unit displayed with the value.
func (g *Gauge) SetUnit(unit string) *Gauge {
	g.unit = unit
	return g
}

// SetMaxValue sets the max value for display purposes.
func (g *Gauge) SetMaxValue(max float64) *Gauge {
	g.maxValue = max
	return g
}

// Draw renders the gauge.
func (g *Gauge) Draw(screen tcell.Screen) {
	g.Box.DrawForSubclass(screen, g)
	x, y, width, height := g.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	successColor := theme.Success()
	warningColor := theme.Warning()
	errorColor := theme.Error()

	// Determine color based on value
	var valueColor tcell.Color
	if g.value >= 0.9 {
		valueColor = errorColor
	} else if g.value >= 0.7 {
		valueColor = warningColor
	} else if g.value >= 0.5 {
		valueColor = accentColor
	} else {
		valueColor = successColor
	}

	// ASCII gauge using block characters
	// ╭────────────╮
	// │  ██████░░  │
	// │    75%     │
	// │   CPU      │
	// ╰────────────╯

	// Draw border
	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
	filledStyle := tcell.StyleDefault.Background(bgColor).Foreground(valueColor)
	emptyStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
	textStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)

	row := y

	// Top border
	screen.SetContent(x, row, '╭', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╮', nil, borderStyle)
	row++

	// Clear inner area
	clearStyle := tcell.StyleDefault.Background(bgColor)
	for r := row; r < y+height-1; r++ {
		screen.SetContent(x, r, '│', nil, borderStyle)
		for col := x + 1; col < x+width-1; col++ {
			screen.SetContent(col, r, ' ', nil, clearStyle)
		}
		screen.SetContent(x+width-1, r, '│', nil, borderStyle)
	}

	// Draw progress bar in middle
	if height >= 3 {
		barRow := y + 1
		barWidth := width - 4
		barStart := x + 2
		filledWidth := int(float64(barWidth) * g.value)

		for i := 0; i < barWidth; i++ {
			var style tcell.Style
			if i < filledWidth {
				style = filledStyle
			} else {
				style = emptyStyle
			}
			screen.SetContent(barStart+i, barRow, '█', nil, style)
		}
	}

	// Draw percentage
	if height >= 4 {
		percentRow := y + 2
		percentStr := itoa(int(g.value*100)) + "%" + g.unit
		percentStart := x + (width-len(percentStr))/2
		for i, r := range percentStr {
			screen.SetContent(percentStart+i, percentRow, r, nil, textStyle)
		}
	}

	// Draw label
	if height >= 5 && g.label != "" {
		labelRow := y + 3
		labelStart := x + (width-len(g.label))/2
		for i, r := range g.label {
			if labelStart+i < x+width-1 {
				screen.SetContent(labelStart+i, labelRow, r, nil, textStyle)
			}
		}
	}

	// Bottom border
	bottomRow := y + height - 1
	screen.SetContent(x, bottomRow, '╰', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, bottomRow, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, bottomRow, '╯', nil, borderStyle)
}

// GetFieldHeight returns the preferred height.
func (g *Gauge) GetFieldHeight() int {
	return 5
}

// Sparkline is a minimal line chart for metrics.
type Sparkline struct {
	*tview.Box

	values   []float64
	maxValue float64
	label    string
}

// NewSparkline creates a new Sparkline.
func NewSparkline() *Sparkline {
	return &Sparkline{
		Box: tview.NewBox(),
	}
}

// SetValues sets the data points.
func (s *Sparkline) SetValues(values []float64) *Sparkline {
	s.values = values
	return s
}

// AddValue appends a value and maintains max length.
func (s *Sparkline) AddValue(value float64, maxLen int) *Sparkline {
	s.values = append(s.values, value)
	if len(s.values) > maxLen {
		s.values = s.values[len(s.values)-maxLen:]
	}
	return s
}

// SetMaxValue sets the maximum value for scaling.
func (s *Sparkline) SetMaxValue(max float64) *Sparkline {
	s.maxValue = max
	return s
}

// SetLabel sets the label.
func (s *Sparkline) SetLabel(label string) *Sparkline {
	s.label = label
	return s
}

// Draw renders the sparkline.
func (s *Sparkline) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen, s)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 || len(s.values) == 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	accentColor := theme.Accent()

	// Sparkline characters (8 levels)
	sparkChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	// Calculate max if not set
	maxVal := s.maxValue
	if maxVal == 0 {
		for _, v := range s.values {
			if v > maxVal {
				maxVal = v
			}
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	row := y

	// Draw label if present
	if s.label != "" && height > 1 {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := x
		for _, r := range s.label {
			if col < x+width {
				screen.SetContent(col, row, r, nil, labelStyle)
				col++
			}
		}
		row++
	}

	// Draw sparkline
	sparkStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)
	col := x
	startIdx := 0
	if len(s.values) > width {
		startIdx = len(s.values) - width
	}

	for i := startIdx; i < len(s.values) && col < x+width; i++ {
		// Map value to character index (0-7)
		normalized := s.values[i] / maxVal
		if normalized > 1 {
			normalized = 1
		}
		if normalized < 0 {
			normalized = 0
		}
		charIdx := int(normalized * 7)
		screen.SetContent(col, row, sparkChars[charIdx], nil, sparkStyle)
		col++
	}
}

// GetFieldHeight returns the preferred height.
func (s *Sparkline) GetFieldHeight() int {
	if s.label != "" {
		return 2
	}
	return 1
}
