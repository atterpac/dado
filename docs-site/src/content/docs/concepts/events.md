---
title: Events
description: The typed event system used by dado components.
---

dado components surface user interactions through typed event structs. Every
event embeds `BaseEvent` which carries the source component.

## Event types

| Type | Fired when |
|---|---|
| `ActivateEvent` | User selects/activates an item (Enter key) |
| `ChangeEvent` | A value changes (text input, select, …) |
| `FocusEvent` | A component gains or loses focus |
| `KeyEvent` | A raw key press reaches a component |
| `SubmitEvent` | A form or input is submitted |

## Subscribing

Each event type has a corresponding `On*` method:

```go
input.OnChange(func(e components.ChangeEvent) {
    fmt.Println("new value:", e.Value)
})

form.OnSubmit(func(e components.SubmitEvent) {
    values := e.Values // map[string]string
})

grid.OnActivate(func(row int) {
    // row is the zero-based selected row index
})

component.OnFocus(func(e components.FocusEvent) {
    fmt.Println("gained focus:", e.Gained)
})
```

## BaseEvent

All event structs embed `BaseEvent`:

```go
type BaseEvent struct {
    Source core.Widget // the component that emitted the event
}
```

## Thread safety

Event callbacks run on the draw goroutine. If your callback performs slow
or blocking work, hand it off to a separate goroutine and use
`app.QueueUpdateDraw` to push UI updates back:

```go
grid.OnActivate(func(row int) {
    go func() {
        result := fetchDetails(row)
        app.QueueUpdateDraw(func() {
            detail.SetText(result)
        })
    }()
})
```
