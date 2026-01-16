package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// MasterDetailConfig provides configuration for NewMasterDetailViewConfig.
type MasterDetailConfig struct {
	MasterTitle     string
	DetailTitle     string
	MasterContent   tview.Primitive
	DetailContent   tview.Primitive // nil shows empty state
	Ratio           float64         // 0.0-1.0, default 0.6
	DetailCollapsed bool            // Start with detail hidden
	Resizable       bool            // Allow keyboard resize (default: true)
	EmptyIcon       string
	EmptyTitle      string
	EmptyMessage    string
}

// MasterDetailView provides a two-panel layout with master list and detail preview.
// It combines Split, Panel, and EmptyState components to reduce boilerplate.
//
// Example usage:
//
//	view := components.NewMasterDetailView().
//	    SetMasterTitle("Workflows").
//	    SetDetailTitle("Preview").
//	    SetMasterContent(table).
//	    SetDetailContent(preview).
//	    SetRatio(0.6)
type MasterDetailView struct {
	*tview.Box

	// Internal layout
	split       *Split
	masterPanel *Panel
	detailPanel *Panel

	// Content
	masterContent     tview.Primitive
	detailContent     tview.Primitive
	detailPlaceholder tview.Primitive // Custom placeholder (overrides empty state)
	detailEmpty       *EmptyState     // Default empty state

	// Configuration
	masterTitle  string
	detailTitle  string
	showDetail   bool // Toggle detail pane visibility
	emptyIcon    string
	emptyTitle   string
	emptyMessage string

	// Focus tracking
	masterFocused bool

	// Search functionality
	searchEnabled   bool
	searchText      string
	baseMasterTitle string // Original title without search suffix
	onSearch        func(query string) // Called when search text changes
	onSearchSubmit  func(query string) // Called when search is submitted
	onSearchCancel  func()             // Called when search is cancelled
	showSearchFunc  func(currentText string, callbacks SearchCallbacks) // External search UI handler

	// Callbacks
	onStart           func()
	onStop            func()
	onSelectionChange func()
	onDetailToggle    func(bool)

	// Key hints (for nav.Component)
	hints []KeyHint
}

// SearchCallbacks provides callbacks for search UI integration.
type SearchCallbacks struct {
	OnChange func(text string)
	OnSubmit func(text string)
	OnCancel func()
}

// NewMasterDetailView creates a new master-detail layout.
func NewMasterDetailView() *MasterDetailView {
	box := tview.NewBox()
	box.SetBackgroundColor(theme.Bg())

	m := &MasterDetailView{
		Box:           box,
		split:         NewSplit(),
		masterPanel:   NewPanel(),
		detailPanel:   NewPanel(),
		detailEmpty:   NewEmptyState(),
		showDetail:    true,
		masterFocused: true,
		emptyIcon:     "󰋼", // Info icon
		emptyTitle:    "No Selection",
		emptyMessage:  "Select an item to view details",
	}

	// Configure split
	m.split.SetDirection(SplitHorizontal).
		SetRatio(0.6).
		SetResizable(true).
		SetShowDivider(false). // Panels have their own borders
		FocusFirst()

	// Set initial panel content
	m.updatePanels()

	// Register for automatic theme updates
	theme.Register(box)

	return m
}

// NewMasterDetailViewConfig creates a MasterDetailView with the given configuration.
func NewMasterDetailViewConfig(config MasterDetailConfig) *MasterDetailView {
	m := NewMasterDetailView()

	if config.MasterTitle != "" {
		m.SetMasterTitle(config.MasterTitle)
	}
	if config.DetailTitle != "" {
		m.SetDetailTitle(config.DetailTitle)
	}
	if config.MasterContent != nil {
		m.SetMasterContent(config.MasterContent)
	}
	if config.DetailContent != nil {
		m.SetDetailContent(config.DetailContent)
	}
	if config.Ratio > 0 {
		m.SetRatio(config.Ratio)
	}
	if config.DetailCollapsed {
		m.HideDetail()
	}
	if !config.Resizable {
		m.SetResizable(false)
	}
	if config.EmptyIcon != "" || config.EmptyTitle != "" || config.EmptyMessage != "" {
		m.ConfigureEmpty(config.EmptyIcon, config.EmptyTitle, config.EmptyMessage)
	}

	return m
}

