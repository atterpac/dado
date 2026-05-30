package advanced

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&FinderDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Finder",
			DemoDescription: "Fuzzy search with preview",
			DemoCategory:    demos.Advanced,
			DemoCode:        finderCode,
		},
	})
}

// FinderDemo demonstrates the Finder component.
type FinderDemo struct {
	demos.DemoBase
	finder          *components.Finder
	showDescription bool
}

// Component returns the demo component.
func (d *FinderDemo) Component() tview.Primitive {
	d.showDescription = true

	d.finder = components.NewFinder().
		SetPrompt("> ").
		SetShowDescription(d.showDescription)

	// Sample items
	items := []components.FinderItem{
		{ID: "1", Label: "main.go", Description: "Application entry point", Category: "Source", Icon: theme.IconFile},
		{ID: "2", Label: "config.go", Description: "Configuration handling", Category: "Source", Icon: theme.IconFile},
		{ID: "3", Label: "server.go", Description: "HTTP server setup", Category: "Source", Icon: theme.IconFile},
		{ID: "4", Label: "handler.go", Description: "Request handlers", Category: "Source", Icon: theme.IconFile},
		{ID: "5", Label: "README.md", Description: "Project documentation", Category: "Docs", Icon: theme.IconFile},
		{ID: "6", Label: "CHANGELOG.md", Description: "Version history", Category: "Docs", Icon: theme.IconFile},
		{ID: "7", Label: "go.mod", Description: "Go module definition", Category: "Config", Icon: theme.IconSettings},
		{ID: "8", Label: "go.sum", Description: "Dependency checksums", Category: "Config", Icon: theme.IconSettings},
		{ID: "9", Label: ".gitignore", Description: "Git ignore rules", Category: "Config", Icon: theme.IconSettings},
		{ID: "10", Label: "Dockerfile", Description: "Container definition", Category: "Deploy", Icon: theme.IconFolder},
	}

	d.finder.SetItems(items)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showDescription", "Show item descriptions",
			func() bool { return d.showDescription },
			func(v bool) { d.showDescription = v; d.finder.SetShowDescription(v) },
			true,
		),
	}
	d.ResetFunc = func() {
		d.finder.SetQuery("")
	}

	return d.finder
}

const finderCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create finder
finder := components.NewFinder().
    SetPrompt("> ").
    SetShowDescription(true).
    SetShowIcons(true).
    SetMaxVisible(15)

// Add searchable items
items := []components.FinderItem{
    {
        ID:          "1",
        Label:       "main.go",
        Description: "Application entry point",
        Category:    "Source",
        Icon:        theme.IconFile,
        Keywords:    []string{"entry", "start"},
    },
}
finder.SetItems(items)

// Category ordering
finder.SetCategories([]components.FinderCategory{
    {Name: "Source", Priority: 1},
    {Name: "Config", Priority: 2},
    {Name: "Docs", Priority: 3},
})

// Preview panel
finder.SetPreviewFunc(func(item components.FinderItem) string {
    return loadFilePreview(item.ID)
})

// Callbacks
finder.SetOnSelect(func(item components.FinderItem) {
    openFile(item.ID)
})

finder.SetOnCancel(func() {
    closeFinder()
})

// Track recent items for priority
finder.AddRecent("1")
`
