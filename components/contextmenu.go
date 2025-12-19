package components

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MenuItem represents a single menu item
type MenuItem struct {
	ID       string
	Label    string
	Icon     string     // Optional icon
	Shortcut string     // Displayed shortcut (e.g., "Ctrl+C")
	Disabled bool       // Grayed out if true
	Danger   bool       // Red color for destructive actions
	Checked  bool       // For toggle items (shows checkmark)
	Submenu  []MenuItem // Nested menu items
	Handler  func()     // Action when selected

	// Internal
	isDivider bool
}

// MenuSection groups items with optional header
type MenuSection struct {
	Header string // Optional section header
	Items  []MenuItem
}

// ContextMenu is the popup menu component
type ContextMenu struct {
	*tview.Box

	items         []MenuItem
	sections      []MenuSection
	flatItems     []MenuItem // Flattened view including dividers
	selectedIndex int
	position      struct{ x, y int }
	menuWidth     int
	visible       bool

	// State
	parent        *ContextMenu
	activeSubmenu *ContextMenu

	// Callbacks
	onSelect func(item MenuItem)
	onClose  func()

	mu sync.RWMutex
}

// NewContextMenu creates a new context menu
func NewContextMenu() *ContextMenu {
	m := &ContextMenu{
		Box:           tview.NewBox(),
		items:         make([]MenuItem, 0),
		sections:      make([]MenuSection, 0),
		flatItems:     make([]MenuItem, 0),
		selectedIndex: 0,
		menuWidth:     20,
	}
	m.SetBorder(true)
	return m
}

// AddItem adds a simple menu item
func (m *ContextMenu) AddItem(id, label string, handler func()) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, MenuItem{
		ID:      id,
		Label:   label,
		Handler: handler,
	})
	m.rebuildFlatItems()
	return m
}

// AddItemWithShortcut adds item with displayed shortcut
func (m *ContextMenu) AddItemWithShortcut(id, label, shortcut string, handler func()) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, MenuItem{
		ID:       id,
		Label:    label,
		Shortcut: shortcut,
		Handler:  handler,
	})
	m.rebuildFlatItems()
	return m
}

// AddItemWithIcon adds item with icon
func (m *ContextMenu) AddItemWithIcon(id, label, icon string, handler func()) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, MenuItem{
		ID:      id,
		Label:   label,
		Icon:    icon,
		Handler: handler,
	})
	m.rebuildFlatItems()
	return m
}

// AddSubmenu adds item that opens a submenu
func (m *ContextMenu) AddSubmenu(id, label string, items []MenuItem) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, MenuItem{
		ID:      id,
		Label:   label,
		Submenu: items,
	})
	m.rebuildFlatItems()
	return m
}

// AddDivider adds a horizontal line separator
func (m *ContextMenu) AddDivider() *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, MenuItem{
		isDivider: true,
	})
	m.rebuildFlatItems()
	return m
}

// AddSection adds a group of items with optional header
func (m *ContextMenu) AddSection(header string, items []MenuItem) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sections = append(m.sections, MenuSection{
		Header: header,
		Items:  items,
	})
	m.rebuildFlatItems()
	return m
}

// SetItems sets all menu items at once
func (m *ContextMenu) SetItems(items []MenuItem) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = items
	m.sections = nil
	m.rebuildFlatItems()
	return m
}

// SetSections sets sectioned items
func (m *ContextMenu) SetSections(sections []MenuSection) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sections = sections
	m.items = nil
	m.rebuildFlatItems()
	return m
}

func (m *ContextMenu) rebuildFlatItems() {
	m.flatItems = make([]MenuItem, 0)
	maxWidth := 0

	// Add items from sections
	for i, section := range m.sections {
		if i > 0 {
			// Add divider between sections
			m.flatItems = append(m.flatItems, MenuItem{isDivider: true})
		}
		for _, item := range section.Items {
			m.flatItems = append(m.flatItems, item)
			w := len(item.Label) + len(item.Shortcut) + 6
			if w > maxWidth {
				maxWidth = w
			}
		}
	}

	// Add standalone items
	for _, item := range m.items {
		m.flatItems = append(m.flatItems, item)
		if !item.isDivider {
			w := len(item.Label) + len(item.Shortcut) + 6
			if len(item.Submenu) > 0 {
				w += 2 // Arrow
			}
			if w > maxWidth {
				maxWidth = w
			}
		}
	}

	if maxWidth > m.menuWidth {
		m.menuWidth = maxWidth
	}
	if m.menuWidth < 20 {
		m.menuWidth = 20
	}

	// Find first selectable item
	m.selectedIndex = m.findNextSelectable(-1)
}

