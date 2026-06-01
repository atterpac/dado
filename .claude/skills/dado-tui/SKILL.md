---
name: dado-tui
description: >-
  Guide for building terminal UIs with the dado Go TUI library (github.com/atterpac/dado,
  built on rivo/tview + gdamore/tcell). Use when creating TUI apps, authoring or wiring
  dado components, working with its theme system, key bindings, navigation/page stack,
  data binding, or async/background work in a dado-based app. Covers the component
  catalog, the core architectural patterns, and the conventions that keep theming and
  threading correct.
---

# Building TUIs with dado

`dado` is a Go TUI component toolkit layered over [`rivo/tview`](https://github.com/rivo/tview) and `gdamore/tcell/v2`. It adds: a scoped theme system with live switching, ~50 components, a page-stack navigator, fluent key bindings, two-way data binding, and async helpers. Use this skill whenever you build or modify a dado-based TUI.

> Module path: `github.com/atterpac/dado`. Requires Go 1.24+.

## When to use this skill

- Starting a new dado TUI app (assemble the shell, push views).
- Authoring a new custom component/widget that draws itself.
- Wiring theme colors, key bindings, navigation, modals, forms, or tables.
- Running background work and updating the UI safely.

## The five load-bearing rules

These are the conventions everything else depends on. Get them right and the rest is mechanical.

1. **Read theme colors at draw time — never cache them in a field.**
   Live theme switching works *because* every `Draw` re-reads the active theme. Read via the component's scoped accessor (`w.th().Bg()` for widgets, `cb.Theme().Bg()` for views), or the package forwarders (`theme.Bg()`, which read `theme.Default()`). Storing a `tcell.Color` in a struct field breaks switching.

2. **All UI mutation happens on the draw thread — marshal via `theme.QueueUpdateDraw(func(){...})`.**
   Do background work on a goroutine, then queue the UI update. `async` and `effect` deliver their callbacks on the UI thread for you; `bus` handlers and `util.TaskRunner` callbacks do **not** — wrap those yourself.

3. **Register cleanup into a `Subscriptions`, exposed via `Subs()`.**
   Every theme registration / subscription returns an unregister `func()`. Add it to the component's `Subscriptions`. `ComponentBase.Stop()` releases its own subs *and* the wrapped widget's subs automatically (LIFO, idempotent). This is how teardown stays leak-free.

4. **Two input-handler conventions exist — pick by what you attach to.**
   - tview-native primitives expect the **event** back (`nil` = consumed): use `KeyBindings.Build()`.
   - dado `ComponentBase.SetInputHandler` expects a **bool** (`true` = consumed): use `KeyBindings.BuildBool()`.
   This is the most common footgun. See `reference/input-nav-layout.md`.

5. **Pick the right base type for what you're building.**
   - Authoring a self-drawing leaf widget → embed `widgetBase`, call `initWidget`.
   - Wrapping a finished primitive as a navigable view → hold a `*ComponentBase`.
   - A view backed by an async fetch (loading/error/success) → `*StatefulComponentBase[T]`.
   See `reference/architecture.md`.

## Minimal app skeleton

```go
package main

import (
    "github.com/atterpac/dado/layout"
    "github.com/atterpac/dado/theme"
    "github.com/atterpac/dado/theme/themes"
)

func main() {
    theme.Default().SetTheme(themes.Get("tokyonight-night"))

    app := layout.NewApp(layout.AppConfig{
        TopBar:     myHeader(),   // any tview.Primitive (optional)
        ShowCrumbs: true,
        BottomBar:  layout.NewMenu(), // hint bar; auto-updates from each view's Hints()
        Debug:      true,             // Ctrl+D event-inspector overlay
    })
    app.EnableThemes(layout.ThemeOptions{}) // Ctrl+T live theme picker, zero per-component wiring

    app.Pages().Push(NewHomeView()) // a nav.Component
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

The app shell stacks rows: TopBar → Crumbs → Pages (main area) → BottomBar. Navigation is a **stack** (`Push`/`Pop`/`Replace`), not URL routes. Pushing a view `Stop()`s the previous top and `Start()`s the new one, updates breadcrumbs, and pushes its `Hints()` into the menu.

## Component selection cheat-sheet

| Need | Reach for |
|---|---|
| Frame/label any child | `Panel` |
| Centered dialog / confirm | `Modal` (presets in `modal_presets.go`) |
| Side / bottom slide-out | `Drawer`, `BottomSheet` |
| Tabular data + selection | `Table`; huge/streamed → `VirtualList`; editable grid → `DataGrid` |
| Hierarchy | `Tree` (lazy-load capable) |
| Form entry | `FormBuilder` (fluent) or `Form` (quick) + `binding.FormBinding` |
| Single/multi field input | `TextField`, `TextArea`, `Select`, `MultiSelect`, `Checkbox`, `RadioGroup` |
| Fuzzy pick / palette | `Finder`, `Autocomplete`, `input.CommandBar` |
| Charts / metrics | `LineGraph`, `BarChart`, `Sparkline`, `Gauge`, `HeatMap`, `MetricCard` |
| Transient feedback | `ToastManager`; long op → `ProgressModal` |
| Whole pre-built view | `recipes.Dashboard`, `recipes.LogViewer`, `recipes.ResourceList[T]` |

Full catalog with constructors, methods, and use cases: `reference/components.md`.

## Reference files

Read the one matching your task:

- **`reference/architecture.md`** — `ComponentBase` / `widgetBase` / `StatefulComponentBase[T]`, the Draw prepare+paint split, `Subscriptions`, value/event/handler interfaces, and the canonical recipe for authoring a new component.
- **`reference/theme.md`** — `Provider` / `Default()`, the full color-accessor palette, runtime switching & subscriptions, status colors, Nerd Font icons, the theme YAML schema, programmatic builder, and gradients.
- **`reference/input-nav-layout.md`** — `KeyBindings` (Build vs BuildBool), vim helpers, `ActionRegistry`, `CommandBar`, the `nav` page stack / modals / breadcrumbs, the `layout` app shell / menu / statusbar, and the `binding` package (`Value[T]`, `FormBinding[T]`, `TableBinding[T]`).
- **`reference/components.md`** — the full tiered component catalog (basic / intermediate / advanced).
- **`reference/support.md`** — `async`, `effect`, `bus`, `style`, `help`, `validators`, `util`, `clipboard`, `recipes`, `testutil`, and the `dado` CLI.

## Quick gotchas

- `components/README.md` documents stale `theme.HasStatus`/`theme.StatusColor` — those don't exist. Use typed `*Status` handles (`theme.DefineStatus`). See `reference/theme.md`.
- `style.BorderSet` custom glyphs are largely a no-op through `Apply` — tview draws borders from its global rune set. Hand-draw borders in your own `Draw` if you need custom glyphs.
- The `dado` CLI does **not** scaffold projects (`dado new` does not exist). It only lists/previews themes and lists components. Bootstrap by hand.
- There are two different `ListNavigator` interfaces (one in `input`, one in `nav`) — don't conflate them.
