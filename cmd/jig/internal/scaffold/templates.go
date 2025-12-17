package scaffold

import "fmt"

// SponsorURL is the GitHub sponsors URL included in all generated apps.
const SponsorURL = "github.com/sponsors/atterpac"

// ThemeToImport converts a theme name to its Go variable name.
func ThemeToImport(themeName string) string {
	switch themeName {
	case "tokyonight-night":
		return "TokyoNightNight"
	case "tokyonight-storm":
		return "TokyoNightStorm"
	case "tokyonight-moon":
		return "TokyoNightMoon"
	case "tokyonight-day":
		return "TokyoNightDay"
	case "catppuccin-mocha":
		return "CatppuccinMocha"
	case "catppuccin-macchiato":
		return "CatppuccinMacchiato"
	case "catppuccin-frappe":
		return "CatppuccinFrappe"
	case "catppuccin-latte":
		return "CatppuccinLatte"
	case "dracula":
		return "Dracula"
	case "dracula-light":
		return "DraculaLight"
	case "gruvbox-dark":
		return "GruvboxDark"
	case "gruvbox-light":
		return "GruvboxLight"
	case "onedark":
		return "OneDark"
	case "onelight":
		return "OneLight"
	case "solarized-dark":
		return "SolarizedDark"
	case "solarized-light":
		return "SolarizedLight"
	case "rosepine":
		return "RosePine"
	case "rosepine-moon":
		return "RosePineMoon"
	case "rosepine-dawn":
		return "RosePineDawn"
	case "kanagawa":
		return "Kanagawa"
	case "everforest-dark":
		return "EverforestDark"
	case "everforest-light":
		return "EverforestLight"
	case "monokai":
		return "Monokai"
	case "github-dark":
		return "GitHubDark"
	case "github-light":
		return "GitHubLight"
	case "nord":
		return "Nord"
	default:
		return "TokyoNightNight"
	}
}

// GoMod generates a go.mod file.
func GoMod(name string) string {
	return fmt.Sprintf(`module %s

go 1.21

require github.com/atterpac/jig v0.1.0
`, name)
}