// SetDisabled enables/disables an item by ID
func (m *ContextMenu) SetDisabled(id string, disabled bool) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.flatItems {
		if m.flatItems[i].ID == id {
			m.flatItems[i].Disabled = disabled
		}
	}
	for i := range m.items {
		if m.items[i].ID == id {
			m.items[i].Disabled = disabled
		}
	}
	for i := range m.sections {
		for j := range m.sections[i].Items {
			if m.sections[i].Items[j].ID == id {
				m.sections[i].Items[j].Disabled = disabled
			}
		}
	}
	return m
}

// SetChecked sets checked state for toggle items
func (m *ContextMenu) SetChecked(id string, checked bool) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.flatItems {
		if m.flatItems[i].ID == id {
			m.flatItems[i].Checked = checked
		}
	}
	for i := range m.items {
		if m.items[i].ID == id {
			m.items[i].Checked = checked
		}
	}
	for i := range m.sections {
		for j := range m.sections[i].Items {
			if m.sections[i].Items[j].ID == id {
				m.sections[i].Items[j].Checked = checked
			}
		}
	}
	return m
}

// SetDanger marks item as destructive (red color)
func (m *ContextMenu) SetDanger(id string, danger bool) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.flatItems {
		if m.flatItems[i].ID == id {
			m.flatItems[i].Danger = danger
		}
	}
	for i := range m.items {
		if m.items[i].ID == id {
			m.items[i].Danger = danger
		}
	}
	for i := range m.sections {
		for j := range m.sections[i].Items {
			if m.sections[i].Items[j].ID == id {
				m.sections[i].Items[j].Danger = danger
			}
		}
	}
	return m
}

// ShowAt displays menu at specific coordinates
func (m *ContextMenu) ShowAt(x, y int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.position.x = x
	m.position.y = y
	m.visible = true
	m.selectedIndex = m.findNextSelectable(-1)
}

// Close closes the menu and any open submenus
func (m *ContextMenu) Close() {
	m.mu.Lock()
	onClose := m.onClose
	m.visible = false
	if m.activeSubmenu != nil {
		m.activeSubmenu.Close()
		m.activeSubmenu = nil
	}
	m.mu.Unlock()

	if onClose != nil {
		onClose()
	}
}

// IsOpen returns true if menu is visible
func (m *ContextMenu) IsOpen() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.visible
}

// SetOnSelect is called when any item is selected
func (m *ContextMenu) SetOnSelect(fn func(item MenuItem)) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onSelect = fn
	return m
}

// SetOnClose is called when menu is closed
func (m *ContextMenu) SetOnClose(fn func()) *ContextMenu {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onClose = fn
	return m
}

func (m *ContextMenu) findNextSelectable(from int) int {
	if len(m.flatItems) == 0 {
		return -1
	}
	for i := from + 1; i < len(m.flatItems); i++ {
		if !m.flatItems[i].isDivider && !m.flatItems[i].Disabled {
			return i
		}
	}
	// Wrap around
	for i := 0; i <= from && i < len(m.flatItems); i++ {
		if !m.flatItems[i].isDivider && !m.flatItems[i].Disabled {
			return i
		}
	}
	return -1
}

func (m *ContextMenu) findPrevSelectable(from int) int {
	if len(m.flatItems) == 0 {
		return -1
	}
	for i := from - 1; i >= 0; i-- {
		if !m.flatItems[i].isDivider && !m.flatItems[i].Disabled {
			return i
		}
	}
	// Wrap around
	for i := len(m.flatItems) - 1; i >= from && i >= 0; i-- {
		if !m.flatItems[i].isDivider && !m.flatItems[i].Disabled {
			return i
		}
	}
	return -1
}

