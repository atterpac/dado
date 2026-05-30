package layout

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

func TestEnableThemesDefaults(t *testing.T) {
	// Synchronous theme application; avoids the auto-refresh goroutine that
	// would otherwise queue a draw against the non-running test app.
	theme.Default().SetAutoRefresh(false)
	defer theme.Default().SetAutoRefresh(true)

	app := NewApp(AppConfig{})
	app.EnableThemes(ThemeOptions{})

	if got := app.CurrentTheme(); got != themes.DefaultName {
		t.Fatalf("CurrentTheme = %q, want %q", got, themes.DefaultName)
	}
	if app.themeState.key != tcell.KeyCtrlT {
		t.Fatalf("default key = %v, want Ctrl+T", app.themeState.key)
	}
	if len(app.themeState.names) != len(themes.All()) {
		t.Fatalf("names = %d, want %d", len(app.themeState.names), len(themes.All()))
	}
	if theme.Get() != themes.All()[themes.DefaultName] {
		t.Fatal("default theme was not applied to provider")
	}
}

func TestEnableThemesExplicit(t *testing.T) {
	theme.Default().SetAutoRefresh(false)
	defer theme.Default().SetAutoRefresh(true)

	set := map[string]theme.Theme{
		"nord":    themes.Nord,
		"dracula": themes.Dracula,
	}
	var changed string
	app := NewApp(AppConfig{})
	app.EnableThemes(ThemeOptions{
		Key:      tcell.KeyCtrlP,
		Themes:   set,
		Default:  "nord",
		OnChange: func(name string) { changed = name },
	})

	if app.CurrentTheme() != "nord" {
		t.Fatalf("CurrentTheme = %q, want nord", app.CurrentTheme())
	}
	if app.themeState.key != tcell.KeyCtrlP {
		t.Fatal("explicit key not honored")
	}
	if theme.Get() != themes.Nord {
		t.Fatal("explicit default theme not applied")
	}
	_ = changed
}

func TestOpenThemeSelectorIdempotent(t *testing.T) {
	theme.Default().SetAutoRefresh(false)
	defer theme.Default().SetAutoRefresh(true)

	app := NewApp(AppConfig{})
	app.EnableThemes(ThemeOptions{Default: "nord"})

	app.openThemeSelector()
	if !app.themeState.open {
		t.Fatal("selector did not mark open")
	}
	depth := app.pages.StackDepth()

	// Second call while open must not push another page.
	app.openThemeSelector()
	if app.pages.StackDepth() != depth {
		t.Fatalf("re-open pushed a duplicate page: depth %d -> %d", depth, app.pages.StackDepth())
	}

	if _, ok := app.pages.Current().(*themeSelectorWrapper); !ok {
		t.Fatal("top page is not the theme selector")
	}
}
