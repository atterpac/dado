package components

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testScreen wraps tcell.SimulationScreen for testing render output.
type testScreen struct {
	tcell.SimulationScreen
}

// newTestScreen creates a new test screen with the given dimensions.
func newTestScreen(width, height int) *testScreen {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		panic(err)
	}
	screen.SetSize(width, height)
	screen.Clear()
	return &testScreen{SimulationScreen: screen}
}

// getContent reads content at position (x, y) for length characters.
func (ts *testScreen) getContent(x, y, length int) string {
	var result strings.Builder
	for i := 0; i < length; i++ {
		mainc, _, _, _ := ts.SimulationScreen.GetContent(x+i, y)
		if mainc == 0 {
			mainc = ' '
		}
		result.WriteRune(mainc)
	}
	return result.String()
}

// dump returns all screen content as a single string.
func (ts *testScreen) dump() string {
	_, h := ts.Size()
	var result strings.Builder
	for row := 0; row < h; row++ {
		line := ts.getContent(0, row, 200)
		result.WriteString(strings.TrimRight(line, " "))
		result.WriteString("\n")
	}
	return result.String()
}

// containsText checks if the screen contains the given text.
func (ts *testScreen) containsText(text string) bool {
	return strings.Contains(ts.dump(), text)
}

// drawPrimitive draws a tview.Primitive to the test screen.
func (ts *testScreen) drawPrimitive(p tview.Primitive) {
	w, h := ts.Size()
	p.SetRect(0, 0, w, h)
	p.Draw(ts.SimulationScreen)
	ts.Show()
}

// TestPanel_Render tests Panel rendering.
func TestPanel_Render(t *testing.T) {
	screen := newTestScreen(40, 10)

	panel := NewPanel().SetTitle("Test Panel")
	panel.SetBorder(true)

	screen.drawPrimitive(panel)

	// Panel should render with border and title
	content := screen.dump()
	assert.Contains(t, content, "Test Panel")
}

// TestTextField_Render tests TextField rendering.
func TestTextField_Render(t *testing.T) {
	screen := newTestScreen(40, 3)

	field := NewTextField("test").
		SetValue("hello world").
		SetPlaceholder("Enter text")

	screen.drawPrimitive(field)

	content := screen.dump()
	assert.Contains(t, content, "hello world")
}

// TestTextField_RenderPlaceholder tests TextField placeholder rendering.
func TestTextField_RenderPlaceholder(t *testing.T) {
	screen := newTestScreen(40, 3)

	field := NewTextField("test").
		SetPlaceholder("Enter text here")

	screen.drawPrimitive(field)

	content := screen.dump()
	assert.Contains(t, content, "Enter text here")
}

// TestCheckbox_RenderChecked tests Checkbox rendering when checked.
func TestCheckbox_RenderChecked(t *testing.T) {
	screen := newTestScreen(40, 3)

	cb := NewCheckbox("test").
		SetLabel("Enable feature").
		SetChecked(true)

	screen.drawPrimitive(cb)

	content := screen.dump()
	assert.Contains(t, content, "Enable feature")
	// Checked checkbox should show a check mark
	assert.True(t, screen.containsText("x") || screen.containsText("✓") || screen.containsText("X"))
}

// TestCheckbox_RenderUnchecked tests Checkbox rendering when unchecked.
func TestCheckbox_RenderUnchecked(t *testing.T) {
	screen := newTestScreen(40, 3)

	cb := NewCheckbox("test").
		SetLabel("Enable feature").
		SetChecked(false)

	screen.drawPrimitive(cb)

	content := screen.dump()
	assert.Contains(t, content, "Enable feature")
}

// TestSelect_RenderClosed tests Select rendering in closed state.
func TestSelect_RenderClosed(t *testing.T) {
	screen := newTestScreen(40, 5)

	sel := NewSelect("test").
		SetOptions([]string{"Option A", "Option B", "Option C"}).
		SetSelected(0)

	screen.drawPrimitive(sel)

	content := screen.dump()
	assert.Contains(t, content, "Option A")
}

// TestRadioGroup_Render tests RadioGroup rendering.
func TestRadioGroup_Render(t *testing.T) {
	screen := newTestScreen(40, 10)

	rg := NewRadioGroup("test").
		SetOptions([]string{"Red", "Green", "Blue"}).
		SetSelected(1)

	screen.drawPrimitive(rg)

	content := screen.dump()
	assert.Contains(t, content, "Red")
	assert.Contains(t, content, "Green")
	assert.Contains(t, content, "Blue")
}

// TestMultiSelect_Render tests MultiSelect rendering.
func TestMultiSelect_Render(t *testing.T) {
	screen := newTestScreen(40, 10)

	ms := NewMultiSelect("test").
		SetOptions([]string{"Tag1", "Tag2", "Tag3"}).
		SetSelected([]int{0, 2})

	screen.drawPrimitive(ms)

	content := screen.dump()
	assert.Contains(t, content, "Tag1")
	assert.Contains(t, content, "Tag2")
	assert.Contains(t, content, "Tag3")
}

// TestTextArea_Render tests TextArea rendering.
func TestTextArea_Render(t *testing.T) {
	screen := newTestScreen(40, 10)

	ta := NewTextArea("test").
		SetValue("Line 1\nLine 2\nLine 3")

	screen.drawPrimitive(ta)

	content := screen.dump()
	assert.Contains(t, content, "Line 1")
	assert.Contains(t, content, "Line 2")
	assert.Contains(t, content, "Line 3")
}

