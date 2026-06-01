package components

import (
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/theme"
)

// =============================================================================
// Badge - Small label/count indicator
// =============================================================================

// BadgeVariant controls the color scheme of a Badge: Default (neutral gray),
// Primary (accent), Success (green), Warning (yellow), Error (red), Info (blue).
type BadgeVariant int

const (
	BadgeDefault BadgeVariant = iota
	BadgePrimary
	BadgeSuccess
	BadgeWarning
	BadgeError
	BadgeInfo
)

// Badge is a small pill-shaped label for counts, statuses, or tags.
type Badge struct {
	widgetBase

	text    string
	variant BadgeVariant
	pill    bool // Rounded pill style
	icon    string
}

// NewBadge creates a new badge
func NewBadge(text string) *Badge {
	b := &Badge{
		text:    text,
		variant: BadgeDefault,
		pill:    true,
	}
	b.initWidget()
	return b
}

// SetText sets the badge text
func (b *Badge) SetText(text string) *Badge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.text = text
	return b
}

// SetVariant sets the badge color variant
func (b *Badge) SetVariant(v BadgeVariant) *Badge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.variant = v
	return b
}

// SetPill enables/disables pill (rounded) style
func (b *Badge) SetPill(pill bool) *Badge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pill = pill
	return b
}

// SetIcon sets an icon prefix
func (b *Badge) SetIcon(icon string) *Badge {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.icon = icon
	return b
}

// GetText returns the badge text
func (b *Badge) GetText() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.text
}

func (b *Badge) getColors() (bg, fg tcell.Color) {
	t := b.th()
	switch b.variant {
	case BadgePrimary:
		return t.Accent(), t.Bg()
	case BadgeSuccess:
		return t.Success(), t.Bg()
	case BadgeWarning:
		return t.Warning(), t.Bg()
	case BadgeError:
		return t.Error(), t.Bg()
	case BadgeInfo:
		return t.Info(), t.Bg()
	default:
		return t.BgLight(), t.Fg()
	}
}

// Draw renders the badge
func (b *Badge) Draw(screen tcell.Screen) {
	b.Box.DrawForSubclass(screen)
	x, y, width, height := b.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	bgColor, fgColor := b.getColors()
	style := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)

	// Build display string with optional pill padding.
	display := b.text
	if b.icon != "" {
		display = b.icon + " " + display
	}
	if b.pill {
		display = " " + display + " "
	}

	fillLine(screen, x, y, width, tcell.StyleDefault.Background(b.th().Bg()))
	drawCentered(screen, x, y, width, display, style)
}

// GetFieldHeight returns preferred height
func (b *Badge) GetFieldHeight() int {
	return 1
}

// Width returns the badge's preferred width
func (b *Badge) Width() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	w := len(b.text)
	if b.icon != "" {
		w += len(b.icon) + 1
	}
	if b.pill {
		w += 2
	}
	return w
}

// =============================================================================
// Chip - Removable tag with optional icon
// =============================================================================

// Chip is a labeled tag with optional remove button and click handler.
// Set removable=true via SetRemovable and wire SetOnRemove to handle dismissal.
type Chip struct {
	widgetBase

	text      string
	icon      string
	removable bool
	selected  bool
	disabled  bool

	onRemove func()
	onClick  func()
}

// NewChip creates a new chip
func NewChip(text string) *Chip {
	c := &Chip{text: text}
	c.initWidget()
	return c
}

// SetText sets the chip text
func (c *Chip) SetText(text string) *Chip {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.text = text
	return c
}

// SetIcon sets the chip icon
func (c *Chip) SetIcon(icon string) *Chip {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.icon = icon
	return c
}

// SetRemovable enables/disables the remove button
func (c *Chip) SetRemovable(removable bool) *Chip {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.removable = removable
	return c
}

// SetSelected sets the selected state
func (c *Chip) SetSelected(selected bool) *Chip {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.selected = selected
	return c
}

// SetDisabled sets the disabled state
func (c *Chip) SetDisabled(disabled bool) *Chip {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disabled = disabled
	return c
}

// SetOnRemove sets the remove callback
func (c *Chip) SetOnRemove(fn func()) *Chip {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onRemove = fn
	return c
}

// SetOnClick sets the click callback
func (c *Chip) SetOnClick(fn func()) *Chip {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onClick = fn
	return c
}

