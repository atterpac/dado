package components

import (
	"fmt"
	"sync"

	"github.com/atterpac/dado/validators"
)

// FormFieldFactory creates a new FormField instance.
type FormFieldFactory func(name string) FormField

var (
	customFieldTypes   = make(map[string]FormFieldFactory)
	customFieldTypesMu sync.RWMutex
)

// RegisterFormFieldType registers a custom field type that can be used in FormBuilder.
// Example:
//
//	components.RegisterFormFieldType("date", func(name string) FormField {
//	    return NewDatePicker(name)
//	})
func RegisterFormFieldType(typeName string, factory FormFieldFactory) {
	customFieldTypesMu.Lock()
	defer customFieldTypesMu.Unlock()
	customFieldTypes[typeName] = factory
}

// UnregisterFormFieldType removes a custom field type registration.
func UnregisterFormFieldType(typeName string) {
	customFieldTypesMu.Lock()
	defer customFieldTypesMu.Unlock()
	delete(customFieldTypes, typeName)
}

func getFieldFactory(typeName string) (FormFieldFactory, bool) {
	customFieldTypesMu.RLock()
	defer customFieldTypesMu.RUnlock()
	factory, ok := customFieldTypes[typeName]
	return factory, ok
}

// =============================================================================
// FormBuilder
// =============================================================================

// FormBuilder provides a fluent API for building forms.
type FormBuilder struct {
	form     *Form
	fields   []FormField
	onSubmit func(values map[string]any)
	onCancel func()

	// For tracking current field being built
	currentField FormField
}

// NewFormBuilder creates a new FormBuilder.
func NewFormBuilder() *FormBuilder {
	return &FormBuilder{
		form: NewForm(),
	}
}

// =============================================================================
// Field Sub-Builders
// =============================================================================

// TextFieldBuilder builds a TextField with fluent API.
type TextFieldBuilder struct {
	parent     *FormBuilder
	field      *TextField
	validators []validators.Validator
}

// Text adds a text field and returns a builder to configure it.
func (fb *FormBuilder) Text(name, label string) *TextFieldBuilder {
	field := NewTextField(name).SetLabel(label)
	return &TextFieldBuilder{
		parent: fb,
		field:  field,
	}
}

// Password adds a password field (masked text input) and returns a builder to configure it.
func (fb *FormBuilder) Password(name, label string) *TextFieldBuilder {
	field := NewTextField(name).SetLabel(label).SetMasked(true)
	return &TextFieldBuilder{
		parent: fb,
		field:  field,
	}
}

// Placeholder sets the placeholder text.
func (tb *TextFieldBuilder) Placeholder(text string) *TextFieldBuilder {
	tb.field.SetPlaceholder(text)
	return tb
}

// Value sets the default value.
func (tb *TextFieldBuilder) Value(value string) *TextFieldBuilder {
	tb.field.SetValue(value)
	return tb
}

// Validate adds validators to the field.
func (tb *TextFieldBuilder) Validate(v ...validators.Validator) *TextFieldBuilder {
	tb.validators = append(tb.validators, v...)
	return tb
}

// OnChange sets the change handler.
func (tb *TextFieldBuilder) OnChange(handler ChangeHandler[string]) *TextFieldBuilder {
	tb.field.SetOnChange(handler)
	return tb
}

// OnSubmit sets the submit handler.
func (tb *TextFieldBuilder) OnSubmit(handler SubmitHandler) *TextFieldBuilder {
	tb.field.SetOnSubmit(handler)
	return tb
}

// Masked enables password masking for the field.
func (tb *TextFieldBuilder) Masked() *TextFieldBuilder {
	tb.field.SetMasked(true)
	return tb
}

// MaskChar sets a custom mask character (default is '•').
func (tb *TextFieldBuilder) MaskChar(char rune) *TextFieldBuilder {
	tb.field.SetMaskChar(char)
	return tb
}

