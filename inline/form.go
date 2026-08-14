package inline

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var ErrFormCancelled = errors.New("inline: form cancelled")

// Choice is one selectable value in a SelectField or MultiSelectField.
type Choice struct {
	Value       string
	Label       string
	Description string
	Disabled    bool
}

// NewChoice creates an enabled choice.
func NewChoice(value, label string) Choice { return Choice{Value: value, Label: label} }

// FormResult contains typed values keyed by field ID.
type FormResult map[string]any

// FormField is a field that can participate in a Form. The private methods
// intentionally keep the first version of the field contract closed.
type FormField interface {
	ID() string
	Value() any
	draw(*Frame, int, int, bool, InlineTheme) int
	handle(*tcell.EventKey)
	validate() error
	summary() string
	labelSummary() string
}

// TextField edits one line of Unicode text.
type TextField struct {
	id, label, placeholder string
	value                  []rune
	cursor                 int
	required, password     bool
	validateFn             func(string) error
	err                    error
}

func NewTextField(id, label string) *TextField { return &TextField{id: id, label: label} }
func (f *TextField) ID() string                { return f.id }
func (f *TextField) Value() any                { return string(f.value) }
func (f *TextField) SetValue(value string) *TextField {
	f.value = []rune(value)
	f.cursor = len(f.value)
	return f
}
func (f *TextField) SetPlaceholder(value string) *TextField    { f.placeholder = value; return f }
func (f *TextField) Required() *TextField                      { f.required = true; return f }
func (f *TextField) Password() *TextField                      { f.password = true; return f }
func (f *TextField) Validate(fn func(string) error) *TextField { f.validateFn = fn; return f }

func (f *TextField) handle(event *tcell.EventKey) {
	f.err = nil
	switch event.Key() {
	case tcell.KeyRune:
		f.value = slices.Insert(f.value, f.cursor, event.Rune())
		f.cursor++
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if f.cursor > 0 {
			f.value = slices.Delete(f.value, f.cursor-1, f.cursor)
			f.cursor--
		}
	case tcell.KeyDelete:
		if f.cursor < len(f.value) {
			f.value = slices.Delete(f.value, f.cursor, f.cursor+1)
		}
	case tcell.KeyLeft:
		f.cursor = max(f.cursor-1, 0)
	case tcell.KeyRight:
		f.cursor = min(f.cursor+1, len(f.value))
	case tcell.KeyHome:
		f.cursor = 0
	case tcell.KeyEnd:
		f.cursor = len(f.value)
	}
}

func (f *TextField) validate() error {
	value := string(f.value)
	if f.required && strings.TrimSpace(value) == "" {
		f.err = errors.New("required")
	} else if f.validateFn != nil {
		f.err = f.validateFn(value)
	} else {
		f.err = nil
	}
	return f.err
}

func (f *TextField) draw(frame *Frame, y, width int, focused bool, theme InlineTheme) int {
	drawThemedLabel(frame, y, width, f.label, f.required, focused, theme)
	display := string(f.value)
	if f.password {
		display = strings.Repeat("•", len(f.value))
	}
	visibleStart := 0
	available := max(width-4, 0)
	for visibleStart < f.cursor && displayWidth(string([]rune(display)[visibleStart:f.cursor])) >= available && available > 0 {
		visibleStart++
	}
	if display != "" {
		display = string([]rune(display)[visibleStart:])
	}
	style := theme.Text
	if display == "" {
		display, style = f.placeholder, theme.Muted
	}
	borderStyle := theme.Border
	if focused {
		borderStyle = theme.FocusedBorder
	}
	drawBox(frame, y+1, width, 3, theme.Borders, borderStyle)
	if focused {
		drawClipped(frame, 0, y+2, theme.Glyphs.Focus, 1, theme.Accent)
	}
	drawClipped(frame, 2, y+2, display, max(width-4, 0), style)
	if focused {
		cursorText := string(f.value[visibleStart:f.cursor])
		if f.password {
			cursorText = strings.Repeat("•", f.cursor-visibleStart)
		}
		frame.ShowCursor(min(2+displayWidth(cursorText), max(width-2, 0)), y+2)
	}
	if f.err != nil {
		drawClipped(frame, 2, y+4, theme.Glyphs.Error+" "+f.err.Error(), max(width-2, 0), theme.Error)
		return 5
	}
	return 4
}

