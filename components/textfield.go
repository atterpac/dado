package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
	"github.com/atterpac/jig/validators"
)

// TextField is a single-line text input with validation.
// It implements ValueProvider[string].
type TextField struct {
	*tview.Box
	BaseEventEmitter

	name        string
	label       string
	placeholder string
	value       string
	cursorPos   int
	offset      int // horizontal scroll offset

	// Validation
	validator func(string) error
	error     string

	// State
	focused bool

	// Typed handlers
	onChange ChangeHandler[string]
	onSubmit SubmitHandler
}

// NewTextField creates a new TextField.
func NewTextField(name string) *TextField {
	return &TextField{
		Box:  tview.NewBox(),
		name: name,
	}
}

// SetLabel sets the field label.
func (t *TextField) SetLabel(label string) *TextField {
	t.label = label
	return t
}

// SetPlaceholder sets the placeholder text.
func (t *TextField) SetPlaceholder(placeholder string) *TextField {
	t.placeholder = placeholder
	return t
}

// SetValue sets the current value.
func (t *TextField) SetValue(value string) *TextField {
	t.value = value
	t.cursorPos = len(value)
	t.validate()
	return t
}

// GetValue returns the current value.
func (t *TextField) GetValue() string {
	return t.value
}

// Value returns the current value (alias for GetValue).
// This method is part of the ValueProvider interface.
func (t *TextField) Value() string {
	return t.value
}

// HasValue returns true if the text field has a non-empty value.
func (t *TextField) HasValue() bool {
	return t.value != ""
}

// Clear resets the text field to an empty value.
func (t *TextField) Clear() {
	t.SetValue("")
}

// GetName returns the field name.
func (t *TextField) GetName() string {
	return t.name
}

// SetValidator sets the validation function.
func (t *TextField) SetValidator(fn func(string) error) *TextField {
	t.validator = fn
	return t
}

// SetValidators sets multiple validators for the field.
// Validators are run in order and the first error is returned.
func (t *TextField) SetValidators(vs ...validators.Validator) *TextField {
	t.validator = func(value string) error {
		for _, v := range vs {
			if err := v(value); err != nil {
				return err
			}
		}
		return nil
	}
	return t
}

// SetOnChange sets the change handler (new API).
func (t *TextField) SetOnChange(handler ChangeHandler[string]) *TextField {
	t.onChange = handler
	return t
}

// SetOnSubmit sets the submit handler (new API).
func (t *TextField) SetOnSubmit(handler SubmitHandler) *TextField {
	t.onSubmit = handler
	return t
}


// emitChange emits a change event to all handlers.
func (t *TextField) emitChange(oldValue, newValue string) {
	event := NewChangeEvent(t.name, oldValue, newValue)

	// Typed handler
	if t.onChange != nil {
		t.onChange(event)
	}

	// Generic event bus
	t.EmitEvent(event)
}

// emitSubmit emits a submit event to all handlers.
func (t *TextField) emitSubmit() {
	event := NewSubmitEvent(t.name, t.value)

	// Typed handler
	if t.onSubmit != nil {
		t.onSubmit(event)
	}

	// Generic event bus
	t.EmitEvent(event)
}

// Validate runs validation and returns any error.
func (t *TextField) Validate() error {
	return t.validate()
}

func (t *TextField) validate() error {
	if t.validator != nil {
		err := t.validator(t.value)
		if err != nil {
			t.error = err.Error()
		} else {
			t.error = ""
		}
		return err
	}
	t.error = ""
	return nil
}

// HasError returns whether the field has a validation error.
func (t *TextField) HasError() bool {
	return t.error != ""
}

// GetError returns the current validation error.
func (t *TextField) GetError() string {
	return t.error
}

