package components

import (
	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
)

// SplitDirection defines the split orientation: Horizontal for a left/right
// arrangement, Vertical for a top/bottom arrangement.
type SplitDirection int

const (
	// SplitHorizontal splits left and right.
	SplitHorizontal SplitDirection = iota
	// SplitVertical splits top and bottom.
	SplitVertical
)

// Split is a two-pane container whose divider can be resized with Ctrl+Arrow
// when SetResizable(true). Tab or Ctrl+W toggles keyboard focus between panes.
type Split struct {
	widgetBase

	direction SplitDirection
	ratio     float64 // 0.0 to 1.0, proportion of first pane
	minSize   int     // minimum size in rows/cols

	first  core.Widget
	second core.Widget

	resizable   bool
	showDivider bool
	focusedPane int // 0 = first, 1 = second

	// Callbacks
	onResize func(ratio float64)
}

// NewSplit creates a new Split component.
func NewSplit() *Split {
	s := &Split{
		direction:   SplitHorizontal,
		ratio:       0.5,
		minSize:     5,
		resizable:   true,
		showDivider: true,
	}
	s.initWidget()
	return s
}

// SetDirection sets the split direction.
func (s *Split) SetDirection(dir SplitDirection) *Split {
	s.direction = dir
	return s
}

// SetRatio sets the split ratio (0.0 to 1.0).
func (s *Split) SetRatio(ratio float64) *Split {
	if ratio < 0.1 {
		ratio = 0.1
	}
	if ratio > 0.9 {
		ratio = 0.9
	}
	s.ratio = ratio
	return s
}

// SetMinSize sets the minimum pane size.
func (s *Split) SetMinSize(size int) *Split {
	s.minSize = size
	return s
}

// SetFirst sets the first pane content (left or top).
func (s *Split) SetFirst(p core.Widget) *Split {
	s.first = p
	return s
}

// SetSecond sets the second pane content (right or bottom).
func (s *Split) SetSecond(p core.Widget) *Split {
	s.second = p
	return s
}

// SetLeft sets the left pane (horizontal split).
func (s *Split) SetLeft(p core.Widget) *Split {
	return s.SetFirst(p)
}

// SetRight sets the right pane (horizontal split).
func (s *Split) SetRight(p core.Widget) *Split {
	return s.SetSecond(p)
}

// SetTop sets the top pane (vertical split).
func (s *Split) SetTop(p core.Widget) *Split {
	return s.SetFirst(p)
}

// SetBottom sets the bottom pane (vertical split).
func (s *Split) SetBottom(p core.Widget) *Split {
	return s.SetSecond(p)
}

// SetResizable enables/disables resizing.
func (s *Split) SetResizable(resizable bool) *Split {
	s.resizable = resizable
	return s
}

// SetShowDivider enables/disables the divider line.
func (s *Split) SetShowDivider(show bool) *Split {
	s.showDivider = show
	return s
}

// SetOnResize sets the callback for resize events.
func (s *Split) SetOnResize(fn func(ratio float64)) *Split {
	s.onResize = fn
	return s
}

// GetRatio returns the current split ratio.
func (s *Split) GetRatio() float64 {
	return s.ratio
}

// FocusFirst focuses the first pane.
func (s *Split) FocusFirst() *Split {
	s.focusedPane = 0
	return s
}

// FocusSecond focuses the second pane.
func (s *Split) FocusSecond() *Split {
	s.focusedPane = 1
	return s
}

// ToggleFocus switches focus between panes.
func (s *Split) ToggleFocus() *Split {
	s.focusedPane = 1 - s.focusedPane
	return s
}

// FocusedPane returns which pane is currently focused (0 = first, 1 = second).
func (s *Split) FocusedPane() int {
	return s.focusedPane
}

