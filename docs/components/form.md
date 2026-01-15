# Form Patterns

Advanced form usage, validation, and custom fields.

## FormBuilder Quick Reference

```go
form := components.NewFormBuilder().
    Text(name, label).Placeholder(...).Validate(...).Done().
    TextArea(name, label).MaxLines(...).Done().
    Select(name, label, options).Default(...).Done().
    Checkbox(name, label).Checked(...).Done().
    Radio(name, label, options).Selected(...).Done().
    MultiSelect(name, label, options).Selected(...).Done().
    OnSubmit(func(values map[string]any) {...}).
    OnCancel(func() {...}).
    Build()
```

---

## Validation

### Built-in Validators

```go
import "github.com/atterpac/jig/validators"

form := components.NewFormBuilder().
    Text("username", "Username").
        Validate(
            validators.Required(),
            validators.MinLength(3),
            validators.MaxLength(20),
            validators.AlphaNumeric(),
        ).
        Done().
    Text("email", "Email").
        Validate(
            validators.Required(),
            validators.Email(),
        ).
        Done().
    Text("password", "Password").
        Validate(
            validators.Required(),
            validators.MinLength(8),
            validators.Pattern(`[A-Z]`, "must contain uppercase"),
            validators.Pattern(`[0-9]`, "must contain number"),
        ).
        Done().
    Build()
```

### Available Validators

| Validator | Description |
|-----------|-------------|
| `Required()` | Field must not be empty |
| `MinLength(n)` | Minimum string length |
| `MaxLength(n)` | Maximum string length |
| `Email()` | Valid email format |
| `URL()` | Valid URL format |
| `Pattern(regex, msg)` | Match regex pattern |
| `AlphaNumeric()` | Letters and numbers only |
| `Numeric()` | Numbers only |
| `In(options...)` | Value must be in list |

### Custom Validators

```go
// Simple custom validator
customValidator := func(value string) error {
    if !strings.HasPrefix(value, "PRJ-") {
        return fmt.Errorf("must start with PRJ-")
    }
    return nil
}

// Composable validator
func NotEmpty() validators.Validator {
    return func(value string) error {
        if strings.TrimSpace(value) == "" {
            return fmt.Errorf("cannot be empty or whitespace")
        }
        return nil
    }
}

// Using custom validators
form := components.NewFormBuilder().
    Text("project_id", "Project ID").
        Validate(validators.Required(), customValidator).
        Done().
    Build()
```

### Validation on Submit

```go
form := components.NewFormBuilder().
    Text("name", "Name").
        Validate(validators.Required()).
        Done().
    OnSubmit(func(values map[string]any) {
        // Only called if all fields pass validation
        name := values["name"].(string)
        saveUser(name)
    }).
    Build()
```

### Manual Validation

```go
// Validate all fields
if err := form.Validate(); err != nil {
    // Handle validation error
    return
}

// Check specific field
if field.HasError() {
    errorMsg := field.GetError()
}
```

---

## Form Values

### Extracting Values

```go
form := components.NewFormBuilder().
    Text("name", "Name").Done().
    Text("email", "Email").Done().
    Select("role", "Role", []string{"Admin", "User"}).Done().
    Checkbox("active", "Active").Done().
    OnSubmit(func(values map[string]any) {
        // Type assertions required
        name := values["name"].(string)
        email := values["email"].(string)
        role := values["role"].(components.SelectOption)
        active := values["active"].(bool)

        log.Printf("Name: %s, Email: %s, Role: %s, Active: %v",
            name, email, role.Value, active)
    }).
    Build()
```

### Value Types by Field

| Field Type | Value Type |
|------------|------------|
| Text | `string` |
| TextArea | `string` |
| Select | `components.SelectOption` |
| MultiSelect | `[]components.SelectOption` |
| Checkbox | `bool` |
| Radio | `string` |

---

## Form as Modal

### Standard Form Modal

```go
// Form modal (Escape doesn't auto-dismiss to prevent data loss)
modal := components.NewFormBuilder().
    Text("name", "Name").Done().
    Text("email", "Email").Done().
    OnSubmit(func(values map[string]any) {
        saveUser(values)
        app.Pages().Pop()
    }).
    OnCancel(func() {
        app.Pages().Pop()
    }).
    AsFormModal("Edit User", 60, 20)

app.Pages().Push(modal)
```

### Confirm Dialog

```go
// Confirm modal (Escape auto-dismisses)
modal := components.NewFormBuilder().
    // Add a "hidden" field or just use for button layout
    OnSubmit(func(values map[string]any) {
        deleteItem()
        app.Pages().Pop()
    }).
    OnCancel(func() {
        app.Pages().Pop()
    }).
    AsConfirmModal("Confirm Delete", 40, 8)

// Customize the modal
modal.SetContent(tview.NewTextView().SetText("Are you sure you want to delete this item?"))
app.Pages().Push(modal)
```