// updatePanels synchronizes panel content with the current state.
func (m *MasterDetailView) updatePanels() {
	// Update master panel
	m.masterPanel.SetTitle(m.masterTitle)
	if m.masterContent != nil {
		m.masterPanel.SetContent(m.masterContent)
	}

	// Update detail panel
	m.detailPanel.SetTitle(m.detailTitle)
	m.updateDetailContent()

	// Update split panes
	m.split.SetFirst(m.masterPanel)
	if m.showDetail {
		m.split.SetSecond(m.detailPanel)
	} else {
		m.split.SetSecond(nil)
	}
}

// updateDetailContent sets the appropriate content in the detail panel.
func (m *MasterDetailView) updateDetailContent() {
	var content tview.Primitive

	if m.detailContent != nil {
		content = m.detailContent
	} else if m.detailPlaceholder != nil {
		content = m.detailPlaceholder
	} else {
		// Use default empty state
		m.detailEmpty.Configure(m.emptyIcon, m.emptyTitle, m.emptyMessage)
		content = m.detailEmpty
	}

	m.detailPanel.SetContent(content)
}

// --- Content Methods ---

// SetMasterTitle sets the master panel's title.
// If search is enabled, this also updates the base title used for search display.
func (m *MasterDetailView) SetMasterTitle(title string) *MasterDetailView {
	m.masterTitle = title
	if m.searchEnabled {
		m.baseMasterTitle = title
		m.updateSearchTitle() // Re-apply search suffix if any
	} else {
		m.masterPanel.SetTitle(title)
	}
	return m
}

// SetDetailTitle sets the detail panel's title.
func (m *MasterDetailView) SetDetailTitle(title string) *MasterDetailView {
	m.detailTitle = title
	m.detailPanel.SetTitle(title)
	return m
}

// SetMasterContent sets the master panel's content.
func (m *MasterDetailView) SetMasterContent(content tview.Primitive) *MasterDetailView {
	m.masterContent = content
	m.masterPanel.SetContent(content)
	return m
}

// SetDetailContent sets the detail panel's content.
// Pass nil to show the empty state or placeholder.
func (m *MasterDetailView) SetDetailContent(content tview.Primitive) *MasterDetailView {
	m.detailContent = content
	m.updateDetailContent()
	return m
}

// SetDetailPlaceholder sets a custom placeholder to show when detail content is nil.
// This overrides the default EmptyState component.
func (m *MasterDetailView) SetDetailPlaceholder(content tview.Primitive) *MasterDetailView {
	m.detailPlaceholder = content
	m.updateDetailContent()
	return m
}

// GetMasterContent returns the master panel's content.
func (m *MasterDetailView) GetMasterContent() tview.Primitive {
	return m.masterContent
}

// GetDetailContent returns the detail panel's content (may be nil).
func (m *MasterDetailView) GetDetailContent() tview.Primitive {
	return m.detailContent
}

// --- Empty State Configuration ---

// SetEmptyIcon sets the icon shown in the empty state.
func (m *MasterDetailView) SetEmptyIcon(icon string) *MasterDetailView {
	m.emptyIcon = icon
	m.updateDetailContent()
	return m
}

// SetEmptyTitle sets the title shown in the empty state.
func (m *MasterDetailView) SetEmptyTitle(title string) *MasterDetailView {
	m.emptyTitle = title
	m.updateDetailContent()
	return m
}

// SetEmptyMessage sets the message shown in the empty state.
func (m *MasterDetailView) SetEmptyMessage(message string) *MasterDetailView {
	m.emptyMessage = message
	m.updateDetailContent()
	return m
}

