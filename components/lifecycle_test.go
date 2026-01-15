package components

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComponentBase_Lifecycle tests ComponentBase Start/Stop lifecycle.
func TestComponentBase_Lifecycle(t *testing.T) {
	panel := NewPanel()
	var startCalled, stopCalled bool

	base := NewComponentBase(panel).
		SetName("test-view").
		SetOnStart(func() { startCalled = true }).
		SetOnStop(func() { stopCalled = true })

	// Initial state
	assert.False(t, startCalled)
	assert.False(t, stopCalled)

	// Start lifecycle
	base.Start()
	assert.True(t, startCalled)
	assert.False(t, stopCalled)

	// Stop lifecycle
	base.Stop()
	assert.True(t, stopCalled)
}

// TestComponentBase_LifecycleOrder tests that lifecycle methods are called in order.
func TestComponentBase_LifecycleOrder(t *testing.T) {
	var order []string

	base := NewComponentBase(tview.NewBox()).
		SetOnStart(func() { order = append(order, "start") }).
		SetOnStop(func() { order = append(order, "stop") })

	base.Start()
	base.Stop()
	base.Start()
	base.Stop()

	assert.Equal(t, []string{"start", "stop", "start", "stop"}, order)
}

// TestComponentBase_NilCallbacks tests that nil callbacks don't panic.
func TestComponentBase_NilCallbacks(t *testing.T) {
	base := NewComponentBase(tview.NewBox())

	// Should not panic
	base.Start()
	base.Stop()
}

// TestComponentBase_Name tests component name accessor.
func TestComponentBase_Name(t *testing.T) {
	base := NewComponentBase(tview.NewBox()).SetName("my-component")
	assert.Equal(t, "my-component", base.Name())
}

// TestComponentBase_ID tests component ID uniqueness.
func TestComponentBase_ID(t *testing.T) {
	ids := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		base := NewComponentBase(tview.NewBox())
		if ids[base.ID()] {
			t.Fatalf("duplicate component ID: %d", base.ID())
		}
		ids[base.ID()] = true
	}
}

// TestComponentBase_Hints tests key hint management.
func TestComponentBase_Hints(t *testing.T) {
	base := NewComponentBase(tview.NewBox()).
		AddHint("Enter", "Select").
		AddHint("q", "Quit")

	hints := base.Hints()
	require.Len(t, hints, 2)
	assert.Equal(t, "Enter", hints[0].Key)
	assert.Equal(t, "Select", hints[0].Description)
	assert.Equal(t, "q", hints[1].Key)
	assert.Equal(t, "Quit", hints[1].Description)
}

// TestComponentBase_HintsCopy tests that Hints returns a copy.
func TestComponentBase_HintsCopy(t *testing.T) {
	base := NewComponentBase(tview.NewBox()).
		AddHint("Enter", "Select")

	hints := base.Hints()
	hints[0].Key = "modified"

	// Original should be unchanged
	assert.Equal(t, "Enter", base.Hints()[0].Key)
}

// TestComponentBase_SetHints tests bulk hint setting.
func TestComponentBase_SetHints(t *testing.T) {
	base := NewComponentBase(tview.NewBox()).
		SetHints([]KeyHint{
			{Key: "a", Description: "Action A"},
			{Key: "b", Description: "Action B"},
		})

	hints := base.Hints()
	require.Len(t, hints, 2)
	assert.Equal(t, "a", hints[0].Key)
	assert.Equal(t, "b", hints[1].Key)
}

// TestComponentBase_Primitive tests underlying primitive access.
func TestComponentBase_Primitive(t *testing.T) {
	panel := NewPanel()
	base := NewComponentBase(panel)

	assert.Same(t, panel, base.Primitive())
}

// TestComponentBase_ThreadSafety tests concurrent access to ComponentBase.
func TestComponentBase_ThreadSafety(t *testing.T) {
	base := NewComponentBase(tview.NewBox())

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent name access
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			base.SetName("name")
			_ = base.Name()
		}(i)
	}

	// Concurrent hint access
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			base.AddHint("key", "desc")
			_ = base.Hints()
		}()
	}

	// Concurrent lifecycle
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			base.Start()
			base.Stop()
		}()
	}

	wg.Wait()
}

