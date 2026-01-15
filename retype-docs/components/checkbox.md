---
label: Checkbox
icon: checkbox
order: 70
---

# Checkbox

Boolean toggle with label.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

checkbox := components.NewCheckbox("active").
    SetLabel("Enable notifications").
    SetChecked(true)

// Get/set value
isChecked := checkbox.Value()  // bool
checkbox.SetChecked(false)
```

---

## Events

```go
checkbox := components.NewCheckbox("notify").
    SetLabel("Email notifications").
    SetOnChange(func(e *components.ChangeEvent[bool]) {
        log.Printf("Checked: %v -> %v", e.OldValue, e.NewValue)
    })
```

---

## Methods

| Method | Description |
|--------|-------------|
| `SetLabel(string)` | Set checkbox label |
| `SetChecked(bool)` | Set checked state |
| `Value()` | Get current state (bool) |
| `Toggle()` | Toggle checked state |
| `Clear()` | Set to unchecked |
| `HasValue()` | Check if checkbox has a value |
| `SetOnChange(func(*ChangeEvent[bool]))` | Change callback |

---

## With FormBuilder

```go
form := components.NewFormBuilder().
    Checkbox("notify", "Email notifications").
        Checked(true).
        Done().
    Checkbox("newsletter", "Subscribe to newsletter").
        Checked(false).
        Done().
    OnSubmit(func(values map[string]any) {
        notify := values["notify"].(bool)
        newsletter := values["newsletter"].(bool)
    }).
    Build()
```

---

## Example

```go
type SettingsView struct {
    *components.ComponentBase
    darkMode    *components.Checkbox
    autoSave    *components.Checkbox
    showTips    *components.Checkbox
}

func NewSettingsView(app *layout.App, settings *Settings) *SettingsView {
    darkMode := components.NewCheckbox("darkMode").
        SetLabel("Dark mode").
        SetChecked(settings.DarkMode)

    autoSave := components.NewCheckbox("autoSave").
        SetLabel("Auto-save").
        SetChecked(settings.AutoSave)

    showTips := components.NewCheckbox("showTips").
        SetLabel("Show tips").
        SetChecked(settings.ShowTips)

    v := &SettingsView{
        darkMode: darkMode,
        autoSave: autoSave,
        showTips: showTips,
    }

    darkMode.SetOnChange(func(e *components.ChangeEvent[bool]) {
        if e.NewValue {
            theme.SetProvider(themes.Dracula())
        } else {
            theme.SetProvider(themes.GitHub())
        }
    })

    flex := tview.NewFlex().SetDirection(tview.FlexRow).
        AddItem(darkMode, 1, 0, true).
        AddItem(autoSave, 1, 0, false).
        AddItem(showTips, 1, 0, false)

    panel := components.NewPanel().
        SetTitle("Settings").
        SetContent(flex)

    v.ComponentBase = components.NewComponentBase(panel).
        SetName("settings")

    return v
}
```

---

## Keyboard Navigation

| Key | Action |
|-----|--------|
| `Space` | Toggle checkbox |
| `Enter` | Toggle checkbox |
| `Tab` | Move to next field |
| `Shift+Tab` | Move to previous field |
