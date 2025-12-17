package nav

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
)

// Component represents a navigable view/page.
// All views pushed to Pages must implement this interface.
type Component interface {
	tview.Primitive

	// Start is called when the component becomes active (shown).
	Start()

	// Stop is called when the component becomes inactive (hidden).
	Stop()

	// Hints returns key binding hints for this component.
	Hints() []components.KeyHint
}

// Pages manages stack-based page navigation.
type Pages struct {
	*tview.Pages
	stack    []Component
	onChange func(Component)
	counter  int // For generating unique page names
}

// NewPages creates a new page stack manager.
func NewPages() *Pages {
	pages := tview.NewPages()
	pages.SetBackgroundColor(theme.Bg())

	p := &Pages{
		Pages: pages,
		stack: make([]Component, 0),
	}

	// Register for automatic theme updates
	theme.Register(pages)

	return p
}

// Push adds a component to the stack and shows it.
// Calls Stop() on the previous component if any.
func (p *Pages) Push(c Component) {
	// Stop current component
	if len(p.stack) > 0 {
		p.stack[len(p.stack)-1].Stop()
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
}

// Pop removes the current component and returns to previous.
// Calls Stop() on current, Start() on previous.
// Returns false if stack is empty or only has one item.
func (p *Pages) Pop() bool {
	if len(p.stack) <= 1 {
		return false
	}

	// Stop and remove current
	current := p.stack[len(p.stack)-1]
	current.Stop()

	// Get current page name and remove it
	name, _ := p.Pages.GetFrontPage()
	p.Pages.RemovePage(name)

	// Update stack
	p.stack = p.stack[:len(p.stack)-1]

	// Show and start previous
	if len(p.stack) > 0 {
		prev := p.stack[len(p.stack)-1]
		prev.Start()
	}

	// Notify listener
	p.notifyChange()

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
	p.Pages = tview.NewPages()
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

// notifyChange calls the onChange callback if set.
func (p *Pages) notifyChange() {
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
}
