package main

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

// ThemePicker is a modal for selecting themes.
type ThemePicker struct {
	*components.Modal
	list     *components.List
	onSelect func(t theme.Theme)
}

// NewThemePicker creates a new theme picker modal.
func NewThemePicker(onSelect func(t theme.Theme)) *ThemePicker {
	list := components.NewList()
	list.SetHighlightFullLine(true)

	themeNames := themes.Names()
	for _, name := range themeNames {
		list.AddItem(name)
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

	list.SetOnSelect(func(_ int, item components.ListItem) {
		if t := themes.Get(item.Text); t != nil {
			if p.onSelect != nil {
				p.onSelect(t)
			}
		}
	})

	return p
}

// Start implements nav.Component.
func (p *ThemePicker) Start() {}

// Stop implements nav.Component.
func (p *ThemePicker) Stop() { p.list.Subs().Release() }

// Hints implements nav.Component.
func (p *ThemePicker) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "Esc", Description: "Close"},
	}
}

func (p *ThemePicker) HandleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j':
			p.list.MoveDown()
			return true
		case 'k':
			p.list.MoveUp()
			return true
		case 'g':
			p.list.MoveToTop()
			return true
		case 'G':
			p.list.MoveToBottom()
			return true
		}
	case tcell.KeyEnter:
		if idx, item, ok := p.list.GetSelected(); ok {
			if t := themes.Get(item.Text); t != nil {
				if p.onSelect != nil {
					p.onSelect(t)
				}
			}
			_ = idx
		}
		return true
	}
	return false
}

// Focus handles focus.
func (p *ThemePicker) Focus() {
	p.list.Focus()
}

// HasFocus returns focus state.
func (p *ThemePicker) HasFocus() bool {
	return p.list.HasFocus()
}
