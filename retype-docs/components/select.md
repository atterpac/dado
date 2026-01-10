---
label: Select
icon: single-select
order: 65
---

# Select

Dropdown selection with single choice.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

// Simple string options
sel := components.NewSelect("role").
    SetLabel("Role").
    SetOptions([]string{"Admin", "User", "Guest"}).
    SetPlaceholder("Select a role").
    SetDefault("User")

// Get value
option := sel.Value()  // SelectOption{Label, Value}
index := sel.SelectedIndex()
```

---

## Custom Options

```go
sel := components.NewSelect("status").
    SetLabel("Status").
    SetOptionsWithValues([]components.SelectOption{
        {Label: "Active", Value: "active"},
        {Label: "Inactive", Value: "inactive"},
        {Label: "Pending Review", Value: "pending"},
    })
```

### SelectOption

```go
type SelectOption struct {
    Label string  // Display text
    Value string  // Underlying value
}
```

---

## Events

```go
sel := components.NewSelect("priority").
    SetOnChange(func(e *components.ChangeEvent[components.SelectOption]) {
        log.Printf("Selected: %s (value: %s)", e.NewValue.Label, e.NewValue.Value)
    })
```

---

## Methods

| Method | Description |
|--------|-------------|
| `SetLabel(string)` | Set field label |
| `SetPlaceholder(string)` | Set placeholder text |
| `SetOptions([]string)` | Set simple string options |
| `SetOptionsWithValues([]SelectOption)` | Set label/value options |
| `SetDefault(string)` | Set default selection by label |
| `Value()` | Get selected option |
| `SelectedIndex()` | Get selected index |
| `Clear()` | Clear selection |
| `SetOnChange(func(*ChangeEvent[SelectOption]))` | Change callback |

---

## With FormBuilder

```go
form := components.NewFormBuilder().
    Select("role", "Role", []string{"Admin", "User", "Guest"}).
        Default("User").
        Done().
    Select("priority", "Priority", []string{"Low", "Medium", "High", "Critical"}).
        Default("Medium").
        Done().
    OnSubmit(func(values map[string]any) {
        role := values["role"].(components.SelectOption)
        priority := values["priority"].(components.SelectOption)

        log.Printf("Role: %s, Priority: %s", role.Value, priority.Value)
    }).
    Build()
```

---

## Example

```go
type FilterView struct {
    *components.ComponentBase
    status   *components.Select
    category *components.Select
    onFilter func(status, category string)
}

func NewFilterView(onFilter func(status, category string)) *FilterView {
    status := components.NewSelect("status").
        SetLabel("Status").
        SetOptionsWithValues([]components.SelectOption{
            {Label: "All", Value: ""},
            {Label: "Active", Value: "active"},
            {Label: "Inactive", Value: "inactive"},
            {Label: "Pending", Value: "pending"},
        })

    category := components.NewSelect("category").
        SetLabel("Category").
        SetOptions([]string{"All", "Development", "Design", "Marketing"})

    v := &FilterView{
        status:   status,
        category: category,
        onFilter: onFilter,
    }

    applyFilter := func() {
        statusVal := status.Value().Value
        catVal := category.Value().Label
        if catVal == "All" {
            catVal = ""
        }
        onFilter(statusVal, catVal)
    }

    status.SetOnChange(func(e *components.ChangeEvent[components.SelectOption]) {
        applyFilter()
    })

    category.SetOnChange(func(e *components.ChangeEvent[components.SelectOption]) {
        applyFilter()
    })

    flex := tview.NewFlex().
        AddItem(status, 0, 1, true).
        AddItem(category, 0, 1, false)

    v.ComponentBase = components.NewComponentBase(flex).
        SetName("filter")

    return v
}
```

---

## Keyboard Navigation

| Key | Action |
|-----|--------|
| `Enter` / `Space` | Open dropdown |
| `j` / `Down` | Move down in list |
| `k` / `Up` | Move up in list |
| `Enter` | Select option |
| `Esc` | Close dropdown |
