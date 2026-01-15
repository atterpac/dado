// Package clipboard provides cross-platform clipboard access for TUI applications.
//
// Basic usage:
//
//	// Copy text
//	clipboard.Copy("Hello, World!")
//
//	// Paste text
//	text, err := clipboard.Paste()
//
//	// Check availability
//	if clipboard.Available() {
//	    // Clipboard is available
//	}
package clipboard

import (
	"bytes"
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	// ErrNotAvailable is returned when clipboard is not available
	ErrNotAvailable = errors.New("clipboard not available")

	// ErrUnsupportedPlatform is returned for unsupported platforms
	ErrUnsupportedPlatform = errors.New("unsupported platform")

	mu sync.Mutex

	// Cached availability check
	available     *bool
	availableMu   sync.Once
)

// Available returns true if clipboard operations are supported
func Available() bool {
	availableMu.Do(func() {
		result := checkAvailable()
		available = &result
	})
	return *available
}

func checkAvailable() bool {
	switch runtime.GOOS {
	case "darwin":
		_, err := exec.LookPath("pbcopy")
		return err == nil
	case "linux":
		// Check for common clipboard tools
		if _, err := exec.LookPath("xclip"); err == nil {
			return true
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			return true
		}
		if _, err := exec.LookPath("wl-copy"); err == nil {
			return true // Wayland
		}
		return false
	case "windows":
		// PowerShell is always available on modern Windows
		return true
	default:
		return false
	}
}

// Copy copies text to the system clipboard
func Copy(text string) error {
	mu.Lock()
	defer mu.Unlock()

	if !Available() {
		return ErrNotAvailable
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")

	case "linux":
		// Try clipboard tools in order of preference
		if path, err := exec.LookPath("wl-copy"); err == nil {
			// Wayland
			cmd = exec.Command(path)
		} else if path, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command(path, "-selection", "clipboard")
		} else if path, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command(path, "--clipboard", "--input")
		} else {
			return ErrNotAvailable
		}

	case "windows":
		// Use clip.exe which is available on all modern Windows
		cmd = exec.Command("clip.exe")

	default:
		return ErrUnsupportedPlatform
	}

	cmd.Stdin = strings.NewReader(text)

	return cmd.Run()
}

// Paste retrieves text from the system clipboard
func Paste() (string, error) {
	mu.Lock()
	defer mu.Unlock()

	if !Available() {
		return "", ErrNotAvailable
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbpaste")

	case "linux":
		// Try clipboard tools in order of preference
		if path, err := exec.LookPath("wl-paste"); err == nil {
			// Wayland
			cmd = exec.Command(path, "--no-newline")
		} else if path, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command(path, "-selection", "clipboard", "-o")
		} else if path, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command(path, "--clipboard", "--output")
		} else {
			return "", ErrNotAvailable
		}

	case "windows":
		// PowerShell to get clipboard
		cmd = exec.Command("powershell.exe", "-NoProfile", "-Command", "Get-Clipboard")

	default:
		return "", ErrUnsupportedPlatform
	}

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := out.String()

	// Windows adds a newline, trim it
	if runtime.GOOS == "windows" {
		result = strings.TrimSuffix(result, "\r\n")
		result = strings.TrimSuffix(result, "\n")
	}

	return result, nil
}

// CopyBytes copies binary data to clipboard (best effort, may not work on all platforms)
func CopyBytes(data []byte) error {
	return Copy(string(data))
}

// Clear clears the clipboard
func Clear() error {
	return Copy("")
}

// WriteAndPaste copies text and returns it (convenience for chains)
func WriteAndPaste(text string) (string, error) {
	if err := Copy(text); err != nil {
		return "", err
	}
	return text, nil
}
