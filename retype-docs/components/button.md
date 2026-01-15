---
label: Button
icon: play
order: 95
---

# Button

Clickable button component with variants and theming.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

button := components.NewButton("Click Me").
    OnClick(func() {
        log.Println("Clicked!")
    })
```

---

## Variants

```go
// Primary button (default)
primary := components.NewButton("Save").
    SetVariant(components.ButtonPrimary)

// Secondary button
secondary := components.NewButton("Cancel").
    SetVariant(components.ButtonSecondary)

// Danger button (for destructive actions)
danger := components.NewButton("Delete").
    SetVariant(components.ButtonDanger)

// Ghost button (minimal style)
ghost := components.NewButton("More Info").
    SetVariant(components.ButtonGhost)
```

| Variant | Description |
|---------|-------------|
| `ButtonDefault` | Accent-colored, standard button |
| `ButtonPrimary` | Same as default |
| `ButtonSecondary` | Dimmed, less prominent |
| `ButtonDanger` | Red/error color for destructive actions |
| `ButtonGhost` | Transparent background, accent text |

---

## Methods

| Method | Description |
|--------|-------------|
| `SetLabel(string)` | Set button text |
| `SetVariant(ButtonVariant)` | Set visual style |
| `SetDisabled(bool)` | Enable/disable the button |
| `SetOnClick(func())` | Set click handler |
| `OnClick(func())` | Alias for SetOnClick |
| `Click()` | Programmatically trigger click |
| `Primitive()` | Get underlying `*tview.Box` |

---

## Disabled State

```go
button := components.NewButton("Submit").
    SetDisabled(true)

// Enable later
button.SetDisabled(false)
```

---

## Examples

### Confirmation Buttons

```go
buttons := components.Row(
    components.NewButton("Cancel").
        SetVariant(components.ButtonSecondary).
        OnClick(func() {
            app.Pages().Pop()
        }),
    components.NewButton("Confirm").
        SetVariant(components.ButtonPrimary).
        OnClick(func() {
            save()
            app.Pages().Pop()
        }),
)
```

### Delete Button

```go
deleteBtn := components.NewButton("Delete").
    SetVariant(components.ButtonDanger).
    OnClick(func() {
        ConfirmAction(app, "Delete this item?", func() {
            deleteItem(item)
        })
    })
```

### In a Form Layout

```go
form := components.Column(
    nameField,
    emailField,
    components.Row(
        components.NewButton("Cancel").
            SetVariant(components.ButtonGhost).
            OnClick(cancel),
        components.NewButton("Save").
            OnClick(save),
    ),
)
```

---

## Keyboard Navigation

| Key | Action |
|-----|--------|
| `Enter` | Click button |
| `Space` | Click button |
| `Tab` | Move to next focusable element |

---

## Mouse Support

Buttons respond to left-click within their bounds.
