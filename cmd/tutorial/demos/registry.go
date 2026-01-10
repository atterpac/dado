package demos

import (
	"github.com/rivo/tview"
)

// Category represents the complexity level of a demo
type Category int

const (
	Basic Category = iota
	Intermediate
	Advanced
)

func (c Category) String() string {
	switch c {
	case Basic:
		return "Basic"
	case Intermediate:
		return "Intermediate"
	case Advanced:
		return "Advanced"
	default:
		return "Unknown"
	}
}

// PropertyType defines the type of a property for the editor
type PropertyType int

const (
	PropString PropertyType = iota
	PropBool
	PropInt
	PropSelect
	PropColor
)

// Property represents an editable component property
type Property struct {
	Name        string
	Type        PropertyType
	Value       any
	Options     []string // For PropSelect type
	Description string
}

// PropertyDescriptor defines a property with getter/setter functions for automatic binding.
// This eliminates the need for switch statements in ApplyProperty and Reset methods.
type PropertyDescriptor struct {
	Name        string
	Type        PropertyType
	Description string
	Options     []string // For PropSelect type

	// Get returns the current value of the property
	Get func() any

	// Set applies a new value to the property (called by ApplyProperty)
	Set func(value any)

	// Default returns the default value (used by Reset)
	Default func() any
}

// ToProperty converts a PropertyDescriptor to a Property (for the property editor)
func (pd *PropertyDescriptor) ToProperty() Property {
	return Property{
		Name:        pd.Name,
		Type:        pd.Type,
		Value:       pd.Get(),
		Options:     pd.Options,
		Description: pd.Description,
	}
}

// Demo represents a component demonstration
type Demo interface {
	// Name returns the display name
	Name() string

	// Description returns a brief description
	Description() string

	// Category returns Basic/Intermediate/Advanced
	Category() Category

	// Component returns the interactive demo component
	Component() tview.Primitive

	// Properties returns editable properties for live editing
	Properties() []Property

	// ApplyProperty updates the demo with a new property value
	ApplyProperty(name string, value any)

	// CodeExample returns the embedded source code
	CodeExample() string

	// Reset resets the demo to its initial state
	Reset()
}

// DemoBase provides default implementations of the Demo interface methods.
// Embed this in your demo struct to reduce boilerplate.
//
// Example usage:
//
//	type CheckboxDemo struct {
//	    demos.DemoBase
//	    checkbox *components.Checkbox
//	}
//
//	func NewCheckboxDemo() *CheckboxDemo {
//	    d := &CheckboxDemo{}
//	    d.DemoBase = demos.DemoBase{
//	        DemoName:        "Checkbox",
//	        DemoDescription: "Boolean toggle input",
//	        DemoCategory:    demos.Basic,
//	        DemoCode:        checkboxCode,
//	    }
//	    return d
//	}
type DemoBase struct {
	DemoName        string
	DemoDescription string
	DemoCategory    Category
	DemoCode        string

	// Props defines property descriptors for automatic property handling.
	// If set, Properties(), ApplyProperty(), and Reset() use these descriptors.
	Props []PropertyDescriptor

	// ComponentFunc creates the demo component. Called by Component().
	// If nil, you must override Component() in the embedding struct.
	ComponentFunc func() any

	// ResetFunc is called by Reset() after resetting all Props to defaults.
	// Use this for any custom reset logic beyond property values.
	ResetFunc func()
}

// Name implements Demo.Name
func (d *DemoBase) Name() string {
	return d.DemoName
}

// Description implements Demo.Description
func (d *DemoBase) Description() string {
	return d.DemoDescription
}

// Category implements Demo.Category
func (d *DemoBase) Category() Category {
	return d.DemoCategory
}

// CodeExample implements Demo.CodeExample
func (d *DemoBase) CodeExample() string {
	return d.DemoCode
}

// Properties implements Demo.Properties using PropertyDescriptors
func (d *DemoBase) Properties() []Property {
	if d.Props == nil {
		return nil
	}
	props := make([]Property, len(d.Props))
	for i, pd := range d.Props {
		props[i] = pd.ToProperty()
	}
	return props
}

