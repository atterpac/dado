package nav

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/bus"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// publishNav emits a navigation event when the bus is enabled.
func (p *Pages) publishNav(kind, op, name string) {
	if !bus.Enabled() {
		return
	}
	bus.Publish(bus.Event{
		Kind:    kind,
		Source:  bus.SourceNav,
		Payload: bus.PageNav{Op: op, Name: name, Depth: len(p.stack)},
	})
}

// Pages manages stack-based page navigation with automatic modal handling.
// It implements core.Widget so it can be placed directly inside a core.Flex.
type Pages struct {
	x, y, w, h int
	bgColor     tcell.Color
	hasFocus    bool

	stack          []Component
	focusStack     []core.Widget // saved focus for modal restoration
	onChange       func(Component)
	onModalDismiss func(Modal)
	counter        int
	fm             *core.FocusManager
	crumbs         *Crumbs
	subs           components.Subscriptions
}

// Subs returns the widget's subscription set; release on app teardown.
func (p *Pages) Subs() *components.Subscriptions { return &p.subs }

// NewPages creates a new page stack manager.
func NewPages() *Pages {
	p := &Pages{
		stack:      make([]Component, 0),
		focusStack: make([]core.Widget, 0),
		bgColor:    theme.Bg(),
	}
	p.subs.Add(theme.RegisterFn(func(c tcell.Color) { p.bgColor = c }))
	return p
}


// --- core.Widget implementation ---

// Draw renders the background and the front component.
func (p *Pages) Draw(screen tcell.Screen) {
	bg := theme.Bg()
	style := tcell.StyleDefault.Background(bg)
	for row := p.y; row < p.y+p.h; row++ {
		for col := p.x; col < p.x+p.w; col++ {
			screen.SetContent(col, row, ' ', nil, style)
		}
	}
	if c := p.Current(); c != nil {
		c.SetRect(p.x, p.y, p.w, p.h)
		c.Draw(screen)
	}
}

// Rect returns the current geometry.
func (p *Pages) Rect() (x, y, w, h int) { return p.x, p.y, p.w, p.h }

// SetRect sets the geometry. Passed down to the front component on Draw.
func (p *Pages) SetRect(x, y, w, h int) { p.x = x; p.y = y; p.w = w; p.h = h }

// Focus marks Pages as focused and forwards to the front component.
func (p *Pages) Focus() { p.hasFocus = true }

// Blur removes focus.
func (p *Pages) Blur() { p.hasFocus = false }

// HasFocus returns whether Pages is focused.
func (p *Pages) HasFocus() bool { return p.hasFocus }

// HandleKey routes the key event to the front component's HandleKey.
func (p *Pages) HandleKey(ev *tcell.EventKey) bool {
	c := p.Current()
	if c == nil {
		return false
	}
	if kh, ok := c.(core.KeyHandler); ok {
		return kh.HandleKey(ev)
	}
	return false
}

// HandleMouse routes mouse events to the front component.
func (p *Pages) HandleMouse(action core.MouseAction, ev *tcell.EventMouse) (bool, core.Widget) {
	c := p.Current()
	if c == nil {
		return false, nil
	}
	if mh, ok := c.(core.MouseHandler); ok {
		return mh.HandleMouse(action, ev)
	}
	return false, nil
}

// --- Focus manager wiring ---

// SetFocusManager sets the core.FocusManager used for modal focus save/restore.
// Call from layout.App after creating the Pages.
func (p *Pages) SetFocusManager(fm *core.FocusManager) {
	p.fm = fm
}

// SetOnModalDismiss sets a callback that fires when any modal is dismissed.
func (p *Pages) SetOnModalDismiss(fn func(Modal)) {
	p.onModalDismiss = fn
}

// Push adds a component to the stack and shows it.
// Calls Stop() on the previous component if any.
// If the component implements Modal, modal behavior is applied automatically.
func (p *Pages) Push(c Component) {
	// Check if a blocking modal is active
	if p.hasBlockingModal() {
		return
	}

	// Stop current component
	if len(p.stack) > 0 {
		current := p.stack[len(p.stack)-1]
		current.Stop()

		// If pushing a modal, save current focus for restoration
		if IsModal(c) && p.fm != nil {
			p.focusStack = append(p.focusStack, p.fm.Focused())
		}
	}

	p.counter++
	_ = fmt.Sprintf("page-%d", p.counter) // preserve counter semantics

	// Add to stack
	p.stack = append(p.stack, c)

	// Start the new component
	c.Start()

	// Notify listener
	p.notifyChange()
	p.publishNav(bus.KindNavPush, "push", c.Name())
}

