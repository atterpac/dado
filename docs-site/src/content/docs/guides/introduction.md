---
title: Introduction
description: What dado is, what it provides, and when to reach for it.
---

**dado** is a Go library for building terminal UIs. It provides a curated set of
ready-made components — data grids, charts, modals, forms, graphs, and more — built
on its own widget core directly over
[gdamore/tcell](https://github.com/gdamore/tcell).

## What dado gives you

| Feature | Detail |
|---|---|
| **Component catalog** | 26+ production-ready widgets |
| **Theme system** | 26 built-in themes, runtime switching |
| **Event system** | Typed focus / key / change / submit events |
| **Layout helpers** | Flex and grid layouts built on the `core` widget primitives |
| **Data binding** | `SliceSource` and `Changeset` for live data updates |

## When to use dado

dado is a good fit when you need:

- A polished internal tool, CLI dashboard, or DevOps TUI
- Consistent theming across many components without hand-rolling styles
- A component you'd otherwise spend days building from raw tcell primitives

dado components are plain `core.Widget`s. You can drop down to the `core`
package (`Box`, `Flex`, `TextView`, …) and mix hand-rolled widgets with the
catalog freely — everything renders through the same `Draw`/`SetRect` surface.

## What dado is not

- A web framework — dado is terminal-only
- An abstraction over multiple TUI backends — dado targets tcell

## Architecture overview

```
your app
  └── dado.App          (owns the tcell screen + event loop)
       ├── theme.Provider   (resolves colours, fonts)
       ├── navigation.Stack (push/pop page stack)
       └── components.*     (the widget catalog)
```

Each component follows the same lifecycle:

1. **Construct** — `components.NewFoo()`
2. **Configure** — chain `Set*` methods
3. **Mount** — add to a layout or set as root
4. **Event** — react via `OnChange`, `OnSubmit`, `OnFocus`, …

See [Lifecycle](/concepts/lifecycle/) for the full picture.
