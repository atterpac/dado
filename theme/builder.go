package theme

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Color is a flexible color type that accepts multiple formats:
//   - tcell.Color: tcell.NewHexColor(0x282828)
//   - string hex: "#282828" or "282828"
//   - int32 hex: 0x282828
//
// Use the zero value to indicate "not set" (will be auto-derived).
type Color struct {
	value tcell.Color
	isSet bool
}

// C creates a Color from various input types.
// Accepts: tcell.Color, string (hex), int32/int (hex value), or nil.
//
// Examples:
//
//	C(tcell.NewHexColor(0x282828))
//	C("#282828")
//	C("282828")
//	C(0x282828)
//	C(nil) // not set, will use default
func C(v any) Color {
	if v == nil {
		return Color{}
	}

	switch val := v.(type) {
	case tcell.Color:
		return Color{value: val, isSet: true}
	case Color:
		return val
	case string:
		if val == "" {
			return Color{}
		}
		c, err := parseColor(val)
		if err != nil {
			return Color{} // Invalid string = not set
		}
		return Color{value: c, isSet: true}
	case int32:
		return Color{value: tcell.NewHexColor(val), isSet: true}
	case int:
		return Color{value: tcell.NewHexColor(int32(val)), isSet: true}
	default:
		return Color{}
	}
}

// parseColor parses a hex color string with or without # prefix.
func parseColor(s string) (tcell.Color, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return 0, fmt.Errorf("invalid hex color: %s", s)
	}
	val, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid hex color: %s", s)
	}
	return tcell.NewHexColor(int32(val)), nil
}

// ThemeColors defines all colors for a theme.
// Required fields: Bg, Fg, Accent, Success, Warning, Error, Info.
// Optional fields will be auto-derived if not set.
//
// Example:
//
//	theme.FromColors(theme.ThemeColors{
//	    Bg:      theme.C("#282828"),
//	    Fg:      theme.C("#ebdbb2"),
//	    Accent:  theme.C("#83a598"),
//	    Success: theme.C("#b8bb26"),
//	    Warning: theme.C("#fabd2f"),
//	    Error:   theme.C("#fb4934"),
//	    Info:    theme.C("#83a598"),
//	})
type ThemeColors struct {
	// Required - Base colors
	Bg Color // Background color (required)
	Fg Color // Foreground/text color (required)

	// Required - Accent colors
	Accent Color // Primary accent color (required)

	// Required - Semantic colors
	Success Color // Success/completed status (required)
	Warning Color // Warning status (required)
	Error   Color // Error/failed status (required)
	Info    Color // Info/running status (required)

	// Optional - Derived from base if not set
	BgLight   Color // Lighter background (default: lighten(Bg, 10%))
	BgDark    Color // Darker background (default: darken(Bg, 10%))
	FgDim     Color // Dimmed foreground (default: darken(Fg, 30%))
	FgMuted   Color // Muted foreground (default: darken(Fg, 40%))
	AccentDim Color // Dimmed accent (default: darken(Accent, 20%))
	Highlight Color // Highlight/selection (default: Accent)

	// Optional - Border colors (derived from base)
	Border      Color // Default border (default: darken(Bg, 20%))
	BorderFocus Color // Focused border (default: Accent)

	// Optional - UI element colors (derived from base/accent)
	Header      Color // Header background (default: BgDark)
	Menu        Color // Menu background (default: Bg)
	TableHeader Color // Table header text (default: Accent)
	Key         Color // Keyboard shortcut text (default: AccentDim)
	Crumb       Color // Breadcrumb text (default: Accent)
	PanelBorder Color // Panel border (default: Border)
	PanelTitle  Color // Panel title (default: Accent)
}

