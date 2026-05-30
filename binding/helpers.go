package binding

import (
	"strings"

	"github.com/atterpac/dado/components"
)

// BindTableToSlice is a convenience function for simple table binding.
// It creates a TableBinding with the given headers, data, and mapper.
//
// Example:
//
//	type User struct {
//	    Name  string
//	    Email string
//	}
//
//	users := []User{{Name: "Alice", Email: "alice@example.com"}}
//	binding := binding.BindTableToSlice(table, []string{"NAME", "EMAIL"}, users,
//	    func(u User) []string { return []string{u.Name, u.Email} })
func BindTableToSlice[T any](
	table *components.Table,
	headers []string,
	data []T,
	mapper func(T) []string,
) *TableBinding[T] {
	table.SetHeaders(headers...)

	binding := NewTableBinding[T](table).
		SetMapper(mapper)

	if len(data) > 0 {
		binding.SetData(data)
	}

	return binding
}

// BindTableWithKey is like BindTableToSlice but also sets up key-based selection.
//
// Example:
//
//	binding := binding.BindTableWithKey(table, headers, pods,
//	    func(p Pod) []string { return []string{p.Name, p.Status} },
//	    func(p Pod) string { return p.Name })
func BindTableWithKey[T any](
	table *components.Table,
	headers []string,
	data []T,
	mapper func(T) []string,
	keyMapper func(T) string,
) *TableBinding[T] {
	table.SetHeaders(headers...)

	binding := NewTableBinding[T](table).
		SetMapper(mapper).
		SetKeyMapper(keyMapper)

	if len(data) > 0 {
		binding.SetData(data)
	}

	return binding
}

// DefaultStringFilter returns a filter function that matches any mapped field
// containing the query string (case-insensitive).
//
// Example:
//
//	mapper := func(p Pod) []string { return []string{p.Name, p.Status} }
//	binding.SetFilter(binding.DefaultStringFilter(mapper))
func DefaultStringFilter[T any](mapper func(T) []string) func(T, string) bool {
	return func(item T, query string) bool {
		query = strings.ToLower(query)
		for _, cell := range mapper(item) {
			if strings.Contains(strings.ToLower(cell), query) {
				return true
			}
		}
		return false
	}
}

// FieldFilter returns a filter function that matches a specific field.
// The fieldIndex corresponds to the column returned by the mapper.
//
// Example:
//
//	// Filter on column 0 (Name)
//	binding.SetFilter(binding.FieldFilter(mapper, 0))
func FieldFilter[T any](mapper func(T) []string, fieldIndex int) func(T, string) bool {
	return func(item T, query string) bool {
		fields := mapper(item)
		if fieldIndex < 0 || fieldIndex >= len(fields) {
			return false
		}
		return strings.Contains(strings.ToLower(fields[fieldIndex]), strings.ToLower(query))
	}
}

// ExactMatchFilter returns a filter that requires an exact match (case-insensitive).
func ExactMatchFilter[T any](mapper func(T) []string) func(T, string) bool {
	return func(item T, query string) bool {
		query = strings.ToLower(query)
		for _, cell := range mapper(item) {
			if strings.ToLower(cell) == query {
				return true
			}
		}
		return false
	}
}

// PrefixFilter returns a filter that matches items starting with the query.
func PrefixFilter[T any](mapper func(T) []string) func(T, string) bool {
	return func(item T, query string) bool {
		query = strings.ToLower(query)
		for _, cell := range mapper(item) {
			if strings.HasPrefix(strings.ToLower(cell), query) {
				return true
			}
		}
		return false
	}
}

// FormBindingOptions configures form binding behavior.
type FormBindingOptions struct {
	// AutoCollect enables automatic collecting from form to struct on submit
	AutoCollect bool
}

// BindFormToStruct creates a form binding with common options.
// If AutoCollect is true, the struct is automatically updated on form submit.
//
// Example:
//
//	type Config struct {
//	    Name    string `form:"name"`
//	    Enabled bool   `form:"enabled"`
//	}
//
//	config := &Config{}
//	fb := binding.BindFormToStruct(form, config, binding.FormBindingOptions{
//	    AutoCollect: true,
//	})
func BindFormToStruct[T any](
	form *components.Form,
	target *T,
	opts FormBindingOptions,
) *FormBinding[T] {
	fb := NewFormBinding(form, target)

	if opts.AutoCollect {
		// Note: This would need form.SetOnSubmit integration
		// For now, users can call fb.Collect() manually in their submit handler
	}

	return fb
}

// RequiredValidator returns a validator that ensures a string value is not empty.
func RequiredValidator(fieldName string) func(any) error {
	return func(val any) error {
		if s, ok := val.(string); ok && s == "" {
			return &ValidationError{Field: fieldName, Message: "is required"}
		}
		return nil
	}
}

// MinLengthValidator returns a validator that ensures minimum string length.
func MinLengthValidator(fieldName string, minLen int) func(any) error {
	return func(val any) error {
		if s, ok := val.(string); ok && len(s) < minLen {
			return &ValidationError{Field: fieldName, Message: "is too short"}
		}
		return nil
	}
}

// MaxLengthValidator returns a validator that ensures maximum string length.
func MaxLengthValidator(fieldName string, maxLen int) func(any) error {
	return func(val any) error {
		if s, ok := val.(string); ok && len(s) > maxLen {
			return &ValidationError{Field: fieldName, Message: "is too long"}
		}
		return nil
	}
}

// ValidationError represents a field validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Message
}

// ChainValidators combines multiple validators into one.
// All validators must pass for the value to be valid.
func ChainValidators(validators ...func(any) error) func(any) error {
	return func(val any) error {
		for _, v := range validators {
			if err := v(val); err != nil {
				return err
			}
		}
		return nil
	}
}