// SimpleMain generates a simple main.go file with full design system.
func SimpleMain(name, themeName string) string {
	themeImport := ThemeToImport(themeName)

	return fmt.Sprintf(`package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/input"
	"github.com/atterpac/jig/layout"
	"github.com/atterpac/jig/nav"
	"github.com/atterpac/jig/theme"
	"github.com/atterpac/jig/theme/themes"
)

const sponsorURL = "%s"

// ASCII logo - customize this for your app
const logo = `+"`"+`
   ____ ____ ____ ____
  ||%s ||%s ||%s ||%s ||
  ||__|||__|||__|||__||
  |/__\|/__\|/__\|/__\|
`+"`"+`

func main() {
	// 1. Initialize theme FIRST
	theme.SetProvider(themes.%s)

	// 2. Show splash screen
	if err := showSplash(); err != nil {
		return
	}

	// 3. Create layout components
	statusBar := layout.NewStatusBar().
		SetTitle("%s")

	menu := layout.NewMenu().
		SetRightText(theme.IconHeart + " " + sponsorURL)

	// 4. Create app with 4-tier layout
	app := layout.NewApp(layout.AppConfig{
		TopBar:     statusBar,
		BottomBar:  menu,
		ShowCrumbs: true,
	})

	// 5. Global input handler
	app.SetInputCapture(globalInputHandler(app))

	// 6. Push home view and run
	home := NewHomeView(app)
	app.Pages().Push(home)
	app.Crumbs().SetPath([]string{"Home"})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func showSplash() error {
	splash := components.NewSplash().
		SetLogo(logo).
		SetStatus("Starting %s...\n\nPress any key to continue").
		SetAutoDismiss(3 * time.Second).
		SetDismissKeys([]components.DismissKey{components.DismissAnyKey})

	// Add sponsor text
	splash.Build()

	app := tview.NewApplication()
	theme.SetApp(app)

	splash.SetOnClose(func() {
		app.Stop()
	})

	app.SetRoot(splash, true)
	return app.Run()
}

func globalInputHandler(app *layout.App) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		// Quit on root view
		if event.Rune() == 'q' && app.Pages().StackDepth() <= 1 {
			app.Stop()
			return nil
		}

		// Go back
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyBackspace {
			if app.Pages().CanPop() {
				app.Pages().Pop()
				return nil
			}
		}

		// Theme selector
		if event.Rune() == 'T' {
			showThemeSelector(app)
			return nil
		}

		// Help
		if event.Rune() == '?' {
			showHelp(app)
			return nil
		}

		return event
	}
}

func showThemeSelector(app *layout.App) {
	selector := theme.NewThemeSelectorModal(themes.Names(), themes.DefaultName)

	selector.SetOnSelect(func(name string) {
		newTheme := themes.Get(name)
		if newTheme != nil {
			theme.SetProvider(newTheme)
		}
		app.Pages().Pop()
	})

	selector.SetOnCancel(func() {
		app.Pages().Pop()
	})

	app.Pages().Push(&themeModalWrapper{modal: selector})
	app.SetFocus(selector)
}

func showHelp(app *layout.App) {
	helpText := `+"`"+`
Keyboard Shortcuts
──────────────────

Navigation:
  j/k, ↑/↓    Move up/down
  Enter       Select item
  Esc         Go back
  Tab         Switch pane

Global:
  T           Theme selector
  ?           Show this help
  p           Toggle preview
  q           Quit (on home)

Press any key to close
`+"`"+`

	modal := components.NewModal(components.ModalConfig{
		Title:  "Help",
		Width:  50,
		Height: 18,
	})

	content := tview.NewTextView().
		SetText(helpText).
		SetDynamicColors(true)

	modal.SetContent(content)
	modal.SetOnClose(func() {
		app.Pages().Pop()
	})

	app.ShowModal(modal)
}

// themeModalWrapper wraps the theme selector for nav.Component
type themeModalWrapper struct {
	modal *theme.ThemeSelectorModal
}

func (t *themeModalWrapper) Start()                      {}
func (t *themeModalWrapper) Stop()                       {}
func (t *themeModalWrapper) Hints() []components.KeyHint { return nil }
func (t *themeModalWrapper) Draw(screen tcell.Screen)    { t.modal.Draw(screen) }
func (t *themeModalWrapper) GetRect() (int, int, int, int) {
	return t.modal.GetRect()
}
func (t *themeModalWrapper) SetRect(x, y, w, h int) { t.modal.SetRect(x, y, w, h) }
func (t *themeModalWrapper) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.modal.InputHandler()
}
func (t *themeModalWrapper) Focus(delegate func(tview.Primitive)) { t.modal.Focus(delegate) }
func (t *themeModalWrapper) Blur()                                 { t.modal.Blur() }
func (t *themeModalWrapper) HasFocus() bool                        { return t.modal.HasFocus() }
func (t *themeModalWrapper) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return t.modal.MouseHandler()
}
func (t *themeModalWrapper) PasteHandler() func(string, func(tview.Primitive)) { return nil }

// HomeView is the main view with two-panel layout
type HomeView struct {
	*tview.Flex
	app         *layout.App
	split       *components.Split
	table       *components.Table
	preview     *tview.TextView
	actions     *input.ActionRegistry
	showPreview bool
}

func NewHomeView(app *layout.App) *HomeView {
	v := &HomeView{
		Flex:        tview.NewFlex(),
		app:         app,
		table:       components.NewTable(),
		preview:     tview.NewTextView(),
		showPreview: true,
	}
	v.setup()
	return v
}

func (v *HomeView) setup() {
	// Configure table
	v.table.SetHeaders("ID", "Name", "Status")
	v.table.SetOnSelect(v.onSelect)

	// Add sample data
	v.table.AddRow("1", "First Item", "Active")
	v.table.AddRow("2", "Second Item", "Pending")
	v.table.AddRow("3", "Third Item", "Complete")

	// Configure preview
	v.preview.SetDynamicColors(true)
	v.preview.SetWordWrap(true)
	v.preview.SetBackgroundColor(theme.Bg())
	theme.Register(v.preview) // Auto-update background on theme change
	v.updatePreview(0)

	// Create panels
	leftPanel := components.NewPanel().
		SetTitle("Items").
		SetContent(v.table)

	rightPanel := components.NewPanel().
		SetTitle("Preview").
		SetContent(v.preview)

	// Create split layout (60/40)
	v.split = components.NewSplit().
		SetDirection(components.SplitHorizontal).
		SetRatio(0.6).
		SetLeft(leftPanel).
		SetRight(rightPanel)

	// Register actions
	v.actions = input.NewActionRegistry().
		AddSimple("toggle_preview", 'p', "Preview", v.togglePreview).
		AddSimple("refresh", 'r', "Refresh", v.refresh)

	v.AddItem(v.split, 0, 1, true)
}

func (v *HomeView) onSelect(row int) {
	v.updatePreview(row)
}

func (v *HomeView) updatePreview(row int) {
	if row < 0 {
		v.preview.SetText("")
		return
	}

	// Get row data and display in preview
	text := fmt.Sprintf("[%s::b]Item Details[-:-:-]\n\n", theme.TagAccent())
	text += fmt.Sprintf("[%s]Row:[%s] %%d\n", theme.TagFgDim(), theme.TagFg())
	text += fmt.Sprintf("[%s]Status:[%s] Active\n\n", theme.TagFgDim(), theme.TagFg())
	text += fmt.Sprintf("[%s]Select an item to see details here.[-]", theme.TagFgDim())

	v.preview.SetText(fmt.Sprintf(text, row+1))
}

func (v *HomeView) togglePreview() {
	v.showPreview = !v.showPreview
	if v.showPreview {
		v.split.SetRatio(0.6)
	} else {
		v.split.SetRatio(1.0)
	}
}

func (v *HomeView) refresh() {
	// Refresh data
	v.table.Clear()
	v.table.SetHeaders("ID", "Name", "Status")
	v.table.AddRow("1", "First Item", "Active")
	v.table.AddRow("2", "Second Item", "Pending")
	v.table.AddRow("3", "Third Item", "Complete")
}

// Start is called when the view becomes active
func (v *HomeView) Start() {
	v.app.Crumbs().SetPath([]string{"Home"})
	// Refresh preview to pick up any theme changes
	row, _ := v.table.Table.GetSelection()
	v.updatePreview(row - 1) // Subtract 1 for header row
}

// Stop is called when the view becomes inactive
func (v *HomeView) Stop() {}

// Hints returns the key hints for this view
func (v *HomeView) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "p", Description: "Preview"},
		{Key: "T", Description: "Theme"},
		{Key: "?", Description: "Help"},
	}
}

func (v *HomeView) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return v.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if v.actions.Handle(event) {
			return
		}
		if handler := v.split.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

var _ nav.Component = (*HomeView)(nil)
`, SponsorURL, name[:1], name[1:2], name[2:3], name[3:4], themeImport, name, name)
}

