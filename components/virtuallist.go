package components

import (
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
)

// VirtualListItem is a single row in a VirtualList. Height > 1 lets an item
// span multiple terminal rows; the RenderFunc receives the same width for all rows.
type VirtualListItem struct {
	ID     string // Unique identifier
	Data   any    // User data
	Height int    // Item height in rows (default 1)
}

// FetchFunc loads items on demand
// - start: index of first item to load
// - count: number of items to load
// Returns the items and total count
type FetchFunc func(start, count int) (items []VirtualListItem, total int)

// RenderFunc renders a single item
// - index: item index in the list
// - item: the item data
// - width: available width
// - selected: whether this item is selected
// Returns the string to display
type RenderFunc func(index int, item VirtualListItem, width int, selected bool) string

// VirtualList renders large lists efficiently by drawing only the rows in the
// visible viewport. Supply items up-front via SetItems or on-demand via
// SetFetchFunc for paginated or lazy-loaded data.
type VirtualList struct {
	widgetBase

	// Data
	items      []VirtualListItem
	totalCount int
	fetchFunc  FetchFunc
	renderFunc RenderFunc

	// State
	selectedIndex int
	offset        int // Scroll offset (first visible item)

	// Cache
	cache      map[int]*VirtualListItem
	cacheStart int
	cacheEnd   int
	cacheMu    sync.RWMutex

	// Options
	overscan          int // Items to render outside visible area
	showScrollbar     bool
	showIndex         bool
	defaultItemHeight int
	pageSize          int // Items per page for PgUp/PgDn (0 = visible height)

	// Callbacks
	onSelect    func(index int, item VirtualListItem)
	onChange    func(index int, item VirtualListItem)
	onScrollEnd func()
}

// NewVirtualList creates a new virtual list component
func NewVirtualList() *VirtualList {
	v := &VirtualList{
		cache:             make(map[int]*VirtualListItem),
		overscan:          5,
		showScrollbar:     true,
		defaultItemHeight: 1,
	}
	v.initWidget()
	return v
}

// SetItems sets all items directly (for smaller lists)
func (v *VirtualList) SetItems(items []VirtualListItem) *VirtualList {
	v.items = items
	v.totalCount = len(items)
	v.fetchFunc = nil
	v.clearCache()
	return v
}

// SetTotalCount sets the total number of items (for lazy loading)
func (v *VirtualList) SetTotalCount(count int) *VirtualList {
	v.totalCount = count
	return v
}

// SetFetchFunc sets the lazy loading function
func (v *VirtualList) SetFetchFunc(fn FetchFunc) *VirtualList {
	v.fetchFunc = fn
	v.items = nil // Clear static items when using fetch
	v.clearCache()
	return v
}

// SetRenderFunc sets the custom render function
func (v *VirtualList) SetRenderFunc(fn RenderFunc) *VirtualList {
	v.renderFunc = fn
	return v
}

// SetDefaultItemHeight sets height for items (default 1)
func (v *VirtualList) SetDefaultItemHeight(height int) *VirtualList {
	v.defaultItemHeight = height
	return v
}

// SetShowScrollbar enables/disables scrollbar
func (v *VirtualList) SetShowScrollbar(show bool) *VirtualList {
	v.showScrollbar = show
	return v
}

// SetShowIndex shows item index/number
func (v *VirtualList) SetShowIndex(show bool) *VirtualList {
	v.showIndex = show
	return v
}

// SetOverscan sets how many items to render outside visible area
func (v *VirtualList) SetOverscan(count int) *VirtualList {
	v.overscan = count
	return v
}

// SetPageSize sets items per page for PgUp/PgDn
func (v *VirtualList) SetPageSize(size int) *VirtualList {
	v.pageSize = size
	return v
}

// SetOnSelect is called when Enter is pressed on an item
func (v *VirtualList) SetOnSelect(fn func(index int, item VirtualListItem)) *VirtualList {
	v.onSelect = fn
	return v
}

// SetOnChange is called when selection changes
func (v *VirtualList) SetOnChange(fn func(index int, item VirtualListItem)) *VirtualList {
	v.onChange = fn
	return v
}

// SetOnScrollEnd is called when scrolled to the end
func (v *VirtualList) SetOnScrollEnd(fn func()) *VirtualList {
	v.onScrollEnd = fn
	return v
}

