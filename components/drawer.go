package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/dado/theme"
)

// DrawerPosition specifies which edge the drawer appears from.
type DrawerPosition int

const (
	// DrawerRight positions the drawer on the right edge.
	DrawerRight DrawerPosition = iota
	// DrawerLeft positions the drawer on the left edge.
	DrawerLeft
)

// DrawerConfig configures drawer dimensions and behavior.
type DrawerConfig struct {
	Title    string
	Width    int            // Fixed width
	Position DrawerPosition // Which edge to attach to
	Backdrop bool           // Draw semi-transparent background
}

// Drawer is a slide-out panel that appears from an edge of the screen.
// Unlike Modal which is centered, Drawer attaches to a screen edge.
// Drawer implements the nav.ModalComponent interface for automatic lifecycle management.
type Drawer struct {
	*tview.Flex
	panel       *Panel
	hintBar     *KeyHintBar
	content     tview.Primitive
	focusTarget tview.Primitive
	config      DrawerConfig
	behavior    ModalBehavior
	onClose     func()
	onDismiss   func() bool
	subs        Subscriptions
}

// Subs returns the widget's subscription set; released by ComponentBase.Stop.
func (d *Drawer) Subs() *Subscriptions { return &d.subs }

// NewDrawer creates a new drawer with the given configuration.
func NewDrawer(config DrawerConfig) *Drawer {
	flex := tview.NewFlex()
	flex.SetBackgroundColor(theme.Bg())

	// Default width
	width := config.Width
	if width == 0 {
		width = 40
	}
	config.Width = width

	behavior := ModalBehavior{
		CapturesAllInput:      true,
		DismissOnEsc:          true,
		RestoreFocusOnDismiss: true,
		Backdrop:              config.Backdrop,
		BlockUntilDismissed:   false,
	}

	d := &Drawer{
		Flex:     flex,
		panel:    NewPanel(),
		hintBar:  NewKeyHintBar(),
		config:   config,
		behavior: behavior,
	}

	if config.Title != "" {
		d.panel.SetTitle(config.Title)
	}

	d.subs.Add(theme.Register(flex))
	d.setupLayout()

	return d
}

// setupLayout builds the drawer's internal structure.
func (d *Drawer) setupLayout() {
	// Inner content area with hint bar at bottom
	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	contentBox := tview.NewBox()
	innerFlex.AddItem(contentBox, 0, 1, true)
	innerFlex.AddItem(d.hintBar, 1, 0, false)
	d.panel.SetContent(innerFlex)

	// Build edge-aligned layout
	d.Flex.SetDirection(tview.FlexColumn)

	switch d.config.Position {
	case DrawerRight:
		// Left spacer (fills available space) | drawer panel (fixed width)
		d.Flex.AddItem(nil, 0, 1, false) // Transparent spacer
		d.Flex.AddItem(d.panel, d.config.Width, 0, true)
	case DrawerLeft:
		// Drawer panel (fixed width) | right spacer (fills available space)
		d.Flex.AddItem(d.panel, d.config.Width, 0, true)
		d.Flex.AddItem(nil, 0, 1, false) // Transparent spacer
	}
}

// SetContent sets the drawer's main content.
func (d *Drawer) SetContent(content tview.Primitive) *Drawer {
	d.content = content

	innerFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	innerFlex.AddItem(content, 0, 1, true)
	innerFlex.AddItem(d.hintBar, 1, 0, false)
	d.panel.SetContent(innerFlex)

	return d
}

// SetHints sets the key hints displayed at bottom.
func (d *Drawer) SetHints(hints []KeyHint) *Drawer {
	d.hintBar.SetHints(hints)
	return d
}

// SetOnClose sets callback when drawer closes.
func (d *Drawer) SetOnClose(fn func()) *Drawer {
	d.onClose = fn
	return d
}

// Close triggers the close callback.
func (d *Drawer) Close() {
	if d.onClose != nil {
		d.onClose()
	}
}

// Draw renders the drawer, optionally with backdrop.
func (d *Drawer) Draw(screen tcell.Screen) {
	d.Flex.SetBackgroundColor(theme.Bg())
	d.hintBar.SetBackgroundColor(theme.Bg())

	if d.config.Backdrop {
		d.drawBackdrop(screen)
	}
	d.Flex.Draw(screen)
}