// StructuredMain generates a main.go for structured projects.
func StructuredMain(name, themeName string) string {
	themeImport := ThemeToImport(themeName)

	return fmt.Sprintf(`package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"%s/internal/views"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/layout"
	"github.com/atterpac/jig/theme"
	"github.com/atterpac/jig/theme/themes"
)

const sponsorURL = "%s"

// ASCII logo - customize this for your app
const logo = `+"`"+`
    ___       ___       ___       ___
   /\  \     /\  \     /\  \     /\  \
  /::\  \   /::\  \   /::\  \   /::\  \
 /:/\:\__\ /::\:\__\ /::\:\__\ /::\:\__\
 \:\/:/  / \/\::/  / \:\:\/  / \:\:\/  /
  \::/  /    /:/  /   \:\/  /   \:\/  /
   \/__/     \/__/     \/__/     \/__/
`+"`"+`

func main() {
	// 1. Initialize theme FIRST
	theme.SetProvider(themes.%s)

	// 2. Show splash screen
	if err := showSplash(); err != nil {
		return
	}

	// 3. Create layout components
	statusBar := layout.NewStatusBar().
		SetTitle("%s")

	menu := layout.NewMenu().
		SetRightText(theme.IconHeart + " " + sponsorURL)

	// 4. Create app with 4-tier layout
	app := layout.NewApp(layout.AppConfig{
		TopBar:     statusBar,
		BottomBar:  menu,
		ShowCrumbs: true,
	})

	// 5. Global input handler
	app.SetInputCapture(globalInputHandler(app))

	// 6. Push home view and run
	home := views.NewHomeView(app)
	app.Pages().Push(home)
	app.Crumbs().SetPath([]string{"Home"})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func showSplash() error {
	splash := components.NewSplash().
		SetLogo(logo).
		SetStatus("Starting %s...\n\nPress any key to continue").
		SetAutoDismiss(3 * time.Second).
		SetDismissKeys([]components.DismissKey{components.DismissAnyKey})

	splash.Build()

	app := tview.NewApplication()
	theme.SetApp(app)

	splash.SetOnClose(func() {
		app.Stop()
	})

	app.SetRoot(splash, true)
	return app.Run()
}

func globalInputHandler(app *layout.App) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		// Quit on root view
		if event.Rune() == 'q' && app.Pages().StackDepth() <= 1 {
			app.Stop()
			return nil
		}

		// Go back
		if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyBackspace {
			if app.Pages().CanPop() {
				app.Pages().Pop()
				return nil
			}
		}

		// Theme selector
		if event.Rune() == 'T' {
			views.ShowThemeSelector(app)
			return nil
		}

		// Help
		if event.Rune() == '?' {
			views.ShowHelp(app)
			return nil
		}

		return event
	}
}
`, name, SponsorURL, themeImport, name, name)
}

