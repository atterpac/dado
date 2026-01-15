package advanced

import (
	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
	"github.com/atterpac/jig/validators"
)

func init() {
	demos.Register(&FormDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Form",
			DemoDescription: "Form with validation",
			DemoCategory:    demos.Advanced,
			DemoCode:        formCode,
		},
	})
}

// FormDemo demonstrates the Form and FormBuilder components.
type FormDemo struct {
	demos.DemoBase
	form *components.Form
}

// Component returns the demo component.
func (d *FormDemo) Component() tview.Primitive {
	// Build a form using the FormBuilder
	d.form = components.NewFormBuilder().
		Text("username", "Username").
			Placeholder("Enter username").
			Validate(validators.Required(), validators.MinLength(3)).
			Done().
		Text("email", "Email").
			Placeholder("user@example.com").
			Validate(validators.Required(), validators.Email()).
			Done().
		Select("role", "Role", []string{"Admin", "User", "Guest"}).
			Default("User").
			Done().
		Checkbox("newsletter", "Subscribe to newsletter").
			Checked(true).
			Done().
		OnSubmit(func(values map[string]any) {
			// Handle form submission
		}).
		Build()

	d.form.SetBackgroundColor(theme.Bg())

	d.ResetFunc = func() {
		d.form.Clear()
	}

	return d.form
}

const formCode = `package main

import (
    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/validators"
)

// Build a form using FormBuilder
form := components.NewFormBuilder().
    Text("username", "Username").
        Placeholder("Enter username").
        Validate(validators.Required, validators.MinLength(3)).
        Done().
    Text("email", "Email").
        Placeholder("user@example.com").
        Validate(validators.Required, validators.Email).
        Done().
    Select("role", "Role", []string{"Admin", "User", "Guest"}).
        Default("User").
        Done().
    Checkbox("newsletter", "Subscribe to newsletter").
        Checked(true).
        Done().
    OnSubmit(func(values map[string]any) {
        username := values["username"].(string)
        email := values["email"].(string)
        fmt.Printf("User: %s, Email: %s\n", username, email)
    }).
    Build()

// Validate all fields
if form.IsValid() {
    form.Submit()
}

// Get values programmatically
values := form.GetValues()

// Set values programmatically
form.SetValues(map[string]any{
    "username": "john",
    "email":    "john@example.com",
})

// Wrap form in a modal
modal := components.NewFormBuilder().
    Text("confirm", "Type 'DELETE' to confirm").
        Validate(validators.OneOf("DELETE")).
        Done().
    AsConfirmModal("Confirm Deletion", 50, 10)
`
