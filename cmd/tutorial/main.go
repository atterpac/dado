package main

import (
	"log"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	_ "github.com/atterpac/dado/cmd/tutorial/demos/advanced"
	_ "github.com/atterpac/dado/cmd/tutorial/demos/basic"
	_ "github.com/atterpac/dado/cmd/tutorial/demos/intermediate"
	"github.com/atterpac/dado/layout"
	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

func main() {
	// Set initial theme
	theme.SetProvider(themes.Default())

	// Create the tutorial app
	tutorial := NewTutorial()

	// Build the app layout
	app := layout.NewApp(layout.AppConfig{
		TopBar:     tutorial.StatusBar(),
		BottomBar:  tutorial.Menu(),
		ShowCrumbs: false,
	})

	// Push the main tutorial view
	app.Pages().Push(tutorial)

	// Set up global input handling for tutorial-specific keys
	app.SetInputCapture(tutorial.HandleGlobalInput)

	// Register demos
	registerDemos()

	// Initialize the tutorial view
	tutorial.Initialize(app)

	// Run the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// registerDemos registers all component demos
func registerDemos() {
	// Demos are auto-registered via init() in each demo package
	// The imports above trigger registration
	_ = demos.DefaultRegistry
}