// ConfigureEmpty sets all empty state properties at once.
func (m *MasterDetailView) ConfigureEmpty(icon, title, message string) *MasterDetailView {
	if icon != "" {
		m.emptyIcon = icon
	}
	if title != "" {
		m.emptyTitle = title
	}
	if message != "" {
		m.emptyMessage = message
	}
	m.updateDetailContent()
	return m
}

// --- Layout Configuration ---

// SetRatio sets the split ratio (0.0 to 1.0, proportion of master pane).
func (m *MasterDetailView) SetRatio(ratio float64) *MasterDetailView {
	m.split.SetRatio(ratio)
	return m
}

// GetRatio returns the current split ratio.
func (m *MasterDetailView) GetRatio() float64 {
	return m.split.GetRatio()
}

// SetResizable enables or disables keyboard resizing.
func (m *MasterDetailView) SetResizable(resizable bool) *MasterDetailView {
	m.split.SetResizable(resizable)
	return m
}

// SetMinSize sets the minimum pane size in columns.
func (m *MasterDetailView) SetMinSize(size int) *MasterDetailView {
	m.split.SetMinSize(size)
	return m
}

// --- Visibility ---

// ShowDetail shows the detail pane.
func (m *MasterDetailView) ShowDetail() *MasterDetailView {
	if !m.showDetail {
		m.showDetail = true
		m.updatePanels()
		if m.onDetailToggle != nil {
			m.onDetailToggle(true)
		}
	}
	return m
}

// HideDetail hides the detail pane, showing only the master.
func (m *MasterDetailView) HideDetail() *MasterDetailView {
	if m.showDetail {
		m.showDetail = false
		m.masterFocused = true
		m.split.FocusFirst()
		m.updatePanels()
		if m.onDetailToggle != nil {
			m.onDetailToggle(false)
		}
	}
	return m
}

// ToggleDetail toggles detail pane visibility.
func (m *MasterDetailView) ToggleDetail() *MasterDetailView {
	if m.showDetail {
		return m.HideDetail()
	}
	return m.ShowDetail()
}

// SetDetailVisible sets detail pane visibility.
func (m *MasterDetailView) SetDetailVisible(visible bool) *MasterDetailView {
	if visible {
		return m.ShowDetail()
	}
	return m.HideDetail()
}

// IsDetailVisible returns whether the detail pane is visible.
func (m *MasterDetailView) IsDetailVisible() bool {
	return m.showDetail
}

// --- Focus Control ---

// FocusMaster focuses the master pane.
func (m *MasterDetailView) FocusMaster() *MasterDetailView {
	m.masterFocused = true
	m.split.FocusFirst()
	return m
}

// FocusDetail focuses the detail pane.
func (m *MasterDetailView) FocusDetail() *MasterDetailView {
	if m.showDetail {
		m.masterFocused = false
		m.split.FocusSecond()
	}
	return m
}

// ToggleFocus switches focus between master and detail panes.
func (m *MasterDetailView) ToggleFocus() *MasterDetailView {
	if !m.showDetail {
		return m
	}
	m.masterFocused = !m.masterFocused
	m.split.ToggleFocus()
	return m
}

// IsMasterFocused returns whether the master pane is focused.
func (m *MasterDetailView) IsMasterFocused() bool {
	return m.masterFocused
}

// --- Callbacks ---

// SetOnStart sets a callback for when the view becomes active.
func (m *MasterDetailView) SetOnStart(fn func()) *MasterDetailView {
	m.onStart = fn
	return m
}

// SetOnStop sets a callback for when the view becomes inactive.
func (m *MasterDetailView) SetOnStop(fn func()) *MasterDetailView {
	m.onStop = fn
	return m
}

// SetOnSelectionChange sets a callback for selection changes in the master list.
func (m *MasterDetailView) SetOnSelectionChange(fn func()) *MasterDetailView {
	m.onSelectionChange = fn
	return m
}