func (f *TextField) summary() string {
	if f.password && len(f.value) > 0 {
		return strings.Repeat("•", len(f.value))
	}
	return string(f.value)
}

// SelectField chooses one value.
type SelectField struct {
	id, label string
	choices   []Choice
	cursor    int
	selected  int
	required  bool
	err       error
}

func NewSelectField(id, label string, choices ...Choice) *SelectField {
	f := &SelectField{id: id, label: label, choices: append([]Choice(nil), choices...), selected: -1}
	f.cursor = f.nextEnabled(-1, 1)
	return f
}
func (f *SelectField) ID() string             { return f.id }
func (f *SelectField) Required() *SelectField { f.required = true; return f }
func (f *SelectField) Value() any {
	if f.selected < 0 || f.selected >= len(f.choices) {
		return ""
	}
	return f.choices[f.selected].Value
}
func (f *SelectField) handle(event *tcell.EventKey) {
	f.err = nil
	switch event.Key() {
	case tcell.KeyUp, tcell.KeyLeft:
		f.cursor = f.nextEnabled(f.cursor, -1)
	case tcell.KeyDown, tcell.KeyRight:
		f.cursor = f.nextEnabled(f.cursor, 1)
	case tcell.KeyRune:
		if event.Rune() != ' ' {
			return
		}
		fallthrough
	case tcell.KeyEnter:
		if f.cursor >= 0 {
			f.selected = f.cursor
		}
	}
}
func (f *SelectField) nextEnabled(from, direction int) int {
	if len(f.choices) == 0 {
		return -1
	}
	for offset := 1; offset <= len(f.choices); offset++ {
		index := (from + direction*offset + len(f.choices)*2) % len(f.choices)
		if !f.choices[index].Disabled {
			return index
		}
	}
	return -1
}
func (f *SelectField) validate() error {
	if f.required && f.selected < 0 {
		f.err = errors.New("select a value")
	} else {
		f.err = nil
	}
	return f.err
}
func (f *SelectField) draw(frame *Frame, y, width int, focused bool, theme InlineTheme) int {
	drawThemedLabel(frame, y, width, f.label, f.required, focused, theme)
	rows := 1
	if focused {
		rows = max(len(f.choices), 1)
	}
	borderStyle := theme.Border
	if focused {
		borderStyle = theme.FocusedBorder
	}
	drawBox(frame, y+1, width, rows+2, theme.Borders, borderStyle)
	height := rows + 3
	if !focused {
		label := "Choose an option"
		style := theme.Muted
		if f.selected >= 0 {
			label = f.choices[f.selected].Label
			style = theme.Text
		}
		drawClipped(frame, 2, y+2, label, max(width-5, 0), style)
		drawRight(frame, y+2, max(width-2, 0), theme.Glyphs.Dropdown, theme.Muted)
	} else {
		for index, choice := range f.choices {
			marker := theme.Glyphs.Unselected
			if index == f.selected {
				marker = theme.Glyphs.Selected
			}
			style := theme.Text
			if choice.Disabled {
				style = theme.Muted
			}
			if index == f.cursor {
				drawClipped(frame, 0, y+2+index, theme.Glyphs.Focus, 1, theme.Accent)
				style = style.Bold(true)
			}
			drawClipped(frame, 2, y+2+index, marker+" "+choice.Label, max(width-4, 0), style)
			if choice.Description != "" {
				start := 4 + displayWidth(choice.Label)
				drawClipped(frame, start, y+2+index, choice.Description, max(width-start-2, 0), theme.Muted)
			}
		}
	}
	if f.err != nil {
		drawClipped(frame, 2, y+height, theme.Glyphs.Error+" "+f.err.Error(), max(width-2, 0), theme.Error)
		height++
	}
	return height
}
func (f *SelectField) summary() string {
	if f.selected < 0 || f.selected >= len(f.choices) {
		return ""
	}
	return f.choices[f.selected].Label
}