// GetText returns the chip text
func (c *Chip) GetText() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.text
}

// IsSelected returns whether the chip is selected
func (c *Chip) IsSelected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.selected
}

// Draw renders the chip
func (c *Chip) Draw(screen tcell.Screen) {
	c.Box.DrawForSubclass(screen)
	x, y, width, height := c.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Get colors
	t := c.th()
	bgColor := t.Bg()
	chipBg := t.BgLight()
	chipFg := t.Fg()

	if c.selected {
		chipBg = t.Accent()
		chipFg = t.Bg()
	}
	if c.disabled {
		chipFg = t.FgDim()
	}

	chipStyle := tcell.StyleDefault.Background(chipBg).Foreground(chipFg)
	removeStyle := tcell.StyleDefault.Background(chipBg).Foreground(t.Error())

	// Body: " text " (plus a trailing space when removable, before the ✕).
	body := c.text
	if c.icon != "" {
		body = c.icon + " " + body
	}
	body = " " + body + " "
	chipWidth := runeLen(body)
	if c.removable {
		chipWidth += 2 // " ✕"
	}

	startX := x + (width-chipWidth)/2
	if startX < x {
		startX = x
	}
	right := x + width

	fillLine(screen, x, y, width, tcell.StyleDefault.Background(bgColor))
	col := drawText(screen, startX, y, right-startX, body, chipStyle)

	if c.removable && col < right {
		col = drawText(screen, col, y, right-col, " ", chipStyle)
		drawText(screen, col, y, right-col, "✕", removeStyle)
	}
}

// GetFieldHeight returns preferred height
func (c *Chip) GetFieldHeight() int {
	return 1
}

// Width returns the chip's preferred width
func (c *Chip) Width() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	w := len(c.text) + 2
	if c.icon != "" {
		w += len(c.icon) + 1
	}
	if c.removable {
		w += 2
	}
	return w
}

// =============================================================================
// Divider - Horizontal or vertical separator
// =============================================================================

// DividerOrientation specifies whether the Divider draws horizontally or vertically.
type DividerOrientation int

const (
	DividerHorizontal DividerOrientation = iota
	DividerVertical
)

// Divider renders a single-line visual separator using the theme border color.
// Default orientation is horizontal; use NewDivider(DividerVertical) for columns.
type Divider struct {
	widgetBase

	orientation DividerOrientation
	label       string
	style       rune // Character to use (default ─ or │)
}

// NewDivider creates a new horizontal divider
func NewDivider() *Divider {
	d := &Divider{
		orientation: DividerHorizontal,
		style:       '─',
	}
	d.initWidget()
	return d
}

// NewVerticalDivider creates a new vertical divider
func NewVerticalDivider() *Divider {
	d := &Divider{
		orientation: DividerVertical,
		style:       '│',
	}
	d.initWidget()
	return d
}

// SetLabel sets an optional centered label
func (d *Divider) SetLabel(label string) *Divider {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.label = label
	return d
}

// SetStyle sets the divider character
func (d *Divider) SetStyle(char rune) *Divider {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.style = char
	return d
}

// SetOrientation sets horizontal or vertical
func (d *Divider) SetOrientation(o DividerOrientation) *Divider {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.orientation = o
	if o == DividerVertical && d.style == '─' {
		d.style = '│'
	} else if o == DividerHorizontal && d.style == '│' {
		d.style = '─'
	}
	return d
}

// Draw renders the divider
func (d *Divider) Draw(screen tcell.Screen) {
	d.Box.DrawForSubclass(screen)
	x, y, width, height := d.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	t := d.th()
	bgColor := t.Bg()
	fgColor := t.FgDim()
	labelColor := t.Fg()

	bgStyle := tcell.StyleDefault.Background(bgColor)
	lineStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(labelColor)

	if d.orientation == DividerHorizontal {
		fillLine(screen, x, y, width, bgStyle)

		if d.label != "" {
			// Draw line with label in middle
			labelStart := x + (width-len(d.label)-2)/2

			for col := x; col < x+width; col++ {
				if col == labelStart-1 {
					screen.SetContent(col, y, ' ', nil, bgStyle)
				} else if col >= labelStart && col < labelStart+len(d.label) {
					screen.SetContent(col, y, rune(d.label[col-labelStart]), nil, labelStyle)
				} else if col == labelStart+len(d.label) {
					screen.SetContent(col, y, ' ', nil, bgStyle)
				} else {
					screen.SetContent(col, y, d.style, nil, lineStyle)
				}
			}
		} else {
			// Simple line
			for col := x; col < x+width; col++ {
				screen.SetContent(col, y, d.style, nil, lineStyle)
			}
		}
	} else {
		// Vertical
		for row := y; row < y+height; row++ {
			screen.SetContent(x, row, d.style, nil, lineStyle)
		}
	}
}

