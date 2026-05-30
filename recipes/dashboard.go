package recipes

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/nav"
	"github.com/atterpac/dado/theme"
)

// DashboardSection represents a section in the dashboard.
type DashboardSection struct {
	Title     string
	Span      int // number of columns to span (1-4)
	Content   tview.Primitive
	Refresh   time.Duration // refresh interval, 0 for no auto-refresh
	OnRefresh func()        // called on refresh
}

// Dashboard is a multi-pane status dashboard.
type Dashboard struct {
	*tview.Box

	title    string
	sections []DashboardSection
	columns  int

	// Refresh management
	stopRefresh map[int]chan struct{}
	tickers     map[int]*time.Ticker

	// Navigation
	focusedSection int
	focusable      []int // indices of focusable sections

	mu sync.RWMutex
}

// NewDashboard creates a new Dashboard.
func NewDashboard() *Dashboard {
	return &Dashboard{
		Box:         tview.NewBox(),
		columns:     4,
		stopRefresh: make(map[int]chan struct{}),
		tickers:     make(map[int]*time.Ticker),
	}
}

// SetTitle sets the dashboard title.
func (d *Dashboard) SetTitle(title string) *Dashboard {
	d.title = title
	return d
}

// SetColumns sets the number of columns in the grid.
func (d *Dashboard) SetColumns(columns int) *Dashboard {
	if columns < 1 {
		columns = 1
	}
	if columns > 6 {
		columns = 6
	}
	d.columns = columns
	return d
}

// AddSection adds a section to the dashboard.
func (d *Dashboard) AddSection(section DashboardSection) *Dashboard {
	if section.Span < 1 {
		section.Span = 1
	}
	if section.Span > d.columns {
		section.Span = d.columns
	}
	d.sections = append(d.sections, section)
	return d
}

// AddWidget adds a simple widget section.
func (d *Dashboard) AddWidget(title string, content tview.Primitive) *Dashboard {
	return d.AddSection(DashboardSection{
		Title:   title,
		Span:    1,
		Content: content,
	})
}

// AddSparkline adds a sparkline section.
func (d *Dashboard) AddSparkline(title string, label string) *components.Sparkline {
	sparkline := components.NewSparkline().SetLabel(label)
	d.AddSection(DashboardSection{
		Title:   title,
		Span:    1,
		Content: sparkline,
	})
	return sparkline
}

// AddGauge adds a gauge section.
func (d *Dashboard) AddGauge(title string, label string) *components.Gauge {
	gauge := components.NewGauge().SetLabel(label)
	d.AddSection(DashboardSection{
		Title:   title,
		Span:    1,
		Content: gauge,
	})
	return gauge
}

// AddProgressBar adds a progress bar section.
func (d *Dashboard) AddProgressBar(title string) *components.ProgressBar {
	bar := components.NewProgressBar().SetLabel(title)
	d.AddSection(DashboardSection{
		Title:   title,
		Span:    1,
		Content: bar,
	})
	return bar
}

// GetSection returns a section by index.
func (d *Dashboard) GetSection(index int) *DashboardSection {
	if index >= 0 && index < len(d.sections) {
		return &d.sections[index]
	}
	return nil
}

// UpdateSection updates a section's content.
func (d *Dashboard) UpdateSection(index int, content tview.Primitive) *Dashboard {
	if index >= 0 && index < len(d.sections) {
		d.sections[index].Content = content
	}
	return d
}

// Start begins the dashboard lifecycle (auto-refresh).
func (d *Dashboard) Start() {
	for i, section := range d.sections {
		if section.Refresh > 0 {
			stopCh := make(chan struct{})
			d.stopRefresh[i] = stopCh
			ticker := time.NewTicker(section.Refresh)
			d.tickers[i] = ticker

			go func(idx int, t *time.Ticker, stop chan struct{}, refresh func()) {
				for {
					select {
					case <-stop:
						return
					case <-t.C:
						if refresh != nil {
							refresh()
							theme.QueueUpdateDraw(func() {})
						}
					}
				}
			}(i, ticker, stopCh, section.OnRefresh)
		}
	}
}

