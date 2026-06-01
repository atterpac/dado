// Package components provides UI primitives for building TUI applications.
//
// # Component Categories
//
// Input components for user data entry:
//   - [TextField] - Single-line text input with placeholder and validation
//   - [TextArea] - Multi-line text input
//   - [Checkbox] - Boolean toggle with label
//   - [Select] - Dropdown selection
//   - [MultiSelect] - Multi-option selection
//   - [RadioGroup] - Single selection from visible options
//
// Container components for layout:
//   - [Panel] - Bordered container with rounded corners and title
//   - [Modal] - Centered dialog overlay with configurable behavior
//   - [Tabs] - Tabbed container for organizing content
//   - [Split] - Resizable split pane container
//
// Display components for data presentation:
//   - [Table] - Data table with headers, selection, and styling
//   - [Tree] - Hierarchical tree view
//   - [VirtualList] - Virtualized list for large datasets
//   - [Progress] - Progress indicator with percentage
//
// # Creating Components
//
// All components use factory functions and fluent configuration:
//
//	field := NewTextField("email").
//	    SetLabel("Email Address").
//	    SetPlaceholder("user@example.com").
//	    SetValidator(validators.Email())
//
// # Component Lifecycle
//
// Components that are pushed to nav.Pages have Start/Stop lifecycle:
//
//	func (v *MyView) Start() {
//	    // Called when view becomes active
//	    // - Load initial data
//	    // - Start timers/pollers
//	}
//
//	func (v *MyView) Stop() {
//	    // Called when view becomes inactive
//	    // - Cancel pending operations
//	    // - Stop timers
//	}
//
// # Using ComponentBase
//
// ComponentBase wraps any core.Widget to implement nav.Component:
//
//	type MyView struct {
//	    *ComponentBase
//	    table *Table
//	}
//
//	func NewMyView() *MyView {
//	    table := NewTable()
//	    v := &MyView{table: table}
//	    v.ComponentBase = NewComponentBase(table).
//	        SetName("my-view").
//	        AddHint("Enter", "Select").
//	        SetOnStart(v.loadData)
//	    return v
//	}
//
// # Form Builder
//
// Use FormBuilder for fluent form construction with validation:
//
//	form := NewFormBuilder().
//	    Text("name", "Name").
//	        Validate(validators.Required()).
//	        Done().
//	    Text("email", "Email").
//	        Validate(validators.Email()).
//	        Done().
//	    OnSubmit(func(values map[string]any) {
//	        // Handle submission
//	    }).
//	    Build()
//
// # Theme Integration
//
// Components automatically integrate with the theme system. Colors are read
// at draw time for live theme switching. Register primitives for automatic
// background updates:
//
//	theme.Register(myBox)  // Auto-updates background on theme change
//
// # Value Interfaces
//
// Input components implement consistent value access:
//
//	type ValueProvider[T any] interface {
//	    Value() T
//	    SetValue(T) ValueProvider[T]
//	    HasValue() bool
//	    Clear()
//	}
//
// Components with validation implement:
//
//	type Validatable interface {
//	    Validate() error
//	    HasError() bool
//	    GetError() string
//	}
package components