// TestStatefulComponentBase_LoadStates tests state transitions.
func TestStatefulComponentBase_LoadStates(t *testing.T) {
	tests := []struct {
		name        string
		transition  func(*StatefulComponentBase[string])
		wantState   LoadState
		wantReady   bool
		wantLoading bool
		wantError   bool
	}{
		{
			name:       "initial idle",
			transition: func(s *StatefulComponentBase[string]) {},
			wantState:  LoadStateIdle,
		},
		{
			name: "set loading",
			transition: func(s *StatefulComponentBase[string]) {
				s.SetLoadState(LoadStateLoading)
			},
			wantState:   LoadStateLoading,
			wantLoading: true,
		},
		{
			name: "set data transitions to success",
			transition: func(s *StatefulComponentBase[string]) {
				s.SetData("test data")
			},
			wantState: LoadStateSuccess,
			wantReady: true,
		},
		{
			name: "set error",
			transition: func(s *StatefulComponentBase[string]) {
				s.SetError(errors.New("test error"))
			},
			wantState: LoadStateError,
			wantError: true,
		},
		{
			name: "reset clears all",
			transition: func(s *StatefulComponentBase[string]) {
				s.SetData("data")
				s.Reset()
			},
			wantState: LoadStateIdle,
		},
		{
			name: "loading to success",
			transition: func(s *StatefulComponentBase[string]) {
				s.SetLoadState(LoadStateLoading)
				s.SetData("loaded")
			},
			wantState: LoadStateSuccess,
			wantReady: true,
		},
		{
			name: "loading to error",
			transition: func(s *StatefulComponentBase[string]) {
				s.SetLoadState(LoadStateLoading)
				s.SetError(errors.New("failed"))
			},
			wantState: LoadStateError,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewStatefulComponentBase[string](tview.NewBox())
			tt.transition(base)

			assert.Equal(t, tt.wantState, base.LoadState())
			assert.Equal(t, tt.wantReady, base.IsReady())
			assert.Equal(t, tt.wantLoading, base.IsLoading())
			assert.Equal(t, tt.wantError, base.HasError())
		})
	}
}

// TestStatefulComponentBase_Data tests data storage and retrieval.
func TestStatefulComponentBase_Data(t *testing.T) {
	base := NewStatefulComponentBase[[]int](tview.NewBox())

	// Initial data should be zero value
	assert.Nil(t, base.Data())

	// Set data
	data := []int{1, 2, 3}
	base.SetData(data)
	assert.Equal(t, data, base.Data())

	// Update data
	base.SetData([]int{4, 5})
	assert.Equal(t, []int{4, 5}, base.Data())
}

// TestStatefulComponentBase_Error tests error handling.
func TestStatefulComponentBase_Error(t *testing.T) {
	base := NewStatefulComponentBase[string](tview.NewBox())

	// Initially no error
	assert.Nil(t, base.Error())
	assert.False(t, base.HasError())

	// Set error
	err := errors.New("something went wrong")
	base.SetError(err)
	assert.Equal(t, err, base.Error())
	assert.True(t, base.HasError())
	assert.Equal(t, LoadStateError, base.LoadState())

	// Reset clears error
	base.Reset()
	assert.Nil(t, base.Error())
	assert.False(t, base.HasError())
}

// TestStatefulComponentBase_OnStateChange tests state change callback.
func TestStatefulComponentBase_OnStateChange(t *testing.T) {
	base := NewStatefulComponentBase[string](tview.NewBox())

	var stateChanges []LoadState
	var dataValues []string

	base.SetOnStateChange(func(state LoadState, data string, err error) {
		stateChanges = append(stateChanges, state)
		dataValues = append(dataValues, data)
	})

	// Loading
	base.SetLoadState(LoadStateLoading)
	assert.Equal(t, []LoadState{LoadStateLoading}, stateChanges)

	// Success with data
	base.SetData("hello")
	assert.Equal(t, []LoadState{LoadStateLoading, LoadStateSuccess}, stateChanges)
	assert.Equal(t, []string{"", "hello"}, dataValues)

	// Error
	base.SetError(errors.New("fail"))
	assert.Equal(t, []LoadState{LoadStateLoading, LoadStateSuccess, LoadStateError}, stateChanges)
}

// TestStatefulComponentBase_OnDataChange tests data change callback.
func TestStatefulComponentBase_OnDataChange(t *testing.T) {
	base := NewStatefulComponentBase[int](tview.NewBox())

	var dataChanges []int
	base.SetOnDataChange(func(data int) {
		dataChanges = append(dataChanges, data)
	})

	base.SetData(10)
	base.SetData(20)
	base.SetData(30)

	assert.Equal(t, []int{10, 20, 30}, dataChanges)
}

// TestStatefulComponentBase_UpdateData tests atomic data updates.
func TestStatefulComponentBase_UpdateData(t *testing.T) {
	base := NewStatefulComponentBase[int](tview.NewBox())
	base.SetData(10)

	base.UpdateData(func(current int) int {
		return current * 2
	})

	assert.Equal(t, 20, base.Data())
}

