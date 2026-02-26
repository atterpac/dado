package components

import (
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// FinderItem represents a searchable item
type FinderItem struct {
	ID          string   // Unique identifier
	Label       string   // Primary display text (searchable)
	Description string   // Secondary text (searchable)
	Category    string   // Group/category name
	Icon        string   // Optional icon
	Keywords    []string // Additional search terms
	Data        any      // User data
	Score       int      // Match score (set by filtering)
	Matches     []int    // Character indices that matched (for highlighting)
}

// FinderCategory represents a group header
type FinderCategory struct {
	Name     string
	Icon     string
	Priority int // Lower = shown first
}

// PreviewFunc generates preview content for an item
type PreviewFunc func(item FinderItem) string

// Finder is a fuzzy search component
type Finder struct {
	*tview.Box

	// Items
	items      []FinderItem
	filtered   []FinderItem
	categories []FinderCategory

	// Configuration
	placeholder     string
	prompt          string
	maxVisible      int
	minScore        int
	showCategories  bool
	showIcons       bool
	showDescription bool
	caseSensitive   bool
	previewFunc     PreviewFunc
	previewRatio    float64

	// State
	query         string
	selectedIndex int
	scrollOffset  int
	showPreview   bool
	recentIDs     []string
	vimMode       bool // When true, j/k navigate and / enters search
	searching     bool // In vim mode, true when typing into search

	// Cached theme colors (set during Draw)
	bgColor     tcell.Color
	fgColor     tcell.Color
	fgDimColor  tcell.Color
	accentColor tcell.Color

	// Callbacks
	onSelect      func(item FinderItem)
	onChange      func(item FinderItem)
	onCancel      func()
	onQueryChange func(query string)

	mu sync.RWMutex
}

// NewFinder creates a new fuzzy finder
func NewFinder() *Finder {
	f := &Finder{
		Box:             tview.NewBox(),
		items:           make([]FinderItem, 0),
		filtered:        make([]FinderItem, 0),
		categories:      make([]FinderCategory, 0),
		prompt:          "> ",
		maxVisible:      10,
		minScore:        0,
		showDescription: true,
		previewRatio:    0.4,
		recentIDs:       make([]string, 0),
	}
	f.SetBorder(true)
	return f
}

// SetItems sets the searchable items
func (f *Finder) SetItems(items []FinderItem) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items = items
	f.filterItems()
	return f
}

// SetCategories defines category ordering
func (f *Finder) SetCategories(categories []FinderCategory) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.categories = categories
	return f
}

// SetPlaceholder sets input placeholder text
func (f *Finder) SetPlaceholder(text string) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.placeholder = text
	return f
}

// SetPrompt sets the input prompt
func (f *Finder) SetPrompt(prompt string) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prompt = prompt
	return f
}

// SetMaxVisible limits visible results
func (f *Finder) SetMaxVisible(max int) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.maxVisible = max
	return f
}

// SetMinScore sets minimum fuzzy match score
func (f *Finder) SetMinScore(score int) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.minScore = score
	return f
}

// SetShowCategories enables category headers
func (f *Finder) SetShowCategories(show bool) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.showCategories = show
	return f
}

// SetShowIcons enables item icons
func (f *Finder) SetShowIcons(show bool) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.showIcons = show
	return f
}

// SetShowDescription enables description column
func (f *Finder) SetShowDescription(show bool) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.showDescription = show
	return f
}

// SetPreview enables preview pane
func (f *Finder) SetPreview(fn PreviewFunc) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.previewFunc = fn
	return f
}

// SetPreviewRatio sets preview pane width ratio
func (f *Finder) SetPreviewRatio(ratio float64) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.previewRatio = ratio
	return f
}

// SetVimMode enables vim-style navigation (j/k to move, / to search, Esc exits search).
// When disabled (default), all typing goes directly to the search query.
func (f *Finder) SetVimMode(enabled bool) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.vimMode = enabled
	f.searching = false
	return f
}

// SetCaseSensitive enables case-sensitive matching
func (f *Finder) SetCaseSensitive(sensitive bool) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.caseSensitive = sensitive
	return f
}

