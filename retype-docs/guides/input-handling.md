---
label: Input Handling
icon: key
order: 75
---

# Input Handling

Keyboard input handling with KeyBindings builder and Vim navigation.

---

## Quick Start

The simplest way to handle input:

```go
import "github.com/atterpac/jig/input"

handler := input.NewKeyBindings().
    Bind('q', func() { app.Stop() }).
    Bind('r', func() { v.refresh() }).
    Bind(input.KeyEnter, func() { v.select() }).
    Bind(input.KeyEscape, func() { app.Pages().Pop() }).
    Build()

v.ComponentBase = components.NewComponentBase(content).
    SetInputHandler(handler)
```

Or use string-based binding for even simpler code:

```go
handler := input.NewKeyBindings().
    BindString("q", func() { app.Stop() }).
    BindString("enter", func() { v.select() }).
    BindString("escape", func() { cancel() }).
    BindString("ctrl+s", func() { save() }).
    Build()
```

---

## Input Event Chain

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

---

## Key Constants

Available key constants:

```go
import "github.com/atterpac/jig/input"

// Special keys
input.KeyEnter
input.KeyEscape  // or input.KeyEsc
input.KeyTab
input.KeyBacktab // Shift+Tab
input.KeyBackspace
input.KeyDelete
input.KeySpace

// Arrow keys
input.KeyUp
input.KeyDown
input.KeyLeft
input.KeyRight

// Navigation
input.KeyHome
input.KeyEnd
input.KeyPageUp
input.KeyPageDown

// Function keys
input.KeyF1 ... input.KeyF12

// Ctrl combinations
input.KeyCtrlS
input.KeyCtrlQ
input.KeyCtrlC
// ... KeyCtrlA through KeyCtrlZ
```

---

## Binding Methods

### Bind (Recommended)

The simplest approach - accepts keys or runes:

```go
handler := input.NewKeyBindings().
    Bind('q', quit).              // Character
    Bind('r', refresh).           // Character
    Bind(input.KeyEnter, submit). // Special key
    Bind(input.KeyCtrlS, save).   // Ctrl combination
    Build()
```

### BindString

Use strings to describe keys - great for configuration:

```go
handler := input.NewKeyBindings().
    BindString("q", quit).
    BindString("enter", submit).
    BindString("escape", cancel).
    BindString("ctrl+s", save).
    BindString("f1", showHelp).
    BindString("up", moveUp).
    Build()
```

Supported strings:
- Single characters: `"q"`, `"r"`, `"1"`, `"?"`, etc.
- Special keys: `"enter"`, `"escape"`, `"esc"`, `"tab"`, `"space"`, `"backspace"`, `"delete"`
- Arrow keys: `"up"`, `"down"`, `"left"`, `"right"`
- Navigation: `"home"`, `"end"`, `"pageup"`, `"pagedown"`
- Function keys: `"f1"` through `"f12"`
- Ctrl combinations: `"ctrl+s"`, `"ctrl+q"`, etc.

### Advanced: OnRune / On

For handlers that need the event object:

```go
handler := input.NewKeyBindings().
    OnRune('/', func(event *tcell.EventKey) bool {
        openSearch()
        return true  // Consumed
    }).
    On(input.KeyEnter, func(event *tcell.EventKey) bool {
        if event.Modifiers()&tcell.ModShift != 0 {
            submitAndContinue()
        } else {
            submit()
        }
        return true
    }).
    Build()
```

---

## Available Methods

### Simple Bindings (recommended)

| Method | Description |
|--------|-------------|
| `Bind(key, func())` | Bind key or rune to simple handler |
| `BindString(string, func())` | Bind using string key name |
| `BindCtrl(rune, func())` | Bind Ctrl+letter (e.g., `BindCtrl('s', save)`) |
| `BindShift(Key, func())` | Bind Shift+key |
| `BindAlt(Key, func())` | Bind Alt+key |

### Advanced (with event access)

| Method | Description |
|--------|-------------|
| `OnRune(rune, Handler)` | Handle character with event access |
| `On(Key, Handler)` | Handle special key with event access |
| `OnCtrlRune(rune, Handler)` | Handle Ctrl+character with event |
| `SetFallback(Handler)` | Handle unmatched keys |
| `Build()` | Build the handler function |

### Example

```go
handler := input.NewKeyBindings().
    Bind('q', quit).
    Bind('r', refresh).
    Bind(input.KeyEnter, submit).
    Bind(input.KeyEscape, cancel).
    BindCtrl('s', save).           // Ctrl+S
    BindCtrl('f', find).           // Ctrl+F
    BindShift(input.KeyTab, prev). // Shift+Tab
    Build()
```

---

## Vim Navigation

### List Navigation

Add vim-style navigation to lists:

```go
// j/k to move, gg/G for top/bottom, / for search
input.VimListBindings(v.list)
```

### Custom Vim Navigator