// TestStatefulComponentBase_UpdateDataCallback tests that UpdateData triggers callbacks.
func TestStatefulComponentBase_UpdateDataCallback(t *testing.T) {
	base := NewStatefulComponentBase[int](tview.NewBox())
	base.SetData(5) // Sets to Success state

	var dataChanges []int
	base.SetOnDataChange(func(data int) {
		dataChanges = append(dataChanges, data)
	})

	base.UpdateData(func(current int) int {
		return current + 1
	})

	require.Len(t, dataChanges, 1)
	assert.Equal(t, 6, dataChanges[0])
}

// TestStatefulComponentBase_ThreadSafety tests concurrent access.
func TestStatefulComponentBase_ThreadSafety(t *testing.T) {
	base := NewStatefulComponentBase[int](tview.NewBox())

	var ops int64
	done := make(chan struct{})

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					_ = base.Data()
					_ = base.LoadState()
					_ = base.IsReady()
					_ = base.IsLoading()
					_ = base.HasError()
					_ = base.Error()
					atomic.AddInt64(&ops, 1)
				}
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func(n int) {
			for {
				select {
				case <-done:
					return
				default:
					base.SetData(n)
					base.SetLoadState(LoadStateLoading)
					atomic.AddInt64(&ops, 1)
				}
			}
		}(i)
	}

	time.Sleep(100 * time.Millisecond)
	close(done)

	// If we get here without data race, test passes
	t.Logf("completed %d concurrent operations", atomic.LoadInt64(&ops))
}

// TestStatefulComponentBase_FluentAPI tests fluent method chaining.
func TestStatefulComponentBase_FluentAPI(t *testing.T) {
	base := NewStatefulComponentBase[string](tview.NewBox()).
		SetName("stateful-view").
		AddHint("r", "Refresh").
		SetOnStart(func() {}).
		SetOnStop(func() {}).
		SetOnStateChange(func(state LoadState, data string, err error) {}).
		SetOnDataChange(func(data string) {})

	assert.Equal(t, "stateful-view", base.Name())
	assert.Len(t, base.Hints(), 1)
}

// TestStatefulComponentBase_InheritsComponentBase tests that StatefulComponentBase
// properly inherits from ComponentBase.
func TestStatefulComponentBase_InheritsComponentBase(t *testing.T) {
	var startCalled, stopCalled bool

	base := NewStatefulComponentBase[string](tview.NewBox()).
		SetName("inherited").
		SetOnStart(func() { startCalled = true }).
		SetOnStop(func() { stopCalled = true })

	// Test ComponentBase methods work
	assert.Equal(t, "inherited", base.Name())
	assert.NotZero(t, base.ID())

	base.Start()
	assert.True(t, startCalled)

	base.Stop()
	assert.True(t, stopCalled)
}

// TestStatefulComponentBase_SetLoadStateClearsError tests that non-error states clear error.
func TestStatefulComponentBase_SetLoadStateClearsError(t *testing.T) {
	base := NewStatefulComponentBase[string](tview.NewBox())

	// Set error
	base.SetError(errors.New("error"))
	require.NotNil(t, base.Error())

	// Set to loading - should NOT clear error since SetLoadState preserves error for LoadStateError
	base.SetLoadState(LoadStateLoading)
	assert.Nil(t, base.Error())
}

// TestStatefulComponentBase_Reset tests full reset behavior.
func TestStatefulComponentBase_Reset(t *testing.T) {
	var stateChanges []LoadState

	base := NewStatefulComponentBase[string](tview.NewBox()).
		SetOnStateChange(func(state LoadState, data string, err error) {
			stateChanges = append(stateChanges, state)
		})

	// Set data and error
	base.SetData("data")
	base.SetError(errors.New("error"))

	// Reset
	base.Reset()

	assert.Equal(t, LoadStateIdle, base.LoadState())
	assert.Equal(t, "", base.Data())
	assert.Nil(t, base.Error())
	assert.Contains(t, stateChanges, LoadStateIdle)
}

// TestKeyHint tests KeyHint struct.
func TestKeyHint(t *testing.T) {
	hints := []KeyHint{
		{Key: "Enter", Description: "Select"},
		{Key: "Esc", Description: "Back"},
		{Key: "q", Description: "Quit"},
	}

	assert.Equal(t, "Enter", hints[0].Key)
	assert.Equal(t, "Select", hints[0].Description)
	assert.Len(t, hints, 3)
}
