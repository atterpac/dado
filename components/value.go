package components

// ValueProvider is implemented by components that hold a typed value.
// This interface provides a consistent API for value access across all form components.
type ValueProvider[T any] interface {
	// Value returns the component's primary value
	Value() T

	// SetValue sets the component's value
	SetValue(value T) ValueProvider[T]

	// HasValue returns true if the component has a non-zero/non-empty value
	HasValue() bool

	// Clear resets the value to its zero state
	Clear()
}

// IndexedValueProvider extends ValueProvider for components with selectable options.
// Components like Select and RadioGroup implement this interface.
type IndexedValueProvider[T any] interface {
	ValueProvider[T]

	// SelectedIndex returns the currently selected index (-1 if none)
	SelectedIndex() int

	// SetSelectedIndex selects by index, returns error if out of range
	SetSelectedIndex(index int) error
}

// MultiValueProvider is implemented by components that hold multiple values.
// Components like MultiSelect implement this interface.
type MultiValueProvider[T any] interface {
	// Values returns all selected values
	Values() []T

	// SetValues sets the selected values
	SetValues(values []T) error

	// HasValue returns true if at least one value is selected
	HasValue() bool

	// Clear deselects all values
	Clear()

	// SelectedIndices returns all selected indices (sorted)
	SelectedIndices() []int

	// SetSelectedIndices selects by indices
	SetSelectedIndices(indices []int) error
}

// Validatable is implemented by components that support validation.
type Validatable interface {
	// Validate runs validation and returns any error
	Validate() error

	// HasError returns whether the component has a validation error
	HasError() bool

	// GetError returns the current validation error message
	GetError() string
}

// Focusable is implemented by components that can receive focus.
type Focusable interface {
	// HasFocus returns whether the component currently has focus
	HasFocus() bool
}

// Named is implemented by components that have a name identifier.
type Named interface {
	// GetName returns the component's name
	GetName() string
}
