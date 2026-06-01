package core

import "sync"

// FocusManager tracks focus state and a push/pop history for modal restoration.
// One instance lives on App; nothing else should own focus state.
type FocusManager struct {
	mu       sync.Mutex
	current  Widget
	history  []Widget
	onChange []onChangeFn
	seq      uint64
}

type onChangeFn struct {
	id uint64
	fn func(prev, next Widget)
}

// NewFocusManager returns an empty FocusManager with no focused widget.
func NewFocusManager() *FocusManager { return &FocusManager{} }

// Focus sets focus to w: calls Blur on previous widget, Focus on next, fires callbacks.
// Passing nil clears focus.
func (f *FocusManager) Focus(w Widget) {
	f.mu.Lock()
	prev := f.current
	f.current = w
	cbs := make([]onChangeFn, len(f.onChange))
	copy(cbs, f.onChange)
	f.mu.Unlock()

	if prev != nil && prev != w {
		prev.Blur()
	}
	if w != nil {
		w.Focus()
	}
	for _, cb := range cbs {
		cb.fn(prev, w)
	}
}

// Focused returns the current focused widget (nil if none).
func (f *FocusManager) Focused() Widget {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.current
}

// Push saves the current focus to history then focuses w.
// Used by nav.Pages when opening a modal.
func (f *FocusManager) Push(w Widget) {
	f.mu.Lock()
	f.history = append(f.history, f.current)
	f.mu.Unlock()
	f.Focus(w)
}

// Pop restores the last saved focus. Returns the restored widget, or nil if history empty.
func (f *FocusManager) Pop() Widget {
	f.mu.Lock()
	if len(f.history) == 0 {
		f.mu.Unlock()
		return nil
	}
	last := f.history[len(f.history)-1]
	f.history = f.history[:len(f.history)-1]
	f.mu.Unlock()
	f.Focus(last)
	return last
}

// OnChange registers a callback fired when focus changes (prev and next Widget).
// Returns an unregister function.
func (f *FocusManager) OnChange(fn func(prev, next Widget)) func() {
	f.mu.Lock()
	f.seq++
	id := f.seq
	f.onChange = append(f.onChange, onChangeFn{id, fn})
	f.mu.Unlock()
	return func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		for i, cb := range f.onChange {
			if cb.id == id {
				f.onChange = append(f.onChange[:i], f.onChange[i+1:]...)
				return
			}
		}
	}
}