// Draw renders the text field.
func (t *TextField) Draw(screen tcell.Screen) {
	t.Box.DrawForSubclass(screen, t)
	x, y, width, height := t.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	errorColor := theme.Error()
	borderColor := theme.Border()
	borderFocusColor := theme.BorderFocus()

	row := y

	// Draw label if present
	if t.label != "" && height > 1 {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := x
		for _, r := range t.label {
			if col < x+width {
				screen.SetContent(col, row, r, nil, labelStyle)
				col++
			}
		}
		row++
	}

	// Determine border color
	var currentBorderColor tcell.Color
	if t.error != "" {
		currentBorderColor = errorColor
	} else if t.focused {
		currentBorderColor = borderFocusColor
	} else {
		currentBorderColor = borderColor
	}

	// Draw input box border
	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(currentBorderColor)
	inputStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	placeholderStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	// Draw top border
	screen.SetContent(x, row, '╭', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╮', nil, borderStyle)
	row++

	// Draw input line
	screen.SetContent(x, row, '│', nil, borderStyle)
	screen.SetContent(x+width-1, row, '│', nil, borderStyle)

	// Calculate visible text area
	inputWidth := width - 4 // 2 for borders, 2 for padding
	inputX := x + 2

	// Adjust offset for scrolling
	if t.cursorPos < t.offset {
		t.offset = t.cursorPos
	}
	if t.cursorPos >= t.offset+inputWidth {
		t.offset = t.cursorPos - inputWidth + 1
	}

	// Clear input area
	clearStyle := tcell.StyleDefault.Background(bgColor)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, ' ', nil, clearStyle)
	}

	// Draw value or placeholder
	if t.value == "" && !t.focused {
		// Draw placeholder
		col := inputX
		for _, r := range t.placeholder {
			if col < inputX+inputWidth {
				screen.SetContent(col, row, r, nil, placeholderStyle)
				col++
			}
		}
	} else {
		// Draw value
		visibleValue := t.value
		if t.offset < len(visibleValue) {
			visibleValue = visibleValue[t.offset:]
		} else {
			visibleValue = ""
		}
		if len(visibleValue) > inputWidth {
			visibleValue = visibleValue[:inputWidth]
		}

		col := inputX
		for i, r := range visibleValue {
			style := inputStyle
			// Draw cursor
			if t.focused && t.offset+i == t.cursorPos {
				style = tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
			}
			screen.SetContent(col, row, r, nil, style)
			col++
		}

		// Draw cursor at end
		if t.focused && t.cursorPos >= len(t.value) {
			cursorStyle := tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
			if col < inputX+inputWidth {
				screen.SetContent(col, row, ' ', nil, cursorStyle)
			}
		}
	}
	row++

	// Draw bottom border
	screen.SetContent(x, row, '╰', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╯', nil, borderStyle)
	row++

	// Draw error message if present
	if t.error != "" && row < y+height {
		errorStyle := tcell.StyleDefault.Background(bgColor).Foreground(errorColor)
		col := x
		for _, r := range "  " + t.error {
			if col < x+width {
				screen.SetContent(col, row, r, nil, errorStyle)
				col++
			}
		}
	}
}

// InputHandler handles keyboard input.
func (t *TextField) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		oldValue := t.value

		switch event.Key() {
		case tcell.KeyLeft:
			if t.cursorPos > 0 {
				t.cursorPos--
			}
		case tcell.KeyRight:
			if t.cursorPos < len(t.value) {
				t.cursorPos++
			}
		case tcell.KeyHome:
			t.cursorPos = 0
		case tcell.KeyEnd:
			t.cursorPos = len(t.value)
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if t.cursorPos > 0 {
				t.value = t.value[:t.cursorPos-1] + t.value[t.cursorPos:]
				t.cursorPos--
				t.validate()
				t.emitChange(oldValue, t.value)
			}
		case tcell.KeyDelete:
			if t.cursorPos < len(t.value) {
				t.value = t.value[:t.cursorPos] + t.value[t.cursorPos+1:]
				t.validate()
				t.emitChange(oldValue, t.value)
			}
		case tcell.KeyEnter:
			t.emitSubmit()
		case tcell.KeyCtrlU:
			t.value = t.value[t.cursorPos:]
			t.cursorPos = 0
			t.validate()
			t.emitChange(oldValue, t.value)
		case tcell.KeyCtrlK:
			t.value = t.value[:t.cursorPos]
			t.validate()
			t.emitChange(oldValue, t.value)
		case tcell.KeyCtrlW:
			// Delete word backward
			if t.cursorPos > 0 {
				pos := t.cursorPos - 1
				for pos > 0 && t.value[pos] == ' ' {
					pos--
				}
				for pos > 0 && t.value[pos-1] != ' ' {
					pos--
				}
				t.value = t.value[:pos] + t.value[t.cursorPos:]
				t.cursorPos = pos
				t.validate()
				t.emitChange(oldValue, t.value)
			}
		case tcell.KeyRune:
			r := event.Rune()
			t.value = t.value[:t.cursorPos] + string(r) + t.value[t.cursorPos:]
			t.cursorPos++
			t.validate()
			t.emitChange(oldValue, t.value)
		}
	})
}

// Focus handles focus.
func (t *TextField) Focus(delegate func(tview.Primitive)) {
	t.focused = true
	t.Box.Focus(delegate)
}

// Blur handles blur.
func (t *TextField) Blur() {
	t.focused = false
	t.Box.Blur()
}

// HasFocus returns whether the field has focus.
func (t *TextField) HasFocus() bool {
	return t.focused
}