// FromColors creates a Theme from a ThemeColors struct.
// Required fields: Bg, Fg, Accent, Success, Warning, Error, Info.
// Missing optional fields are auto-derived with sensible defaults.
//
// Example:
//
//	myTheme := theme.FromColors(theme.ThemeColors{
//	    Bg:      theme.C("#282828"),
//	    Fg:      theme.C("#ebdbb2"),
//	    Accent:  theme.C("#83a598"),
//	    Success: theme.C("#b8bb26"),
//	    Warning: theme.C("#fabd2f"),
//	    Error:   theme.C("#fb4934"),
//	    Info:    theme.C("#83a598"),
//	})
//	theme.SetProvider(myTheme)
func FromColors(colors ThemeColors) (Theme, error) {
	// Validate required fields
	if !colors.Bg.isSet {
		return nil, fmt.Errorf("Bg is required")
	}
	if !colors.Fg.isSet {
		return nil, fmt.Errorf("Fg is required")
	}
	if !colors.Accent.isSet {
		return nil, fmt.Errorf("Accent is required")
	}
	if !colors.Success.isSet {
		return nil, fmt.Errorf("Success is required")
	}
	if !colors.Warning.isSet {
		return nil, fmt.Errorf("Warning is required")
	}
	if !colors.Error.isSet {
		return nil, fmt.Errorf("Error is required")
	}
	if !colors.Info.isSet {
		return nil, fmt.Errorf("Info is required")
	}

	t := &builtTheme{}

	// Set required colors
	t.bg = colors.Bg.value
	t.fg = colors.Fg.value
	t.accent = colors.Accent.value
	t.success = colors.Success.value
	t.warning = colors.Warning.value
	t.err = colors.Error.value
	t.info = colors.Info.value

	// Derive optional base colors
	t.bgLight = deriveColor(colors.BgLight, func() tcell.Color {
		return lightenColor(t.bg, 0.1)
	})
	t.bgDark = deriveColor(colors.BgDark, func() tcell.Color {
		return darkenColor(t.bg, 0.1)
	})
	t.fgDim = deriveColor(colors.FgDim, func() tcell.Color {
		return darkenColor(t.fg, 0.3)
	})
	t.fgMuted = deriveColor(colors.FgMuted, func() tcell.Color {
		return darkenColor(t.fg, 0.4)
	})
	t.accentDim = deriveColor(colors.AccentDim, func() tcell.Color {
		return darkenColor(t.accent, 0.2)
	})
	t.highlight = deriveColor(colors.Highlight, func() tcell.Color {
		return t.accent
	})

	// Derive border colors
	t.border = deriveColor(colors.Border, func() tcell.Color {
		return darkenColor(t.bg, 0.2)
	})
	t.borderFocus = deriveColor(colors.BorderFocus, func() tcell.Color {
		return t.accent
	})

	// Derive UI element colors
	t.header = deriveColor(colors.Header, func() tcell.Color {
		return t.bgDark
	})
	t.menu = deriveColor(colors.Menu, func() tcell.Color {
		return t.bg
	})
	t.tableHeader = deriveColor(colors.TableHeader, func() tcell.Color {
		return t.accent
	})
	t.key = deriveColor(colors.Key, func() tcell.Color {
		return t.accentDim
	})
	t.crumb = deriveColor(colors.Crumb, func() tcell.Color {
		return t.accent
	})
	t.panelBorder = deriveColor(colors.PanelBorder, func() tcell.Color {
		return t.border
	})
	t.panelTitle = deriveColor(colors.PanelTitle, func() tcell.Color {
		return t.accent
	})

	return t, nil
}

// MustFromColors is like FromColors but panics on error.
// Use this when you know the colors are valid.
func MustFromColors(colors ThemeColors) Theme {
	t, err := FromColors(colors)
	if err != nil {
		panic(fmt.Sprintf("theme.MustFromColors: %v", err))
	}
	return t
}

// deriveColor returns the color if set, otherwise calls the derive function.
func deriveColor(c Color, derive func() tcell.Color) tcell.Color {
	if c.isSet {
		return c.value
	}
	return derive()
}

// builtTheme implements Theme from builder/struct.
type builtTheme struct {
	// Base colors
	bg, bgLight, bgDark tcell.Color
	fg, fgDim, fgMuted  tcell.Color

	// Accent colors
	accent, accentDim, highlight tcell.Color

	// Semantic colors
	success, warning, err, info tcell.Color

	// Border colors
	border, borderFocus tcell.Color

	// UI element colors
	header, menu, tableHeader tcell.Color
	key, crumb                tcell.Color
	panelBorder, panelTitle   tcell.Color
}

