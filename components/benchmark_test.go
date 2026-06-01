package components

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/validators"
)

// ============================================================================
// Event System Benchmarks
// ============================================================================

// BenchmarkEventEmission measures event emission performance.
func BenchmarkEventEmission(b *testing.B) {
	emitter := &BaseEventEmitter{}

	// Register handlers
	emitter.OnEvent(func(e Event) {})
	emitter.OnEvent(func(e Event) {})
	emitter.OnEvent(func(e Event) {})

	event := NewActivateEvent("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emitter.EmitEvent(event)
	}
}

// BenchmarkEventEmission_ManyHandlers measures emission with many handlers.
func BenchmarkEventEmission_ManyHandlers(b *testing.B) {
	emitter := &BaseEventEmitter{}

	// Register 100 handlers
	for i := 0; i < 100; i++ {
		emitter.OnEvent(func(e Event) {})
	}

	event := NewActivateEvent("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emitter.EmitEvent(event)
	}
}

// BenchmarkEventCreation measures event creation performance.
func BenchmarkEventCreation(b *testing.B) {
	b.Run("ChangeEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewChangeEvent("field", "old", "new")
		}
	})

	b.Run("SubmitEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewSubmitEvent("form", "value")
		}
	})

	b.Run("ActivateEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewActivateEvent("button")
		}
	})
}

// BenchmarkTextField_ChangeEvent measures change event triggering.
func BenchmarkTextField_ChangeEvent(b *testing.B) {
	field := NewTextField("test").SetValue("")

	field.SetOnChange(func(e *ChangeEvent[string]) {
		// Simulate handler work
		_ = e.NewValue
	})

	event := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		field.HandleKey(event)
	}
}

// ============================================================================
// Validation Benchmarks
// ============================================================================

// BenchmarkValidation_Required measures required validator performance.
func BenchmarkValidation_Required(b *testing.B) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.Required()(v)
		}).
		SetValue("test value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = field.Validate()
	}
}

// BenchmarkValidation_Email measures email validator performance.
func BenchmarkValidation_Email(b *testing.B) {
	field := NewTextField("email").
		SetValidator(func(v string) error {
			return validators.Email()(v)
		}).
		SetValue("test@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = field.Validate()
	}
}

// BenchmarkValidation_Pattern measures pattern validator performance.
func BenchmarkValidation_Pattern(b *testing.B) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.Pattern(`^\d{3}-\d{4}$`)(v)
		}).
		SetValue("123-4567")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = field.Validate()
	}
}

// BenchmarkValidation_Composite measures composite validator performance.
func BenchmarkValidation_Composite(b *testing.B) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.All(
				validators.Required(),
				validators.MinLength(3),
				validators.MaxLength(100),
				validators.Alphanumeric(),
			)(v)
		}).
		SetValue("abc123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = field.Validate()
	}
}

// BenchmarkForm_ValidateAll measures form validation with multiple fields.
func BenchmarkForm_ValidateAll(b *testing.B) {
	form := NewFormBuilder().
		Text("name", "Name").
		Validate(validators.Required()).
		Value("John").
		Done().
		Text("email", "Email").
		Validate(validators.Required(), validators.Email()).
		Value("john@example.com").
		Done().
		Text("phone", "Phone").
		Validate(validators.Required(), validators.Pattern(`^\d{3}-\d{4}$`)).
		Value("123-4567").
		Done().
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = form.ValidateAll()
	}
}

// ============================================================================
// Rendering Benchmarks
// ============================================================================

// BenchmarkTextField_Render measures TextField rendering performance.
func BenchmarkTextField_Render(b *testing.B) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(80, 24)

	field := NewTextField("test").SetValue("benchmark text field value")
	field.SetRect(0, 0, 80, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.Clear()
		field.Draw(screen)
		screen.Show()
	}
}

// BenchmarkPanel_Render measures Panel rendering performance.
func BenchmarkPanel_Render(b *testing.B) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(80, 24)

	panel := NewPanel().SetTitle("Benchmark Panel")
	panel.SetBorder(true)
	panel.SetRect(0, 0, 80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.Clear()
		panel.Draw(screen)
		screen.Show()
	}
}

// BenchmarkForm_Render measures Form rendering performance.
func BenchmarkForm_Render(b *testing.B) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(80, 24)

	form := NewFormBuilder().
		Text("name", "Name").Value("John Doe").Done().
		Text("email", "Email").Value("john@example.com").Done().
		Checkbox("subscribe", "Subscribe").Checked(true).Done().
		Select("role", "Role", []string{"Admin", "User", "Guest"}).Default("User").Done().
		Build()

	form.SetRect(0, 0, 80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.Clear()
		form.Draw(screen)
		screen.Show()
	}
}

// BenchmarkList_Render measures List rendering with many items.
func BenchmarkList_Render(b *testing.B) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(80, 24)

	items := make([]ListItem, 100)
	for i := range items {
		items[i] = ListItem{
			Text:      "Item " + string(rune('A'+i%26)),
			Secondary: "Description for item",
		}
	}

	list := NewList()
	list.SetItems(items)
	list.SetRect(0, 0, 80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.Clear()
		list.Draw(screen)
		screen.Show()
	}
}

