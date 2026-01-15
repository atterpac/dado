# Modal Patterns

Modal dialog patterns for confirmations, forms, and overlays.

## Basic Modal

```go
content := tview.NewTextView().
    SetText("This is a modal dialog.\n\nPress Escape to close.")

modal := components.NewModal(components.ModalConfig{
    Title:    "Information",
    Width:    50,
    Height:   10,
    Backdrop: true,
})

modal.SetContent(content)

app.Pages().Push(modal)
```

---

## Modal Configuration

```go
config := components.ModalConfig{
    Title:     "Modal Title",
    Width:     60,       // Fixed width (0 = auto)
    Height:    20,       // Fixed height (0 = auto)
    MinWidth:  40,       // Minimum width
    MaxWidth:  100,      // Maximum width
    MinHeight: 10,       // Minimum height
    MaxHeight: 40,       // Maximum height
    Backdrop:  true,     // Dark overlay behind modal
}

modal := components.NewModal(config)
```

---

## Modal Behavior

```go
modal := components.NewModal(config)

// Configure behavior
modal.SetBehavior(components.ModalBehavior{
    CapturesAllInput:      true,   // Block input to underlying views
    DismissOnEsc:          true,   // Escape closes modal
    RestoreFocusOnDismiss: true,   // Restore focus to previous component
    BlockUntilDismissed:   false,  // Allow other navigation while open
    Backdrop:              true,   // Draw dark overlay
})

// Or set individual properties
modal.SetDismissOnEsc(true)
modal.SetBlockUntilDismissed(true)
```

---

## Confirm Dialog

```go
func ShowConfirmDialog(app *layout.App, message string, onConfirm func()) {
    content := tview.NewFlex().SetDirection(tview.FlexRow)

    // Message
    text := tview.NewTextView().
        SetText(message).
        SetTextAlign(tview.AlignCenter)
    content.AddItem(text, 0, 1, false)

    // Buttons
    buttons := tview.NewFlex()
    confirmBtn := tview.NewButton("Confirm")
    cancelBtn := tview.NewButton("Cancel")
    buttons.AddItem(nil, 0, 1, false)
    buttons.AddItem(confirmBtn, 10, 0, true)
    buttons.AddItem(nil, 2, 0, false)
    buttons.AddItem(cancelBtn, 10, 0, false)
    buttons.AddItem(nil, 0, 1, false)
    content.AddItem(buttons, 1, 0, true)

    modal := components.NewModal(components.ModalConfig{
        Title:    "Confirm",
        Width:    50,
        Height:   8,
        Backdrop: true,
    }).SetContent(content)

    modal.SetOnSubmit(func() {
        app.Pages().Pop()
        onConfirm()
    })
    modal.SetOnCancel(func() {
        app.Pages().Pop()
    })

    app.Pages().Push(modal)
}

// Usage
ShowConfirmDialog(app, "Delete this item?", func() {
    deleteItem()
})
```

---

## Form Modal

```go
modal := components.NewFormBuilder().
    Text("name", "Name").
        Validate(validators.Required()).
        Done().
    Text("email", "Email").
        Validate(validators.Required(), validators.Email()).
        Done().
    OnSubmit(func(values map[string]any) {
        name := values["name"].(string)
        email := values["email"].(string)
        createUser(name, email)
        app.Pages().Pop()
    }).
    OnCancel(func() {
        app.Pages().Pop()
    }).
    AsFormModal("New User", 60, 15)

// Form modals don't dismiss on Escape to prevent data loss
// Use AsConfirmModal for dialogs that should dismiss on Escape

app.Pages().Push(modal)
```

---

## Blocking Modal

Prevent navigation until modal is dismissed:

```go
modal := components.NewModal(components.ModalConfig{
    Title:    "Critical Action",
    Width:    50,
    Height:   10,
    Backdrop: true,
})

modal.SetBlockUntilDismissed(true)
modal.SetDismissOnEsc(false)  // Must explicitly close

modal.SetOnSubmit(func() {
    performCriticalAction()
    app.Pages().Pop()
})

app.Pages().Push(modal)

// While this modal is active:
// - Push() will be ignored
// - Pop() will be blocked
// - Only the modal's submit/cancel will work
```

---

## Dismiss Prevention

Confirm before closing:

