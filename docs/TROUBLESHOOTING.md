# Troubleshooting Guide

Common issues and solutions when working with jig.

## Common Issues

### UI Not Updating

**Symptom**: You update component state but the UI doesn't reflect changes.

**Cause**: UI updates from goroutines must be queued to the main thread.

**Solution**:

```go
// WRONG: Direct mutation from goroutine
go func() {
    data := fetchData()
    table.SetData(data)  // Won't update!
}()

// CORRECT: Queue update to main thread
go func() {
    data := fetchData()
    app.QueueUpdateDraw(func() {
        table.SetData(data)
    })
}()
```

If using the async package, this is handled automatically:

```go
async.NewLoader[[]Item]().
    OnSuccess(func(data []Item) {
        // Already on main thread
        table.SetData(data)
    }).
    Run(fetchData)
```

---

### Deadlocks / App Freezes

**Symptom**: App freezes, no response to input, may need Ctrl+C to exit.

**Common Causes**:

1. **Calling QueueUpdateDraw from Draw()**

```go
// WRONG: Will deadlock
func (c *MyComponent) Draw(screen tcell.Screen) {
    app.QueueUpdateDraw(func() {  // DEADLOCK!
        c.updateSomething()
    })
    c.Box.Draw(screen)
}

// CORRECT: Update state outside Draw
func (c *MyComponent) refresh() {
    c.updateSomething()
    app.Draw()  // Request redraw
}
```

2. **Blocking in event handlers**

```go
// WRONG: Blocking the event loop
func (v *View) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Rune() == 'r' {
        data := fetchDataSync()  // BLOCKS UI!
        v.setData(data)
        return nil
    }
    return event
}

// CORRECT: Use async
func (v *View) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Rune() == 'r' {
        go func() {
            data := fetchDataSync()
            app.QueueUpdateDraw(func() {
                v.setData(data)
            })
        }()
        return nil
    }
    return event
}
```

3. **Dismiss causing deadlock in modal**

The App handles modal dismiss via goroutine to avoid deadlocks. If you're implementing custom modal handling:

```go
// CORRECT: App already does this internally
go func() {
    app.QueueUpdateDraw(func() {
        app.Pages().DismissModal()
    })
}()
```

---

### Focus Problems

**Symptom**: Focus jumps unexpectedly or keyboard input goes to wrong component.

**Cause**: Multiple components fighting for focus, or focus not properly delegated.

**Solutions**:

1. **Implement Focus correctly**

```go
func (v *MyView) Focus(delegate func(tview.Primitive)) {
    // Delegate to the component that should receive input
    delegate(v.activeChild)
}
```

2. **Queue focus changes from goroutines**

```go
// WRONG: Direct focus from goroutine
go func() {
    time.Sleep(time.Second)
    app.SetFocus(myComponent)  // Race condition!
}()

// CORRECT: Queue focus change
go func() {
    time.Sleep(time.Second)
    app.QueueUpdate(func() {
        app.SetFocus(myComponent)
    })
}()
```

3. **Check InputHandler chain**

Events flow: App.SetInputCapture -> Pages -> Component -> Primitive

If a handler returns `nil`, the event is consumed. Return the event to pass it through:

```go
func (v *View) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Rune() == 'r' {
        v.refresh()
        return nil  // Consumed - won't reach primitive
    }
    return event  // Pass through to primitive's handler
}
```

---

### Theme Not Applying

**Symptom**: Component uses wrong colors, or doesn't update when theme changes.

**Cause**: Component not registered with theme system, or caching colors.

**Solutions**:

1. **Register primitives for background updates**

```go
func NewMyComponent() *MyComponent {
    box := tview.NewBox()
    theme.Register(box)  // Auto-updates background on theme change
    return &MyComponent{Box: box}
}
```

2. **Don't cache theme colors**

```go
// WRONG: Cached at creation time
func NewMyComponent() *MyComponent {
    return &MyComponent{
        bgColor: theme.Bg(),  // Won't update on theme change!
    }
}

// CORRECT: Read at draw time
func (c *MyComponent) Draw(screen tcell.Screen) {
    style := tcell.StyleDefault.
        Background(theme.Bg()).  // Fresh read every draw
        Foreground(theme.Fg())
    // ...
}
```

3. **Implement Refreshable for custom refresh**

```go
type MyComponent struct {
    *tview.Box
    cachedStyles Styles
}

func (c *MyComponent) RefreshTheme() {
    c.cachedStyles = computeStyles()  // Recompute on theme change
}

func init() {
    theme.RegisterRefreshable(myComponent)
}
```

---

### Memory Leaks

**Symptom**: Memory usage grows over time, especially when navigating between views.

**Cause**: Subscriptions not unsubscribed, components not cleaned up.

**Solution**:

