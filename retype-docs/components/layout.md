---
label: Layout
icon: rows
order: 92
---

# Layout

Simple layout helpers that abstract `tview.Flex` complexity.

---

## Quick Start

```go
import "github.com/atterpac/jig/components"

// Vertical layout (items stacked top to bottom)
layout := components.Column(
    header,
    content,
    footer,
)

// Horizontal layout (items arranged left to right)
layout := components.Row(
    sidebar,
    mainContent,
)
```

---

## Column Layout

Stack items vertically:

```go
layout := components.Column(
    components.NewLabel("Title").SetBold(true),
    components.NewLabel("Description"),
    table,
    buttonRow,
)
```

---

## Row Layout

Arrange items horizontally:

```go
layout := components.Row(
    components.NewLabel("Name:"),
    nameField,
    components.NewButton("Save"),
)
```

---

## Builder Pattern

For more control, use the builder API:

```go
// Column with fixed and weighted items
layout := components.NewColumn().
    Fixed(header, 3).           // Fixed height of 3
    Add(content).               // Flexible, equal weight
    Weighted(sidebar, 2).       // 2x weight
    FixedSpacer(1).            // Empty space
    Fixed(footer, 1).          // Fixed height of 1
    Build()

// Row with specific weights
layout := components.NewRow().
    Fixed(icon, 4).            // Fixed width of 4
    Add(label).                // Flexible
    Spacer().                  // Push remaining to right
    Fixed(button, 10).         // Fixed width of 10
    Build()
```

---

## Methods

### Layout

| Method | Description |
|--------|-------------|
| `Horizontal()` | Set direction to horizontal (row) |
| `Vertical()` | Set direction to vertical (column) |
| `Add(Primitive)` | Add item with equal weight |
| `AddFixed(Primitive, size)` | Add item with fixed size |
| `AddWeighted(Primitive, weight)` | Add item with specific weight |
| `AddFocused(Primitive)` | Add item that receives focus |
| `AddSpacer()` | Add flexible empty space |
| `AddFixedSpacer(size)` | Add fixed empty space |
| `Primitive()` | Get underlying `*tview.Flex` |

### RowBuilder / ColumnBuilder

| Method | Description |
|--------|-------------|
| `Add(Primitive)` | Add flexible item |
| `Fixed(Primitive, size)` | Add fixed-size item |
| `Weighted(Primitive, weight)` | Add weighted item |
| `Focused(Primitive)` | Add item with focus |
| `Spacer()` | Add flexible spacer |
| `FixedSpacer(size)` | Add fixed spacer |
| `Build()` | Return the completed Layout |

---

## Examples

### App Layout

```go
layout := components.NewColumn().
    Fixed(topBar, 1).
    Fixed(breadcrumbs, 1).
    Add(mainContent).
    Fixed(statusBar, 1).
    Build()
```

### Sidebar Layout

```go
layout := components.NewRow().
    Fixed(sidebar, 30).
    Add(content).
    Build()
```

### Form with Buttons

```go
form := components.NewColumn().
    Add(nameField).
    Add(emailField).
    FixedSpacer(1).
    Fixed(
        components.NewRow().
            Spacer().
            Fixed(cancelBtn, 10).
            FixedSpacer(1).
            Fixed(saveBtn, 10).
            Build(),
        1,
    ).
    Build()
```

### Center Content

```go
centered := components.NewRow().
    Spacer().
    Add(content).
    Spacer().
    Build()
```

### Dashboard Grid

```go
dashboard := components.NewColumn().
    Add(components.NewRow().
        Add(statsPanel).
        Add(chartPanel).
        Build()).
    Add(components.NewRow().
        Add(tablePanel).
        Add(activityPanel).
        Build()).
    Build()
```

---

## Nesting Layouts

Layouts can be nested for complex UIs:

```go
app := components.Column(
    // Top bar
    components.Row(logo, spacer, userMenu),
    // Main content
    components.Row(
        sidebar,
        components.Column(
            breadcrumbs,
            content,
        ),
    ),
    // Bottom bar
    statusBar,
)
```

---

## Accessing the Primitive

For advanced tview features:

```go
layout := components.Column(items...)
flex := layout.Primitive()

// Use tview-specific features
flex.SetFullScreen(true)
```
