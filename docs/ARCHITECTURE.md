# Jig Architecture Guide

This guide covers the internals of jig for developers who want to understand how the framework works and how to use it correctly.

## 1. Overview

### Design Philosophy

Jig is built around these core principles:

1. **Composition over Inheritance** - Components wrap primitives, not extend them
2. **Explicit Lifecycle** - Start/Stop methods make state transitions predictable
3. **Lock-Free Theme Reads** - Atomic storage prevents Draw() deadlocks
4. **Progressive Disclosure** - Simple APIs for common cases, power features when needed

### Key Abstractions

```
┌─────────────────────────────────────────────────────────────┐
│                        layout.App                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                     TopBar                           │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   nav.Crumbs                         │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   nav.Pages                          │   │
│  │  ┌─────────────────────────────────────────────┐    │   │
│  │  │            nav.Component                     │    │   │
│  │  │  ┌───────────────────────────────────────┐  │    │   │
│  │  │  │        tview.Primitive                 │  │    │   │
│  │  │  │   (Your UI built with components)      │  │    │   │
│  │  │  └───────────────────────────────────────┘  │    │   │
│  │  └─────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   BottomBar (Menu)                   │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

- **layout.App**: Root container managing layout zones and input capture
- **nav.Pages**: Stack-based navigation manager with modal support
- **nav.Component**: Interface for navigable views with lifecycle
- **tview.Primitive**: Base interface for all visual elements

### Package Organization

```
jig/
├── layout/      # App shell and layout primitives
├── nav/         # Navigation and page management
├── components/  # UI components (Panel, Form, Table, etc.)
├── theme/       # Theming system with 20+ built-in themes
├── binding/     # Data binding utilities (Value[T], FormBinding)
├── async/       # Async operation helpers with indicators
├── input/       # Input handling (KeyBindings builder, Vim nav)
├── validators/  # Form validation
├── recipes/     # Pre-built component compositions
└── util/        # Internal utilities
```

---

## 2. Threading Model

### Main Event Loop

Jig uses tview's event loop, which runs on the main goroutine:

```
┌─────────────────────────────────────────────────────────────┐
│                     Main Goroutine                          │
│                                                             │
│  app.Run() ──► Event Loop ──┬──► Input Events               │
│                             │                               │
│                             ├──► Draw Cycle                 │
│                             │    └──► Component.Draw()      │
│                             │                               │
│                             └──► QueueUpdateDraw callbacks  │
└─────────────────────────────────────────────────────────────┘
```

**Critical Rule**: All UI mutations must happen on the main goroutine.

### QueueUpdate vs QueueUpdateDraw

```go
// QueueUpdate: Schedule UI update without immediate redraw
// Use when batching multiple updates
app.QueueUpdate(func() {
    table.SetCell(0, 0, "Updated")
    table.SetCell(0, 1, "Values")
    // No redraw yet
})

// QueueUpdateDraw: Schedule UI update AND trigger redraw
// Use for single updates that should be visible immediately
app.QueueUpdateDraw(func() {
    statusBar.SetText("Loading complete")
    // Redraws after this function returns
})
```

**When to use each**:

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

    // Must queue UI updates
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

1. **Deadlock in Draw()**: Never call `QueueUpdateDraw` from within a `Draw()` method - this blocks waiting for itself
2. **Focus from goroutine**: Always queue focus changes via `QueueUpdate`
3. **Theme access**: Safe from any goroutine (lock-free atomic reads)
4. **Component state**: Not thread-safe unless documented - use mutexes or queue updates

---

## 3. Component Lifecycle

### The nav.Component Interface

Every navigable view must implement this interface:

```go
// nav.Component in nav/component.go
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
  └─► ComponentA.Start()

Push(ComponentB)
  ├─► ComponentA.Stop()
  └─► ComponentB.Start()

Pop()
  ├─► ComponentB.Stop()
  └─► ComponentA.Start()
```

### Using ComponentBase

`ComponentBase` wraps any `tview.Primitive` to implement `nav.Component`:

```go
type MyView struct {
    *components.ComponentBase
    table *components.Table
}

func NewMyView() *MyView {
    table := components.NewTable()

    v := &MyView{table: table}
    v.ComponentBase = components.NewComponentBase(table).
        SetName("my-view").
        AddHint("Enter", "Select").
        AddHint("q", "Quit").
        SetOnStart(v.loadData).
        SetOnStop(v.cleanup)

    return v
}

