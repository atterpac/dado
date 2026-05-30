package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

// ThemePicker is a modal for selecting themes.
type ThemePicker struct {
	*components.Modal
	list     *tview.List
	onSelect func(t theme.Theme)
	subs     components.Subscriptions
}

// Subs returns the picker's subscription set; release on dismissal.
func (p *ThemePicker) Subs() *components.Subscriptions { return &p.subs }

// NewThemePicker creates a new theme picker modal.
func NewThemePicker(onSelect func(t theme.Theme)) *ThemePicker {
	list := tview.NewList()
	list.SetBackgroundColor(theme.Bg())
	list.SetMainTextColor(theme.Fg())
	list.SetSelectedBackgroundColor(theme.Accent())
	list.SetSelectedTextColor(theme.Bg())
	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)

	// Add all themes
	themeNames := themes.Names()
	for _, name := range themeNames {
		list.AddItem(name, "", 0, nil)
	}

	modal := components.NewModal(components.ModalConfig{
		Title:     "Select Theme",
		Width:     40,
		Height:    20,
		MinWidth:  30,
		MinHeight: 10,
		Backdrop:  true,
	})
	modal.SetContent(list)
	modal.SetDismissOnEsc(true)
	modal.SetFocusOnShow(list)

	// Set up hints
	modal.SetHints([]components.KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "Esc", Description: "Close"},
	})

	p := &ThemePicker{
		Modal:    modal,
		list:     list,
		onSelect: onSelect,
	}

	// Handle selection
	list.SetSelectedFunc(func(index int, name string, _ string, _ rune) {
		if t := themes.Get(name); t != nil {
			if p.onSelect != nil {
				p.onSelect(t)
			}
		}
	})

	p.subs.Add(theme.Register(list))

	return p
}

// Start implements nav.Component.
func (p *ThemePicker) Start() {}

// Stop implements nav.Component.
func (p *ThemePicker) Stop() { p.subs.Release() }

// Hints implements nav.Component.
func (p *ThemePicker) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "Esc", Description: "Close"},
	}
}

// InputHandler handles keyboard input.
func (p *ThemePicker) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return p.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// Handle navigation and Enter - Escape is handled by App's SetInputCapture
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				p.list.SetCurrentItem(p.list.GetCurrentItem() + 1)
				return
			case 'k':
				idx := p.list.GetCurrentItem() - 1
				if idx < 0 {
					idx = 0
				}
				p.list.SetCurrentItem(idx)
				return
			case 'g':
				p.list.SetCurrentItem(0)
				return
			case 'G':
				p.list.SetCurrentItem(p.list.GetItemCount() - 1)
				return
			}
		case tcell.KeyEnter:
			// Pass Enter to list for selection
			if handler := p.list.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
			return
		}
		// Don't pass other events - let App handle Escape
	})
}

// Focus handles focus.
func (p *ThemePicker) Focus(delegate func(tview.Primitive)) {
	delegate(p.list)
}

// HasFocus returns focus state.
func (p *ThemePicker) HasFocus() bool {
	return p.list.HasFocus()
}
