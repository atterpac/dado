package components

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/theme"
)

// ProgressModal is a modal dialog for long-running operations. Defaults to
// indeterminate (spinner) mode; call SetProgress(0.0–1.0) to switch to a
// determinate bar. Call Complete or Fail to update the state; the modal
// stays visible until Close is called or the user presses a key.
type ProgressModal struct {
	widgetBase

	// Content
	title      string
	message    string
	subMessage string
	progress   float64 // 0.0 to 1.0, or -1 for indeterminate

	// Configuration
	width         int
	cancelable    bool
	showBackdrop  bool
	indeterminate bool

	// State
	complete bool
	failed   bool
	err      error

	// Spinner animation
	spinnerFrame int
	spinnerStop  chan struct{}

	// Callbacks
	onCancel   func()
	onComplete func()
	onClose    func()

	// Focus management
}

// NewProgressModal creates a new progress modal
func NewProgressModal() *ProgressModal {
	p := &ProgressModal{
		width:        50,
		showBackdrop: true,
		progress:     -1, // Indeterminate by default
		spinnerStop:  make(chan struct{}),
	}
	p.initWidget()
	p.SetBorder(true)
	return p
}

// SetTitle sets the modal title
func (p *ProgressModal) SetTitle(title string) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.title = title
	return p
}

// SetWidth sets modal width
func (p *ProgressModal) SetWidth(width int) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.width = width
	return p
}

// SetCancelable enables/disables cancel functionality
func (p *ProgressModal) SetCancelable(cancelable bool) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cancelable = cancelable
	return p
}

// SetShowBackdrop enables/disables backdrop overlay
func (p *ProgressModal) SetShowBackdrop(show bool) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.showBackdrop = show
	return p
}

// SetIndeterminate sets indeterminate (spinner) mode
func (p *ProgressModal) SetIndeterminate(indeterminate bool) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.indeterminate = indeterminate
	if indeterminate {
		p.progress = -1
	}
	return p
}

// SetProgress sets the progress percentage (0.0 - 1.0)
// Values < 0 switch to indeterminate mode
func (p *ProgressModal) SetProgress(progress float64) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.progress = progress
	if progress < 0 {
		p.indeterminate = true
	} else {
		p.indeterminate = false
	}
	return p
}

// SetMessage sets the status message
func (p *ProgressModal) SetMessage(message string) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.message = message
	return p
}

// SetSubMessage sets a secondary/detail message
func (p *ProgressModal) SetSubMessage(message string) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subMessage = message
	return p
}

// Complete marks the operation as successfully completed
func (p *ProgressModal) Complete(message string) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.complete = true
	p.failed = false
	p.progress = 1.0
	p.message = message
	p.stopSpinner()
	if p.onComplete != nil {
		go p.onComplete()
	}
	return p
}

// Fail marks the operation as failed
func (p *ProgressModal) Fail(err error) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failed = true
	p.complete = false
	p.err = err
	if err != nil {
		p.message = err.Error()
	}
	p.stopSpinner()
	return p
}

// IsComplete returns true if operation completed successfully
func (p *ProgressModal) IsComplete() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.complete
}

// IsFailed returns true if operation failed
func (p *ProgressModal) IsFailed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.failed
}

// SetOnCancel is called when user cancels
func (p *ProgressModal) SetOnCancel(fn func()) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onCancel = fn
	return p
}

// SetOnComplete is called when Complete() is called
func (p *ProgressModal) SetOnComplete(fn func()) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onComplete = fn
	return p
}

// SetOnClose is called when modal is closed
func (p *ProgressModal) SetOnClose(fn func()) *ProgressModal {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onClose = fn
	return p
}