// HomeView generates the home view file with two-panel layout.
func HomeView(_ string) string {
	return `package views

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/input"
	"github.com/atterpac/jig/layout"
	"github.com/atterpac/jig/nav"
	"github.com/atterpac/jig/theme"
	"github.com/atterpac/jig/theme/themes"
)

// HomeView is the main application view with two-panel layout
type HomeView struct {
	*tview.Flex
	app         *layout.App
	split       *components.Split
	table       *components.Table
	preview     *tview.TextView
	actions     *input.ActionRegistry
	showPreview bool
}

// NewHomeView creates a new home view
func NewHomeView(app *layout.App) *HomeView {
	v := &HomeView{
		Flex:        tview.NewFlex(),
		app:         app,
		table:       components.NewTable(),
		preview:     tview.NewTextView(),
		showPreview: true,
	}
	v.setup()
	return v
}

func (v *HomeView) setup() {
	// Configure table
	v.table.SetHeaders("ID", "Name", "Status")
	v.table.SetOnSelect(v.onSelect)

	// Add sample data
	v.table.AddRow("1", "First Item", "Active")
	v.table.AddRow("2", "Second Item", "Pending")
	v.table.AddRow("3", "Third Item", "Complete")

	// Configure preview
	v.preview.SetDynamicColors(true)
	v.preview.SetWordWrap(true)
	v.preview.SetBackgroundColor(theme.Bg())
	theme.Register(v.preview) // Auto-update background on theme change
	v.updatePreview(0)

	// Create panels
	leftPanel := components.NewPanel().
		SetTitle("Items").
		SetContent(v.table)

	rightPanel := components.NewPanel().
		SetTitle("Preview").
		SetContent(v.preview)

	// Create split layout (60/40)
	v.split = components.NewSplit().
		SetDirection(components.SplitHorizontal).
		SetRatio(0.6).
		SetLeft(leftPanel).
		SetRight(rightPanel)

	// Register actions
	v.actions = input.NewActionRegistry().
		AddSimple("toggle_preview", 'p', "Preview", v.togglePreview).
		AddSimple("refresh", 'r', "Refresh", v.refresh)

	v.AddItem(v.split, 0, 1, true)
}

func (v *HomeView) onSelect(row int) {
	v.updatePreview(row)
}

func (v *HomeView) updatePreview(row int) {
	if row < 0 {
		v.preview.SetText("")
		return
	}

	text := fmt.Sprintf("[%%s::b]Item Details[-:-:-]\n\n", theme.TagAccent())
	text += fmt.Sprintf("[%%s]Row:[%%s] %%d\n", theme.TagFgDim(), theme.TagFg())
	text += fmt.Sprintf("[%%s]Status:[%%s] Active\n\n", theme.TagFgDim(), theme.TagFg())
	text += fmt.Sprintf("[%%s]Select an item to see details here.[-]", theme.TagFgDim())

	v.preview.SetText(fmt.Sprintf(text, row+1))
}

func (v *HomeView) togglePreview() {
	v.showPreview = !v.showPreview
	if v.showPreview {
		v.split.SetRatio(0.6)
	} else {
		v.split.SetRatio(1.0)
	}
}

func (v *HomeView) refresh() {
	v.table.Clear()
	v.table.SetHeaders("ID", "Name", "Status")
	v.table.AddRow("1", "First Item", "Active")
	v.table.AddRow("2", "Second Item", "Pending")
	v.table.AddRow("3", "Third Item", "Complete")
}

// Start is called when the view becomes active
func (v *HomeView) Start() {
	v.app.Crumbs().SetPath([]string{"Home"})
	// Refresh preview to pick up any theme changes
	row, _ := v.table.Table.GetSelection()
	v.updatePreview(row - 1) // Subtract 1 for header row
}

// Stop is called when the view becomes inactive
func (v *HomeView) Stop() {}

// Hints returns the key hints for this view
func (v *HomeView) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "p", Description: "Preview"},
		{Key: "T", Description: "Theme"},
		{Key: "?", Description: "Help"},
	}
}

// InputHandler handles keyboard input
func (v *HomeView) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return v.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if v.actions.Handle(event) {
			return
		}
		if handler := v.split.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

// Ensure HomeView implements nav.Component
var _ nav.Component = (*HomeView)(nil)

// ShowThemeSelector displays the theme selector modal
func ShowThemeSelector(app *layout.App) {
	selector := theme.NewThemeSelectorModal(themes.Names(), themes.DefaultName)

	// Live preview when navigating
	selector.SetOnPreview(func(name string) {
		newTheme := themes.Get(name)
		if newTheme != nil {
			theme.SetProvider(newTheme)
		}
	})

	selector.SetOnSelect(func(name string) {
		newTheme := themes.Get(name)
		if newTheme != nil {
			theme.SetProvider(newTheme)
		}
		app.Pages().Pop()
	})

	selector.SetOnCancel(func() {
		app.Pages().Pop()
	})

	app.Pages().Push(&themeModalWrapper{modal: selector})
	app.SetFocus(selector)
}

// ShowHelp displays the help modal
func ShowHelp(app *layout.App) {
	helpText := `+"`"+`
Keyboard Shortcuts
──────────────────

Navigation:
  j/k, ↑/↓    Move up/down
  Enter       Select item
  Esc         Go back
  Tab         Switch pane

Global:
  T           Theme selector
  ?           Show this help
  p           Toggle preview
  q           Quit (on home)

Press any key to close
`+"`"+`

	modal := components.NewModal(components.ModalConfig{
		Title:  "Help",
		Width:  50,
		Height: 18,
	})

	content := tview.NewTextView().
		SetText(helpText).
		SetDynamicColors(true)

	modal.SetContent(content)
	modal.SetOnClose(func() {
		app.Pages().Pop()
	})

	app.ShowModal(modal)
}

// themeModalWrapper wraps the theme selector for nav.Component
type themeModalWrapper struct {
	modal *theme.ThemeSelectorModal
}

func (t *themeModalWrapper) Start()                      {}
func (t *themeModalWrapper) Stop()                       {}
func (t *themeModalWrapper) Hints() []components.KeyHint { return nil }
func (t *themeModalWrapper) Draw(screen tcell.Screen)    { t.modal.Draw(screen) }
func (t *themeModalWrapper) GetRect() (int, int, int, int) {
	return t.modal.GetRect()
}
func (t *themeModalWrapper) SetRect(x, y, w, h int) { t.modal.SetRect(x, y, w, h) }
func (t *themeModalWrapper) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.modal.InputHandler()
}
func (t *themeModalWrapper) Focus(delegate func(tview.Primitive)) { t.modal.Focus(delegate) }
func (t *themeModalWrapper) Blur()                                 { t.modal.Blur() }
func (t *themeModalWrapper) HasFocus() bool                        { return t.modal.HasFocus() }
func (t *themeModalWrapper) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return t.modal.MouseHandler()
}
func (t *themeModalWrapper) PasteHandler() func(string, func(tview.Primitive)) { return nil }
`
}

