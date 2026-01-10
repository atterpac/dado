package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// Checkbox is a boolean toggle component.
// It implements ValueProvider[bool].
type Checkbox struct {
	*tview.Box
	BaseEventEmitter

	name    string
	label   string
	checked bool
	focused bool

	// Typed handler
	onChange ChangeHandler[bool]
}

// NewCheckbox creates a new Checkbox.
func NewCheckbox(name string) *Checkbox {
	return &Checkbox{
		Box:  tview.NewBox(),
		name: name,
	}
}

// SetLabel sets the checkbox label.
func (c *Checkbox) SetLabel(label string) *Checkbox {
	c.label = label
	return c
}

// SetChecked sets the checked state.
func (c *Checkbox) SetChecked(checked bool) *Checkbox {
	c.checked = checked
	return c
}

// Checked returns the checked state.
func (c *Checkbox) Checked() bool {
	return c.checked
}

// Value returns the checked state (alias for Checked).
// This method is part of the ValueProvider interface.
func (c *Checkbox) Value() bool {
	return c.checked
}

// GetValue returns the checked state.
func (c *Checkbox) GetValue() bool {
	return c.checked
}

// HasValue returns true (checkbox always has a value).
func (c *Checkbox) HasValue() bool {
	return true
}

// Clear resets the checkbox to unchecked.
func (c *Checkbox) Clear() {
	c.SetChecked(false)
}

// GetName returns the field name.
func (c *Checkbox) GetName() string {
	return c.name
}

// SetOnChange sets the change handler (new API).
func (c *Checkbox) SetOnChange(handler ChangeHandler[bool]) *Checkbox {
	c.onChange = handler
	return c
}

// emitChange emits a change event to all handlers.
func (c *Checkbox) emitChange(oldValue, newValue bool) {
	event := NewChangeEvent(c.name, oldValue, newValue)

	// Typed handler
	if c.onChange != nil {
		c.onChange(event)
	}

	// Generic event bus
	c.EmitEvent(event)
}

// Toggle toggles the checked state.
func (c *Checkbox) Toggle() *Checkbox {
	oldValue := c.checked
	c.checked = !c.checked
	c.emitChange(oldValue, c.checked)
	return c
}

// Draw renders the checkbox.
func (c *Checkbox) Draw(screen tcell.Screen) {
	c.Box.DrawForSubclass(screen, c)
	x, y, width, height := c.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	accentColor := theme.Accent()
	successColor := theme.Success()

	style := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	if c.focused {
		style = style.Foreground(accentColor)
	}

	col := x

	// Draw checkbox
	checkStyle := style
	if c.checked {
		checkStyle = tcell.StyleDefault.Background(bgColor).Foreground(successColor)
	}

	screen.SetContent(col, y, '[', nil, checkStyle)
	col++
	if c.checked {
		screen.SetContent(col, y, '✓', nil, checkStyle)
	} else {
		screen.SetContent(col, y, ' ', nil, checkStyle)
	}
	col++
	screen.SetContent(col, y, ']', nil, checkStyle)
	col++

	// Draw label
	col++
	labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	for _, r := range c.label {
		if col < x+width {
			screen.SetContent(col, y, r, nil, labelStyle)
			col++
		}
	}
}

// InputHandler handles keyboard input.
func (c *Checkbox) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// Space to toggle (Enter is reserved for form submit)
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			c.Toggle()
		}
	})
}

// Focus handles focus.
func (c *Checkbox) Focus(delegate func(tview.Primitive)) {
	c.focused = true
	c.Box.Focus(delegate)
}

// Blur handles blur.
func (c *Checkbox) Blur() {
	c.focused = false
	c.Box.Blur()
}

// HasFocus returns whether the checkbox has focus.
func (c *Checkbox) HasFocus() bool {
	return c.focused
}

// GetFieldHeight returns the preferred height for this field.
func (c *Checkbox) GetFieldHeight() int {
	return 1
}

// RadioGroup is a single-choice option group.
// It implements IndexedValueProvider[string].
type RadioGroup struct {
	*tview.Box
	BaseEventEmitter

	name     string
	label    string
	options  []string
	selected int
	cursor   int
	focused  bool

	// Typed handler
	onChange ChangeHandler[string]
}

// NewRadioGroup creates a new RadioGroup.
func NewRadioGroup(name string) *RadioGroup {
	return &RadioGroup{
		Box:      tview.NewBox(),
		name:     name,
		selected: -1,
	}
}

// SetLabel sets the group label.
func (r *RadioGroup) SetLabel(label string) *RadioGroup {
	r.label = label
	return r
}

// SetOptions sets the available options.
func (r *RadioGroup) SetOptions(options []string) *RadioGroup {
	r.options = options
	return r
}

// SetSelected sets the selected index.
func (r *RadioGroup) SetSelected(index int) *RadioGroup {
	if index >= 0 && index < len(r.options) {
		r.selected = index
	}
	return r
}

// SetSelectedIndex sets the selected index, returning an error if out of range.
func (r *RadioGroup) SetSelectedIndex(index int) error {
	if index < -1 || index >= len(r.options) {
		return fmt.Errorf("index %d out of range [0, %d)", index, len(r.options))
	}
	r.selected = index
	return nil
}