// GetSelectedIndex returns the currently selected index
func (v *VirtualList) GetSelectedIndex() int {
	return v.selectedIndex
}

// GetSelectedItem returns the currently selected item
func (v *VirtualList) GetSelectedItem() *VirtualListItem {
	return v.getItem(v.selectedIndex)
}

// SetSelectedIndex sets selection by index
func (v *VirtualList) SetSelectedIndex(index int) *VirtualList {
	if index >= 0 && index < v.totalCount {
		v.selectedIndex = index
		v.ensureVisible()
		v.triggerOnChange()
	}
	return v
}

// SetSelectedID sets selection by item ID
func (v *VirtualList) SetSelectedID(id string) *VirtualList {
	for i := 0; i < v.totalCount; i++ {
		item := v.getItem(i)
		if item != nil && item.ID == id {
			v.selectedIndex = i
			v.ensureVisible()
			v.triggerOnChange()
			break
		}
	}
	return v
}

// ScrollTo scrolls to show a specific index
func (v *VirtualList) ScrollTo(index int) {
	if index < 0 {
		index = 0
	}
	if index >= v.totalCount {
		index = v.totalCount - 1
	}
	v.offset = index
}

// ScrollToTop scrolls to the first item
func (v *VirtualList) ScrollToTop() {
	v.offset = 0
	v.selectedIndex = 0
	v.triggerOnChange()
}

// ScrollToBottom scrolls to the last item
func (v *VirtualList) ScrollToBottom() {
	_, _, _, height := v.GetInnerRect()
	if height <= 0 {
		height = 20
	}
	if v.totalCount > height {
		v.offset = v.totalCount - height
	}
	v.selectedIndex = v.totalCount - 1
	v.triggerOnChange()
}

// GetItem returns item at index (may trigger fetch)
func (v *VirtualList) GetItem(index int) *VirtualListItem {
	return v.getItem(index)
}

// GetVisibleRange returns currently visible index range
func (v *VirtualList) GetVisibleRange() (start, end int) {
	_, _, _, height := v.GetInnerRect()
	start = v.offset
	end = v.offset + height
	if end > v.totalCount {
		end = v.totalCount
	}
	return
}

// GetTotalCount returns total number of items
func (v *VirtualList) GetTotalCount() int {
	return v.totalCount
}

// Refresh re-fetches visible items and redraws
func (v *VirtualList) Refresh() {
	v.clearCache()
}

// Clear removes all items
func (v *VirtualList) Clear() {
	v.items = nil
	v.totalCount = 0
	v.selectedIndex = 0
	v.offset = 0
	v.clearCache()
}

func (v *VirtualList) clearCache() {
	v.cacheMu.Lock()
	v.cache = make(map[int]*VirtualListItem)
	v.cacheStart = 0
	v.cacheEnd = 0
	v.cacheMu.Unlock()
}

func (v *VirtualList) getItem(index int) *VirtualListItem {
	if index < 0 || index >= v.totalCount {
		return nil
	}

	// Static items
	if v.items != nil && index < len(v.items) {
		return &v.items[index]
	}

	// Check cache
	v.cacheMu.RLock()
	if item, ok := v.cache[index]; ok {
		v.cacheMu.RUnlock()
		return item
	}
	v.cacheMu.RUnlock()

	// Fetch if function available
	if v.fetchFunc != nil {
		// Fetch a batch around the requested index
		batchSize := 50
		start := index - batchSize/2
		if start < 0 {
			start = 0
		}

		items, total := v.fetchFunc(start, batchSize)
		if total > 0 {
			v.totalCount = total
		}

		v.cacheMu.Lock()
		// Evict cache entries far from current fetch window
		maxCacheSize := batchSize * 3
		if len(v.cache) > maxCacheSize {
			for k := range v.cache {
				if k < start-batchSize || k > start+batchSize*2 {
					delete(v.cache, k)
				}
			}
		}
		for i, item := range items {
			itemCopy := item
			v.cache[start+i] = &itemCopy
		}
		v.cacheMu.Unlock()

		// Return the requested item
		v.cacheMu.RLock()
		item := v.cache[index]
		v.cacheMu.RUnlock()
		return item
	}

	return nil
}

