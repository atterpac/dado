---
label: ComponentBase
icon: stack
order: 55
---

# ComponentBase

Wrapper to make any `tview.Primitive` implement `nav.Component`.

---

## Overview

`ComponentBase` is the foundation for creating navigable views in Jig. It wraps any tview primitive and provides:

- Lifecycle methods (`Start`, `Stop`)
- Key binding hints
- Input handling
- Navigation integration

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

type MyView struct {
    *components.ComponentBase
    content *tview.TextView
}

func NewMyView() *MyView {
    content := tview.NewTextView().SetText("Hello!")

    v := &MyView{content: content}
    v.ComponentBase = components.NewComponentBase(content).
        SetName("my-view")

    return v
}
```

---

## Configuration

```go
v.ComponentBase = components.NewComponentBase(primitive).
    SetName("my-view").                    // View identifier
    AddHint("r", "Refresh").               // Add key hint
    AddHint("Enter", "Select").            // Add another hint
    SetHints(hints).                       // Replace all hints
    SetOnStart(v.loadData).                // Lifecycle: shown
    SetOnStop(v.cleanup).                  // Lifecycle: hidden
    SetInputHandler(v.handleInput)         // Input handling
```

---

## Lifecycle

### Start

Called when the component becomes active (shown):

```go
func (v *MyView) Start() {
    // ComponentBase.Start() is called automatically
    // Your SetOnStart callback runs here

    // Good place to:
    // - Load data
    // - Start timers
    // - Register handlers
}
```

### Stop

Called when the component becomes inactive (hidden):

```go
func (v *MyView) Stop() {
    // ComponentBase.Stop() is called automatically
    // Your SetOnStop callback runs here

    // Good place to:
    // - Cancel pending operations
    // - Stop timers
    // - Unsubscribe from bindings
}
```

### Lifecycle Sequence

```
Push(ComponentA)
  `-- ComponentA.Start()

Push(ComponentB)
  |-- ComponentA.Stop()
  `-- ComponentB.Start()

Pop()
  |-- ComponentB.Stop()
  `-- ComponentA.Start()
```

---

## Input Handling

```go
func (v *MyView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Rune() {
    case 'r':
        v.refresh()
        return nil  // Event consumed
    case 'q':
        v.app.Pages().Pop()
        return nil
    }

    switch event.Key() {
    case tcell.KeyEnter:
        v.select()
        return nil
    case tcell.KeyEscape:
        v.app.Pages().Pop()
        return nil
    }

    return event  // Pass through to wrapped primitive
}
```

!!!info Return Values
- Return `nil` to consume the event (stop propagation)
- Return `event` to pass through to the wrapped primitive
!!!

---

## Methods

| Method | Description |
|--------|-------------|
| `SetName(string)` | Set view identifier |
| `GetName()` | Get view identifier |
| `AddHint(key, description)` | Add a key hint |
| `SetHints([]KeyHint)` | Replace all hints |
| `Hints()` | Get current hints |
| `SetOnStart(func())` | Set start callback |
| `SetOnStop(func())` | Set stop callback |
| `SetInputHandler(func(*tcell.EventKey, func(tview.Primitive)) *tcell.EventKey)` | Set input handler |

---

## Complete Example

```go
type UsersView struct {
    *components.ComponentBase
    table         *components.Table
    app           *layout.App
    subscriptions []func()
}

func NewUsersView(app *layout.App) *UsersView {
    table := components.NewTable()
    table.SetHeaders("Name", "Email", "Status")

    v := &UsersView{
        table: table,
        app:   app,
    }

    v.ComponentBase = components.NewComponentBase(table).
        SetName("users").
        AddHint("r", "Refresh").
        AddHint("Enter", "View").
        AddHint("n", "New").
        AddHint("d", "Delete").
        AddHint("Esc", "Back").
        SetOnStart(v.onStart).
        SetOnStop(v.onStop).
        SetInputHandler(v.handleInput)

    return v
}

func (v *UsersView) onStart() {
    // Load data
    v.loadUsers()

    // Subscribe to updates
    unsub := userStore.Subscribe(func(users []User) {
        v.updateTable(users)
    })
    v.subscriptions = append(v.subscriptions, unsub)
}

func (v *UsersView) onStop() {
    // Unsubscribe from updates
    for _, unsub := range v.subscriptions {
        unsub()
    }
    v.subscriptions = nil
}

func (v *UsersView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Rune() {
    case 'r':
        v.loadUsers()
        return nil
    case 'n':
        v.app.Pages().Push(NewUserForm(v.app))
        return nil
    case 'd':
        v.deleteSelected()
        return nil
    }

    switch event.Key() {
    case tcell.KeyEnter:
        if row := v.table.GetSelectedRow(); row >= 0 {
            v.app.Pages().Push(NewUserDetail(v.app, v.users[row]))
        }
        return nil
    case tcell.KeyEscape:
        v.app.Pages().Pop()
        return nil
    }

    return event
}

func (v *UsersView) loadUsers() {
    async.Load(
        func(ctx context.Context) ([]User, error) {
            return api.GetUsers(ctx)
        },
        func(users []User) {
            v.users = users
            v.updateTable(users)
        },
        func(err error) {
            components.ShowToast(v.app, "Failed to load users", 3*time.Second)
        },
    )
}

func (v *UsersView) updateTable(users []User) {
    v.table.ClearRows()
    for _, u := range users {
        v.table.AddRow(u.Name, u.Email, u.Status)
    }
}
```

---

## nav.Component Interface

`ComponentBase` implements the `nav.Component` interface:

```go
type Component interface {
    tview.Primitive

    // Start is called when the component becomes active
    Start()

    // Stop is called when the component becomes inactive
    Stop()

    // Hints returns key binding hints for this component
    Hints() []components.KeyHint
}
```