// SetRecentItems sets recently used items
func (f *Finder) SetRecentItems(ids []string) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recentIDs = ids
	return f
}

// AddRecent adds an item to recent list
func (f *Finder) AddRecent(id string) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Remove if exists
	for i, rid := range f.recentIDs {
		if rid == id {
			f.recentIDs = append(f.recentIDs[:i], f.recentIDs[i+1:]...)
			break
		}
	}
	// Add to front
	f.recentIDs = append([]string{id}, f.recentIDs...)
	// Keep max 10 recent
	if len(f.recentIDs) > 10 {
		f.recentIDs = f.recentIDs[:10]
	}
	return f
}

// ClearRecent clears recent items
func (f *Finder) ClearRecent() *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recentIDs = make([]string, 0)
	return f
}

// SetOnSelect is called when Enter is pressed on an item
func (f *Finder) SetOnSelect(fn func(item FinderItem)) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onSelect = fn
	return f
}

// SetOnChange is called when selection changes
func (f *Finder) SetOnChange(fn func(item FinderItem)) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onChange = fn
	return f
}

// SetOnCancel is called when Esc is pressed
func (f *Finder) SetOnCancel(fn func()) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onCancel = fn
	return f
}

// SetOnQueryChange is called when search text changes
func (f *Finder) SetOnQueryChange(fn func(query string)) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onQueryChange = fn
	return f
}

// GetQuery returns current search text
func (f *Finder) GetQuery() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.query
}

// SetQuery sets search text programmatically
func (f *Finder) SetQuery(query string) *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.query = query
	f.filterItems()
	return f
}

// GetSelected returns currently highlighted item
func (f *Finder) GetSelected() *FinderItem {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.selectedIndex >= 0 && f.selectedIndex < len(f.filtered) {
		item := f.filtered[f.selectedIndex]
		return &item
	}
	return nil
}

// GetFiltered returns all items matching current query
func (f *Finder) GetFiltered() []FinderItem {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]FinderItem, len(f.filtered))
	copy(result, f.filtered)
	return result
}

// Clear resets the finder
func (f *Finder) Clear() *Finder {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.query = ""
	f.selectedIndex = 0
	f.scrollOffset = 0
	f.filterItems()
	return f
}

func (f *Finder) filterItems() {
	if f.query == "" {
		// Show all items, with recent items first
		f.filtered = make([]FinderItem, len(f.items))
		copy(f.filtered, f.items)
		f.sortFiltered()
	} else {
		// Fuzzy filter
		f.filtered = make([]FinderItem, 0)
		query := f.query
		if !f.caseSensitive {
			query = strings.ToLower(query)
		}

		for _, item := range f.items {
			score, matches := f.fuzzyScore(query, item)
			if score >= f.minScore {
				item.Score = score
				item.Matches = matches
				f.filtered = append(f.filtered, item)
			}
		}
		f.sortFiltered()
	}

	// Reset selection if out of bounds
	if f.selectedIndex >= len(f.filtered) {
		f.selectedIndex = 0
	}
	f.scrollOffset = 0
}

func (f *Finder) fuzzyScore(query string, item FinderItem) (int, []int) {
	target := item.Label
	if !f.caseSensitive {
		target = strings.ToLower(target)
	}

	// Try to match in label
	score, matches := f.fuzzyMatch(query, target)
	if score > 0 {
		return score, matches
	}

	// Try description
	if item.Description != "" {
		desc := item.Description
		if !f.caseSensitive {
			desc = strings.ToLower(desc)
		}
		score, _ = f.fuzzyMatch(query, desc)
		if score > 0 {
			return score / 2, nil // Lower score for description matches
		}
	}

	// Try keywords
	for _, kw := range item.Keywords {
		kwLower := kw
		if !f.caseSensitive {
			kwLower = strings.ToLower(kw)
		}
		if strings.Contains(kwLower, query) {
			return 50, nil
		}
	}

	return 0, nil
}

