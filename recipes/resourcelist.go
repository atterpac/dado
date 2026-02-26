package recipes

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/input"
	"github.com/atterpac/jig/nav"
	"github.com/atterpac/jig/theme"
)

// ResourceList is a K9s-style filterable resource list with actions.
// It provides a complete view for listing, filtering, and acting on resources.
type ResourceList[T any] struct {
	*tview.Flex

	// Components
	table     *components.Table
	filterBar *input.CommandBar
	hintBar   *components.KeyHintBar

	// Data
	resources   []T
	filtered    []T
	columns     []string
	rowMapper   func(T) []string
	filterQuery string

	// Actions
	actions  *input.ActionRegistry
	onSelect func(T)

	// Fetching
	fetcher         func() ([]T, error)
	refreshInterval time.Duration
	refreshTicker   *time.Ticker
	stopRefresh     chan struct{}
	lastError       error

	// State
	loading bool
	mu      sync.RWMutex
}

// NewResourceList creates a new ResourceList.
func NewResourceList[T any]() *ResourceList[T] {
	r := &ResourceList[T]{
		Flex:     tview.NewFlex().SetDirection(tview.FlexRow),
		table:    components.NewTable(),
		filterBar: input.NewCommandBar(),
		hintBar:  components.NewKeyHintBar(),
		actions:  input.NewActionRegistry(),
	}

	// Setup table
	r.table.SetMultiSelect(true)

	// Setup filter bar
	r.filterBar.Hide()
	r.filterBar.SetOnSubmit(func(cmdType input.CommandType, cmd string) {
		r.filterQuery = cmd
		r.applyFilter()
		r.filterBar.Hide()
	})
	r.filterBar.SetOnCancel(func() {
		r.filterBar.Hide()
	})
	r.filterBar.SetOnChange(func(cmd string) {
		// Live filtering
		r.filterQuery = cmd
		r.applyFilter()
	})

	// Default hints
	r.updateHints()

	// Layout
	r.AddItem(r.filterBar, 1, 0, false)
	r.AddItem(r.table, 0, 1, true)
	r.AddItem(r.hintBar, 1, 0, false)

	return r
}

// SetColumns sets the table columns.
func (r *ResourceList[T]) SetColumns(columns []string) *ResourceList[T] {
	r.columns = columns
	r.table.SetHeaders(columns...)
	return r
}

// SetRowMapper sets the function to convert resources to table rows.
func (r *ResourceList[T]) SetRowMapper(mapper func(T) []string) *ResourceList[T] {
	r.rowMapper = mapper
	return r
}

// SetFetcher sets the function to fetch resources.
func (r *ResourceList[T]) SetFetcher(fetcher func() ([]T, error)) *ResourceList[T] {
	r.fetcher = fetcher
	return r
}

// SetRefreshInterval sets the auto-refresh interval. Pass 0 to disable.
func (r *ResourceList[T]) SetRefreshInterval(interval time.Duration) *ResourceList[T] {
	r.refreshInterval = interval
	return r
}

// SetOnSelect sets the callback for when a resource is selected (Enter).
func (r *ResourceList[T]) SetOnSelect(fn func(T)) *ResourceList[T] {
	r.onSelect = fn
	r.table.SetOnSelect(func(row int) {
		if resource := r.getResourceAtRow(row); resource != nil {
			fn(*resource)
		}
	})
	return r
}

// AddAction adds a single-resource action.
func (r *ResourceList[T]) AddAction(key rune, description string, handler func(T)) *ResourceList[T] {
	r.actions.AddSimple(string(key), key, description, func() {
		if resource := r.GetSelected(); resource != nil {
			handler(*resource)
		}
	})
	r.updateHints()
	return r
}

// AddBulkAction adds a multi-resource action.
func (r *ResourceList[T]) AddBulkAction(key rune, description string, handler func([]T)) *ResourceList[T] {
	actionKey := "bulk_" + string(key)
	r.actions.AddSimple(actionKey, key, description, func() {
		selected := r.GetSelectedResources()
		if len(selected) > 0 {
			handler(selected)
		}
	})
	r.updateHints()
	return r
}

// GetSelected returns the currently highlighted resource.
func (r *ResourceList[T]) GetSelected() *T {
	row, _ := r.table.GetSelection()
	return r.getResourceAtRow(row)
}

// GetSelectedResources returns all multi-selected resources.
func (r *ResourceList[T]) GetSelectedResources() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	selectedRows := r.table.GetSelectedRows()
	var resources []T
	for _, row := range selectedRows {
		if resource := r.getResourceAtRow(row); resource != nil {
			resources = append(resources, *resource)
		}
	}
	return resources
}