func (v *MyView) loadData() {
    // Called when view becomes active
    go func() {
        data := fetchData()
        app.QueueUpdateDraw(func() {
            v.table.SetData(data)
        })
    }()
}

func (v *MyView) cleanup() {
    // Called when view becomes inactive
    // Cancel pending operations, stop timers
}
```

### Start/Stop Best Practices

**Start()** - Do this:
- Load initial data (async)
- Start timers/pollers
- Register global handlers

**Stop()** - Do this:
- Cancel pending operations
- Stop timers
- Unsubscribe from bindings

```go
type PollingView struct {
    *components.ComponentBase
    ticker *time.Ticker
    done   chan struct{}
}

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

func (v *PollingView) Stop() {
    if v.ticker != nil {
        v.ticker.Stop()
    }
    if v.done != nil {
        close(v.done)
    }
}
```

### Input Handling Chain

Input events flow through this chain:

```
Input Event
    │
    ▼
App.SetInputCapture (global shortcuts, modal auto-dismiss)
    │
    ▼
Pages.InputHandler (modal input capture)
    │
    ▼
Current Component.InputHandler
    │
    ▼
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
            return nil  // Consumed
        }
        return event  // Pass through to primitive
    })
```

---

## 4. Navigation System

### Stack-Based Model

```
┌─────────────────────────────────────────┐
│              nav.Pages                   │
│                                         │
│  Stack:                                 │
│  ┌─────────────────────────────────┐   │
│  │ [2] SettingsModal  ◄── Current  │   │
│  ├─────────────────────────────────┤   │
│  │ [1] DetailView                  │   │
│  ├─────────────────────────────────┤   │
│  │ [0] HomeView                    │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Push(view) ──► Add to top             │
│  Pop()      ──► Remove from top        │
│  Replace()  ──► Swap top               │
└─────────────────────────────────────────┘
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

### Modal Behavior

Modals are components that implement the `nav.Modal` interface:

```go
// nav.Modal interface
type Modal interface {
    Component

    // ModalBehavior returns configuration for this modal
    ModalBehavior() components.ModalBehavior

    // OnDismiss is called before dismiss. Return false to cancel.
    OnDismiss() bool
}
```

Modal behavior configuration:

```go
type ModalBehavior struct {
    CapturesAllInput      bool  // Block input to underlying views
    DismissOnEsc          bool  // Auto-dismiss on Escape key
    RestoreFocusOnDismiss bool  // Return focus to previous component
    BlockUntilDismissed   bool  // Prevent Push/Pop while active
}
```

Using the built-in Modal component:

```go
modal := components.NewModal(components.ModalConfig{
    Title:    "Confirm Delete",
    Width:    50,
    Height:   10,
    MinWidth: 40,
}).SetContent(confirmContent)

app.Pages().Push(modal)  // Modal lifecycle handled automatically
```

### Focus Restoration

When pushing a modal, Pages saves the current focus:

```
Push(Modal)
  ├─► Save current focus to focusStack
  └─► Focus modal content

Pop() [with RestoreFocusOnDismiss=true]
  ├─► Remove modal
  └─► Restore focus from focusStack
```

---

## 5. Theme System

### Lock-Free Design

The theme system uses `atomic.Value` for lock-free reads:

```go
// In theme/provider.go
var activeTheme atomic.Value  // stores *themeHolder

// Safe to call from any goroutine, including Draw()
func Get() Theme {
    if holder := activeTheme.Load(); holder != nil {
        return holder.(*themeHolder).theme
    }
    return nil
}
```

**Why lock-free matters**: `Draw()` is called frequently and must never block. Theme colors are read during every draw cycle. A mutex here would cause deadlocks when `SetProvider()` is called during drawing.

### Reading Theme Colors

Always read colors at draw time, not at creation time:

```go
// CORRECT: Read at draw time
func (c *MyComponent) Draw(screen tcell.Screen) {
    style := tcell.StyleDefault.
        Background(theme.Bg()).
        Foreground(theme.Fg())
    // ... draw with style
}

// INCORRECT: Cache colors at creation
func NewMyComponent() *MyComponent {
    return &MyComponent{
        bgColor: theme.Bg(),  // Won't update on theme change!
    }
}
```