func (f *Finder) fuzzyMatch(query, target string) (int, []int) {
	if len(query) == 0 {
		return 100, nil
	}
	if len(target) == 0 {
		return 0, nil
	}

	queryRunes := []rune(query)
	targetRunes := []rune(target)

	matches := make([]int, 0, len(queryRunes))
	queryIdx := 0
	score := 0
	consecutive := 0
	prevMatchIdx := -1

	for i, r := range targetRunes {
		if queryIdx >= len(queryRunes) {
			break
		}

		qr := queryRunes[queryIdx]
		if r == qr {
			matches = append(matches, i)

			// Bonus for consecutive matches
			if i == prevMatchIdx+1 {
				consecutive++
				score += consecutive * 5
			} else {
				consecutive = 0
			}

			// Bonus for match at word boundary
			if i == 0 || !unicode.IsLetter(targetRunes[i-1]) || unicode.IsUpper(r) {
				score += 10
			}

			// Bonus for match at start
			if i == 0 {
				score += 15
			}

			prevMatchIdx = i
			queryIdx++
		}
	}

	// Check if all query chars matched
	if queryIdx < len(queryRunes) {
		return 0, nil
	}

	// Base score for complete match
	score += 50

	// Bonus for shorter targets (more precise match)
	score += 100 - len(target)

	return score, matches
}

func (f *Finder) sortFiltered() {
	// Build category priority map
	catPriority := make(map[string]int)
	for _, cat := range f.categories {
		catPriority[cat.Name] = cat.Priority
	}

	// Build recent set
	recentSet := make(map[string]int)
	for i, id := range f.recentIDs {
		recentSet[id] = i
	}

	sort.SliceStable(f.filtered, func(i, j int) bool {
		itemI := f.filtered[i]
		itemJ := f.filtered[j]

		// Recent items first (when no query)
		if f.query == "" {
			recentI, okI := recentSet[itemI.ID]
			recentJ, okJ := recentSet[itemJ.ID]
			if okI && !okJ {
				return true
			}
			if !okI && okJ {
				return false
			}
			if okI && okJ {
				return recentI < recentJ
			}
		}

		// By category priority
		if f.showCategories {
			catI := catPriority[itemI.Category]
			catJ := catPriority[itemJ.Category]
			if catI != catJ {
				return catI < catJ
			}
		}

		// By match score (higher first)
		if itemI.Score != itemJ.Score {
			return itemI.Score > itemJ.Score
		}

		// Alphabetically
		return itemI.Label < itemJ.Label
	})
}

func (f *Finder) isRecent(id string) bool {
	for _, rid := range f.recentIDs {
		if rid == id {
			return true
		}
	}
	return false
}

