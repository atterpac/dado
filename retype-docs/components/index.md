---
label: Components
icon: package
order: 80
expanded: true
---

# Component Reference

Quick reference for all Jig components with usage examples.

---

## Component Categories

+++ Primitives
### Primitive Components

These replace raw tview usage for common patterns:

| Component | Description |
|-----------|-------------|
| [Label](label.md) | Simple text display |
| [Button](button.md) | Clickable button with variants |
| [List](list.md) | Selectable list with vim navigation |
| [Layout](layout.md) | Row/Column layout helpers |

+++ Input
### Input Components

| Component | Description | Value Type |
|-----------|-------------|------------|
| [TextField](textfield.md) | Single-line text input | `string` |
| [TextArea](textarea.md) | Multi-line text input | `string` |
| [Checkbox](checkbox.md) | Boolean toggle | `bool` |
| [Select](select.md) | Dropdown selection | `SelectOption` |
| MultiSelect | Multi-option selection | `[]SelectOption` |
| RadioGroup | Single selection from options | `string` |
| Autocomplete | Text input with suggestions | `string` |

+++ Containers
### Container Components

| Component | Description |
|-----------|-------------|
| [Panel](panel.md) | Bordered container with title |
| [Modal](modal.md) | Centered dialog overlay |
| [Tabs](tabs.md) | Tabbed container |
| [Split](split.md) | Resizable split panes |
| MasterDetail | List-detail pattern |
| Drawer | Slide-out drawer panel |

+++ Display
### Display Components

| Component | Description |
|-----------|-------------|
| [Table](table.md) | Data table with selection |
| [Tree](tree.md) | Hierarchical tree view |
| VirtualList | Virtualized scrolling list |
| [Progress](progress.md) | Progress indicator |
| Toast | Notification messages |
| DiffViewer | Side-by-side diff view |
| CodeView | Syntax-highlighted code viewer |

+++ Navigation
### Navigation Components

| Component | Description |
|-----------|-------------|
| [ComponentBase](componentbase.md) | Base wrapper for nav.Component |
| KeyHintBar | Keyboard shortcut display |
| Crumbs | Breadcrumb navigation |
+++

---

## Common Interfaces

### ValueProvider

Implemented by all input components:

```go
type ValueProvider[T any] interface {
    Value() T
    SetValue(T) ValueProvider[T]
    HasValue() bool
    Clear()
}
```

### Validatable

Implemented by components with validation:

```go
type Validatable interface {
    Validate() error
    HasError() bool
    GetError() string
}
```

### Named

Implemented by named components:

```go
type Named interface {
    GetName() string
}
```

---

## Form Builder

Fluent API for building forms with validation.

```go
form := components.NewFormBuilder().
    Text("name", "Name").
        Placeholder("Enter name").
        Validate(validators.Required(), validators.MinLength(2)).
        Done().
    Text("email", "Email").
        Validate(validators.Required(), validators.Email()).
        Done().
    Select("role", "Role", []string{"Admin", "User", "Guest"}).
        Default("User").
        Done().
    Checkbox("notify", "Email notifications").
        Checked(true).
        Done().
    OnSubmit(func(values map[string]any) {
        name := values["name"].(string)
        email := values["email"].(string)
        role := values["role"].(components.SelectOption)
        notify := values["notify"].(bool)
        log.Printf("Submitted: %s, %s, %s, %v", name, email, role.Value, notify)
    }).
    OnCancel(func() {
        app.Pages().Pop()
    }).
    Build()
```

### As Modal

```go
modal := components.NewFormBuilder().
    Text("name", "Name").Done().
    OnSubmit(func(values map[string]any) { /* ... */ }).
    AsFormModal("Edit User", 60, 20)
```

---

## Validators

Standard validators in `validators/validators.go`:

| Validator | Description |
|-----------|-------------|
| `Required()` | Non-empty value |
| `MinLength(n)` | Minimum character count |
| `MaxLength(n)` | Maximum character count |
| `Email()` | Valid email format |
| `URL()` | Valid URL format |
| `Regex(pattern)` | Regex pattern matching |
| `Min(n)` | Minimum numeric value |
| `Max(n)` | Maximum numeric value |

Validators can be composed:

```go
Validate(validators.Required(), validators.MinLength(8), validators.Email())
```
