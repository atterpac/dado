package binding

import (
	"sync"
	"time"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
)

// TableBinding binds a slice of data to a Table component.
// It provides automatic data mapping, filtering, fetching, and selection handling.
type TableBinding[T any] struct {
	table      *components.Table
	data       []T
	filtered   []T
	mapper     func(T) []string
	keyMapper  func(T) string
	filter     func(T, string) bool
	filterText string

	// Auto-refresh
	fetcher         func() ([]T, error)
	refreshInterval time.Duration
	refreshTicker   *time.Ticker
	stopRefresh     chan struct{}
	lastError       error

	// Callbacks
	onRefresh func([]T, error)
	onSelect  func(T)

	// State
	loading bool
	started bool
	mu      sync.RWMutex
}

// NewTableBinding creates a binding between data and a table.
func NewTableBinding[T any](table *components.Table) *TableBinding[T] {
	return &TableBinding[T]{
		table:       table,
		stopRefresh: make(chan struct{}),
	}
}

// SetMapper sets the function that converts each item to table row strings.
func (b *TableBinding[T]) SetMapper(fn func(T) []string) *TableBinding[T] {
	b.mapper = fn
	return b
}

// SetKeyMapper sets the function that returns a unique key for each item.
// This enables stable selection across data updates.
func (b *TableBinding[T]) SetKeyMapper(fn func(T) string) *TableBinding[T] {
	b.keyMapper = fn
	return b
}

// SetFilter sets the filter function (item, query) -> matches.
func (b *TableBinding[T]) SetFilter(fn func(T, string) bool) *TableBinding[T] {
	b.filter = fn
	return b
}

// SetFetcher sets the data source function for pull-based updates.
func (b *TableBinding[T]) SetFetcher(fn func() ([]T, error)) *TableBinding[T] {
	b.fetcher = fn
	return b
}

// SetRefreshInterval enables auto-refresh at the given interval.
// Pass 0 to disable auto-refresh.
func (b *TableBinding[T]) SetRefreshInterval(interval time.Duration) *TableBinding[T] {
	b.refreshInterval = interval
	return b
}

// SetOnRefresh sets callback after data refresh completes.
func (b *TableBinding[T]) SetOnRefresh(fn func([]T, error)) *TableBinding[T] {
	b.onRefresh = fn
	return b
}

// SetOnSelect sets callback when a table row is selected (Enter key).
func (b *TableBinding[T]) SetOnSelect(fn func(T)) *TableBinding[T] {
	b.onSelect = fn
	b.table.SetOnSelect(func(row int) {
		if item := b.GetItem(row); item != nil {
			fn(*item)
		}
	})
	return b
}

// SetData directly sets the bound data (push update).
// This triggers a table rebuild with the new data.
func (b *TableBinding[T]) SetData(data []T) {
	b.mu.Lock()
	b.data = data
	b.mu.Unlock()
	b.applyFilter()
}

// GetData returns a copy of the current data.
func (b *TableBinding[T]) GetData() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]T, len(b.data))
	copy(result, b.data)
	return result
}

// GetFiltered returns a copy of the currently filtered data.
func (b *TableBinding[T]) GetFiltered() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]T, len(b.filtered))
	copy(result, b.filtered)
	return result
}

// GetItem returns the item at the given table row (accounting for header).
// Returns nil if out of bounds.
func (b *TableBinding[T]) GetItem(row int) *T {
	b.mu.RLock()
	defer b.mu.RUnlock()

	dataRow := row - 1 // Account for header
	if dataRow < 0 || dataRow >= len(b.filtered) {
		return nil
	}
	return &b.filtered[dataRow]
}

// GetSelected returns the currently highlighted item.
func (b *TableBinding[T]) GetSelected() *T {
	row, _ := b.table.GetSelection()
	return b.GetItem(row)
}

// GetSelectedItems returns all multi-selected items.
func (b *TableBinding[T]) GetSelectedItems() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()

	selectedRows := b.table.GetSelectedRows()
	items := make([]T, 0, len(selectedRows))
	for _, row := range selectedRows {
		dataRow := row - 1
		if dataRow >= 0 && dataRow < len(b.filtered) {
			items = append(items, b.filtered[dataRow])
		}
	}
	return items
}

// Filter applies a filter query to the data.
func (b *TableBinding[T]) Filter(query string) {
	b.mu.Lock()
	b.filterText = query
	b.mu.Unlock()
	b.applyFilter()
}

// ClearFilter removes the current filter.
func (b *TableBinding[T]) ClearFilter() {
	b.Filter("")
}

