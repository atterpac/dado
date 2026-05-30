package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SplitDirection defines the split orientation.
type SplitDirection int

const (
	// SplitHorizontal splits left and right.
	SplitHorizontal SplitDirection = iota
	// SplitVertical splits top and bottom.
	SplitVertical
)

// Split is a resizable split pane container.
type Split struct {
	widgetBase

	direction SplitDirection
	ratio     float64 // 0.0 to 1.0, proportion of first pane
	minSize   int     // minimum size in rows/cols

	first  tview.Primitive
	second tview.Primitive

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
	s.initWidget(tview.NewBox())
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
func (s *Split) SetFirst(p tview.Primitive) *Split {
	s.first = p
	return s
}

// SetSecond sets the second pane content (right or bottom).
func (s *Split) SetSecond(p tview.Primitive) *Split {
	s.second = p
	return s
}

// SetLeft sets the left pane (horizontal split).
func (s *Split) SetLeft(p tview.Primitive) *Split {
	return s.SetFirst(p)
}

// SetRight sets the right pane (horizontal split).
func (s *Split) SetRight(p tview.Primitive) *Split {
	return s.SetSecond(p)
}

// SetTop sets the top pane (vertical split).
func (s *Split) SetTop(p tview.Primitive) *Split {
	return s.SetFirst(p)
}

// SetBottom sets the bottom pane (vertical split).
func (s *Split) SetBottom(p tview.Primitive) *Split {
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

	s.Box.DrawForSubclass(screen, s)
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

// InputHandler handles keyboard input.
func (s *Split) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return s.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// Check for resize keys with Ctrl modifier
		if s.resizable && event.Modifiers()&tcell.ModCtrl != 0 {
			switch event.Key() {
			case tcell.KeyLeft:
				if s.direction == SplitHorizontal {
					s.adjustRatio(-0.05)
					return
				}
			case tcell.KeyRight:
				if s.direction == SplitHorizontal {
					s.adjustRatio(0.05)
					return
				}
			case tcell.KeyUp:
				if s.direction == SplitVertical {
					s.adjustRatio(-0.05)
					return
				}
			case tcell.KeyDown:
				if s.direction == SplitVertical {
					s.adjustRatio(0.05)
					return
				}
			}
		}

		// Check for focus switching
		switch event.Key() {
		case tcell.KeyTab:
			s.ToggleFocus()
			s.focusCurrentPane(setFocus)
			return
		case tcell.KeyRune:
			switch event.Rune() {
			case 'w':
				if event.Modifiers()&tcell.ModCtrl != 0 {
					s.ToggleFocus()
					s.focusCurrentPane(setFocus)
					return
				}
			}
		}

		// Pass to focused pane
		var focusedContent tview.Primitive
		if s.focusedPane == 0 {
			focusedContent = s.first
		} else {
			focusedContent = s.second
		}

		if focusedContent != nil {
			if handler := focusedContent.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
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

func (s *Split) focusCurrentPane(setFocus func(tview.Primitive)) {
	var target tview.Primitive
	if s.focusedPane == 0 {
		target = s.first
	} else {
		target = s.second
	}
	if target != nil {
		setFocus(target)
	}
}

// MouseHandler handles mouse input.
func (s *Split) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return s.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		x, y, width, height := s.GetInnerRect()
		mx, my := event.Position()

		if !s.InRect(mx, my) {
			return false, nil
		}

		// Check for divider drag
		if s.resizable && s.showDivider {
			if s.direction == SplitHorizontal {
				dividerX := x + int(float64(width)*s.ratio)
				if mx == dividerX && action == tview.MouseLeftClick {
					// Start drag - this would need state tracking for full drag support
					return true, s
				}
			} else {
				dividerY := y + int(float64(height)*s.ratio)
				if my == dividerY && action == tview.MouseLeftClick {
					return true, s
				}
			}
		}

		// Determine which pane was clicked
		var firstW, firstH int
		if s.direction == SplitHorizontal {
			firstW = int(float64(width) * s.ratio)
			firstH = height
		} else {
			firstW = width
			firstH = int(float64(height) * s.ratio)
		}

		if mx < x+firstW && my < y+firstH {
			s.focusedPane = 0
			if s.first != nil {
				if handler := s.first.MouseHandler(); handler != nil {
					return handler(action, event, setFocus)
				}
			}
		} else {
			s.focusedPane = 1
			if s.second != nil {
				if handler := s.second.MouseHandler(); handler != nil {
					return handler(action, event, setFocus)
				}
			}
		}

		return false, nil
	})
}

// Focus handles focus.
func (s *Split) Focus(delegate func(tview.Primitive)) {
	var target tview.Primitive
	if s.focusedPane == 0 {
		target = s.first
	} else {
		target = s.second
	}
	if target != nil {
		delegate(target)
	} else {
		s.Box.Focus(delegate)
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