// Theme interface implementation - Base colors
func (t *builtTheme) Bg() tcell.Color      { return t.bg }
func (t *builtTheme) BgLight() tcell.Color { return t.bgLight }
func (t *builtTheme) BgDark() tcell.Color  { return t.bgDark }
func (t *builtTheme) Fg() tcell.Color      { return t.fg }
func (t *builtTheme) FgDim() tcell.Color   { return t.fgDim }
func (t *builtTheme) FgMuted() tcell.Color { return t.fgMuted }

// Theme interface implementation - Accent colors
func (t *builtTheme) Accent() tcell.Color    { return t.accent }
func (t *builtTheme) AccentDim() tcell.Color { return t.accentDim }
func (t *builtTheme) Highlight() tcell.Color { return t.highlight }

// Theme interface implementation - Semantic colors
func (t *builtTheme) Success() tcell.Color { return t.success }
func (t *builtTheme) Warning() tcell.Color { return t.warning }
func (t *builtTheme) Error() tcell.Color   { return t.err }
func (t *builtTheme) Info() tcell.Color    { return t.info }

// Theme interface implementation - Border colors
func (t *builtTheme) Border() tcell.Color      { return t.border }
func (t *builtTheme) BorderFocus() tcell.Color { return t.borderFocus }

// Theme interface implementation - UI element colors
func (t *builtTheme) Header() tcell.Color      { return t.header }
func (t *builtTheme) Menu() tcell.Color        { return t.menu }
func (t *builtTheme) TableHeader() tcell.Color { return t.tableHeader }
func (t *builtTheme) Key() tcell.Color         { return t.key }
func (t *builtTheme) Crumb() tcell.Color       { return t.crumb }
func (t *builtTheme) PanelBorder() tcell.Color { return t.panelBorder }
func (t *builtTheme) PanelTitle() tcell.Color  { return t.panelTitle }

// Verify interface compliance at compile time
var _ Theme = (*builtTheme)(nil)

// Builder provides a fluent API for constructing themes.
//
// Example:
//
//	myTheme := theme.NewBuilder().
//	    Bg("#282828").
//	    Fg("#ebdbb2").
//	    Accent("#83a598").
//	    Success("#b8bb26").
//	    Warning("#fabd2f").
//	    Error("#fb4934").
//	    Info("#83a598").
//	    Build()
type Builder struct {
	colors ThemeColors
}

// NewBuilder creates a new theme builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Build creates the Theme from the builder configuration.
// Returns an error if required colors are missing.
func (b *Builder) Build() (Theme, error) {
	return FromColors(b.colors)
}

// MustBuild is like Build but panics on error.
func (b *Builder) MustBuild() Theme {
	return MustFromColors(b.colors)
}

// Bg sets the background color.
func (b *Builder) Bg(v any) *Builder {
	b.colors.Bg = C(v)
	return b
}

// BgLight sets the lighter background color.
func (b *Builder) BgLight(v any) *Builder {
	b.colors.BgLight = C(v)
	return b
}

// BgDark sets the darker background color.
func (b *Builder) BgDark(v any) *Builder {
	b.colors.BgDark = C(v)
	return b
}

// Fg sets the foreground/text color.
func (b *Builder) Fg(v any) *Builder {
	b.colors.Fg = C(v)
	return b
}

// FgDim sets the dimmed foreground color.
func (b *Builder) FgDim(v any) *Builder {
	b.colors.FgDim = C(v)
	return b
}

// FgMuted sets the muted foreground color.
func (b *Builder) FgMuted(v any) *Builder {
	b.colors.FgMuted = C(v)
	return b
}

// Accent sets the primary accent color.
func (b *Builder) Accent(v any) *Builder {
	b.colors.Accent = C(v)
	return b
}

// AccentDim sets the dimmed accent color.
func (b *Builder) AccentDim(v any) *Builder {
	b.colors.AccentDim = C(v)
	return b
}