// Refresh fetches new data from the fetcher.
// This is a blocking call - use RefreshAsync for non-blocking refresh.
func (b *TableBinding[T]) Refresh() error {
	if b.fetcher == nil {
		return nil
	}

	b.mu.Lock()
	b.loading = true
	b.mu.Unlock()

	data, err := b.fetcher()

	b.mu.Lock()
	b.loading = false
	b.lastError = err
	if err == nil {
		b.data = data
	}
	b.mu.Unlock()

	if err == nil {
		b.applyFilter()
	}

	if b.onRefresh != nil {
		b.onRefresh(data, err)
	}

	return err
}

// RefreshAsync fetches data asynchronously and updates the table.
func (b *TableBinding[T]) RefreshAsync() {
	go func() {
		err := b.Refresh()
		theme.QueueUpdateDraw(func() {
			// Trigger redraw after data update
			_ = err
		})
	}()
}

// Start begins the binding lifecycle (initial fetch, auto-refresh).
func (b *TableBinding[T]) Start() {
	if b.started {
		return
	}
	b.started = true

	// Initial fetch
	if b.fetcher != nil {
		b.RefreshAsync()
	}

	// Start refresh ticker if interval is set
	if b.refreshInterval > 0 {
		b.refreshTicker = time.NewTicker(b.refreshInterval)
		go func() {
			for {
				select {
				case <-b.stopRefresh:
					return
				case <-b.refreshTicker.C:
					b.RefreshAsync()
				}
			}
		}()
	}
}

// Stop halts auto-refresh and cleans up.
func (b *TableBinding[T]) Stop() {
	if b.refreshTicker != nil {
		b.refreshTicker.Stop()
	}
	if b.started {
		close(b.stopRefresh)
		b.stopRefresh = make(chan struct{})
		b.started = false
	}
}

// IsLoading returns whether a fetch is in progress.
func (b *TableBinding[T]) IsLoading() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.loading
}

// LastError returns the error from the last fetch, if any.
func (b *TableBinding[T]) LastError() error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastError
}

// Count returns the number of items (after filtering).
func (b *TableBinding[T]) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.filtered)
}

// TotalCount returns the total number of items (before filtering).
func (b *TableBinding[T]) TotalCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.data)
}

// applyFilter filters data and rebuilds the table.
func (b *TableBinding[T]) applyFilter() {
	b.mu.Lock()

	if b.filterText == "" || b.filter == nil {
		b.filtered = b.data
	} else {
		b.filtered = nil
		for _, item := range b.data {
			if b.filter(item, b.filterText) {
				b.filtered = append(b.filtered, item)
			}
		}
	}

	b.mu.Unlock()

	theme.QueueUpdate(func() {
		b.rebuildTable()
	})
}

// rebuildTable clears and repopulates the table from filtered data.
func (b *TableBinding[T]) rebuildTable() {
	b.table.ClearRows()

	b.mu.RLock()
	defer b.mu.RUnlock()

	for i, item := range b.filtered {
		if b.mapper != nil {
			b.table.AddRow(b.mapper(item)...)
			// Set row key if keyMapper is provided
			if b.keyMapper != nil {
				b.table.SetRowKey(i, b.keyMapper(item))
			}
		}
	}
}

// UpdateItem updates a single item in the data and refreshes the table.
// Uses keyMapper to find the item. Returns false if item not found.
func (b *TableBinding[T]) UpdateItem(newItem T) bool {
	if b.keyMapper == nil {
		return false
	}

	newKey := b.keyMapper(newItem)

	b.mu.Lock()
	found := false
	for i, item := range b.data {
		if b.keyMapper(item) == newKey {
			b.data[i] = newItem
			found = true
			break
		}
	}
	b.mu.Unlock()

	if found {
		b.applyFilter()
	}
	return found
}

// RemoveItem removes an item by key. Returns false if not found.
func (b *TableBinding[T]) RemoveItem(key string) bool {
	if b.keyMapper == nil {
		return false
	}

	b.mu.Lock()
	found := false
	for i, item := range b.data {
		if b.keyMapper(item) == key {
			b.data = append(b.data[:i], b.data[i+1:]...)
			found = true
			break
		}
	}
	b.mu.Unlock()

	if found {
		b.applyFilter()
	}
	return found
}

// AddItem adds an item to the data and refreshes the table.
func (b *TableBinding[T]) AddItem(item T) {
	b.mu.Lock()
	b.data = append(b.data, item)
	b.mu.Unlock()
	b.applyFilter()
}
