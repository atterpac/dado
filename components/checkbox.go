package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// Checkbox is a boolean toggle component.
type Checkbox struct {
	*tview.Box

	name    string
	label   string
	checked bool
	focused bool

	onChange func(checked bool)
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

// IsChecked returns the checked state.
func (c *Checkbox) IsChecked() bool {
	return c.checked
}

// GetValue returns the checked state as interface{}.
func (c *Checkbox) GetValue() bool {
	return c.checked
}

// GetName returns the field name.
func (c *Checkbox) GetName() string {
	return c.name
}

// SetOnChange sets the callback for state changes.
func (c *Checkbox) SetOnChange(fn func(checked bool)) *Checkbox {
	c.onChange = fn
	return c
}

// Toggle toggles the checked state.
func (c *Checkbox) Toggle() *Checkbox {
	c.checked = !c.checked
	if c.onChange != nil {
		c.onChange(c.checked)
	}
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
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyRune:
			if event.Key() == tcell.KeyEnter || event.Rune() == ' ' {
				c.Toggle()
			}
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
type RadioGroup struct {
	*tview.Box

	name     string
	label    string
	options  []string
	selected int
	cursor   int
	focused  bool

	onChange func(index int, value string)
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

// GetSelected returns the selected index and value.
func (r *RadioGroup) GetSelected() (int, string) {
	if r.selected >= 0 && r.selected < len(r.options) {
		return r.selected, r.options[r.selected]
	}
	return -1, ""
}

// GetValue returns the selected value.
func (r *RadioGroup) GetValue() string {
	if r.selected >= 0 && r.selected < len(r.options) {
		return r.options[r.selected]
	}
	return ""
}

// GetName returns the field name.
func (r *RadioGroup) GetName() string {
	return r.name
}

// GetOptions returns the available options.
func (r *RadioGroup) GetOptions() []string {
	return r.options
}

// SetOnChange sets the callback for selection changes.
func (r *RadioGroup) SetOnChange(fn func(index int, value string)) *RadioGroup {
	r.onChange = fn
	return r
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
		case tcell.KeyEnter, tcell.KeyRune:
			if event.Key() == tcell.KeyEnter || event.Rune() == ' ' {
				r.selected = r.cursor
				if r.onChange != nil {
					r.onChange(r.selected, r.options[r.selected])
				}
			}
			// Vim keys
			if event.Key() == tcell.KeyRune {
				switch event.Rune() {
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