// Stop ends the dashboard lifecycle.
func (d *Dashboard) Stop() {
	for _, ticker := range d.tickers {
		ticker.Stop()
	}
	for _, stopCh := range d.stopRefresh {
		close(stopCh)
	}
	d.tickers = make(map[int]*time.Ticker)
	d.stopRefresh = make(map[int]chan struct{})
}

// Name returns the display name for breadcrumbs.
func (d *Dashboard) Name() string {
	if d.title != "" {
		return d.title
	}
	return "Dashboard"
}

// Hints returns the current key hints.
func (d *Dashboard) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "Tab", Description: "Next section"},
		{Key: "Shift+Tab", Description: "Prev section"},
		{Key: "r", Description: "Refresh all"},
	}
}

// Draw renders the dashboard.
func (d *Dashboard) Draw(screen tcell.Screen) {
	d.Box.DrawForSubclass(screen, d)
	x, y, width, height := d.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	borderColor := theme.Border()
	borderFocusColor := theme.BorderFocus()

	titleStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)
	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(borderColor)
	focusBorderStyle := tcell.StyleDefault.Background(bgColor).Foreground(borderFocusColor)

	row := y

	// Draw title if set
	if d.title != "" {
		col := x
		for _, r := range d.title {
			if col < x+width {
				screen.SetContent(col, row, r, nil, titleStyle)
				col++
			}
		}
		row += 2
	}

	// Calculate section layout
	type sectionLayout struct {
		x, y, w, h int
		section    *DashboardSection
		index      int
	}

	var layouts []sectionLayout
	colWidth := width / d.columns
	sectionHeight := (height - row + y) / ((len(d.sections) + d.columns - 1) / d.columns)
	if sectionHeight < 5 {
		sectionHeight = 5
	}

	currentCol := 0
	currentRow := row
	for i := range d.sections {
		section := &d.sections[i]

		// Check if this section fits in current row
		if currentCol+section.Span > d.columns {
			currentCol = 0
			currentRow += sectionHeight
		}

		// Check if we've run out of vertical space
		if currentRow >= y+height {
			break
		}

		sectionWidth := colWidth * section.Span
		actualHeight := sectionHeight
		if currentRow+actualHeight > y+height {
			actualHeight = y + height - currentRow
		}

		layouts = append(layouts, sectionLayout{
			x:       x + currentCol*colWidth,
			y:       currentRow,
			w:       sectionWidth,
			h:       actualHeight,
			section: section,
			index:   i,
		})

		currentCol += section.Span
	}

	// Update focusable indices
	d.focusable = nil
	for i, layout := range layouts {
		if layout.section.Content != nil {
			d.focusable = append(d.focusable, i)
		}
	}

	// Draw sections
	for layoutIdx, layout := range layouts {
		section := layout.section
		sx, sy, sw, sh := layout.x, layout.y, layout.w, layout.h

		// Determine if this section is focused
		isFocused := layoutIdx == d.focusedSection
		currentBorderStyle := borderStyle
		if isFocused {
			currentBorderStyle = focusBorderStyle
		}

		// Draw section border
		// Top border with title
		screen.SetContent(sx, sy, '╭', nil, currentBorderStyle)
		col := sx + 1

		// Draw title in border
		if section.Title != "" {
			titleText := " " + section.Title + " "
			for _, r := range titleText {
				if col < sx+sw-1 {
					screen.SetContent(col, sy, r, nil, titleStyle)
					col++
				}
			}
		}

		// Complete top border
		for col < sx+sw-1 {
			screen.SetContent(col, sy, '─', nil, currentBorderStyle)
			col++
		}
		screen.SetContent(sx+sw-1, sy, '╮', nil, currentBorderStyle)

		// Side borders and clear interior
		clearStyle := tcell.StyleDefault.Background(bgColor)
		for row := sy + 1; row < sy+sh-1; row++ {
			screen.SetContent(sx, row, '│', nil, currentBorderStyle)
			for col := sx + 1; col < sx+sw-1; col++ {
				screen.SetContent(col, row, ' ', nil, clearStyle)
			}
			screen.SetContent(sx+sw-1, row, '│', nil, currentBorderStyle)
		}

		// Bottom border
		screen.SetContent(sx, sy+sh-1, '╰', nil, currentBorderStyle)
		for col := sx + 1; col < sx+sw-1; col++ {
			screen.SetContent(col, sy+sh-1, '─', nil, currentBorderStyle)
		}
		screen.SetContent(sx+sw-1, sy+sh-1, '╯', nil, currentBorderStyle)

		// Draw content
		if section.Content != nil {
			contentX := sx + 1
			contentY := sy + 1
			contentW := sw - 2
			contentH := sh - 2

			if contentW > 0 && contentH > 0 {
				section.Content.SetRect(contentX, contentY, contentW, contentH)
				section.Content.Draw(screen)
			}
		}
	}

	_ = fgColor
	_ = fgDimColor
}

