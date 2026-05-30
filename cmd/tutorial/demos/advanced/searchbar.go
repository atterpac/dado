package advanced

import (
	"strings"

	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
)

func init() {
	demos.Register(&SearchBarDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "SearchBar",
			DemoDescription: "Search input with results dropdown",
			DemoCategory:    demos.Advanced,
			DemoCode:        searchBarCode,
		},
	})
}

// SearchBarDemo demonstrates the SearchBar component.
type SearchBarDemo struct {
	demos.DemoBase
	search   *components.SearchBar
	allItems []components.SearchResult
}

// Component returns the demo component.
func (d *SearchBarDemo) Component() tview.Primitive {
	// Sample items to search through
	d.allItems = []components.SearchResult{
		{Text: "main.go", Description: "src/", Icon: ""},
		{Text: "app.go", Description: "src/", Icon: ""},
		{Text: "config.yaml", Description: "etc/", Icon: ""},
		{Text: "README.md", Description: "./", Icon: ""},
		{Text: "components.go", Description: "pkg/", Icon: ""},
		{Text: "utils.go", Description: "pkg/", Icon: ""},
		{Text: "handler.go", Description: "api/", Icon: ""},
		{Text: "router.go", Description: "api/", Icon: ""},
		{Text: "Makefile", Description: "./", Icon: ""},
		{Text: "go.mod", Description: "./", Icon: ""},
	}

	d.search = components.NewSearchBar().
		SetPlaceholder("Search files...").
		SetIcon("").
		SetMaxResults(5)

	// Handle search input changes
	d.search.SetOnChange(func(query string) {
		if query == "" {
			d.search.ClearResults()
			return
		}

		// Filter items
		var matches []components.SearchResult
		queryLower := strings.ToLower(query)
		for _, item := range d.allItems {
			if strings.Contains(strings.ToLower(item.Text), queryLower) {
				matches = append(matches, item)
			}
		}
		d.search.SetResults(matches)
	})

	// Handle selection
	d.search.SetOnSelect(func(result components.SearchResult) {
		d.search.SetQuery(result.Text)
		d.search.ClearResults()
	})

	d.ResetFunc = func() {
		d.search.Clear()
	}

	return d.search
}

const searchBarCode = `package main

import "github.com/atterpac/dado/components"

// Create a search bar
search := components.NewSearchBar().
    SetPlaceholder("Search...").
    SetIcon("")

// Handle query changes (for live search)
search.SetOnChange(func(query string) {
    if query == "" {
        search.ClearResults()
        return
    }

    // Filter/fetch results
    results := filterItems(query)
    search.SetResults(results)
})

// Handle result selection
search.SetOnSelect(func(result components.SearchResult) {
    openFile(result.Text)
    search.Clear()
})

// Handle search submission
search.SetOnSearch(func(query string) {
    performSearch(query)
})

// Handle cancel (Escape)
search.SetOnCancel(func() {
    search.Clear()
})

// Set results manually
search.SetResults([]components.SearchResult{
    {Text: "main.go", Description: "src/", Icon: ""},
    {Text: "app.go", Description: "pkg/", Icon: ""},
})

// Show loading spinner
search.SetShowSpinner(true)

// Programmatic control
search.SetQuery("search term")
search.GetQuery()
search.Clear()

// Keyboard:
// Up/Down  - navigate results
// Enter    - select result or submit
// Escape   - cancel/clear
// Ctrl+U/K - clear line
// Ctrl+W   - delete word
`
