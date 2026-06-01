---
title: Data Binding
description: How to feed live data into dado components with SliceSource and Changeset.
---

dado provides two helpers for keeping component data in sync with your
application state: `SliceSource` for list-style data and `Changeset` for
tracking mutations.

## SliceSource

`SliceSource` wraps a slice and exposes it through the interface expected by
list components like `DataGrid` and `VirtualList`.

```go
type User struct {
    Name  string
    Email string
    Role  string
}

users := []User{…}

source := components.NewSliceSource(users, func(u User) []string {
    return []string{u.Name, u.Email, u.Role}
})

grid := components.NewDataGrid().
    SetColumns([]string{"Name", "Email", "Role"}).
    SetSource(source)
```

To update the data, mutate the slice and call `source.Refresh()`:

```go
users = append(users, newUser)
source.Refresh()
app.Draw()
```

## Changeset

`Changeset` tracks a set of mutations (adds, updates, deletes) and can replay
them into a component. Useful when you receive diffs from a server or database
watcher.

```go
cs := components.NewChangeset[User]()
cs.Add(newUser)
cs.Update(existingID, updatedUser)
cs.Delete(removedID)

source.Apply(cs)
app.Draw()
```

## Background updates

Always push data changes back to the draw goroutine via `QueueUpdateDraw`:

```go
go func() {
    for update := range stream {
        app.QueueUpdateDraw(func() {
            source.Refresh()
        })
    }
}()
```

Updating source data from a background goroutine without `QueueUpdateDraw` will
cause race conditions and potentially corrupt the display.