// Draw renders the split panes.
func (s *Split) Draw(screen tcell.Screen) {
	th := s.th()
	// Update background color from theme
	s.Box.SetBackgroundColor(th.Bg())

	s.Box.DrawForSubclass(screen)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := th.Bg()
	borderColor := th.Border()
	borderFocusColor := th.BorderFocus()

	var firstX, firstY, firstW, firstH int
	var secondX, secondY, secondW, secondH int
	var dividerX, dividerY, dividerLen int
	var dividerChar rune

	if s.direction == SplitHorizontal {
		// Calculate widths
		dividerWidth := 0
		if s.showDivider {
			dividerWidth = 1
		}

		availableWidth := width - dividerWidth
		firstW = int(float64(availableWidth) * s.ratio)
		if firstW < s.minSize {
			firstW = s.minSize
		}
		if firstW > availableWidth-s.minSize {
			firstW = availableWidth - s.minSize
		}
		secondW = availableWidth - firstW

		firstX, firstY = x, y
		firstH = height

		if s.showDivider {
			dividerX = x + firstW
			dividerY = y
			dividerLen = height
			dividerChar = '│'
		}

		secondX = x + firstW + dividerWidth
		secondY = y
		secondH = height
	} else {
		// Calculate heights
		dividerHeight := 0
		if s.showDivider {
			dividerHeight = 1
		}

		availableHeight := height - dividerHeight
		firstH = int(float64(availableHeight) * s.ratio)
		if firstH < s.minSize {
			firstH = s.minSize
		}
		if firstH > availableHeight-s.minSize {
			firstH = availableHeight - s.minSize
		}
		secondH = availableHeight - firstH

		firstX, firstY = x, y
		firstW = width

		if s.showDivider {
			dividerX = x
			dividerY = y + firstH
			dividerLen = width
			dividerChar = '─'
		}

		secondX = x
		secondY = y + firstH + dividerHeight
		secondW = width
	}

	// Draw divider
	if s.showDivider {
		dividerStyle := tcell.StyleDefault.Background(bgColor).Foreground(borderColor)
		if s.HasFocus() {
			dividerStyle = dividerStyle.Foreground(borderFocusColor)
		}

		if s.direction == SplitHorizontal {
			for i := 0; i < dividerLen; i++ {
				screen.SetContent(dividerX, dividerY+i, dividerChar, nil, dividerStyle)
			}
		} else {
			for i := 0; i < dividerLen; i++ {
				screen.SetContent(dividerX+i, dividerY, dividerChar, nil, dividerStyle)
			}
		}
	}

	// Draw panes
	if s.first != nil && firstW > 0 && firstH > 0 {
		s.first.SetRect(firstX, firstY, firstW, firstH)
		s.first.Draw(screen)
	}

	if s.second != nil && secondW > 0 && secondH > 0 {
		s.second.SetRect(secondX, secondY, secondW, secondH)
		s.second.Draw(screen)
	}
}

// HandleKey processes a key event for the Split.
func (s *Split) HandleKey(ev *tcell.EventKey) bool {
	// Check for resize keys with Ctrl modifier
	if s.resizable && ev.Modifiers()&tcell.ModCtrl != 0 {
		switch ev.Key() {
		case tcell.KeyLeft:
			if s.direction == SplitHorizontal {
				s.adjustRatio(-0.05)
				return true
			}
		case tcell.KeyRight:
			if s.direction == SplitHorizontal {
				s.adjustRatio(0.05)
				return true
			}
		case tcell.KeyUp:
			if s.direction == SplitVertical {
				s.adjustRatio(-0.05)
				return true
			}
		case tcell.KeyDown:
			if s.direction == SplitVertical {
				s.adjustRatio(0.05)
				return true
			}
		}
	}

	// Check for focus switching
	switch ev.Key() {
	case tcell.KeyTab:
		s.ToggleFocus()
		return true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'w':
			if ev.Modifiers()&tcell.ModCtrl != 0 {
				s.ToggleFocus()
				return true
			}
		}
	}

	// Pass to focused pane
	var focusedContent core.Widget
	if s.focusedPane == 0 {
		focusedContent = s.first
	} else {
		focusedContent = s.second
	}

	if focusedContent != nil {
		if kh, ok := focusedContent.(core.KeyHandler); ok {
			kh.HandleKey(ev)
		}
	}
	return false
}

func (s *Split) adjustRatio(delta float64) {
	newRatio := s.ratio + delta
	if newRatio < 0.1 {
		newRatio = 0.1
	}
	if newRatio > 0.9 {
		newRatio = 0.9
	}
	s.ratio = newRatio
	if s.onResize != nil {
		s.onResize(s.ratio)
	}
}

// Focus handles focus.
func (s *Split) Focus() {
	var target core.Widget
	if s.focusedPane == 0 {
		target = s.first
	} else {
		target = s.second
	}
	if target != nil {
		target.Focus()
	} else {
		s.Box.Focus()
	}
}

// Blur handles blur.
func (s *Split) Blur() {
	if s.first != nil {
		s.first.Blur()
	}
	if s.second != nil {
		s.second.Blur()
	}
	s.Box.Blur()
}

// HasFocus returns whether any pane has focus.
func (s *Split) HasFocus() bool {
	if s.first != nil && s.first.HasFocus() {
		return true
	}
	if s.second != nil && s.second.HasFocus() {
		return true
	}
	return s.Box.HasFocus()
}
