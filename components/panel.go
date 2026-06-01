package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
)

// Align specifies text alignment.
type Align int

const (
	// AlignCenter centers the content (default).
	AlignCenter Align = iota
	// AlignLeft aligns content to the left.
	AlignLeft
	// AlignRight aligns content to the right.
	AlignRight
)

// TitleAlign is an alias for Align for backward compatibility.
type TitleAlign = Align

// Title alignment constants (aliases for backward compatibility).
const (
	TitleAlignCenter = AlignCenter
	TitleAlignLeft   = AlignLeft
	TitleAlignRight  = AlignRight
)

// Panel is a container with rounded borders and optional title.
// It delegates focus and input handling to its content.
type Panel struct {
	widgetBase
	content    core.Widget
	title      string
	titleColor tcell.Color // 0 means use theme default (Accent)
	titleAlign TitleAlign  // Title alignment (default: center)
	focused    bool        // Manual focus state for visual indication
}

// NewPanel creates a new Panel container.
func NewPanel() *Panel {
	p := &Panel{}
	p.initWidget()
	return p
}

// SetContent sets the inner content widget.
func (p *Panel) SetContent(content core.Widget) *Panel {
	p.content = content
	return p
}

// SetTitle sets the title displayed in the top border.
func (p *Panel) SetTitle(title string) *Panel {
	p.title = title
	return p
}

// SetTitleColor overrides the title color. Pass 0 to use theme default.
func (p *Panel) SetTitleColor(color tcell.Color) *Panel {
	p.titleColor = color
	return p
}

// SetTitleAlign sets the title alignment (Left, Center, or Right).
func (p *Panel) SetTitleAlign(align TitleAlign) *Panel {
	p.titleAlign = align
	return p
}

// SetFocused sets the manual focus state for visual indication.
// When focused, the panel border uses the accent color.
func (p *Panel) SetFocused(focused bool) *Panel {
	p.focused = focused
	return p
}

// IsFocused returns the manual focus state.
func (p *Panel) IsFocused() bool {
	return p.focused
}

// GetContent returns the inner content widget.
func (p *Panel) GetContent() core.Widget {
	return p.content
}

// Draw renders the panel with rounded borders.
func (p *Panel) Draw(screen tcell.Screen) {
	p.Box.DrawForSubclass(screen)

	x, y, width, height := p.GetInnerRect()
	if width < 2 || height < 2 {
		return
	}

	// Get colors from theme at draw time
	th := p.th()
	bgColor := th.Bg()
	borderColor := th.PanelBorder()
	if p.focused {
		borderColor = th.Accent()
	}
	titleColor := p.titleColor
	if titleColor == 0 {
		if p.focused {
			titleColor = th.Accent()
		} else {
			titleColor = th.PanelTitle()
		}
	}

	style := tcell.StyleDefault.Background(bgColor).Foreground(borderColor)
	titleStyle := tcell.StyleDefault.Background(bgColor).Foreground(titleColor)

	// Draw corners
	screen.SetContent(x, y, '╭', nil, style)
	screen.SetContent(x+width-1, y, '╮', nil, style)
	screen.SetContent(x, y+height-1, '╰', nil, style)
	screen.SetContent(x+width-1, y+height-1, '╯', nil, style)

	// Draw horizontal borders
	for i := x + 1; i < x+width-1; i++ {
		screen.SetContent(i, y, '─', nil, style)
		screen.SetContent(i, y+height-1, '─', nil, style)
	}

	// Draw vertical borders
	for i := y + 1; i < y+height-1; i++ {
		screen.SetContent(x, i, '│', nil, style)
		screen.SetContent(x+width-1, i, '│', nil, style)
	}

	// Draw title in top border if set
	if p.title != "" {
		titleText := " " + p.title + " "
		titleRunes := []rune(titleText)
		titleLen := len(titleRunes)

		// Calculate title start position based on alignment
		var titleStart int
		switch p.titleAlign {
		case TitleAlignLeft:
			titleStart = x + 2 // Leave space after corner
		case TitleAlignRight:
			titleStart = x + width - titleLen - 2 // Leave space before corner
		default: // TitleAlignCenter
			titleStart = x + (width-titleLen)/2
		}

		for i, r := range titleRunes {
			if titleStart+i > x && titleStart+i < x+width-1 {
				screen.SetContent(titleStart+i, y, r, nil, titleStyle)
			}
		}
	}

	// Fill background
	bgStyle := tcell.StyleDefault.Background(bgColor)
	fillRect(screen, x+1, y+1, width-2, height-2, bgStyle)

	// Draw content inside border.
	if p.content != nil {
		p.content.SetRect(x+1, y+1, width-2, height-2)
		p.content.Draw(screen)
	}
}

// Focus delegates to content.
func (p *Panel) Focus() {
	if p.content != nil {
		p.content.Focus()
	}
}

// Blur delegates to content.
func (p *Panel) Blur() {
	if p.content != nil {
		p.content.Blur()
	}
}

// HasFocus delegates to content.
func (p *Panel) HasFocus() bool {
	if p.content != nil {
		return p.content.HasFocus()
	}
	return false
}

// HandleKey processes a key event for the Panel.
func (p *Panel) HandleKey(ev *tcell.EventKey) bool {
	if p.content != nil {
		if kh, ok := p.content.(core.KeyHandler); ok {
			return kh.HandleKey(ev)
		}
	}
	return false
}