// MultiSelectField chooses zero or more values.
type MultiSelectIndicator int

const (
	IndicatorTheme MultiSelectIndicator = iota
	IndicatorCheck
	IndicatorDot
	IndicatorDiamond
	IndicatorBox
	IndicatorMinimal
	IndicatorNone
)

// IndicatorGlyphs defines a custom selected, unselected, and disabled marker set.
type IndicatorGlyphs struct{ Selected, Unselected, Disabled string }

type MultiSelectField struct {
	id, label        string
	choices          []Choice
	cursor           int
	selected         map[int]bool
	minimum, maximum int
	err              error
	indicator        MultiSelectIndicator
	indicatorGlyphs  *IndicatorGlyphs
}

func NewMultiSelectField(id, label string, choices ...Choice) *MultiSelectField {
	f := &MultiSelectField{id: id, label: label, choices: append([]Choice(nil), choices...), selected: make(map[int]bool)}
	f.cursor = f.nextEnabled(-1, 1)
	return f
}
func (f *MultiSelectField) ID() string { return f.id }
func (f *MultiSelectField) MinSelected(value int) *MultiSelectField {
	f.minimum = max(value, 0)
	return f
}
func (f *MultiSelectField) MaxSelected(value int) *MultiSelectField {
	f.maximum = max(value, 0)
	return f
}

// SetIndicator selects a built-in marker preset.
func (f *MultiSelectField) SetIndicator(indicator MultiSelectIndicator) *MultiSelectField {
	f.indicator = indicator
	f.indicatorGlyphs = nil
	return f
}

// SetIndicatorGlyphs supplies a custom marker set.
func (f *MultiSelectField) SetIndicatorGlyphs(selected, unselected, disabled string) *MultiSelectField {
	f.indicatorGlyphs = &IndicatorGlyphs{selected, unselected, disabled}
	return f
}
func (f *MultiSelectField) Value() any {
	values := make([]string, 0, len(f.selected))
	for index, choice := range f.choices {
		if f.selected[index] {
			values = append(values, choice.Value)
		}
	}
	return values
}
func (f *MultiSelectField) nextEnabled(from, direction int) int {
	if len(f.choices) == 0 {
		return -1
	}
	for offset := 1; offset <= len(f.choices); offset++ {
		index := (from + direction*offset + len(f.choices)*2) % len(f.choices)
		if !f.choices[index].Disabled {
			return index
		}
	}
	return -1
}
func (f *MultiSelectField) handle(event *tcell.EventKey) {
	f.err = nil
	switch event.Key() {
	case tcell.KeyUp, tcell.KeyLeft:
		f.cursor = f.nextEnabled(f.cursor, -1)
	case tcell.KeyDown, tcell.KeyRight:
		f.cursor = f.nextEnabled(f.cursor, 1)
	case tcell.KeyRune:
		if event.Rune() == ' ' && f.cursor >= 0 {
			if f.selected[f.cursor] {
				delete(f.selected, f.cursor)
			} else if f.maximum == 0 || len(f.selected) < f.maximum {
				f.selected[f.cursor] = true
			}
		}
	}
}
func (f *MultiSelectField) validate() error {
	count := len(f.selected)
	switch {
	case count < f.minimum:
		f.err = fmt.Errorf("select at least %d", f.minimum)
	case f.maximum > 0 && count > f.maximum:
		f.err = fmt.Errorf("select at most %d", f.maximum)
	default:
		f.err = nil
	}
	return f.err
}
func (f *MultiSelectField) draw(frame *Frame, y, width int, focused bool, theme InlineTheme) int {
	drawThemedLabel(frame, y, width, f.label, f.minimum > 0, focused, theme)
	rows := 1
	if focused {
		rows = max(len(f.choices), 1)
	}
	borderStyle := theme.Border
	if focused {
		borderStyle = theme.FocusedBorder
	}
	drawBox(frame, y+1, width, rows+2, theme.Borders, borderStyle)
	height := rows + 3
	if !focused {
		value := f.summary()
		style := theme.Text
		if value == "" {
			value = "No selections"
			style = theme.Muted
		}
		drawClipped(frame, 2, y+2, value, max(width-4, 0), style)
	} else {
		for index, choice := range f.choices {
			selected := f.selected[index]
			marker := f.indicatorMarker(selected, choice.Disabled, theme)
			style := theme.Text
			if choice.Disabled {
				style = theme.Muted
			} else if f.indicator == IndicatorNone && selected {
				style = theme.Accent.Reverse(true)
			}
			if index == f.cursor {
				drawClipped(frame, 0, y+2+index, theme.Glyphs.Focus, 1, theme.Accent)
				style = style.Bold(true)
			}
			chip := choice.Label
			if marker != "" {
				chip = marker + " " + chip
			}
			drawClipped(frame, 2, y+2+index, chip, max(width-4, 0), style)
		}
	}
	if f.err != nil {
		drawClipped(frame, 2, y+height, theme.Glyphs.Error+" "+f.err.Error(), max(width-2, 0), theme.Error)
		height++
	}
	return height
}

