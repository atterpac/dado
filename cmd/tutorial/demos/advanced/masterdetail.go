package advanced

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
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
func (d *MasterDetailDemo) Component() tview.Primitive {
	// Create master content (list)
	master := tview.NewList()
	master.SetBackgroundColor(theme.Bg())
	master.SetMainTextColor(theme.Fg())
	master.SetSecondaryTextColor(theme.FgDim())
	master.AddItem("Item 1", "First item description", 0, nil)
	master.AddItem("Item 2", "Second item description", 0, nil)
	master.AddItem("Item 3", "Third item description", 0, nil)

	// Create detail content (preview)
	detail := tview.NewTextView()
	detail.SetText("Select an item to see details\n\nThis panel shows a preview of the selected item.")
	detail.SetBackgroundColor(theme.Bg())
	detail.SetTextColor(theme.Fg())

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
