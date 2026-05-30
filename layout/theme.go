package layout

import (
	"sort"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

// ThemeOptions configures App.EnableThemes.
type ThemeOptions struct {
	// Key opens the theme selector. Defaults to Ctrl+T when zero.
	Key tcell.Key

	// Themes is the name->theme set offered by the selector. Defaults to
	// the built-in set (themes.All()) when nil.
	Themes map[string]theme.Theme

	// Names is the display order in the selector. Defaults to the sorted
	// keys of Themes when nil.
	Names []string

	// Default is the theme applied immediately on EnableThemes. Falls back
	// to themes.DefaultName (if present) or the first name otherwise.
	Default string

	// OnChange, if set, is called with the theme name whenever the user
	// commits a selection. Use it to persist the choice across runs.
	OnChange func(name string)
}

// themeState holds the live theme-selector wiring for an App.
type themeState struct {
	key      tcell.Key
	byName   map[string]theme.Theme
	names    []string
	current  string
	onChange func(string)
	open     bool
}

// EnableThemes wires the built-in theme selector into the app: it applies a
// default theme and binds a key (default Ctrl+T) that opens a live-preview
// selector. Selecting a theme switches it app-wide via theme.Default(), which
// auto-refreshes every registered primitive and queues a redraw — downstream
// apps get theme switching with no per-component wiring.
//
//	app.EnableThemes(layout.ThemeOptions{Default: "nord"})
func (a *App) EnableThemes(opts ThemeOptions) *App {
	if opts.Themes == nil {
		opts.Themes = themes.All()
	}
	names := opts.Names
	if names == nil {
		names = sortedThemeNames(opts.Themes)
	}
	key := opts.Key
	if key == 0 {
		key = tcell.KeyCtrlT
	}

	def := opts.Default
	if _, ok := opts.Themes[def]; !ok {
		if _, ok := opts.Themes[themes.DefaultName]; ok {
			def = themes.DefaultName
		} else if len(names) > 0 {
			def = names[0]
		}
	}

	a.themeState = &themeState{
		key:      key,
		byName:   opts.Themes,
		names:    names,
		current:  def,
		onChange: opts.OnChange,
	}

	if t := opts.Themes[def]; t != nil {
		theme.Default().SetTheme(t)
	}
	return a
}

// CurrentTheme returns the name of the active theme, or "" if EnableThemes
// was never called.
func (a *App) CurrentTheme() string {
	if a.themeState == nil {
		return ""
	}
	return a.themeState.current
}

func sortedThemeNames(m map[string]theme.Theme) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// openThemeSelector pushes the live-preview theme selector onto the page stack.
func (a *App) openThemeSelector() {
	ts := a.themeState
	if ts == nil || ts.open {
		return
	}
	ts.open = true

	original := ts.current
	apply := func(name string) {
		if t := ts.byName[name]; t != nil {
			theme.Default().SetTheme(t)
		}
	}

	modal := theme.NewThemeSelectorModal(ts.names, ts.current)
	modal.SetOnPreview(apply)
	modal.SetOnSelect(func(name string) {
		ts.current = name
		ts.open = false
		apply(name)
		if ts.onChange != nil {
			ts.onChange(name)
		}
		a.pages.Pop()
	})
	modal.SetOnCancel(func() {
		ts.current = original
		ts.open = false
		apply(original)
		a.pages.Pop()
	})

	a.pages.Push(&themeSelectorWrapper{ThemeSelectorModal: modal})
}

// themeSelectorWrapper adapts theme.ThemeSelectorModal to nav.Component.
// The modal already implements tview.Primitive; this adds the lifecycle and
// hint methods nav.Pages requires.
type themeSelectorWrapper struct {
	*theme.ThemeSelectorModal
}

func (w *themeSelectorWrapper) Name() string { return "Theme" }
func (w *themeSelectorWrapper) Start()       {}
func (w *themeSelectorWrapper) Stop()        {}

func (w *themeSelectorWrapper) Hints() []components.KeyHint {
	return []components.KeyHint{
		{Key: "j/k", Description: "Navigate"},
		{Key: "Enter", Description: "Select"},
		{Key: "Esc", Description: "Cancel"},
	}
}