// InputHandler handles keyboard input.
func (d *Dashboard) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyTab:
			d.focusNext()
		case tcell.KeyBacktab:
			d.focusPrev()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'r':
				// Refresh all sections
				for _, section := range d.sections {
					if section.OnRefresh != nil {
						section.OnRefresh()
					}
				}
			case 'j', 'l':
				d.focusNext()
			case 'k', 'h':
				d.focusPrev()
			}
		}

		// Pass to focused section
		if d.focusedSection >= 0 && d.focusedSection < len(d.sections) {
			section := d.sections[d.focusedSection]
			if section.Content != nil {
				if handler := section.Content.InputHandler(); handler != nil {
					handler(event, setFocus)
				}
			}
		}
	})
}

func (d *Dashboard) focusNext() {
	if len(d.focusable) == 0 {
		return
	}

	// Find current position in focusable list
	currentPos := -1
	for i, idx := range d.focusable {
		if idx == d.focusedSection {
			currentPos = i
			break
		}
	}

	nextPos := (currentPos + 1) % len(d.focusable)
	d.focusedSection = d.focusable[nextPos]
}

func (d *Dashboard) focusPrev() {
	if len(d.focusable) == 0 {
		return
	}

	// Find current position in focusable list
	currentPos := -1
	for i, idx := range d.focusable {
		if idx == d.focusedSection {
			currentPos = i
			break
		}
	}

	prevPos := currentPos - 1
	if prevPos < 0 {
		prevPos = len(d.focusable) - 1
	}
	d.focusedSection = d.focusable[prevPos]
}

// Focus handles focus.
func (d *Dashboard) Focus(delegate func(tview.Primitive)) {
	if d.focusedSection >= 0 && d.focusedSection < len(d.sections) {
		section := d.sections[d.focusedSection]
		if section.Content != nil {
			delegate(section.Content)
			return
		}
	}
	d.Box.Focus(delegate)
}

// HasFocus returns whether the dashboard has focus.
func (d *Dashboard) HasFocus() bool {
	for _, section := range d.sections {
		if section.Content != nil && section.Content.HasFocus() {
			return true
		}
	}
	return d.Box.HasFocus()
}

// MouseHandler handles mouse input.
func (d *Dashboard) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return d.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		if !d.InRect(event.Position()) {
			return false, nil
		}

		// Find which section was clicked
		mx, my := event.Position()
		for i, section := range d.sections {
			if section.Content != nil {
				cx, cy, cw, ch := section.Content.GetRect()
				if mx >= cx && mx < cx+cw && my >= cy && my < cy+ch {
					d.focusedSection = i
					if handler := section.Content.MouseHandler(); handler != nil {
						return handler(action, event, setFocus)
					}
					return true, section.Content
				}
			}
		}

		return false, nil
	})
}

// Ensure Dashboard implements nav.Component
var _ nav.Component = (*Dashboard)(nil)
