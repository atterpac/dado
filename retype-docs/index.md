---
label: Home
icon: home
order: 100
---

# Jig

A Go TUI component library built on [tview](https://github.com/rivo/tview) for creating terminal applications with consistent design patterns.

!!!warning Caution
This is a personal project for building personal tools and may not work as expected.
!!!

---

## Features

+++ Components
### 15+ UI Components

Panels, modals, tables, forms, trees, tabs, progress bars, and more - all styled consistently and ready to use.

```go
panel := components.NewPanel().
    SetTitle("Welcome").
    SetContent(content)
```
+++ Themes
### 13+ Built-in Themes

Tokyo Night, Catppuccin, Dracula, Nord, Gruvbox, and more. Switch themes at runtime with a single line.

```go
theme.SetProvider(themes.TokyoNight())
```
+++ Navigation
### Vim-style Navigation

`j/k` movement, `g/G` for top/bottom. Stack-based page navigation with modal support.

```go
app.Pages().Push(NewDetailView(item))
app.Pages().Pop()
```
+++ Data Binding
### Reactive Data Binding

Observable values with automatic UI updates. Two-way form and table binding with struct tags.

```go
status := binding.NewValue("Ready")
status.SetAndDraw("Loading...")
```
+++

---

## Quick Start

### Installation

```bash
go get github.com/atterpac/jig
```

### Hello World

```go
package main

import (
    "log"

    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/layout"
    "github.com/atterpac/jig/theme"
    "github.com/atterpac/jig/theme/themes"
    "github.com/rivo/tview"
)

func main() {
    theme.SetProvider(themes.TokyoNight())

    app := layout.NewApp(layout.AppConfig{
        BottomBar: layout.NewMenu(),
    })

    app.Pages().Push(NewHomeView(app))

    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}

type HomeView struct {
    *components.ComponentBase
    app *layout.App
}

func NewHomeView(app *layout.App) *HomeView {
    content := tview.NewTextView().
        SetText("Hello, Jig!\n\nPress 'q' to quit.").
        SetTextAlign(tview.AlignCenter)

    panel := components.NewPanel().
        SetTitle("Welcome").
        SetContent(content)

    v := &HomeView{app: app}
    v.ComponentBase = components.NewComponentBase(panel).
        SetName("home").
        AddHint("q", "Quit")

    return v
}
```

---

## Package Structure

| Package | Description |
|---------|-------------|
| `components/` | UI primitives (Panel, Form, Table, etc.) |
| `theme/` | Theming system with 13+ built-in themes |
| `layout/` | App shell and layout management |
| `nav/` | Navigation (pages, breadcrumbs) |
| `input/` | Key bindings, Vim navigation |
| `binding/` | Data binding utilities |
| `async/` | Async operation helpers |
| `validators/` | Form validation |
| `recipes/` | Pre-built templates |

---

## Next Steps

[!ref icon="rocket" text="Getting Started"](getting-started.md)
[!ref icon="package" text="Components"](components/index.md)
[!ref icon="paintbrush" text="Theming"](guides/theming.md)
[!ref icon="book" text="Architecture"](guides/architecture.md)