func (v *VirtualList) ensureVisible() {
	_, _, _, height := v.GetInnerRect()
	if height <= 0 {
		return
	}

	if v.selectedIndex < v.offset {
		v.offset = v.selectedIndex
	}
	if v.selectedIndex >= v.offset+height {
		v.offset = v.selectedIndex - height + 1
	}
}

func (v *VirtualList) triggerOnChange() {
	if v.onChange != nil {
		item := v.getItem(v.selectedIndex)
		if item != nil {
			v.onChange(v.selectedIndex, *item)
		}
	}
}

// vlSnapshot is the immutable geometry the paint phase consumes.
type vlSnapshot struct {
	contentWidth int
	indexWidth   int
	bg, fg       tcell.Color
	fgDim        tcell.Color
	accent       tcell.Color
}

// Draw renders the virtual list. State mutation (ensureVisible,
// prefetch) is contained in prepareDraw; paint writes only to screen.
func (v *VirtualList) Draw(screen tcell.Screen) {
	v.Box.DrawForSubclass(screen)
	x, y, width, height := v.GetInnerRect()
	if width <= 0 || height <= 0 || v.totalCount == 0 {
		return
	}

	snap := v.prepareDraw(width, height)
	v.paint(screen, x, y, width, height, snap)

	if v.offset+height >= v.totalCount && v.onScrollEnd != nil {
		v.onScrollEnd()
	}
}

// prepareDraw fits the viewport, prefetches the visible window, and
// returns the geometry the paint phase needs.
func (v *VirtualList) prepareDraw(width, height int) vlSnapshot {
	th := v.th()
	snap := vlSnapshot{
		contentWidth: width,
		bg:           th.Bg(),
		fg:           th.Fg(),
		fgDim:        th.FgDim(),
		accent:       th.Accent(),
	}
	if v.showScrollbar && v.totalCount > height {
		snap.contentWidth--
	}

	v.ensureVisible()

	startFetch := v.offset - v.overscan
	if startFetch < 0 {
		startFetch = 0
	}
	endFetch := v.offset + height + v.overscan
	if endFetch > v.totalCount {
		endFetch = v.totalCount
	}
	for i := startFetch; i < endFetch; i++ {
		v.getItem(i)
	}

	if v.showIndex {
		snap.indexWidth = len(string(rune(v.totalCount))) + 2
		if snap.indexWidth < 4 {
			snap.indexWidth = 4
		}
	}
	return snap
}

// paint writes the list rows to screen. Does not mutate v.* state.
func (v *VirtualList) paint(screen tcell.Screen, x, y, width, height int, snap vlSnapshot) {
	for i := 0; i < height && v.offset+i < v.totalCount; i++ {
		index := v.offset + i
		item := v.getItem(index)
		rowY := y + i
		isSelected := index == v.selectedIndex

		rowStyle := tcell.StyleDefault.Background(snap.bg).Foreground(snap.fg)
		if isSelected {
			rowStyle = tcell.StyleDefault.Background(snap.accent).Foreground(snap.bg)
		}

		fillLine(screen, x, rowY, snap.contentWidth, rowStyle)

		col := x

		if v.showIndex {
			indexStyle := rowStyle
			if !isSelected {
				indexStyle = tcell.StyleDefault.Background(snap.bg).Foreground(snap.fgDim)
			}
			indexStr := padLeft(itoa(index+1), snap.indexWidth-1) + " "
			for _, ch := range indexStr {
				if col < x+snap.contentWidth {
					screen.SetContent(col, rowY, ch, nil, indexStyle)
					col++
				}
			}
		}

		if item != nil {
			var text string
			if v.renderFunc != nil {
				text = v.renderFunc(index, *item, snap.contentWidth-snap.indexWidth, isSelected)
			} else if item.ID != "" {
				text = item.ID
			} else if item.Data != nil {
				text = toString(item.Data)
			}

			col += core.PrintTagged(screen, text, col, rowY, x+snap.contentWidth-col, rowStyle)
		}
	}

	if v.showScrollbar && v.totalCount > height {
		v.drawScrollbar(screen, x+width-1, y, height)
	}
}

