package util

import (
	"bytes"
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// Clipboard errors.
var (
	ErrNoClipboard    = errors.New("clipboard not available")
	ErrClipboardWrite = errors.New("failed to write to clipboard")
	ErrClipboardRead  = errors.New("failed to read from clipboard")
)

// CopyToClipboard copies text to the system clipboard.
// Returns an error if clipboard is not available.
func CopyToClipboard(text string) error {
	switch runtime.GOOS {
	case "darwin":
		return copyWithCommand("pbcopy", nil, text)

	case "linux":
		// Try Wayland first
		if hasCommand("wl-copy") {
			return copyWithCommand("wl-copy", nil, text)
		}
		// Try xclip
		if hasCommand("xclip") {
			return copyWithCommand("xclip", []string{"-selection", "clipboard"}, text)
		}
		// Try xsel
		if hasCommand("xsel") {
			return copyWithCommand("xsel", []string{"--clipboard", "--input"}, text)
		}
		return ErrNoClipboard

	case "windows":
		return copyWithCommand("clip", nil, text)

	default:
		return ErrNoClipboard
	}
}

// ReadFromClipboard reads text from the system clipboard.
// Returns an error if clipboard is not available.
func ReadFromClipboard() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return readWithCommand("pbpaste", nil)

	case "linux":
		// Try Wayland first
		if hasCommand("wl-paste") {
			return readWithCommand("wl-paste", nil)
		}
		// Try xclip
		if hasCommand("xclip") {
			return readWithCommand("xclip", []string{"-selection", "clipboard", "-o"})
		}
		// Try xsel
		if hasCommand("xsel") {
			return readWithCommand("xsel", []string{"--clipboard", "--output"})
		}
		return "", ErrNoClipboard

	case "windows":
		return readWithCommand("powershell", []string{"-command", "Get-Clipboard"})

	default:
		return "", ErrNoClipboard
	}
}

// HasClipboard checks if clipboard functionality is available.
func HasClipboard() bool {
	switch runtime.GOOS {
	case "darwin":
		return hasCommand("pbcopy")
	case "linux":
		return hasCommand("wl-copy") || hasCommand("xclip") || hasCommand("xsel")
	case "windows":
		return hasCommand("clip")
	default:
		return false
	}
}

// copyWithCommand copies text using a command that reads from stdin.
func copyWithCommand(name string, args []string, text string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)

	if err := cmd.Run(); err != nil {
		return ErrClipboardWrite
	}
	return nil
}

// readWithCommand reads text using a command that writes to stdout.
func readWithCommand(name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", ErrClipboardRead
	}

	return strings.TrimSuffix(out.String(), "\n"), nil
}

// hasCommand checks if a command is available in PATH.
func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