// Done finalizes the field and returns to the FormBuilder.
func (tb *TextFieldBuilder) Done() *FormBuilder {
	// Set combined validator
	if len(tb.validators) > 0 {
		tb.field.SetValidator(func(value string) error {
			for _, v := range tb.validators {
				if err := v(value); err != nil {
					return err
				}
			}
			return nil
		})
	}
	tb.parent.fields = append(tb.parent.fields, tb.field)
	return tb.parent
}

// =============================================================================
// TextAreaBuilder
// =============================================================================

// TextAreaBuilder builds a TextArea with fluent API.
type TextAreaBuilder struct {
	parent *FormBuilder
	field  *TextArea
}

// TextArea adds a text area field and returns a builder to configure it.
func (fb *FormBuilder) TextArea(name, label string) *TextAreaBuilder {
	field := NewTextArea(name).SetLabel(label)
	return &TextAreaBuilder{
		parent: fb,
		field:  field,
	}
}

// Placeholder sets the placeholder text.
func (tb *TextAreaBuilder) Placeholder(text string) *TextAreaBuilder {
	tb.field.SetPlaceholder(text)
	return tb
}

// Value sets the default value.
func (tb *TextAreaBuilder) Value(value string) *TextAreaBuilder {
	tb.field.SetValue(value)
	return tb
}

// MaxLines sets the maximum number of lines.
func (tb *TextAreaBuilder) MaxLines(n int) *TextAreaBuilder {
	tb.field.SetMaxLines(n)
	return tb
}

// OnChange sets the change handler.
func (tb *TextAreaBuilder) OnChange(handler ChangeHandler[string]) *TextAreaBuilder {
	tb.field.SetOnChange(handler)
	return tb
}

// Done finalizes the field and returns to the FormBuilder.
func (tb *TextAreaBuilder) Done() *FormBuilder {
	tb.parent.fields = append(tb.parent.fields, tb.field)
	return tb.parent
}

// =============================================================================
// SelectBuilder
// =============================================================================

// SelectBuilder builds a Select with fluent API.
type SelectBuilder struct {
	parent     *FormBuilder
	field      *Select
	validators []validators.Validator
}

// Select adds a select dropdown and returns a builder to configure it.
func (fb *FormBuilder) Select(name, label string, options []string) *SelectBuilder {
	field := NewSelect(name).SetLabel(label).SetOptions(options)
	return &SelectBuilder{
		parent: fb,
		field:  field,
	}
}

// SelectWithValues adds a select dropdown with custom label/value pairs.
func (fb *FormBuilder) SelectWithValues(name, label string, options []SelectOption) *SelectBuilder {
	field := NewSelect(name).SetLabel(label).SetOptionsWithValues(options)
	return &SelectBuilder{
		parent: fb,
		field:  field,
	}
}

// Placeholder sets the placeholder text.
func (sb *SelectBuilder) Placeholder(text string) *SelectBuilder {
	sb.field.SetPlaceholder(text)
	return sb
}

// Default sets the default selected value.
func (sb *SelectBuilder) Default(value string) *SelectBuilder {
	sb.field.SetDefault(value)
	return sb
}

// Selected sets the default selected index.
func (sb *SelectBuilder) Selected(index int) *SelectBuilder {
	sb.field.SetSelected(index)
	return sb
}

// Validate adds validators to the field.
func (sb *SelectBuilder) Validate(v ...validators.Validator) *SelectBuilder {
	sb.validators = append(sb.validators, v...)
	return sb
}

// OnChange sets the change handler.
func (sb *SelectBuilder) OnChange(handler ChangeHandler[SelectOption]) *SelectBuilder {
	sb.field.SetOnChange(handler)
	return sb
}

// Done finalizes the field and returns to the FormBuilder.
func (sb *SelectBuilder) Done() *FormBuilder {
	sb.parent.fields = append(sb.parent.fields, sb.field)
	return sb.parent
}

// =============================================================================
// MultiSelectBuilder
// =============================================================================

// MultiSelectBuilder builds a MultiSelect with fluent API.
type MultiSelectBuilder struct {
	parent     *FormBuilder
	field      *MultiSelect
	validators []validators.Validator
}

