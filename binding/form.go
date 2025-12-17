package binding

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/atterpac/jig/components"
)

// FormBinding provides two-way binding between a struct and a Form.
// Use struct tags `form:"fieldname"` to map struct fields to form fields.
type FormBinding[T any] struct {
	form       *components.Form
	target     *T
	fieldMap   map[string]string // struct field name -> form field name
	validators map[string]func(any) error
	onChange   func(*T)
}

// NewFormBinding creates a binding between a struct and a form.
// The struct should use `form:"fieldname"` tags to map fields.
//
// Example:
//
//	type Config struct {
//	    Name    string `form:"name"`
//	    Enabled bool   `form:"enabled"`
//	}
//
//	config := &Config{}
//	fb := binding.NewFormBinding(form, config)
func NewFormBinding[T any](form *components.Form, target *T) *FormBinding[T] {
	fb := &FormBinding[T]{
		form:       form,
		target:     target,
		fieldMap:   make(map[string]string),
		validators: make(map[string]func(any) error),
	}
	fb.buildFieldMap()
	return fb
}

// SetOnChange sets callback when any bound value changes.
func (fb *FormBinding[T]) SetOnChange(fn func(*T)) *FormBinding[T] {
	fb.onChange = fn
	return fb
}

// SetValidator sets a validation function for a specific form field.
func (fb *FormBinding[T]) SetValidator(fieldName string, fn func(any) error) *FormBinding[T] {
	fb.validators[fieldName] = fn
	return fb
}

// buildFieldMap extracts form tags from the struct type.
func (fb *FormBinding[T]) buildFieldMap() {
	t := reflect.TypeOf(*fb.target)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("form"); tag != "" {
			fb.fieldMap[field.Name] = tag
		}
	}
}

// Sync synchronizes struct values to form fields.
// Call this after modifying the target struct programmatically.
func (fb *FormBinding[T]) Sync() error {
	v := reflect.ValueOf(fb.target).Elem()

	for structField, formField := range fb.fieldMap {
		fieldVal := v.FieldByName(structField)
		if !fieldVal.IsValid() {
			continue
		}

		formFieldObj := fb.form.GetField(formField)
		if formFieldObj == nil {
			continue
		}

		switch f := formFieldObj.(type) {
		case *components.TextField:
			f.SetValue(toString(fieldVal))
		case *components.TextArea:
			f.SetValue(toString(fieldVal))
		case *components.Checkbox:
			if fieldVal.Kind() == reflect.Bool {
				f.SetChecked(fieldVal.Bool())
			}
		case *components.Select:
			f.SetDefault(toString(fieldVal))
		case *components.RadioGroup:
			switch fieldVal.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				f.SetSelected(int(fieldVal.Int()))
			case reflect.String:
				// Find the option index by value
				for i, opt := range f.GetOptions() {
					if opt == fieldVal.String() {
						f.SetSelected(i)
						break
					}
				}
			}
		case *components.MultiSelect:
			if fieldVal.Kind() == reflect.Slice && fieldVal.Type().Elem().Kind() == reflect.String {
				values := make([]string, fieldVal.Len())
				for i := 0; i < fieldVal.Len(); i++ {
					values[i] = fieldVal.Index(i).String()
				}
				// Find indices for the values
				indices := make([]int, 0, len(values))
				for _, val := range values {
					for i, opt := range f.GetSelected() {
						if opt.Value == val {
							indices = append(indices, i)
							break
						}
					}
				}
				f.SetSelected(indices)
			}
		}
	}

	return nil
}

// Collect reads form values into the target struct.
// Returns validation errors if any.
func (fb *FormBinding[T]) Collect() error {
	v := reflect.ValueOf(fb.target).Elem()
	values := fb.form.GetValues()

	var errs []error

	for structField, formField := range fb.fieldMap {
		val, ok := values[formField]
		if !ok {
			continue
		}

		// Run custom validator if exists
		if validator, ok := fb.validators[formField]; ok {
			if err := validator(val); err != nil {
				errs = append(errs, err)
				continue
			}
		}

		fieldVal := v.FieldByName(structField)
		if !fieldVal.IsValid() || !fieldVal.CanSet() {
			continue
		}

		if err := setFieldValue(fieldVal, val); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// CollectAndNotify collects values and calls onChange if successful.
func (fb *FormBinding[T]) CollectAndNotify() error {
	if err := fb.Collect(); err != nil {
		return err
	}

	if fb.onChange != nil {
		fb.onChange(fb.target)
	}

	return nil
}

// Validate runs form validation and custom validators.
func (fb *FormBinding[T]) Validate() error {
	// First run form's built-in validation
	if err := fb.form.Validate(); err != nil {
		return err
	}

	// Then run custom validators
	values := fb.form.GetValues()
	var errs []error

	for formField, validator := range fb.validators {
		if val, ok := values[formField]; ok {
			if err := validator(val); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// Clear resets the form and target struct to zero values.
func (fb *FormBinding[T]) Clear() {
	fb.form.Clear()
	v := reflect.ValueOf(fb.target).Elem()
	v.Set(reflect.Zero(v.Type()))
}

// GetTarget returns a pointer to the bound struct.
func (fb *FormBinding[T]) GetTarget() *T {
	return fb.target
}

// SetTarget sets a new target struct and syncs to form.
func (fb *FormBinding[T]) SetTarget(target *T) {
	fb.target = target
	fb.Sync()
}

// Helper functions

func toString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		return ""
	}
}

func setFieldValue(field reflect.Value, val any) error {
	switch field.Kind() {
	case reflect.String:
		if s, ok := val.(string); ok {
			field.SetString(s)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := val.(type) {
		case int:
			field.SetInt(int64(v))
		case int64:
			field.SetInt(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				field.SetInt(i)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch v := val.(type) {
		case int:
			field.SetUint(uint64(v))
		case uint64:
			field.SetUint(v)
		case string:
			if i, err := strconv.ParseUint(v, 10, 64); err == nil {
				field.SetUint(i)
			}
		}
	case reflect.Float32, reflect.Float64:
		switch v := val.(type) {
		case float64:
			field.SetFloat(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				field.SetFloat(f)
			}
		}
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			field.SetBool(b)
		}
	case reflect.Slice:
		if ss, ok := val.([]string); ok && field.Type().Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(ss))
		}
	}
	return nil
}
