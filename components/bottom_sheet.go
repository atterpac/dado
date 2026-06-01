package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// BottomSheetConfig configures bottom sheet dimensions and behavior.
type BottomSheetConfig struct {
	Title    string
	Height   int  // Fixed height (0 = auto-size to content)
	Backdrop bool // Draw semi-transparent background above the sheet
}

// BottomSheet is a panel that appears anchored to the bottom of the screen.
// Unlike Modal which is centered, BottomSheet attaches to the bottom edge
// and spans the full width (with small horizontal margins).
// BottomSheet implements the nav.ModalComponent interface for automatic lifecycle management.
type BottomSheet struct {
	*core.Flex
	panel       *Panel
	hintBar     *KeyHintBar
	content     core.Widget
	focusTarget core.Widget
	config      BottomSheetConfig
	behavior    ModalBehavior
	onClose     func()
	onDismiss   func() bool
	subs        Subscriptions
}

// Subs returns the widget's subscription set; released by ComponentBase.Stop.
func (b *BottomSheet) Subs() *Subscriptions { return &b.subs }

// NewBottomSheet creates a new bottom sheet with the given configuration.
func NewBottomSheet(config BottomSheetConfig) *BottomSheet {
	flex := core.NewFlex()
	flex.Box.SetBackgroundColor(tcell.ColorDefault)

	// Default height
	height := config.Height
	if height == 0 {
		height = 10
	}
	config.Height = height

	behavior := ModalBehavior{
		CapturesAllInput:      true,
		DismissOnEsc:          true,
		RestoreFocusOnDismiss: true,
		Backdrop:              config.Backdrop,
		BlockUntilDismissed:   false,
	}

	b := &BottomSheet{
		Flex:     flex,
		panel:    NewPanel(),
		hintBar:  NewKeyHintBar(),
		config:   config,
		behavior: behavior,
	}

	if config.Title != "" {
		b.panel.SetTitle(config.Title)
	}

	b.subs.Add(theme.RegisterFn(func(c tcell.Color) { flex.Box.SetBackgroundColor(c) }))
	b.setupLayout()

	return b
}

// setupLayout builds the bottom sheet's internal structure.
func (b *BottomSheet) setupLayout() {
	// Inner content area with hint bar at bottom
	innerFlex := core.NewFlex().SetDirection(core.Column)
	contentBox := new(core.Box)
	innerFlex.AddItem(contentBox, 0, 1, true)
	innerFlex.AddItem(b.hintBar, 1, 0, false)
	b.panel.SetContent(innerFlex)

	// Build bottom-anchored layout: top spacer (fills) | panel at bottom
	b.Flex.SetDirection(core.Column)
	b.Flex.AddItem(new(core.Box), 0, 1, false)        // Top spacer (transparent)
	b.Flex.AddItem(b.panel, b.config.Height, 0, true) // Panel at bottom
}

// SetContent sets the bottom sheet's main content.
func (b *BottomSheet) SetContent(content core.Widget) *BottomSheet {
	b.content = content

	innerFlex := core.NewFlex().SetDirection(core.Column)
	innerFlex.AddItem(content, 0, 1, true)
	innerFlex.AddItem(b.hintBar, 1, 0, false)
	b.panel.SetContent(innerFlex)

	return b
}

// SetHeight updates the sheet height and rebuilds layout.
func (b *BottomSheet) SetHeight(height int) *BottomSheet {
	b.config.Height = height

	// Rebuild flex layout with new height
	b.Flex.Clear()
	b.Flex.SetDirection(core.Column)
	b.Flex.AddItem(new(core.Box), 0, 1, false)
	b.Flex.AddItem(b.panel, height, 0, true)

	return b
}

// SetHints sets the key hints displayed at bottom.
func (b *BottomSheet) SetHints(hints []KeyHint) *BottomSheet {
	b.hintBar.SetHints(hints)
	return b
}

