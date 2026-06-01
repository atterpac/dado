---
title: Lifecycle
description: How dado components are constructed, configured, mounted, and torn down.
---

Every dado component follows the same four-phase lifecycle.

## 1. Construct

Instantiate with the `New` constructor:

```go
badge := components.NewBadge()
grid  := components.NewDataGrid()
modal := components.NewModal()
```

Constructors return a fully initialised component with sensible defaults. No
options are required.

## 2. Configure

Chain `Set*` methods before mounting:

```go
badge.
    SetLabel("OK").
    SetVariant(components.BadgeVariantSuccess)
```

All `Set*` methods return the receiver so chains compose cleanly.

## 3. Mount

Components satisfy `core.Widget`, so they slot directly into core layouts
with no adapter code:

```go
flex := core.NewFlex().
    AddItem(sidebar, 24, 0, false).
    AddItem(grid, 0, 1, true)
```

That layout becomes visible by living inside a **view** that you push onto the
app's page stack. See [Views & Navigation](/concepts/navigation/) for how to
wrap components in a view and move between screens.

## 4. Events

Wire callbacks after constructing:

```go
grid.OnActivate(func(row int) { … })
form.OnSubmit(func(e components.SubmitEvent) { … })
input.OnChange(func(e components.ChangeEvent) { … })
```

Callbacks are called on the draw goroutine. For background work use
`app.QueueUpdateDraw` to push updates back to the UI thread.

## ComponentBase and StatefulComponentBase

All dado widgets embed one of these two base types:

| Base | When used |
|---|---|
| `ComponentBase` | Stateless display components (Badge, Spinner, …) |
| `StatefulComponentBase` | Components with mutable internal state (DataGrid, Form, …) |

You rarely interact with these directly, but they appear in the API docs as the
source of shared methods (`SetBorder`, `SetTitle`, `SetTheme`, …).

## Teardown

dado components have no explicit teardown. When a component is removed from the
view tree and goes out of scope the Go garbage collector reclaims it. Long-lived
background goroutines spawned inside a component should be stopped explicitly
before unmounting.
