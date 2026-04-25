package nav

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atterpac/jig/components"
)

// mockComponent implements nav.Component for testing.
type mockComponent struct {
	tview.Primitive
	name         string
	startCalled  bool
	stopCalled   bool
	startCount   int
	stopCount    int
	hints        []components.KeyHint
	onStart      func()
	onStop       func()
}

func newMockComponent(name string) *mockComponent {
	return &mockComponent{
		Primitive: tview.NewBox(),
		name:      name,
		hints:     make([]components.KeyHint, 0),
	}
}

func (m *mockComponent) Start() {
	m.startCalled = true
	m.startCount++
	if m.onStart != nil {
		m.onStart()
	}
}

func (m *mockComponent) Stop() {
	m.stopCalled = true
	m.stopCount++
	if m.onStop != nil {
		m.onStop()
	}
}

func (m *mockComponent) Name() string { return m.name }

func (m *mockComponent) Hints() []components.KeyHint {
	return m.hints
}

func (m *mockComponent) reset() {
	m.startCalled = false
	m.stopCalled = false
}

// mockModal implements nav.Modal for testing.
type mockModal struct {
	*mockComponent
	behavior        components.ModalBehavior
	onDismissReturn bool
	dismissCalled   bool
}

func newMockModal(name string) *mockModal {
	return &mockModal{
		mockComponent: newMockComponent(name),
		behavior: components.ModalBehavior{
			CapturesAllInput:      true,
			DismissOnEsc:          true,
			RestoreFocusOnDismiss: true,
		},
		onDismissReturn: true,
	}
}

func (m *mockModal) ModalBehavior() components.ModalBehavior {
	return m.behavior
}

func (m *mockModal) OnDismiss() bool {
	m.dismissCalled = true
	return m.onDismissReturn
}

// TestPages_NewPages tests Pages creation.
func TestPages_NewPages(t *testing.T) {
	pages := NewPages()

	assert.NotNil(t, pages)
	assert.Equal(t, 0, pages.StackDepth())
	assert.Nil(t, pages.Current())
}

// TestPages_Push tests pushing components onto the stack.
func TestPages_Push(t *testing.T) {
	pages := NewPages()
	c1 := newMockComponent("view1")

	pages.Push(c1)

	assert.Equal(t, 1, pages.StackDepth())
	assert.Same(t, c1, pages.Current())
	assert.True(t, c1.startCalled)
	assert.False(t, c1.stopCalled)
}

// TestPages_PushMultiple tests pushing multiple components.
func TestPages_PushMultiple(t *testing.T) {
	pages := NewPages()
	c1 := newMockComponent("view1")
	c2 := newMockComponent("view2")

	pages.Push(c1)
	assert.Equal(t, 1, pages.StackDepth())
	assert.True(t, c1.startCalled)

	pages.Push(c2)
	assert.Equal(t, 2, pages.StackDepth())
	assert.True(t, c1.stopCalled, "previous should be stopped")
	assert.True(t, c2.startCalled)
	assert.Same(t, c2, pages.Current())
}

// TestPages_Pop tests popping components from the stack.
func TestPages_Pop(t *testing.T) {
	pages := NewPages()
	c1 := newMockComponent("view1")
	c2 := newMockComponent("view2")

	pages.Push(c1)
	pages.Push(c2)

	c1.reset()

	result := pages.Pop()

	assert.True(t, result)
	assert.True(t, c2.stopCalled)
	assert.True(t, c1.startCalled, "previous should be re-started")
	assert.Equal(t, 1, pages.StackDepth())
	assert.Same(t, c1, pages.Current())
}

// TestPages_PopEmpty tests popping from empty stack.
func TestPages_PopEmpty(t *testing.T) {
	pages := NewPages()

	result := pages.Pop()

	assert.False(t, result)
}

// TestPages_PopSingleItem tests that we can't pop the last item.
func TestPages_PopSingleItem(t *testing.T) {
	pages := NewPages()
	c := newMockComponent("only")
	pages.Push(c)

	result := pages.Pop()

	assert.False(t, result, "should not pop last item")
	assert.Equal(t, 1, pages.StackDepth())
}

// TestPages_Clear tests clearing all components.
func TestPages_Clear(t *testing.T) {
	pages := NewPages()
	c1 := newMockComponent("view1")
	c2 := newMockComponent("view2")
	c3 := newMockComponent("view3")

	pages.Push(c1)
	pages.Push(c2)
	pages.Push(c3)

	pages.Clear()

	assert.Equal(t, 0, pages.StackDepth())
	assert.Nil(t, pages.Current())
	assert.True(t, c1.stopCalled)
	assert.True(t, c2.stopCalled)
	assert.True(t, c3.stopCalled)
}

// TestPages_Replace tests replacing the current component.
func TestPages_Replace(t *testing.T) {
	pages := NewPages()
	c1 := newMockComponent("view1")
	c2 := newMockComponent("view2")
	c3 := newMockComponent("replacement")

	pages.Push(c1)
	pages.Push(c2)
	depth := pages.StackDepth()

	pages.Replace(c3)

	assert.Equal(t, depth, pages.StackDepth(), "depth should not change")
	assert.True(t, c2.stopCalled)
	assert.True(t, c3.startCalled)
	assert.Same(t, c3, pages.Current())
}