// SetOnClose sets callback when bottom sheet closes.
func (b *BottomSheet) SetOnClose(fn func()) *BottomSheet {
	b.onClose = fn
	return b
}

// Close triggers the close callback.
func (b *BottomSheet) Close() {
	if b.onClose != nil {
		b.onClose()
	}
}

// Draw renders the bottom sheet, optionally with backdrop.
func (b *BottomSheet) Draw(screen tcell.Screen) {
	b.Flex.Box.SetBackgroundColor(theme.Bg())
	b.hintBar.SetBackgroundColor(theme.Bg())

	if b.config.Backdrop {
		b.drawBackdrop(screen)
	}
	b.Flex.Draw(screen)
}

// drawBackdrop draws a semi-transparent dark overlay above the sheet.
func (b *BottomSheet) drawBackdrop(screen tcell.Screen) {
	x, y, width, height := b.GetRect()

	bgColor := theme.Bg()
	r, g, bb := bgColor.RGB()
	darkBg := tcell.NewRGBColor(int32(r/2), int32(g/2), int32(bb/2))
	style := tcell.StyleDefault.Background(darkBg)

	// Darken the area above the sheet
	backdropHeight := height - b.config.Height
	for row := y; row < y+backdropHeight; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, style)
		}
	}
}

// handleBaseInput handles Escape for close.
func (b *BottomSheet) handleBaseInput(event *tcell.EventKey) bool {
	if event.Key() == tcell.KeyEscape {
		b.Close()
		return true
	}
	return false
}

// Focus delegates focus to the box.
func (b *BottomSheet) Focus() {
	b.Flex.Box.Focus()
}

// SetFocusOnShow sets a specific widget to focus when the bottom sheet is shown.
func (b *BottomSheet) SetFocusOnShow(p core.Widget) *BottomSheet {
	b.focusTarget = p
	return b
}

// GetPanel returns the bottom sheet's panel for customization.
func (b *BottomSheet) GetPanel() *Panel {
	return b.panel
}

// GetHintBar returns the hint bar for direct manipulation.
func (b *BottomSheet) GetHintBar() *KeyHintBar {
	return b.hintBar
}

// GetBehavior returns the bottom sheet's behavior configuration.
func (b *BottomSheet) GetBehavior() ModalBehavior {
	return b.behavior
}

// SetBehavior configures the bottom sheet's behavior.
func (b *BottomSheet) SetBehavior(beh ModalBehavior) *BottomSheet {
	b.behavior = beh
	return b
}

// SetDismissOnEsc sets whether Escape key dismisses the bottom sheet.
func (b *BottomSheet) SetDismissOnEsc(dismiss bool) *BottomSheet {
	b.behavior.DismissOnEsc = dismiss
	return b
}

// SetOnDismiss sets a handler called before the bottom sheet is dismissed.
// Return false from the handler to cancel the dismiss.
func (b *BottomSheet) SetOnDismiss(fn func() bool) *BottomSheet {
	b.onDismiss = fn
	return b
}

// --- nav.Component interface implementation ---

// Name returns the bottom sheet title for breadcrumbs.
func (b *BottomSheet) Name() string {
	return b.config.Title
}

// Start is called when the bottom sheet becomes active.
func (b *BottomSheet) Start() {}

// Stop is called when the bottom sheet becomes inactive.
func (b *BottomSheet) Stop() {}

// Hints returns key binding hints for this bottom sheet.
func (b *BottomSheet) Hints() []KeyHint {
	hints := []KeyHint{}
	if b.behavior.DismissOnEsc {
		hints = append(hints, KeyHint{Key: "Esc", Description: "Close"})
	}
	return hints
}

// --- nav.Modal interface implementation ---

// ModalBehavior implements nav.Modal.
func (b *BottomSheet) ModalBehavior() ModalBehavior {
	return b.behavior
}

// OnDismiss implements nav.Modal.
func (b *BottomSheet) OnDismiss() bool {
	if b.onDismiss != nil {
		return b.onDismiss()
	}
	return true
}
