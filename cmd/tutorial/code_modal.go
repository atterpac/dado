package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
)

// CodeModal displays source code in a modal overlay.
type CodeModal struct {
	*components.Modal
	codeView *tview.TextView
	code     string
	subs     components.Subscriptions
}

// Subs returns the modal's subscription set; release on dismissal.
func (m *CodeModal) Subs() *components.Subscriptions { return &m.subs }

// NewCodeModal creates a new code modal.
func NewCodeModal(title string, code string) *CodeModal {
	codeView := tview.NewTextView()
	codeView.SetDynamicColors(true)
	codeView.SetScrollable(true)
	codeView.SetWrap(false)
	codeView.SetBackgroundColor(theme.Bg())
	codeView.SetTextColor(theme.Fg())

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

	m.subs.Add(theme.Register(codeView))

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

// InputHandler handles keyboard input.
func (m *CodeModal) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	// Use Box.WrapInputHandler directly instead of Modal.WrapInputHandler
	// to avoid Modal's handleBaseInput which also handles Escape.
	// Escape is handled by App's SetInputCapture via DismissModal.
	return m.Box.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				row, _ := m.codeView.GetScrollOffset()
				m.codeView.ScrollTo(row+1, 0)
			case 'k':
				row, _ := m.codeView.GetScrollOffset()
				if row > 0 {
					m.codeView.ScrollTo(row-1, 0)
				}
			case 'g':
				m.codeView.ScrollToBeginning()
			case 'G':
				m.codeView.ScrollToEnd()
			}
		case tcell.KeyPgDn, tcell.KeyCtrlD:
			row, _ := m.codeView.GetScrollOffset()
			_, _, _, height := m.codeView.GetInnerRect()
			m.codeView.ScrollTo(row+height/2, 0)
		case tcell.KeyPgUp, tcell.KeyCtrlU:
			row, _ := m.codeView.GetScrollOffset()
			_, _, _, height := m.codeView.GetInnerRect()
			newRow := row - height/2
			if newRow < 0 {
				newRow = 0
			}
			m.codeView.ScrollTo(newRow, 0)
		case tcell.KeyEscape:
			// Let App's SetInputCapture handle this - do nothing here
			return
		}
	})
}

// Focus handles focus.
func (m *CodeModal) Focus(delegate func(tview.Primitive)) {
	delegate(m.codeView)
}

// HasFocus returns focus state.
func (m *CodeModal) HasFocus() bool {
	return m.codeView.HasFocus()
}
