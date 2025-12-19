package components

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ToastLevel indicates the severity/type of notification
type ToastLevel int

const (
	ToastInfo    ToastLevel = iota // Blue info icon
	ToastSuccess                   // Green checkmark
	ToastWarning                   // Yellow warning
	ToastError                     // Red error
)

// ToastAction represents a clickable action in a toast
type ToastAction struct {
	Label   string
	Handler func()
}

// Toast represents a single notification
type Toast struct {
	ID        string
	Message   string
	Level     ToastLevel
	Duration  time.Duration // 0 = persistent (must dismiss manually)
	Actions   []ToastAction
	CreatedAt time.Time

	// State
	dismissed bool
	timer     *time.Timer
}

// Icon returns the icon for this toast level
func (t *Toast) Icon() string {
	switch t.Level {
	case ToastInfo:
		return "ℹ"
	case ToastSuccess:
		return "✓"
	case ToastWarning:
		return "⚠"
	case ToastError:
		return "✗"
	}
	return ""
}

// ToastPosition determines where toasts appear
type ToastPosition int

const (
	ToastTopRight ToastPosition = iota
	ToastTopLeft
	ToastBottomRight
	ToastBottomLeft
	ToastTopCenter
	ToastBottomCenter
)

// ToastManager manages all toasts in the application
type ToastManager struct {
	toasts          []*Toast
	position        ToastPosition
	maxVisible      int           // Max toasts shown at once
	maxWidth        int           // Max toast width
	defaultDuration time.Duration

	app *tview.Application // For QueueUpdateDraw

	// Callbacks
	onShow    func(toast *Toast)
	onDismiss func(toast *Toast)

	mu sync.RWMutex
}

// NewToastManager creates a new toast manager
func NewToastManager(app *tview.Application) *ToastManager {
	return &ToastManager{
		toasts:          make([]*Toast, 0),
		position:        ToastTopRight,
		maxVisible:      5,
		maxWidth:        40,
		defaultDuration: 3 * time.Second,
		app:             app,
	}
}

// SetPosition sets where toasts appear
func (m *ToastManager) SetPosition(pos ToastPosition) *ToastManager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.position = pos
	return m
}

// SetMaxVisible limits how many toasts show at once
func (m *ToastManager) SetMaxVisible(max int) *ToastManager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxVisible = max
	return m
}

// SetMaxWidth sets maximum toast width
func (m *ToastManager) SetMaxWidth(width int) *ToastManager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxWidth = width
	return m
}

// SetDefaultDuration sets default auto-dismiss time
func (m *ToastManager) SetDefaultDuration(d time.Duration) *ToastManager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDuration = d
	return m
}

// SetOnShow is called when a toast is displayed
func (m *ToastManager) SetOnShow(fn func(toast *Toast)) *ToastManager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onShow = fn
	return m
}

// SetOnDismiss is called when a toast is dismissed
func (m *ToastManager) SetOnDismiss(fn func(toast *Toast)) *ToastManager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onDismiss = fn
	return m
}

// Show displays a simple toast message
func (m *ToastManager) Show(message string, level ToastLevel) *Toast {
	return m.ShowWithDuration(message, level, m.defaultDuration)
}

// ShowWithDuration displays a toast with custom duration
func (m *ToastManager) ShowWithDuration(message string, level ToastLevel, duration time.Duration) *Toast {
	toast := &Toast{
		ID:        generateToastID(),
		Message:   message,
		Level:     level,
		Duration:  duration,
		CreatedAt: time.Now(),
	}

	m.mu.Lock()
	m.toasts = append(m.toasts, toast)
	onShow := m.onShow
	m.mu.Unlock()

	if onShow != nil {
		onShow(toast)
	}

	if duration > 0 && m.app != nil {
		toast.timer = time.AfterFunc(duration, func() {
			m.app.QueueUpdateDraw(func() {
				m.Dismiss(toast.ID)
			})
		})
	}

	return toast
}

// ShowPersistent displays a toast that must be manually dismissed
func (m *ToastManager) ShowPersistent(message string, level ToastLevel) *Toast {
	return m.ShowWithDuration(message, level, 0)
}

