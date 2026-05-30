package components

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SelectOption represents an option in a select dropdown.
type SelectOption struct {
	Label string
	Value string
}

// Select is a dropdown selection component.
// It implements IndexedValueProvider[string].
type Select struct {
	widgetBase
	BaseEventEmitter

	name        string
	label       string
	placeholder string
	options     []SelectOption
	selected    int
	expanded    bool

	focused bool

	// Typed handler
	onChange ChangeHandler[SelectOption]
}

// NewSelect creates a new Select component.
func NewSelect(name string) *Select {
	s := &Select{
		name:     name,
		selected: -1,
	}
	s.initWidget(tview.NewBox())
	return s
}

// SetLabel sets the field label.
func (s *Select) SetLabel(label string) *Select {
	s.label = label
	return s
}

// SetPlaceholder sets the placeholder text.
func (s *Select) SetPlaceholder(placeholder string) *Select {
	s.placeholder = placeholder
	return s
}

// SetOptions sets the available options.
func (s *Select) SetOptions(options []string) *Select {
	s.options = make([]SelectOption, len(options))
	for i, opt := range options {
		s.options[i] = SelectOption{Label: opt, Value: opt}
	}
	return s
}

// SetOptionsWithValues sets options with custom values.
func (s *Select) SetOptionsWithValues(options []SelectOption) *Select {
	s.options = options
	return s
}

// SetDefault sets the default selected option by value.
func (s *Select) SetDefault(value string) *Select {
	for i, opt := range s.options {
		if opt.Value == value {
			s.selected = i
			break
		}
	}
	return s
}

// SetSelected sets the selected index.
func (s *Select) SetSelected(index int) *Select {
	if index >= 0 && index < len(s.options) {
		s.selected = index
	}
	return s
}

// SetSelectedIndex sets the selected index, returning an error if out of range.
func (s *Select) SetSelectedIndex(index int) error {
	if index < -1 || index >= len(s.options) {
		return fmt.Errorf("index %d out of range [0, %d)", index, len(s.options))
	}
	s.selected = index
	return nil
}

// SetSelectedValue sets the selected option by value.
func (s *Select) SetSelectedValue(value string) error {
	for i, opt := range s.options {
		if opt.Value == value {
			s.selected = i
			return nil
		}
	}
	return fmt.Errorf("value %q not found in options", value)
}

// SelectedIndex returns the selected index (-1 if none).
func (s *Select) SelectedIndex() int {
	return s.selected
}

// SelectedOption returns the selected option.
func (s *Select) SelectedOption() SelectOption {
	if s.selected >= 0 && s.selected < len(s.options) {
		return s.options[s.selected]
	}
	return SelectOption{}
}

// Value returns the selected value.
// This method is part of the ValueProvider interface.
func (s *Select) Value() string {
	if s.selected >= 0 && s.selected < len(s.options) {
		return s.options[s.selected].Value
	}
	return ""
}

// GetValue returns the selected value.
func (s *Select) GetValue() string {
	return s.Value()
}

// HasValue returns true if an option is selected.
func (s *Select) HasValue() bool {
	return s.selected >= 0 && s.selected < len(s.options)
}

// Clear resets the selection to none.
func (s *Select) Clear() {
	s.selected = -1
}

// GetName returns the field name.
func (s *Select) GetName() string {
	return s.name
}

// SetOnChange sets the change handler (new API).
func (s *Select) SetOnChange(handler ChangeHandler[SelectOption]) *Select {
	s.onChange = handler
	return s
}

// emitChange emits a change event to all handlers.
func (s *Select) emitChange(oldIndex, newIndex int) {
	var oldOption, newOption SelectOption
	if oldIndex >= 0 && oldIndex < len(s.options) {
		oldOption = s.options[oldIndex]
	}
	if newIndex >= 0 && newIndex < len(s.options) {
		newOption = s.options[newIndex]
	}

	event := NewChangeEvent(s.name, oldOption, newOption).WithIndex(newIndex)

	// Typed handler
	if s.onChange != nil {
		s.onChange(event)
	}

	// Generic event bus
	s.EmitEvent(event)
}

