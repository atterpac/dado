package advanced

import (
	"strings"

	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
)

func init() {
	demos.Register(&AutocompleteDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Autocomplete",
			DemoDescription: "Input with suggestions",
			DemoCategory:    demos.Advanced,
			DemoCode:        autocompleteCode,
		},
	})
}

// AutocompleteDemo demonstrates the AutocompleteInput component.
type AutocompleteDemo struct {
	demos.DemoBase
	input          *components.AutocompleteInput
	maxSuggestions int
}

// Component returns the demo component.
func (d *AutocompleteDemo) Component() tview.Primitive {
	d.maxSuggestions = 8

	d.input = components.NewAutocompleteInput().
		SetPrompt("> ").
		SetPlaceholder("Type a fruit name...").
		SetTitle("Fruit Search").
		SetMaxSuggestions(d.maxSuggestions)

	// Sample data
	fruits := []string{
		"Apple", "Apricot", "Avocado",
		"Banana", "Blackberry", "Blueberry",
		"Cherry", "Coconut", "Cranberry",
		"Date", "Dragonfruit",
		"Elderberry",
		"Fig",
		"Grape", "Grapefruit", "Guava",
		"Honeydew",
		"Kiwi", "Kumquat",
		"Lemon", "Lime", "Lychee",
		"Mango", "Melon",
		"Orange",
		"Papaya", "Peach", "Pear", "Pineapple", "Plum", "Pomegranate",
		"Raspberry",
		"Strawberry",
		"Tangerine",
		"Watermelon",
	}

	d.input.SetSuggestionProvider(func(text string, cursorPos int) []components.Suggestion {
		if text == "" {
			return nil
		}
		text = strings.ToLower(text)
		var suggestions []components.Suggestion
		for _, f := range fruits {
			if strings.Contains(strings.ToLower(f), text) {
				suggestions = append(suggestions, components.Suggestion{
					Text:        f,
					Description: "A delicious fruit",
				})
			}
		}
		return suggestions
	})

	d.Props = []demos.PropertyDescriptor{
		demos.IntProp("maxSuggestions", "Max visible suggestions",
			func() int { return d.maxSuggestions },
			func(v int) { d.maxSuggestions = v; d.input.SetMaxSuggestions(v) },
			8,
		),
	}
	d.ResetFunc = func() {
		d.input.SetText("")
	}

	return d.input
}

const autocompleteCode = `package main

import (
    "github.com/atterpac/jig/components"
)

// Create autocomplete input
input := components.NewAutocompleteInput().
    SetPrompt("> ").
    SetPlaceholder("Type to search...").
    SetTitle("Search").
    SetMaxSuggestions(10)

// Provide suggestions based on input
input.SetSuggestionProvider(func(text string, cursorPos int) []components.Suggestion {
    return []components.Suggestion{
        {Text: "Option 1", Description: "First option"},
        {Text: "Option 2", Description: "Second option", Category: "Group A"},
    }
})

// Optional: History navigation (up/down when no suggestions)
input.SetHistoryProvider(func(direction int) string {
    // direction: -1 for previous, +1 for next
    return getHistoryEntry(direction)
})

// Callbacks
input.SetOnSubmit(func(text string) {
    executeSearch(text)
})

input.SetOnSelect(func(suggestion components.Suggestion) {
    fmt.Printf("Selected: %s\n", suggestion.Text)
})

input.SetOnChange(func(text string) {
    // Text changed
})
`