// Draw renders the finder
func (f *Finder) Draw(screen tcell.Screen) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Set background color from theme
	f.Box.SetBackgroundColor(theme.Bg())
	f.Box.DrawForSubclass(screen, f)
	x, y, width, height := f.GetInnerRect()

	if width < 10 || height < 3 {
		return
	}

	// Get theme colors
	bg := theme.Bg()
	fg := theme.Fg()
	fgDim := theme.FgDim()
	fgMuted := theme.FgMuted()
	accent := theme.Accent()
	border := theme.Border()

	// Base style with theme background
	baseStyle := tcell.StyleDefault.Background(bg).Foreground(fg)

	// Calculate preview width if enabled
	listWidth := width
	previewWidth := 0
	previewX := x
	if f.previewFunc != nil && f.showPreview {
		previewWidth = int(float64(width) * f.previewRatio)
		listWidth = width - previewWidth - 1 // -1 for separator
		previewX = x + listWidth + 1
	}

	// Draw input line
	inputStyle := baseStyle
	promptStyle := baseStyle.Foreground(accent)

	// In vim mode, show "/" prompt when searching, otherwise show the normal prompt
	displayPrompt := f.prompt
	if f.vimMode && f.searching {
		displayPrompt = "/ "
	}

	// Draw prompt
	for i, r := range displayPrompt {
		screen.SetContent(x+i, y, r, nil, promptStyle)
	}

	// Draw query or placeholder
	inputX := x + len(displayPrompt)
	if f.query == "" && f.placeholder != "" {
		placeholderStyle := baseStyle.Foreground(fgMuted)
		for i, r := range f.placeholder {
			if inputX+i >= x+listWidth {
				break
			}
			screen.SetContent(inputX+i, y, r, nil, placeholderStyle)
		}
	} else {
		for i, r := range f.query {
			if inputX+i >= x+listWidth {
				break
			}
			screen.SetContent(inputX+i, y, r, nil, inputStyle)
		}
	}

	// Draw cursor (only when actively searching, or always in default mode)
	if !f.vimMode || f.searching {
		cursorX := inputX + len(f.query)
		if cursorX < x+listWidth {
			screen.SetContent(cursorX, y, '_', nil, baseStyle.Foreground(accent))
		}
	}

	// Draw separator
	separatorY := y + 1
	separatorStyle := baseStyle.Foreground(border)
	for i := 0; i < listWidth; i++ {
		screen.SetContent(x+i, separatorY, '─', nil, separatorStyle)
	}

	// Store colors for item drawing
	f.bgColor = bg
	f.fgColor = fg
	f.fgDimColor = fgDim
	f.accentColor = accent

	// Draw items
	itemsStartY := separatorY + 1
	itemsHeight := height - 3 // Subtract input, separator, footer

	visibleItems := f.maxVisible
	if itemsHeight < visibleItems {
		visibleItems = itemsHeight
	}

	// Adjust scroll offset
	if f.selectedIndex < f.scrollOffset {
		f.scrollOffset = f.selectedIndex
	}
	if f.selectedIndex >= f.scrollOffset+visibleItems {
		f.scrollOffset = f.selectedIndex - visibleItems + 1
	}

	lastCategory := ""
	itemsDrawn := 0

	for i := f.scrollOffset; i < len(f.filtered) && itemsDrawn < visibleItems; i++ {
		item := f.filtered[i]
		itemY := itemsStartY + itemsDrawn

		// Draw category header if changed
		if f.showCategories && item.Category != lastCategory && item.Category != "" {
			lastCategory = item.Category
			catStyle := baseStyle.Foreground(accent).Bold(true)
			catHeader := []rune("── " + item.Category + " ")
			col := 0
			for _, r := range catHeader {
				if col >= listWidth {
					break
				}
				screen.SetContent(x+col, itemY, r, nil, catStyle)
				col++
			}
			for col < listWidth {
				screen.SetContent(x+col, itemY, '─', nil, catStyle)
				col++
			}
			itemsDrawn++
			itemY++
			if itemsDrawn >= visibleItems {
				break
			}
		}

		// Draw item
		selected := i == f.selectedIndex
		f.drawItem(screen, x, itemY, listWidth, item, selected)
		itemsDrawn++
	}

	// Draw footer
	footerY := y + height - 1
	footerStyle := baseStyle.Foreground(fgMuted)
	footerText := ""
	if len(f.filtered) > 0 {
		footerText = itoa(len(f.filtered)) + " of " + itoa(len(f.items)) + " items"
	} else {
		footerText = "No matches"
	}
	for i, r := range footerText {
		screen.SetContent(x+i, footerY, r, nil, footerStyle)
	}

	// Draw hints on right side of footer
	var hints string
	if f.vimMode && !f.searching {
		hints = "[/] Search  [j/k] Navigate  [Enter] Select  [q] Cancel"
	} else {
		hints = "[Enter] Select  [Esc] Cancel"
		if f.previewFunc != nil {
			hints = "[Tab] Preview  " + hints
		}
	}
	hintsX := x + listWidth - len(hints)
	if hintsX > x+len(footerText)+2 {
		for i, r := range hints {
			screen.SetContent(hintsX+i, footerY, r, nil, footerStyle)
		}
	}

	// Draw preview if enabled
	if f.previewFunc != nil && f.showPreview && previewWidth > 0 {
		// Draw vertical separator
		for row := y; row < y+height; row++ {
			screen.SetContent(previewX-1, row, '│', nil, separatorStyle)
		}

		// Draw preview content
		if f.selectedIndex >= 0 && f.selectedIndex < len(f.filtered) {
			item := f.filtered[f.selectedIndex]
			preview := f.previewFunc(item)
			lines := strings.Split(preview, "\n")
			previewStyle := baseStyle
			for row, line := range lines {
				if row >= height {
					break
				}
				for col, r := range line {
					if col >= previewWidth {
						break
					}
					screen.SetContent(previewX+col, y+row, r, nil, previewStyle)
				}
			}
		}
	}
}

