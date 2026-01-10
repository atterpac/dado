---
label: Troubleshooting
icon: tools
order: 50
---

# Troubleshooting

Common issues and solutions when working with Jig.

---

## UI Not Updating

### Problem
Changes to component state don't appear on screen.

### Solution
Ensure you're using `QueueUpdateDraw` for async updates:

```go
// WRONG: Direct mutation from goroutine
go func() {
    data := fetchData()
    table.SetData(data)  // Won't show!
}()

// CORRECT: Queue the update
go func() {
    data := fetchData()
    app.QueueUpdateDraw(func() {
        table.SetData(data)
    })
}()
```

---

## Deadlock on Theme Change

### Problem
Application freezes when switching themes.

### Cause
Calling `QueueUpdateDraw` from within a `Draw()` method.

### Solution
Never call `QueueUpdateDraw` from `Draw()`. The theme system uses lock-free atomic reads specifically to avoid this:

```go
// WRONG: This can deadlock
func (c *MyComponent) Draw(screen tcell.Screen) {
    app.QueueUpdateDraw(func() {
        // ...
    })
}

// CORRECT: Read theme colors directly (lock-free)
func (c *MyComponent) Draw(screen tcell.Screen) {
    style := tcell.StyleDefault.
        Background(theme.Bg()).
        Foreground(theme.Fg())
    // ...
}
```

---

## Theme Colors Not Updating

### Problem
Component colors don't change when switching themes.

### Cause
Theme colors were cached at creation time.

### Solution
Read theme colors at draw time:

```go
// WRONG: Cached at creation
func NewMyComponent() *MyComponent {
    return &MyComponent{
        bgColor: theme.Bg(),  // Won't update!
    }
}

// CORRECT: Read at draw time
func (c *MyComponent) Draw(screen tcell.Screen) {
    bg := theme.Bg()  // Always current
    // ...
}
```

Or register for theme updates:

```go
theme.Register(myBox)  // Auto-updates background
```

---

## Memory Leaks

### Problem
Memory usage grows over time.

### Cause
Subscriptions not being cleaned up.

### Solution
Always unsubscribe in `Stop()`:

```go
type MyView struct {
    *components.ComponentBase
    subscriptions []func()
}

func (v *MyView) Start() {
    unsub := status.Subscribe(func(old, new string) {
        v.update(new)
    })
    v.subscriptions = append(v.subscriptions, unsub)
}

func (v *MyView) Stop() {
    for _, unsub := range v.subscriptions {
        unsub()
    }
    v.subscriptions = nil
}
```

---

## Focus Issues

### Problem
Can't focus a component or focus is lost.

### Solution

1. Ensure the component is focusable:
```go
box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    return event
})
```

2. Queue focus changes from goroutines:
```go
// WRONG: Direct focus from goroutine
go func() {
    app.SetFocus(myComponent)  // Race condition!
}()

// CORRECT: Queue the focus change
go func() {
    app.QueueUpdate(func() {
        app.SetFocus(myComponent)
    })
}()
```

---

## Input Not Captured

### Problem
Key presses aren't being handled.

### Cause
Event being consumed earlier in the chain.

### Solution

Check the input chain order:
1. App.SetInputCapture (global)
2. Pages.InputHandler (modal)
3. Component.InputHandler
4. Primitive.InputHandler

Make sure you're returning `event` to pass through:

```go
func (v *MyView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Rune() == 'r' {
        v.refresh()
        return nil  // Consumed
    }
    return event  // Pass through!
}
```

---

## Component Not Starting

### Problem
`Start()` method isn't being called.

### Cause
Component not pushed to Pages.

### Solution
Ensure you're using `app.Pages().Push()`:

```go
// WRONG: Just setting content
app.SetRoot(myComponent, true)

// CORRECT: Push to pages stack
app.Pages().Push(myComponent)
```

---

## Modal Not Dismissing

### Problem
Modal stays on screen after Escape is pressed.

### Solution
Enable dismiss on escape:

```go
modal := components.NewModal(config).
    SetDismissOnEsc(true)
```

Or handle it manually:

```go
func (m *MyModal) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyEscape {
        app.Pages().Pop()
        return nil
    }
    return event
}
```

---

## Form Validation Not Working

### Problem
Form submits despite validation errors.

### Cause
Validation not being checked on submit.

### Solution
FormBuilder handles this automatically. If using manual form:

```go
func (v *FormView) submit() {
    if err := v.form.Validate(); err != nil {
        // Show error
        return
    }
    // Proceed with submit
}
```

---

## Table Selection Issues

### Problem
Can't select rows or multi-select doesn't work.

### Solution

Enable selection:
```go
table := components.NewTable()
table.SetSelectable(true, false)  // Row selectable, not cell
table.SetMultiSelect(true)        // For multi-select
```

Handle selection:
```go
table.SetOnSelect(func(row int) {
    data := table.GetRowData(row)
    // ...
})
```

---

## Async Loading Issues

### Problem
Loading indicator doesn't show or data doesn't appear.

### Solution
Use the async helpers correctly:

```go
async.NewLoader[[]Item]().
    WithIndicator(async.Toast("Loading...")).
    OnSuccess(func(items []Item) {
        // This runs on main goroutine
        table.SetItems(items)
    }).
    OnError(func(err error) {
        // Handle error
    }).
    Run(func(ctx context.Context) ([]Item, error) {
        // This runs in goroutine
        return fetchItems(ctx)
    })
```

---

## Debug Tips

### Check Component State

```go
log.Printf("Current page: %s", app.Pages().Current().GetName())
log.Printf("Stack depth: %d", app.Pages().StackDepth())
log.Printf("Can pop: %v", app.Pages().CanPop())
```

### Check Theme

```go
log.Printf("Current theme Bg: %v", theme.Bg())
```

### Check Focus

```go
focused := app.GetFocus()
log.Printf("Focused: %T", focused)
```

---

## Getting Help

If you're still stuck:

1. Check the [Architecture Guide](architecture.md) for understanding the internals
2. Review the [Getting Started](../getting-started.md) example
3. Run the interactive tutorial: `go run ./cmd/tutorial`
4. File an issue on GitHub with a minimal reproduction case
