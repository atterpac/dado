---
label: Tabs
icon: browser
order: 50
---

# Tabs

Tabbed container for organizing content.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

tabs := components.NewTabs().
    AddTab("Overview", overviewContent).
    AddTab("Details", detailsContent).
    AddTab("Settings", settingsContent)
```

---

## Configuration

```go
tabs := components.NewTabs().
    AddTab("Tab 1", content1).
    AddTab("Tab 2", content2).
    AddTab("Tab 3", content3).
    SetActiveTab(0).
    SetOnTabChange(func(index int, name string) {
        log.Printf("Switched to tab: %s (index: %d)", name, index)
    })
```

---

## Methods

| Method | Description |
|--------|-------------|
| `AddTab(name string, content tview.Primitive)` | Add a new tab |
| `RemoveTab(index int)` | Remove tab at index |
| `SetActiveTab(index int)` | Switch to tab |
| `GetActiveTab()` | Get current tab index |
| `NextTab()` | Switch to next tab |
| `PrevTab()` | Switch to previous tab |
| `SetOnTabChange(func(int, string))` | Tab change callback |
| `GetTabCount()` | Get number of tabs |

---

## Example

```go
type DetailView struct {
    *components.ComponentBase
    tabs *components.Tabs
    app  *layout.App
}

func NewDetailView(app *layout.App, item *Item) *DetailView {
    // Overview tab
    overview := tview.NewTextView().
        SetText(fmt.Sprintf("Name: %s\nStatus: %s\nCreated: %s",
            item.Name, item.Status, item.CreatedAt))

    // Properties tab
    properties := components.NewTable()
    properties.SetHeaders("Key", "Value")
    for k, v := range item.Properties {
        properties.AddRow(k, v)
    }

    // History tab
    history := components.NewTable()
    history.SetHeaders("Date", "Action", "User")
    for _, h := range item.History {
        history.AddRow(h.Date, h.Action, h.User)
    }

    tabs := components.NewTabs().
        AddTab("Overview", overview).
        AddTab("Properties", properties).
        AddTab("History", history)

    panel := components.NewPanel().
        SetTitle(item.Name).
        SetContent(tabs)

    v := &DetailView{tabs: tabs, app: app}
    v.ComponentBase = components.NewComponentBase(panel).
        SetName("detail").
        AddHint("Tab", "Next Tab").
        AddHint("Shift+Tab", "Prev Tab").
        AddHint("Esc", "Back").
        SetInputHandler(v.handleInput)

    return v
}

func (v *DetailView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Key() {
    case tcell.KeyTab:
        v.tabs.NextTab()
        return nil
    case tcell.KeyBacktab:
        v.tabs.PrevTab()
        return nil
    case tcell.KeyEscape:
        v.app.Pages().Pop()
        return nil
    }

    // Number keys to switch tabs directly
    if event.Rune() >= '1' && event.Rune() <= '9' {
        index := int(event.Rune() - '1')
        if index < v.tabs.GetTabCount() {
            v.tabs.SetActiveTab(index)
        }
        return nil
    }

    return event
}
```

---

## Keyboard Navigation

| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `1`-`9` | Switch to tab by number |