func (f *Finder) drawItem(screen tcell.Screen, x, y, width int, item FinderItem, selected bool) {
	// Use cached theme colors
	bg := f.bgColor
	fg := f.fgColor
	fgDim := f.fgDimColor
	accent := f.accentColor

	// Get selection colors from theme for high contrast
	selBg := theme.SelectionBg()
	selFg := theme.SelectionFg()

	style := tcell.StyleDefault.Background(bg).Foreground(fg)
	if selected {
		style = tcell.StyleDefault.Background(selBg).Foreground(selFg).Bold(true)
	}

	// Clear line
	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, ' ', nil, style)
	}

	col := x

	// Draw selection indicator
	if selected {
		screen.SetContent(col, y, '→', nil, style)
	} else {
		screen.SetContent(col, y, ' ', nil, style)
	}
	col += 2

	// Draw icon if enabled
	if f.showIcons && item.Icon != "" {
		for _, r := range item.Icon {
			screen.SetContent(col, y, r, nil, style)
			col++
		}
		col++ // Space after icon
	}

	// Draw label with match highlighting
	labelRunes := []rune(item.Label)
	matchSet := make(map[int]bool)
	for _, idx := range item.Matches {
		matchSet[idx] = true
	}

	for i, r := range labelRunes {
		if col >= x+width {
			break
		}
		charStyle := style
		if matchSet[i] && !selected {
			charStyle = style.Foreground(accent).Bold(true)
		}
		screen.SetContent(col, y, r, nil, charStyle)
		col++
	}

	// Draw description if enabled
	if f.showDescription && item.Description != "" {
		descStyle := style.Foreground(fgDim)
		if selected {
			// Use slightly lighter color for description when selected
			descStyle = style.Foreground(selFg).Bold(false)
		}

		// Calculate remaining space
		remainingWidth := (x + width) - col - 2
		if remainingWidth > 0 {
			col += 2 // Spacing
			desc := item.Description
			if len(desc) > remainingWidth {
				desc = desc[:remainingWidth-3] + "..."
			}
			for _, r := range desc {
				if col >= x+width {
					break
				}
				screen.SetContent(col, y, r, nil, descStyle)
				col++
			}
		}
	}

	// Draw recent badge
	if f.isRecent(item.ID) && f.query == "" {
		badgeText := "Recent"
		badgeX := x + width - len(badgeText) - 1
		if badgeX > col+2 {
			badgeStyle := style.Foreground(theme.Highlight())
			for i, r := range badgeText {
				screen.SetContent(badgeX+i, y, r, nil, badgeStyle)
			}
		}
	}
}

// InputHandler handles keyboard input
func (f *Finder) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return f.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		f.mu.Lock()
		defer f.mu.Unlock()

		if f.vimMode {
			f.handleVimInput(event)
		} else {
			f.handleDefaultInput(event)
		}
	})
}

func (f *Finder) handleDefaultInput(event *tcell.EventKey) {
	key := event.Key()
	switch key {
	case tcell.KeyDown, tcell.KeyCtrlN:
		f.moveDown()
	case tcell.KeyUp, tcell.KeyCtrlP:
		f.moveUp()
	case tcell.KeyPgDn:
		for i := 0; i < f.maxVisible; i++ {
			f.moveDown()
		}
	case tcell.KeyPgUp:
		for i := 0; i < f.maxVisible; i++ {
			f.moveUp()
		}
	case tcell.KeyHome:
		f.selectedIndex = 0
		f.scrollOffset = 0
		f.notifyChange()
	case tcell.KeyEnd:
		if len(f.filtered) > 0 {
			f.selectedIndex = len(f.filtered) - 1
		}
		f.notifyChange()
	case tcell.KeyEnter:
		f.selectCurrent()
	case tcell.KeyEsc:
		if f.onCancel != nil {
			go f.onCancel()
		}
	case tcell.KeyTab:
		if f.previewFunc != nil {
			f.showPreview = !f.showPreview
		}
	case tcell.KeyCtrlU:
		f.clearQuery()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		f.deleteChar()
	case tcell.KeyRune:
		f.appendChar(event.Rune())
	}
}

