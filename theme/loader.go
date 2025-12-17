package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ThemeConfig represents a theme loaded from a config file.
type ThemeConfig struct {
	Name   string       `json:"name"`
	Colors ColorConfig  `json:"colors"`
	Status []StatusDef  `json:"statuses,omitempty"`
}

// ColorConfig holds all theme color definitions.
type ColorConfig struct {
	// Base colors
	Bg      string `json:"bg"`
	BgLight string `json:"bg_light,omitempty"`
	BgDark  string `json:"bg_dark,omitempty"`
	Fg      string `json:"fg"`
	FgDim   string `json:"fg_dim"`
	FgMuted string `json:"fg_muted"`

	// Accent colors
	Accent    string `json:"accent"`
	AccentDim string `json:"accent_dim,omitempty"`
	Highlight string `json:"highlight"`

	// Semantic colors
	Success string `json:"success"`
	Warning string `json:"warning"`
	Error   string `json:"error"`
	Info    string `json:"info"`

	// Border colors
	Border      string `json:"border"`
	BorderFocus string `json:"border_focus"`

	// UI element colors (optional, will derive from base colors if not set)
	Header      string `json:"header,omitempty"`
	Menu        string `json:"menu,omitempty"`
	TableHeader string `json:"table_header,omitempty"`
	Key         string `json:"key,omitempty"`
	Crumb       string `json:"crumb,omitempty"`
	PanelBorder string `json:"panel_border,omitempty"`
	PanelTitle  string `json:"panel_title,omitempty"`
}

// StatusDef defines a status color and icon.
type StatusDef struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// ConfigTheme implements Theme from a config file.
type ConfigTheme struct {
	config ThemeConfig
	colors struct {
		// Base colors
		bg, bgLight, bgDark          tcell.Color
		fg, fgDim, fgMuted           tcell.Color

		// Accent colors
		accent, accentDim, highlight tcell.Color

		// Semantic colors
		success, warning, err, info  tcell.Color

		// Border colors
		border, borderFocus          tcell.Color

		// UI element colors
		header, menu, tableHeader    tcell.Color
		key, crumb                   tcell.Color
		panelBorder, panelTitle      tcell.Color
	}
}

// LoadFromFile loads a theme from a JSON or YAML file.
// Supported formats: .json, .yaml, .yml
func LoadFromFile(path string) (*ConfigTheme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return LoadFromJSON(data)
	case ".yaml", ".yml":
		return LoadFromYAML(data)
	default:
		return nil, fmt.Errorf("unsupported theme file format: %s", ext)
	}
}

// LoadFromJSON loads a theme from JSON data.
func LoadFromJSON(data []byte) (*ConfigTheme, error) {
	var config ThemeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return newConfigTheme(config)
}

// LoadFromYAML loads a theme from YAML data.
// Uses a simple YAML parser to avoid external dependencies.
func LoadFromYAML(data []byte) (*ConfigTheme, error) {
	config, err := parseSimpleYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return newConfigTheme(config)
}

// parseSimpleYAML parses a simple YAML theme config.
// This is a lightweight parser that handles the theme format without external deps.
func parseSimpleYAML(data []byte) (ThemeConfig, error) {
	var config ThemeConfig
	lines := strings.Split(string(data), "\n")

	var currentSection string
	var currentStatus *StatusDef

	for _, line := range lines {
		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Calculate indentation
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Check for section headers
		if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " ") {
			section := strings.TrimSuffix(trimmed, ":")
			if indent == 0 {
				currentSection = section
				currentStatus = nil
			} else if currentSection == "statuses" && indent >= 2 {
				// New status entry
				if currentStatus != nil {
					config.Status = append(config.Status, *currentStatus)
				}
				currentStatus = &StatusDef{Name: section}
			}
			continue
		}

		// Parse key: value
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		switch currentSection {
		case "":
			if key == "name" {
				config.Name = value
			}
		case "colors":
			switch key {
			// Base colors
			case "bg":
				config.Colors.Bg = value
			case "bg_light":
				config.Colors.BgLight = value
			case "bg_dark":
				config.Colors.BgDark = value
			case "fg":
				config.Colors.Fg = value
			case "fg_dim":
				config.Colors.FgDim = value
			case "fg_muted":
				config.Colors.FgMuted = value
			// Accent colors
			case "accent":
				config.Colors.Accent = value
			case "accent_dim":
				config.Colors.AccentDim = value
			case "highlight":
				config.Colors.Highlight = value
			// Semantic colors
			case "success":
				config.Colors.Success = value
			case "warning":
				config.Colors.Warning = value
			case "error":
				config.Colors.Error = value
			case "info":
				config.Colors.Info = value
			// Border colors
			case "border":
				config.Colors.Border = value
			case "border_focus":
				config.Colors.BorderFocus = value
			// UI element colors
			case "header":
				config.Colors.Header = value
			case "menu":
				config.Colors.Menu = value
			case "table_header":
				config.Colors.TableHeader = value
			case "key":
				config.Colors.Key = value
			case "crumb":
				config.Colors.Crumb = value
			case "panel_border":
				config.Colors.PanelBorder = value
			case "panel_title":
				config.Colors.PanelTitle = value
			}
		case "statuses":
			if currentStatus != nil {
				switch key {
				case "color":
					currentStatus.Color = value
				case "icon":
					currentStatus.Icon = value
				}
			}
		}
	}

	// Add last status if any
	if currentStatus != nil {
		config.Status = append(config.Status, *currentStatus)
	}

	return config, nil
}

