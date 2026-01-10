package testutil

import (
	"github.com/atterpac/jig/components"
)

// TextFieldBuilder provides fluent test setup for TextField.
type TextFieldBuilder struct {
	field *components.TextField
}

// NewTestTextField creates a new TextFieldBuilder for testing.
func NewTestTextField(name string) *TextFieldBuilder {
	return &TextFieldBuilder{
		field: components.NewTextField(name),
	}
}

// WithValue sets the field value.
func (b *TextFieldBuilder) WithValue(v string) *TextFieldBuilder {
	b.field.SetValue(v)
	return b
}

// WithLabel sets the field label.
func (b *TextFieldBuilder) WithLabel(l string) *TextFieldBuilder {
	b.field.SetLabel(l)
	return b
}

// WithPlaceholder sets the placeholder text.
func (b *TextFieldBuilder) WithPlaceholder(p string) *TextFieldBuilder {
	b.field.SetPlaceholder(p)
	return b
}

// WithValidator sets a validation function.
func (b *TextFieldBuilder) WithValidator(fn func(string) error) *TextFieldBuilder {
	b.field.SetValidator(fn)
	return b
}

// WithOnChange sets a change handler.
func (b *TextFieldBuilder) WithOnChange(fn components.ChangeHandler[string]) *TextFieldBuilder {
	b.field.SetOnChange(fn)
	return b
}

// WithOnSubmit sets a submit handler.
func (b *TextFieldBuilder) WithOnSubmit(fn components.SubmitHandler) *TextFieldBuilder {
	b.field.SetOnSubmit(fn)
	return b
}

// Build returns the configured TextField.
func (b *TextFieldBuilder) Build() *components.TextField {
	return b.field
}

// CheckboxBuilder provides fluent test setup for Checkbox.
type CheckboxBuilder struct {
	checkbox *components.Checkbox
}

// NewTestCheckbox creates a new CheckboxBuilder for testing.
func NewTestCheckbox(name string) *CheckboxBuilder {
	return &CheckboxBuilder{
		checkbox: components.NewCheckbox(name),
	}
}

// WithLabel sets the checkbox label.
func (b *CheckboxBuilder) WithLabel(l string) *CheckboxBuilder {
	b.checkbox.SetLabel(l)
	return b
}

// Checked sets the checked state.
func (b *CheckboxBuilder) Checked(c bool) *CheckboxBuilder {
	b.checkbox.SetChecked(c)
	return b
}

// WithOnChange sets a change handler.
func (b *CheckboxBuilder) WithOnChange(fn components.ChangeHandler[bool]) *CheckboxBuilder {
	b.checkbox.SetOnChange(fn)
	return b
}

// Build returns the configured Checkbox.
func (b *CheckboxBuilder) Build() *components.Checkbox {
	return b.checkbox
}

// SelectBuilder provides fluent test setup for Select.
type SelectBuilder struct {
	sel *components.Select
}

// NewTestSelect creates a new SelectBuilder for testing.
func NewTestSelect(name string) *SelectBuilder {
	return &SelectBuilder{
		sel: components.NewSelect(name),
	}
}

// WithLabel sets the select label.
func (b *SelectBuilder) WithLabel(l string) *SelectBuilder {
	b.sel.SetLabel(l)
	return b
}

// WithPlaceholder sets the placeholder text.
func (b *SelectBuilder) WithPlaceholder(p string) *SelectBuilder {
	b.sel.SetPlaceholder(p)
	return b
}

// WithOptions sets the available options.
func (b *SelectBuilder) WithOptions(opts ...string) *SelectBuilder {
	b.sel.SetOptions(opts)
	return b
}

// WithOptionsAndValues sets options with custom label/value pairs.
func (b *SelectBuilder) WithOptionsAndValues(opts []components.SelectOption) *SelectBuilder {
	b.sel.SetOptionsWithValues(opts)
	return b
}

// WithDefault sets the default selected value.
func (b *SelectBuilder) WithDefault(v string) *SelectBuilder {
	b.sel.SetDefault(v)
	return b
}

// WithSelectedIndex sets the selected index.
func (b *SelectBuilder) WithSelectedIndex(idx int) *SelectBuilder {
	b.sel.SetSelected(idx)
	return b
}

// WithOnChange sets a change handler.
func (b *SelectBuilder) WithOnChange(fn components.ChangeHandler[components.SelectOption]) *SelectBuilder {
	b.sel.SetOnChange(fn)
	return b
}

// Build returns the configured Select.
func (b *SelectBuilder) Build() *components.Select {
	return b.sel
}

// RadioGroupBuilder provides fluent test setup for RadioGroup.
type RadioGroupBuilder struct {
	rg *components.RadioGroup
}

// NewTestRadioGroup creates a new RadioGroupBuilder for testing.
func NewTestRadioGroup(name string) *RadioGroupBuilder {
	return &RadioGroupBuilder{
		rg: components.NewRadioGroup(name),
	}
}

// WithLabel sets the group label.
func (b *RadioGroupBuilder) WithLabel(l string) *RadioGroupBuilder {
	b.rg.SetLabel(l)
	return b
}

