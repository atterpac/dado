package inline

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	// ErrFormCancelled is returned when the user cancels an interactive form.
	ErrFormCancelled = errors.New("inline: form cancelled")
	// ErrFormInterrupted is returned when the user requests that the enclosing
	// interactive application exit instead of navigating back from the form. It
	// unwraps to ErrFormCancelled so callers can handle all aborted forms alike.
	ErrFormInterrupted error = formInterruptedError{}
	// ErrFormPrevious is returned when boundary navigation is enabled and the
	// user presses Shift+Tab on the first field.
	ErrFormPrevious = errors.New("inline: navigate to previous form")
	// ErrFormHasNoFields is returned when a form has nothing to edit.
	ErrFormHasNoFields = errors.New("inline: form has no fields")
	// ErrNilFormField is returned when a form contains a nil field.
	ErrNilFormField = errors.New("inline: nil form field")
	// ErrInvalidFieldID is returned when a field ID is blank.
	ErrInvalidFieldID = errors.New("inline: form field ID cannot be empty")
	// ErrDuplicateField is returned when multiple fields use the same ID.
	ErrDuplicateField = errors.New("inline: duplicate form field ID")
)

type formInterruptedError struct{}

func (formInterruptedError) Error() string { return "inline: form interrupted" }
func (formInterruptedError) Unwrap() error { return ErrFormCancelled }

// ChoiceTone applies a semantic theme color to a choice.
type ChoiceTone int

const (
	ChoiceToneDefault ChoiceTone = iota
	ChoiceToneSuccess
	ChoiceToneDanger
	ChoiceToneMuted
)

// Choice is one selectable value in a SelectField or MultiSelectField.
type Choice struct {
	Value       string
	Label       string
	Description string
	Disabled    bool
	Tone        ChoiceTone
}

// NewChoice creates an enabled choice.
func NewChoice(value, label string) Choice { return Choice{Value: value, Label: label} }

// WithTone applies a semantic theme color to the choice.
func (c Choice) WithTone(tone ChoiceTone) Choice { c.Tone = tone; return c }

func choiceStyle(choice Choice, theme InlineTheme) tcell.Style {
	if choice.Disabled {
		return theme.Muted
	}
	switch choice.Tone {
	case ChoiceToneSuccess:
		return theme.Success
	case ChoiceToneDanger:
		return theme.Error
	case ChoiceToneMuted:
		return theme.Muted
	default:
		return theme.Text
	}
}

// choiceViewport keeps a cursor-positioned window over a list of choices. The
// positions passed to it are relative to the current list, which lets Select
// reuse it after filtering while MultiSelect uses choice indices directly.
type choiceViewport struct {
	maximum int
	offset  int
}

func (v *choiceViewport) setMaximum(maximum, cursorPosition, total int) {
	v.maximum = max(maximum, 0)
	v.keepVisible(cursorPosition, total)
}

func (v *choiceViewport) keepVisible(cursorPosition, total int) {
	if v.maximum <= 0 || total <= v.maximum {
		v.offset = 0
		return
	}
	limit := min(v.maximum, total)
	v.offset = min(max(v.offset, 0), total-limit)
	if cursorPosition < 0 || cursorPosition >= total {
		return
	}
	if cursorPosition < v.offset {
		v.offset = cursorPosition
	} else if cursorPosition >= v.offset+limit {
		v.offset = cursorPosition - limit + 1
	}
}

func (v *choiceViewport) bounds(total int) (start, end int, clipped bool) {
	if total <= 0 {
		v.offset = 0
		return 0, 0, false
	}
	if v.maximum <= 0 || total <= v.maximum {
		v.offset = 0
		return 0, total, false
	}
	limit := min(v.maximum, total)
	v.offset = min(max(v.offset, 0), total-limit)
	return v.offset, v.offset + limit, true
}

func (v *choiceViewport) layout(total int) (start, end, rows int, clipped bool) {
	start, end, clipped = v.bounds(total)
	rows = max(end-start, 1)
	if clipped {
		rows++
	}
	return
}

func choicePosition(indices []int, choiceIndex int) int {
	for position, index := range indices {
		if index == choiceIndex {
			return position
		}
	}
	return -1
}