// Draw renders the select component.
func (s *Select) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen, s)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	th := s.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	accentColor := th.Accent()
	borderColor := th.Border()
	borderFocusColor := th.BorderFocus()

	row := y

	// Draw label if present
	if s.label != "" {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		drawText(screen, x, row, width, s.label, labelStyle)
		row++
	}

	// Determine border color
	currentBorderColor := borderColor
	if s.focused {
		currentBorderColor = borderFocusColor
	}

	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(currentBorderColor)
	textStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	placeholderStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	// Draw select box border
	screen.SetContent(x, row, '╭', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╮', nil, borderStyle)
	row++

	// Draw select value line
	screen.SetContent(x, row, '│', nil, borderStyle)
	screen.SetContent(x+width-1, row, '│', nil, borderStyle)

	// Clear inner area
	clearStyle := tcell.StyleDefault.Background(bgColor)
	fillLine(screen, x+1, row, width-2, clearStyle)

	// Draw selected value or placeholder
	col := x + 2
	var displayText string
	var displayStyle tcell.Style
	if s.selected >= 0 && s.selected < len(s.options) {
		displayText = s.options[s.selected].Label
		displayStyle = textStyle
	} else {
		displayText = s.placeholder
		displayStyle = placeholderStyle
	}

	drawText(screen, col, row, width-6, displayText, displayStyle)

	// Draw dropdown indicator
	indicatorStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)
	if s.expanded {
		screen.SetContent(x+width-3, row, '▲', nil, indicatorStyle)
	} else {
		screen.SetContent(x+width-3, row, '▼', nil, indicatorStyle)
	}
	row++

	// Draw bottom border
	screen.SetContent(x, row, '╰', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╯', nil, borderStyle)
	row++

	// Draw dropdown options if expanded
	if s.expanded && row < y+height {
		optionBg := bgColor
		optionFg := fgColor
		selectedBg := accentColor
		selectedFg := bgColor

		for i, opt := range s.options {
			if row >= y+height {
				break
			}

			var optStyle tcell.Style
			if i == s.selected {
				optStyle = tcell.StyleDefault.Background(selectedBg).Foreground(selectedFg)
			} else {
				optStyle = tcell.StyleDefault.Background(optionBg).Foreground(optionFg)
			}

			// Draw option background
			fillLine(screen, x, row, width, optStyle)

			// Draw option text
			drawText(screen, x+2, row, width-4, opt.Label, optStyle)
			row++
		}
	}
}

// InputHandler handles keyboard input.
func (s *Select) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return s.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		oldSelected := s.selected

		switch event.Key() {
		case tcell.KeyRune:
			// Space to toggle dropdown (Enter is reserved for form submit)
			if event.Rune() == ' ' {
				if s.expanded {
					s.expanded = false
					if s.selected >= 0 {
						s.emitChange(oldSelected, s.selected)
					}
				} else {
					s.expanded = true
				}
				return
			}
			// Vim keys
			switch event.Rune() {
			case 'j':
				if s.expanded && s.selected < len(s.options)-1 {
					s.selected++
				}
			case 'k':
				if s.expanded && s.selected > 0 {
					s.selected--
				}
			}
		case tcell.KeyEscape:
			s.expanded = false
		case tcell.KeyUp:
			if s.expanded {
				if s.selected > 0 {
					s.selected--
				}
			} else {
				s.expanded = true
			}
		case tcell.KeyDown:
			if s.expanded {
				if s.selected < len(s.options)-1 {
					s.selected++
				}
			} else {
				s.expanded = true
			}
		}
	})
}

// Focus handles focus.
func (s *Select) Focus(delegate func(tview.Primitive)) {
	s.focused = true
	s.Box.Focus(delegate)
}

// Blur handles blur.
func (s *Select) Blur() {
	s.focused = false
	s.expanded = false
	s.Box.Blur()
}

// HasFocus returns whether the select has focus.
func (s *Select) HasFocus() bool {
	return s.focused
}

// SetFieldValue sets the select's value from an any. Accepts a string
// (matched against option values) or an int (option index). Implements FormField.
func (s *Select) SetFieldValue(value any) error {
	switch v := value.(type) {
	case string:
		s.SetDefault(v)
		return nil
	case int:
		s.SetSelected(v)
		return nil
	default:
		return fmt.Errorf("expected string or int, got %T", value)
	}
}

// FieldValue returns the selected option value as an any. Implements FormField.
func (s *Select) FieldValue() any { return s.GetValue() }

// ClearField resets the selection. Implements FormField.
func (s *Select) ClearField() { s.SetSelected(-1) }

// GetFieldHeight returns the preferred height for this field.
func (s *Select) GetFieldHeight() int {
	height := 3 // border top, value, border bottom
	if s.label != "" {
		height++
	}
	if s.expanded {
		height += len(s.options)
	}
	return height
}

