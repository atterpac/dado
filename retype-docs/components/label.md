---
label: Label
icon: typography
order: 100
---

# Label

Simple text display component with themed defaults.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

label := components.NewLabel("Hello, World!")
```

Or use the shorthand:

```go
label := components.Text("Hello, World!")
```

---

## Configuration

```go
label := components.NewLabel("Welcome to Jig").
    SetAlign(components.AlignCenter).
    SetColor(theme.Accent()).
    SetBold(true).
    SetWordWrap(true)
```

---

## Alignment

| Value | Description |
|-------|-------------|
| `AlignLeft` | Left-aligned text |
| `AlignCenter` | Center-aligned text (default) |
| `AlignRight` | Right-aligned text |

```go
label := components.NewLabel("Centered").
    SetAlign(components.AlignCenter)
```

---

## Methods

| Method | Description |
|--------|-------------|
| `SetText(string)` | Set the label text |
| `SetAlign(Align)` | Set text alignment |
| `SetColor(tcell.Color)` | Set text color |
| `SetBold(bool)` | Enable/disable bold |
| `SetWordWrap(bool)` | Enable/disable word wrapping |
| `SetScrollable(bool)` | Enable/disable scrolling |
| `SetDynamicColors(bool)` | Enable `[color]text[-]` syntax |
| `SetRegions(bool)` | Enable `["region"]text[""]` syntax |
| `Primitive()` | Get underlying `*tview.TextView` |

---

## Dynamic Colors

When dynamic colors are enabled (default), you can use color tags:

```go
label := components.NewLabel("[red]Error:[-] Something went wrong").
    SetDynamicColors(true)
```

Available colors: `red`, `green`, `blue`, `yellow`, `cyan`, `white`, `black`, and hex codes like `[#ff0000]`.

---

## Examples

### Simple Status Label

```go
status := components.NewLabel("Ready").
    SetColor(theme.Success())
```

### Multiline Content

```go
content := components.NewLabel(`
Welcome to Jig!

Press 'q' to quit.
Press 's' for settings.
`).SetWordWrap(true)
```

### In a Layout

```go
layout := components.Column(
    components.NewLabel("Header").SetBold(true),
    components.NewLabel("Content goes here"),
    components.NewLabel("Footer").SetColor(theme.FgDim()),
)
```

---

## Accessing the Primitive

For advanced usage, access the underlying `tview.TextView`:

```go
label := components.NewLabel("Text")
tv := label.Primitive()

// Use tview-specific features
tv.SetChangedFunc(func() {
    // Handle text changes
})
```
