package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// SelectOption represents an option in a select dropdown.
type SelectOption struct {
	Label string
	Value string
}

// Select is a dropdown selection component.
type Select struct {
	*tview.Box

	name        string
	label       string
	placeholder string
	options     []SelectOption
	selected    int
	expanded    bool

	focused bool

	onChange func(index int, option SelectOption)
}

// NewSelect creates a new Select component.
func NewSelect(name string) *Select {
	return &Select{
		Box:      tview.NewBox(),
		name:     name,
		selected: -1,
	}
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

// GetSelected returns the selected option.
func (s *Select) GetSelected() (int, SelectOption) {
	if s.selected >= 0 && s.selected < len(s.options) {
		return s.selected, s.options[s.selected]
	}
	return -1, SelectOption{}
}

// GetValue returns the selected value.
func (s *Select) GetValue() string {
	if s.selected >= 0 && s.selected < len(s.options) {
		return s.options[s.selected].Value
	}
	return ""
}

// GetName returns the field name.
func (s *Select) GetName() string {
	return s.name
}

// SetOnChange sets the callback for selection changes.
func (s *Select) SetOnChange(fn func(index int, option SelectOption)) *Select {
	s.onChange = fn
	return s
}

// Draw renders the select component.
func (s *Select) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen, s)
	x, y, width, height := s.GetInnerRect()

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
	if s.label != "" {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := x
		for _, r := range s.label {
			if col < x+width {
				screen.SetContent(col, row, r, nil, labelStyle)
				col++
			}
		}
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
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, row, ' ', nil, clearStyle)
	}

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

	for _, r := range displayText {
		if col < x+width-4 {
			screen.SetContent(col, row, r, nil, displayStyle)
			col++
		}
	}

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
			for col := x; col < x+width; col++ {
				screen.SetContent(col, row, ' ', nil, optStyle)
			}

			// Draw option text
			col := x + 2
			for _, r := range opt.Label {
				if col < x+width-2 {
					screen.SetContent(col, row, r, nil, optStyle)
					col++
				}
			}
			row++
		}
	}
}

// InputHandler handles keyboard input.
func (s *Select) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return s.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyRune:
			if event.Key() == tcell.KeyRune && event.Rune() != ' ' {
				return
			}
			if s.expanded {
				s.expanded = false
				if s.onChange != nil && s.selected >= 0 {
					s.onChange(s.selected, s.options[s.selected])
				}
			} else {
				s.expanded = true
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

		// Vim keys
		if event.Key() == tcell.KeyRune {
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
type MultiSelect struct {
	*tview.Box

	name     string
	label    string
	options  []SelectOption
	selected map[int]bool
	cursor   int
	expanded bool

	focused bool

	onChange func(selected []SelectOption)
}

// NewMultiSelect creates a new MultiSelect component.
func NewMultiSelect(name string) *MultiSelect {
	return &MultiSelect{
		Box:      tview.NewBox(),
		name:     name,
		selected: make(map[int]bool),
	}
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

// GetSelected returns the selected options.
func (m *MultiSelect) GetSelected() []SelectOption {
	var result []SelectOption
	for i := 0; i < len(m.options); i++ {
		if m.selected[i] {
			result = append(result, m.options[i])
		}
	}
	return result
}

// GetValues returns the selected values.
func (m *MultiSelect) GetValues() []string {
	var result []string
	for i := 0; i < len(m.options); i++ {
		if m.selected[i] {
			result = append(result, m.options[i].Value)
		}
	}
	return result
}

// GetName returns the field name.
func (m *MultiSelect) GetName() string {
	return m.name
}

// SetOnChange sets the callback for selection changes.
func (m *MultiSelect) SetOnChange(fn func(selected []SelectOption)) *MultiSelect {
	m.onChange = fn
	return m
}

// Draw renders the multi-select component.
func (m *MultiSelect) Draw(screen tcell.Screen) {
	m.Box.DrawForSubclass(screen, m)
	x, y, width, height := m.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	successColor := theme.Success()
	borderColor := theme.Border()
	borderFocusColor := theme.BorderFocus()

	row := y

	// Draw label if present
	if m.label != "" {
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		col := x
		for _, r := range m.label {
			if col < x+width {
				screen.SetContent(col, row, r, nil, labelStyle)
				col++
			}
		}
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
		for col := x + 1; col < x+width-1; col++ {
			screen.SetContent(col, row, ' ', nil, rowStyle)
		}

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
		for _, r := range opt.Label {
			if col < x+width-2 {
				screen.SetContent(col, row, r, nil, rowStyle)
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
		case tcell.KeyEnter, tcell.KeyRune:
			if event.Key() == tcell.KeyEnter || event.Rune() == ' ' {
				m.selected[m.cursor] = !m.selected[m.cursor]
				if m.onChange != nil {
					m.onChange(m.GetSelected())
				}
			}
			// Vim keys
			if event.Key() == tcell.KeyRune {
				switch event.Rune() {
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

// GetFieldHeight returns the preferred height for this field.
func (m *MultiSelect) GetFieldHeight() int {
	height := 2 + len(m.options) // borders + options
	if m.label != "" {
		height++
	}
	return height
}