// TestForm_Render tests Form rendering with multiple fields.
func TestForm_Render(t *testing.T) {
	screen := newTestScreen(60, 20)

	form := NewFormBuilder().
		Text("name", "Name").
		Value("John Doe").
		Done().
		Text("email", "Email").
		Value("john@example.com").
		Done().
		Checkbox("subscribe", "Subscribe").
		Checked(true).
		Done().
		Build()

	screen.drawPrimitive(form)

	content := screen.dump()
	// Form should render field labels and values
	assert.Contains(t, content, "Name")
	assert.Contains(t, content, "Email")
	assert.Contains(t, content, "Subscribe")
}

// TestModal_Render tests Modal rendering.
func TestModal_Render(t *testing.T) {
	screen := newTestScreen(80, 24)

	modal := NewModal(ModalConfig{
		Title:  "Confirm Action",
		Width:  40,
		Height: 10,
	})

	textContent := tview.NewTextView()
	textContent.SetText("Are you sure?")
	modal.SetContent(textContent)

	screen.drawPrimitive(modal)

	dump := screen.dump()
	assert.Contains(t, dump, "Confirm Action")
	assert.Contains(t, dump, "Are you sure?")
}

// TestScreen_GetContent tests reading specific positions.
func TestScreen_GetContent(t *testing.T) {
	screen := newTestScreen(20, 5)

	// Draw a simple text view at known position
	tv := tview.NewTextView()
	tv.SetText("ABCDE")
	tv.SetRect(0, 0, 20, 5)

	screen.drawPrimitive(tv)

	// Read content at the start
	content := screen.getContent(0, 0, 5)
	assert.Equal(t, "ABCDE", content)
}

// TestScreen_ContainsText tests text search.
func TestScreen_ContainsText(t *testing.T) {
	screen := newTestScreen(40, 10)

	tv := tview.NewTextView()
	tv.SetText("Hello World")
	tv.SetRect(0, 0, 40, 10)

	screen.drawPrimitive(tv)

	assert.True(t, screen.containsText("Hello"))
	assert.True(t, screen.containsText("World"))
	assert.False(t, screen.containsText("Goodbye"))
}

// TestTextField_RenderWithValidationError tests error display.
func TestTextField_RenderWithValidationError(t *testing.T) {
	screen := newTestScreen(40, 5)

	field := NewTextField("test").
		SetValue("").
		SetValidator(func(v string) error {
			if v == "" {
				return assert.AnError
			}
			return nil
		})

	// Trigger validation
	field.Validate()

	screen.drawPrimitive(field)

	// Field should indicate error state
	assert.True(t, field.HasError())
}

// TestComponent_FocusRender tests rendering with focus state.
func TestComponent_FocusRender(t *testing.T) {
	screen := newTestScreen(40, 5)

	field := NewTextField("test").SetValue("focused field")

	// Focus the field
	field.Focus(func(p tview.Primitive) {})

	screen.drawPrimitive(field)

	content := screen.dump()
	assert.Contains(t, content, "focused field")
}

// TestPanel_RenderWithContent tests Panel with content.
func TestPanel_RenderWithContent(t *testing.T) {
	screen := newTestScreen(50, 15)

	panel := NewPanel().SetTitle("Content Panel")
	panel.SetBorder(true)

	textContent := tview.NewTextView()
	textContent.SetText("Panel body text goes here")
	panel.SetContent(textContent)

	screen.drawPrimitive(panel)

	dump := screen.dump()
	assert.Contains(t, dump, "Content Panel")
	assert.Contains(t, dump, "Panel body text goes here")
}

// TestList_Render tests List component rendering.
func TestList_Render(t *testing.T) {
	screen := newTestScreen(40, 15)

	list := NewList()
	list.SetItems([]ListItem{
		{Text: "Item 1", Secondary: "Description 1"},
		{Text: "Item 2", Secondary: "Description 2"},
		{Text: "Item 3", Secondary: "Description 3"},
	})

	screen.drawPrimitive(list)

	dump := screen.dump()
	assert.Contains(t, dump, "Item 1")
	assert.Contains(t, dump, "Item 2")
	assert.Contains(t, dump, "Item 3")
}

// TestScreen_Clear tests screen clearing.
func TestScreen_Clear(t *testing.T) {
	screen := newTestScreen(20, 5)

	// Draw something
	tv := tview.NewTextView()
	tv.SetText("Content")
	tv.SetRect(0, 0, 20, 5)
	screen.drawPrimitive(tv)

	require.True(t, screen.containsText("Content"))

	// Clear the screen
	screen.Clear()
	screen.Show()

	// Content should be gone
	assert.False(t, screen.containsText("Content"))
}

// TestScreen_Size tests screen dimensions.
func TestScreen_Size(t *testing.T) {
	screen := newTestScreen(80, 24)

	w, h := screen.Size()
	assert.Equal(t, 80, w)
	assert.Equal(t, 24, h)
}

// TestScreen_SimulateResize tests screen resize simulation.
func TestScreen_SimulateResize(t *testing.T) {
	screen := newTestScreen(80, 24)

	// Resize the screen
	screen.SetSize(100, 30)

	w, h := screen.Size()
	assert.Equal(t, 100, w)
	assert.Equal(t, 30, h)
}
