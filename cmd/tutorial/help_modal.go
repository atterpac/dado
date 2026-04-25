package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
)

// HelpModal displays help information.
type HelpModal struct {
	*components.Modal
	content *tview.TextView
	subs    components.Subscriptions
}

// Subs returns the modal's subscription set; release on dismissal.
func (m *HelpModal) Subs() *components.Subscriptions { return &m.subs }

// NewHelpModal creates a new help modal.
func NewHelpModal() *HelpModal {
	content := tview.NewTextView()
	content.SetDynamicColors(true)
	content.SetScrollable(true)
	content.SetBackgroundColor(theme.Bg())
	content.SetTextColor(theme.Fg())

	helpText := `[::b]Jig Tutorial - Keyboard Shortcuts[::-]

[` + theme.TagAccent() + `]Navigation[::-]
  [` + theme.TagFgDim() + `]j / k[` + theme.TagFg() + `]       Move down / up in list
  [` + theme.TagFgDim() + `]h / l[` + theme.TagFg() + `]       Collapse / expand categories
  [` + theme.TagFgDim() + `]Enter[` + theme.TagFg() + `]       Select component
  [` + theme.TagFgDim() + `]Tab[` + theme.TagFg() + `]         Switch between sidebar and demo

[` + theme.TagAccent() + `]Actions[::-]
  [` + theme.TagFgDim() + `]p[` + theme.TagFg() + `]           Open property editor
  [` + theme.TagFgDim() + `]c[` + theme.TagFg() + `]           View code example
  [` + theme.TagFgDim() + `]t[` + theme.TagFg() + `]           Open theme picker
  [` + theme.TagFgDim() + `]?[` + theme.TagFg() + `]           Show this help
  [` + theme.TagFgDim() + `]q[` + theme.TagFg() + `]           Quit application

[` + theme.TagAccent() + `]In Modals[::-]
  [` + theme.TagFgDim() + `]Tab[` + theme.TagFg() + `]         Next field
  [` + theme.TagFgDim() + `]j / k[` + theme.TagFg() + `]       Scroll / navigate
  [` + theme.TagFgDim() + `]Space[` + theme.TagFg() + `]       Toggle / select
  [` + theme.TagFgDim() + `]Esc[` + theme.TagFg() + `]         Close modal

[::d]Press Esc to close this help[::-]`

	content.SetText(helpText)

	modal := components.NewModal(components.ModalConfig{
		Title:     "Help",
		Width:     60,
		Height:    22,
		MinWidth:  40,
		MinHeight: 16,
		Backdrop:  true,
	})
	modal.SetContent(content)
	modal.SetDismissOnEsc(true)
	modal.SetFocusOnShow(content)

	// Set up hints
	modal.SetHints([]components.KeyHint{
		{Key: "j/k", Description: "Scroll"},
		{Key: "Esc", Description: "Close"},
	})

	m := &HelpModal{
		Modal:   modal,
		content: content,
	}

	m.subs.Add(theme.Register(content))

	return m
}

// Start implements nav.Component.
func (m *HelpModal) Start() {}

// Stop implements nav.Component.
func (m *HelpModal) Stop() { m.subs.Release() }

// Hints implements nav.Component.
func (m *HelpModal) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "Esc", Description: "Close"},
	}
}

// Focus handles focus.
func (m *HelpModal) Focus(delegate func(tview.Primitive)) {
	delegate(m.content)
}

// HasFocus returns focus state.
func (m *HelpModal) HasFocus() bool {
	return m.content.HasFocus()
}

// InputHandler handles keyboard input.
func (m *HelpModal) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				row, _ := m.content.GetScrollOffset()
				m.content.ScrollTo(row+1, 0)
				return
			case 'k':
				row, _ := m.content.GetScrollOffset()
				if row > 0 {
					m.content.ScrollTo(row-1, 0)
				}
				return
			}
		}

		// Pass to modal for Escape handling
		if handler := m.Modal.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}
