package components

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ProgressModal is a modal dialog for displaying progress of long-running operations
type ProgressModal struct {
	*tview.Box

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
	previousFocus tview.Primitive

	mu sync.RWMutex
}

// NewProgressModal creates a new progress modal
func NewProgressModal() *ProgressModal {
	p := &ProgressModal{
		Box:          tview.NewBox(),
		width:        50,
		showBackdrop: true,
		progress:     -1, // Indeterminate by default
		spinnerStop:  make(chan struct{}),
	}
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
func (p *ProgressModal) StartSpinner(app *tview.Application) {
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
				app.QueueUpdateDraw(func() {})
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

	screenWidth, screenHeight := screen.Size()

	// Draw backdrop if enabled
	if p.showBackdrop {
		backdropStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorDarkGray)
		for y := 0; y < screenHeight; y++ {
			for x := 0; x < screenWidth; x++ {
				screen.SetContent(x, y, '░', nil, backdropStyle)
			}
		}
	}

	// Calculate modal dimensions
	modalWidth := p.width
	modalHeight := 9 // Title + border + progress + messages + footer

	if p.subMessage != "" {
		modalHeight++
	}

	// Center modal
	modalX := (screenWidth - modalWidth) / 2
	modalY := (screenHeight - modalHeight) / 2

	// Draw modal background
	bgStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	for y := modalY; y < modalY+modalHeight; y++ {
		for x := modalX; x < modalX+modalWidth; x++ {
			screen.SetContent(x, y, ' ', nil, bgStyle)
		}
	}

	// Draw border
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	p.drawBorder(screen, modalX, modalY, modalWidth, modalHeight, borderStyle)

	// Draw title
	titleStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true)
	title := p.title
	if p.complete {
		title = title + " ✓"
		titleStyle = titleStyle.Foreground(tcell.ColorGreen)
	} else if p.failed {
		title = title + " ✗"
		titleStyle = titleStyle.Foreground(tcell.ColorRed)
	}
	titleX := modalX + (modalWidth-len(title))/2
	p.drawText(screen, titleX, modalY+1, title, titleStyle)

	// Draw separator
	p.drawHorizontalLine(screen, modalX, modalY+2, modalWidth, borderStyle)

	// Draw progress bar or spinner
	progressY := modalY + 4
	if p.indeterminate && !p.complete && !p.failed {
		// Draw spinner
		frames := spinnerFrames[SpinnerCircle]
		spinnerChar := frames[p.spinnerFrame%len(frames)]
		spinnerText := fmt.Sprintf("%s Loading...", spinnerChar)
		spinnerX := modalX + (modalWidth-len(spinnerText))/2
		spinnerStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
		p.drawText(screen, spinnerX, progressY, spinnerText, spinnerStyle)
	} else {
		// Draw progress bar
		barWidth := modalWidth - 8
		barX := modalX + 3
		p.drawProgressBar(screen, barX, progressY, barWidth)
	}

	// Draw message
	messageY := progressY + 2
	if p.message != "" {
		messageStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
		if p.failed {
			messageStyle = messageStyle.Foreground(tcell.ColorRed)
		}
		msgX := modalX + (modalWidth-len(p.message))/2
		if len(p.message) > modalWidth-4 {
			p.message = p.message[:modalWidth-7] + "..."
			msgX = modalX + 2
		}
		p.drawText(screen, msgX, messageY, p.message, messageStyle)
	}

	// Draw sub-message
	if p.subMessage != "" {
		subMessageY := messageY + 1
		subStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
		subX := modalX + (modalWidth-len(p.subMessage))/2
		if len(p.subMessage) > modalWidth-4 {
			p.subMessage = p.subMessage[:modalWidth-7] + "..."
			subX = modalX + 2
		}
		p.drawText(screen, subX, subMessageY, p.subMessage, subStyle)
	}

	// Draw footer separator
	footerY := modalY + modalHeight - 3
	p.drawHorizontalLine(screen, modalX, footerY, modalWidth, borderStyle)

	// Draw footer text
	footerTextY := modalY + modalHeight - 2
	var footerText string
	if p.complete || p.failed {
		footerText = "[Enter] Close"
	} else if p.cancelable {
		footerText = "[Esc] Cancel"
	}
	if footerText != "" {
		footerStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
		footerX := modalX + (modalWidth-len(footerText))/2
		p.drawText(screen, footerX, footerTextY, footerText, footerStyle)
	}
}

func (p *ProgressModal) drawProgressBar(screen tcell.Screen, x, y, width int) {
	progress := p.progress
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	fillWidth := int(progress * float64(width))
	emptyWidth := width - fillWidth

	fillStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	emptyStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)

	// Draw filled portion
	for i := 0; i < fillWidth; i++ {
		screen.SetContent(x+i, y, '█', nil, fillStyle)
	}

	// Draw empty portion
	for i := 0; i < emptyWidth; i++ {
		screen.SetContent(x+fillWidth+i, y, '░', nil, emptyStyle)
	}

	// Draw percentage
	pct := fmt.Sprintf("%3.0f%%", progress*100)
	pctX := x + width + 1
	pctStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	p.drawText(screen, pctX, y, pct, pctStyle)
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

// InputHandler handles keyboard input
func (p *ProgressModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		p.mu.RLock()
		complete := p.complete
		failed := p.failed
		cancelable := p.cancelable
		onCancel := p.onCancel
		onClose := p.onClose
		p.mu.RUnlock()

		switch event.Key() {
		case tcell.KeyEsc:
			if complete || failed {
				// Close on Esc when done
				if onClose != nil {
					onClose()
				}
			} else if cancelable && onCancel != nil {
				// Cancel operation
				onCancel()
			}
		case tcell.KeyEnter:
			if complete || failed {
				if onClose != nil {
					onClose()
				}
			}
		}
	})
}

// MouseHandler handles mouse events
func (p *ProgressModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return p.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Consume all mouse events to prevent interaction with background
		return true, nil
	})
}

// Focus is called when the modal receives focus
func (p *ProgressModal) Focus(delegate func(p tview.Primitive)) {
	p.Box.Focus(delegate)
}

// HasFocus returns whether the modal has focus
func (p *ProgressModal) HasFocus() bool {
	return p.Box.HasFocus()
}