```go
var hasChanges bool

modal := components.NewModal(config).
    SetContent(form)

modal.SetOnDismiss(func() bool {
    if hasChanges {
        // Show nested confirm dialog
        showUnsavedChangesDialog(func(discard bool) {
            if discard {
                hasChanges = false
                app.Pages().Pop()  // Now dismiss succeeds
            }
        })
        return false  // Cancel this dismiss attempt
    }
    return true  // Allow dismiss
})

form.SetOnChange(func() {
    hasChanges = true
})

app.Pages().Push(modal)
```

---

## Custom Modal Component

For complex modals, create a component:

```go
type SettingsModal struct {
    *components.ComponentBase
    modal   *components.Modal
    app     *layout.App
    onSave  func(settings Settings)
}

func NewSettingsModal(app *layout.App, current Settings) *SettingsModal {
    form := buildSettingsForm(current)

    modal := components.NewModal(components.ModalConfig{
        Title:  "Settings",
        Width:  70,
        Height: 25,
    }).SetContent(form)

    s := &SettingsModal{
        modal: modal,
        app:   app,
    }

    // Wrap modal with ComponentBase
    s.ComponentBase = components.NewComponentBase(modal).
        SetName("settings-modal").
        SetOnStart(s.onStart).
        SetInputHandler(s.handleInput)

    return s
}

func (s *SettingsModal) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Key() {
    case tcell.KeyEscape:
        s.app.Pages().Pop()
        return nil
    case tcell.KeyCtrlS:
        s.save()
        return nil
    }
    return event
}

func (s *SettingsModal) save() {
    if s.onSave != nil {
        s.onSave(s.getSettings())
    }
    s.app.Pages().Pop()
}

// Implement nav.Modal interface
func (s *SettingsModal) ModalBehavior() components.ModalBehavior {
    return components.DefaultModalBehavior()
}

func (s *SettingsModal) OnDismiss() bool {
    return true
}

// Usage
modal := NewSettingsModal(app, currentSettings)
modal.onSave = func(settings Settings) {
    applySettings(settings)
}
app.Pages().Push(modal)
```

---

## Progress Modal

Show progress for long operations:

```go
progress := components.NewProgress().
    SetLabel("Processing...")

modal := components.NewModal(components.ModalConfig{
    Title:  "Please Wait",
    Width:  50,
    Height: 6,
}).SetContent(progress)

modal.SetBlockUntilDismissed(true)
modal.SetDismissOnEsc(false)

app.Pages().Push(modal)

go func() {
    for i := 0; i <= 100; i++ {
        time.Sleep(50 * time.Millisecond)
        app.QueueUpdateDraw(func() {
            progress.SetProgress(float64(i) / 100)
        })
    }
    app.QueueUpdateDraw(func() {
        app.Pages().Pop()
    })
}()
```

---

## Focus Management

```go
// Focus specific element when modal opens
input := components.NewTextField("search")

modal := components.NewModal(config).
    SetContent(form).
    SetFocusOnShow(input)  // Focus this when modal opens

app.Pages().Push(modal)

// Focus is automatically restored to previous component on dismiss
// (when RestoreFocusOnDismiss is true)
```

---

## Nested Modals

Modals can be stacked:

```go
// First modal
modal1 := components.NewModal(config1).SetContent(content1)
app.Pages().Push(modal1)

// Second modal (stacks on top)
modal2 := components.NewModal(config2).SetContent(content2)
app.Pages().Push(modal2)

// Pop in order
app.Pages().Pop()  // Removes modal2
app.Pages().Pop()  // Removes modal1
```

---

## Best Practices

1. **Use backdrop** for modals that capture all input
2. **Prevent data loss** - use `OnDismiss` to confirm unsaved changes
3. **Set appropriate size** - use min/max for responsive sizing
4. **Provide clear escape** - always have a way to close the modal
5. **Use blocking sparingly** - only for critical operations

```go
// Good modal pattern
modal := components.NewModal(components.ModalConfig{
    Title:    "Edit Item",
    Width:    60,
    MinWidth: 40,
    MaxWidth: 80,
    Height:   20,
    Backdrop: true,
}).
    SetContent(form).
    SetDismissOnEsc(false).  // Prevent accidental close
    SetOnDismiss(func() bool {
        if hasChanges {
            showConfirm("Discard changes?")
            return false
        }
        return true
    })
```