// StartSpinner starts the spinner animation for indeterminate mode
func (p *ProgressModal) StartSpinner() {
	p.mu.Lock()
	// Stop any existing spinner goroutine before starting a new one
	if p.spinnerStop != nil {
		select {
		case <-p.spinnerStop:
			// Already closed
		default:
			close(p.spinnerStop)
		}
	}
	p.spinnerStop = make(chan struct{})
	p.mu.Unlock()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-p.spinnerStop:
				return
			case <-ticker.C:
				p.mu.Lock()
				p.spinnerFrame = (p.spinnerFrame + 1) % len(spinnerFrames)
				p.mu.Unlock()
				theme.QueueUpdateDraw(func() {})
			}
		}
	}()
}

func (p *ProgressModal) stopSpinner() {
	select {
	case <-p.spinnerStop:
		// Already closed
	default:
		close(p.spinnerStop)
	}
}

// Close closes the modal
func (p *ProgressModal) Close() {
	p.mu.Lock()
	onClose := p.onClose
	p.stopSpinner()
	p.mu.Unlock()

	if onClose != nil {
		onClose()
	}
}

// Draw renders the progress modal
func (p *ProgressModal) Draw(screen tcell.Screen) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	th := p.th()
	bg := th.Bg()
	fg := th.Fg()
	dim := th.FgDim()
	border := th.BorderFocus()
	accent := th.Accent()
	success := th.Success()
	errColor := th.Error()
	warning := th.Warning()

	screenWidth, screenHeight := screen.Size()

	if p.showBackdrop {
		backdropStyle := tcell.StyleDefault.Background(bg).Foreground(dim)
		for y := range screenHeight {
			for x := range screenWidth {
				screen.SetContent(x, y, '░', nil, backdropStyle)
			}
		}
	}

	modalWidth := p.width
	modalHeight := 9
	if p.subMessage != "" {
		modalHeight++
	}

	modalX := (screenWidth - modalWidth) / 2
	modalY := (screenHeight - modalHeight) / 2

	bgStyle := tcell.StyleDefault.Background(bg).Foreground(fg)
	fillRect(screen, modalX, modalY, modalWidth, modalHeight, bgStyle)

	borderStyle := tcell.StyleDefault.Background(bg).Foreground(border)
	p.drawBorder(screen, modalX, modalY, modalWidth, modalHeight, borderStyle)

	titleStyle := tcell.StyleDefault.Background(bg).Foreground(fg).Bold(true)
	title := p.title
	if p.complete {
		title = title + " ✓"
		titleStyle = titleStyle.Foreground(success)
	} else if p.failed {
		title = title + " ✗"
		titleStyle = titleStyle.Foreground(errColor)
	}
	titleX := modalX + (modalWidth-len(title))/2
	p.drawText(screen, titleX, modalY+1, title, titleStyle)

	p.drawHorizontalLine(screen, modalX, modalY+2, modalWidth, borderStyle)

	progressY := modalY + 4
	if p.indeterminate && !p.complete && !p.failed {
		frames := spinnerFrames[SpinnerCircle]
		spinnerChar := frames[p.spinnerFrame%len(frames)]
		spinnerText := fmt.Sprintf("%s Loading...", spinnerChar)
		spinnerX := modalX + (modalWidth-len(spinnerText))/2
		spinnerStyle := tcell.StyleDefault.Background(bg).Foreground(warning)
		p.drawText(screen, spinnerX, progressY, spinnerText, spinnerStyle)
	} else {
		barWidth := modalWidth - 8
		barX := modalX + 3
		p.drawProgressBar(screen, barX, progressY, barWidth, bg, fg, dim, accent, errColor)
	}

	messageY := progressY + 2
	if p.message != "" {
		messageStyle := tcell.StyleDefault.Background(bg).Foreground(fg)
		if p.failed {
			messageStyle = messageStyle.Foreground(errColor)
		}
		msgX := modalX + (modalWidth-len(p.message))/2
		if len(p.message) > modalWidth-4 {
			p.message = p.message[:modalWidth-7] + "..."
			msgX = modalX + 2
		}
		p.drawText(screen, msgX, messageY, p.message, messageStyle)
	}

	if p.subMessage != "" {
		subMessageY := messageY + 1
		subStyle := tcell.StyleDefault.Background(bg).Foreground(dim)
		subX := modalX + (modalWidth-len(p.subMessage))/2
		if len(p.subMessage) > modalWidth-4 {
			p.subMessage = p.subMessage[:modalWidth-7] + "..."
			subX = modalX + 2
		}
		p.drawText(screen, subX, subMessageY, p.subMessage, subStyle)
	}

	footerY := modalY + modalHeight - 3
	p.drawHorizontalLine(screen, modalX, footerY, modalWidth, borderStyle)

	footerTextY := modalY + modalHeight - 2
	var footerText string
	if p.complete || p.failed {
		footerText = "[Enter] Close"
	} else if p.cancelable {
		footerText = "[Esc] Cancel"
	}
	if footerText != "" {
		footerStyle := tcell.StyleDefault.Background(bg).Foreground(dim)
		footerX := modalX + (modalWidth-len(footerText))/2
		p.drawText(screen, footerX, footerTextY, footerText, footerStyle)
	}
}