// Draw renders the context menu
func (m *ContextMenu) Draw(screen tcell.Screen) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.visible {
		return
	}

	screenWidth, screenHeight := screen.Size()

	// Calculate menu dimensions
	menuHeight := len(m.flatItems) + 2 // +2 for border

	// Adjust position to stay on screen
	x, y := m.position.x, m.position.y
	if x+m.menuWidth > screenWidth {
		x = screenWidth - m.menuWidth - 1
	}
	if y+menuHeight > screenHeight {
		y = screenHeight - menuHeight - 1
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Draw background
	bgStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	for row := y; row < y+menuHeight && row < screenHeight; row++ {
		for col := x; col < x+m.menuWidth && col < screenWidth; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw border
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	m.drawBorder(screen, x, y, m.menuWidth, menuHeight, borderStyle)

	// Draw items
	for i, item := range m.flatItems {
		itemY := y + 1 + i
		if itemY >= screenHeight-1 {
			break
		}

		if item.isDivider {
			m.drawDivider(screen, x, itemY, m.menuWidth, borderStyle)
		} else {
			m.drawItem(screen, x+1, itemY, m.menuWidth-2, item, i == m.selectedIndex)
		}
	}

	// Draw active submenu
	if m.activeSubmenu != nil {
		m.activeSubmenu.Draw(screen)
	}
}

func (m *ContextMenu) drawBorder(screen tcell.Screen, x, y, width, height int, style tcell.Style) {
	// Corners
	screen.SetContent(x, y, '╭', nil, style)
	screen.SetContent(x+width-1, y, '╮', nil, style)
	screen.SetContent(x, y+height-1, '╰', nil, style)
	screen.SetContent(x+width-1, y+height-1, '╯', nil, style)

	// Top and bottom edges
	for i := 1; i < width-1; i++ {
		screen.SetContent(x+i, y, '─', nil, style)
		screen.SetContent(x+i, y+height-1, '─', nil, style)
	}

	// Left and right edges
	for i := 1; i < height-1; i++ {
		screen.SetContent(x, y+i, '│', nil, style)
		screen.SetContent(x+width-1, y+i, '│', nil, style)
	}
}

func (m *ContextMenu) drawDivider(screen tcell.Screen, x, y, width int, style tcell.Style) {
	screen.SetContent(x, y, '├', nil, style)
	screen.SetContent(x+width-1, y, '┤', nil, style)
	for i := 1; i < width-1; i++ {
		screen.SetContent(x+i, y, '─', nil, style)
	}
}

func (m *ContextMenu) drawItem(screen tcell.Screen, x, y, width int, item MenuItem, selected bool) {
	// Determine style based on state
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	if item.Disabled {
		style = style.Foreground(tcell.ColorDarkGray)
	} else if item.Danger {
		style = style.Foreground(tcell.ColorRed)
	}

	if selected && !item.Disabled {
		style = tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
		if item.Danger {
			style = style.Background(tcell.ColorDarkRed)
		}
	}

	// Clear line
	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, ' ', nil, style)
	}

	col := x

	// Draw checkbox or bullet
	if item.Checked {
		screen.SetContent(col, y, '●', nil, style.Foreground(tcell.ColorGreen))
		col += 2
	} else if item.Icon != "" {
		for _, r := range item.Icon {
			screen.SetContent(col, y, r, nil, style)
			col++
		}
		col++ // Space after icon
	} else {
		col += 2 // Indent
	}

	// Draw label
	for _, r := range item.Label {
		if col >= x+width-2 {
			break
		}
		screen.SetContent(col, y, r, nil, style)
		col++
	}

	// Draw submenu arrow or shortcut on right
	if len(item.Submenu) > 0 {
		arrowX := x + width - 2
		arrowStyle := style
		if !selected {
			arrowStyle = style.Foreground(tcell.ColorDarkGray)
		}
		screen.SetContent(arrowX, y, '→', nil, arrowStyle)
	} else if item.Shortcut != "" {
		shortcutStyle := style
		if !selected {
			shortcutStyle = style.Foreground(tcell.ColorDarkGray)
		}
		shortcutX := x + width - len(item.Shortcut) - 1
		if shortcutX > col+2 {
			for i, r := range item.Shortcut {
				screen.SetContent(shortcutX+i, y, r, nil, shortcutStyle)
			}
		}
	}
}

