package main

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// CodeModal displays source code in a modal overlay.
type CodeModal struct {
	*components.Modal
	codeView *core.TextView
	code     string
	subs     components.Subscriptions
}

// Subs returns the modal's subscription set; release on dismissal.
func (m *CodeModal) Subs() *components.Subscriptions { return &m.subs }

// NewCodeModal creates a new code modal.
func NewCodeModal(title string, code string) *CodeModal {
	codeView := core.NewTextView()
	codeView.SetDynamicColors(true)
	codeView.SetScrollable(true)
	codeView.Box.SetBackgroundColor(theme.Bg())

	// Format code with syntax highlighting (basic)
	formattedCode := formatGoCode(code)
	codeView.SetText(formattedCode)

	modal := components.NewModal(components.ModalConfig{
		Title:     "Code: " + title,
		Width:     80,
		Height:    30,
		MinWidth:  40,
		MinHeight: 15,
		Backdrop:  true,
	})
	modal.SetContent(codeView)
	modal.SetDismissOnEsc(true)
	modal.SetFocusOnShow(codeView)

	// Set up hints
	modal.SetHints([]components.KeyHint{
		{Key: "j/k", Description: "Scroll"},
		{Key: "Esc", Description: "Close"},
	})

	m := &CodeModal{
		Modal:    modal,
		codeView: codeView,
		code:     code,
	}

	m.subs.Add(theme.RegisterFn(func(c tcell.Color) { codeView.Box.SetBackgroundColor(c) }))

	return m
}

// formatGoCode applies basic syntax highlighting to Go code.
func formatGoCode(code string) string {
	// For now, return plain code. Could add syntax highlighting later.
	return code
}

// Start implements nav.Component.
func (m *CodeModal) Start() {}

// Stop implements nav.Component.
func (m *CodeModal) Stop() { m.subs.Release() }

// Hints implements nav.Component.
func (m *CodeModal) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "j/k", Description: "Scroll"},
		{Key: "Esc", Description: "Close"},
	}
}

func (m *CodeModal) HandleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j':
			row, _ := m.codeView.GetScrollOffset()
			m.codeView.ScrollTo(row+1, 0)
		case 'k':
			row, _ := m.codeView.GetScrollOffset()
			if row > 0 {
				m.codeView.ScrollTo(row-1, 0)
			}
		case 'g':
			m.codeView.ScrollTo(0, 0)
		case 'G':
			m.codeView.ScrollTo(9999, 0)
		}
	case tcell.KeyPgDn, tcell.KeyCtrlD:
		row, _ := m.codeView.GetScrollOffset()
		_, _, _, height := m.codeView.Box.GetInnerRect()
		m.codeView.ScrollTo(row+height/2, 0)
	case tcell.KeyPgUp, tcell.KeyCtrlU:
		row, _ := m.codeView.GetScrollOffset()
		_, _, _, height := m.codeView.Box.GetInnerRect()
		newRow := row - height/2
		if newRow < 0 {
			newRow = 0
		}
		m.codeView.ScrollTo(newRow, 0)
	case tcell.KeyEscape:
		return false
	}
	return false
}

// Focus handles focus.
func (m *CodeModal) Focus() {
	m.codeView.Focus()
}

// HasFocus returns focus state.
func (m *CodeModal) HasFocus() bool {
	return m.codeView.HasFocus()
}