// SetOnDetailToggle sets a callback for when detail visibility changes.
func (m *MasterDetailView) SetOnDetailToggle(fn func(visible bool)) *MasterDetailView {
	m.onDetailToggle = fn
	return m
}

// SetOnResize sets a callback for when the split ratio changes.
func (m *MasterDetailView) SetOnResize(fn func(ratio float64)) *MasterDetailView {
	m.split.SetOnResize(fn)
	return m
}

// --- Search Functionality ---

// EnableSearch enables the built-in search functionality.
// When enabled, pressing '/' will trigger the search UI and the title
// will update to show the current search term (e.g., "Workflows (/term)").
//
// The showSearchFunc is called to display the search UI. It receives the
// current search text and callbacks for handling user input.
//
// Example usage:
//
//	view.EnableSearch(func(current string, cb components.SearchCallbacks) {
//	    app.ShowFilterMode(current, FilterModeCallbacks{
//	        OnChange: cb.OnChange,
//	        OnSubmit: cb.OnSubmit,
//	        OnCancel: cb.OnCancel,
//	    })
//	})
func (m *MasterDetailView) EnableSearch(showSearchFunc func(currentText string, callbacks SearchCallbacks)) *MasterDetailView {
	m.searchEnabled = true
	m.showSearchFunc = showSearchFunc
	m.baseMasterTitle = m.masterTitle
	return m
}

// SetOnSearch sets a callback that fires when the search text changes (live filtering).
func (m *MasterDetailView) SetOnSearch(fn func(query string)) *MasterDetailView {
	m.onSearch = fn
	return m
}

// SetOnSearchSubmit sets a callback that fires when search is submitted (Enter pressed).
func (m *MasterDetailView) SetOnSearchSubmit(fn func(query string)) *MasterDetailView {
	m.onSearchSubmit = fn
	return m
}

// SetOnSearchCancel sets a callback that fires when search is cancelled (Escape pressed).
func (m *MasterDetailView) SetOnSearchCancel(fn func()) *MasterDetailView {
	m.onSearchCancel = fn
	return m
}

// GetSearchText returns the current search text.
func (m *MasterDetailView) GetSearchText() string {
	return m.searchText
}

// SetSearchText sets the search text and updates the title.
func (m *MasterDetailView) SetSearchText(text string) *MasterDetailView {
	m.searchText = text
	m.updateSearchTitle()
	return m
}

// ClearSearch clears the search text and resets the title.
func (m *MasterDetailView) ClearSearch() *MasterDetailView {
	m.searchText = ""
	m.updateSearchTitle()
	if m.onSearchCancel != nil {
		m.onSearchCancel()
	}
	return m
}

// IsSearchEnabled returns whether search is enabled.
func (m *MasterDetailView) IsSearchEnabled() bool {
	return m.searchEnabled
}

// ShowSearch triggers the search UI if search is enabled.
func (m *MasterDetailView) ShowSearch() {
	if !m.searchEnabled || m.showSearchFunc == nil {
		return
	}

	m.showSearchFunc(m.searchText, SearchCallbacks{
		OnChange: func(text string) {
			m.searchText = text
			m.updateSearchTitle()
			if m.onSearch != nil {
				m.onSearch(text)
			}
		},
		OnSubmit: func(text string) {
			m.searchText = text
			m.updateSearchTitle()
			if m.onSearchSubmit != nil {
				m.onSearchSubmit(text)
			}
		},
		OnCancel: func() {
			if m.onSearchCancel != nil {
				m.onSearchCancel()
			}
		},
	})
}

// updateSearchTitle updates the master title to include search term.
func (m *MasterDetailView) updateSearchTitle() {
	if m.baseMasterTitle == "" {
		m.baseMasterTitle = m.masterTitle
	}

	if m.searchText == "" {
		m.masterPanel.SetTitle(m.baseMasterTitle)
	} else {
		m.masterPanel.SetTitle(m.baseMasterTitle + " (/" + m.searchText + ")")
	}
}