// Pop removes the current component and returns to previous.
// Calls Stop() on current, Start() on previous.
// Returns false if stack is empty, only has one item, or if a modal's OnDismiss() returns false.
func (p *Pages) Pop() bool {
	if len(p.stack) <= 1 {
		return false
	}

	current := p.stack[len(p.stack)-1]

	// Handle modal dismiss
	if modal := AsModal(current); modal != nil {
		if !modal.OnDismiss() {
			return false
		}
		if p.onModalDismiss != nil {
			p.onModalDismiss(modal)
		}

		// Restore focus if configured
		behavior := modal.ModalBehavior()
		if behavior.RestoreFocusOnDismiss && len(p.focusStack) > 0 && p.fm != nil {
			restoreTo := p.focusStack[len(p.focusStack)-1]
			p.focusStack = p.focusStack[:len(p.focusStack)-1]
			if restoreTo != nil {
				p.fm.Focus(restoreTo)
			}
		}
	}

	// Stop and remove current
	current.Stop()
	p.stack = p.stack[:len(p.stack)-1]

	// Show and start previous
	var prevName string
	if len(p.stack) > 0 {
		prev := p.stack[len(p.stack)-1]
		prev.Start()
		prevName = prev.Name()
	}

	// Notify listener
	p.notifyChange()
	p.publishNav(bus.KindNavPop, "pop", prevName)

	return true
}

// Current returns the active component, or nil if stack is empty.
func (p *Pages) Current() Component {
	if len(p.stack) == 0 {
		return nil
	}
	return p.stack[len(p.stack)-1]
}

// Clear removes all components from the stack.
// Calls Stop() on each component.
func (p *Pages) Clear() {
	for i := len(p.stack) - 1; i >= 0; i-- {
		p.stack[i].Stop()
	}
	p.stack = make([]Component, 0)
	p.notifyChange()
}

// StackDepth returns the number of components in stack.
func (p *Pages) StackDepth() int {
	return len(p.stack)
}

// SetOnChange sets callback when active component changes.
func (p *Pages) SetOnChange(fn func(Component)) {
	p.onChange = fn
}

// SetCrumbs sets the breadcrumb component to update on navigation.
func (p *Pages) SetCrumbs(crumbs *Crumbs) {
	p.crumbs = crumbs
}

// notifyChange calls the onChange callback if set and updates crumbs.
func (p *Pages) notifyChange() {
	if p.crumbs != nil {
		var path []string
		for _, c := range p.stack {
			if !IsModal(c) {
				path = append(path, c.Name())
			}
		}
		p.crumbs.SetPath(path)
	}
	if p.onChange != nil {
		p.onChange(p.Current())
	}
}

// CanPop returns true if there's a previous page to return to.
func (p *Pages) CanPop() bool {
	return len(p.stack) > 1
}

// GetStack returns a copy of the component stack.
func (p *Pages) GetStack() []Component {
	result := make([]Component, len(p.stack))
	copy(result, p.stack)
	return result
}

// Replace replaces the current component without affecting stack depth.
func (p *Pages) Replace(c Component) {
	if len(p.stack) == 0 {
		p.Push(c)
		return
	}
	if p.hasBlockingModal() {
		return
	}

	current := p.stack[len(p.stack)-1]
	current.Stop()
	p.stack[len(p.stack)-1] = c
	c.Start()

	p.notifyChange()
	p.publishNav(bus.KindNavReplace, "replace", c.Name())
}

// CurrentIsModal returns true if the current (front) page is a modal.
func (p *Pages) CurrentIsModal() bool {
	if c := p.Current(); c != nil {
		return IsModal(c)
	}
	return false
}

// CurrentModalBehavior returns the modal behavior if the current page is a modal.
func (p *Pages) CurrentModalBehavior() *components.ModalBehavior {
	if c := p.Current(); c != nil {
		return GetModalBehavior(c)
	}
	return nil
}

// DismissModal attempts to dismiss the current modal.
func (p *Pages) DismissModal() bool {
	if !p.CurrentIsModal() {
		return false
	}
	return p.Pop()
}

// hasBlockingModal returns true if a blocking modal is currently active.
func (p *Pages) hasBlockingModal() bool {
	if behavior := p.CurrentModalBehavior(); behavior != nil {
		return behavior.BlockUntilDismissed
	}
	return false
}

// ModalCount returns the number of modals currently in the stack.
func (p *Pages) ModalCount() int {
	count := 0
	for _, c := range p.stack {
		if IsModal(c) {
			count++
		}
	}
	return count
}

// HasModal returns true if any modal is currently in the stack.
func (p *Pages) HasModal() bool {
	for _, c := range p.stack {
		if IsModal(c) {
			return true
		}
	}
	return false
}
