package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
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
	*core.Flex
	panel       *Panel
	hintBar     *KeyHintBar
	content     core.Widget
	focusTarget core.Widget
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
	flex := core.NewFlex()
	flex.Box.SetBackgroundColor(theme.Bg())

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

	d.subs.Add(theme.RegisterFn(func(c tcell.Color) { flex.Box.SetBackgroundColor(c) }))
	d.setupLayout()

	return d
}

// setupLayout builds the drawer's internal structure.
func (d *Drawer) setupLayout() {
	// Inner content area with hint bar at bottom
	innerFlex := core.NewFlex().SetDirection(core.Column)
	contentBox := new(core.Box)
	innerFlex.AddItem(contentBox, 0, 1, true)
	innerFlex.AddItem(d.hintBar, 1, 0, false)
	d.panel.SetContent(innerFlex)

	// Build edge-aligned layout
	d.Flex.SetDirection(core.Row)

	switch d.config.Position {
	case DrawerRight:
		// Left spacer (fills available space) | drawer panel (fixed width)
		d.Flex.AddItem(new(core.Box), 0, 1, false) // Transparent spacer
		d.Flex.AddItem(d.panel, d.config.Width, 0, true)
	case DrawerLeft:
		// Drawer panel (fixed width) | right spacer (fills available space)
		d.Flex.AddItem(d.panel, d.config.Width, 0, true)
		d.Flex.AddItem(new(core.Box), 0, 1, false) // Transparent spacer
	}
}

// SetContent sets the drawer's main content.
func (d *Drawer) SetContent(content core.Widget) *Drawer {
	d.content = content

	innerFlex := core.NewFlex().SetDirection(core.Column)
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
	d.Flex.Box.SetBackgroundColor(theme.Bg())
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

// HandleKey processes a key event for the Drawer.
func (d *Drawer) HandleKey(ev *tcell.EventKey) bool {
	if d.handleBaseInput(ev) {
		return true
	}

	if d.content != nil {
		if kh, ok := d.content.(interface{ HandleKey(*tcell.EventKey) bool }); ok {
			kh.HandleKey(ev)
		}
	}
	return false
}

// handleBaseInput handles Escape for close.
func (d *Drawer) handleBaseInput(event *tcell.EventKey) bool {
	if event.Key() == tcell.KeyEscape {
		d.Close()
		return true
	}
	return false
}

// Focus delegates focus to the box.
func (d *Drawer) Focus() {
	d.Flex.Box.Focus()
}

// SetFocusOnShow sets a specific widget to focus when the drawer is shown.
func (d *Drawer) SetFocusOnShow(p core.Widget) *Drawer {
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
