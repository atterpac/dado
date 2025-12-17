package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// Tab represents a single tab with content.
type Tab struct {
	Name    string
	Icon    string
	Badge   int
	Content tview.Primitive
}

// Tabs is a tabbed container component.
type Tabs struct {
	*tview.Box

	tabs        []*Tab
	activeIndex int

	// Appearance
	showIcons  bool
	showBadges bool
	closable   bool

	// Callbacks
	onChange func(index int, name string)
	onClose  func(index int) bool
}

// NewTabs creates a new Tabs component.
func NewTabs() *Tabs {
	return &Tabs{
		Box:        tview.NewBox(),
		showIcons:  true,
		showBadges: true,
	}
}

// AddTab adds a new tab.
func (t *Tabs) AddTab(name string, content tview.Primitive) *Tabs {
	t.tabs = append(t.tabs, &Tab{
		Name:    name,
		Content: content,
	})
	return t
}

// AddTabWithIcon adds a new tab with an icon.
func (t *Tabs) AddTabWithIcon(name, icon string, content tview.Primitive) *Tabs {
	t.tabs = append(t.tabs, &Tab{
		Name:    name,
		Icon:    icon,
		Content: content,
	})
	return t
}

// RemoveTab removes a tab by index.
func (t *Tabs) RemoveTab(index int) *Tabs {
	if index >= 0 && index < len(t.tabs) {
		t.tabs = append(t.tabs[:index], t.tabs[index+1:]...)
		if t.activeIndex >= len(t.tabs) {
			t.activeIndex = len(t.tabs) - 1
		}
		if t.activeIndex < 0 {
			t.activeIndex = 0
		}
	}
	return t
}

// SetActive sets the active tab by index.
func (t *Tabs) SetActive(index int) *Tabs {
	if index >= 0 && index < len(t.tabs) {
		t.activeIndex = index
		if t.onChange != nil {
			t.onChange(index, t.tabs[index].Name)
		}
	}
	return t
}

// SetActiveByName sets the active tab by name.
func (t *Tabs) SetActiveByName(name string) *Tabs {
	for i, tab := range t.tabs {
		if tab.Name == name {
			return t.SetActive(i)
		}
	}
	return t
}

// GetActive returns the active tab index.
func (t *Tabs) GetActive() int {
	return t.activeIndex
}

// GetActiveTab returns the active tab.
func (t *Tabs) GetActiveTab() *Tab {
	if t.activeIndex >= 0 && t.activeIndex < len(t.tabs) {
		return t.tabs[t.activeIndex]
	}
	return nil
}

// SetBadge sets a badge count on a tab.
func (t *Tabs) SetBadge(name string, count int) *Tabs {
	for _, tab := range t.tabs {
		if tab.Name == name {
			tab.Badge = count
			break
		}
	}
	return t
}

// ClearBadge clears the badge on a tab.
func (t *Tabs) ClearBadge(name string) *Tabs {
	return t.SetBadge(name, 0)
}

// SetShowIcons enables/disables icon display.
func (t *Tabs) SetShowIcons(show bool) *Tabs {
	t.showIcons = show
	return t
}

// SetShowBadges enables/disables badge display.
func (t *Tabs) SetShowBadges(show bool) *Tabs {
	t.showBadges = show
	return t
}

// SetClosable enables/disables tab closing.
func (t *Tabs) SetClosable(closable bool) *Tabs {
	t.closable = closable
	return t
}

// SetOnChange sets the callback for tab changes.
func (t *Tabs) SetOnChange(fn func(index int, name string)) *Tabs {
	t.onChange = fn
	return t
}

// SetOnClose sets the callback for tab close requests.
// Return false to prevent closing.
func (t *Tabs) SetOnClose(fn func(index int) bool) *Tabs {
	t.onClose = fn
	return t
}

// TabCount returns the number of tabs.
func (t *Tabs) TabCount() int {
	return len(t.tabs)
}

// Draw renders the tabs.
func (t *Tabs) Draw(screen tcell.Screen) {
	t.Box.DrawForSubclass(screen, t)
	x, y, width, height := t.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	borderColor := theme.Border()
	errorColor := theme.Error()

	tabBarHeight := 1

	// Draw tab bar background
	barStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	for col := x; col < x+width; col++ {
		screen.SetContent(col, y, ' ', nil, barStyle)
	}

	// Draw tabs
	col := x
	for i, tab := range t.tabs {
		if col >= x+width {
			break
		}

		// Build tab label
		var label string
		if t.showIcons && tab.Icon != "" {
			label = tab.Icon + " "
		}
		label += tab.Name

		// Add badge
		if t.showBadges && tab.Badge > 0 {
			label += " (" + itoa(tab.Badge) + ")"
		}

		// Add close button
		if t.closable {
			label += " ×"
		}

		// Add padding
		label = " " + label + " "

		// Determine style
		var style tcell.Style
		if i == t.activeIndex {
			style = tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
		} else {
			style = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		}

		// Draw tab
		for _, r := range label {
			if col < x+width {
				screen.SetContent(col, y, r, nil, style)
				col++
			}
		}

		// Draw separator
		if i < len(t.tabs)-1 && col < x+width {
			sepStyle := tcell.StyleDefault.Background(bgColor).Foreground(borderColor)
			screen.SetContent(col, y, '│', nil, sepStyle)
			col++
		}
	}

	// Fill remaining tab bar
	for col < x+width {
		screen.SetContent(col, y, ' ', nil, barStyle)
		col++
	}

	// Draw underline for active tab indicator
	underlineStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)
	for col := x; col < x+width; col++ {
		screen.SetContent(col, y+1, '─', nil, underlineStyle)
	}

	// Draw active tab content
	if t.activeIndex >= 0 && t.activeIndex < len(t.tabs) {
		content := t.tabs[t.activeIndex].Content
		if content != nil {
			contentY := y + tabBarHeight + 1
			contentHeight := height - tabBarHeight - 1
			if contentHeight > 0 {
				content.SetRect(x, contentY, width, contentHeight)
				content.Draw(screen)
			}
		}
	}

	// Draw focus indicator if needed
	_ = errorColor // Available for badge highlight if needed
}

