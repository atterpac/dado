package themes

import "github.com/atterpac/jig/theme"

// All returns all built-in themes as a map.
func All() map[string]theme.Theme {
	return map[string]theme.Theme{
		// TokyoNight variants
		"tokyonight-night": TokyoNightNight,
		"tokyonight-storm": TokyoNightStorm,
		"tokyonight-moon":  TokyoNightMoon,
		"tokyonight-day":   TokyoNightDay,

		// Catppuccin variants
		"catppuccin-mocha":     CatppuccinMocha,
		"catppuccin-macchiato": CatppuccinMacchiato,
		"catppuccin-frappe":    CatppuccinFrappe,
		"catppuccin-latte":     CatppuccinLatte,

		// Dracula variants
		"dracula":       Dracula,
		"dracula-light": DraculaLight,

		// Gruvbox variants
		"gruvbox-dark":  GruvboxDark,
		"gruvbox-light": GruvboxLight,

		// One Dark/Light variants
		"onedark":  OneDark,
		"onelight": OneLight,

		// Solarized variants
		"solarized-dark":  SolarizedDark,
		"solarized-light": SolarizedLight,

		// Rosé Pine variants
		"rosepine":      RosePine,
		"rosepine-moon": RosePineMoon,
		"rosepine-dawn": RosePineDawn,

		// Kanagawa
		"kanagawa": Kanagawa,

		// Everforest variants
		"everforest-dark":  EverforestDark,
		"everforest-light": EverforestLight,

		// Monokai
		"monokai": Monokai,

		// GitHub variants
		"github-dark":  GitHubDark,
		"github-light": GitHubLight,

		// Nord
		"nord": Nord,
	}
}

// Names returns a sorted list of all built-in theme names.
func Names() []string {
	return []string{
		"catppuccin-frappe",
		"catppuccin-latte",
		"catppuccin-macchiato",
		"catppuccin-mocha",
		"dracula",
		"dracula-light",
		"everforest-dark",
		"everforest-light",
		"github-dark",
		"github-light",
		"gruvbox-dark",
		"gruvbox-light",
		"kanagawa",
		"monokai",
		"nord",
		"onedark",
		"onelight",
		"rosepine",
		"rosepine-dawn",
		"rosepine-moon",
		"solarized-dark",
		"solarized-light",
		"tokyonight-day",
		"tokyonight-moon",
		"tokyonight-night",
		"tokyonight-storm",
	}
}

// Get returns a theme by name, or nil if not found.
func Get(name string) theme.Theme {
	return All()[name]
}

// Default returns the default theme (TokyoNight Night).
func Default() theme.Theme {
	return TokyoNightNight
}

// DefaultName is the name of the default theme.
const DefaultName = "tokyonight-night"
