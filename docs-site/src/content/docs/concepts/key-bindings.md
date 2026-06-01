---
title: Key Bindings
description: How dado handles keyboard input and how to add custom bindings.
---

dado exposes keyboard input through two layers: per-component key handlers and
global application-level bindings.

## Per-component key events

Subscribe to raw key events on any component:

```go
grid.OnKey(func(e components.KeyEvent) bool {
    if e.Key == tcell.KeyRune && e.Rune == 'd' {
        deleteSelected()
        return true // consumed — stop propagation
    }
    return false // pass through to default handler
})
```

Return `true` to consume the event; `false` to let the component process it
normally.

## Global bindings

Register app-level shortcuts that fire regardless of which component has focus:

```go
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    switch event.Key() {
    case tcell.KeyCtrlQ:
        app.Stop()
        return nil
    case tcell.KeyF1:
        pushHelpModal()
        return nil
    }
    return event // pass through
})
```

Returning `nil` consumes the event. Returning the event unchanged passes it
down to the focused component.

## Common key constants

```go
tcell.KeyEnter
tcell.KeyEscape
tcell.KeyTab
tcell.KeyBacktab   // Shift+Tab
tcell.KeyCtrlC
tcell.KeyF1 … tcell.KeyF12
tcell.KeyRune      // printable character — check event.Rune()
```

## Hint bars

Use `components.NewKeyHintBar` to display a context-sensitive key legend at the
bottom of the screen. Update it when focus changes:

```go
hints := components.NewKeyHintBar().
    SetHints([]components.Hint{
        {Key: "↑↓", Label: "navigate"},
        {Key: "Enter", Label: "open"},
        {Key: "q", Label: "quit"},
    })
```