func (f *MultiSelectField) indicatorMarker(selected, disabled bool, theme InlineTheme) string {
	if f.indicatorGlyphs != nil {
		if disabled {
			return f.indicatorGlyphs.Disabled
		}
		if selected {
			return f.indicatorGlyphs.Selected
		}
		return f.indicatorGlyphs.Unselected
	}
	if disabled {
		return theme.Glyphs.Disabled
	}
	var on, off string
	switch f.indicator {
	case IndicatorCheck:
		on, off = "✓", "○"
	case IndicatorDot:
		on, off = "●", "○"
	case IndicatorDiamond:
		on, off = "◆", "◇"
	case IndicatorBox:
		on, off = "▣", "□"
	case IndicatorMinimal:
		on, off = "✓", "·"
	case IndicatorNone:
		return ""
	default:
		on, off = theme.Glyphs.Checked, theme.Glyphs.Unchecked
	}
	if selected {
		return on
	}
	return off
}
func (f *MultiSelectField) summary() string {
	labels := make([]string, 0, len(f.selected))
	for i, choice := range f.choices {
		if f.selected[i] {
			labels = append(labels, choice.Label)
		}
	}
	return strings.Join(labels, ", ")
}

// Form composes fields, focus traversal, validation, submission, and summary rendering.
type Form struct {
	title                string
	fields               []FormField
	focus                int
	submitted, cancelled bool
	result               FormResult
	theme                InlineTheme
}

func NewForm(title string) *Form                 { return &Form{title: title, theme: RoundedInlineTheme()} }
func (f *Form) Add(fields ...FormField) *Form    { f.fields = append(f.fields, fields...); return f }
func (f *Form) SetTheme(theme InlineTheme) *Form { f.theme = normalizedInlineTheme(theme); return f }
func (f *Form) Result() FormResult {
	result := make(FormResult, len(f.result))
	for k, v := range f.result {
		result[k] = v
	}
	return result
}
func (f *Form) Done() bool { return f.submitted || f.cancelled }
func (f *Form) HandleKey(event *tcell.EventKey) {
	if f.Done() || len(f.fields) == 0 {
		return
	}
	switch event.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		f.cancelled = true
		return
	case tcell.KeyTab:
		f.focus = (f.focus + 1) % len(f.fields)
		return
	case tcell.KeyBacktab:
		f.focus = (f.focus - 1 + len(f.fields)) % len(f.fields)
		return
	case tcell.KeyEnter:
		if _, selectField := f.fields[f.focus].(*SelectField); selectField {
			f.fields[f.focus].handle(event)
		}
		if f.focus < len(f.fields)-1 {
			f.focus++
			return
		}
		f.submit()
		return
	}
	f.fields[f.focus].handle(event)
}
func (f *Form) submit() {
	for index, field := range f.fields {
		if field.validate() != nil {
			f.focus = index
			return
		}
	}
	f.result = make(FormResult, len(f.fields))
	for _, field := range f.fields {
		f.result[field.ID()] = field.Value()
	}
	f.submitted = true
}
func (f *Form) Frame(width int) *Frame {
	if f.submitted {
		return f.summaryFrame(width)
	}
	if f.cancelled {
		frame := NewFrame(max(width, 0), 1)
		drawClipped(frame, 0, 0, "Form cancelled", width, f.theme.Status.CancelledStyle)
		return frame
	}
	height := 0
	if f.title != "" {
		height = 2
	}
	for index, field := range f.fields {
		height += fieldHeight(field, index == f.focus, f.theme) + f.theme.FieldGap
	}
	height++
	frame := NewFrame(max(width, 0), height)
	y := 0
	if f.title != "" {
		drawClipped(frame, 0, y, f.title, width, f.theme.Accent.Bold(true))
		drawClipped(frame, displayWidth(f.title)+1, y, strings.Repeat(f.theme.Borders.Horizontal, max(width-displayWidth(f.title)-1, 0)), max(width-displayWidth(f.title)-1, 0), f.theme.Border)
		y += 2
	}
	for index, field := range f.fields {
		y += field.draw(frame, y, width, index == f.focus, f.theme)
		y += f.theme.FieldGap
	}
	drawKeyHints(frame, height-1, width, f.theme)
	return frame
}
func (f *Form) summaryFrame(width int) *Frame {
	frame := NewFrame(max(width, 0), len(f.fields)+3)
	drawBox(frame, 0, width, len(f.fields)+3, f.theme.Borders, f.theme.Success)
	drawClipped(frame, 2, 1, f.theme.Glyphs.Success+" "+f.title, max(width-4, 0), f.theme.Success.Bold(true))
	for i, field := range f.fields {
		drawClipped(frame, 2, i+2, field.labelSummary(), max(width/3-2, 0), f.theme.Muted)
		drawClipped(frame, max(width/3, 2), i+2, field.summary(), max(width-width/3-2, 0), f.theme.Text)
	}
	return frame
}