// MultiSelect adds a multi-select field and returns a builder to configure it.
func (fb *FormBuilder) MultiSelect(name, label string, options []string) *MultiSelectBuilder {
	field := NewMultiSelect(name).SetLabel(label).SetOptions(options)
	return &MultiSelectBuilder{
		parent: fb,
		field:  field,
	}
}

// MultiSelectWithValues adds a multi-select with custom label/value pairs.
func (fb *FormBuilder) MultiSelectWithValues(name, label string, options []SelectOption) *MultiSelectBuilder {
	field := NewMultiSelect(name).SetLabel(label).SetOptionsWithValues(options)
	return &MultiSelectBuilder{
		parent: fb,
		field:  field,
	}
}

// Selected sets the initially selected indices.
func (mb *MultiSelectBuilder) Selected(indices []int) *MultiSelectBuilder {
	mb.field.SetSelected(indices)
	return mb
}

// Validate adds validators to the field.
func (mb *MultiSelectBuilder) Validate(v ...validators.Validator) *MultiSelectBuilder {
	mb.validators = append(mb.validators, v...)
	return mb
}

// OnChange sets the change handler.
func (mb *MultiSelectBuilder) OnChange(handler ChangeHandler[[]SelectOption]) *MultiSelectBuilder {
	mb.field.SetOnChange(handler)
	return mb
}

// Done finalizes the field and returns to the FormBuilder.
func (mb *MultiSelectBuilder) Done() *FormBuilder {
	mb.parent.fields = append(mb.parent.fields, mb.field)
	return mb.parent
}

// =============================================================================
// CheckboxBuilder
// =============================================================================

// CheckboxBuilder builds a Checkbox with fluent API.
type CheckboxBuilder struct {
	parent *FormBuilder
	field  *Checkbox
}

// Checkbox adds a checkbox field and returns a builder to configure it.
func (fb *FormBuilder) Checkbox(name, label string) *CheckboxBuilder {
	field := NewCheckbox(name).SetLabel(label)
	return &CheckboxBuilder{
		parent: fb,
		field:  field,
	}
}

// Checked sets the initial checked state.
func (cb *CheckboxBuilder) Checked(checked bool) *CheckboxBuilder {
	cb.field.SetChecked(checked)
	return cb
}

// OnChange sets the change handler.
func (cb *CheckboxBuilder) OnChange(handler ChangeHandler[bool]) *CheckboxBuilder {
	cb.field.SetOnChange(handler)
	return cb
}

// Done finalizes the field and returns to the FormBuilder.
func (cb *CheckboxBuilder) Done() *FormBuilder {
	cb.parent.fields = append(cb.parent.fields, cb.field)
	return cb.parent
}

// =============================================================================
// RadioGroupBuilder
// =============================================================================

// RadioGroupBuilder builds a RadioGroup with fluent API.
type RadioGroupBuilder struct {
	parent     *FormBuilder
	field      *RadioGroup
	validators []validators.Validator
}

// Radio adds a radio group and returns a builder to configure it.
func (fb *FormBuilder) Radio(name, label string, options []string) *RadioGroupBuilder {
	field := NewRadioGroup(name).SetLabel(label).SetOptions(options)
	return &RadioGroupBuilder{
		parent: fb,
		field:  field,
	}
}

// Selected sets the initially selected index.
func (rb *RadioGroupBuilder) Selected(index int) *RadioGroupBuilder {
	rb.field.SetSelected(index)
	return rb
}

// Validate adds validators to the field.
func (rb *RadioGroupBuilder) Validate(v ...validators.Validator) *RadioGroupBuilder {
	rb.validators = append(rb.validators, v...)
	return rb
}

// OnChange sets the change handler.
func (rb *RadioGroupBuilder) OnChange(handler ChangeHandler[string]) *RadioGroupBuilder {
	rb.field.SetOnChange(handler)
	return rb
}

// Done finalizes the field and returns to the FormBuilder.
func (rb *RadioGroupBuilder) Done() *FormBuilder {
	rb.parent.fields = append(rb.parent.fields, rb.field)
	return rb.parent
}

// =============================================================================
// CustomFieldBuilder
// =============================================================================

