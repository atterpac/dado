---
title: Quick Start
description: Build a simple TUI dashboard in under 50 lines.
---

This guide walks you through a minimal dado app: a data grid themed with Nord.

## 1. Create a new module

```sh
mkdir my-tui && cd my-tui
go mod init example.com/my-tui
go get github.com/atterpac/dado@latest
```

## 2. Write the app

```go
package main

import (
    "fmt"

    "github.com/atterpac/dado"
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/theme"
)

type Process struct {
    Name   string
    PID    int
    Status string
}

var processes = []Process{
    {"nginx", 1001, "running"},
    {"postgres", 1042, "running"},
    {"redis", 2200, "sleeping"},
}

func main() {
    app := dado.New()
    app.SetTheme(theme.Nord)

    rows := make([][]string, len(processes))
    for i, p := range processes {
        rows[i] = []string{p.Name, fmt.Sprint(p.PID), p.Status}
    }

    grid := components.NewDataGrid().
        SetColumns([]string{"Name", "PID", "Status"}).
        SetRows(rows).
        OnActivate(func(row int) {
            // called when the user presses Enter on a row
        })

    app.SetRoot(grid, true)
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

## 3. Run it

```sh
go run .
```

A full-screen data grid with keyboard navigation (arrow keys / `j`/`k` to move,
`Enter` to activate a row).

## Next steps

- [Themes](/concepts/themes/) — switch or customise the colour scheme
- [Key Bindings](/concepts/key-bindings/) — add global and per-component shortcuts
- [Events](/concepts/events/) — respond to focus, change, and submit events
- [Components](/components/data-grid/) — explore the full catalog