// GetFieldHeight returns preferred height
func (d *Divider) GetFieldHeight() int {
	return 1
}

// =============================================================================
// Skeleton - Loading placeholder with animation
// =============================================================================

// SkeletonVariant controls the shape rendered during animated loading:
// Text for multi-line stubs, Block for rectangular placeholders, Circle for avatars.
type SkeletonVariant int

const (
	SkeletonText   SkeletonVariant = iota // Single line text
	SkeletonBlock                         // Rectangular block
	SkeletonCircle                        // Circle/avatar
)

// Skeleton renders an animated placeholder in place of content while data loads.
// Size it identically to the content it replaces so layout does not shift on load.
type Skeleton struct {
	widgetBase

	variant  SkeletonVariant
	lines    int // Number of text lines
	animated bool
	frame    int
	interval time.Duration
	running  bool
	stopCh   chan struct{}
}

// NewSkeleton creates a new skeleton loader
func NewSkeleton() *Skeleton {
	s := &Skeleton{
		variant:  SkeletonText,
		lines:    1,
		animated: true,
		interval: 150 * time.Millisecond,
	}
	s.initWidget()
	return s
}

// SetVariant sets the skeleton shape
func (s *Skeleton) SetVariant(v SkeletonVariant) *Skeleton {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.variant = v
	return s
}

// SetLines sets number of text lines
func (s *Skeleton) SetLines(lines int) *Skeleton {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lines = lines
	return s
}

// SetAnimated enables/disables animation
func (s *Skeleton) SetAnimated(animated bool) *Skeleton {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.animated = animated
	return s
}

// Start begins the animation
func (s *Skeleton) Start() *Skeleton {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running || !s.animated {
		return s
	}

	s.running = true
	s.stopCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.mu.Lock()
				s.frame = (s.frame + 1) % 3
				s.mu.Unlock()
				theme.QueueUpdateDraw(func() {})
			}
		}
	}()

	return s
}

// Stop ends the animation
func (s *Skeleton) Stop() *Skeleton {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return s
	}

	s.running = false
	close(s.stopCh)
	return s
}

// Draw renders the skeleton
func (s *Skeleton) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	t := s.th()
	bgColor := t.Bg()
	// Animate between shades
	var fgColor tcell.Color
	switch s.frame {
	case 0:
		fgColor = t.BgLight()
	case 1:
		fgColor = t.FgDim()
	case 2:
		fgColor = t.BgLight()
	}

	bgStyle := tcell.StyleDefault.Background(bgColor)
	skelStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)

	fillRect(screen, x, y, width, height, bgStyle)

	switch s.variant {
	case SkeletonText:
		// Draw text-like lines with varying widths
		for line := 0; line < s.lines && y+line < y+height; line++ {
			lineWidth := width
			if line == s.lines-1 && s.lines > 1 {
				lineWidth = width * 2 / 3 // Last line shorter
			}
			for col := x; col < x+lineWidth; col++ {
				screen.SetContent(col, y+line, '░', nil, skelStyle)
			}
		}

	case SkeletonBlock:
		// Fill entire area
		for row := y; row < y+height; row++ {
			for col := x; col < x+width; col++ {
				screen.SetContent(col, row, '░', nil, skelStyle)
			}
		}

	case SkeletonCircle:
		// Draw a simple circle approximation
		size := height
		if width < size {
			size = width
		}
		centerX := x + width/2
		centerY := y + height/2
		radius := size / 2

		for row := y; row < y+height; row++ {
			for col := x; col < x+width; col++ {
				dx := col - centerX
				dy := (row - centerY) * 2 // Adjust for character aspect ratio
				if dx*dx+dy*dy <= radius*radius {
					screen.SetContent(col, row, '░', nil, skelStyle)
				}
			}
		}
	}
}

// GetFieldHeight returns preferred height
func (s *Skeleton) GetFieldHeight() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.variant == SkeletonText {
		return s.lines
	}
	return 3
}