func (p *ProgressModal) drawProgressBar(screen tcell.Screen, x, y, width int, bg, _, dim, accent, errColor tcell.Color) {
	progress := p.progress
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	fillWidth := int(progress * float64(width))
	emptyWidth := width - fillWidth

	barColor := accent
	if p.failed {
		barColor = errColor
	}
	fillStyle := tcell.StyleDefault.Background(bg).Foreground(barColor)
	emptyStyle := tcell.StyleDefault.Background(bg).Foreground(dim)

	for i := range fillWidth {
		screen.SetContent(x+i, y, '█', nil, fillStyle)
	}
	for i := range emptyWidth {
		screen.SetContent(x+fillWidth+i, y, '░', nil, emptyStyle)
	}

	// Draw percentage overlaid centered on the bar, colored like the bar
	pct := fmt.Sprintf(" %.0f%% ", progress*100)
	pctX := x + (width-len(pct))/2
	pctStyle := tcell.StyleDefault.Background(bg).Foreground(barColor).Bold(true)
	for i, r := range pct {
		col := pctX + i
		if col < x || col >= x+width {
			continue
		}
		screen.SetContent(col, y, r, nil, pctStyle)
	}
}

func (p *ProgressModal) drawBorder(screen tcell.Screen, x, y, width, height int, style tcell.Style) {
	// Corners
	screen.SetContent(x, y, '╭', nil, style)
	screen.SetContent(x+width-1, y, '╮', nil, style)
	screen.SetContent(x, y+height-1, '╰', nil, style)
	screen.SetContent(x+width-1, y+height-1, '╯', nil, style)

	// Top and bottom edges
	for i := 1; i < width-1; i++ {
		screen.SetContent(x+i, y, '─', nil, style)
		screen.SetContent(x+i, y+height-1, '─', nil, style)
	}

	// Left and right edges
	for i := 1; i < height-1; i++ {
		screen.SetContent(x, y+i, '│', nil, style)
		screen.SetContent(x+width-1, y+i, '│', nil, style)
	}
}

func (p *ProgressModal) drawHorizontalLine(screen tcell.Screen, x, y, width int, style tcell.Style) {
	screen.SetContent(x, y, '├', nil, style)
	screen.SetContent(x+width-1, y, '┤', nil, style)
	for i := 1; i < width-1; i++ {
		screen.SetContent(x+i, y, '─', nil, style)
	}
}

func (p *ProgressModal) drawText(screen tcell.Screen, x, y int, text string, style tcell.Style) {
	for i, r := range text {
		screen.SetContent(x+i, y, r, nil, style)
	}
}

// Focus is called when the modal receives focus
// HasFocus returns whether the modal has focus
func (p *ProgressModal) HasFocus() bool {
	return p.Box.HasFocus()
}