```go
type MyView struct {
    *components.ComponentBase
    subscriptions []func()  // Track unsubscribe functions
}

func (v *MyView) Start() {
    // Subscribe to values
    unsub := status.Subscribe(func(old, new string) {
        v.updateStatus(new)
    })
    v.subscriptions = append(v.subscriptions, unsub)

    // Subscribe to theme changes
    unsub2 := theme.OnThemeChange(func() {
        v.refreshStyles()
    })
    v.subscriptions = append(v.subscriptions, unsub2)
}

func (v *MyView) Stop() {
    // Unsubscribe all
    for _, unsub := range v.subscriptions {
        unsub()
    }
    v.subscriptions = nil
}
```

---

### Navigation Not Working

**Symptom**: Push/Pop don't work, or component lifecycle not called.

**Cause**: Blocking modal or incorrect component interface.

**Solutions**:

1. **Check for blocking modals**

```go
// Blocking modals prevent other navigation
modal := components.NewModal(ModalConfig{
    BlockUntilDismissed: true,  // This blocks Push/Pop!
})
```

2. **Ensure component implements nav.Component**

```go
type MyView struct {
    *components.ComponentBase  // This implements nav.Component
    // ...
}

// Or implement manually:
func (v *MyView) Start() {}
func (v *MyView) Stop() {}
func (v *MyView) Hints() []components.KeyHint { return nil }
// Plus all tview.Primitive methods
```

3. **Check Start/Stop are called**

Add logging to verify lifecycle:

```go
func (v *MyView) Start() {
    log.Printf("MyView.Start() called")
    v.loadData()
}

func (v *MyView) Stop() {
    log.Printf("MyView.Stop() called")
    v.cleanup()
}
```

---

### Form Validation Not Triggering

**Symptom**: Form submits without validating, or errors don't show.

**Solution**:

```go
// Make sure validators are added correctly
form := components.NewFormBuilder().
    Text("email", "Email").
        Validate(validators.Required(), validators.Email()).  // Add validators
        Done().
    OnSubmit(func(values map[string]any) {
        // Only called if validation passes
    }).
    Build()

// Check if form has errors before processing
if err := form.Validate(); err != nil {
    // Handle validation error
    return
}
```

---

## Debugging Techniques

### Logging

```go
import "log"

// Log in Start/Stop to trace lifecycle
func (v *MyView) Start() {
    log.Printf("[%s] Start() called", v.Name())
}

// Log state changes
func (v *MyView) SetData(data []Item) {
    log.Printf("[%s] SetData: %d items", v.Name(), len(data))
    v.data = data
}

// Log in Draw (careful - called frequently!)
func (c *MyComponent) Draw(screen tcell.Screen) {
    x, y, w, h := c.GetRect()
    log.Printf("Draw at (%d,%d) size %dx%d", x, y, w, h)
    c.Box.Draw(screen)
}
```

### State Inspection

Add a debug key to dump state:

```go
func (v *View) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyF12 {
        log.Printf("State dump:")
        log.Printf("  Stack depth: %d", app.Pages().StackDepth())
        log.Printf("  Current: %v", app.Pages().Current())
        log.Printf("  Has modal: %v", app.Pages().HasModal())
        return nil
    }
    return event
}
```

### Performance Profiling

```go
import "runtime/pprof"

// CPU profiling
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// Memory profiling
f, _ := os.Create("mem.prof")
defer func() {
    pprof.WriteHeapProfile(f)
    f.Close()
}()

// Analyze with:
// go tool pprof cpu.prof
// go tool pprof mem.prof
```

---

## FAQ

### Why doesn't Ctrl+C exit my app?

tview captures Ctrl+C by default. Handle it explicitly:

```go
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyCtrlC {
        app.Stop()
        return nil
    }
    return event
})
```

### How do I run shell commands from my app?

Use `app.Suspend()` to temporarily give up the terminal:

```go
import "os/exec"

app.Suspend(func() {
    cmd := exec.Command("vim", filename)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()
})
// App resumes automatically after function returns
```

### Can I use mouse input?

Yes, enable mouse support:

```go
app.GetApplication().EnableMouse(true)
```

Components implement `MouseHandler()` for mouse events.

### How do I show a loading indicator?

Use the async package with an indicator:

```go
async.NewLoader[Data]().
    WithIndicator(async.Toast("Loading...")).
    OnSuccess(func(data Data) { /* ... */ }).
    Run(fetchData)
```

### Why is my theme not applying to all components?

Components need to either:
1. Be registered with `theme.Register(primitive)`
2. Read theme colors at draw time, not creation time
3. Implement `Refreshable` interface for custom refresh logic

### How do I handle form cancellation?

```go
form := components.NewFormBuilder().
    Text("name", "Name").Done().
    OnSubmit(func(values map[string]any) {
        // Handle submit
        app.Pages().Pop()
    }).
    OnCancel(func() {
        // Handle cancel (e.g., user pressed Escape in form)
        app.Pages().Pop()
    }).
    Build()
```

### How do I update the menu hints dynamically?

```go
// Hints are updated automatically when component changes
// But you can update them manually:
app.UpdateMenuHints([]components.KeyHint{
    {Key: "Enter", Description: "Select"},
    {Key: "Esc", Description: "Back"},
})
```
