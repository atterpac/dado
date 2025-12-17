package scaffold

import (
	"os"
	"path/filepath"

	"github.com/atterpac/jig/cmd/jig/internal/ui"
)

// CreateSimpleProject creates a simple single-file project.
func CreateSimpleProject(name, themeName string) {
	// Create directory
	if err := os.MkdirAll(name, 0755); err != nil {
		ui.PrintError("Failed to create directory: " + err.Error())
		os.Exit(1)
	}

	// Create files
	writeFile(filepath.Join(name, "go.mod"), GoMod(name))
	writeFile(filepath.Join(name, "main.go"), SimpleMain(name, themeName))
	writeFile(filepath.Join(name, "README.md"), Readme(name))
}

// CreateStructuredProject creates a project with cmd/internal layout.
func CreateStructuredProject(name, themeName string) {
	// Create directories
	dirs := []string{
		filepath.Join(name, "cmd", name),
		filepath.Join(name, "internal", "views"),
		filepath.Join(name, "internal", "models"),
		filepath.Join(name, "internal", "actions"),
		filepath.Join(name, "internal", "config"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			ui.PrintError("Failed to create directory: " + err.Error())
			os.Exit(1)
		}
	}

	// Create files
	writeFile(filepath.Join(name, "go.mod"), GoMod(name))
	writeFile(filepath.Join(name, "cmd", name, "main.go"), StructuredMain(name, themeName))
	writeFile(filepath.Join(name, "internal", "views", "home.go"), HomeView(name))
	writeFile(filepath.Join(name, "internal", "actions", "registry.go"), Actions())
	writeFile(filepath.Join(name, "internal", "config", "config.go"), Config(name))
	writeFile(filepath.Join(name, "Taskfile.yml"), Taskfile(name))
	writeFile(filepath.Join(name, "README.md"), Readme(name))
}

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		ui.PrintError("Failed to write " + path + ": " + err.Error())
		os.Exit(1)
	}
	ui.PrintFileCreated(path)
}
