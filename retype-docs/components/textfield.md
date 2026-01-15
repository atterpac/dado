---
label: TextField
icon: typography
order: 75
---

# TextField

Single-line text input with validation and placeholder support.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

field := components.NewTextField("email").
    SetLabel("Email Address").
    SetPlaceholder("user@example.com")

// Get/set value
value := field.Value()
field.SetValue("test@example.com")
```

---

## Validation

```go
import "github.com/atterpac/jig/validators"

field := components.NewTextField("email").
    SetLabel("Email Address").
    SetValidator(func(value string) error {
        return validators.Email()(value)
    })

// Validate
if err := field.Validate(); err != nil {
    log.Printf("Validation error: %v", err)
}

// Check error state
if field.HasError() {
    errorMsg := field.GetError()
}
```

### Multiple Validators

```go
field := components.NewTextField("username").
    SetLabel("Username").
    SetValidator(func(value string) error {
        if err := validators.Required()(value); err != nil {
            return err
        }
        if err := validators.MinLength(3)(value); err != nil {
            return err
        }
        if err := validators.MaxLength(20)(value); err != nil {
            return err
        }
        return nil
    })
```

---

## Events

```go
field := components.NewTextField("search").
    SetOnChange(func(e *components.ChangeEvent[string]) {
        log.Printf("Changed: %q -> %q", e.OldValue, e.NewValue)
    }).
    SetOnSubmit(func(e *components.SubmitEvent) {
        log.Printf("Submitted: %s", e.Value)
    })
```

### Event Types

```go
type ChangeEvent[T any] struct {
    OldValue T
    NewValue T
    Index    int  // For selections
}

type SubmitEvent struct {
    Value any
}
```

---

## Methods

| Method | Description |
|--------|-------------|
| `SetLabel(string)` | Set field label |
| `SetPlaceholder(string)` | Set placeholder text |
| `SetValue(string)` | Set field value |
| `Value()` | Get current value |
| `SetValidator(func(string) error)` | Set validation function |
| `Validate()` | Run validation |
| `HasError()` | Check if validation failed |
| `GetError()` | Get error message |
| `Clear()` | Clear the field |
| `HasValue()` | Check if field has value |
| `SetOnChange(func(*ChangeEvent[string]))` | Change callback |
| `SetOnSubmit(func(*SubmitEvent))` | Submit callback |

---

## With FormBuilder

```go
form := components.NewFormBuilder().
    Text("name", "Name").
        Placeholder("Enter your name").
        Validate(validators.Required(), validators.MinLength(2)).
        Done().
    Text("email", "Email").
        Placeholder("user@example.com").
        Validate(validators.Required(), validators.Email()).
        Done().
    OnSubmit(func(values map[string]any) {
        name := values["name"].(string)
        email := values["email"].(string)
    }).
    Build()
```

---

## Example

```go
type SearchView struct {
    *components.ComponentBase
    field   *components.TextField
    results *components.Table
}

func NewSearchView(app *layout.App) *SearchView {
    field := components.NewTextField("search").
        SetLabel("Search").
        SetPlaceholder("Type to search...")

    results := components.NewTable()
    results.SetHeaders("Name", "Type")

    v := &SearchView{
        field:   field,
        results: results,
    }

    // Debounced search on change
    var debounce *time.Timer
    field.SetOnChange(func(e *components.ChangeEvent[string]) {
        if debounce != nil {
            debounce.Stop()
        }
        debounce = time.AfterFunc(300*time.Millisecond, func() {
            v.search(e.NewValue)
        })
    })

    flex := tview.NewFlex().SetDirection(tview.FlexRow).
        AddItem(field, 3, 0, true).
        AddItem(results, 0, 1, false)

    v.ComponentBase = components.NewComponentBase(flex).
        SetName("search")

    return v
}
```