func (r *ResourceList[T]) getResourceAtRow(row int) *T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Account for header
	dataRow := row - 1
	if dataRow < 0 || dataRow >= len(r.filtered) {
		return nil
	}
	return &r.filtered[dataRow]
}

// Refresh fetches and updates the resource list.
func (r *ResourceList[T]) Refresh() error {
	if r.fetcher == nil {
		return nil
	}

	r.mu.Lock()
	r.loading = true
	r.mu.Unlock()

	resources, err := r.fetcher()

	r.mu.Lock()
	r.loading = false
	r.lastError = err
	if err == nil {
		r.resources = resources
	}
	r.mu.Unlock()

	if err != nil {
		return err
	}

	r.applyFilter()
	return nil
}

func (r *ResourceList[T]) applyFilter() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.filterQuery == "" {
		r.filtered = r.resources
	} else {
		r.filtered = nil
		for _, resource := range r.resources {
			if r.matchesFilter(resource) {
				r.filtered = append(r.filtered, resource)
			}
		}
	}

	r.rebuildTable()
}

func (r *ResourceList[T]) matchesFilter(resource T) bool {
	if r.rowMapper == nil {
		return true
	}
	row := r.rowMapper(resource)
	query := r.filterQuery
	for _, cell := range row {
		if containsIgnoreCase(cell, query) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return contains(sLower, substrLower)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (r *ResourceList[T]) rebuildTable() {
	r.table.ClearRows()

	for _, resource := range r.filtered {
		if r.rowMapper != nil {
			row := r.rowMapper(resource)
			r.table.AddRow(row...)
		}
	}
}

func (r *ResourceList[T]) updateHints() {
	hints := r.actions.Hints()
	// Add default hints
	defaultHints := []components.KeyHint{
		{Key: "/", Description: "Filter"},
		{Key: "Enter", Description: "Select"},
		{Key: "Space", Description: "Multi-select"},
		{Key: "r", Description: "Refresh"},
	}
	hints = append(defaultHints, hints...)
	r.hintBar.SetHints(hints)
}

// Name returns the display name for breadcrumbs.
func (r *ResourceList[T]) Name() string { return "Resources" }

// Start begins the resource list lifecycle (fetching, refresh).
func (r *ResourceList[T]) Start() {
	// Initial fetch
	go func() {
		r.Refresh()
		theme.QueueUpdateDraw(func() {})
	}()

	// Start refresh ticker if interval is set
	if r.refreshInterval > 0 {
		r.stopRefresh = make(chan struct{})
		r.refreshTicker = time.NewTicker(r.refreshInterval)
		go func() {
			for {
				select {
				case <-r.stopRefresh:
					return
				case <-r.refreshTicker.C:
					r.Refresh()
					theme.QueueUpdateDraw(func() {})
				}
			}
		}()
	}
}

// Stop ends the resource list lifecycle.
func (r *ResourceList[T]) Stop() {
	if r.refreshTicker != nil {
		r.refreshTicker.Stop()
	}
	if r.stopRefresh != nil {
		close(r.stopRefresh)
	}
}

// Hints returns the current key hints.
func (r *ResourceList[T]) Hints() []components.KeyHint {
	return r.hintBar.Hints
}

// Draw renders the resource list.
func (r *ResourceList[T]) Draw(screen tcell.Screen) {
	r.Flex.Draw(screen)
}

// InputHandler handles keyboard input.
func (r *ResourceList[T]) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		// Handle filter bar if visible
		if r.filterBar.IsVisible() {
			if handler := r.filterBar.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
			return
		}

		// Check for filter key
		if event.Key() == tcell.KeyRune && event.Rune() == '/' {
			r.filterBar.Show(input.CommandTypeFilter)
			return
		}

		// Check for refresh
		if event.Key() == tcell.KeyRune && event.Rune() == 'r' {
			go func() {
				r.Refresh()
				theme.QueueUpdateDraw(func() {})
			}()
			return
		}

		// Check for clear filter
		if event.Key() == tcell.KeyEscape {
			if r.filterQuery != "" {
				r.filterQuery = ""
				r.applyFilter()
				return
			}
		}

		// Check registered actions
		if r.actions.Handle(event) {
			return
		}

		// Pass to table
		if handler := r.table.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

// Focus handles focus.
func (r *ResourceList[T]) Focus(delegate func(tview.Primitive)) {
	if r.filterBar.IsVisible() {
		delegate(r.filterBar)
	} else {
		delegate(r.table)
	}
}

// HasFocus returns whether the list has focus.
func (r *ResourceList[T]) HasFocus() bool {
	return r.table.HasFocus() || r.filterBar.HasFocus()
}

// Ensure ResourceList implements nav.Component
var _ nav.Component = (*ResourceList[any])(nil)