// Actions generates the actions registry file.
func Actions() string {
	return `package actions

import (
	"github.com/atterpac/jig/input"
	"github.com/atterpac/jig/theme"
)

// GlobalActions returns the global action registry
func GlobalActions() *input.ActionRegistry {
	actions := input.NewActionRegistry()

	actions.AddSimple("quit", 'q', "Quit", func() {
		theme.GetApp().Stop()
	})

	return actions
}
`
}

// Config generates the config file.
func Config(name string) string {
	return fmt.Sprintf(`package config

// Config holds application configuration
type Config struct {
	AppName string
	Version string
	Theme   string
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		AppName: "%s",
		Version: "0.1.0",
		Theme:   "tokyonight-night",
	}
}
`, name)
}

// Taskfile generates a Taskfile.yml.
func Taskfile(name string) string {
	return fmt.Sprintf(`# https://taskfile.dev
version: '3'

vars:
  APP_NAME: %s

tasks:
  default:
    desc: Run the application
    cmds:
      - go run ./cmd/{{.APP_NAME}}

  build:
    desc: Build the binary
    cmds:
      - go build -o {{.APP_NAME}} ./cmd/{{.APP_NAME}}

  test:
    desc: Run tests
    cmds:
      - go test ./...

  lint:
    desc: Run linter
    cmds:
      - golangci-lint run

  clean:
    desc: Remove build artifacts
    cmds:
      - rm -f {{.APP_NAME}}

  tidy:
    desc: Tidy dependencies
    cmds:
      - go mod tidy
`, name)
}

