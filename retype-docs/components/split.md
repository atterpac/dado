---
label: Split
icon: columns
order: 45
---

# Split

Resizable split pane container.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

split := components.NewSplit().
    SetPrimary(leftPanel).
    SetSecondary(rightPanel)
```

---

## Configuration

```go
split := components.NewSplit().
    SetDirection(components.SplitHorizontal). // or SplitVertical
    SetPrimary(leftPanel).
    SetSecondary(rightPanel).
    SetRatio(0.3).      // 30% / 70% split
    SetMinSize(20).     // Minimum panel size in cells
    SetResizable(true)  // Allow runtime resizing
```

### Split Directions

| Direction | Description |
|-----------|-------------|
| `SplitHorizontal` | Side by side (left/right) |
| `SplitVertical` | Stacked (top/bottom) |

---

## Methods

| Method | Description |
|--------|-------------|
| `SetDirection(SplitDirection)` | Set split orientation |
| `SetPrimary(tview.Primitive)` | Set primary (left/top) pane |
| `SetSecondary(tview.Primitive)` | Set secondary (right/bottom) pane |
| `SetRatio(float64)` | Set primary pane ratio (0.0-1.0) |
| `SetMinSize(int)` | Set minimum pane size |
| `SetResizable(bool)` | Enable/disable resizing |
| `ToggleDirection()` | Switch between horizontal/vertical |
| `GetRatio()` | Get current ratio |
| `FocusPrimary()` | Focus primary pane |
| `FocusSecondary()` | Focus secondary pane |

---

## Example

```go
type EditorView struct {
    *components.ComponentBase
    split   *components.Split
    files   *components.Tree
    editor  *tview.TextArea
    app     *layout.App
}

func NewEditorView(app *layout.App) *EditorView {
    // File tree on the left
    files := components.NewTree()
    files.SetRoot(buildFileTree("."))

    // Editor on the right
    editor := tview.NewTextArea()
    editor.SetPlaceholder("Select a file to edit...")

    split := components.NewSplit().
        SetDirection(components.SplitHorizontal).
        SetPrimary(files).
        SetSecondary(editor).
        SetRatio(0.25).      // 25% file tree
        SetMinSize(15).
        SetResizable(true)

    v := &EditorView{
        split:  split,
        files:  files,
        editor: editor,
        app:    app,
    }

    files.SetOnSelect(func(node *components.TreeNode) {
        if path, ok := node.GetReference().(string); ok {
            v.loadFile(path)
        }
    })

    v.ComponentBase = components.NewComponentBase(split).
        SetName("editor").
        AddHint("Ctrl+B", "Toggle Sidebar").
        AddHint("Ctrl+\\", "Switch Pane").
        AddHint("Ctrl+S", "Save").
        SetInputHandler(v.handleInput)

    return v
}

func (v *EditorView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyCtrlB {
        // Toggle sidebar visibility
        if v.split.GetRatio() > 0 {
            v.split.SetRatio(0)
        } else {
            v.split.SetRatio(0.25)
        }
        return nil
    }

    if event.Key() == tcell.KeyCtrlBackslash {
        // Switch focus between panes
        // Implementation depends on current focus
        return nil
    }

    return event
}

func (v *EditorView) loadFile(path string) {
    content, err := os.ReadFile(path)
    if err != nil {
        return
    }
    v.editor.SetText(string(content), true)
    v.split.FocusSecondary()
}
```

---

## Nested Splits

```go
// Create a complex layout with nested splits
leftSplit := components.NewSplit().
    SetDirection(components.SplitVertical).
    SetPrimary(fileTree).
    SetSecondary(outline).
    SetRatio(0.6)

mainSplit := components.NewSplit().
    SetDirection(components.SplitHorizontal).
    SetPrimary(leftSplit).
    SetSecondary(editor).
    SetRatio(0.2)
```

---

## Keyboard Hints

When using splits, consider adding hints for:

| Key | Suggested Action |
|-----|------------------|
| `Ctrl+\` | Switch focus between panes |
| `Ctrl+B` | Toggle sidebar |
| `Ctrl+[` | Decrease primary ratio |
| `Ctrl+]` | Increase primary ratio |
