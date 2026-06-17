// Command inlinedemo showcases dado's inline/CLI rendering helpers
// (github.com/atterpac/dado/inline) — ANSI printed straight to stdout, no
// alt-screen, no tcell. Contrast with the full-screen component demos
// in cmd/tutorial, cmd/graphdemo, etc.
//
//	go run ./cmd/inlinedemo
package main

import (
	"fmt"
	"time"

	"github.com/atterpac/dado/inline"
)

func main() {
	// Banner ----------------------------------------------------------
	inline.PrintVersion()

	// Interactive prompts gate the rest of the demo. On a non-TTY
	// (piped / CI) these return their defaults so the demo still runs
	// end-to-end without blocking.
	inline.PrintSection("SETUP")
	name := inline.Input("Project name", "myapp")
	module := inline.Input("Module path", "github.com/me/"+name)
	choice := inline.SelectOption("Template", []string{"minimal", "dashboard", "wizard"},
		[]string{"Single view", "Multi-pane layout", "Step-by-step flow"})
	templates := []string{"minimal", "dashboard", "wizard"}
	template := templates[choice]
	fmt.Println()
	if !inline.Confirm(fmt.Sprintf("Scaffold %q (%s template)?", name, template), true) {
		inline.PrintInfo("Aborted.")
		return
	}
	fmt.Println()

	// Status messages -------------------------------------------------
	inline.PrintSection("STATUS MESSAGES")
	inline.PrintSuccess("Build completed in 1.2s")
	inline.PrintInfo("Fetching dependencies")
	inline.PrintStep("Resolving module graph")
	inline.PrintStep("Downloading 4 packages")
	inline.PrintWarning("Using a pre-release toolchain")
	inline.PrintError("Lint found 2 issues")
	fmt.Println()

	// Spinner (animated, in-place) ------------------------------------
	inline.PrintSection("LONG-RUNNING STEP")
	sp := inline.NewSpinner("Running go mod tidy...")
	sp.Start()
	time.Sleep(1200 * time.Millisecond)
	sp.Stop("Dependencies resolved")
	fmt.Println()

	// Progress bar (in-place redraw) ----------------------------------
	inline.PrintSection("DOWNLOAD")
	for i := 0; i <= 20; i++ {
		inline.ProgressBar("packages", float64(i)/20, i == 20)
		time.Sleep(40 * time.Millisecond)
	}
	fmt.Println()

	// Multi-step status list (in-place redraw) ------------------------
	inline.PrintSection("PIPELINE")
	steps := inline.NewStatusList("Lint", "Test", "Build", "Package")
	for i := range 4 {
		steps.Set(i, inline.StepRunning)
		time.Sleep(300 * time.Millisecond)
		steps.Set(i, inline.StepDone)
	}
	fmt.Println()

	// Diff -----------------------------------------------------------
	inline.PrintSection("go.mod CHANGES")
	inline.PrintDiff(
		[]string{"module myapp", "go 1.21", "require tcell v2.7.3"},
		[]string{"module myapp", "go 1.22", "require tcell v2.7.4", "require x/term v0.18.0"},
	)
	fmt.Println()

	// File tree (scaffolder style) ------------------------------------
	inline.PrintSection("GENERATED FILES")
	inline.PrintTree([]inline.TreeNode{
		{Label: "main.go"},
		{Label: "go.mod"},
		{Label: "internal", Children: []inline.TreeNode{
			{Label: "app", Children: []inline.TreeNode{{Label: "app.go"}}},
			{Label: "views", Children: []inline.TreeNode{
				{Label: "home.go"},
				{Label: "about.go"},
			}},
		}},
	})
	fmt.Println()

	// Command help (aligned, no header chrome) ------------------------
	inline.PrintSection("COMMANDS")
	inline.PrintCommand("new", "<name>", "Scaffold a new dado app")
	inline.PrintCommand("component", "[section]", "Browse the component catalog")
	inline.PrintCommand("theme", "[name]", "Preview a color theme")
	inline.PrintCommand("version", "", "Print the version")
	fmt.Println()

	// Table — for genuinely tabular data ------------------------------
	inline.PrintSection("DEPENDENCIES")
	inline.PrintTable(
		[]string{"MODULE", "VERSION", "STATUS"},
		[][]string{
			{"github.com/gdamore/tcell/v2", "v2.7.4", "ok"},
			{"golang.org/x/term", "v0.18.0", "ok"},
			{"github.com/atterpac/dado", "v0.1.0", "local"},
		},
	)
	fmt.Println()

	// Key/value summary -----------------------------------------------
	inline.PrintSection("PROJECT")
	inline.PrintKV(
		[2]string{"name", name},
		[2]string{"module", module},
		[2]string{"template", template},
		[2]string{"dado", "v0.1.0"},
		[2]string{"docs", inline.Hyperlink("getgalaxy.io/dado", "https://getgalaxy.io/dado")},
	)
	fmt.Println()

	// Bulleted list ---------------------------------------------------
	inline.PrintSection("FEATURES")
	inline.PrintList(
		"Full-screen component framework",
		"Inline CLI rendering helpers",
		"Truecolor theming",
	)
	fmt.Println()

	// Bordered box ----------------------------------------------------
	inline.PrintBox("NEXT STEPS", []string{
		"cd myapp",
		"go mod tidy",
		"go run .",
	})
	fmt.Println()

	// Truecolor swatches ----------------------------------------------
	inline.PrintSection("THEME SWATCHES")
	swatches := []struct{ name, hex string }{
		{"primary", "#7c3aed"},
		{"success", "#22c55e"},
		{"warning", "#f59e0b"},
		{"danger", "#ef4444"},
		{"info", "#3b82f6"},
	}
	fmt.Print("    ")
	for _, s := range swatches {
		fmt.Printf("%s  %s ", inline.ColorBg(s.hex), inline.Reset)
	}
	fmt.Println()
	fmt.Print("    ")
	for _, s := range swatches {
		fmt.Printf("%s%-9s%s", inline.ColorFg(s.hex), s.name, inline.Reset)
	}
	fmt.Println()
	fmt.Println()
}
