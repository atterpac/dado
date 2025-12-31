package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// =============================================================================
// SetValues Error Types
// =============================================================================

// SetValueError represents an error setting a specific field's value.
type SetValueError struct {
	Field string
	Err   error
}

func (e SetValueError) Error() string {
	return fmt.Sprintf("field %q: %v", e.Field, e.Err)
}

// SetValuesError collects all errors from SetValues.
type SetValuesError struct {
	Errors []SetValueError
}

func (e SetValuesError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("failed to set %d field(s): %s", len(e.Errors), strings.Join(msgs, "; "))
}

// FormField is the interface for form fields.
type FormField interface {
	tview.Primitive
	GetName() string
	GetFieldHeight() int
}

// Form is a container for form fields with focus management.
type Form struct {
	*tview.Box

	fields       []FormField
	focusedIndex int
	offset       int // scroll offset

	// Callbacks
	onSubmit func(values map[string]any)
	onCancel func()
}

// NewForm creates a new Form container.
func NewForm() *Form {
	return &Form{
		Box: tview.NewBox(),
	}
}

// AddField adds a field to the form.
func (f *Form) AddField(field FormField) *Form {
	f.fields = append(f.fields, field)
	return f
}

// AddTextField adds a text field to the form.
func (f *Form) AddTextField(name, label, placeholder string) *Form {
	field := NewTextField(name).
		SetLabel(label).
		SetPlaceholder(placeholder)
	return f.AddField(field)
}

// AddSelect adds a select field to the form.
func (f *Form) AddSelect(name, label string, options []string) *Form {
	field := NewSelect(name).
		SetLabel(label).
		SetOptions(options)
	return f.AddField(field)
}

// AddCheckbox adds a checkbox to the form.
func (f *Form) AddCheckbox(name, label string) *Form {
	field := NewCheckbox(name).SetLabel(label)
	return f.AddField(field)
}

// AddRadioGroup adds a radio group to the form.
func (f *Form) AddRadioGroup(name, label string, options []string) *Form {
	field := NewRadioGroup(name).
		SetLabel(label).
		SetOptions(options)
	return f.AddField(field)
}

// SetOnSubmit sets the callback for form submission.
func (f *Form) SetOnSubmit(fn func(values map[string]any)) *Form {
	f.onSubmit = fn
	return f
}

// SetOnCancel sets the callback for form cancellation.
func (f *Form) SetOnCancel(fn func()) *Form {
	f.onCancel = fn
	return f
}

// GetField returns a field by name.
func (f *Form) GetField(name string) FormField {
	for _, field := range f.fields {
		if field.GetName() == name {
			return field
		}
	}
	return nil
}

// =============================================================================
// Type-Safe Field Accessors
// =============================================================================

// GetTextField returns a TextField by name.
// Returns nil, false if not found or if the field is not a TextField.
func (f *Form) GetTextField(name string) (*TextField, bool) {
	field := f.GetField(name)
	if field == nil {
		return nil, false
	}
	tf, ok := field.(*TextField)
	return tf, ok
}

// GetTextArea returns a TextArea by name.
// Returns nil, false if not found or if the field is not a TextArea.
func (f *Form) GetTextArea(name string) (*TextArea, bool) {
	field := f.GetField(name)
	if field == nil {
		return nil, false
	}
	ta, ok := field.(*TextArea)
	return ta, ok
}

// GetSelect returns a Select by name.
// Returns nil, false if not found or if the field is not a Select.
func (f *Form) GetSelect(name string) (*Select, bool) {
	field := f.GetField(name)
	if field == nil {
		return nil, false
	}
	s, ok := field.(*Select)
	return s, ok
}

// GetMultiSelect returns a MultiSelect by name.
// Returns nil, false if not found or if the field is not a MultiSelect.
func (f *Form) GetMultiSelect(name string) (*MultiSelect, bool) {
	field := f.GetField(name)
	if field == nil {
		return nil, false
	}
	ms, ok := field.(*MultiSelect)
	return ms, ok
}

// GetCheckbox returns a Checkbox by name.
// Returns nil, false if not found or if the field is not a Checkbox.
func (f *Form) GetCheckbox(name string) (*Checkbox, bool) {
	field := f.GetField(name)
	if field == nil {
		return nil, false
	}
	cb, ok := field.(*Checkbox)
	return cb, ok
}

// GetRadioGroup returns a RadioGroup by name.
// Returns nil, false if not found or if the field is not a RadioGroup.
func (f *Form) GetRadioGroup(name string) (*RadioGroup, bool) {
	field := f.GetField(name)
	if field == nil {
		return nil, false
	}
	rg, ok := field.(*RadioGroup)
	return rg, ok
}

// GetFormField is a generic helper for type-safe field access.
// Returns the typed field and true if found and type matches.
//
// Example:
//
//	tf, ok := components.GetFormField[*TextField](form, "username")
//	if ok {
//	    value := tf.GetValue()
//	}
func GetFormField[T FormField](f *Form, name string) (T, bool) {
	field := f.GetField(name)
	if field == nil {
		var zero T
		return zero, false
	}
	typed, ok := field.(T)
	return typed, ok
}