// ShowWithAction displays a toast with action buttons
func (m *ToastManager) ShowWithAction(message string, level ToastLevel, actions ...ToastAction) *Toast {
	toast := &Toast{
		ID:        generateToastID(),
		Message:   message,
		Level:     level,
		Duration:  0, // Persistent by default when actions are present
		Actions:   actions,
		CreatedAt: time.Now(),
	}

	m.mu.Lock()
	m.toasts = append(m.toasts, toast)
	onShow := m.onShow
	m.mu.Unlock()

	if onShow != nil {
		onShow(toast)
	}

	return toast
}

// ShowWithUndo displays a success toast with an Undo action
func (m *ToastManager) ShowWithUndo(message string, undoFn func()) *Toast {
	return m.ShowWithAction(message, ToastSuccess,
		ToastAction{Label: "Undo", Handler: undoFn},
		ToastAction{Label: "Dismiss", Handler: func() {}},
	)
}

// Info shows an info toast
func (m *ToastManager) Info(message string) *Toast {
	return m.Show(message, ToastInfo)
}

// Success shows a success toast
func (m *ToastManager) Success(message string) *Toast {
	return m.Show(message, ToastSuccess)
}

// Warning shows a warning toast
func (m *ToastManager) Warning(message string) *Toast {
	return m.Show(message, ToastWarning)
}

// Error shows an error toast
func (m *ToastManager) Error(message string) *Toast {
	return m.Show(message, ToastError)
}

// Dismiss removes a specific toast
func (m *ToastManager) Dismiss(id string) {
	m.mu.Lock()
	var dismissed *Toast
	for i, t := range m.toasts {
		if t.ID == id {
			dismissed = t
			t.dismissed = true
			if t.timer != nil {
				t.timer.Stop()
			}
			m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
			break
		}
	}
	onDismiss := m.onDismiss
	m.mu.Unlock()

	if dismissed != nil && onDismiss != nil {
		onDismiss(dismissed)
	}
}

// DismissAll removes all toasts
func (m *ToastManager) DismissAll() {
	m.mu.Lock()
	for _, t := range m.toasts {
		t.dismissed = true
		if t.timer != nil {
			t.timer.Stop()
		}
	}
	m.toasts = make([]*Toast, 0)
	m.mu.Unlock()
}

// GetActive returns all currently visible toasts
func (m *ToastManager) GetActive() []*Toast {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Toast, len(m.toasts))
	copy(result, m.toasts)
	return result
}

// HasActive returns true if any toasts are visible
func (m *ToastManager) HasActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.toasts) > 0
}

// Draw renders all visible toasts
func (m *ToastManager) Draw(screen tcell.Screen, screenWidth, screenHeight int) {
	m.mu.RLock()
	toasts := m.toasts
	maxVisible := m.maxVisible
	maxWidth := m.maxWidth
	position := m.position
	m.mu.RUnlock()

	if len(toasts) == 0 {
		return
	}

	// Limit to max visible
	visibleToasts := toasts
	if len(visibleToasts) > maxVisible {
		visibleToasts = visibleToasts[len(visibleToasts)-maxVisible:]
	}

	for i, toast := range visibleToasts {
		m.drawToast(screen, toast, screenWidth, screenHeight, maxWidth, position, i)
	}
}

