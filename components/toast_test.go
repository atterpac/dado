package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToast_Icon tests toast icon for each level.
func TestToast_Icon(t *testing.T) {
	tests := []struct {
		level    ToastLevel
		expected string
	}{
		{ToastInfo, "ℹ"},
		{ToastSuccess, "✓"},
		{ToastWarning, "⚠"},
		{ToastError, "✗"},
	}

	for _, tt := range tests {
		toast := &Toast{Level: tt.level}
		assert.Equal(t, tt.expected, toast.Icon())
	}
}

// TestToastManager_NewToastManager tests ToastManager creation.
func TestToastManager_NewToastManager(t *testing.T) {
	manager := NewToastManager()

	assert.NotNil(t, manager)
	assert.False(t, manager.HasActive())
	assert.Empty(t, manager.GetActive())
}

// TestToastManager_SetPosition tests position setting.
func TestToastManager_SetPosition(t *testing.T) {
	manager := NewToastManager()

	positions := []ToastPosition{
		ToastTopRight,
		ToastTopLeft,
		ToastBottomRight,
		ToastBottomLeft,
		ToastTopCenter,
		ToastBottomCenter,
	}

	for _, pos := range positions {
		result := manager.SetPosition(pos)
		assert.Same(t, manager, result)
	}
}

// TestToastManager_SetMaxVisible tests max visible setting.
func TestToastManager_SetMaxVisible(t *testing.T) {
	manager := NewToastManager()

	result := manager.SetMaxVisible(3)

	assert.Same(t, manager, result)
}

// TestToastManager_SetMaxWidth tests max width setting.
func TestToastManager_SetMaxWidth(t *testing.T) {
	manager := NewToastManager()

	result := manager.SetMaxWidth(50)

	assert.Same(t, manager, result)
}

// TestToastManager_SetDefaultDuration tests default duration setting.
func TestToastManager_SetDefaultDuration(t *testing.T) {
	manager := NewToastManager()

	result := manager.SetDefaultDuration(5 * time.Second)

	assert.Same(t, manager, result)
}

// TestToastManager_Show tests showing a basic toast.
func TestToastManager_Show(t *testing.T) {
	manager := NewToastManager()

	toast := manager.Show("Test message", ToastInfo)

	require.NotNil(t, toast)
	assert.Equal(t, "Test message", toast.Message)
	assert.Equal(t, ToastInfo, toast.Level)
	assert.NotEmpty(t, toast.ID)
	assert.True(t, manager.HasActive())
	assert.Len(t, manager.GetActive(), 1)
}

// TestToastManager_ShowWithDuration tests showing with custom duration.
func TestToastManager_ShowWithDuration(t *testing.T) {
	manager := NewToastManager()

	toast := manager.ShowWithDuration("Test", ToastWarning, 10*time.Second)

	require.NotNil(t, toast)
	assert.Equal(t, 10*time.Second, toast.Duration)
	assert.Equal(t, ToastWarning, toast.Level)
}

// TestToastManager_ShowPersistent tests persistent toasts.
func TestToastManager_ShowPersistent(t *testing.T) {
	manager := NewToastManager()

	toast := manager.ShowPersistent("Important!", ToastError)

	require.NotNil(t, toast)
	assert.Equal(t, time.Duration(0), toast.Duration)
}

// TestToastManager_ShowWithAction tests toasts with actions.
func TestToastManager_ShowWithAction(t *testing.T) {
	manager := NewToastManager()

	var actionCalled bool
	toast := manager.ShowWithAction("Action toast", ToastSuccess,
		ToastAction{Label: "Confirm", Handler: func() { actionCalled = true }},
		ToastAction{Label: "Cancel", Handler: func() {}},
	)

	require.NotNil(t, toast)
	assert.Len(t, toast.Actions, 2)
	assert.Equal(t, "Confirm", toast.Actions[0].Label)
	assert.Equal(t, time.Duration(0), toast.Duration) // Persistent by default

	// Trigger action
	toast.Actions[0].Handler()
	assert.True(t, actionCalled)
}

// TestToastManager_ShowWithUndo tests undo toast helper.
func TestToastManager_ShowWithUndo(t *testing.T) {
	manager := NewToastManager()

	var undoCalled bool
	toast := manager.ShowWithUndo("Item deleted", func() { undoCalled = true })

	require.NotNil(t, toast)
	assert.Len(t, toast.Actions, 2)
	assert.Equal(t, "Undo", toast.Actions[0].Label)
	assert.Equal(t, "Dismiss", toast.Actions[1].Label)

	toast.Actions[0].Handler()
	assert.True(t, undoCalled)
}

// TestToastManager_ConvenienceMethods tests Info/Success/Warning/Error methods.
func TestToastManager_ConvenienceMethods(t *testing.T) {
	manager := NewToastManager()

	t.Run("Info", func(t *testing.T) {
		toast := manager.Info("Info message")
		assert.Equal(t, ToastInfo, toast.Level)
	})

	t.Run("Success", func(t *testing.T) {
		toast := manager.Success("Success message")
		assert.Equal(t, ToastSuccess, toast.Level)
	})

	t.Run("Warning", func(t *testing.T) {
		toast := manager.Warning("Warning message")
		assert.Equal(t, ToastWarning, toast.Level)
	})

	t.Run("Error", func(t *testing.T) {
		toast := manager.Error("Error message")
		assert.Equal(t, ToastError, toast.Level)
	})
}

