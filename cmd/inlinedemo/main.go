// Command inlinedemo showcases dado's static CLI helpers alongside its
// renderer-native activity components. It stays on the normal screen and
// preserves terminal scrollback.
//
//	go run ./cmd/inlinedemo
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/atterpac/dado/inline"
	"golang.org/x/term"
)

func main() {
	// Banner ----------------------------------------------------------
	inline.PrintVersion()

	// Interactive prompts gate the rest of the demo. On a non-TTY
	// (piped / CI) these return their defaults so the demo still runs
	// end-to-end without blocking.
	inline.PrintSection("SETUP")
	name, module, template, proceed, err := runSetupForm()
	if err != nil {
		panic(err)
	}
	if !proceed {
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

	// Renderer-native spinner -----------------------------------------
	inline.PrintSection("LONG-RUNNING STEP")
	if err := renderSpinner(); err != nil {
		panic(err)
	}
	fmt.Println()

	// Renderer-native progress ----------------------------------------
	inline.PrintSection("DOWNLOAD")
	if err := renderProgress(); err != nil {
		panic(err)
	}
	fmt.Println()

	// Renderer-native ordered workflow --------------------------------
	inline.PrintSection("PIPELINE")
	if err := renderStepper(); err != nil {
		panic(err)
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

func runSetupForm() (string, string, string, bool, error) {
	form := inline.NewForm("Setup").Add(
		inline.NewTextField("name", "Project name").Required().SetValue("myapp"),
		inline.NewTextField("module", "Module path").Required().SetValue("github.com/me/myapp"),
		inline.NewSelectField("template", "Template",
			inline.Choice{Value: "minimal", Label: "Minimal", Description: "Single view"},
			inline.Choice{Value: "dashboard", Label: "Dashboard", Description: "Multi-pane layout"},
			inline.Choice{Value: "wizard", Label: "Wizard", Description: "Step-by-step flow"},
		).Required(),
		inline.NewSelectField("action", "Continue?", inline.NewChoice("yes", "Proceed"), inline.NewChoice("no", "Cancel")).Required(),
	)
	renderer := inline.NewRenderer(inline.WithOutput(os.Stdout))
	options := []inline.SessionOption{inline.WithSessionWidth(60)}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		options = append(options, inline.WithSessionInput(bytes.NewBufferString("\r\r\r\r")))
	}
	result, err := inline.NewSession(renderer, options...).Run(context.Background(), form)
	closeErr := renderer.Close()
	if err != nil {
		return "", "", "", false, errors.Join(err, closeErr)
	}
	if closeErr != nil {
		return "", "", "", false, closeErr
	}
	return result["name"].(string), result["module"].(string), result["template"].(string), result["action"] == "yes", nil
}

func renderSpinner() error {
	renderer := inline.NewRenderer(inline.WithOutput(os.Stdout))
	spinner := inline.NewSpinner("Running go mod tidy")
	spinner.Begin()
	if term.IsTerminal(int(os.Stdout.Fd())) {
		for frame := range 15 {
			spinner.SetDetail(fmt.Sprintf("resolving dependency %d/15", frame+1))
			if err := renderer.Render(spinner.Frame(60)); err != nil {
				return err
			}
			time.Sleep(80 * time.Millisecond)
		}
	}
	spinner.Succeed("Dependencies resolved")
	if err := renderer.Render(spinner.Frame(60)); err != nil {
		return err
	}
	return renderer.Close()
}

func renderProgress() error {
	renderer := inline.NewRenderer(inline.WithOutput(os.Stdout))
	progress := inline.NewProgress("Download packages", 20)
	if term.IsTerminal(int(os.Stdout.Fd())) {
		for value := range 21 {
			progress.Set(int64(value))
			if err := renderer.Render(progress.Frame(60)); err != nil {
				return err
			}
			time.Sleep(40 * time.Millisecond)
		}
	} else {
		progress.Complete()
	}
	if err := renderer.Render(progress.Frame(60)); err != nil {
		return err
	}
	return renderer.Close()
}

func renderStepper() error {
	renderer := inline.NewRenderer(inline.WithOutput(os.Stdout))
	stepper := inline.NewStepper("Pipeline")
	for _, label := range []string{"Lint", "Test", "Build", "Package"} {
		if err := stepper.Add(label, label); err != nil {
			return err
		}
	}
	if err := stepper.Start(); err != nil {
		return err
	}
	for range 4 {
		if term.IsTerminal(int(os.Stdout.Fd())) {
			if err := renderer.Render(stepper.Frame(60)); err != nil {
				return err
			}
			time.Sleep(300 * time.Millisecond)
		}
		if err := stepper.Advance(); err != nil {
			return err
		}
	}
	if err := renderer.Render(stepper.Frame(60)); err != nil {
		return err
	}
	return renderer.Close()
}
