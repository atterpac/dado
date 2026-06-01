package advanced

import (
	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&MasterDetailDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "MasterDetail",
			DemoDescription: "List with preview panel",
			DemoCategory:    demos.Advanced,
			DemoCode:        masterDetailCode,
		},
	})
}

// MasterDetailDemo demonstrates the MasterDetailView component.
type MasterDetailDemo struct {
	demos.DemoBase
	view *components.MasterDetailView
}

// Component returns the demo component.
func (d *MasterDetailDemo) Component() core.Widget {
	// Create master content (list)
	master := components.NewList()
	master.SetItems([]components.ListItem{
		{Text: "Item 1", Secondary: "First item description"},
		{Text: "Item 2", Secondary: "Second item description"},
		{Text: "Item 3", Secondary: "Third item description"},
	})

	// Create detail content (preview)
	detail := core.NewTextView()
	detail.SetText("Select an item to see details\n\nThis panel shows a preview of the selected item.")
	detail.Box.SetBackgroundColor(theme.Bg())

	d.view = components.NewMasterDetailView().
		SetMasterTitle("Items").
		SetDetailTitle("Preview").
		SetMasterContent(master).
		SetDetailContent(detail).
		SetRatio(0.4)

	return d.view
}

const masterDetailCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create master-detail view
view := components.NewMasterDetailView().
    SetMasterTitle("Workflows").
    SetDetailTitle("Preview").
    SetMasterContent(listComponent).
    SetDetailContent(previewComponent).
    SetRatio(0.4) // 40% for master

// Or use config
view := components.NewMasterDetailViewConfig(components.MasterDetailConfig{
    MasterTitle:   "Items",
    DetailTitle:   "Details",
    MasterContent: list,
    DetailContent: detail,
    Ratio:         0.5,
    Resizable:     true,
    EmptyIcon:     theme.IconList,
    EmptyTitle:    "No Selection",
    EmptyMessage:  "Select an item to view details",
})

// Toggle detail panel
view.ToggleDetail()
view.SetShowDetail(false)

// Update detail content dynamically
view.SetDetailContent(newPreview)

// Clear to show empty state
view.ClearDetail()

// Callbacks
view.SetOnSelectionChange(func() {
    updatePreview()
})

view.SetOnDetailToggle(func(visible bool) {
    fmt.Printf("Detail panel: %v\n", visible)
})

// Focus management
view.FocusMaster()
view.FocusDetail()
`
