---
label: Data Binding
icon: link
order: 80
---

# Data Binding

Reactive data binding with observable values and automatic UI updates.

---

## Value[T] Observable

Observable values notify listeners when changed:

```go
import "github.com/atterpac/jig/binding"

// Create observable value
count := binding.NewValue(0)

// Subscribe to changes
unsubscribe := count.Subscribe(func(old, new int) {
    log.Printf("Count changed: %d -> %d", old, new)
})
defer unsubscribe()

// Update value (notifies listeners)
count.Set(1)

// Get current value
current := count.Get()
```

---

## Basic Operations

### Get and Set

```go
status := binding.NewValue("Ready")

// Get value
current := status.Get()  // "Ready"

// Set value (notifies listeners)
status.Set("Loading")

// Set and trigger redraw
status.SetAndDraw("Complete")
```

### Update with Function

```go
count := binding.NewValue(0)

// Atomic update
count.Update(func(n int) int {
    return n + 1
})

// Update and redraw
count.UpdateAndDraw(func(n int) int {
    return n * 2
})
```

---

## Subscribing to Changes

### Basic Subscription

```go
status := binding.NewValue("Ready")

unsubscribe := status.Subscribe(func(old, new string) {
    log.Printf("Status: %q -> %q", old, new)
})

// Later: unsubscribe to prevent memory leaks
unsubscribe()
```

### Subscribe with Draw

Automatically triggers UI redraw on change:

```go
counter := binding.NewValue(0)

unsubscribe := counter.SubscribeWithDraw(func(old, new int) {
    label.SetText(fmt.Sprintf("Count: %d", new))
})
```

---

## Binding to UI

### One-Way Binding

```go
connectionStatus := binding.NewValue("Disconnected")

unsubscribe := connectionStatus.BindToWithDraw(func(s string) {
    statusLabel.SetText("Status: " + s)
})
defer unsubscribe()

// Now any Set() call updates the label
connectionStatus.Set("Connecting...")
connectionStatus.Set("Connected")
```

### Example: Status Indicator

```go
type StatusIndicator struct {
    status      *binding.Value[string]
    label       *tview.TextView
    unsubscribe func()
}

func NewStatusIndicator() *StatusIndicator {
    status := binding.NewValue("Ready")
    label := tview.NewTextView()

    si := &StatusIndicator{
        status: status,
        label:  label,
    }

    si.unsubscribe = status.BindToWithDraw(func(s string) {
        color := theme.Success()
        switch s {
        case "Error":
            color = theme.Error()
        case "Warning":
            color = theme.Warning()
        case "Loading":
            color = theme.Info()
        }
        label.SetText(s).SetTextColor(color)
    })

    return si
}

func (si *StatusIndicator) SetStatus(s string) {
    si.status.Set(s)
}

func (si *StatusIndicator) Destroy() {
    si.unsubscribe()
}
```

---

## Computed Values

### Same Type Transformation

```go
firstName := binding.NewValue("John")

// Computed value updates automatically
upperFirst := firstName.Computed(strings.ToUpper)

firstName.Set("Jane")
fmt.Println(upperFirst.Get())  // "JANE"
```

### Type Transformation

```go
count := binding.NewValue(0)

// Transform int to string
displayText := binding.ComputedTo(count, func(n int) string {
    return fmt.Sprintf("Count: %d", n)
})

count.Set(5)
fmt.Println(displayText.Get())  // "Count: 5"
```

### Combining Values

```go
firstName := binding.NewValue("John")
lastName := binding.NewValue("Doe")

// Computed from multiple sources
fullName := binding.ComputedTo(firstName, func(first string) string {
    return first + " " + lastName.Get()
})

// Subscribe to both for full reactivity
firstName.Subscribe(func(_, _ string) {
    // fullName updates automatically
})
lastName.Subscribe(func(_, _ string) {
    // Need to manually trigger fullName recalculation
    firstName.Set(firstName.Get())
})
```