---

## Pre-Populating Forms

### Set Default Values

```go
form := components.NewFormBuilder().
    Text("name", "Name").
        Value("John Doe").  // Default value
        Done().
    Text("email", "Email").
        Value("john@example.com").
        Done().
    Select("role", "Role", []string{"Admin", "User", "Guest"}).
        Default("User").  // Default by value
        Done().
    Checkbox("active", "Active").
        Checked(true).  // Default checked
        Done().
    Build()
```

### Edit Mode Pattern

```go
func NewEditUserForm(user *User) *components.Form {
    return components.NewFormBuilder().
        Text("name", "Name").
            Value(user.Name).
            Validate(validators.Required()).
            Done().
        Text("email", "Email").
            Value(user.Email).
            Validate(validators.Required(), validators.Email()).
            Done().
        Select("role", "Role", []string{"Admin", "User", "Guest"}).
            Default(user.Role).
            Done().
        Checkbox("active", "Active").
            Checked(user.Active).
            Done().
        OnSubmit(func(values map[string]any) {
            user.Name = values["name"].(string)
            user.Email = values["email"].(string)
            user.Role = values["role"].(components.SelectOption).Value
            user.Active = values["active"].(bool)
            saveUser(user)
        }).
        Build()
}
```

---

## Custom Field Types

### Registering Custom Fields

```go
// Register a custom date picker field type
components.RegisterFormFieldType("date", func(name string) components.FormField {
    return NewDatePicker(name)
})

// Use in FormBuilder
form := components.NewFormBuilder().
    Text("name", "Name").Done().
    Custom("birthday", "date", "Birthday").
        Configure(func(field components.FormField) {
            datePicker := field.(*DatePicker)
            datePicker.SetFormat("2006-01-02")
        }).
        Done().
    Build()
```

### Custom Field Requirements

Custom fields must implement `FormField` interface:

```go
type FormField interface {
    tview.Primitive
    Named
    ValueProvider[any]  // Or specific type
    Validatable
}
```

---

## Form Layout

### Standard Form

The default Form component arranges fields vertically:

```go
form := components.NewFormBuilder().
    Text("field1", "Field 1").Done().
    Text("field2", "Field 2").Done().
    Text("field3", "Field 3").Done().
    Build()

// Fields are stacked vertically
// Submit/Cancel buttons at bottom
```

### Custom Layout

For complex layouts, build manually:

```go
// Create fields individually
nameField := components.NewTextField("name").SetLabel("Name")
emailField := components.NewTextField("email").SetLabel("Email")

// Custom layout
flex := tview.NewFlex().SetDirection(tview.FlexRow)

row1 := tview.NewFlex()
row1.AddItem(nameField, 0, 1, true)
row1.AddItem(emailField, 0, 1, false)
flex.AddItem(row1, 3, 0, true)

// Add to form manually
form := components.NewForm()
form.AddField(nameField)
form.AddField(emailField)
form.SetLayout(flex)  // Custom layout
```

---

## Change Handlers

### Field-Level Changes

```go
form := components.NewFormBuilder().
    Text("username", "Username").
        OnChange(func(e *components.ChangeEvent[string]) {
            // Validate username availability
            if isUsernameTaken(e.NewValue) {
                showWarning("Username is taken")
            }
        }).
        Done().
    Select("country", "Country", countries).
        OnChange(func(e *components.ChangeEvent[components.SelectOption]) {
            // Update states/provinces based on country
            updateStates(e.NewValue.Value)
        }).
        Done().
    Build()
```

### Form-Level Submission

```go
form := components.NewFormBuilder().
    Text("name", "Name").Done().
    OnSubmit(func(values map[string]any) {
        // All fields passed validation
        saveData(values)
    }).
    OnCancel(func() {
        // User cancelled (Escape or Cancel button)
        discardChanges()
    }).
    Build()
```

---

## Best Practices

1. **Always validate on submit** - Don't rely solely on field-level validation
2. **Provide clear error messages** - Use descriptive validator messages
3. **Pre-populate edit forms** - Users expect to see current values
4. **Handle cancel gracefully** - Ask for confirmation if there are unsaved changes
5. **Use appropriate field types** - Select for fixed options, Text for free-form

```go
// Example: Confirm before discarding changes
var hasChanges bool

form := components.NewFormBuilder().
    Text("name", "Name").
        OnChange(func(e *components.ChangeEvent[string]) {
            hasChanges = true
        }).
        Done().
    OnCancel(func() {
        if hasChanges {
            showConfirmDialog("Discard changes?", func() {
                app.Pages().Pop()
            })
        } else {
            app.Pages().Pop()
        }
    }).
    Build()
```