// drawBackdrop draws a semi-transparent dark overlay on the non-drawer area.
func (d *Drawer) drawBackdrop(screen tcell.Screen) {
	x, y, width, height := d.GetRect()

	bgColor := theme.Bg()
	r, g, b := bgColor.RGB()
	darkBg := tcell.NewRGBColor(int32(r/2), int32(g/2), int32(b/2))
	style := tcell.StyleDefault.Background(darkBg)

	// Calculate the area to darken (everything except the drawer)
	var backdropX, backdropWidth int
	switch d.config.Position {
	case DrawerRight:
		backdropX = x
		backdropWidth = width - d.config.Width
	case DrawerLeft:
		backdropX = x + d.config.Width
		backdropWidth = width - d.config.Width
	}

	fillRect(screen, backdropX, y, backdropWidth, height, style)
}

// InputHandler handles input with drawer behavior.
func (d *Drawer) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if d.handleBaseInput(event) {
			return
		}

		if d.content != nil {
			if handler := d.content.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

// handleBaseInput handles Escape for close.
func (d *Drawer) handleBaseInput(event *tcell.EventKey) bool {
	if event.Key() == tcell.KeyEscape {
		d.Close()
		return true
	}
	return false
}

// WrapInputHandler wraps a custom handler with drawer's base handler.
func (d *Drawer) WrapInputHandler(handler func(*tcell.EventKey, func(tview.Primitive))) func(*tcell.EventKey, func(tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if d.handleBaseInput(event) {
			return
		}
		handler(event, setFocus)
	}
}

// Focus delegates to focusTarget, content, or panel.
func (d *Drawer) Focus(delegate func(tview.Primitive)) {
	if d.focusTarget != nil {
		delegate(d.focusTarget)
	} else if d.content != nil {
		delegate(d.content)
	} else {
		delegate(d.panel)
	}
}

// SetFocusOnShow sets a specific primitive to focus when the drawer is shown.
func (d *Drawer) SetFocusOnShow(p tview.Primitive) *Drawer {
	d.focusTarget = p
	return d
}

// GetPanel returns the drawer's panel for customization.
func (d *Drawer) GetPanel() *Panel {
	return d.panel
}

// GetHintBar returns the hint bar for direct manipulation.
func (d *Drawer) GetHintBar() *KeyHintBar {
	return d.hintBar
}

// GetBehavior returns the drawer's behavior configuration.
func (d *Drawer) GetBehavior() ModalBehavior {
	return d.behavior
}

// SetBehavior configures the drawer's behavior.
func (d *Drawer) SetBehavior(b ModalBehavior) *Drawer {
	d.behavior = b
	return d
}

// SetDismissOnEsc sets whether Escape key dismisses the drawer.
func (d *Drawer) SetDismissOnEsc(dismiss bool) *Drawer {
	d.behavior.DismissOnEsc = dismiss
	return d
}

// SetOnDismiss sets a handler called before the drawer is dismissed.
// Return false from the handler to cancel the dismiss.
func (d *Drawer) SetOnDismiss(fn func() bool) *Drawer {
	d.onDismiss = fn
	return d
}

// --- nav.Component interface implementation ---

// Start is called when the drawer becomes active.
func (d *Drawer) Start() {}

// Stop is called when the drawer becomes inactive.
func (d *Drawer) Stop() {}

// Hints returns key binding hints for this drawer.
func (d *Drawer) Hints() []KeyHint {
	hints := []KeyHint{}
	if d.behavior.DismissOnEsc {
		hints = append(hints, KeyHint{Key: "Esc", Description: "Close"})
	}
	return hints
}

// --- nav.Modal interface implementation ---

// ModalBehavior implements nav.Modal.
func (d *Drawer) ModalBehavior() ModalBehavior {
	return d.behavior
}

// OnDismiss implements nav.Modal.
func (d *Drawer) OnDismiss() bool {
	if d.onDismiss != nil {
		return d.onDismiss()
	}
	return true
}
