package components_test

import (
	"fmt"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/validators"
)

// This file contains testable examples that appear in godoc.
// Run with: go test -v -run Example

func ExampleNewTextField() {
	// Create a text field with label and placeholder
	field := components.NewTextField("email").
		SetLabel("Email Address").
		SetPlaceholder("user@example.com")

	// Set a value
	field.SetValue("test@example.com")

	// Get the value
	fmt.Println(field.Value())
	// Output: test@example.com
}

func ExampleTextField_SetValidator() {
	// Create a field with validation
	field := components.NewTextField("email").
		SetLabel("Email").
		SetValidator(func(value string) error {
			return validators.Email()(value)
		})

	// Validate the field
	field.SetValue("invalid-email")
	if err := field.Validate(); err != nil {
		fmt.Println("Validation failed:", err)
	}

	field.SetValue("valid@example.com")
	if err := field.Validate(); err == nil {
		fmt.Println("Validation passed")
	}
	// Output:
	// Validation failed: invalid email format
	// Validation passed
}

// ExampleSelect is a kitchen-sink walkthrough of the Select API: distinct
// display labels and stored values, a default selection, a change handler,
// reading the current selection, and selecting programmatically by value.
func ExampleSelect() {
	sel := components.NewSelect("region").
		SetLabel("Region").
		SetPlaceholder("Choose a region").
		SetOptionsWithValues([]components.SelectOption{
			{Label: "North America", Value: "na"},
			{Label: "Europe", Value: "eu"},
			{Label: "Asia Pacific", Value: "apac"},
		}).
		SetDefault("eu")

	// React to selection changes. The handler receives the typed option so
	// you get both the human-readable label and the stored value.
	sel.SetOnChange(func(e *components.ChangeEvent[components.SelectOption]) {
		fmt.Printf("changed to %s (%s)\n", e.NewValue.Label, e.NewValue.Value)
	})

	// Read the current selection.
	fmt.Println("value:", sel.Value())
	fmt.Println("label:", sel.SelectedOption().Label)

	// Select programmatically by value (does not fire onChange).
	_ = sel.SetSelectedValue("apac")
	fmt.Println("value:", sel.Value())

	// Output:
	// value: eu
	// label: Europe
	// value: apac
}

func ExampleNewPanel() {
	// Create a panel container
	panel := components.NewPanel().
		SetTitle("Settings").
		SetTitleAlign(components.AlignLeft)

	fmt.Println(panel != nil)
	// Output: true
}

func ExampleNewTable() {
	// Create a table with headers
	table := components.NewTable()
	table.SetHeaders("Name", "Status", "Created")

	// Add rows
	table.AddRow("Alice", "Active", "2024-01-15")
	table.AddRow("Bob", "Inactive", "2024-01-10")

	fmt.Println(table.RowCount())
	// Output: 2
}

func ExampleTable_AddRowWithStatus() {
	// Create a table with status-colored rows
	table := components.NewTable()
	table.SetHeaders("Name", "Status", "Created")

	// Note: In actual usage, theme.StatusSuccess() returns a status
	// that provides color and icon automatically
	table.AddRow("Alice", "Active", "2024-01-15")

	fmt.Println(table.RowCount())
	// Output: 1
}

func ExampleNewFormBuilder() {
	// Build a form with the FormBuilder API
	form := components.NewFormBuilder().
		Text("name", "Name").
		Placeholder("Enter your name").
		Validate(validators.Required()).
		Done().
		Text("email", "Email").
		Validate(validators.Required(), validators.Email()).
		Done().
		Checkbox("notify", "Email notifications").
		Checked(true).
		Done().
		OnSubmit(func(values map[string]any) {
			name := values["name"].(string)
			email := values["email"].(string)
			notify := values["notify"].(bool)
			fmt.Printf("Name: %s, Email: %s, Notify: %v\n", name, email, notify)
		}).
		Build()

	fmt.Println(form != nil)
	// Output: true
}

func ExampleNewModal() {
	// Create a modal dialog
	modal := components.NewModal(components.ModalConfig{
		Title:    "Confirm",
		Width:    50,
		Height:   10,
		Backdrop: true,
	})

	modal.SetOnSubmit(func() {
		fmt.Println("Submitted")
	})

	modal.SetOnCancel(func() {
		fmt.Println("Cancelled")
	})

	fmt.Println(modal != nil)
	// Output: true
}

func ExampleNewComponentBase() {
	// ComponentBase wraps a primitive to implement nav.Component
	panel := components.NewPanel().SetTitle("My View")

	base := components.NewComponentBase(panel).
		SetName("my-view").
		AddHint("Enter", "Select").
		AddHint("q", "Quit").
		SetOnStart(func() {
			fmt.Println("View started")
		}).
		SetOnStop(func() {
			fmt.Println("View stopped")
		})

	// ComponentBase now implements nav.Component:
	// - Start() calls the onStart callback
	// - Stop() calls the onStop callback
	// - Hints() returns the configured hints

	fmt.Println(base.Name())
	// Output: my-view
}

func ExampleKeyHint() {
	// KeyHint represents a keyboard shortcut
	hints := []components.KeyHint{
		{Key: "Enter", Description: "Select"},
		{Key: "Esc", Description: "Back"},
		{Key: "q", Description: "Quit"},
	}

	for _, h := range hints {
		fmt.Printf("%s: %s\n", h.Key, h.Description)
	}
	// Output:
	// Enter: Select
	// Esc: Back
	// q: Quit
}
