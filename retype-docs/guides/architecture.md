---
label: Architecture
icon: workflow
order: 95
---

# Architecture Guide

This guide covers the internals of Jig for developers who want to understand how the framework works.

---

## Design Philosophy

Jig is built around these core principles:

1. **Composition over Inheritance** - Components wrap primitives, not extend them
2. **Explicit Lifecycle** - Start/Stop methods make state transitions predictable
3. **Lock-Free Theme Reads** - Atomic storage prevents Draw() deadlocks
4. **Progressive Disclosure** - Simple APIs for common cases, power features when needed

---

## Key Abstractions

```
+-------------------------------------------------------------+
|                        layout.App                            |
|  +-------------------------------------------------------+  |
|  |                     TopBar                             |  |
|  +-------------------------------------------------------+  |
|  +-------------------------------------------------------+  |
|  |                   nav.Crumbs                           |  |
|  +-------------------------------------------------------+  |
|  +-------------------------------------------------------+  |
|  |                   nav.Pages                            |  |
|  |  +---------------------------------------------------+|  |
|  |  |            nav.Component                          ||  |
|  |  |  +---------------------------------------------+  ||  |
|  |  |  |        tview.Primitive                      |  ||  |
|  |  |  |   (Your UI built with components)           |  ||  |
|  |  |  +---------------------------------------------+  ||  |
|  |  +---------------------------------------------------+|  |
|  +-------------------------------------------------------+  |
|  +-------------------------------------------------------+  |
|  |                   BottomBar (Menu)                     |  |
|  +-------------------------------------------------------+  |
+-------------------------------------------------------------+
```

| Abstraction | Description |
|-------------|-------------|
| `layout.App` | Root container managing layout zones and input capture |
| `nav.Pages` | Stack-based navigation manager with modal support |
| `nav.Component` | Interface for navigable views with lifecycle |
| `tview.Primitive` | Base interface for all visual elements |

---

## Package Organization

```
jig/
+-- layout/      # App shell and layout primitives
+-- nav/         # Navigation and page management
+-- components/  # UI components (Panel, Form, Table, etc.)
+-- theme/       # Theming system with 13+ built-in themes
+-- binding/     # Data binding utilities (Value[T], FormBinding)
+-- async/       # Async operation helpers with indicators
+-- input/       # Input handling (KeyBindings builder, Vim nav)
+-- validators/  # Form validation
+-- recipes/     # Pre-built component compositions
+-- util/        # Internal utilities
```

---

## Threading Model

### Main Event Loop

Jig uses tview's event loop, which runs on the main goroutine:

```
+-------------------------------------------------------------+
|                     Main Goroutine                           |
|                                                              |
|  app.Run() --> Event Loop --+-- Input Events                 |
|                             |                                |
|                             +-- Draw Cycle                   |
|                             |   +-- Component.Draw()         |
|                             |                                |
|                             +-- QueueUpdateDraw callbacks    |
+-------------------------------------------------------------+
```

!!!danger Critical Rule
All UI mutations must happen on the main goroutine.
!!!

### QueueUpdate vs QueueUpdateDraw

+++ QueueUpdate
Schedule UI update without immediate redraw. Use when batching multiple updates.

```go
app.QueueUpdate(func() {
    table.SetCell(0, 0, "Updated")
    table.SetCell(0, 1, "Values")
    // No redraw yet
})
```
+++ QueueUpdateDraw
Schedule UI update AND trigger redraw. Use for single updates that should be visible immediately.

```go
app.QueueUpdateDraw(func() {
    statusBar.SetText("Loading complete")
    // Redraws after this function returns
})
```
+++

### When to Use Each

| Scenario | Method |
|----------|--------|
| Single value update | `QueueUpdateDraw` |
| Multiple related updates | `QueueUpdate` + `app.Draw()` at end |
| Response to user input | Usually automatic (event loop redraws) |
| Async callback result | `QueueUpdateDraw` |

### Safe Async Patterns