// HandleSearchKey checks if the event is a search trigger and handles it.
// Returns true if the event was handled.
func (m *MasterDetailView) HandleSearchKey(event *tcell.EventKey) bool {
	if !m.searchEnabled {
		return false
	}

	if event.Key() == tcell.KeyRune && event.Rune() == '/' {
		m.ShowSearch()
		return true
	}

	return false
}

// --- Key Hints (nav.Component support) ---

// SetHints sets the key hints for this view.
func (m *MasterDetailView) SetHints(hints []KeyHint) *MasterDetailView {
	m.hints = hints
	return m
}

// AddHint adds a single key hint.
func (m *MasterDetailView) AddHint(key, description string) *MasterDetailView {
	m.hints = append(m.hints, KeyHint{Key: key, Description: description})
	return m
}

// ClearHints removes all key hints.
func (m *MasterDetailView) ClearHints() *MasterDetailView {
	m.hints = nil
	return m
}

// --- nav.Component Interface ---

// Start is called when the view becomes active.
func (m *MasterDetailView) Start() {
	if m.onStart != nil {
		m.onStart()
	}
}

// Stop is called when the view becomes inactive.
func (m *MasterDetailView) Stop() {
	if m.onStop != nil {
		m.onStop()
	}
}

// Hints returns the configured key hints plus built-in hints.
func (m *MasterDetailView) Hints() []KeyHint {
	hints := make([]KeyHint, len(m.hints))
	copy(hints, m.hints)

	// Add built-in hints
	if m.showDetail {
		hints = append(hints, KeyHint{Key: "Tab", Description: "Switch pane"})
	}

	return hints
}

// --- tview.Primitive Implementation ---

// Draw renders the master-detail view.
func (m *MasterDetailView) Draw(screen tcell.Screen) {
	// Update background color from theme
	m.Box.SetBackgroundColor(theme.Bg())
	m.Box.DrawForSubclass(screen, m)

	x, y, width, height := m.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	if m.showDetail {
		// Draw split with both panels
		m.split.SetRect(x, y, width, height)
		m.split.Draw(screen)
	} else {
		// Draw only master panel
		m.masterPanel.SetRect(x, y, width, height)
		m.masterPanel.Draw(screen)
	}
}

// InputHandler handles keyboard input.
func (m *MasterDetailView) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if m.showDetail {
			// Delegate to split (handles Tab, Ctrl+arrows, etc.)
			if handler := m.split.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		} else {
			// Only master panel visible
			if handler := m.masterPanel.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

// MouseHandler handles mouse input.
func (m *MasterDetailView) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		if m.showDetail {
			if handler := m.split.MouseHandler(); handler != nil {
				return handler(action, event, setFocus)
			}
		} else {
			if handler := m.masterPanel.MouseHandler(); handler != nil {
				return handler(action, event, setFocus)
			}
		}
		return false, nil
	})
}

// Focus handles focus.
func (m *MasterDetailView) Focus(delegate func(tview.Primitive)) {
	if m.showDetail {
		delegate(m.split)
	} else {
		delegate(m.masterPanel)
	}
}

// Blur handles blur.
func (m *MasterDetailView) Blur() {
	m.split.Blur()
	m.masterPanel.Blur()
	m.detailPanel.Blur()
}

// HasFocus returns whether this view has focus.
func (m *MasterDetailView) HasFocus() bool {
	if m.showDetail {
		return m.split.HasFocus()
	}
	return m.masterPanel.HasFocus()
}

// GetRect returns the current drawing rectangle.
func (m *MasterDetailView) GetRect() (int, int, int, int) {
	return m.Box.GetRect()
}

// SetRect sets the drawing rectangle.
func (m *MasterDetailView) SetRect(x, y, width, height int) {
	m.Box.SetRect(x, y, width, height)
}
