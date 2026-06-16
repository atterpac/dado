package inline

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI color codes. These are variables (not constants) so they can be
// blanked out when color is disabled (NO_COLOR, non-TTY stdout). Use
// DisableColor / EnableColor to toggle; ColorEnabled reports the state.
var (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
)

// colorVars points at every escape-code variable so they can be toggled
// as a group. Order matches the var block above.
var colorVars = []*string{
	&Reset, &Bold, &Dim, &Italic, &Underline,
	&Black, &Red, &Green, &Yellow, &Blue, &Magenta, &Cyan, &White,
	&BrightBlack, &BrightRed, &BrightGreen, &BrightYellow,
	&BrightBlue, &BrightMagenta, &BrightCyan, &BrightWhite,
	&BgBlue, &BgMagenta, &BgCyan,
}

// colorDefaults stores the original escape codes so EnableColor can restore.
var colorDefaults = func() []string {
	d := make([]string, len(colorVars))
	for i, p := range colorVars {
		d[i] = *p
	}
	return d
}()

var (
	colorEnabled = true
	// stdoutIsTTY reports whether color/animation are appropriate. It is
	// also used to gate spinners and in-place redraws.
	stdoutIsTTY = term.IsTerminal(int(os.Stdout.Fd()))
)

func init() {
	// Honor NO_COLOR (https://no-color.org) and disable color when stdout
	// is not a terminal (piped output, CI logs).
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor || !stdoutIsTTY {
		DisableColor()
	}
}

// ColorEnabled reports whether ANSI color output is currently on.
func ColorEnabled() bool { return colorEnabled }

// DisableColor blanks all escape codes so output is plain text.
func DisableColor() {
	colorEnabled = false
	for _, p := range colorVars {
		*p = ""
	}
}

// EnableColor restores ANSI escape codes.
func EnableColor() {
	colorEnabled = true
	for i, p := range colorVars {
		*p = colorDefaults[i]
	}
}

// ColorBg returns an ANSI background color from a hex code (empty if color
// is disabled).
func ColorBg(hex string) string {
	if !colorEnabled {
		return ""
	}
	r, g, b := HexToRGB(hex)
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// ColorFg returns an ANSI foreground color from a hex code (empty if color
// is disabled).
func ColorFg(hex string) string {
	if !colorEnabled {
		return ""
	}
	r, g, b := HexToRGB(hex)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// HexToRGB converts a hex color to RGB values.
func HexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 255, 255, 255
	}
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}
