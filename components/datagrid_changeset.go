package components

import "sync"

// CellPosition identifies a cell in the grid by row and column index.
type CellPosition struct {
	Row int
	Col int
}

// CellChange records a single cell modification.
type CellChange struct {
	Position CellPosition
	OldValue string
	NewValue string
}

// Changeset tracks all pending cell modifications in a DataGrid.
// It maintains insertion order for review and supports overlay reads.
type Changeset struct {
	changes map[CellPosition]*CellChange
	order   []CellPosition // Insertion order for iteration
	mu      sync.RWMutex
}

// NewChangeset creates an empty changeset.
func NewChangeset() *Changeset {
	return &Changeset{
		changes: make(map[CellPosition]*CellChange),
	}
}

// RecordChange records a cell modification.
// If the cell was already changed, updates NewValue while keeping the original OldValue.
// If the new value equals the original value, the change is removed (reverted to original).
func (cs *Changeset) RecordChange(pos CellPosition, oldValue, newValue string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if existing, ok := cs.changes[pos]; ok {
		if existing.OldValue == newValue {
			// Reverted to original — remove from changeset
			delete(cs.changes, pos)
			cs.removeFromOrder(pos)
			return
		}
		existing.NewValue = newValue
		return
	}

	if oldValue == newValue {
		return
	}

	cs.changes[pos] = &CellChange{
		Position: pos,
		OldValue: oldValue,
		NewValue: newValue,
	}
	cs.order = append(cs.order, pos)
}

// RevertCell removes a single cell change and returns it.
// Returns nil if the cell has no pending change.
func (cs *Changeset) RevertCell(pos CellPosition) *CellChange {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	change, ok := cs.changes[pos]
	if !ok {
		return nil
	}

	delete(cs.changes, pos)
	cs.removeFromOrder(pos)
	return change
}

// RevertAll clears all changes and returns them in insertion order.
func (cs *Changeset) RevertAll() []*CellChange {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	result := make([]*CellChange, 0, len(cs.order))
	for _, pos := range cs.order {
		if change, ok := cs.changes[pos]; ok {
			result = append(result, change)
		}
	}

	cs.changes = make(map[CellPosition]*CellChange)
	cs.order = nil
	return result
}

// IsDirty returns whether the given cell has a pending change.
func (cs *Changeset) IsDirty(pos CellPosition) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	_, ok := cs.changes[pos]
	return ok
}

// GetChange returns the pending change for a cell, or nil.
func (cs *Changeset) GetChange(pos CellPosition) *CellChange {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.changes[pos]
}

// HasChanges returns whether any changes exist.
func (cs *Changeset) HasChanges() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.changes) > 0
}

// Count returns the number of pending changes.
func (cs *Changeset) Count() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.changes)
}

// Changes returns all changes in insertion order.
func (cs *Changeset) Changes() []*CellChange {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make([]*CellChange, 0, len(cs.order))
	for _, pos := range cs.order {
		if change, ok := cs.changes[pos]; ok {
			result = append(result, change)
		}
	}
	return result
}

// DirtyRows returns the set of row indices that have dirty cells.
func (cs *Changeset) DirtyRows() map[int]bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	rows := make(map[int]bool)
	for pos := range cs.changes {
		rows[pos.Row] = true
	}
	return rows
}

// DirtyCells returns all dirty cell positions.
func (cs *Changeset) DirtyCells() []CellPosition {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make([]CellPosition, 0, len(cs.changes))
	for pos := range cs.changes {
		result = append(result, pos)
	}
	return result
}

// Clear removes all changes without returning them.
func (cs *Changeset) Clear() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.changes = make(map[CellPosition]*CellChange)
	cs.order = nil
}

func (cs *Changeset) removeFromOrder(pos CellPosition) {
	for i, p := range cs.order {
		if p == pos {
			cs.order = append(cs.order[:i], cs.order[i+1:]...)
			return
		}
	}
}