// WithOptions sets the available options.
func (b *RadioGroupBuilder) WithOptions(opts ...string) *RadioGroupBuilder {
	b.rg.SetOptions(opts)
	return b
}

// WithSelected sets the selected index.
func (b *RadioGroupBuilder) WithSelected(idx int) *RadioGroupBuilder {
	b.rg.SetSelected(idx)
	return b
}

// WithOnChange sets a change handler.
func (b *RadioGroupBuilder) WithOnChange(fn components.ChangeHandler[string]) *RadioGroupBuilder {
	b.rg.SetOnChange(fn)
	return b
}

// Build returns the configured RadioGroup.
func (b *RadioGroupBuilder) Build() *components.RadioGroup {
	return b.rg
}

// MultiSelectBuilder provides fluent test setup for MultiSelect.
type MultiSelectBuilder struct {
	ms *components.MultiSelect
}

// NewTestMultiSelect creates a new MultiSelectBuilder for testing.
func NewTestMultiSelect(name string) *MultiSelectBuilder {
	return &MultiSelectBuilder{
		ms: components.NewMultiSelect(name),
	}
}

// WithLabel sets the multi-select label.
func (b *MultiSelectBuilder) WithLabel(l string) *MultiSelectBuilder {
	b.ms.SetLabel(l)
	return b
}

// WithOptions sets the available options.
func (b *MultiSelectBuilder) WithOptions(opts ...string) *MultiSelectBuilder {
	b.ms.SetOptions(opts)
	return b
}

// WithSelected sets the selected indices.
func (b *MultiSelectBuilder) WithSelected(indices ...int) *MultiSelectBuilder {
	b.ms.SetSelected(indices)
	return b
}

// WithOnChange sets a change handler.
func (b *MultiSelectBuilder) WithOnChange(fn components.ChangeHandler[[]components.SelectOption]) *MultiSelectBuilder {
	b.ms.SetOnChange(fn)
	return b
}

// Build returns the configured MultiSelect.
func (b *MultiSelectBuilder) Build() *components.MultiSelect {
	return b.ms
}

// FormBuilder provides fluent test setup for Form.
type FormBuilder struct {
	form *components.Form
}

// NewTestForm creates a new FormBuilder for testing.
func NewTestForm() *FormBuilder {
	return &FormBuilder{
		form: components.NewForm(),
	}
}

// WithTextField adds a text field to the form.
func (b *FormBuilder) WithTextField(name, label, placeholder string) *FormBuilder {
	b.form.AddTextField(name, label, placeholder)
	return b
}

// WithCheckbox adds a checkbox to the form.
func (b *FormBuilder) WithCheckbox(name, label string) *FormBuilder {
	b.form.AddCheckbox(name, label)
	return b
}

// WithSelect adds a select field to the form.
func (b *FormBuilder) WithSelect(name, label string, options []string) *FormBuilder {
	b.form.AddSelect(name, label, options)
	return b
}

// WithRadioGroup adds a radio group to the form.
func (b *FormBuilder) WithRadioGroup(name, label string, options []string) *FormBuilder {
	b.form.AddRadioGroup(name, label, options)
	return b
}

// WithField adds a custom field to the form.
func (b *FormBuilder) WithField(field components.FormField) *FormBuilder {
	b.form.AddField(field)
	return b
}

// WithOnSubmit sets the submit handler.
func (b *FormBuilder) WithOnSubmit(fn func(map[string]any)) *FormBuilder {
	b.form.SetOnSubmit(fn)
	return b
}

// WithOnCancel sets the cancel handler.
func (b *FormBuilder) WithOnCancel(fn func()) *FormBuilder {
	b.form.SetOnCancel(fn)
	return b
}

// Build returns the configured Form.
func (b *FormBuilder) Build() *components.Form {
	return b.form
}

// ComponentBaseBuilder provides fluent test setup for ComponentBase.
type ComponentBaseBuilder struct {
	base *components.ComponentBase
}

// NewTestComponentBase creates a new ComponentBaseBuilder for testing.
func NewTestComponentBase() *ComponentBaseBuilder {
	return &ComponentBaseBuilder{
		base: components.NewComponentBase(components.NewPanel()),
	}
}

// WithName sets the component name.
func (b *ComponentBaseBuilder) WithName(name string) *ComponentBaseBuilder {
	b.base.SetName(name)
	return b
}

// WithHints sets the key hints.
func (b *ComponentBaseBuilder) WithHints(hints []components.KeyHint) *ComponentBaseBuilder {
	b.base.SetHints(hints)
	return b
}

// AddHint adds a single key hint.
func (b *ComponentBaseBuilder) AddHint(key, description string) *ComponentBaseBuilder {
	b.base.AddHint(key, description)
	return b
}

// WithOnStart sets the start callback.
func (b *ComponentBaseBuilder) WithOnStart(fn func()) *ComponentBaseBuilder {
	b.base.SetOnStart(fn)
	return b
}

// WithOnStop sets the stop callback.
func (b *ComponentBaseBuilder) WithOnStop(fn func()) *ComponentBaseBuilder {
	b.base.SetOnStop(fn)
	return b
}

// Build returns the configured ComponentBase.
func (b *ComponentBaseBuilder) Build() *components.ComponentBase {
	return b.base
}