// GetFieldHeight returns the preferred height for this field.
func (t *TextField) GetFieldHeight() int {
	height := 3 // border top, input, border bottom
	if t.label != "" {
		height++
	}
	if t.error != "" {
		height++
	}
	return height
}

// TextArea is a multi-line text input.
// It implements ValueProvider[string].
type TextArea struct {
	*tview.Box
	BaseEventEmitter

	name        string
	label       string
	placeholder string
	lines       []string
	cursorRow   int
	cursorCol   int
	offsetRow   int
	offsetCol   int

	maxLines int

	focused bool

	// Typed handler
	onChange ChangeHandler[string]
}

// NewTextArea creates a new TextArea.
func NewTextArea(name string) *TextArea {
	return &TextArea{
		Box:      tview.NewBox(),
		name:     name,
		lines:    []string{""},
		maxLines: 100,
	}
}

// SetLabel sets the field label.
func (t *TextArea) SetLabel(label string) *TextArea {
	t.label = label
	return t
}

// SetPlaceholder sets the placeholder text.
func (t *TextArea) SetPlaceholder(placeholder string) *TextArea {
	t.placeholder = placeholder
	return t
}

// SetValue sets the current value.
func (t *TextArea) SetValue(value string) *TextArea {
	t.lines = strings.Split(value, "\n")
	if len(t.lines) == 0 {
		t.lines = []string{""}
	}
	t.cursorRow = 0
	t.cursorCol = 0
	return t
}

// GetValue returns the current value.
func (t *TextArea) GetValue() string {
	return strings.Join(t.lines, "\n")
}

// Value returns the current value (alias for GetValue).
// This method is part of the ValueProvider interface.
func (t *TextArea) Value() string {
	return t.GetValue()
}

// HasValue returns true if the text area has a non-empty value.
func (t *TextArea) HasValue() bool {
	return len(t.lines) > 1 || (len(t.lines) == 1 && t.lines[0] != "")
}

// Clear resets the text area to an empty value.
func (t *TextArea) Clear() {
	t.SetValue("")
}

// GetName returns the field name.
func (t *TextArea) GetName() string {
	return t.name
}

// SetMaxLines sets the maximum number of lines.
func (t *TextArea) SetMaxLines(max int) *TextArea {
	t.maxLines = max
	return t
}

// SetOnChange sets the change handler (new API).
func (t *TextArea) SetOnChange(handler ChangeHandler[string]) *TextArea {
	t.onChange = handler
	return t
}

// emitChange emits a change event to all handlers.
func (t *TextArea) emitChange(oldValue, newValue string) {
	event := NewChangeEvent(t.name, oldValue, newValue)

	// Typed handler
	if t.onChange != nil {
		t.onChange(event)
	}

	// Generic event bus
	t.EmitEvent(event)
}

