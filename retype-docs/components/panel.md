---
label: Panel
icon: container
order: 90
---

# Panel

Bordered container with rounded corners and optional title.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

content := tview.NewTextView().SetText("Hello, World!")

panel := components.NewPanel().
    SetTitle("Welcome").
    SetContent(content)
```

---

## Configuration

```go
panel := components.NewPanel().
    SetTitle("My Panel").
    SetTitleColor(theme.Accent()).
    SetTitleAlign(components.TitleAlignLeft).
    SetContent(content).
    SetFocused(true)  // Highlight border
```

### Title Alignment

| Value | Description |
|-------|-------------|
| `TitleAlignLeft` | Left-aligned title (default) |
| `TitleAlignCenter` | Center-aligned title |
| `TitleAlignRight` | Right-aligned title |

---

## Methods

| Method | Description |
|--------|-------------|
| `SetTitle(string)` | Set panel title |
| `SetTitleColor(tcell.Color)` | Set title text color |
| `SetTitleAlign(TitleAlign)` | Set title alignment |
| `SetContent(tview.Primitive)` | Set panel content |
| `SetFocused(bool)` | Highlight border when focused |
| `GetContent()` | Get the content primitive |

---

## Example

```go
type DetailView struct {
    *components.ComponentBase
    panel *components.Panel
}

func NewDetailView(title, text string) *DetailView {
    content := tview.NewTextView().
        SetText(text).
        SetTextAlign(tview.AlignLeft).
        SetDynamicColors(true)

    panel := components.NewPanel().
        SetTitle(title).
        SetContent(content)

    v := &DetailView{panel: panel}
    v.ComponentBase = components.NewComponentBase(panel).
        SetName("detail")

    return v
}
```

---

## Theme Integration

Panel automatically uses theme colors:

- `theme.PanelBorder()` for border color
- `theme.PanelTitle()` for title color
- `theme.BorderFocus()` when focused

To respond to theme changes:

```go
theme.Register(panel)  // Auto-update background
```