// MultiSelect allows multiple option selection.
// It implements MultiValueProvider[string].
type MultiSelect struct {
	widgetBase
	BaseEventEmitter

	name     string
	label    string
	options  []SelectOption
	selected map[int]bool
	cursor   int
	expanded bool

	focused bool

	// Typed handler
	onChange ChangeHandler[[]SelectOption]
}

// NewMultiSelect creates a new MultiSelect component.
func NewMultiSelect(name string) *MultiSelect {
	m := &MultiSelect{
		name:     name,
		selected: make(map[int]bool),
	}
	m.initWidget(tview.NewBox())
	return m
}

// SetLabel sets the field label.
func (m *MultiSelect) SetLabel(label string) *MultiSelect {
	m.label = label
	return m
}

// SetOptions sets the available options.
func (m *MultiSelect) SetOptions(options []string) *MultiSelect {
	m.options = make([]SelectOption, len(options))
	for i, opt := range options {
		m.options[i] = SelectOption{Label: opt, Value: opt}
	}
	return m
}

// SetOptionsWithValues sets options with custom values.
func (m *MultiSelect) SetOptionsWithValues(options []SelectOption) *MultiSelect {
	m.options = options
	return m
}

// SetSelected sets the selected indices.
func (m *MultiSelect) SetSelected(indices []int) *MultiSelect {
	m.selected = make(map[int]bool)
	for _, idx := range indices {
		if idx >= 0 && idx < len(m.options) {
			m.selected[idx] = true
		}
	}
	return m
}

// SetSelectedIndices sets the selected indices, returning an error if any index is out of range.
func (m *MultiSelect) SetSelectedIndices(indices []int) error {
	m.selected = make(map[int]bool)
	for _, idx := range indices {
		if idx < 0 || idx >= len(m.options) {
			return fmt.Errorf("index %d out of range [0, %d)", idx, len(m.options))
		}
		m.selected[idx] = true
	}
	return nil
}

// SetSelectedValues sets the selected options by value.
func (m *MultiSelect) SetSelectedValues(values []string) error {
	m.selected = make(map[int]bool)
	valueSet := make(map[string]bool)
	for _, v := range values {
		valueSet[v] = true
	}
	for i, opt := range m.options {
		if valueSet[opt.Value] {
			m.selected[i] = true
			delete(valueSet, opt.Value)
		}
	}
	if len(valueSet) > 0 {
		for v := range valueSet {
			return fmt.Errorf("value %q not found in options", v)
		}
	}
	return nil
}

// SelectedIndices returns all selected indices (sorted).
func (m *MultiSelect) SelectedIndices() []int {
	indices := make([]int, 0, len(m.selected))
	for i := range m.selected {
		if m.selected[i] {
			indices = append(indices, i)
		}
	}
	sort.Ints(indices)
	return indices
}

// SelectedOptions returns all selected options.
func (m *MultiSelect) SelectedOptions() []SelectOption {
	var result []SelectOption
	for i := 0; i < len(m.options); i++ {
		if m.selected[i] {
			result = append(result, m.options[i])
		}
	}
	return result
}

// Values returns all selected value strings.
// This method is part of the MultiValueProvider interface.
func (m *MultiSelect) Values() []string {
	var result []string
	for i := 0; i < len(m.options); i++ {
		if m.selected[i] {
			result = append(result, m.options[i].Value)
		}
	}
	return result
}

// HasValue returns true if at least one option is selected.
func (m *MultiSelect) HasValue() bool {
	return len(m.selected) > 0
}

// Clear deselects all options.
func (m *MultiSelect) Clear() {
	m.selected = make(map[int]bool)
}

// GetName returns the field name.
func (m *MultiSelect) GetName() string {
	return m.name
}

// SetOnChange sets the change handler (new API).
func (m *MultiSelect) SetOnChange(handler ChangeHandler[[]SelectOption]) *MultiSelect {
	m.onChange = handler
	return m
}

// emitChange emits a change event to all handlers.
func (m *MultiSelect) emitChange(oldSelected, newSelected []SelectOption) {
	event := NewChangeEvent(m.name, oldSelected, newSelected)

	// Typed handler
	if m.onChange != nil {
		m.onChange(event)
	}

	// Generic event bus
	m.EmitEvent(event)
}