func choiceRange(start, end, total int) string {
	return fmt.Sprintf("%d–%d of %d", start+1, end, total)
}

func drawChoiceRange(frame *Frame, y, width int, text string, style tcell.Style) {
	available := max(width-4, 0)
	x := max(2, width-2-displayWidth(text))
	drawClipped(frame, x, y, text, min(displayWidth(text), available), style)
}

// FormResult contains typed values keyed by field ID.
type FormResult map[string]any

// FormKeyHint describes one key and action in a Form footer.
type FormKeyHint struct {
	Key   string
	Label string
}

// FormField is a field that can participate in a Form. The private methods
// intentionally keep the first version of the field contract closed.
type FormField interface {
	ID() string
	Value() any
	draw(*Frame, int, int, bool, InlineTheme) int
	height(bool) int
	handle(*tcell.EventKey)
	validate() error
	summary() string
	labelSummary() string
}

// TextField edits one line of Unicode text.
type TextField struct {
	id, label, placeholder string
	value                  []string
	cursor                 int
	required, password     bool
	validateFn             func(string) error
	err                    error
}

func NewTextField(id, label string) *TextField { return &TextField{id: id, label: label} }
func (f *TextField) ID() string                { return f.id }
func (f *TextField) Value() any                { return strings.Join(f.value, "") }
func (f *TextField) SetValue(value string) *TextField {
	f.value = splitGraphemes(value)
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
		f.insertRune(event.Rune())
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if f.cursor > 0 {
			f.value = append(f.value[:f.cursor-1], f.value[f.cursor:]...)
			f.cursor--
		}
	case tcell.KeyDelete:
		if f.cursor < len(f.value) {
			f.value = append(f.value[:f.cursor], f.value[f.cursor+1:]...)
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

func (f *TextField) insertRune(value rune) {
	before := strings.Join(f.value[:f.cursor], "")
	after := strings.Join(f.value[f.cursor:], "")
	prefix := before + string(value)
	f.value = splitGraphemes(prefix + after)
	f.cursor = len(splitGraphemes(prefix))
}

func (f *TextField) validate() error {
	value := strings.Join(f.value, "")
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
	displayGraphemes := f.value
	if f.password {
		displayGraphemes = splitGraphemes(strings.Repeat("•", len(f.value)))
	}
	visibleStart := 0
	available := max(width-4, 0)
	for visibleStart < f.cursor && graphemesWidth(displayGraphemes[visibleStart:f.cursor]) >= available && available > 0 {
		visibleStart++
	}
	display := strings.Join(displayGraphemes[visibleStart:], "")
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
		cursorWidth := graphemesWidth(displayGraphemes[visibleStart:f.cursor])
		frame.ShowCursor(min(2+cursorWidth, max(width-2, 0)), y+2)
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
	return strings.Join(f.value, "")
}

func (f *TextField) height(_ bool) int {
	if f.err != nil {
		return 5
	}
	return 4
}

// SelectField chooses one value.
type SelectField struct {
	id, label  string
	choices    []Choice
	cursor     int
	selected   int
	viewport   choiceViewport
	required   bool
	filterable bool
	query      []string
	err        error
}

func NewSelectField(id, label string, choices ...Choice) *SelectField {
	f := &SelectField{id: id, label: label, choices: append([]Choice(nil), choices...), selected: -1}
	f.cursor = f.nextEnabled(-1, 1)
	return f
}
func (f *SelectField) ID() string             { return f.id }
func (f *SelectField) Required() *SelectField { f.required = true; return f }

// MaxVisibleChoices bounds the number of choices rendered while the field is
// focused. Values less than one retain the unbounded behavior.
func (f *SelectField) MaxVisibleChoices(maximum int) *SelectField {
	visible := f.visibleChoices()
	f.viewport.setMaximum(maximum, choicePosition(visible, f.cursor), len(visible))
	return f
}

// Filterable enables incremental, case-insensitive filtering by choice value,
// label, and description. Typing updates the query and Backspace removes it.
func (f *SelectField) Filterable(enabled bool) *SelectField {
	f.filterable = enabled
	if !enabled {
		f.query = nil
	}
	f.ensureVisibleCursor()
	return f
}
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
		f.keepCursorVisible()
	case tcell.KeyDown, tcell.KeyRight:
		f.cursor = f.nextEnabled(f.cursor, 1)
		f.keepCursorVisible()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if f.filterable && len(f.query) > 0 {
			f.query = f.query[:len(f.query)-1]
			f.ensureVisibleCursor()
		}
	case tcell.KeyRune:
		if f.filterable {
			f.query = append(f.query, string(event.Rune()))
			f.ensureVisibleCursor()
			return
		}
		switch event.Rune() {
		case 'h', 'k':
			f.cursor = f.nextEnabled(f.cursor, -1)
			f.keepCursorVisible()
			return
		case 'j', 'l':
			f.cursor = f.nextEnabled(f.cursor, 1)
			f.keepCursorVisible()
			return
		case ' ':
		default:
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
	visible := f.visibleChoices()
	if len(visible) == 0 {
		return -1
	}
	position := -1
	for index, choiceIndex := range visible {
		if choiceIndex == from {
			position = index
			break
		}
	}
	for offset := 1; offset <= len(visible); offset++ {
		index := (position + direction*offset + len(visible)*2) % len(visible)
		choiceIndex := visible[index]
		if !f.choices[choiceIndex].Disabled {
			return choiceIndex
		}
	}
	return -1
}
func (f *SelectField) visibleChoices() []int {
	query := strings.ToLower(strings.TrimSpace(strings.Join(f.query, "")))
	visible := make([]int, 0, len(f.choices))
	for index, choice := range f.choices {
		if query == "" || strings.Contains(strings.ToLower(choice.Value+" "+choice.Label+" "+choice.Description), query) {
			visible = append(visible, index)
		}
	}
	return visible
}
func (f *SelectField) ensureVisibleCursor() {
	visible := f.visibleChoices()
	for position, index := range visible {
		if index == f.cursor && !f.choices[index].Disabled {
			f.viewport.keepVisible(position, len(visible))
			return
		}
	}
	f.cursor = f.nextEnabled(-1, 1)
	f.viewport.keepVisible(choicePosition(visible, f.cursor), len(visible))
}
func (f *SelectField) keepCursorVisible() {
	visible := f.visibleChoices()
	f.viewport.keepVisible(choicePosition(visible, f.cursor), len(visible))
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
	visible := f.visibleChoices()
	start, end, clipped := 0, len(visible), false
	if focused {
		f.keepCursorVisible()
		start, end, rows, clipped = f.viewport.layout(len(visible))
		if f.filterable {
			rows++
		}
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
		choiceY := y + 2
		if f.filterable {
			query := strings.Join(f.query, "")
			if query == "" {
				query = "Type to filter…"
			}
			drawClipped(frame, 2, choiceY, "/ "+query, max(width-4, 0), theme.Muted)
			choiceY++
		}
		if len(visible) == 0 {
			drawClipped(frame, 2, choiceY, "No matching options", max(width-4, 0), theme.Muted)
		}
		for row, index := range visible[start:end] {
			choice := f.choices[index]
			marker := theme.Glyphs.Unselected
			if index == f.selected {
				marker = theme.Glyphs.Selected
			}
			style := choiceStyle(choice, theme)
			if index == f.cursor {
				drawClipped(frame, 0, choiceY+row, theme.Glyphs.Focus, 1, theme.Accent)
				style = style.Bold(true)
			}
			chip := choice.Label
			if marker != "" {
				chip = marker + " " + chip
			}
			drawClipped(frame, 2, choiceY+row, chip, max(width-4, 0), style)
			if choice.Description != "" {
				start := 3 + displayWidth(chip)
				drawClipped(frame, start, choiceY+row, choice.Description, max(width-start-2, 0), theme.Muted)
			}
		}
		if clipped {
			rangeY := choiceY + max(end-start, 1)
			drawChoiceRange(frame, rangeY, width, choiceRange(start, end, len(visible)), theme.Muted)
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

func (f *SelectField) height(focused bool) int {
	rows := 1
	if focused {
		visible := f.visibleChoices()
		_, _, rows, _ = f.viewport.layout(len(visible))
		if f.filterable {
			rows++
		}
	}
	height := rows + 3
	if f.err != nil {
		height++
	}
	return height
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
	viewport         choiceViewport
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

// MaxVisibleChoices bounds the number of choices rendered while the field is
// focused. Values less than one retain the unbounded behavior.
func (f *MultiSelectField) MaxVisibleChoices(maximum int) *MultiSelectField {
	f.viewport.setMaximum(maximum, f.cursor, len(f.choices))
	return f
}

// SetSelectedValues replaces the current selection with choices matching the
// supplied values. Unknown values are ignored.
func (f *MultiSelectField) SetSelectedValues(values ...string) *MultiSelectField {
	wanted := make(map[string]bool, len(values))
	for _, value := range values {
		wanted[value] = true
	}
	f.selected = make(map[int]bool)
	for index, choice := range f.choices {
		if wanted[choice.Value] {
			f.selected[index] = true
		}
	}
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
	return nextEnabledChoice(f.choices, from, direction)
}

func nextEnabledChoice(choices []Choice, from, direction int) int {
	if len(choices) == 0 {
		return -1
	}
	for offset := 1; offset <= len(choices); offset++ {
		index := (from + direction*offset + len(choices)*2) % len(choices)
		if !choices[index].Disabled {
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
		f.viewport.keepVisible(f.cursor, len(f.choices))
	case tcell.KeyDown, tcell.KeyRight:
		f.cursor = f.nextEnabled(f.cursor, 1)
		f.viewport.keepVisible(f.cursor, len(f.choices))
	case tcell.KeyRune:
		switch event.Rune() {
		case 'h', 'k':
			f.cursor = f.nextEnabled(f.cursor, -1)
			f.viewport.keepVisible(f.cursor, len(f.choices))
		case 'j', 'l':
			f.cursor = f.nextEnabled(f.cursor, 1)
			f.viewport.keepVisible(f.cursor, len(f.choices))
		case ' ':
			if f.cursor < 0 {
				return
			}
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
	start, end, clipped := 0, len(f.choices), false
	if focused {
		f.viewport.keepVisible(f.cursor, len(f.choices))
		start, end, rows, clipped = f.viewport.layout(len(f.choices))
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
		for row, choice := range f.choices[start:end] {
			index := start + row
			selected := f.selected[index]
			marker := f.indicatorMarker(selected, choice.Disabled, theme)
			style := choiceStyle(choice, theme)
			if !choice.Disabled && f.indicator == IndicatorNone && selected {
				style = style.Reverse(true)
			}
			if index == f.cursor {
				drawClipped(frame, 0, y+2+row, theme.Glyphs.Focus, 1, theme.Accent)
				style = style.Bold(true)
			}
			chip := choice.Label
			if marker != "" {
				chip = marker + " " + chip
			}
			drawClipped(frame, 2, y+2+row, chip, max(width-4, 0), style)
		}
		if clipped {
			rangeY := y + 2 + max(end-start, 1)
			drawChoiceRange(frame, rangeY, width, choiceRange(start, end, len(f.choices)), theme.Muted)
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

func (f *MultiSelectField) height(focused bool) int {
	rows := 1
	if focused {
		_, _, rows, _ = f.viewport.layout(len(f.choices))
	}
	height := rows + 3
	if f.err != nil {
		height++
	}
	return height
}

// Form composes fields, focus traversal, validation, submission, and summary rendering.
type Form struct {
	title                                       string
	fields                                      []FormField
	focus                                       int
	focusChanged                                func(int, FormField)
	submitted, cancelled, interrupted, previous bool
	quitOnQ, navigateAtBoundaries               bool
	keyHintOverride                             []FormKeyHint
	header                                      FrameProvider
	headerGap                                   int
	result                                      FormResult
	theme                                       InlineTheme
}

func NewForm(title string) *Form { return &Form{title: title, theme: RoundedInlineTheme()} }
func (f *Form) Add(fields ...FormField) *Form {
	wasEmpty := len(f.fields) == 0
	f.fields = append(f.fields, fields...)
	if wasEmpty {
		f.notifyFocusChange()
	}
	return f
}
func (f *Form) SetTheme(theme InlineTheme) *Form { f.theme = normalizedInlineTheme(theme); return f }

// SetHeader renders another inline component above the form. A nil provider
// removes the header.
func (f *Form) SetHeader(header FrameProvider) *Form { f.header = header; return f }

// SetHeaderGap controls the blank rows between the header and form.
func (f *Form) SetHeaderGap(rows int) *Form { f.headerGap = max(rows, 0); return f }

// OnFocusChange calls callback with the newly focused field. The current field
// is reported immediately when callback is set on a non-empty form.
func (f *Form) OnFocusChange(callback func(int, FormField)) *Form {
	f.focusChanged = callback
	f.notifyFocusChange()
	return f
}

// QuitOnQ makes q interrupt the form when the focused field is not accepting
// text. Text fields and filterable selects continue to receive q as input.
func (f *Form) QuitOnQ(enabled bool) *Form { f.quitOnQ = enabled; return f }

// NavigateAtBoundaries makes Tab submit from the last field and Shift+Tab exit
// from the first field with ErrFormPrevious. Disabled forms retain wrapping
// Tab navigation.
func (f *Form) NavigateAtBoundaries(enabled bool) *Form {
	f.navigateAtBoundaries = enabled
	return f
}

// SetKeyHints replaces the contextual footer hints. Calling it with no hints
// restores the automatically generated hints.
func (f *Form) SetKeyHints(hints ...FormKeyHint) *Form {
	f.keyHintOverride = append([]FormKeyHint(nil), hints...)
	return f
}

func (f *Form) validateStructure() error {
	if len(f.fields) == 0 {
		return ErrFormHasNoFields
	}
	ids := make(map[string]struct{}, len(f.fields))
	for _, field := range f.fields {
		if nilFormField(field) {
			return ErrNilFormField
		}
		id := strings.TrimSpace(field.ID())
		if id == "" {
			return ErrInvalidFieldID
		}
		if _, exists := ids[id]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateField, id)
		}
		ids[id] = struct{}{}
	}
	return nil
}

func nilFormField(field FormField) bool {
	if field == nil {
		return true
	}
	switch value := field.(type) {
	case *TextField:
		return value == nil
	case *SelectField:
		return value == nil
	case *MultiSelectField:
		return value == nil
	default:
		return false
	}
}
func (f *Form) Result() FormResult {
	result := make(FormResult, len(f.result))
	for k, v := range f.result {
		result[k] = v
	}
	return result
}
func (f *Form) Done() bool { return f.submitted || f.cancelled || f.previous }
func (f *Form) HandleKey(event *tcell.EventKey) {
	if f.Done() || event == nil || f.validateStructure() != nil {
		return
	}
	if event.Key() == tcell.KeyTab && event.Modifiers()&tcell.ModShift != 0 {
		f.handlePreviousField()
		return
	}
	switch event.Key() {
	case tcell.KeyCtrlC:
		f.cancelled = true
		f.interrupted = true
		return
	case tcell.KeyEscape:
		f.cancelled = true
		return
	case tcell.KeyRune:
		if f.quitOnQ && event.Rune() == 'q' && !formFieldAcceptsText(f.fields[f.focus]) {
			f.cancelled = true
			f.interrupted = true
			return
		}
	case tcell.KeyTab:
		if f.navigateAtBoundaries && f.focus == len(f.fields)-1 {
			f.commitFocusedSelect()
			f.submit()
			return
		}
		f.setFocus((f.focus + 1) % len(f.fields))
		return
	case tcell.KeyBacktab:
		f.handlePreviousField()
		return
	case tcell.KeyEnter:
		f.commitFocusedSelect()
		if f.focus < len(f.fields)-1 {
			f.setFocus(f.focus + 1)
			return
		}
		f.submit()
		return
	}
	f.fields[f.focus].handle(event)
}

func (f *Form) commitFocusedSelect() {
	if field, ok := f.fields[f.focus].(*SelectField); ok {
		field.handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	}
}

func (f *Form) handlePreviousField() {
	if f.navigateAtBoundaries && f.focus == 0 {
		f.previous = true
		return
	}
	f.setFocus((f.focus - 1 + len(f.fields)) % len(f.fields))
}

func (f *Form) setFocus(index int) {
	if index == f.focus {
		return
	}
	f.focus = index
	f.notifyFocusChange()
}

func (f *Form) notifyFocusChange() {
	if f.focusChanged == nil || f.focus < 0 || f.focus >= len(f.fields) || nilFormField(f.fields[f.focus]) {
		return
	}
	f.focusChanged(f.focus, f.fields[f.focus])
}

func formFieldAcceptsText(field FormField) bool {
	switch typed := field.(type) {
	case *TextField:
		return true
	case *SelectField:
		return typed.filterable
	default:
		return false
	}
}
func (f *Form) submit() {
	for index, field := range f.fields {
		if field.validate() != nil {
			f.setFocus(index)
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
	body := f.formFrame(width)
	if f.header == nil {
		return body
	}
	return StackFrames(f.headerGap, f.header.Frame(width), body)
}

func (f *Form) formFrame(width int) *Frame {
	if err := f.validateStructure(); err != nil {
		frame := NewFrame(max(width, 0), 1)
		drawClipped(frame, 0, 0, err.Error(), width, f.theme.Error)
		return frame
	}
	if f.submitted {
		return f.summaryFrame(width)
	}
	if f.previous {
		frame := NewFrame(max(width, 0), 1)
		drawClipped(frame, 0, 0, "Returning to previous form", width, f.theme.Muted)
		return frame
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
		height += field.height(index == f.focus) + f.theme.FieldGap
	}
	hints := f.keyHints()
	hintRows := keyHintsHeight(width, hints)
	height += hintRows
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
	drawKeyHints(frame, height-hintRows, width, f.theme, hints)
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

func (f *Form) keyHints() []FormKeyHint {
	if len(f.keyHintOverride) > 0 {
		return append([]FormKeyHint(nil), f.keyHintOverride...)
	}
	hints := make([]FormKeyHint, 0, 7)
	switch f.fields[f.focus].(type) {
	case *SelectField:
		hints = append(hints, FormKeyHint{"↑/↓", "Move"})
	case *MultiSelectField:
		hints = append(hints, FormKeyHint{"↑/↓", "Move"}, FormKeyHint{"Space", "Toggle"})
	}
	if f.navigateAtBoundaries {
		hints = append(hints, FormKeyHint{"Tab", "Next"}, FormKeyHint{"Shift+Tab", "Previous"})
	} else if len(f.fields) > 1 {
		hints = append(hints, FormKeyHint{"Tab", "Next field"}, FormKeyHint{"Shift+Tab", "Previous field"})
	}
	if f.focus == len(f.fields)-1 {
		hints = append(hints, FormKeyHint{"Enter", "Submit"})
	} else {
		hints = append(hints, FormKeyHint{"Enter", "Next field"})
	}
	hints = append(hints, FormKeyHint{"Esc", "Cancel"})
	return hints
}

func keyHintsHeight(width int, hints []FormKeyHint) int {
	if width <= 0 || len(hints) == 0 {
		return 0
	}
	rows, x := 1, 0
	for _, hint := range hints {
		hintWidth := displayWidth("["+hint.Key+"]") + 1 + displayWidth(hint.Label)
		if x > 0 && x+hintWidth > width {
			rows++
			x = 0
		}
		x += hintWidth + 2
	}
	return rows
}

func drawKeyHints(frame *Frame, y, width int, theme InlineTheme, hints []FormKeyHint) {
	x := 0
	for _, hint := range hints {
		piece := "[" + hint.Key + "]"
		hintWidth := displayWidth(piece) + 1 + displayWidth(hint.Label)
		if x > 0 && x+hintWidth > width {
			y++
			x = 0
		}
		drawClipped(frame, x, y, piece, max(width-x, 0), theme.Accent.Bold(true))
		x += displayWidth(piece) + 1
		drawClipped(frame, x, y, hint.Label, max(width-x, 0), theme.Muted)
		x += displayWidth(hint.Label) + 2
	}
}
