---
label: TextArea
icon: note
order: 35
---

# TextArea

Multi-line text input.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

area := components.NewTextArea("description").
    SetLabel("Description").
    SetPlaceholder("Enter description...")
```

---

## Configuration

```go
area := components.NewTextArea("notes").
    SetLabel("Notes").
    SetPlaceholder("Add your notes here...").
    SetMaxLines(10).
    SetValue("Initial content\nwith multiple lines")
```

---

## Methods

| Method | Description |
|--------|-------------|
| `SetLabel(string)` | Set field label |
| `SetPlaceholder(string)` | Set placeholder text |
| `SetValue(string)` | Set text content |
| `Value()` | Get current content |
| `SetMaxLines(int)` | Set maximum line count |
| `Clear()` | Clear the field |
| `HasValue()` | Check if field has content |
| `SetOnChange(func(*ChangeEvent[string]))` | Change callback |

---

## Events

```go
area := components.NewTextArea("bio").
    SetOnChange(func(e *components.ChangeEvent[string]) {
        lines := strings.Count(e.NewValue, "\n") + 1
        log.Printf("Lines: %d", lines)
    })
```

---

## With FormBuilder

```go
form := components.NewFormBuilder().
    Text("title", "Title").
        Validate(validators.Required()).
        Done().
    TextArea("body", "Body").
        Placeholder("Write your content...").
        MaxLines(20).
        Done().
    OnSubmit(func(values map[string]any) {
        title := values["title"].(string)
        body := values["body"].(string)
        log.Printf("Title: %s\nBody: %s", title, body)
    }).
    Build()
```

---

## Example: Note Editor

```go
type NoteEditor struct {
    *components.ComponentBase
    title   *components.TextField
    content *components.TextArea
    app     *layout.App
    note    *Note
}

func NewNoteEditor(app *layout.App, note *Note) *NoteEditor {
    title := components.NewTextField("title").
        SetLabel("Title").
        SetValue(note.Title)

    content := components.NewTextArea("content").
        SetLabel("Content").
        SetValue(note.Content).
        SetMaxLines(50)

    // Auto-save on change
    var saveTimer *time.Timer
    save := func() {
        note.Title = title.Value()
        note.Content = content.Value()
        note.Save()
    }

    autoSave := func() {
        if saveTimer != nil {
            saveTimer.Stop()
        }
        saveTimer = time.AfterFunc(1*time.Second, save)
    }

    title.SetOnChange(func(e *components.ChangeEvent[string]) {
        autoSave()
    })

    content.SetOnChange(func(e *components.ChangeEvent[string]) {
        autoSave()
    })

    flex := tview.NewFlex().SetDirection(tview.FlexRow).
        AddItem(title, 3, 0, true).
        AddItem(content, 0, 1, false)

    panel := components.NewPanel().
        SetTitle("Edit Note").
        SetContent(flex)

    v := &NoteEditor{
        title:   title,
        content: content,
        app:     app,
        note:    note,
    }

    v.ComponentBase = components.NewComponentBase(panel).
        SetName("note-editor").
        AddHint("Ctrl+S", "Save").
        AddHint("Esc", "Back").
        SetInputHandler(v.handleInput)

    return v
}

func (v *NoteEditor) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyCtrlS {
        v.note.Title = v.title.Value()
        v.note.Content = v.content.Value()
        v.note.Save()
        components.ShowToast(v.app, "Saved!", 1*time.Second)
        return nil
    }

    if event.Key() == tcell.KeyEscape {
        v.app.Pages().Pop()
        return nil
    }

    return event
}
```

---

## Keyboard Navigation

| Key | Action |
|-----|--------|
| `Tab` | Move to next field |
| `Shift+Tab` | Move to previous field |
| `Enter` | New line |
| Standard text editing shortcuts | Cut, copy, paste, etc. |