func (f *Finder) handleVimInput(event *tcell.EventKey) {
	key := event.Key()

	// In search mode, typing goes to query
	if f.searching {
		switch key {
		case tcell.KeyEsc:
			f.searching = false
		case tcell.KeyEnter:
			f.searching = false
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(f.query) > 0 {
				f.deleteChar()
			} else {
				f.searching = false
			}
		case tcell.KeyCtrlU:
			f.clearQuery()
		case tcell.KeyRune:
			f.appendChar(event.Rune())
		case tcell.KeyDown, tcell.KeyCtrlN:
			f.moveDown()
		case tcell.KeyUp, tcell.KeyCtrlP:
			f.moveUp()
		}
		return
	}

	// Normal vim mode — j/k navigate, / enters search
	switch key {
	case tcell.KeyDown, tcell.KeyCtrlN:
		f.moveDown()
	case tcell.KeyUp, tcell.KeyCtrlP:
		f.moveUp()
	case tcell.KeyEnter:
		f.selectCurrent()
	case tcell.KeyEsc:
		if f.query != "" {
			f.clearQuery()
		} else if f.onCancel != nil {
			go f.onCancel()
		}
	case tcell.KeyTab:
		if f.previewFunc != nil {
			f.showPreview = !f.showPreview
		}
	case tcell.KeyPgDn:
		for i := 0; i < f.maxVisible; i++ {
			f.moveDown()
		}
	case tcell.KeyPgUp:
		for i := 0; i < f.maxVisible; i++ {
			f.moveUp()
		}
	case tcell.KeyRune:
		switch event.Rune() {
		case '/':
			f.searching = true
		case 'j':
			f.moveDown()
		case 'k':
			f.moveUp()
		case 'g':
			f.selectedIndex = 0
			f.scrollOffset = 0
			f.notifyChange()
		case 'G':
			if len(f.filtered) > 0 {
				f.selectedIndex = len(f.filtered) - 1
			}
			f.notifyChange()
		case 'q':
			if f.onCancel != nil {
				go f.onCancel()
			}
		}
	}
}

func (f *Finder) selectCurrent() {
	if len(f.filtered) > 0 && f.selectedIndex >= 0 && f.selectedIndex < len(f.filtered) {
		item := f.filtered[f.selectedIndex]
		if f.onSelect != nil {
			go f.onSelect(item)
		}
	}
}

func (f *Finder) clearQuery() {
	f.query = ""
	f.filterItems()
	if f.onQueryChange != nil {
		go f.onQueryChange(f.query)
	}
}

func (f *Finder) deleteChar() {
	if len(f.query) > 0 {
		runes := []rune(f.query)
		f.query = string(runes[:len(runes)-1])
		f.filterItems()
		if f.onQueryChange != nil {
			go f.onQueryChange(f.query)
		}
	}
}

func (f *Finder) appendChar(r rune) {
	f.query += string(r)
	f.filterItems()
	if f.onQueryChange != nil {
		go f.onQueryChange(f.query)
	}
}

func (f *Finder) moveDown() {
	if f.selectedIndex < len(f.filtered)-1 {
		f.selectedIndex++
		f.notifyChange()
	}
}

func (f *Finder) moveUp() {
	if f.selectedIndex > 0 {
		f.selectedIndex--
		f.notifyChange()
	}
}

func (f *Finder) notifyChange() {
	if f.onChange != nil && f.selectedIndex >= 0 && f.selectedIndex < len(f.filtered) {
		item := f.filtered[f.selectedIndex]
		go f.onChange(item)
	}
}

// Focus is called when the finder receives focus
func (f *Finder) Focus(delegate func(p tview.Primitive)) {
	f.Box.Focus(delegate)
}

// HasFocus returns whether the finder has focus
func (f *Finder) HasFocus() bool {
	return f.Box.HasFocus()
}