// =============================================================================
// Bulk Value Setting
// =============================================================================

// SetValues sets multiple field values from a map.
// Keys are field names, values must match the expected type for each field:
//   - TextField, TextArea, Select, RadioGroup: string
//   - Checkbox: bool
//   - MultiSelect: []string or []int (for indices)
//
// Returns nil if all values were set successfully.
// Returns SetValuesError containing details of any failures.
func (f *Form) SetValues(values map[string]any) error {
	var errs []SetValueError

	for name, value := range values {
		field := f.GetField(name)
		if field == nil {
			errs = append(errs, SetValueError{
				Field: name,
				Err:   fmt.Errorf("field not found"),
			})
			continue
		}

		if err := f.setFieldValue(field, value); err != nil {
			errs = append(errs, SetValueError{
				Field: name,
				Err:   err,
			})
		}
	}

	if len(errs) > 0 {
		return SetValuesError{Errors: errs}
	}
	return nil
}

// setFieldValue sets a single field's value with type checking.
func (f *Form) setFieldValue(field FormField, value any) error {
	switch fld := field.(type) {
	case *TextField:
		if v, ok := value.(string); ok {
			fld.SetValue(v)
			return nil
		}
		return fmt.Errorf("expected string, got %T", value)

	case *TextArea:
		if v, ok := value.(string); ok {
			fld.SetValue(v)
			return nil
		}
		return fmt.Errorf("expected string, got %T", value)

	case *Select:
		switch v := value.(type) {
		case string:
			fld.SetDefault(v)
			return nil
		case int:
			fld.SetSelected(v)
			return nil
		default:
			return fmt.Errorf("expected string or int, got %T", value)
		}

	case *MultiSelect:
		switch v := value.(type) {
		case []string:
			// Find indices for the given values
			indices := make([]int, 0, len(v))
			for _, val := range v {
				for i, opt := range fld.options {
					if opt.Value == val {
						indices = append(indices, i)
						break
					}
				}
			}
			fld.SetSelected(indices)
			return nil
		case []int:
			fld.SetSelected(v)
			return nil
		default:
			return fmt.Errorf("expected []string or []int, got %T", value)
		}

	case *Checkbox:
		if v, ok := value.(bool); ok {
			fld.SetChecked(v)
			return nil
		}
		return fmt.Errorf("expected bool, got %T", value)

	case *RadioGroup:
		switch v := value.(type) {
		case string:
			for i, opt := range fld.options {
				if opt == v {
					fld.SetSelected(i)
					return nil
				}
			}
			return fmt.Errorf("option %q not found", v)
		case int:
			fld.SetSelected(v)
			return nil
		default:
			return fmt.Errorf("expected string or int, got %T", value)
		}

	default:
		return fmt.Errorf("unsupported field type %T", field)
	}
}

// GetValues returns all field values as a map.
func (f *Form) GetValues() map[string]any {
	values := make(map[string]any)
	for _, field := range f.fields {
		switch fld := field.(type) {
		case *TextField:
			values[fld.GetName()] = fld.GetValue()
		case *TextArea:
			values[fld.GetName()] = fld.GetValue()
		case *Select:
			values[fld.GetName()] = fld.GetValue()
		case *MultiSelect:
			values[fld.GetName()] = fld.GetValues()
		case *Checkbox:
			values[fld.GetName()] = fld.GetValue()
		case *RadioGroup:
			values[fld.GetName()] = fld.GetValue()
		}
	}
	return values
}

