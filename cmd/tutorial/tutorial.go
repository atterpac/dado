package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/layout"
	"github.com/atterpac/jig/theme"
)

// Tutorial is the main tutorial application component.
type Tutorial struct {
	*components.Split

	app           *layout.App
	statusBar     *layout.StatusBar
	menu          *layout.Menu
	sidebar       *Sidebar
	demoPanel     *components.Panel
	demoContent   *components.Layout
	currentDemo   demos.Demo
	demoComponent tview.Primitive // Cached to avoid re-creation
	subs          components.Subscriptions
}

// Subs returns the tutorial's subscription set; release on app teardown.
func (t *Tutorial) Subs() *components.Subscriptions { return &t.subs }

// NewTutorial creates a new tutorial application.
func NewTutorial() *Tutorial {
	t := &Tutorial{
		Split:       components.NewSplit(),
		statusBar:   layout.NewStatusBar(),
		menu:        layout.NewMenu(),
		sidebar:     NewSidebar(),
		demoPanel:   components.NewPanel(),
		demoContent: components.NewLayout().Vertical(),
	}

	// Configure status bar
	t.statusBar.SetTitle("Jig Tutorial")
	t.statusBar.SetTitleAlign(components.AlignCenter)

	// Configure demo panel
	t.demoPanel.SetTitle("Demo")
	t.demoPanel.SetTitleAlign(components.AlignLeft)
	t.demoPanel.SetContent(t.demoContent)

	// Configure split layout (sidebar on left, demo on right)
	t.Split.SetDirection(components.SplitHorizontal)
	t.Split.SetRatio(0.25) // 25% for sidebar
	t.Split.SetMinSize(20)
	t.Split.SetLeft(t.sidebar)
	t.Split.SetRight(t.demoPanel)

	// Wire up sidebar selection
	t.sidebar.SetOnSelect(func(demo demos.Demo) {
		t.selectDemo(demo)
	})

	// Update demo when navigating sidebar
	t.sidebar.SetOnHighlight(func(demo demos.Demo) {
		t.selectDemo(demo)
	})

	t.subs.Add(theme.Register(t.demoContent))

	return t
}

// Initialize sets up the tutorial after the app is created.
func (t *Tutorial) Initialize(app *layout.App) {
	t.app = app

	// Populate sidebar with registered demos
	t.sidebar.PopulateFromRegistry(demos.DefaultRegistry)

	// Select the first demo if available
	if all := demos.DefaultRegistry.All(); len(all) > 0 {
		t.selectDemo(all[0])
	}

	// Update menu hints
	t.updateHints()
}

// selectDemo switches to displaying the given demo.
func (t *Tutorial) selectDemo(demo demos.Demo) {
	if demo == t.currentDemo {
		return
	}

	t.currentDemo = demo
	t.demoComponent = nil
	t.demoContent.Clear()

	if demo == nil {
		t.showEmptyState()
		return
	}

	// Update panel title
	t.demoPanel.SetTitle(demo.Name() + " - " + demo.Description())

	// Add demo component (cache it)
	t.demoComponent = demo.Component()
	if t.demoComponent != nil {
		t.demoContent.AddItem(t.demoComponent, 0, 1, true)
	}

	// Update status bar
	t.statusBar.ClearSections()
	t.statusBar.AddSection(layout.StatusSection{
		Icon: theme.IconFolder,
		Text: demo.Category().String(),
	})
	t.statusBar.AddSection(layout.StatusSection{
		Icon:      theme.IconFile,
		Text:      demo.Name(),
		ColorFunc: theme.Accent,
	})
}

// showEmptyState displays a placeholder when no demo is selected.
func (t *Tutorial) showEmptyState() {
	empty := components.NewEmptyState().
		SetTitle("Select a Component").
		SetMessage("Choose a component from the sidebar to see its demo").
		SetIcon(theme.IconList)

	t.demoContent.Clear()
	t.demoContent.AddItem(empty, 0, 1, false)
	t.demoPanel.SetTitle("Demo")
}

// updateHints updates the menu with current key hints.
func (t *Tutorial) updateHints() {
	hints := []components.KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "Tab", Description: "Switch Pane"},
		{Key: "p", Description: "Properties"},
		{Key: "c", Description: "Code"},
		{Key: "t", Description: "Theme"},
		{Key: "?", Description: "Help"},
		{Key: "q", Description: "Quit"},
	}
	t.menu.SetHints(hints)
}

// StatusBar returns the status bar component.
func (t *Tutorial) StatusBar() *layout.StatusBar {
	return t.statusBar
}

// Menu returns the menu component.
func (t *Tutorial) Menu() *layout.Menu {
	return t.menu
}

// HandleGlobalInput handles global keyboard shortcuts.
func (t *Tutorial) HandleGlobalInput(event *tcell.EventKey) *tcell.EventKey {
	// Don't process global shortcuts when a modal is active
	// (except 'q' to quit which should always work)
	if t.app.Pages().CurrentIsModal() {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			t.app.Stop()
			return nil
		}
		return event
	}

	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			t.app.Stop()
			return nil
		case 'c':
			t.showCodeModal()
			return nil
		case 'p':
			t.showPropertyModal()
			return nil
		case 't':
			t.showThemePicker()
			return nil
		case '?':
			t.showHelpModal()
			return nil
		}
	}
	return event
}

// showCodeModal displays the code example for the current demo.
func (t *Tutorial) showCodeModal() {
	if t.currentDemo == nil {
		return
	}

	code := t.currentDemo.CodeExample()
	if code == "" {
		code = "// No code example available for this demo"
	}

	modal := NewCodeModal(t.currentDemo.Name(), code)
	t.app.Pages().Push(modal)
}

// showPropertyModal displays the property editor modal.
func (t *Tutorial) showPropertyModal() {
	if t.currentDemo == nil {
		return
	}

	modal := NewPropertyModal(t.currentDemo, func() {
		t.app.Pages().Pop()
	})
	t.app.Pages().Push(modal)
}

// showThemePicker displays the theme selection modal.
func (t *Tutorial) showThemePicker() {
	picker := NewThemePicker(func(provider theme.Theme) {
		theme.SetProvider(provider)
		t.app.Pages().Pop()
	})
	t.app.Pages().Push(picker)
}

// showHelpModal displays help information.
func (t *Tutorial) showHelpModal() {
	modal := NewHelpModal()
	t.app.Pages().Push(modal)
}

// Name implements nav.Component.
func (t *Tutorial) Name() string { return "tutorial" }

// Start implements nav.Component.
func (t *Tutorial) Start() {
	t.Split.FocusFirst()
}

// Stop implements nav.Component.
func (t *Tutorial) Stop() {}

// Hints implements nav.Component.
func (t *Tutorial) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "p", Description: "Properties"},
		{Key: "c", Description: "Code"},
		{Key: "t", Description: "Theme"},
		{Key: "q", Description: "Quit"},
	}
}

// Draw renders the tutorial view.
func (t *Tutorial) Draw(screen tcell.Screen) {
	t.Split.Draw(screen)
}

// InputHandler handles input for the tutorial.
func (t *Tutorial) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.Split.InputHandler()
}

// Focus handles focus delegation.
func (t *Tutorial) Focus(delegate func(tview.Primitive)) {
	t.Split.Focus(delegate)
}

// HasFocus returns whether the tutorial has focus.
func (t *Tutorial) HasFocus() bool {
	return t.Split.HasFocus()
}

// MouseHandler handles mouse events.
func (t *Tutorial) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return t.Split.MouseHandler()
}
