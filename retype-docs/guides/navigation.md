---
label: Navigation
icon: arrow-switch
order: 85
---

# Navigation

Stack-based page navigation with modal support.

---

## Overview

Jig uses a stack-based navigation model where views are pushed onto and popped from a stack:

```
+---------------------------------------------+
|              nav.Pages                       |
|                                              |
|  Stack:                                      |
|  +--------------------------------------+    |
|  | [2] SettingsModal  <-- Current       |    |
|  +--------------------------------------+    |
|  | [1] DetailView                       |    |
|  +--------------------------------------+    |
|  | [0] HomeView                         |    |
|  +--------------------------------------+    |
|                                              |
|  Push(view) --> Add to top                   |
|  Pop()      --> Remove from top              |
|  Replace()  --> Swap top                     |
+---------------------------------------------+
```

---

## Basic Navigation

```go
// Push a new view (stops current, starts new)
app.Pages().Push(NewDetailView(item))

// Pop back to previous view (stops current, starts previous)
app.Pages().Pop()

// Replace current view (same stack depth)
app.Pages().Replace(NewAlternateView())

// Check if we can go back
if app.Pages().CanPop() {
    app.Pages().Pop()
}

// Get current view
current := app.Pages().Current()

// Get stack depth
depth := app.Pages().StackDepth()
```

---

## Lifecycle Integration

Navigation triggers component lifecycle methods:

```
Push(ComponentA)
  +-- ComponentA.Start()

Push(ComponentB)
  +-- ComponentA.Stop()
  +-- ComponentB.Start()

Pop()
  +-- ComponentB.Stop()
  +-- ComponentA.Start()
```

---

## Modal Behavior

Modals are components that implement the `nav.Modal` interface:

```go
type Modal interface {
    Component

    // ModalBehavior returns configuration for this modal
    ModalBehavior() components.ModalBehavior

    // OnDismiss is called before dismiss. Return false to cancel.
    OnDismiss() bool
}
```

### ModalBehavior Configuration

```go
type ModalBehavior struct {
    CapturesAllInput      bool  // Block input to underlying views
    DismissOnEsc          bool  // Auto-dismiss on Escape key
    RestoreFocusOnDismiss bool  // Return focus to previous component
    BlockUntilDismissed   bool  // Prevent Push/Pop while active
}
```

### Using Built-in Modal

```go
modal := components.NewModal(components.ModalConfig{
    Title:    "Confirm Delete",
    Width:    50,
    Height:   10,
    MinWidth: 40,
}).SetContent(confirmContent)

app.Pages().Push(modal)  // Modal lifecycle handled automatically
```

---

## Focus Management

### Focus Restoration

When pushing a modal, Pages saves the current focus:

```
Push(Modal)
  +-- Save current focus to focusStack
  +-- Focus modal content

Pop() [with RestoreFocusOnDismiss=true]
  +-- Remove modal
  +-- Restore focus from focusStack
```

### Manual Focus Control

```go
// Set focus to a specific primitive
app.SetFocus(myComponent)

// Get current focus
focused := app.GetFocus()
```

---

## Navigation Patterns

### Master-Detail

```go
type MasterView struct {
    *components.ComponentBase
    list *components.Table
    app  *layout.App
}

func (v *MasterView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyEnter {
        row := v.list.GetSelectedRow()
        item := v.items[row]
        v.app.Pages().Push(NewDetailView(v.app, item))
        return nil
    }
    return event
}

type DetailView struct {
    *components.ComponentBase
    app *layout.App
}

func (v *DetailView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyEscape {
        v.app.Pages().Pop()
        return nil
    }
    return event
}
```

### Confirmation Dialog

```go
func ConfirmAction(app *layout.App, message string, onConfirm func()) {
    content := tview.NewTextView().
        SetText(message).
        SetTextAlign(tview.AlignCenter)

    modal := components.NewModal(components.ModalConfig{
        Title:  "Confirm",
        Width:  50,
        Height: 8,
    }).
        SetContent(content).
        SetOnSubmit(func() {
            app.Pages().Pop()
            onConfirm()
        }).
        SetOnCancel(func() {
            app.Pages().Pop()
        }).
        SetDismissOnEsc(true)

    app.Pages().Push(modal)
}

// Usage
ConfirmAction(app, "Delete this item?", func() {
    deleteItem(item)
})
```

### Wizard Flow

```go
type WizardStep1 struct {
    *components.ComponentBase
    app  *layout.App
    data *WizardData
}

func (v *WizardStep1) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyEnter {
        // Validate and proceed
        v.app.Pages().Push(NewWizardStep2(v.app, v.data))
        return nil
    }
    if event.Key() == tcell.KeyEscape {
        v.app.Pages().Pop()  // Cancel wizard
        return nil
    }
    return event
}

type WizardStep2 struct {
    *components.ComponentBase
    app  *layout.App
    data *WizardData
}

func (v *WizardStep2) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyEnter {
        // Complete wizard
        v.data.Save()
        // Pop both steps
        v.app.Pages().Pop()
        v.app.Pages().Pop()
        return nil
    }
    if event.Key() == tcell.KeyEscape {
        v.app.Pages().Pop()  // Back to step 1
        return nil
    }
    return event
}
```

---

## Breadcrumbs

Enable breadcrumb navigation:

```go
app := layout.NewApp(layout.AppConfig{
    ShowCrumbs: true,
    BottomBar:  layout.NewMenu(),
})
```

Breadcrumbs automatically update based on the component name:

```go
v.ComponentBase = components.NewComponentBase(panel).
    SetName("settings")  // Shows "Settings" in breadcrumbs
```

---

## App Configuration

```go
app := layout.NewApp(layout.AppConfig{
    TopBar:     myTopBar,      // Custom top bar
    BottomBar:  layout.NewMenu(),  // Key hints menu
    ShowCrumbs: true,          // Enable breadcrumbs
})
```

### Layout Zones

```
+-------------------------------------------------------+
|                     TopBar                             |
+-------------------------------------------------------+
|                   nav.Crumbs                           |
+-------------------------------------------------------+
|                                                        |
|                   nav.Pages                            |
|                (Current Component)                     |
|                                                        |
+-------------------------------------------------------+
|                   BottomBar (Menu)                     |
+-------------------------------------------------------+
```
