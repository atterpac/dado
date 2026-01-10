package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
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
	*tview.Box

	label    string
	variant  ButtonVariant
	disabled bool
	focused  bool

	onClick func()
}

// NewButton creates a new Button with the given label.
func NewButton(label string) *Button {
	box := tview.NewBox()
	box.SetBackgroundColor(theme.Bg())

	b := &Button{
		Box:     box,
		label:   label,
		variant: ButtonDefault,
	}

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
	b.Box.DrawForSubclass(screen, b)

	x, y, width, height := b.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Get colors based on variant and state
	var bgColor, fgColor tcell.Color

	if b.disabled {
		bgColor = theme.BgLight()
		fgColor = theme.FgMuted()
	} else if b.focused {
		switch b.variant {
		case ButtonDanger:
			bgColor = theme.Error()
			fgColor = theme.Bg()
		case ButtonGhost:
			bgColor = theme.Accent()
			fgColor = theme.Bg()
		case ButtonSecondary:
			bgColor = theme.AccentDim()
			fgColor = theme.Bg()
		default: // ButtonDefault, ButtonPrimary
			bgColor = theme.Accent()
			fgColor = theme.Bg()
		}
	} else {
		switch b.variant {
		case ButtonDanger:
			bgColor = theme.BgLight()
			fgColor = theme.Error()
		case ButtonGhost:
			bgColor = theme.Bg()
			fgColor = theme.Accent()
		case ButtonSecondary:
			bgColor = theme.BgLight()
			fgColor = theme.FgDim()
		default: // ButtonDefault, ButtonPrimary
			bgColor = theme.BgLight()
			fgColor = theme.Accent()
		}
	}

	style := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)

	// Calculate button position (centered vertically)
	buttonY := y + (height-1)/2

	// Draw button background
	for col := x; col < x+width; col++ {
		screen.SetContent(col, buttonY, ' ', nil, style)
	}

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

// InputHandler handles keyboard input.
func (b *Button) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return b.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if b.disabled {
			return
		}

		switch event.Key() {
		case tcell.KeyEnter:
			b.Click()
		}

		switch event.Rune() {
		case ' ':
			b.Click()
		}
	})
}

// MouseHandler handles mouse input.
func (b *Button) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return b.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		if b.disabled {
			return false, nil
		}

		x, y := event.Position()
		bx, by, bw, bh := b.GetInnerRect()

		// Check if click is within button bounds
		if x >= bx && x < bx+bw && y >= by && y < by+bh {
			if action == tview.MouseLeftClick {
				setFocus(b)
				b.Click()
				return true, b
			}
		}

		return false, nil
	})
}

// Focus handles focus.
func (b *Button) Focus(delegate func(tview.Primitive)) {
	b.focused = true
	b.Box.Focus(delegate)
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

// Primitive returns the underlying tview.Box for advanced usage.
func (b *Button) Primitive() *tview.Box {
	return b.Box
}

// GetFieldHeight returns the preferred height for this button.
func (b *Button) GetFieldHeight() int {
	return 1
}