// SetSelectedValue sets the selected option by value.
func (r *RadioGroup) SetSelectedValue(value string) error {
	for i, opt := range r.options {
		if opt == value {
			r.selected = i
			return nil
		}
	}
	return fmt.Errorf("value %q not found in options", value)
}

// SelectedIndex returns the selected index (-1 if none).
func (r *RadioGroup) SelectedIndex() int {
	return r.selected
}

// Value returns the selected value.
// This method is part of the ValueProvider interface.
func (r *RadioGroup) Value() string {
	if r.selected >= 0 && r.selected < len(r.options) {
		return r.options[r.selected]
	}
	return ""
}

// GetValue returns the selected value.
func (r *RadioGroup) GetValue() string {
	return r.Value()
}

// HasValue returns true if an option is selected.
func (r *RadioGroup) HasValue() bool {
	return r.selected >= 0 && r.selected < len(r.options)
}

// Clear resets the selection to none.
func (r *RadioGroup) Clear() {
	r.selected = -1
}

// GetName returns the field name.
func (r *RadioGroup) GetName() string {
	return r.name
}

// GetOptions returns the available options.
func (r *RadioGroup) GetOptions() []string {
	return r.options
}

// SetOnChange sets the change handler (new API).
func (r *RadioGroup) SetOnChange(handler ChangeHandler[string]) *RadioGroup {
	r.onChange = handler
	return r
}

// emitChange emits a change event to all handlers.
func (r *RadioGroup) emitChange(oldIndex, newIndex int) {
	var oldValue, newValue string
	if oldIndex >= 0 && oldIndex < len(r.options) {
		oldValue = r.options[oldIndex]
	}
	if newIndex >= 0 && newIndex < len(r.options) {
		newValue = r.options[newIndex]
	}

	event := NewChangeEvent(r.name, oldValue, newValue).WithIndex(newIndex)

	// Typed handler
	if r.onChange != nil {
		r.onChange(event)
	}

	// Generic event bus
	r.EmitEvent(event)
}

// Draw renders the radio group.
func (r *RadioGroup) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	accentColor := theme.Accent()
	successColor := theme.Success()

	row := y

	// Draw label if present
	if r.label != "" {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := x
		for _, ch := range r.label {
			if col < x+width {
				screen.SetContent(col, row, ch, nil, labelStyle)
				col++
			}
		}
		row++
	}

	// Draw options
	for i, opt := range r.options {
		if row >= y+height {
			break
		}

		col := x

		// Determine style
		var rowBg, rowFg tcell.Color
		if r.focused && i == r.cursor {
			rowBg = accentColor
			rowFg = bgColor
		} else {
			rowBg = bgColor
			rowFg = fgColor
		}
		rowStyle := tcell.StyleDefault.Background(rowBg).Foreground(rowFg)

		// Clear row
		clearStyle := tcell.StyleDefault.Background(rowBg)
		for c := x; c < x+width; c++ {
			screen.SetContent(c, row, ' ', nil, clearStyle)
		}

		// Draw radio button
		radioStyle := rowStyle
		if i == r.selected {
			radioStyle = tcell.StyleDefault.Background(rowBg).Foreground(successColor)
			if r.focused && i == r.cursor {
				radioStyle = rowStyle
			}
		}

		screen.SetContent(col, row, '(', nil, radioStyle)
		col++
		if i == r.selected {
			screen.SetContent(col, row, '●', nil, radioStyle)
		} else {
			screen.SetContent(col, row, ' ', nil, radioStyle)
		}
		col++
		screen.SetContent(col, row, ')', nil, radioStyle)
		col++

		// Draw option label
		col++
		for _, ch := range opt {
			if col < x+width {
				screen.SetContent(col, row, ch, nil, rowStyle)
				col++
			}
		}
		row++
	}
}

// InputHandler handles keyboard input.
func (r *RadioGroup) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			if r.cursor > 0 {
				r.cursor--
			}
		case tcell.KeyDown:
			if r.cursor < len(r.options)-1 {
				r.cursor++
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ':
				// Space to select (Enter is reserved for form submit)
				oldSelected := r.selected
				r.selected = r.cursor
				r.emitChange(oldSelected, r.selected)
			case 'j':
				if r.cursor < len(r.options)-1 {
					r.cursor++
				}
			case 'k':
				if r.cursor > 0 {
					r.cursor--
				}
			}
		}
	})
}

// Focus handles focus.
func (r *RadioGroup) Focus(delegate func(tview.Primitive)) {
	r.focused = true
	r.Box.Focus(delegate)
}

// Blur handles blur.
func (r *RadioGroup) Blur() {
	r.focused = false
	r.Box.Blur()
}

// HasFocus returns whether the radio group has focus.
func (r *RadioGroup) HasFocus() bool {
	return r.focused
}

// GetFieldHeight returns the preferred height for this field.
func (r *RadioGroup) GetFieldHeight() int {
	height := len(r.options)
	if r.label != "" {
		height++
	}
	return height
}