func (m *ToastManager) drawToast(screen tcell.Screen, toast *Toast, screenWidth, screenHeight, maxWidth int, position ToastPosition, index int) {
	// Calculate toast dimensions
	toastWidth := len(toast.Message) + 4 // Icon + space + message + padding
	if toastWidth > maxWidth {
		toastWidth = maxWidth
	}
	if toastWidth < 20 {
		toastWidth = 20
	}

	toastHeight := 3 // Top border, content, bottom border
	hasActions := len(toast.Actions) > 0
	if hasActions {
		toastHeight = 4 // Add action row
	}

	// Calculate position
	padding := 2
	spacing := 1
	x, y := m.calculatePosition(screenWidth, screenHeight, toastWidth, toastHeight, position, padding, spacing, index)

	// Get level colors
	bgColor, fgColor, iconColor := m.getLevelColors(toast.Level)

	bgStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(iconColor)
	iconStyle := tcell.StyleDefault.Background(bgColor).Foreground(iconColor)

	// Draw background
	for row := y; row < y+toastHeight; row++ {
		for col := x; col < x+toastWidth; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw border
	m.drawToastBorder(screen, x, y, toastWidth, toastHeight, borderStyle)

	// Draw icon
	icon := toast.Icon()
	for i, r := range icon {
		screen.SetContent(x+2+i, y+1, r, nil, iconStyle)
	}

	// Draw message (truncate if needed)
	message := toast.Message
	maxMsgLen := toastWidth - 6 // Account for icon, spaces, padding
	if len(message) > maxMsgLen {
		message = message[:maxMsgLen-3] + "..."
	}
	msgStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	msgX := x + 4 // After icon
	for i, r := range message {
		screen.SetContent(msgX+i, y+1, r, nil, msgStyle)
	}

	// Draw actions if present
	if hasActions {
		actionY := y + 2
		actionX := x + toastWidth - 2

		for i := len(toast.Actions) - 1; i >= 0; i-- {
			action := toast.Actions[i]
			label := "[" + action.Label + "]"
			actionX -= len(label) + 1

			actionStyle := tcell.StyleDefault.Background(bgColor).Foreground(tcell.ColorBlue)
			for j, r := range label {
				screen.SetContent(actionX+j, actionY, r, nil, actionStyle)
			}
		}
	}
}

func (m *ToastManager) calculatePosition(screenWidth, screenHeight, toastWidth, toastHeight int, position ToastPosition, padding, spacing, index int) (x, y int) {
	switch position {
	case ToastTopRight:
		x = screenWidth - toastWidth - padding
		y = padding + index*(toastHeight+spacing)
	case ToastBottomRight:
		x = screenWidth - toastWidth - padding
		y = screenHeight - (index+1)*(toastHeight+spacing) - padding
	case ToastTopLeft:
		x = padding
		y = padding + index*(toastHeight+spacing)
	case ToastBottomLeft:
		x = padding
		y = screenHeight - (index+1)*(toastHeight+spacing) - padding
	case ToastTopCenter:
		x = (screenWidth - toastWidth) / 2
		y = padding + index*(toastHeight+spacing)
	case ToastBottomCenter:
		x = (screenWidth - toastWidth) / 2
		y = screenHeight - (index+1)*(toastHeight+spacing) - padding
	}

	// Ensure within bounds
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+toastWidth > screenWidth {
		x = screenWidth - toastWidth
	}
	if y+toastHeight > screenHeight {
		y = screenHeight - toastHeight
	}

	return x, y
}

func (m *ToastManager) getLevelColors(level ToastLevel) (bg, fg, icon tcell.Color) {
	switch level {
	case ToastInfo:
		return tcell.ColorNavy, tcell.ColorWhite, tcell.ColorBlue
	case ToastSuccess:
		return tcell.ColorDarkGreen, tcell.ColorWhite, tcell.ColorGreen
	case ToastWarning:
		return tcell.ColorOlive, tcell.ColorWhite, tcell.ColorYellow
	case ToastError:
		return tcell.ColorDarkRed, tcell.ColorWhite, tcell.ColorRed
	default:
		return tcell.ColorBlack, tcell.ColorWhite, tcell.ColorWhite
	}
}

func (m *ToastManager) drawToastBorder(screen tcell.Screen, x, y, width, height int, style tcell.Style) {
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

// generateToastID generates a unique toast ID
func generateToastID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return "toast_" + hex.EncodeToString(bytes)
}

// HandleAction triggers an action by index for the most recent toast
func (m *ToastManager) HandleAction(actionIndex int) bool {
	m.mu.RLock()
	if len(m.toasts) == 0 {
		m.mu.RUnlock()
		return false
	}
	toast := m.toasts[len(m.toasts)-1]
	if actionIndex >= len(toast.Actions) {
		m.mu.RUnlock()
		return false
	}
	action := toast.Actions[actionIndex]
	toastID := toast.ID
	m.mu.RUnlock()

	if action.Handler != nil {
		action.Handler()
	}
	m.Dismiss(toastID)
	return true
}
