package inline

import (
	"io"
	"os"

	"golang.org/x/term"
)

// out is the destination for all inline rendering. Defaults to stdout;
// override with SetOutput (e.g. to write to stderr or capture in tests).
var out io.Writer = os.Stdout

// Output returns the current output writer.
func Output() io.Writer { return out }

// SetOutput redirects all inline output to w and re-evaluates color and
// animation gating for the new target: color/redraws are enabled only when
// w is a terminal and NO_COLOR is unset.
func SetOutput(w io.Writer) {
	out = w
	stdoutIsTTY = fdIsTTY(w)
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor || !stdoutIsTTY {
		DisableColor()
	} else {
		EnableColor()
	}
}

// fdIsTTY reports whether w is a terminal-backed file.
func fdIsTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}