// CustomFieldBuilder builds a custom field type.
type CustomFieldBuilder struct {
	parent *FormBuilder
	field  FormField
}

// Custom adds a custom field type by its registered name.
// The field type must be registered with RegisterFormFieldType first.
func (fb *FormBuilder) Custom(name, typeName, label string) *CustomFieldBuilder {
	factory, ok := getFieldFactory(typeName)
	if !ok {
		panic(fmt.Sprintf("unknown form field type: %s", typeName))
	}
	field := factory(name)
	return &CustomFieldBuilder{
		parent: fb,
		field:  field,
	}
}

// Configure allows custom configuration of the field.
func (cfb *CustomFieldBuilder) Configure(fn func(field FormField)) *CustomFieldBuilder {
	fn(cfb.field)
	return cfb
}

// Done finalizes the field and returns to the FormBuilder.
func (cfb *CustomFieldBuilder) Done() *FormBuilder {
	cfb.parent.fields = append(cfb.parent.fields, cfb.field)
	return cfb.parent
}

// =============================================================================
// FormBuilder Methods
// =============================================================================

// AddField adds a pre-built field directly to the form.
func (fb *FormBuilder) AddField(field FormField) *FormBuilder {
	fb.fields = append(fb.fields, field)
	return fb
}

// OnSubmit sets the form submission callback.
func (fb *FormBuilder) OnSubmit(fn func(values map[string]any)) *FormBuilder {
	fb.onSubmit = fn
	return fb
}

// OnCancel sets the form cancellation callback.
func (fb *FormBuilder) OnCancel(fn func()) *FormBuilder {
	fb.onCancel = fn
	return fb
}

// Build creates the final Form with all configured fields.
func (fb *FormBuilder) Build() *Form {
	form := NewForm()
	for _, field := range fb.fields {
		form.AddField(field)
	}
	if fb.onSubmit != nil {
		form.SetOnSubmit(fb.onSubmit)
	}
	if fb.onCancel != nil {
		form.SetOnCancel(fb.onCancel)
	}
	return form
}

// =============================================================================
// Modal Integration
// =============================================================================

// AsModal wraps the form in a modal and returns it.
func (fb *FormBuilder) AsModal(title string, width, height int) *Modal {
	form := fb.Build()

	modal := NewModal(ModalConfig{
		Title:    title,
		Width:    width,
		Height:   height,
		MinWidth: 30,
		Backdrop: true,
	}).SetContent(form)

	// Wire up form callbacks to modal
	originalSubmit := fb.onSubmit
	form.SetOnSubmit(func(values map[string]any) {
		if originalSubmit != nil {
			originalSubmit(values)
		}
	})

	originalCancel := fb.onCancel
	form.SetOnCancel(func() {
		if originalCancel != nil {
			originalCancel()
		}
	})

	return modal
}

// AsConfirmModal creates a modal that dismisses on Escape.
func (fb *FormBuilder) AsConfirmModal(title string, width, height int) *Modal {
	modal := fb.AsModal(title, width, height)
	modal.SetDismissOnEsc(true)
	return modal
}

// AsFormModal creates a modal that does NOT dismiss on Escape (to prevent data loss).
func (fb *FormBuilder) AsFormModal(title string, width, height int) *Modal {
	modal := fb.AsModal(title, width, height)
	modal.SetDismissOnEsc(false)
	return modal
}

// =============================================================================
// Validation Result Types
// =============================================================================

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult contains all validation errors from a form.
type ValidationResult struct {
	Errors []FieldError
}

// HasErrors returns true if there are any validation errors.
func (vr ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// Error returns the first error message, or empty string if no errors.
func (vr ValidationResult) Error() string {
	if len(vr.Errors) == 0 {
		return ""
	}
	return vr.Errors[0].Error()
}

// ErrorFor returns the error message for a specific field, or empty string.
func (vr ValidationResult) ErrorFor(fieldName string) string {
	for _, e := range vr.Errors {
		if e.Field == fieldName {
			return e.Message
		}
	}
	return ""
}