func (v *VirtualList) drawScrollbar(screen tcell.Screen, x, y, height int) {
	th := v.th()
	bgColor := th.BgLight()
	thumbColor := th.FgDim()

	// Calculate thumb size and position
	thumbSize := height * height / v.totalCount
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > height {
		thumbSize = height
	}

	thumbPos := 0
	if v.totalCount > height {
		thumbPos = v.offset * (height - thumbSize) / (v.totalCount - height)
	}

	// Draw track
	trackStyle := tcell.StyleDefault.Background(bgColor)
	for i := 0; i < height; i++ {
		screen.SetContent(x, y+i, ' ', nil, trackStyle)
	}

	// Draw thumb
	thumbStyle := tcell.StyleDefault.Background(thumbColor)
	for i := 0; i < thumbSize; i++ {
		if thumbPos+i < height {
			screen.SetContent(x, y+thumbPos+i, ' ', nil, thumbStyle)
		}
	}
}

// HandleKey processes a key event for the VirtualList.
func (v *VirtualList) HandleKey(ev *tcell.EventKey) bool {
	if v.totalCount == 0 {
		return false
	}

	prevIndex := v.selectedIndex

	switch ev.Key() {
	case tcell.KeyDown:
		v.moveDown()
	case tcell.KeyUp:
		v.moveUp()
	case tcell.KeyHome:
		v.selectedIndex = 0
		v.ensureVisible()
	case tcell.KeyEnd:
		v.selectedIndex = v.totalCount - 1
		v.ensureVisible()
	case tcell.KeyPgDn:
		v.pageDown()
	case tcell.KeyPgUp:
		v.pageUp()
	case tcell.KeyEnter:
		if v.onSelect != nil {
			item := v.getItem(v.selectedIndex)
			if item != nil {
				v.onSelect(v.selectedIndex, *item)
			}
		}
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j':
			v.moveDown()
		case 'k':
			v.moveUp()
		case 'g':
			v.selectedIndex = 0
			v.ensureVisible()
		case 'G':
			v.selectedIndex = v.totalCount - 1
			v.ensureVisible()
		}
	case tcell.KeyCtrlD:
		v.halfPageDown()
	case tcell.KeyCtrlU:
		v.halfPageUp()
	}

	if v.selectedIndex != prevIndex {
		v.triggerOnChange()
	}
	return false
}

func (v *VirtualList) moveDown() {
	if v.selectedIndex < v.totalCount-1 {
		v.selectedIndex++
		v.ensureVisible()
	}
}

func (v *VirtualList) moveUp() {
	if v.selectedIndex > 0 {
		v.selectedIndex--
		v.ensureVisible()
	}
}

func (v *VirtualList) pageDown() {
	_, _, _, height := v.GetInnerRect()
	pageSize := v.pageSize
	if pageSize <= 0 {
		pageSize = height
	}
	v.selectedIndex += pageSize
	if v.selectedIndex >= v.totalCount {
		v.selectedIndex = v.totalCount - 1
	}
	v.ensureVisible()
}

func (v *VirtualList) pageUp() {
	_, _, _, height := v.GetInnerRect()
	pageSize := v.pageSize
	if pageSize <= 0 {
		pageSize = height
	}
	v.selectedIndex -= pageSize
	if v.selectedIndex < 0 {
		v.selectedIndex = 0
	}
	v.ensureVisible()
}

func (v *VirtualList) halfPageDown() {
	_, _, _, height := v.GetInnerRect()
	v.selectedIndex += height / 2
	if v.selectedIndex >= v.totalCount {
		v.selectedIndex = v.totalCount - 1
	}
	v.ensureVisible()
}

func (v *VirtualList) halfPageUp() {
	_, _, _, height := v.GetInnerRect()
	v.selectedIndex -= height / 2
	if v.selectedIndex < 0 {
		v.selectedIndex = 0
	}
	v.ensureVisible()
}

// Focus handles focus
// HasFocus returns whether the component has focus
func (v *VirtualList) HasFocus() bool {
	return v.Box.HasFocus()
}

// Helper functions

func padLeft(s string, length int) string {
	for len(s) < length {
		s = " " + s
	}
	return s
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if s, ok := v.(interface{ String() string }); ok {
		return s.String()
	}
	return ""
}