```go
// CORRECT: Update UI from async callback
go func() {
    data, err := fetchData()

    app.QueueUpdateDraw(func() {
        if err != nil {
            showError(err)
            return
        }
        table.SetData(data)
    })
}()

// INCORRECT: Direct UI mutation from goroutine
go func() {
    data, _ := fetchData()
    table.SetData(data)  // RACE CONDITION!
}()
```

### Common Pitfalls

| Issue | Solution |
|-------|----------|
| Deadlock in Draw() | Never call `QueueUpdateDraw` from within a `Draw()` method |
| Focus from goroutine | Always queue focus changes via `QueueUpdate` |
| Theme access | Safe from any goroutine (lock-free atomic reads) |
| Component state | Not thread-safe unless documented - use mutexes or queue updates |

---

## Component Lifecycle

### The nav.Component Interface

```go
type Component interface {
    tview.Primitive

    // Start is called when the component becomes active (shown)
    Start()

    // Stop is called when the component becomes inactive (hidden)
    Stop()

    // Hints returns key binding hints for this component
    Hints() []components.KeyHint
}
```

### Lifecycle Sequence

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

### Start/Stop Best Practices

+++ Start()
**Do this:**
- Load initial data (async)
- Start timers/pollers
- Register global handlers

```go
func (v *PollingView) Start() {
    v.done = make(chan struct{})
    v.ticker = time.NewTicker(5 * time.Second)

    go func() {
        for {
            select {
            case <-v.ticker.C:
                v.refresh()
            case <-v.done:
                return
            }
        }
    }()
}
```
+++ Stop()
**Do this:**
- Cancel pending operations
- Stop timers
- Unsubscribe from bindings

```go
func (v *PollingView) Stop() {
    if v.ticker != nil {
        v.ticker.Stop()
    }
    if v.done != nil {
        close(v.done)
    }
}
```
+++

---

## Input Handling Chain

Input events flow through this chain:

```
Input Event
    |
    v
App.SetInputCapture (global shortcuts, modal auto-dismiss)
    |
    v
Pages.InputHandler (modal input capture)
    |
    v
Current Component.InputHandler
    |
    v
Focused Primitive.InputHandler
```

Set a custom input handler on ComponentBase:

```go
v.ComponentBase = components.NewComponentBase(table).
    SetInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
        switch event.Rune() {
        case 'r':
            v.refresh()
            return nil  // Consumed
        case 'n':
            v.createNew()
            return nil
        }
        return event  // Pass through to primitive
    })
```

---

## Navigation System

### Stack-Based Model

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

### Basic Navigation

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

## Best Practices

### Memory Management

1. Always unsubscribe when component is destroyed
2. Unregister from theme when component is destroyed
3. Stop timers in Stop() method
4. Cancel contexts for async operations

```go
func (v *MyView) Stop() {
    // Unsubscribe all bindings
    for _, unsub := range v.subscriptions {
        unsub()
    }

    // Stop timers
    if v.ticker != nil {
        v.ticker.Stop()
    }

    // Cancel async operations
    if v.cancel != nil {
        v.cancel()
    }
}
```

### Performance Tips

1. **Batch updates**: Use `QueueUpdate` for multiple changes, `Draw()` once
2. **Lazy loading**: Load data in `Start()`, not constructor
3. **Debounce input**: For search fields, debounce rapid keystrokes
4. **Read theme at draw time**: Don't cache theme colors

### Error Handling

```go
// Show errors in UI, don't panic
go func() {
    data, err := fetchData(ctx)
    app.QueueUpdateDraw(func() {
        if err != nil {
            v.showError(err)
            return
        }
        v.setData(data)
    })
}()
```

### Testing Strategies

Separate state logic from UI:

```go
// Test state logic independently
func TestWorkflowState(t *testing.T) {
    state := NewWorkflowState()
    state.Add(Workflow{Name: "test"})
    assert.Len(t, state.Workflows, 1)
}
```
