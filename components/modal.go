package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/theme"
)

// ModalConfig configures modal dimensions and behavior.
type ModalConfig struct {
	Title     string
	Width     int  // Fixed width (0 = use min/max calculation)
	Height    int  // Fixed height (0 = use min/max calculation)
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
	Backdrop  bool // Draw dark semi-transparent background
}

// Modal is a configurable modal dialog with centered positioning.
type Modal struct {
	*tview.Flex
	panel       *Panel
	hintBar     *KeyHintBar
	content     tview.Primitive
	focusTarget tview.Primitive // Optional: specific primitive to focus when modal opens
	config      ModalConfig
	onClose     func()
	onSubmit    func()
	onCancel    func()
}

// NewModal creates a new modal with the given configuration.
func NewModal(config ModalConfig) *Modal {
	flex := tview.NewFlex()
	flex.SetBackgroundColor(theme.Bg())

	m := &Modal{
		Flex:    flex,
		panel:   NewPanel(),
		hintBar: NewKeyHintBar(),
		config:  config,
	}

	if config.Title != "" {
		m.panel.SetTitle(config.Title)
	}

	// Register for automatic theme updates (Panel and KeyHintBar register themselves)
	theme.Register(flex)

	m.setupLayout()
	return m
}

// setupLayout builds the modal's internal structure.
func (m *Modal) setupLayout() {
	// Inner content area
	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Content placeholder (will be replaced by SetContent)
	contentBox := tview.NewBox()
	innerFlex.AddItem(contentBox, 0, 1, true)

	// Hint bar at bottom
	innerFlex.AddItem(m.hintBar, 1, 0, false)

	m.panel.SetContent(innerFlex)

	// Calculate dimensions
	width := m.config.Width
	height := m.config.Height

	if width == 0 {
		width = m.config.MinWidth
		if width == 0 {
			width = 40
		}
	}
	if height == 0 {
		height = m.config.MinHeight
		if height == 0 {
			height = 10
		}
	}

	// Clamp to max
	if m.config.MaxWidth > 0 && width > m.config.MaxWidth {
		width = m.config.MaxWidth
	}
	if m.config.MaxHeight > 0 && height > m.config.MaxHeight {
		height = m.config.MaxHeight
	}

	// Build centering layout
	m.Flex.SetDirection(tview.FlexRow)
	m.Flex.AddItem(nil, 0, 1, false) // Top spacer

	centerRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	centerRow.AddItem(nil, 0, 1, false)      // Left spacer
	centerRow.AddItem(m.panel, width, 0, true) // Modal panel
	centerRow.AddItem(nil, 0, 1, false)      // Right spacer

	m.Flex.AddItem(centerRow, height, 0, true)
	m.Flex.AddItem(nil, 0, 1, false) // Bottom spacer
}

// SetContent sets the modal's main content.
func (m *Modal) SetContent(content tview.Primitive) *Modal {
	m.content = content

	// Rebuild inner flex with new content
	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	innerFlex.AddItem(content, 0, 1, true)
	innerFlex.AddItem(m.hintBar, 1, 0, false)
	m.panel.SetContent(innerFlex)

	return m
}

// SetHints sets the key hints displayed at bottom.
func (m *Modal) SetHints(hints []KeyHint) *Modal {
	m.hintBar.SetHints(hints)
	return m
}

// SetOnClose sets callback when modal closes.
func (m *Modal) SetOnClose(fn func()) *Modal {
	m.onClose = fn
	return m
}

// SetOnSubmit sets callback for submit action.
func (m *Modal) SetOnSubmit(fn func()) *Modal {
	m.onSubmit = fn
	return m
}

// SetOnCancel sets callback for cancel action.
func (m *Modal) SetOnCancel(fn func()) *Modal {
	m.onCancel = fn
	return m
}

// Close triggers the close callback.
func (m *Modal) Close() {
	if m.onClose != nil {
		m.onClose()
	}
}

// Submit triggers the submit callback.
func (m *Modal) Submit() {
	if m.onSubmit != nil {
		m.onSubmit()
	}
}

// Cancel triggers the cancel callback.
func (m *Modal) Cancel() {
	if m.onCancel != nil {
		m.onCancel()
	}
}

// Draw renders the modal, optionally with backdrop.
func (m *Modal) Draw(screen tcell.Screen) {
	// Update background colors from theme
	m.Flex.SetBackgroundColor(theme.Bg())
	m.hintBar.SetBackgroundColor(theme.Bg())

	if m.config.Backdrop {
		m.drawBackdrop(screen)
	}
	m.Flex.Draw(screen)
}

// drawBackdrop draws a semi-transparent dark overlay.
func (m *Modal) drawBackdrop(screen tcell.Screen) {
	x, y, width, height := m.GetRect()

	// Use a dark background color
	bgColor := theme.Bg()
	r, g, b := bgColor.RGB()
	// Darken the background
	darkBg := tcell.NewRGBColor(int32(r/2), int32(g/2), int32(b/2))

	style := tcell.StyleDefault.Background(darkBg)

	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, style)
		}
	}
}

// InputHandler handles input with base modal behavior.
func (m *Modal) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// Base modal input handling
		handled := m.handleBaseInput(event)
		if handled {
			return
		}

		// Delegate to content
		if m.content != nil {
			if handler := m.content.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

// handleBaseInput handles Enter for submit and Esc for cancel.
func (m *Modal) handleBaseInput(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyEnter:
		if m.onSubmit != nil {
			m.onSubmit()
			return true
		}
	case tcell.KeyEscape:
		if m.onCancel != nil {
			m.onCancel()
		} else {
			m.Close()
		}
		return true
	}
	return false
}

// WrapInputHandler wraps a custom handler with modal's base handler.
func (m *Modal) WrapInputHandler(handler func(*tcell.EventKey, func(tview.Primitive))) func(*tcell.EventKey, func(tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// First try base handling
		if m.handleBaseInput(event) {
			return
		}
		// Then custom handler
		handler(event, setFocus)
	}
}

// Focus delegates to focusTarget, content, or panel.
func (m *Modal) Focus(delegate func(tview.Primitive)) {
	if m.focusTarget != nil {
		delegate(m.focusTarget)
	} else if m.content != nil {
		delegate(m.content)
	} else {
		delegate(m.panel)
	}
}

// SetFocusOnShow sets a specific primitive to focus when the modal is shown.
// This is useful when the content is a container and you want to focus a child.
func (m *Modal) SetFocusOnShow(p tview.Primitive) *Modal {
	m.focusTarget = p
	return m
}

// GetPanel returns the modal's panel for customization.
func (m *Modal) GetPanel() *Panel {
	return m.panel
}

// GetHintBar returns the hint bar for direct manipulation.
func (m *Modal) GetHintBar() *KeyHintBar {
	return m.hintBar
}