// InputHandler handles keyboard input
func (m *ContextMenu) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		m.mu.Lock()
		defer m.mu.Unlock()

		if !m.visible {
			return
		}

		key := event.Key()
		switch key {
		case tcell.KeyDown:
			m.moveDown()
		case tcell.KeyUp:
			m.moveUp()
		case tcell.KeyRight:
			// Open submenu if available
			if m.selectedIndex >= 0 && m.selectedIndex < len(m.flatItems) {
				item := m.flatItems[m.selectedIndex]
				if len(item.Submenu) > 0 {
					m.openSubmenu(item)
				}
			}
		case tcell.KeyLeft:
			if m.parent != nil {
				m.visible = false
			}
		case tcell.KeyHome:
			m.selectedIndex = m.findNextSelectable(-1)
		case tcell.KeyEnd:
			m.selectedIndex = m.findPrevSelectable(len(m.flatItems))
		case tcell.KeyEnter:
			m.selectCurrent()
		case tcell.KeyEsc:
			m.visible = false
			if m.onClose != nil {
				go m.onClose()
			}
		case tcell.KeyRune:
			// Handle vim-style navigation
			switch event.Rune() {
			case 'j':
				m.moveDown()
			case 'k':
				m.moveUp()
			case 'h':
				if m.parent != nil {
					m.visible = false
				}
			case 'l':
				if m.selectedIndex >= 0 && m.selectedIndex < len(m.flatItems) {
					item := m.flatItems[m.selectedIndex]
					if len(item.Submenu) > 0 {
						m.openSubmenu(item)
					}
				}
			}
		}
	})
}

func (m *ContextMenu) moveDown() {
	next := m.findNextSelectable(m.selectedIndex)
	if next >= 0 {
		m.selectedIndex = next
	}
}

func (m *ContextMenu) moveUp() {
	prev := m.findPrevSelectable(m.selectedIndex)
	if prev >= 0 {
		m.selectedIndex = prev
	}
}

func (m *ContextMenu) selectCurrent() {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.flatItems) {
		return
	}

	item := m.flatItems[m.selectedIndex]
	if item.Disabled || item.isDivider {
		return
	}

	// If has submenu, open it instead
	if len(item.Submenu) > 0 {
		m.openSubmenu(item)
		return
	}

	// Trigger handler
	if item.Handler != nil {
		go item.Handler()
	}

	// Call onSelect callback
	if m.onSelect != nil {
		go m.onSelect(item)
	}

	// Close menu
	m.visible = false
	if m.onClose != nil {
		go m.onClose()
	}
}

func (m *ContextMenu) openSubmenu(item MenuItem) {
	if m.activeSubmenu != nil {
		m.activeSubmenu.Close()
	}

	submenu := NewContextMenu().SetItems(item.Submenu)
	submenu.parent = m

	// Position submenu to the right
	subX := m.position.x + m.menuWidth
	subY := m.position.y + m.selectedIndex + 1

	submenu.ShowAt(subX, subY)
	m.activeSubmenu = submenu
}

// MouseHandler handles mouse events
func (m *ContextMenu) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		m.mu.Lock()
		defer m.mu.Unlock()

		if !m.visible {
			return false, nil
		}

		mx, my := event.Position()

		// Check if click is within menu bounds
		menuHeight := len(m.flatItems) + 2
		if mx >= m.position.x && mx < m.position.x+m.menuWidth &&
			my >= m.position.y && my < m.position.y+menuHeight {

			itemIndex := my - m.position.y - 1
			if itemIndex >= 0 && itemIndex < len(m.flatItems) {
				item := m.flatItems[itemIndex]
				if !item.isDivider && !item.Disabled {
					m.selectedIndex = itemIndex

					if action == tview.MouseLeftClick {
						m.selectCurrent()
					}
				}
			}
			return true, nil
		}

		// Click outside menu - close it
		if action == tview.MouseLeftClick {
			m.visible = false
			if m.onClose != nil {
				go m.onClose()
			}
		}

		return true, nil
	})
}

// Focus is called when the menu receives focus
func (m *ContextMenu) Focus(delegate func(p tview.Primitive)) {
	m.Box.Focus(delegate)
}

// HasFocus returns whether the menu has focus
func (m *ContextMenu) HasFocus() bool {
	return m.Box.HasFocus()
}
