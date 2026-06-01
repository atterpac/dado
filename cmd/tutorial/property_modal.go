package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// PropertyModal displays a property editor for a demo component.
type PropertyModal struct {
	*components.Modal
	form    *components.Form
	demo    demos.Demo
	onClose func()
	subs    components.Subscriptions
}

// Subs returns the modal's subscription set; release on dismissal.
func (m *PropertyModal) Subs() *components.Subscriptions { return &m.subs }

// NewPropertyModal creates a new property editor modal.
func NewPropertyModal(demo demos.Demo, onClose func()) *PropertyModal {
	properties := demo.Properties()

	modal := components.NewModal(components.ModalConfig{
		Title:     "Properties: " + demo.Name(),
		Width:     50,
		Height:    calculateModalHeight(properties),
		MinWidth:  40,
		MinHeight: 8,
		Backdrop:  true,
	})

	m := &PropertyModal{
		Modal:   modal,
		demo:    demo,
		onClose: onClose,
	}

	if len(properties) == 0 {
		// Show message when no properties
		noProps := core.NewTextView()
		noProps.SetText("No editable properties for this component")
		noProps.Box.SetBackgroundColor(theme.Bg())
		modal.SetContent(noProps)
		modal.SetFocusOnShow(noProps)
	} else {
		m.buildForm(properties)
		modal.SetContent(m.form)
		modal.SetFocusOnShow(m.form)
	}

	modal.SetDismissOnEsc(true)
	modal.SetOnClose(onClose)

	modal.SetHints([]components.KeyHint{
		{Key: "Tab", Description: "Next Field"},
		{Key: "Space", Description: "Toggle/Select"},
		{Key: "Esc", Description: "Close"},
	})

	if m.form != nil {
		m.subs.Add(theme.RegisterFn(func(c tcell.Color) { m.form.SetBackgroundColor(c) }))
	}

	return m
}

// calculateModalHeight calculates the modal height based on number of properties.
func calculateModalHeight(properties []demos.Property) int {
	// Base height + 2 lines per property + padding
	height := 6 + len(properties)*3
	if height > 20 {
		height = 20
	}
	if height < 10 {
		height = 10
	}
	return height
}

// buildForm creates the property editor form.
func (m *PropertyModal) buildForm(properties []demos.Property) {
	form := components.NewForm()
	form.SetBackgroundColor(theme.Bg())

	for _, prop := range properties {
		propName := prop.Name

		switch prop.Type {
		case demos.PropString:
			val, _ := prop.Value.(string)
			field := components.NewTextField(propName).
				SetLabel(propName).
				SetValue(val).
				SetPlaceholder(prop.Description)
			field.SetOnChange(func(event *components.ChangeEvent[string]) {
				m.demo.ApplyProperty(propName, event.NewValue)
			})
			form.AddField(field)

		case demos.PropBool:
			val, _ := prop.Value.(bool)
			field := components.NewCheckbox(propName).
				SetLabel(propName).
				SetChecked(val)
			field.SetOnChange(func(event *components.ChangeEvent[bool]) {
				m.demo.ApplyProperty(propName, event.NewValue)
			})
			form.AddField(field)

		case demos.PropInt:
			val, _ := prop.Value.(int)
			field := components.NewTextField(propName).
				SetLabel(propName).
				SetValue(fmt.Sprintf("%d", val)).
				SetPlaceholder(prop.Description)
			field.SetOnChange(func(event *components.ChangeEvent[string]) {
				var intVal int
				if _, err := fmt.Sscanf(event.NewValue, "%d", &intVal); err == nil {
					m.demo.ApplyProperty(propName, intVal)
				}
			})
			form.AddField(field)

		case demos.PropSelect:
			if len(prop.Options) > 0 {
				initialVal, _ := prop.Value.(string)
				field := components.NewSelect(propName).
					SetLabel(propName).
					SetOptions(prop.Options).
					SetDefault(initialVal)
				field.SetOnChange(func(event *components.ChangeEvent[components.SelectOption]) {
					m.demo.ApplyProperty(propName, event.NewValue.Value)
				})
				form.AddField(field)
			}
		}
	}

	m.form = form
}

// Start implements nav.Component.
func (m *PropertyModal) Start() {}

// Stop implements nav.Component.
func (m *PropertyModal) Stop() { m.subs.Release() }

// Hints implements nav.Component.
func (m *PropertyModal) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "Tab", Description: "Next"},
		{Key: "Esc", Description: "Close"},
	}
}

// Focus handles focus.
func (m *PropertyModal) Focus() {
	if m.form != nil {
		m.form.Focus()
	} else {
		m.Modal.Focus()
	}
}

// HasFocus returns focus state.
func (m *PropertyModal) HasFocus() bool {
	if m.form != nil {
		return m.form.HasFocus()
	}
	return m.Modal.HasFocus()
}