// TestToastManager_Dismiss tests dismissing a toast.
func TestToastManager_Dismiss(t *testing.T) {
	manager := NewToastManager()

	toast := manager.Show("Test", ToastInfo)
	require.True(t, manager.HasActive())

	manager.Dismiss(toast.ID)

	assert.False(t, manager.HasActive())
	assert.Empty(t, manager.GetActive())
}

// TestToastManager_DismissAll tests dismissing all toasts.
func TestToastManager_DismissAll(t *testing.T) {
	manager := NewToastManager()

	manager.Show("Toast 1", ToastInfo)
	manager.Show("Toast 2", ToastWarning)
	manager.Show("Toast 3", ToastError)
	require.Len(t, manager.GetActive(), 3)

	manager.DismissAll()

	assert.False(t, manager.HasActive())
	assert.Empty(t, manager.GetActive())
}

// TestToastManager_GetActive tests getting active toasts.
func TestToastManager_GetActive(t *testing.T) {
	manager := NewToastManager()

	manager.Show("Toast 1", ToastInfo)
	manager.Show("Toast 2", ToastWarning)

	active := manager.GetActive()

	assert.Len(t, active, 2)
	assert.Equal(t, "Toast 1", active[0].Message)
	assert.Equal(t, "Toast 2", active[1].Message)
}

// TestToastManager_SetOnShow tests show callback.
func TestToastManager_SetOnShow(t *testing.T) {
	manager := NewToastManager()

	var shownToast *Toast
	result := manager.SetOnShow(func(toast *Toast) {
		shownToast = toast
	})

	assert.Same(t, manager, result)

	toast := manager.Show("Test", ToastInfo)

	assert.Same(t, toast, shownToast)
}

// TestToastManager_SetOnDismiss tests dismiss callback.
func TestToastManager_SetOnDismiss(t *testing.T) {
	manager := NewToastManager()

	var dismissedToast *Toast
	manager.SetOnDismiss(func(toast *Toast) {
		dismissedToast = toast
	})

	toast := manager.Show("Test", ToastInfo)
	manager.Dismiss(toast.ID)

	assert.Equal(t, toast.ID, dismissedToast.ID)
}

// TestToastManager_HandleAction tests action handling.
func TestToastManager_HandleAction(t *testing.T) {
	t.Run("valid action", func(t *testing.T) {
		manager := NewToastManager()

		var actionCalled bool
		manager.ShowWithAction("Test", ToastInfo,
			ToastAction{Label: "Action", Handler: func() { actionCalled = true }},
		)

		result := manager.HandleAction(0)

		assert.True(t, result)
		assert.True(t, actionCalled)
		assert.False(t, manager.HasActive()) // Toast should be dismissed
	})

	t.Run("invalid action index", func(t *testing.T) {
		manager := NewToastManager()
		manager.ShowWithAction("Test", ToastInfo,
			ToastAction{Label: "Action", Handler: func() {}},
		)

		result := manager.HandleAction(5)

		assert.False(t, result)
		assert.True(t, manager.HasActive()) // Toast should remain
	})

	t.Run("no toasts", func(t *testing.T) {
		manager := NewToastManager()

		result := manager.HandleAction(0)

		assert.False(t, result)
	})
}

// TestToastManager_FluentAPI tests method chaining.
func TestToastManager_FluentAPI(t *testing.T) {
	var showCalled, dismissCalled bool

	manager := NewToastManager().
		SetPosition(ToastBottomRight).
		SetMaxVisible(3).
		SetMaxWidth(50).
		SetDefaultDuration(5 * time.Second).
		SetOnShow(func(toast *Toast) { showCalled = true }).
		SetOnDismiss(func(toast *Toast) { dismissCalled = true })

	toast := manager.Show("Test", ToastInfo)
	manager.Dismiss(toast.ID)

	assert.True(t, showCalled)
	assert.True(t, dismissCalled)
}

// TestToastManager_UniqueIDs tests that toast IDs are unique.
func TestToastManager_UniqueIDs(t *testing.T) {
	manager := NewToastManager()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		toast := manager.Show("Test", ToastInfo)
		assert.False(t, ids[toast.ID], "Duplicate ID found: %s", toast.ID)
		ids[toast.ID] = true
	}
}

// TestToastManager_ThreadSafety tests concurrent access.
func TestToastManager_ThreadSafety(t *testing.T) {
	manager := NewToastManager()

	done := make(chan bool)

	// Spawn multiple goroutines showing toasts
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				toast := manager.Show("Test", ToastInfo)
				manager.GetActive()
				manager.HasActive()
				manager.Dismiss(toast.ID)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and state should be consistent
	assert.NotNil(t, manager.GetActive())
}

// TestToast_DefaultIcon tests unknown level icon.
func TestToast_DefaultIcon(t *testing.T) {
	toast := &Toast{Level: ToastLevel(99)} // Unknown level
	assert.Equal(t, "", toast.Icon())
}