// Validate validates all fields and returns the first error.
func (f *Form) Validate() error {
	for _, field := range f.fields {
		if tf, ok := field.(*TextField); ok {
			if err := tf.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateAll validates all fields and returns all errors.
// This is useful for showing multiple validation errors at once.
func (f *Form) ValidateAll() ValidationResult {
	var errors []FieldError

	for _, field := range f.fields {
		switch fld := field.(type) {
		case *TextField:
			if err := fld.Validate(); err != nil {
				errors = append(errors, FieldError{
					Field:   fld.GetName(),
					Message: err.Error(),
				})
			}
		}
	}

	return ValidationResult{Errors: errors}
}

// IsValid returns true if all fields pass validation.
func (f *Form) IsValid() bool {
	return !f.ValidateAll().HasErrors()
}

// Clear resets all fields to their default values.
func (f *Form) Clear() *Form {
	for _, field := range f.fields {
		switch fld := field.(type) {
		case *TextField:
			fld.SetValue("")
		case *TextArea:
			fld.SetValue("")
		case *Select:
			fld.SetSelected(-1)
		case *MultiSelect:
			fld.SetSelected(nil)
		case *Checkbox:
			fld.SetChecked(false)
		case *RadioGroup:
			fld.SetSelected(-1)
		}
	}
	return f
}

// FocusIndex focuses the field at the given index.
func (f *Form) FocusIndex(index int) *Form {
	if index >= 0 && index < len(f.fields) {
		f.focusedIndex = index
	}
	return f
}

// Draw renders the form.
func (f *Form) Draw(screen tcell.Screen) {
	f.Box.DrawForSubclass(screen, f)
	x, y, width, height := f.GetInnerRect()

	if width <= 0 || height <= 0 || len(f.fields) == 0 {
		return
	}

	// Calculate total height needed
	totalHeight := 0
	fieldPositions := make([]int, len(f.fields))
	for i, field := range f.fields {
		fieldPositions[i] = totalHeight
		totalHeight += field.GetFieldHeight() + 1 // +1 for spacing
	}

	// Adjust offset to keep focused field visible
	focusedPos := fieldPositions[f.focusedIndex]
	focusedHeight := f.fields[f.focusedIndex].GetFieldHeight()

	if focusedPos < f.offset {
		f.offset = focusedPos
	}
	if focusedPos+focusedHeight > f.offset+height {
		f.offset = focusedPos + focusedHeight - height
	}
	if f.offset < 0 {
		f.offset = 0
	}

	// Draw fields
	currentY := y - f.offset
	for i, field := range f.fields {
		fieldHeight := field.GetFieldHeight()

		// Skip if completely above viewport
		if currentY+fieldHeight <= y {
			currentY += fieldHeight + 1
			continue
		}

		// Stop if below viewport
		if currentY >= y+height {
			break
		}

		// Set field rect and draw
		field.SetRect(x, currentY, width, fieldHeight)
		field.Draw(screen)

		currentY += fieldHeight + 1
		_ = i
	}
}

// InputHandler handles keyboard input.
func (f *Form) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return f.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if len(f.fields) == 0 {
			return
		}

		// Check for form-level keys
		switch event.Key() {
		case tcell.KeyTab:
			f.focusNext()
			return
		case tcell.KeyBacktab:
			f.focusPrev()
			return
		case tcell.KeyEscape:
			if f.onCancel != nil {
				f.onCancel()
			}
			return
		case tcell.KeyCtrlS:
			if f.onSubmit != nil {
				f.onSubmit(f.GetValues())
			}
			return
		}

		// Pass to focused field
		if f.focusedIndex >= 0 && f.focusedIndex < len(f.fields) {
			field := f.fields[f.focusedIndex]
			if handler := field.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

func (f *Form) focusNext() {
	if len(f.fields) == 0 {
		return
	}

	// Blur current
	if f.focusedIndex >= 0 && f.focusedIndex < len(f.fields) {
		f.fields[f.focusedIndex].Blur()
	}

	f.focusedIndex++
	if f.focusedIndex >= len(f.fields) {
		f.focusedIndex = 0
	}

	// Visually focus the new field (no-op delegate so tview focus stays on Form)
	f.fields[f.focusedIndex].Focus(func(tview.Primitive) {})
}

func (f *Form) focusPrev() {
	if len(f.fields) == 0 {
		return
	}

	// Blur current
	if f.focusedIndex >= 0 && f.focusedIndex < len(f.fields) {
		f.fields[f.focusedIndex].Blur()
	}

	f.focusedIndex--
	if f.focusedIndex < 0 {
		f.focusedIndex = len(f.fields) - 1
	}

	// Visually focus the new field (no-op delegate so tview focus stays on Form)
	f.fields[f.focusedIndex].Focus(func(tview.Primitive) {})
}

// Focus handles focus.
// The Form keeps focus itself so it can intercept Tab/BackTab for field navigation.
// Individual fields receive input via the Form's InputHandler delegation.
// Note: We do NOT call delegate - this keeps tview focus on the Form.
func (f *Form) Focus(delegate func(tview.Primitive)) {
	// Don't call delegate(f) - that would cause infinite recursion
	// tview will keep focus on the Form since we don't delegate elsewhere

	// Visually focus the current field (no-op delegate for visual state only)
	if len(f.fields) > 0 && f.focusedIndex >= 0 && f.focusedIndex < len(f.fields) {
		f.fields[f.focusedIndex].Focus(func(tview.Primitive) {})
	}
}

// Blur handles blur.
func (f *Form) Blur() {
	for _, field := range f.fields {
		field.Blur()
	}
	f.Box.Blur()
}

// HasFocus returns whether any field has focus.
func (f *Form) HasFocus() bool {
	for _, field := range f.fields {
		if field.HasFocus() {
			return true
		}
	}
	return false
}

// MouseHandler handles mouse input.
func (f *Form) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return f.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		if !f.InRect(event.Position()) {
			return false, nil
		}

		// Find which field was clicked
		_, my := event.Position()
		x, y, width, _ := f.GetInnerRect()

		currentY := y - f.offset
		for i, field := range f.fields {
			fieldHeight := field.GetFieldHeight()
			if my >= currentY && my < currentY+fieldHeight {
				f.focusedIndex = i
				field.SetRect(x, currentY, width, fieldHeight)
				if handler := field.MouseHandler(); handler != nil {
					return handler(action, event, setFocus)
				}
				return true, field
			}
			currentY += fieldHeight + 1
		}

		return false, nil
	})
}