// Readme generates a README.md file.
func Readme(name string) string {
	return fmt.Sprintf(`# %s

A terminal user interface application built with [jig](https://github.com/atterpac/jig).

## Features

- Two-panel layout (list + preview)
- Theme selector (20+ themes)
- Keyboard-first navigation
- Splash screen with gradient logo

## Getting Started

### Prerequisites

- Go 1.21 or later
- [Task](https://taskfile.dev) (optional, for task runner)
- A Nerd Font for icons (recommended)

### Running

`+"`"+`bash
# Direct
go run .

# With Task
task
`+"`"+`

### Building

`+"`"+`bash
# Direct
go build -o %s

# With Task
task build
`+"`"+`

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| j/k | Navigate up/down |
| Enter | Select item |
| p | Toggle preview |
| T | Theme selector |
| ? | Help |
| Esc | Go back |
| q | Quit (on home) |

## Customization

### Theme

Press `+"`T`"+` to open the theme selector. Available themes include:
- Tokyo Night (night, storm, moon, day)
- Catppuccin (mocha, macchiato, frappe, latte)
- Dracula
- Nord
- Gruvbox
- And more...

### Logo

Edit the `+"`logo`"+` constant in main.go to customize the splash screen.

## License

MIT

---

Built with [jig](https://github.com/atterpac/jig) | [Sponsor](https://%s)
`, name, name, SponsorURL)
}
