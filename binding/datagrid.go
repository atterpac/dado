package binding

import (
	"sync"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
)

// DataGridBinding binds typed data to a DataGrid component.
// It supports both in-memory data (SetData) and lazy loading (SetFetcher).
type DataGridBinding[T any] struct {
	grid      *components.DataGrid
	data      []T
	colMapper func(T) []string                         // Convert item to column string values
	colNames  func() []string                          // Column names
	fetcher   func(start, count int) ([]T, int, error) // Lazy: fetch page, returns (items, total, err)

	mu sync.RWMutex
}

// NewDataGridBinding creates a new binding between typed data and a DataGrid.
func NewDataGridBinding[T any](grid *components.DataGrid) *DataGridBinding[T] {
	return &DataGridBinding[T]{
		grid: grid,
	}
}

// SetColMapper sets the function that converts each item to column string values.
func (b *DataGridBinding[T]) SetColMapper(fn func(T) []string) *DataGridBinding[T] {
	b.colMapper = fn
	return b
}

// SetColNames sets the function that returns column names.
func (b *DataGridBinding[T]) SetColNames(fn func() []string) *DataGridBinding[T] {
	b.colNames = fn
	return b
}

// SetFetcher sets the lazy loading function.
// The fetcher receives (start, count) and returns (items, totalCount, error).
func (b *DataGridBinding[T]) SetFetcher(fn func(start, count int) ([]T, int, error)) *DataGridBinding[T] {
	b.fetcher = fn
	return b
}

// SetData sets in-memory data and rebuilds the grid source.
func (b *DataGridBinding[T]) SetData(data []T) {
	b.mu.Lock()
	b.data = data
	b.mu.Unlock()

	b.rebuildSource()
}

// GetData returns a copy of the current data.
func (b *DataGridBinding[T]) GetData() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]T, len(b.data))
	copy(result, b.data)
	return result
}

// GetItem returns the item at the given row index.
func (b *DataGridBinding[T]) GetItem(row int) *T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if row < 0 || row >= len(b.data) {
		return nil
	}
	return &b.data[row]
}

// GetSelected returns the item at the cursor row.
func (b *DataGridBinding[T]) GetSelected() *T {
	row := b.grid.GetCursorRowIndex()
	return b.GetItem(row)
}

// GetSelectedValue returns the item at the cursor row directly.
func (b *DataGridBinding[T]) GetSelectedValue() (T, bool) {
	if item := b.GetSelected(); item != nil {
		return *item, true
	}
	var zero T
	return zero, false
}

// Count returns the number of items.
func (b *DataGridBinding[T]) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.data)
}

// ApplyChangeset returns all pending changes as a map of CellPosition to new value string.
// This is useful for generating SQL UPDATE statements or similar.
func (b *DataGridBinding[T]) ApplyChangeset() map[components.CellPosition]string {
	cs := b.grid.GetChangeset()
	changes := cs.Changes()
	result := make(map[components.CellPosition]string, len(changes))
	for _, change := range changes {
		result[change.Position] = change.NewValue
	}
	return result
}

// Refresh re-fetches data using the fetcher and updates the grid.
func (b *DataGridBinding[T]) Refresh() error {
	if b.fetcher == nil {
		return nil
	}

	b.mu.RLock()
	count := len(b.data)
	b.mu.RUnlock()

	if count == 0 {
		count = 100 // Default initial fetch size
	}

	items, _, err := b.fetcher(0, count)
	if err != nil {
		return err
	}

	b.SetData(items)
	return nil
}

// RefreshAsync fetches data asynchronously.
func (b *DataGridBinding[T]) RefreshAsync() {
	go func() {
		_ = b.Refresh()
	}()
}

// rebuildSource converts the current data into a SliceSource and sets it on the grid.
func (b *DataGridBinding[T]) rebuildSource() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.colMapper == nil || b.colNames == nil {
		return
	}

	colNames := b.colNames()
	cols := make([]components.GridColumn, len(colNames))
	for i, name := range colNames {
		cols[i] = components.GridColumn{Name: name}
	}

	rows := make([][]components.GridCell, len(b.data))
	for i, item := range b.data {
		values := b.colMapper(item)
		row := make([]components.GridCell, len(values))
		for j, val := range values {
			row[j] = components.GridCell{Value: val, RawValue: val}
		}
		rows[i] = row
	}

	source := components.NewSliceSource(cols, rows)

	theme.QueueUpdateDraw(func() {
		b.grid.SetSource(source)
	})
}

// bindingLazySource implements DataGridSource with lazy fetching via the binding's fetcher.
type bindingLazySource[T any] struct {
	binding *DataGridBinding[T]
	total   int
	colDefs []components.GridColumn
	cache   map[int][]components.GridCell
	cacheMu sync.RWMutex
}

// NewLazySource creates a lazy-loading DataGridSource backed by the binding's fetcher.
// This is an alternative to SetData for large datasets.
func (b *DataGridBinding[T]) NewLazySource(totalCount int) *bindingLazySource[T] {
	colNames := b.colNames()
	cols := make([]components.GridColumn, len(colNames))
	for i, name := range colNames {
		cols[i] = components.GridColumn{Name: name}
	}

	src := &bindingLazySource[T]{
		binding: b,
		total:   totalCount,
		colDefs: cols,
		cache:   make(map[int][]components.GridCell),
	}
	return src
}

func (s *bindingLazySource[T]) RowCount() int                    { return s.total }
func (s *bindingLazySource[T]) ColCount() int                    { return len(s.colDefs) }
func (s *bindingLazySource[T]) Columns() []components.GridColumn { return s.colDefs }

func (s *bindingLazySource[T]) Cell(row, col int) components.GridCell {
	s.cacheMu.RLock()
	if cells, ok := s.cache[row]; ok && col < len(cells) {
		s.cacheMu.RUnlock()
		return cells[col]
	}
	s.cacheMu.RUnlock()
	return components.GridCell{Value: "…"}
}

func (s *bindingLazySource[T]) FetchRange(startRow, endRow int) {
	if s.binding.fetcher == nil || s.binding.colMapper == nil {
		return
	}

	// Check if we already have this range cached
	s.cacheMu.RLock()
	allCached := true
	for i := startRow; i < endRow; i++ {
		if _, ok := s.cache[i]; !ok {
			allCached = false
			break
		}
	}
	s.cacheMu.RUnlock()
	if allCached {
		return
	}

	count := endRow - startRow
	items, total, err := s.binding.fetcher(startRow, count)
	if err != nil {
		return
	}

	if total > 0 {
		s.total = total
	}

	s.cacheMu.Lock()
	for i, item := range items {
		values := s.binding.colMapper(item)
		cells := make([]components.GridCell, len(values))
		for j, val := range values {
			cells[j] = components.GridCell{Value: val, RawValue: val}
		}
		s.cache[startRow+i] = cells
	}
	s.cacheMu.Unlock()
}