func newConfigTheme(config ThemeConfig) (*ConfigTheme, error) {
	t := &ConfigTheme{config: config}

	var err error

	// Parse required base colors
	t.colors.bg, err = parseHexColor(config.Colors.Bg)
	if err != nil {
		return nil, fmt.Errorf("invalid bg color: %w", err)
	}
	t.colors.fg, err = parseHexColor(config.Colors.Fg)
	if err != nil {
		return nil, fmt.Errorf("invalid fg color: %w", err)
	}
	t.colors.fgDim, err = parseHexColor(config.Colors.FgDim)
	if err != nil {
		return nil, fmt.Errorf("invalid fg_dim color: %w", err)
	}
	t.colors.fgMuted, err = parseHexColor(config.Colors.FgMuted)
	if err != nil {
		return nil, fmt.Errorf("invalid fg_muted color: %w", err)
	}
	t.colors.accent, err = parseHexColor(config.Colors.Accent)
	if err != nil {
		return nil, fmt.Errorf("invalid accent color: %w", err)
	}
	t.colors.highlight, err = parseHexColor(config.Colors.Highlight)
	if err != nil {
		return nil, fmt.Errorf("invalid highlight color: %w", err)
	}
	t.colors.success, err = parseHexColor(config.Colors.Success)
	if err != nil {
		return nil, fmt.Errorf("invalid success color: %w", err)
	}
	t.colors.warning, err = parseHexColor(config.Colors.Warning)
	if err != nil {
		return nil, fmt.Errorf("invalid warning color: %w", err)
	}
	t.colors.err, err = parseHexColor(config.Colors.Error)
	if err != nil {
		return nil, fmt.Errorf("invalid error color: %w", err)
	}
	t.colors.info, err = parseHexColor(config.Colors.Info)
	if err != nil {
		return nil, fmt.Errorf("invalid info color: %w", err)
	}
	t.colors.border, err = parseHexColor(config.Colors.Border)
	if err != nil {
		return nil, fmt.Errorf("invalid border color: %w", err)
	}
	t.colors.borderFocus, err = parseHexColor(config.Colors.BorderFocus)
	if err != nil {
		return nil, fmt.Errorf("invalid border_focus color: %w", err)
	}

	// Parse optional colors with sensible defaults

	// BgLight - default to slightly lighter bg
	if config.Colors.BgLight != "" {
		t.colors.bgLight, err = parseHexColor(config.Colors.BgLight)
		if err != nil {
			return nil, fmt.Errorf("invalid bg_light color: %w", err)
		}
	} else {
		t.colors.bgLight = lightenColor(t.colors.bg, 0.1)
	}

	// BgDark - default to slightly darker bg
	if config.Colors.BgDark != "" {
		t.colors.bgDark, err = parseHexColor(config.Colors.BgDark)
		if err != nil {
			return nil, fmt.Errorf("invalid bg_dark color: %w", err)
		}
	} else {
		t.colors.bgDark = darkenColor(t.colors.bg, 0.1)
	}

	// AccentDim - default to dimmed accent
	if config.Colors.AccentDim != "" {
		t.colors.accentDim, err = parseHexColor(config.Colors.AccentDim)
		if err != nil {
			return nil, fmt.Errorf("invalid accent_dim color: %w", err)
		}
	} else {
		t.colors.accentDim = darkenColor(t.colors.accent, 0.2)
	}

	// UI element colors - default to base colors
	if config.Colors.Header != "" {
		t.colors.header, _ = parseHexColor(config.Colors.Header)
	} else {
		t.colors.header = t.colors.bgDark
	}

	if config.Colors.Menu != "" {
		t.colors.menu, _ = parseHexColor(config.Colors.Menu)
	} else {
		t.colors.menu = t.colors.bg
	}

	if config.Colors.TableHeader != "" {
		t.colors.tableHeader, _ = parseHexColor(config.Colors.TableHeader)
	} else {
		t.colors.tableHeader = t.colors.accent
	}

	if config.Colors.Key != "" {
		t.colors.key, _ = parseHexColor(config.Colors.Key)
	} else {
		t.colors.key = t.colors.accentDim
	}

	if config.Colors.Crumb != "" {
		t.colors.crumb, _ = parseHexColor(config.Colors.Crumb)
	} else {
		t.colors.crumb = t.colors.accent
	}

	if config.Colors.PanelBorder != "" {
		t.colors.panelBorder, _ = parseHexColor(config.Colors.PanelBorder)
	} else {
		t.colors.panelBorder = t.colors.border
	}

	if config.Colors.PanelTitle != "" {
		t.colors.panelTitle, _ = parseHexColor(config.Colors.PanelTitle)
	} else {
		t.colors.panelTitle = t.colors.accent
	}

	return t, nil
}