// TestPages_ReplaceEmpty tests replace on empty stack.
func TestPages_ReplaceEmpty(t *testing.T) {
	pages := NewPages()
	c := newMockComponent("view")

	pages.Replace(c)

	assert.Equal(t, 1, pages.StackDepth())
	assert.Same(t, c, pages.Current())
	assert.True(t, c.startCalled)
}

// TestPages_OnChange tests the onChange callback.
func TestPages_OnChange(t *testing.T) {
	pages := NewPages()

	var changes []Component
	pages.SetOnChange(func(c Component) {
		changes = append(changes, c)
	})

	c1 := newMockComponent("view1")
	c2 := newMockComponent("view2")

	pages.Push(c1)
	pages.Push(c2)
	pages.Pop()

	require.Len(t, changes, 3)
	assert.Same(t, c1, changes[0])
	assert.Same(t, c2, changes[1])
	assert.Same(t, c1, changes[2])
}

// TestPages_GetStack tests getting a copy of the stack.
func TestPages_GetStack(t *testing.T) {
	pages := NewPages()
	c1 := newMockComponent("view1")
	c2 := newMockComponent("view2")

	pages.Push(c1)
	pages.Push(c2)

	stack := pages.GetStack()
	require.Len(t, stack, 2)
	assert.Same(t, c1, stack[0])
	assert.Same(t, c2, stack[1])

	// Verify it's a copy - modifying returned stack shouldn't affect internal
	stack[0] = nil
	assert.NotNil(t, pages.GetStack()[0])
}

// TestPages_CanPop tests the CanPop helper.
func TestPages_CanPop(t *testing.T) {
	pages := NewPages()

	assert.False(t, pages.CanPop(), "empty stack")

	pages.Push(newMockComponent("view1"))
	assert.False(t, pages.CanPop(), "single item")

	pages.Push(newMockComponent("view2"))
	assert.True(t, pages.CanPop(), "multiple items")
}

// TestPages_Modal tests modal push/pop behavior.
func TestPages_Modal(t *testing.T) {
	pages := NewPages()
	view := newMockComponent("main")
	modal := newMockModal("confirm")

	pages.Push(view)
	assert.False(t, pages.HasModal())
	assert.False(t, pages.CurrentIsModal())
	assert.Equal(t, 0, pages.ModalCount())

	pages.Push(modal)
	assert.True(t, pages.HasModal())
	assert.True(t, pages.CurrentIsModal())
	assert.Equal(t, 1, pages.ModalCount())

	// Pop modal
	result := pages.Pop()
	assert.True(t, result)
	assert.False(t, pages.HasModal())
	assert.True(t, modal.dismissCalled)
}

// TestPages_ModalOnDismiss tests that modal OnDismiss is called.
func TestPages_ModalOnDismiss(t *testing.T) {
	pages := NewPages()
	view := newMockComponent("main")
	modal := newMockModal("confirm")

	pages.Push(view)
	pages.Push(modal)

	pages.Pop()

	assert.True(t, modal.dismissCalled)
}

// TestPages_ModalDismissCancelled tests cancelling modal dismiss.
func TestPages_ModalDismissCancelled(t *testing.T) {
	pages := NewPages()
	view := newMockComponent("main")
	modal := newMockModal("confirm")
	modal.onDismissReturn = false // Cancel dismiss

	pages.Push(view)
	pages.Push(modal)

	result := pages.Pop()

	assert.False(t, result, "pop should be cancelled")
	assert.Equal(t, 2, pages.StackDepth())
	assert.True(t, modal.dismissCalled)
}

// TestPages_DismissModal tests the DismissModal helper.
func TestPages_DismissModal(t *testing.T) {
	pages := NewPages()
	view := newMockComponent("main")
	modal := newMockModal("confirm")

	pages.Push(view)
	pages.Push(modal)

	result := pages.DismissModal()

	assert.True(t, result)
	assert.Equal(t, 1, pages.StackDepth())
}

// TestPages_DismissModalNotModal tests DismissModal when current is not a modal.
func TestPages_DismissModalNotModal(t *testing.T) {
	pages := NewPages()
	view := newMockComponent("main")

	pages.Push(view)

	result := pages.DismissModal()

	assert.False(t, result)
	assert.Equal(t, 1, pages.StackDepth())
}

// TestPages_BlockingModal tests that blocking modal prevents push.
func TestPages_BlockingModal(t *testing.T) {
	pages := NewPages()
	view := newMockComponent("main")
	modal := newMockModal("blocking")
	modal.behavior.BlockUntilDismissed = true

	pages.Push(view)
	pages.Push(modal)

	// Try to push another component
	another := newMockComponent("another")
	pages.Push(another)

	assert.Equal(t, 2, pages.StackDepth(), "should not push while blocking modal active")
	assert.False(t, another.startCalled)
}