### Theme Interface

All themes implement this interface:

```go
type Theme interface {
    // Base colors
    Bg() tcell.Color
    BgLight() tcell.Color
    BgDark() tcell.Color
    Fg() tcell.Color
    FgDim() tcell.Color
    FgMuted() tcell.Color

    // Accent colors
    Accent() tcell.Color
    AccentDim() tcell.Color
    Highlight() tcell.Color

    // Semantic colors
    Success() tcell.Color
    Warning() tcell.Color
    Error() tcell.Color
    Info() tcell.Color

    // Border colors
    Border() tcell.Color
    BorderFocus() tcell.Color

    // UI element colors
    Header() tcell.Color
    Menu() tcell.Color
    TableHeader() tcell.Color
    Key() tcell.Color
    Crumb() tcell.Color
    PanelBorder() tcell.Color
    PanelTitle() tcell.Color
}
```

### Registration Patterns

```go
// 1. Auto-background update (simplest)
// Automatically calls SetBackgroundColor(theme.Bg()) on theme change
theme.Register(myBox)

// 2. Custom refresh logic
// Calls myComponent.RefreshTheme() on theme change
unregister := theme.RegisterRefreshable(myComponent)
defer unregister()  // Important: prevent memory leak

// 3. Callback subscription
unsubscribe := theme.OnThemeChange(func() {
    // Custom handling
})
defer unsubscribe()
```

### Runtime Theme Switching

```go
// Switch theme - all registered components update automatically
theme.SetProvider(theme.TokyoNight())

// Available built-in themes
theme.SetProvider(theme.Catppuccin())
theme.SetProvider(theme.Dracula())
theme.SetProvider(theme.Nord())
theme.SetProvider(theme.Gruvbox())
// ... 20+ themes available
```

### Custom Theme Creation

```go
// Using the builder
custom := theme.NewBuilder().
    SetBg(tcell.ColorBlack).
    SetFg(tcell.ColorWhite).
    SetAccent(tcell.ColorBlue).
    Build()

// Load from JSON file
custom, err := theme.LoadFromFile("mytheme.json")

// Implement Theme interface directly
type MyTheme struct{}
func (t *MyTheme) Bg() tcell.Color { return tcell.ColorBlack }
// ... implement all methods
```

---

## 6. Data Binding

### Value[T] Observable

Observable values notify listeners when changed:

```go
// Create observable value
count := binding.NewValue(0)

// Subscribe to changes
unsubscribe := count.Subscribe(func(old, new int) {
    log.Printf("Count changed: %d -> %d", old, new)
})
defer unsubscribe()

// Update value (notifies listeners)
count.Set(1)

// Update and trigger redraw
count.SetAndDraw(2)

// Update with function
count.Update(func(n int) int { return n + 1 })
```

### Binding to UI

```go
// One-way binding: value -> UI
status := binding.NewValue("Ready")
unsubscribe := status.BindToWithDraw(func(s string) {
    statusLabel.SetText(s)
})
defer unsubscribe()

// Now status.Set("Loading...") updates the label and redraws
```

### Computed Values

```go
firstName := binding.NewValue("John")
lastName := binding.NewValue("Doe")

// Computed from single source (same type)
upperFirst := firstName.Computed(strings.ToUpper)

// Computed with type transformation
fullName := binding.ComputedTo(firstName, func(first string) string {
    return first + " " + lastName.Get()
})
```

### Memory Management

Always unsubscribe to prevent memory leaks:

```go
type MyView struct {
    *components.ComponentBase
    subscriptions []func()
}

func (v *MyView) Start() {
    unsub := status.Subscribe(func(old, new string) {
        v.updateStatus(new)
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

## 7. Best Practices

### Memory Management

1. **Always unsubscribe** when component is destroyed
2. **Unregister from theme** when component is destroyed
3. **Stop timers** in Stop() method
4. **Cancel contexts** for async operations

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

// Test components with mock app (if needed)
func TestWorkflowView(t *testing.T) {
    app := testutil.NewMockApp()
    view := NewWorkflowView(app)

    view.Start()
    // Assert initial state

    // Simulate async completion
    // Assert final state
}
```
