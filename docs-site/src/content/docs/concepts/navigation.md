---
title: Views & Navigation
description: How to build navigable views and move between them with the page stack.
---

A **view** is any screen the app can navigate to. Components (Badge, DataGrid,
Form, …) are the building blocks *inside* a view; a view is the thing you push
onto the navigation stack.

## What makes a view

dado components satisfy `core.Widget`, but that alone is **not** enough to be
navigable. A view must implement `nav.Component`:

```go
type Component interface {
    core.Widget       // it can render and receive input

    Name() string                 // label shown in breadcrumbs
    Start()                       // called when the view becomes active
    Stop()                        // called when the view is hidden
    Hints() []components.KeyHint  // key hints shown in the bottom bar
}
```

dado components do **not** implement these four methods on their own, so you
write a small view type that wraps them. The simplest way to satisfy
`core.Widget` is to embed an existing primitive — usually a `*core.Flex` —
and then add the four methods:

```go
package main

import (
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/core"
)

// HomeView embeds *components.ComponentBase, which implements the whole
// nav.Component contract for you: Name/Start/Stop/Hints come for free, and
// rendering is delegated to the wrapped core.Widget — here a *core.Flex.
// Configure it fluently in the constructor; no lifecycle stubs needed.
type HomeView struct {
    *components.ComponentBase
}

func NewHomeView() *HomeView {
    flex := core.NewFlex().SetDirection(core.Row)

    // Compose components into the view.
    flex.AddItem(components.NewBadge("Ready").SetVariant(components.BadgeSuccess), 0, 1, false)
    flex.AddItem(components.NewBadge("3").SetVariant(components.BadgeError).SetPill(true), 0, 1, false)

    return &HomeView{
        ComponentBase: components.NewComponentBase(flex).
            SetName("Home").
            AddHint("Enter", "Open").
            AddHint("q", "Quit"),
    }
}
```

`ComponentBase` also exposes `SetOnStart`/`SetOnStop` hooks (e.g. to kick off
a data load) and releases any registered subscriptions automatically on
`Stop()`.

:::note
`nav.Component` is the underlying interface every pushed view must satisfy:

```go
type Component interface {
    // core.Widget rendering surface
    Draw(screen tcell.Screen)
    SetRect(x, y, w, h int)
    Blur()
    HasFocus() bool

    Name() string                // breadcrumb label
    Start()                      // becomes active
    Stop()                       // becomes inactive
    Hints() []components.KeyHint  // feeds the menu bar
}
```

Embedding `*components.ComponentBase` is the recommended way to satisfy it —
but you can implement these four methods by hand on any `core.Widget` (a
single component, a `core.Pages`, a custom layout) when you need full control.
:::

## The page stack

The app owns a stack of views, reached through `app.Pages()`. Pushing a view
makes it the active screen; popping returns to the previous one.

```go
app := layout.NewApp(layout.AppConfig{ShowCrumbs: true})
app.Pages().Push(NewHomeView())

if err := app.Run(); err != nil {
    log.Fatal(err)
}
```

`Pages()` exposes the full stack API:

| Method | Effect |
|---|---|
| `Push(c Component)` | Push `c` as the new top view |
| `PushFactory(f ComponentFactory)` | Push a view built lazily by `f` (constructed on entry) |
| `Pop()` | Remove the top view, returning to the one beneath |
| `Replace(c Component)` | Swap the top view without growing the stack |
| `Clear()` | Empty the stack |
| `Depth() int` | Number of views currently stacked |
| `Current() Component` | The active (top) view |
| `OnStackChange(fn func(depth int))` | Run `fn` whenever the depth changes |

A typical drill-down flow:

```go
func (v *ListView) Hints() []components.KeyHint {
    return []components.KeyHint{{Key: "Enter", Description: "Details"}}
}

// when a row is activated:
v.app.Pages().Push(NewDetailView(selected))

// inside the detail view, going back:
v.app.Pages().Pop()
```

## Start and Stop

`Start()` runs when a view becomes active (including when it is revealed again
after a `Pop`), and `Stop()` runs when it is hidden. Use them to manage anything
tied to the view being on screen — background polling, tickers, subscriptions:

```go
func (v *FeedView) Start() {
    v.cancel = make(chan struct{})
    go v.pollFeed(v.cancel) // updates via app.QueueUpdateDraw
}

func (v *FeedView) Stop() {
    close(v.cancel) // tear down the goroutine when navigated away
}
```

Stateless views can leave both methods empty.

## Breadcrumbs

When `AppConfig.ShowCrumbs` is `true`, the top bar shows a breadcrumb trail
built from each view's `Name()` as the stack grows and shrinks — `Home › Users ›
Details` — giving users a sense of where they are with no extra wiring.

## Persistent chrome

A view fills the region between the optional top and bottom bars. Set those once
on the app and they persist across every navigation:

```go
app := layout.NewApp(layout.AppConfig{
    TopBar:     myStatusBar,   // any core.Widget
    BottomBar:  myMenu,
    ShowCrumbs: true,
})
```

The bottom bar is where each view's `Hints()` are surfaced, so navigating
between views automatically updates the visible key hints.
