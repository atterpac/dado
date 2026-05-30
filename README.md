# Dado

> **Caution:** This is a personal project for building personal tools and may not work as expected.

A Go TUI component library built on [tview](https://github.com/rivo/tview) for creating terminal applications with consistent design patterns.

## Features

- **15+ UI Components** - Panels, modals, tables, forms, trees, tabs, and more
- **20+ Themes** - Tokyo Night, Catppuccin, Dracula, Nord, Gruvbox, etc.
- **Runtime Theme Switching** - Press `T` to change themes on the fly
- **Vim-style Navigation** - `j/k` movement, `g/G` for top/bottom
- **Data Binding** - Two-way form and table binding with struct tags
- **CLI Scaffolding** - `dado new <name>` to bootstrap projects

## Installation

```bash
go get github.com/atterpac/dado
```

## Quick Start

```go
package main

import (
    "github.com/atterpac/dado/layout"
    "github.com/atterpac/dado/theme"
)

func main() {
    theme.SetTheme("tokyonight")
    app := layout.NewApp("My App")
    app.Run()
}
```

## Components

Panels, Modals, Tables, Forms, Trees, Tabs, Text Fields, Checkboxes, Selects, Progress Bars, Splits, Splash Screens, Empty States, Key Hints

## Structure

```
components/   # UI primitives
theme/        # Theming system
layout/       # App layout management
nav/          # Navigation (pages, breadcrumbs)
input/        # Key bindings, command bar
binding/      # Data binding utilities
recipes/      # Pre-built templates
```

## License

MIT
