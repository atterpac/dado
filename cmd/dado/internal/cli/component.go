package cli

import (
	"fmt"

	ui "github.com/atterpac/dado/inline"
)

// ComponentInfo contains component metadata.
type ComponentInfo struct {
	Name string
	Desc string
}

// ComponentSection groups related components.
type ComponentSection struct {
	Name       string
	Components []ComponentInfo
}

var componentSections = []ComponentSection{
	{
		Name: "Core",
		Components: []ComponentInfo{
			{"Panel", "Rounded border container with title"},
			{"Modal", "Centered modal dialog"},
			{"Table", "Enhanced table with multi-select and sorting"},
			{"Tree", "Collapsible tree view with lazy loading"},
			{"Tabs", "Tabbed container with badges"},
			{"Split", "Resizable split panes"},
			{"KeyHintBar", "Key hints display bar"},
			{"Empty", "Empty/loading/error state component"},
		},
	},
	{
		Name: "Forms",
		Components: []ComponentInfo{
			{"TextField", "Single-line text input with validation"},
			{"TextArea", "Multi-line text input"},
			{"Select", "Dropdown selection"},
			{"MultiSelect", "Multi-choice selection"},
			{"Checkbox", "Boolean toggle"},
			{"RadioGroup", "Single choice from options"},
			{"Form", "Form container with focus management"},
		},
	},
	{
		Name: "Progress",
		Components: []ComponentInfo{
			{"ProgressBar", "Horizontal progress bar"},
			{"Spinner", "Animated loading indicator"},
			{"Gauge", "Arc-style progress indicator"},
			{"Sparkline", "Minimal line chart"},
		},
	},
	{
		Name: "Recipes",
		Components: []ComponentInfo{
			{"ResourceList", "K9s-style filterable list with actions"},
			{"LogViewer", "Streaming log display with search"},
			{"Dashboard", "Multi-pane status dashboard"},
		},
	},
}

// RunComponent handles the "component" command.
func RunComponent(args []string) {
	if len(args) == 0 || args[0] == "list" {
		printComponentList()
		return
	}

	ui.PrintError(fmt.Sprintf("Unknown component command: %s", args[0]))
}

func printComponentList() {
	fmt.Println()
	ui.PrintLogo()
	fmt.Printf("  %sAvailable components%s\n", ui.Dim, ui.Reset)

	for _, section := range componentSections {
		fmt.Printf("\n  %s%s%s%s\n", ui.Bold, ui.BrightWhite, section.Name, ui.Reset)
		for _, comp := range section.Components {
			fmt.Printf("    %s%-14s%s %s%s%s\n", ui.Cyan, comp.Name, ui.Reset, ui.Dim, comp.Desc, ui.Reset)
		}
	}

	fmt.Println()
	fmt.Printf("  %sDocumentation:%s https://github.com/atterpac/dado/docs\n\n", ui.Dim, ui.Reset)
}