// Draw renders the text area.
func (t *TextArea) Draw(screen tcell.Screen) {
	t.Box.DrawForSubclass(screen, t)
	x, y, width, height := t.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	borderColor := theme.Border()
	borderFocusColor := theme.BorderFocus()

	row := y

	// Draw label if present
	if t.label != "" {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := x
		for _, r := range t.label {
			if col < x+width {
				screen.SetContent(col, row, r, nil, labelStyle)
				col++
			}
		}
		row++
	}

	// Determine border color
	currentBorderColor := borderColor
	if t.focused {
		currentBorderColor = borderFocusColor
	}

	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(currentBorderColor)
	textStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	placeholderStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	// Draw top border
	screen.SetContent(x, row, '╭', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╮', nil, borderStyle)
	row++

	// Calculate text area dimensions
	textHeight := height - 3 // label + borders
	if t.label == "" {
		textHeight++
	}

	// Adjust offset for cursor visibility
	if t.cursorRow < t.offsetRow {
		t.offsetRow = t.cursorRow
	}
	if t.cursorRow >= t.offsetRow+textHeight {
		t.offsetRow = t.cursorRow - textHeight + 1
	}

	// Draw text lines
	for i := 0; i < textHeight; i++ {
		lineIdx := t.offsetRow + i
		screen.SetContent(x, row, '│', nil, borderStyle)
		screen.SetContent(x+width-1, row, '│', nil, borderStyle)

		// Clear line
		clearStyle := tcell.StyleDefault.Background(bgColor)
		for col := x + 1; col < x+width-1; col++ {
			screen.SetContent(col, row, ' ', nil, clearStyle)
		}

		if lineIdx < len(t.lines) {
			line := t.lines[lineIdx]
			col := x + 2
			for j, r := range line {
				if col >= x+width-2 {
					break
				}
				style := textStyle
				if t.focused && lineIdx == t.cursorRow && j == t.cursorCol {
					style = tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
				}
				screen.SetContent(col, row, r, nil, style)
				col++
			}
			// Draw cursor at end of line
			if t.focused && lineIdx == t.cursorRow && t.cursorCol >= len(line) {
				cursorStyle := tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
				if col < x+width-2 {
					screen.SetContent(col, row, ' ', nil, cursorStyle)
				}
			}
		} else if lineIdx == 0 && len(t.lines) == 1 && t.lines[0] == "" && !t.focused {
			// Draw placeholder on first empty line
			col := x + 2
			for _, r := range t.placeholder {
				if col >= x+width-2 {
					break
				}
				screen.SetContent(col, row, r, nil, placeholderStyle)
				col++
			}
		}
		row++
	}

	// Draw bottom border
	screen.SetContent(x, row, '╰', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╯', nil, borderStyle)
}

// InputHandler handles keyboard input.
func (t *TextArea) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		oldValue := t.GetValue()
		currentLine := t.lines[t.cursorRow]

		switch event.Key() {
		case tcell.KeyUp:
			if t.cursorRow > 0 {
				t.cursorRow--
				if t.cursorCol > len(t.lines[t.cursorRow]) {
					t.cursorCol = len(t.lines[t.cursorRow])
				}
			}
		case tcell.KeyDown:
			if t.cursorRow < len(t.lines)-1 {
				t.cursorRow++
				if t.cursorCol > len(t.lines[t.cursorRow]) {
					t.cursorCol = len(t.lines[t.cursorRow])
				}
			}
		case tcell.KeyLeft:
			if t.cursorCol > 0 {
				t.cursorCol--
			} else if t.cursorRow > 0 {
				t.cursorRow--
				t.cursorCol = len(t.lines[t.cursorRow])
			}
		case tcell.KeyRight:
			if t.cursorCol < len(currentLine) {
				t.cursorCol++
			} else if t.cursorRow < len(t.lines)-1 {
				t.cursorRow++
				t.cursorCol = 0
			}
		case tcell.KeyHome:
			t.cursorCol = 0
		case tcell.KeyEnd:
			t.cursorCol = len(currentLine)
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if t.cursorCol > 0 {
				t.lines[t.cursorRow] = currentLine[:t.cursorCol-1] + currentLine[t.cursorCol:]
				t.cursorCol--
			} else if t.cursorRow > 0 {
				// Merge with previous line
				prevLine := t.lines[t.cursorRow-1]
				t.cursorCol = len(prevLine)
				t.lines[t.cursorRow-1] = prevLine + currentLine
				t.lines = append(t.lines[:t.cursorRow], t.lines[t.cursorRow+1:]...)
				t.cursorRow--
			}
			t.emitChange(oldValue, t.GetValue())
		case tcell.KeyDelete:
			if t.cursorCol < len(currentLine) {
				t.lines[t.cursorRow] = currentLine[:t.cursorCol] + currentLine[t.cursorCol+1:]
			} else if t.cursorRow < len(t.lines)-1 {
				// Merge with next line
				t.lines[t.cursorRow] = currentLine + t.lines[t.cursorRow+1]
				t.lines = append(t.lines[:t.cursorRow+1], t.lines[t.cursorRow+2:]...)
			}
			t.emitChange(oldValue, t.GetValue())
		case tcell.KeyEnter:
			if len(t.lines) < t.maxLines {
				// Split line at cursor
				newLine := currentLine[t.cursorCol:]
				t.lines[t.cursorRow] = currentLine[:t.cursorCol]
				t.lines = append(t.lines[:t.cursorRow+1], append([]string{newLine}, t.lines[t.cursorRow+1:]...)...)
				t.cursorRow++
				t.cursorCol = 0
				t.emitChange(oldValue, t.GetValue())
			}
		case tcell.KeyRune:
			r := event.Rune()
			t.lines[t.cursorRow] = currentLine[:t.cursorCol] + string(r) + currentLine[t.cursorCol:]
			t.cursorCol++
			t.emitChange(oldValue, t.GetValue())
		}
	})
}

// Focus handles focus.
func (t *TextArea) Focus(delegate func(tview.Primitive)) {
	t.focused = true
	t.Box.Focus(delegate)
}

// Blur handles blur.
func (t *TextArea) Blur() {
	t.focused = false
	t.Box.Blur()
}

// HasFocus returns whether the field has focus.
func (t *TextArea) HasFocus() bool {
	return t.focused
}

// GetFieldHeight returns the preferred height for this field.
func (t *TextArea) GetFieldHeight() int {
	height := 10 // Default height for text area
	if t.label != "" {
		height++
	}
	return height
}