// Draw renders the multi-select component.
func (m *MultiSelect) Draw(screen tcell.Screen) {
	m.Box.DrawForSubclass(screen, m)
	x, y, width, height := m.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	th := m.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	accentColor := th.Accent()
	successColor := th.Success()
	borderColor := th.Border()
	borderFocusColor := th.BorderFocus()

	row := y

	// Draw label if present
	if m.label != "" {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		drawText(screen, x, row, width, m.label, labelStyle)
		row++
	}

	// Determine border color
	currentBorderColor := borderColor
	if m.focused {
		currentBorderColor = borderFocusColor
	}

	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(currentBorderColor)

	// Draw top border
	screen.SetContent(x, row, '╭', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╮', nil, borderStyle)
	row++

	// Draw options
	for i, opt := range m.options {
		if row >= y+height-1 {
			break
		}

		screen.SetContent(x, row, '│', nil, borderStyle)
		screen.SetContent(x+width-1, row, '│', nil, borderStyle)

		// Determine row style
		var rowBg, rowFg tcell.Color
		if m.focused && i == m.cursor {
			rowBg = accentColor
			rowFg = bgColor
		} else {
			rowBg = bgColor
			rowFg = fgColor
		}
		rowStyle := tcell.StyleDefault.Background(rowBg).Foreground(rowFg)

		// Clear row
		fillLine(screen, x+1, row, width-2, rowStyle)

		// Draw checkbox
		col := x + 2
		checkStyle := rowStyle
		if m.selected[i] {
			checkStyle = tcell.StyleDefault.Background(rowBg).Foreground(successColor)
			if m.focused && i == m.cursor {
				checkStyle = rowStyle
			}
			screen.SetContent(col, row, '[', nil, checkStyle)
			col++
			screen.SetContent(col, row, '✓', nil, checkStyle)
			col++
			screen.SetContent(col, row, ']', nil, checkStyle)
			col++
		} else {
			screen.SetContent(col, row, '[', nil, rowStyle)
			col++
			screen.SetContent(col, row, ' ', nil, rowStyle)
			col++
			screen.SetContent(col, row, ']', nil, rowStyle)
			col++
		}

		// Draw label
		col++
		drawText(screen, col, row, x+width-2-col, opt.Label, rowStyle)
		row++
	}

	// Draw bottom border
	screen.SetContent(x, row, '╰', nil, borderStyle)
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, '─', nil, borderStyle)
	}
	screen.SetContent(x+width-1, row, '╯', nil, borderStyle)

	_ = fgDimColor // Available for future use
}

// InputHandler handles keyboard input.
func (m *MultiSelect) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tcell.KeyDown:
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ':
				// Space to toggle selection (Enter is reserved for form submit)
				oldSelected := m.SelectedOptions()
				m.selected[m.cursor] = !m.selected[m.cursor]
				m.emitChange(oldSelected, m.SelectedOptions())
			case 'j':
				if m.cursor < len(m.options)-1 {
					m.cursor++
				}
			case 'k':
				if m.cursor > 0 {
					m.cursor--
				}
			}
		}
	})
}

// Focus handles focus.
func (m *MultiSelect) Focus(delegate func(tview.Primitive)) {
	m.focused = true
	m.Box.Focus(delegate)
}

// Blur handles blur.
func (m *MultiSelect) Blur() {
	m.focused = false
	m.Box.Blur()
}

// HasFocus returns whether the multi-select has focus.
func (m *MultiSelect) HasFocus() bool {
	return m.focused
}

// SetFieldValue sets the multi-select's value from an any. Accepts []string
// (matched against option values) or []int (option indices). Implements FormField.
func (m *MultiSelect) SetFieldValue(value any) error {
	switch v := value.(type) {
	case []string:
		indices := make([]int, 0, len(v))
		for _, val := range v {
			for i, opt := range m.options {
				if opt.Value == val {
					indices = append(indices, i)
					break
				}
			}
		}
		m.SetSelected(indices)
		return nil
	case []int:
		m.SetSelected(v)
		return nil
	default:
		return fmt.Errorf("expected []string or []int, got %T", value)
	}
}

// FieldValue returns the selected values as an any ([]string). Implements FormField.
func (m *MultiSelect) FieldValue() any { return m.Values() }

// ClearField resets the selection. Implements FormField.
func (m *MultiSelect) ClearField() { m.SetSelected(nil) }

// GetFieldHeight returns the preferred height for this field.
func (m *MultiSelect) GetFieldHeight() int {
	height := 2 + len(m.options) // borders + options
	if m.label != "" {
		height++
	}
	return height
}