func fieldHeight(field FormField, focused bool, theme InlineTheme) int {
	scratch := NewFrame(1, 100)
	return field.draw(scratch, 0, 1, focused, theme)
}
func drawThemedLabel(frame *Frame, y, width int, label string, required, focused bool, theme InlineTheme) {
	style := theme.Label
	if focused {
		style = style.Bold(true)
	}
	drawClipped(frame, 0, y, label, width, style)
	if required {
		drawClipped(frame, displayWidth(label)+1, y, theme.Glyphs.Required, max(width-displayWidth(label)-1, 0), theme.Error)
	}
}
func (f *TextField) labelSummary() string        { return f.label }
func (f *SelectField) labelSummary() string      { return f.label }
func (f *MultiSelectField) labelSummary() string { return f.label }

func drawBox(frame *Frame, y, width, height int, border BorderSet, style tcell.Style) {
	if width < 2 || height < 2 {
		return
	}
	drawClipped(frame, 0, y, border.TopLeft, 1, style)
	drawClipped(frame, width-1, y, border.TopRight, 1, style)
	drawClipped(frame, 1, y, strings.Repeat(border.Horizontal, max(width-2, 0)), max(width-2, 0), style)
	for row := 1; row < height-1; row++ {
		drawClipped(frame, 0, y+row, border.Vertical, 1, style)
		drawClipped(frame, width-1, y+row, border.Vertical, 1, style)
	}
	drawClipped(frame, 0, y+height-1, border.BottomLeft, 1, style)
	drawClipped(frame, width-1, y+height-1, border.BottomRight, 1, style)
	drawClipped(frame, 1, y+height-1, strings.Repeat(border.Horizontal, max(width-2, 0)), max(width-2, 0), style)
}

func drawKeyHints(frame *Frame, y, width int, theme InlineTheme) {
	x := 0
	for _, hint := range []struct{ key, label string }{{"Tab", "Next"}, {"Shift+Tab", "Previous"}, {"Enter", "Submit"}, {"Esc", "Cancel"}} {
		piece := "[" + hint.key + "]"
		drawClipped(frame, x, y, piece, max(width-x, 0), theme.Accent.Bold(true))
		x += displayWidth(piece) + 1
		drawClipped(frame, x, y, hint.label, max(width-x, 0), theme.Muted)
		x += displayWidth(hint.label) + 2
		if x >= width {
			return
		}
	}
}
