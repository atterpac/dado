---
label: Modal
icon: screen-full
order: 85
---

# Modal

Centered dialog overlay with configurable behavior.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

content := tview.NewTextView().SetText("Are you sure?")

modal := components.NewModal(components.ModalConfig{
    Title:  "Confirm",
    Width:  50,
    Height: 10,
}).SetContent(content)

app.Pages().Push(modal)
```

---

## Configuration

```go
modal := components.NewModal(components.ModalConfig{
    Title:     "Confirm",
    Width:     50,
    Height:    10,
    MinWidth:  40,
    MaxWidth:  80,
    Backdrop:  true,
})
```

### ModalConfig Options

| Option | Type | Description |
|--------|------|-------------|
| `Title` | `string` | Modal title |
| `Width` | `int` | Preferred width |
| `Height` | `int` | Preferred height |
| `MinWidth` | `int` | Minimum width |
| `MaxWidth` | `int` | Maximum width |
| `Backdrop` | `bool` | Show backdrop overlay |

---

## Behavior Configuration

```go
modal.SetDismissOnEsc(true).
    SetBlockUntilDismissed(false).
    SetRestoreFocusOnDismiss(true)
```

### ModalBehavior

| Setting | Description |
|---------|-------------|
| `CapturesAllInput` | Block input to underlying views |
| `DismissOnEsc` | Auto-dismiss on Escape key |
| `RestoreFocusOnDismiss` | Return focus to previous component |
| `BlockUntilDismissed` | Prevent Push/Pop while active |

---

## Methods

| Method | Description |
|--------|-------------|
| `SetContent(tview.Primitive)` | Set modal content |
| `SetHints([]KeyHint)` | Set key binding hints |
| `SetOnSubmit(func())` | Submit callback |
| `SetOnCancel(func())` | Cancel callback |
| `SetDismissOnEsc(bool)` | Auto-dismiss on Escape |
| `SetBlockUntilDismissed(bool)` | Block navigation |

---

## Examples

### Confirmation Dialog

```go
func ConfirmDelete(app *layout.App, name string, onConfirm func()) {
    content := tview.NewTextView().
        SetText(fmt.Sprintf("Delete %q?\n\nThis action cannot be undone.", name)).
        SetTextAlign(tview.AlignCenter)

    modal := components.NewModal(components.ModalConfig{
        Title:  "Confirm Delete",
        Width:  50,
        Height: 8,
    }).
        SetContent(content).
        SetHints([]components.KeyHint{
            {Key: "Enter", Description: "Confirm"},
            {Key: "Esc", Description: "Cancel"},
        }).
        SetOnSubmit(func() {
            onConfirm()
            app.Pages().Pop()
        }).
        SetOnCancel(func() {
            app.Pages().Pop()
        }).
        SetDismissOnEsc(true)

    app.Pages().Push(modal)
}
```

### Form Modal

```go
modal := components.NewFormBuilder().
    Text("name", "Name").
        Validate(validators.Required()).
        Done().
    Text("email", "Email").
        Validate(validators.Email()).
        Done().
    OnSubmit(func(values map[string]any) {
        // Handle submit
        app.Pages().Pop()
    }).
    OnCancel(func() {
        app.Pages().Pop()
    }).
    AsFormModal("Edit User", 60, 15)

app.Pages().Push(modal)
```

### Progress Modal

```go
func ShowProgress(app *layout.App, title string) *components.ProgressModal {
    progress := components.NewProgressModal(components.ModalConfig{
        Title:  title,
        Width:  50,
        Height: 5,
    }).SetBlockUntilDismissed(true)

    app.Pages().Push(progress)
    return progress
}

// Usage
progress := ShowProgress(app, "Uploading...")
go func() {
    for i := 0; i <= 100; i += 10 {
        app.QueueUpdateDraw(func() {
            progress.SetProgress(float64(i) / 100)
        })
        time.Sleep(200 * time.Millisecond)
    }
    app.QueueUpdateDraw(func() {
        app.Pages().Pop()
    })
}()
```

---

## Modal Interface

Modals implement the `nav.Modal` interface:

```go
type Modal interface {
    Component

    // ModalBehavior returns configuration for this modal
    ModalBehavior() components.ModalBehavior

    // OnDismiss is called before dismiss. Return false to cancel.
    OnDismiss() bool
}
```

---

## Focus Management

When pushing a modal, Pages saves the current focus:

```
Push(Modal)
  |-- Save current focus to focusStack
  `-- Focus modal content

Pop() [with RestoreFocusOnDismiss=true]
  |-- Remove modal
  `-- Restore focus from focusStack
```
