# Jig Interactive Tutorial

An interactive tutorial application that demonstrates jig components and patterns.

## Running the Tutorial

```bash
go run ./cmd/tutorial
```

## Features

- **Component Showcase**: Browse all available components organized by category
- **Live Property Editor**: Modify component properties and see changes in real-time
- **Code Examples**: View source code for each component (press `c`)
- **Theme Picker**: Switch between 20+ built-in themes (press `t`)
- **Help System**: Quick reference for keyboard shortcuts (press `?`)

## Keyboard Navigation

| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down in lists |
| `Enter` | Select component |
| `Tab` | Switch between sidebar and demo |
| `c` | View code example |
| `t` | Open theme picker |
| `p` | Toggle properties panel |
| `?` | Show help |
| `q` | Quit |

## Component Categories

### Basic
- **TextField** - Single-line text input with placeholder and validation
- **Checkbox** - Boolean toggle with label
- **Progress** - Progress indicator with percentage

### Intermediate
- **Table** - Data table with headers, selection, and multi-select

### Advanced
- **Form** - Complete form with validation using FormBuilder

## Architecture

```
cmd/tutorial/
├── main.go           # Entry point
├── tutorial.go       # Main tutorial component
├── sidebar.go        # Component list sidebar
├── demo_view.go      # Demo display and property editor
├── theme_picker.go   # Theme selection modal
├── code_modal.go     # Code viewer modal
├── help_modal.go     # Help information modal
└── demos/
    ├── registry.go   # Demo registration system
    ├── basic/        # Basic component demos
    ├── intermediate/ # Intermediate demos
    └── advanced/     # Advanced demos
```

## Adding New Demos

1. Create a new file in the appropriate category folder:

```go
package basic

import (
    "github.com/rivo/tview"
    "github.com/atterpac/jig/cmd/tutorial/demos"
    "github.com/atterpac/jig/components"
)

func init() {
    demos.Register(&MyDemo{})
}

type MyDemo struct {
    component *components.MyComponent
}

func (d *MyDemo) Name() string        { return "MyComponent" }
func (d *MyDemo) Description() string { return "Brief description" }
func (d *MyDemo) Category() demos.Category { return demos.Basic }

func (d *MyDemo) Component() tview.Primitive {
    d.component = components.NewMyComponent()
    return d.component
}

func (d *MyDemo) Properties() []demos.Property {
    return []demos.Property{
        {Name: "prop1", Type: demos.PropString, Value: "value"},
        {Name: "prop2", Type: demos.PropBool, Value: true},
    }
}

func (d *MyDemo) ApplyProperty(name string, value any) {
    switch name {
    case "prop1":
        d.component.SetProp1(value.(string))
    case "prop2":
        d.component.SetProp2(value.(bool))
    }
}

func (d *MyDemo) CodeExample() string {
    return `// Code example shown in modal
component := components.NewMyComponent()
component.SetProp1("value")
`
}

func (d *MyDemo) Reset() {
    // Reset to initial state
}
```

2. Import the package in `main.go`:

```go
import (
    _ "github.com/atterpac/jig/cmd/tutorial/demos/basic"
)
```

## Property Types

| Type | Description | Value Type |
|------|-------------|------------|
| `PropString` | Text input | `string` |
| `PropBool` | Checkbox | `bool` |
| `PropInt` | Number input | `int` |
| `PropSelect` | Dropdown | `string` (uses `Options` field) |

## Learning Path

1. **Start with Basic**: TextField, Checkbox, Progress
2. **Move to Intermediate**: Table with selection
3. **Explore Advanced**: Form with validation
4. **Experiment with Themes**: Try different color schemes
5. **Study Code Examples**: Press `c` on any component

## Related Documentation

- [Getting Started](../../docs/GETTING_STARTED.md)
- [Architecture Guide](../../docs/ARCHITECTURE.md)
- [Component Reference](../../docs/components/README.md)
