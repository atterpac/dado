package intermediate

import (
	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&TableDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Table",
			DemoDescription: "Data table with selection",
			DemoCategory:    demos.Intermediate,
			DemoCode:        tableCode,
		},
	})
}

// TableDemo demonstrates the Table component.
type TableDemo struct {
	demos.DemoBase
	table       *components.Table
	multiSelect bool
}

// Component returns the demo component.
func (d *TableDemo) Component() core.Widget {
	d.multiSelect = false

	d.table = components.NewTable()
	d.table.SetHeaders("ID", "Name", "Status", "Created")
	d.table.AddRow("1", "Project Alpha", "Active", "2024-01-15")
	d.table.AddRow("2", "Project Beta", "Pending", "2024-02-20")
	d.table.AddRow("3", "Project Gamma", "Complete", "2024-03-10")
	d.table.AddRow("4", "Project Delta", "Active", "2024-04-05")
	d.table.AddRow("5", "Project Epsilon", "Archived", "2024-05-12")
	d.table.SetMultiSelect(d.multiSelect)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("multiSelect", "Allow multiple selection",
			func() bool { return d.multiSelect },
			func(v bool) { d.multiSelect = v; d.table.SetMultiSelect(v) },
			false,
		),
	}

	return d.table
}

const tableCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create a table
table := components.NewTable()

// Set headers (varargs)
table.SetHeaders("ID", "Name", "Status")

// Add data rows
table.AddRow("1", "Project Alpha", "Active")
table.AddRow("2", "Project Beta", "Pending")
table.AddRow("3", "Project Gamma", "Complete")

// Enable multi-select
table.SetMultiSelect(true)

// Handle selection
table.SetOnSelect(func(row int) {
    fmt.Printf("Selected row: %d\n", row)
})

// Add colored rows
table.AddColoredRow(
    []string{"4", "Error", "Failed"},
    []tcell.Color{theme.Fg(), theme.Error(), theme.Error()},
)

// Use key-based selection for stable selection during data updates
table.SetRowKey(func(row, col int, cell *core.TableCell) string {
    return fmt.Sprintf("row-%d", row)
})
`