// ApplyProperty implements Demo.ApplyProperty using PropertyDescriptors
func (d *DemoBase) ApplyProperty(name string, value any) {
	for _, pd := range d.Props {
		if pd.Name == name && pd.Set != nil {
			pd.Set(value)
			return
		}
	}
}

// Reset implements Demo.Reset using PropertyDescriptors
func (d *DemoBase) Reset() {
	// Reset all properties to their defaults
	for _, pd := range d.Props {
		if pd.Default != nil && pd.Set != nil {
			pd.Set(pd.Default())
		}
	}
	// Call custom reset function if provided
	if d.ResetFunc != nil {
		d.ResetFunc()
	}
}

// StringProp creates a string PropertyDescriptor with getter/setter closures.
// Example:
//
//	demos.StringProp("label", "Field label", d.getLabel, d.setLabel, "Default")
func StringProp(name, desc string, get func() string, set func(string), def string) PropertyDescriptor {
	return PropertyDescriptor{
		Name:        name,
		Type:        PropString,
		Description: desc,
		Get:         func() any { return get() },
		Set:         func(v any) { if s, ok := v.(string); ok { set(s) } },
		Default:     func() any { return def },
	}
}

// BoolProp creates a bool PropertyDescriptor with getter/setter closures.
// Example:
//
//	demos.BoolProp("enabled", "Enable feature", d.getEnabled, d.setEnabled, true)
func BoolProp(name, desc string, get func() bool, set func(bool), def bool) PropertyDescriptor {
	return PropertyDescriptor{
		Name:        name,
		Type:        PropBool,
		Description: desc,
		Get:         func() any { return get() },
		Set:         func(v any) { if b, ok := v.(bool); ok { set(b) } },
		Default:     func() any { return def },
	}
}

// IntProp creates an int PropertyDescriptor with getter/setter closures.
// Example:
//
//	demos.IntProp("count", "Item count", d.getCount, d.setCount, 10)
func IntProp(name, desc string, get func() int, set func(int), def int) PropertyDescriptor {
	return PropertyDescriptor{
		Name:        name,
		Type:        PropInt,
		Description: desc,
		Get:         func() any { return get() },
		Set:         func(v any) { if i, ok := v.(int); ok { set(i) } },
		Default:     func() any { return def },
	}
}

// SelectProp creates a select PropertyDescriptor with getter/setter closures.
// Example:
//
//	demos.SelectProp("align", "Alignment", []string{"left", "center", "right"}, d.getAlign, d.setAlign, "center")
func SelectProp(name, desc string, options []string, get func() string, set func(string), def string) PropertyDescriptor {
	return PropertyDescriptor{
		Name:        name,
		Type:        PropSelect,
		Description: desc,
		Options:     options,
		Get:         func() any { return get() },
		Set:         func(v any) { if s, ok := v.(string); ok { set(s) } },
		Default:     func() any { return def },
	}
}

// Registry holds all registered demos
type Registry struct {
	demos    []Demo
	byName   map[string]Demo
	byCategory map[Category][]Demo
}

// NewRegistry creates a new demo registry
func NewRegistry() *Registry {
	return &Registry{
		demos:      make([]Demo, 0),
		byName:     make(map[string]Demo),
		byCategory: make(map[Category][]Demo),
	}
}

// Register adds a demo to the registry
func (r *Registry) Register(demo Demo) {
	r.demos = append(r.demos, demo)
	r.byName[demo.Name()] = demo
	r.byCategory[demo.Category()] = append(r.byCategory[demo.Category()], demo)
}

// Get returns a demo by name
func (r *Registry) Get(name string) (Demo, bool) {
	demo, ok := r.byName[name]
	return demo, ok
}

// ByCategory returns all demos in a category
func (r *Registry) ByCategory(cat Category) []Demo {
	return r.byCategory[cat]
}

// All returns all registered demos
func (r *Registry) All() []Demo {
	return r.demos
}

// Categories returns all categories with demos
func (r *Registry) Categories() []Category {
	cats := make([]Category, 0, len(r.byCategory))
	for cat := range r.byCategory {
		cats = append(cats, cat)
	}
	return cats
}

// DefaultRegistry is the global demo registry
var DefaultRegistry = NewRegistry()

// Register adds a demo to the default registry
func Register(demo Demo) {
	DefaultRegistry.Register(demo)
}
