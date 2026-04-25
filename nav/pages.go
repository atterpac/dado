package nav

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/bus"
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
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
type Pages struct {
	*tview.Pages
	stack          []Component
	focusStack     []tview.Primitive // Saved focus for modal restoration
	onChange       func(Component)
	onModalDismiss func(Modal) // Optional callback when any modal dismisses
	counter        int         // For generating unique page names
	app            *tview.Application // Reference for focus management
	crumbs         *Crumbs            // Optional breadcrumb component
	subs           components.Subscriptions
}

// Subs returns the widget's subscription set; release on app teardown.
func (p *Pages) Subs() *components.Subscriptions { return &p.subs }

// NewPages creates a new page stack manager.
func NewPages() *Pages {
	pages := tview.NewPages()
	pages.SetBackgroundColor(theme.Bg())

	p := &Pages{
		Pages:      pages,
		stack:      make([]Component, 0),
		focusStack: make([]tview.Primitive, 0),
	}

	p.subs.Add(theme.Register(pages))

	return p
}

// SetApplication sets the tview.Application reference for focus management.
// This should be called by App during initialization.
func (p *Pages) SetApplication(app *tview.Application) {
	p.app = app
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
		return // Cannot push while blocking modal is active
	}

	// Stop current component
	if len(p.stack) > 0 {
		current := p.stack[len(p.stack)-1]
		current.Stop()

		// If pushing a modal, save current focus for restoration
		if IsModal(c) && p.app != nil {
			p.focusStack = append(p.focusStack, p.app.GetFocus())
		}
	}

	// Generate unique page name
	p.counter++
	name := fmt.Sprintf("page-%d", p.counter)

	// Add to stack and pages
	p.stack = append(p.stack, c)
	p.Pages.AddPage(name, c, true, true)

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
		// Check OnDismiss callback
		if !modal.OnDismiss() {
			return false // Dismiss was cancelled
		}
		// Notify modal dismiss callback
		if p.onModalDismiss != nil {
			p.onModalDismiss(modal)
		}

		// Restore focus if configured
		behavior := modal.ModalBehavior()
		if behavior.RestoreFocusOnDismiss && len(p.focusStack) > 0 && p.app != nil {
			restoreTo := p.focusStack[len(p.focusStack)-1]
			p.focusStack = p.focusStack[:len(p.focusStack)-1]
			// Set focus directly - we're already on the UI thread via QueueUpdateDraw
			// from App's modal dismiss handler. Calling QueueUpdateDraw again here
			// can cause deadlocks.
			if restoreTo != nil {
				p.app.SetFocus(restoreTo)
			}
		}
	}

	// Stop and remove current
	current.Stop()

	// Get current page name and remove it
	name, _ := p.Pages.GetFrontPage()
	p.Pages.RemovePage(name)

	// Update stack
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

	// Remove all pages from the existing tview.Pages instead of replacing it.
	// Replacing with tview.NewPages() would orphan the primitive from the layout,
	// breaking focus management after profile switches.
	for _, name := range p.Pages.GetPageNames(true) {
		p.Pages.RemovePage(name)
	}

	p.notifyChange()
}

// StackDepth returns the number of components in stack.
func (p *Pages) StackDepth() int {
	return len(p.stack)
}

// SetOnChange sets callback when active component changes.
// The callback receives the new current component (may be nil).
func (p *Pages) SetOnChange(fn func(Component)) {
	p.onChange = fn
}

// SetCrumbs sets the breadcrumb component to update on navigation.
func (p *Pages) SetCrumbs(crumbs *Crumbs) {
	p.crumbs = crumbs
}

// notifyChange calls the onChange callback if set and updates crumbs.
func (p *Pages) notifyChange() {
	// Update crumbs from stack
	if p.crumbs != nil {
		var path []string
		for _, c := range p.stack {
			// Skip modals in breadcrumb path
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

// Draw renders the pages with current theme background.
func (p *Pages) Draw(screen tcell.Screen) {
	p.Pages.SetBackgroundColor(theme.Bg())
	p.Pages.Draw(screen)
}

// Replace replaces the current component without affecting stack depth.
// Useful for swapping views at the same level.
func (p *Pages) Replace(c Component) {
	if len(p.stack) == 0 {
		p.Push(c)
		return
	}

	// Check if a blocking modal is active
	if p.hasBlockingModal() {
		return
	}

	// Stop current
	current := p.stack[len(p.stack)-1]
	current.Stop()

	// Remove current page
	name, _ := p.Pages.GetFrontPage()
	p.Pages.RemovePage(name)

	// Add new component
	p.counter++
	newName := fmt.Sprintf("page-%d", p.counter)
	p.stack[len(p.stack)-1] = c
	p.Pages.AddPage(newName, c, true, true)

	// Start new component
	c.Start()

	// Notify listener
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
// Returns nil if the current page is not a modal.
func (p *Pages) CurrentModalBehavior() *components.ModalBehavior {
	if c := p.Current(); c != nil {
		return GetModalBehavior(c)
	}
	return nil
}

// DismissModal attempts to dismiss the current modal.
// Returns false if no modal is active, or if the modal's OnDismiss() cancelled it.
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
