package components

import (
	"github.com/gdamore/tcell/v2"
)

// ButtonVariant defines the visual style of a button.
type ButtonVariant int

const (
	// ButtonDefault uses theme accent colors.
	ButtonDefault ButtonVariant = iota
	// ButtonPrimary uses theme accent colors (same as default).
	ButtonPrimary
	// ButtonSecondary uses dimmed accent colors.
	ButtonSecondary
	// ButtonDanger uses error/red colors.
	ButtonDanger
	// ButtonGhost uses transparent background with accent text.
	ButtonGhost
)

// Button is a clickable button component.
// It wraps button functionality with themed defaults and a cleaner API.
type Button struct {
	widgetBase

	label    string
	variant  ButtonVariant
	disabled bool
	focused  bool

	onClick func()
}

// NewButton creates a new Button with the given label.
func NewButton(label string) *Button {
	b := &Button{
		label:   label,
		variant: ButtonDefault,
	}
	b.initWidget()
	return b
}

// SetLabel sets the button label.
func (b *Button) SetLabel(label string) *Button {
	b.label = label
	return b
}

// SetVariant sets the button visual variant.
func (b *Button) SetVariant(variant ButtonVariant) *Button {
	b.variant = variant
	return b
}

// SetDisabled sets whether the button is disabled.
func (b *Button) SetDisabled(disabled bool) *Button {
	b.disabled = disabled
	return b
}

// SetOnClick sets the click handler.
func (b *Button) SetOnClick(handler func()) *Button {
	b.onClick = handler
	return b
}

// OnClick is an alias for SetOnClick for a more fluent API.
func (b *Button) OnClick(handler func()) *Button {
	return b.SetOnClick(handler)
}

// Click programmatically triggers the button click.
func (b *Button) Click() {
	if !b.disabled && b.onClick != nil {
		b.onClick()
	}
}

// Draw renders the button.
func (b *Button) Draw(screen tcell.Screen) {
	b.Box.DrawForSubclass(screen)

	x, y, width, height := b.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Get colors based on variant and state
	var bgColor, fgColor tcell.Color
	th := b.th()

	if b.disabled {
		bgColor = th.BgLight()
		fgColor = th.FgMuted()
	} else if b.focused {
		switch b.variant {
		case ButtonDanger:
			bgColor = th.Error()
			fgColor = th.Bg()
		case ButtonGhost:
			bgColor = th.Accent()
			fgColor = th.Bg()
		case ButtonSecondary:
			bgColor = th.AccentDim()
			fgColor = th.Bg()
		default: // ButtonDefault, ButtonPrimary
			bgColor = th.Accent()
			fgColor = th.Bg()
		}
	} else {
		switch b.variant {
		case ButtonDanger:
			bgColor = th.BgLight()
			fgColor = th.Error()
		case ButtonGhost:
			bgColor = th.Bg()
			fgColor = th.Accent()
		case ButtonSecondary:
			bgColor = th.BgLight()
			fgColor = th.FgDim()
		default: // ButtonDefault, ButtonPrimary
			bgColor = th.BgLight()
			fgColor = th.Accent()
		}
	}

	style := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)

	// Calculate button position (centered vertically)
	buttonY := y + (height-1)/2

	// Draw button background
	fillLine(screen, x, buttonY, width, style)

	// Draw label centered
	labelRunes := []rune(b.label)
	labelStart := x + (width-len(labelRunes))/2
	for i, r := range labelRunes {
		if labelStart+i >= x && labelStart+i < x+width {
			screen.SetContent(labelStart+i, buttonY, r, nil, style)
		}
	}

	// Draw focus indicator brackets
	if b.focused && !b.disabled {
		bracketStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor).Bold(true)
		if x < x+width {
			screen.SetContent(x, buttonY, '[', nil, bracketStyle)
		}
		if x+width-1 >= x {
			screen.SetContent(x+width-1, buttonY, ']', nil, bracketStyle)
		}
	}
}

func (b *Button) HandleKey(ev *tcell.EventKey) bool {
	if b.disabled {
		return false
	}

	switch ev.Key() {
	case tcell.KeyEnter:
		b.Click()
		return true
	}

	switch ev.Rune() {
	case ' ':
		b.Click()
		return true
	}

	return false
}

// Focus handles focus.
func (b *Button) Focus() {
	b.focused = true
	b.Box.Focus()
}

// Blur handles blur.
func (b *Button) Blur() {
	b.focused = false
	b.Box.Blur()
}

// HasFocus returns whether the button has focus.
func (b *Button) HasFocus() bool {
	return b.focused
}

// GetFieldHeight returns the preferred height for this button.
func (b *Button) GetFieldHeight() int {
	return 1
}