func parseHexColor(s string) (tcell.Color, error) {
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

// lightenColor returns a lighter version of the color
func lightenColor(c tcell.Color, factor float64) tcell.Color {
	r, g, b := c.RGB()
	r = int32(float64(r) + float64(255-r)*factor)
	g = int32(float64(g) + float64(255-g)*factor)
	b = int32(float64(b) + float64(255-b)*factor)
	return tcell.NewRGBColor(r, g, b)
}

// darkenColor returns a darker version of the color
func darkenColor(c tcell.Color, factor float64) tcell.Color {
	r, g, b := c.RGB()
	r = int32(float64(r) * (1 - factor))
	g = int32(float64(g) * (1 - factor))
	b = int32(float64(b) * (1 - factor))
	return tcell.NewRGBColor(r, g, b)
}

// Theme interface implementation - Base colors
func (t *ConfigTheme) Bg() tcell.Color      { return t.colors.bg }
func (t *ConfigTheme) BgLight() tcell.Color { return t.colors.bgLight }
func (t *ConfigTheme) BgDark() tcell.Color  { return t.colors.bgDark }
func (t *ConfigTheme) Fg() tcell.Color      { return t.colors.fg }
func (t *ConfigTheme) FgDim() tcell.Color   { return t.colors.fgDim }
func (t *ConfigTheme) FgMuted() tcell.Color { return t.colors.fgMuted }

// Theme interface implementation - Accent colors
func (t *ConfigTheme) Accent() tcell.Color    { return t.colors.accent }
func (t *ConfigTheme) AccentDim() tcell.Color { return t.colors.accentDim }
func (t *ConfigTheme) Highlight() tcell.Color { return t.colors.highlight }

// Theme interface implementation - Semantic colors
func (t *ConfigTheme) Success() tcell.Color { return t.colors.success }
func (t *ConfigTheme) Warning() tcell.Color { return t.colors.warning }
func (t *ConfigTheme) Error() tcell.Color   { return t.colors.err }
func (t *ConfigTheme) Info() tcell.Color    { return t.colors.info }

// Theme interface implementation - Border colors
func (t *ConfigTheme) Border() tcell.Color      { return t.colors.border }
func (t *ConfigTheme) BorderFocus() tcell.Color { return t.colors.borderFocus }

// Theme interface implementation - UI element colors
func (t *ConfigTheme) Header() tcell.Color      { return t.colors.header }
func (t *ConfigTheme) Menu() tcell.Color        { return t.colors.menu }
func (t *ConfigTheme) TableHeader() tcell.Color { return t.colors.tableHeader }
func (t *ConfigTheme) Key() tcell.Color         { return t.colors.key }
func (t *ConfigTheme) Crumb() tcell.Color       { return t.colors.crumb }
func (t *ConfigTheme) PanelBorder() tcell.Color { return t.colors.panelBorder }
func (t *ConfigTheme) PanelTitle() tcell.Color  { return t.colors.panelTitle }

// Name returns the theme name from config.
func (t *ConfigTheme) Name() string { return t.config.Name }

// RegisterStatuses registers all statuses defined in the theme config.
func (t *ConfigTheme) RegisterStatuses() error {
	for _, s := range t.config.Status {
		color, err := parseHexColor(s.Color)
		if err != nil {
			return fmt.Errorf("invalid status color for %s: %w", s.Name, err)
		}
		RegisterStatus(s.Name, StatusStyle{Color: color, Icon: s.Icon})
	}
	return nil
}