// InputHandler handles keyboard input.
func (t *Tabs) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// First, try to pass to active content
		if t.activeIndex >= 0 && t.activeIndex < len(t.tabs) {
			content := t.tabs[t.activeIndex].Content
			if content != nil {
				if handler := content.InputHandler(); handler != nil {
					// Check if this is a tab-specific key first
					if !t.isTabKey(event) {
						handler(event, setFocus)
						return
					}
				}
			}
		}

		switch event.Key() {
		case tcell.KeyTab:
			t.nextTab()
		case tcell.KeyBacktab:
			t.prevTab()
		case tcell.KeyRune:
			switch event.Rune() {
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				idx := int(event.Rune() - '1')
				if idx < len(t.tabs) {
					t.SetActive(idx)
				}
			case 'H':
				t.prevTab()
			case 'L':
				t.nextTab()
			case 'x':
				if t.closable {
					t.closeCurrentTab()
				}
			}
		}
	})
}

func (t *Tabs) isTabKey(event *tcell.EventKey) bool {
	if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab {
		return true
	}
	if event.Key() == tcell.KeyRune {
		switch event.Rune() {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', 'H', 'L':
			return true
		case 'x':
			return t.closable
		}
	}
	return false
}

func (t *Tabs) nextTab() {
	if len(t.tabs) == 0 {
		return
	}
	t.SetActive((t.activeIndex + 1) % len(t.tabs))
}

func (t *Tabs) prevTab() {
	if len(t.tabs) == 0 {
		return
	}
	idx := t.activeIndex - 1
	if idx < 0 {
		idx = len(t.tabs) - 1
	}
	t.SetActive(idx)
}

func (t *Tabs) closeCurrentTab() {
	if len(t.tabs) == 0 {
		return
	}
	if t.onClose != nil && !t.onClose(t.activeIndex) {
		return
	}
	t.RemoveTab(t.activeIndex)
}

// MouseHandler handles mouse input.
func (t *Tabs) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return t.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		x, y, _, _ := t.GetInnerRect()
		mx, my := event.Position()

		if !t.InRect(mx, my) {
			return false, nil
		}

		// Check if click is on tab bar
		if my == y {
			if action == tview.MouseLeftClick {
				setFocus(t)
				// Find which tab was clicked
				col := x
				for i, tab := range t.tabs {
					tabWidth := t.calcTabWidth(tab)
					if mx >= col && mx < col+tabWidth {
						// Check if close button was clicked
						if t.closable && mx == col+tabWidth-2 {
							if t.onClose == nil || t.onClose(i) {
								t.RemoveTab(i)
							}
						} else {
							t.SetActive(i)
						}
						return true, t
					}
					col += tabWidth + 1 // +1 for separator
				}
				return true, t
			}
		}

		// Pass to content
		if t.activeIndex >= 0 && t.activeIndex < len(t.tabs) {
			content := t.tabs[t.activeIndex].Content
			if content != nil {
				if handler := content.MouseHandler(); handler != nil {
					return handler(action, event, setFocus)
				}
			}
		}

		return false, nil
	})
}

func (t *Tabs) calcTabWidth(tab *Tab) int {
	width := 2 // padding
	if t.showIcons && tab.Icon != "" {
		width += len(tab.Icon) + 1
	}
	width += len(tab.Name)
	if t.showBadges && tab.Badge > 0 {
		width += 3 + len(itoa(tab.Badge))
	}
	if t.closable {
		width += 2
	}
	return width
}

// Focus handles focus.
func (t *Tabs) Focus(delegate func(tview.Primitive)) {
	if t.activeIndex >= 0 && t.activeIndex < len(t.tabs) {
		content := t.tabs[t.activeIndex].Content
		if content != nil {
			delegate(content)
			return
		}
	}
	t.Box.Focus(delegate)
}

// HasFocus returns whether the tabs or content has focus.
func (t *Tabs) HasFocus() bool {
	if t.activeIndex >= 0 && t.activeIndex < len(t.tabs) {
		content := t.tabs[t.activeIndex].Content
		if content != nil && content.HasFocus() {
			return true
		}
	}
	return t.Box.HasFocus()
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