---

## Memory Management

!!!danger Important
Always unsubscribe to prevent memory leaks!
!!!

### Pattern: Track Subscriptions

```go
type MyView struct {
    *components.ComponentBase
    subscriptions []func()
}

func (v *MyView) Start() {
    unsub1 := status.Subscribe(func(old, new string) {
        v.updateStatus(new)
    })
    v.subscriptions = append(v.subscriptions, unsub1)

    unsub2 := count.SubscribeWithDraw(func(old, new int) {
        v.updateCount(new)
    })
    v.subscriptions = append(v.subscriptions, unsub2)
}

func (v *MyView) Stop() {
    for _, unsub := range v.subscriptions {
        unsub()
    }
    v.subscriptions = nil
}
```

---

## Form Binding

Two-way binding between structs and form fields:

```go
type User struct {
    Name  string `form:"name"`
    Email string `form:"email"`
    Role  string `form:"role"`
}

// Bind struct to form
user := &User{Name: "John", Email: "john@example.com", Role: "Admin"}
binding := binding.NewFormBinding(user, form)

// Changes to form update struct
// Changes to struct update form

// Get bound values
binding.Read()  // Updates struct from form
binding.Write() // Updates form from struct
```

---

## Table Binding

Bind a table to a slice of structs:

```go
type Item struct {
    Name   string `table:"Name"`
    Status string `table:"Status"`
    Count  int    `table:"Count"`
}

items := []Item{
    {Name: "Item 1", Status: "Active", Count: 10},
    {Name: "Item 2", Status: "Inactive", Count: 5},
}

table := components.NewTable()
binding := binding.NewTableBinding(items, table)

// Table automatically populated from items
// Changes to items update table
binding.Refresh()
```

---

## Practical Examples

### Loading State

```go
type LoadingState struct {
    IsLoading *binding.Value[bool]
    Error     *binding.Value[error]
    Data      *binding.Value[[]Item]
}

func NewLoadingState() *LoadingState {
    return &LoadingState{
        IsLoading: binding.NewValue(false),
        Error:     binding.NewValue[error](nil),
        Data:      binding.NewValue[[]Item](nil),
    }
}

func (s *LoadingState) Load(ctx context.Context) {
    s.IsLoading.SetAndDraw(true)
    s.Error.Set(nil)

    go func() {
        data, err := fetchData(ctx)

        app.QueueUpdateDraw(func() {
            s.IsLoading.Set(false)
            if err != nil {
                s.Error.Set(err)
            } else {
                s.Data.Set(data)
            }
        })
    }()
}
```

### Shopping Cart

```go
type CartItem struct {
    Name     string
    Price    float64
    Quantity *binding.Value[int]
}

type Cart struct {
    Items []*CartItem
    Total *binding.Value[float64]
}

func NewCart() *Cart {
    cart := &Cart{
        Total: binding.NewValue(0.0),
    }
    return cart
}

func (c *Cart) AddItem(name string, price float64) {
    item := &CartItem{
        Name:     name,
        Price:    price,
        Quantity: binding.NewValue(1),
    }

    // Update total when quantity changes
    item.Quantity.Subscribe(func(_, _ int) {
        c.recalculateTotal()
    })

    c.Items = append(c.Items, item)
    c.recalculateTotal()
}

func (c *Cart) recalculateTotal() {
    var total float64
    for _, item := range c.Items {
        total += item.Price * float64(item.Quantity.Get())
    }
    c.Total.SetAndDraw(total)
}
```

---

## Best Practices

1. **Always unsubscribe** - Track subscriptions and clean up in `Stop()`
2. **Use SetAndDraw** - When you need immediate visual feedback
3. **Batch updates** - Use `Set` for multiple updates, then call `app.Draw()` once
4. **Computed values** - Prefer computed values over manual synchronization
5. **Thread safety** - Value operations are thread-safe, but UI updates must use `QueueUpdateDraw`