// Highlight sets the highlight/selection color.
func (b *Builder) Highlight(v any) *Builder {
	b.colors.Highlight = C(v)
	return b
}

// Success sets the success/completed status color.
func (b *Builder) Success(v any) *Builder {
	b.colors.Success = C(v)
	return b
}

// Warning sets the warning status color.
func (b *Builder) Warning(v any) *Builder {
	b.colors.Warning = C(v)
	return b
}

// Error sets the error/failed status color.
func (b *Builder) Error(v any) *Builder {
	b.colors.Error = C(v)
	return b
}

// Info sets the info/running status color.
func (b *Builder) Info(v any) *Builder {
	b.colors.Info = C(v)
	return b
}

// Border sets the default border color.
func (b *Builder) Border(v any) *Builder {
	b.colors.Border = C(v)
	return b
}

// BorderFocus sets the focused border color.
func (b *Builder) BorderFocus(v any) *Builder {
	b.colors.BorderFocus = C(v)
	return b
}

// Header sets the header background color.
func (b *Builder) Header(v any) *Builder {
	b.colors.Header = C(v)
	return b
}

// Menu sets the menu background color.
func (b *Builder) Menu(v any) *Builder {
	b.colors.Menu = C(v)
	return b
}

// TableHeader sets the table header text color.
func (b *Builder) TableHeader(v any) *Builder {
	b.colors.TableHeader = C(v)
	return b
}

// Key sets the keyboard shortcut text color.
func (b *Builder) Key(v any) *Builder {
	b.colors.Key = C(v)
	return b
}

// Crumb sets the breadcrumb text color.
func (b *Builder) Crumb(v any) *Builder {
	b.colors.Crumb = C(v)
	return b
}

// PanelBorder sets the panel border color.
func (b *Builder) PanelBorder(v any) *Builder {
	b.colors.PanelBorder = C(v)
	return b
}

// PanelTitle sets the panel title color.
func (b *Builder) PanelTitle(v any) *Builder {
	b.colors.PanelTitle = C(v)
	return b
}

// FromTheme creates a builder pre-populated with colors from an existing theme.
// This is useful for creating theme variants by modifying only a few colors.
//
// Example:
//
//	darkVariant := theme.FromTheme(themes.TokyoNightNight).
//	    Bg("#0f0f14").
//	    BgDark("#000000").
//	    Build()
func FromTheme(t Theme) *Builder {
	if t == nil {
		return NewBuilder()
	}
	return &Builder{
		colors: ThemeColors{
			// Base colors
			Bg:      Color{value: t.Bg(), isSet: true},
			BgLight: Color{value: t.BgLight(), isSet: true},
			BgDark:  Color{value: t.BgDark(), isSet: true},
			Fg:      Color{value: t.Fg(), isSet: true},
			FgDim:   Color{value: t.FgDim(), isSet: true},
			FgMuted: Color{value: t.FgMuted(), isSet: true},

			// Accent colors
			Accent:    Color{value: t.Accent(), isSet: true},
			AccentDim: Color{value: t.AccentDim(), isSet: true},
			Highlight: Color{value: t.Highlight(), isSet: true},

			// Semantic colors
			Success: Color{value: t.Success(), isSet: true},
			Warning: Color{value: t.Warning(), isSet: true},
			Error:   Color{value: t.Error(), isSet: true},
			Info:    Color{value: t.Info(), isSet: true},

			// Border colors
			Border:      Color{value: t.Border(), isSet: true},
			BorderFocus: Color{value: t.BorderFocus(), isSet: true},

			// UI element colors
			Header:      Color{value: t.Header(), isSet: true},
			Menu:        Color{value: t.Menu(), isSet: true},
			TableHeader: Color{value: t.TableHeader(), isSet: true},
			Key:         Color{value: t.Key(), isSet: true},
			Crumb:       Color{value: t.Crumb(), isSet: true},
			PanelBorder: Color{value: t.PanelBorder(), isSet: true},
			PanelTitle:  Color{value: t.PanelTitle(), isSet: true},
		},
	}
}
