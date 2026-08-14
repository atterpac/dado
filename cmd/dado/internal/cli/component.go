package cli

import "fmt"

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
	{Name: "Core", Components: []ComponentInfo{
		{"Panel", "Rounded border container with title"}, {"Modal", "Centered modal dialog"},
		{"Table", "Enhanced table with multi-select and sorting"}, {"Tree", "Collapsible tree view with lazy loading"},
		{"Tabs", "Tabbed container with badges"}, {"Split", "Resizable split panes"},
		{"KeyHintBar", "Key hints display bar"}, {"Empty", "Empty/loading/error state component"},
	}},
	{Name: "Forms", Components: []ComponentInfo{
		{"TextField", "Single-line text input with validation"}, {"TextArea", "Multi-line text input"},
		{"Select", "Dropdown selection"}, {"MultiSelect", "Multi-choice selection"},
		{"Checkbox", "Boolean toggle"}, {"RadioGroup", "Single choice from options"},
		{"Form", "Form container with focus management"},
	}},
	{Name: "Progress", Components: []ComponentInfo{
		{"ProgressBar", "Horizontal progress bar"}, {"Spinner", "Animated loading indicator"},
		{"Gauge", "Arc-style progress indicator"}, {"Sparkline", "Minimal line chart"},
	}},
	{Name: "Recipes", Components: []ComponentInfo{
		{"ResourceList", "K9s-style filterable list with actions"}, {"LogViewer", "Streaming log display with search"},
		{"Dashboard", "Multi-pane status dashboard"},
	}},
}

// RunComponent handles the "component" command.
func RunComponent(args []string) {
	if len(args) == 0 || args[0] == "list" {
		printComponentList()
		return
	}
	renderCLILines([]cliLine{errorLine(fmt.Sprintf("Unknown component command: %s", args[0]))})
}

func printComponentList() {
	lines := append([]cliLine{blank()}, logoLines()...)
	lines = append(lines, line(styled("  Available components", cliMuted)))
	for _, section := range componentSections {
		lines = append(lines, blank(), sectionLine(section.Name))
		for _, component := range section.Components {
			lines = append(lines, line(styled("    "+padASCII(component.Name, 14)+" ", cliAccent), styled(component.Desc, cliMuted)))
		}
	}
	lines = append(lines, blank(), line(styled("  Documentation: ", cliMuted), text("https://github.com/atterpac/dado/docs")), blank())
	renderCLILines(lines)
}