```go
bindings := input.NewKeyBindings().
    AddVimNavigation(input.VimNavigator{
        Up:     list.MoveUp,
        Down:   list.MoveDown,
        Left:   list.Collapse,
        Right:  list.Expand,
        Top:    list.MoveToTop,
        Bottom: list.MoveToBottom,
        Select: list.SelectCurrent,
        Back:   func() { app.Pages().Pop() },
    })
```

### Vim Keys

| Key | Action |
|-----|--------|
| `j` | Move down |
| `k` | Move up |
| `h` | Move left / collapse |
| `l` | Move right / expand |
| `g` | Go to top |
| `G` | Go to bottom |
| `Enter` | Select |
| `Esc` | Back |

---

## Action Registry

Named actions with key bindings:

```go
import "github.com/atterpac/jig/input"

actions := input.NewActionRegistry()

// Register actions
actions.Register(&input.Action{
    Name:        "save",
    Description: "Save changes",
    Key:         "Ctrl+S",
    Handler:     func() { save() },
})

actions.Register(&input.Action{
    Name:        "quit",
    Description: "Quit application",
    Key:         "q",
    Handler:     func() { app.Stop() },
})

actions.Register(&input.Action{
    Name:        "refresh",
    Description: "Refresh data",
    Key:         "r",
    Handler:     func() { refresh() },
})

// Bind additional keys to actions
actions.BindRune('R', "refresh")

// Get key hints for display
hints := actions.KeyHints()

// Execute action programmatically
actions.Execute("save")
```

---

## Key Hints

### Adding Hints

```go
v.ComponentBase = components.NewComponentBase(content).
    AddHint("r", "Refresh").
    AddHint("Enter", "Select").
    AddHint("q", "Quit")
```

### From Actions

```go
hints := actions.KeyHints()
v.ComponentBase = components.NewComponentBase(content).
    SetHints(hints)
```

### KeyHint Structure

```go
type KeyHint struct {
    Key         string  // Display text for key
    Description string  // Action description
}
```

---

## Practical Examples

### CRUD Operations

```go
func (v *ItemsView) setupInput() {
    v.ComponentBase = components.NewComponentBase(v.table).
        SetName("items").
        AddHint("n", "New").
        AddHint("e", "Edit").
        AddHint("d", "Delete").
        AddHint("r", "Refresh").
        AddHint("Enter", "View").
        AddHint("Esc", "Back").
        SetInputHandler(v.handleInput)
}

func (v *ItemsView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Rune() {
    case 'n':
        v.app.Pages().Push(NewItemForm(v.app, nil))
        return nil
    case 'e':
        if item := v.getSelectedItem(); item != nil {
            v.app.Pages().Push(NewItemForm(v.app, item))
        }
        return nil
    case 'd':
        if item := v.getSelectedItem(); item != nil {
            ConfirmDelete(v.app, item.Name, func() {
                v.deleteItem(item)
            })
        }
        return nil
    case 'r':
        v.loadItems()
        return nil
    }

    switch event.Key() {
    case tcell.KeyEnter:
        if item := v.getSelectedItem(); item != nil {
            v.app.Pages().Push(NewItemDetail(v.app, item))
        }
        return nil
    case tcell.KeyEscape:
        v.app.Pages().Pop()
        return nil
    }

    return event
}
```

### Search with Debounce

```go
type SearchView struct {
    *components.ComponentBase
    field    *components.TextField
    debounce *time.Timer
}

func NewSearchView(app *layout.App) *SearchView {
    field := components.NewTextField("search").
        SetPlaceholder("Type to search...")

    v := &SearchView{field: field}

    field.SetOnChange(func(e *components.ChangeEvent[string]) {
        // Cancel previous timer
        if v.debounce != nil {
            v.debounce.Stop()
        }

        // Start new timer
        v.debounce = time.AfterFunc(300*time.Millisecond, func() {
            app.QueueUpdateDraw(func() {
                v.search(e.NewValue)
            })
        })
    })

    v.ComponentBase = components.NewComponentBase(field).
        SetName("search").
        AddHint("Esc", "Clear")

    return v
}
```

### Mode-Based Input

```go
type EditorView struct {
    *components.ComponentBase
    mode string  // "normal" or "insert"
}

func (v *EditorView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if v.mode == "normal" {
        return v.handleNormalMode(event)
    }
    return v.handleInsertMode(event)
}

func (v *EditorView) handleNormalMode(event *tcell.EventKey) *tcell.EventKey {
    switch event.Rune() {
    case 'i':
        v.mode = "insert"
        v.updateStatusLine()
        return nil
    case 'j':
        v.moveCursorDown()
        return nil
    case 'k':
        v.moveCursorUp()
        return nil
    case ':':
        v.openCommandBar()
        return nil
    }
    return event
}

func (v *EditorView) handleInsertMode(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyEscape {
        v.mode = "normal"
        v.updateStatusLine()
        return nil
    }
    return event  // Pass through for text input
}
```

---

## Global Shortcuts

Set at the app level:

```go
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    // Global shortcuts
    switch event.Key() {
    case tcell.KeyCtrlQ:
        app.Stop()
        return nil
    case tcell.KeyCtrlT:
        cycleTheme()
        return nil
    }
    return event
})
```