// TestPages_BlockingModalReplace tests that blocking modal prevents replace.
func TestPages_BlockingModalReplace(t *testing.T) {
	pages := NewPages()
	view := newMockComponent("main")
	modal := newMockModal("blocking")
	modal.behavior.BlockUntilDismissed = true

	pages.Push(view)
	pages.Push(modal)

	// Try to replace
	replacement := newMockComponent("replacement")
	pages.Replace(replacement)

	// Should still have the modal
	assert.Same(t, modal.mockComponent, pages.Current().(*mockModal).mockComponent)
	assert.False(t, replacement.startCalled)
}

// TestPages_NestedModals tests multiple modals in stack.
func TestPages_NestedModals(t *testing.T) {
	pages := NewPages()
	view := newMockComponent("main")
	modal1 := newMockModal("modal1")
	modal2 := newMockModal("modal2")

	pages.Push(view)
	pages.Push(modal1)
	pages.Push(modal2)

	assert.Equal(t, 2, pages.ModalCount())

	pages.Pop()
	assert.Equal(t, 1, pages.ModalCount())

	pages.Pop()
	assert.Equal(t, 0, pages.ModalCount())
}

// TestPages_OnModalDismiss tests the onModalDismiss callback.
func TestPages_OnModalDismiss(t *testing.T) {
	pages := NewPages()

	var dismissed Modal
	pages.SetOnModalDismiss(func(m Modal) {
		dismissed = m
	})

	view := newMockComponent("main")
	modal := newMockModal("confirm")

	pages.Push(view)
	pages.Push(modal)
	pages.Pop()

	require.NotNil(t, dismissed)
	assert.Same(t, modal, dismissed)
}

// TestPages_CurrentModalBehavior tests getting current modal behavior.
func TestPages_CurrentModalBehavior(t *testing.T) {
	pages := NewPages()

	// No pages
	assert.Nil(t, pages.CurrentModalBehavior())

	// Non-modal page
	view := newMockComponent("main")
	pages.Push(view)
	assert.Nil(t, pages.CurrentModalBehavior())

	// Modal page
	modal := newMockModal("confirm")
	modal.behavior.CapturesAllInput = true
	modal.behavior.DismissOnEsc = false
	pages.Push(modal)

	behavior := pages.CurrentModalBehavior()
	require.NotNil(t, behavior)
	assert.True(t, behavior.CapturesAllInput)
	assert.False(t, behavior.DismissOnEsc)
}

// TestIsModal tests the IsModal helper function.
func TestIsModal(t *testing.T) {
	component := newMockComponent("view")
	modal := newMockModal("modal")

	assert.False(t, IsModal(component))
	assert.True(t, IsModal(modal))
}

// TestAsModal tests the AsModal helper function.
func TestAsModal(t *testing.T) {
	component := newMockComponent("view")
	modal := newMockModal("modal")

	assert.Nil(t, AsModal(component))
	assert.Same(t, modal, AsModal(modal))
}

// TestGetModalBehavior tests the GetModalBehavior helper function.
func TestGetModalBehavior(t *testing.T) {
	component := newMockComponent("view")
	modal := newMockModal("modal")

	assert.Nil(t, GetModalBehavior(component))
	assert.NotNil(t, GetModalBehavior(modal))
}

// TestPages_LifecycleOrdering tests lifecycle call ordering during navigation.
func TestPages_LifecycleOrdering(t *testing.T) {
	pages := NewPages()

	var order []string

	c1 := newMockComponent("view1")
	c2 := newMockComponent("view2")

	// Set callbacks to record order
	c1.onStart = func() {
		order = append(order, "c1-start")
	}
	c1.onStop = func() {
		order = append(order, "c1-stop")
	}
	c2.onStart = func() {
		order = append(order, "c2-start")
	}
	c2.onStop = func() {
		order = append(order, "c2-stop")
	}

	// Navigate: push c1 -> push c2 -> pop
	pages.Push(c1)
	pages.Push(c2)
	pages.Pop()

	expected := []string{
		"c1-start",  // Initial push
		"c1-stop",   // c1 stopped when c2 pushed
		"c2-start",  // c2 starts
		"c2-stop",   // c2 stopped when popped
		"c1-start",  // c1 re-started
	}
	assert.Equal(t, expected, order)
}

// TestPages_StackIntegrity tests stack remains consistent through various operations.
func TestPages_StackIntegrity(t *testing.T) {
	pages := NewPages()

	// Create components
	components := make([]*mockComponent, 5)
	for i := range components {
		components[i] = newMockComponent("view")
	}

	// Push all
	for _, c := range components {
		pages.Push(c)
	}
	assert.Equal(t, 5, pages.StackDepth())

	// Pop some
	pages.Pop()
	pages.Pop()
	assert.Equal(t, 3, pages.StackDepth())

	// Replace
	replacement := newMockComponent("replacement")
	pages.Replace(replacement)
	assert.Equal(t, 3, pages.StackDepth())

	// Verify stack integrity
	stack := pages.GetStack()
	assert.Same(t, components[0], stack[0])
	assert.Same(t, components[1], stack[1])
	assert.Same(t, replacement, stack[2])
}