// BenchmarkModal_Render measures Modal rendering performance.
func BenchmarkModal_Render(b *testing.B) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(80, 24)

	modal := NewModal(ModalConfig{
		Title:  "Benchmark Modal",
		Width:  40,
		Height: 10,
	})

	content := core.NewTextView()
	content.SetText("Modal content for benchmarking")
	modal.SetContent(content)
	modal.SetRect(0, 0, 80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.Clear()
		modal.Draw(screen)
		screen.Show()
	}
}

// ============================================================================
// Component Operation Benchmarks
// ============================================================================

// BenchmarkCheckbox_Toggle measures checkbox toggle performance.
func BenchmarkCheckbox_Toggle(b *testing.B) {
	cb := NewCheckbox("test").SetChecked(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Toggle()
	}
}

// BenchmarkSelect_SetSelected measures selection change performance.
func BenchmarkSelect_SetSelected(b *testing.B) {
	sel := NewSelect("test").SetOptions([]string{"A", "B", "C", "D", "E"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sel.SetSelected(i % 5)
	}
}

// BenchmarkTextField_SetValue measures value setting performance.
func BenchmarkTextField_SetValue(b *testing.B) {
	field := NewTextField("test")
	values := []string{"short", "medium length value", "this is a longer value for testing"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		field.SetValue(values[i%3])
	}
}

// BenchmarkForm_GetValues measures form value extraction performance.
func BenchmarkForm_GetValues(b *testing.B) {
	form := NewFormBuilder().
		Text("name", "Name").Value("John").Done().
		Text("email", "Email").Value("john@example.com").Done().
		Text("phone", "Phone").Value("123-4567").Done().
		Checkbox("active", "Active").Checked(true).Done().
		Select("status", "Status", []string{"pending", "active", "done"}).Default("active").Done().
		Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = form.GetValues()
	}
}

// BenchmarkForm_SetValues measures bulk value setting performance.
func BenchmarkForm_SetValues(b *testing.B) {
	form := NewFormBuilder().
		Text("name", "Name").Done().
		Text("email", "Email").Done().
		Checkbox("notify", "Notify").Done().
		Select("role", "Role", []string{"admin", "user"}).Done().
		Build()

	values := map[string]any{
		"name":   "Alice",
		"email":  "alice@example.com",
		"notify": true,
		"role":   "admin",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = form.SetValues(values)
	}
}

// ============================================================================
// Lifecycle Benchmarks
// ============================================================================

// BenchmarkComponentBase_Lifecycle measures Start/Stop cycle performance.
func BenchmarkComponentBase_Lifecycle(b *testing.B) {
	base := NewComponentBase(new(core.Box)).
		SetOnStart(func() {}).
		SetOnStop(func() {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base.Start()
		base.Stop()
	}
}

// BenchmarkStatefulComponentBase_SetData measures data setting performance.
func BenchmarkStatefulComponentBase_SetData(b *testing.B) {
	base := NewStatefulComponentBase[string](new(core.Box))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base.SetData("test data")
	}
}

// BenchmarkStatefulComponentBase_UpdateData measures update cycle performance.
func BenchmarkStatefulComponentBase_UpdateData(b *testing.B) {
	base := NewStatefulComponentBase[int](new(core.Box))
	base.SetData(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		base.UpdateData(func(current int) int {
			return current + 1
		})
	}
}

// ============================================================================
// Input Handling Benchmarks
// ============================================================================

// BenchmarkTextField_HandleKey measures input handling performance.
func BenchmarkTextField_HandleKey(b *testing.B) {
	field := NewTextField("test")
	event := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		field.HandleKey(event)
	}
}

// BenchmarkTextField_HandleKey_WithValidation measures input with validation.
func BenchmarkTextField_HandleKey_WithValidation(b *testing.B) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.MaxLength(100)(v)
		})

	event := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		field.HandleKey(event)
	}
}

// BenchmarkTextField_TypeString measures typing a string character by character.
func BenchmarkTextField_TypeString(b *testing.B) {
	text := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		field := NewTextField("test")
		for _, r := range text {
			event := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
			field.HandleKey(event)
		}
	}
}

// ============================================================================
// Memory Allocation Benchmarks
// ============================================================================

// BenchmarkTextField_Allocs measures TextField allocations.
func BenchmarkTextField_Allocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewTextField("test").
			SetValue("test value").
			SetPlaceholder("Enter value")
	}
}

// BenchmarkForm_Allocs measures Form allocations.
func BenchmarkForm_Allocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewFormBuilder().
			Text("name", "Name").Value("John").Done().
			Text("email", "Email").Value("john@example.com").Done().
			Checkbox("subscribe", "Subscribe").Done().
			Build()
	}
}

// BenchmarkEvent_Allocs measures event allocation.
func BenchmarkEvent_Allocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewChangeEvent("field", "old", "new")
	}
}

// ============================================================================
// String Benchmarks (for internal use)
// ============================================================================

// BenchmarkStringBuilder measures strings.Builder performance for dumps.
func BenchmarkStringBuilder(b *testing.B) {
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "This is line content for benchmarking"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		for _, line := range lines {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		_ = sb.String()
	}
}
