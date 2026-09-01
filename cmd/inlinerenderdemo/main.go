// Command inlinerenderdemo contains all renderer-native inline showcases.
//
// Run every showcase or select one with:
//
//	go run ./cmd/inlinerenderdemo -demo all
//	go run ./cmd/inlinerenderdemo -demo renderer
//	go run ./cmd/inlinerenderdemo -demo progress
//	go run ./cmd/inlinerenderdemo -demo spinner
//	go run ./cmd/inlinerenderdemo -demo multi
//	go run ./cmd/inlinerenderdemo -demo stepper
//	go run ./cmd/inlinerenderdemo -demo form
//	go run ./cmd/inlinerenderdemo -demo tree
//	go run ./cmd/inlinerenderdemo -demo table
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/inline"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"
)

type demoOptions struct {
	width     int
	delay     time.Duration
	terminal  bool
	automated bool
	theme     inline.InlineTheme
	indicator inline.MultiSelectIndicator
}

type showcase struct {
	name string
	run  func(context.Context, *inline.Renderer, demoOptions) error
}

func main() {
	selected := flag.String("demo", "all", "showcase: renderer, progress, spinner, multi, stepper, text, select, multiselect, form, tree, table, or all")
	requestedWidth := flag.Int("width", 64, "maximum demo width in terminal cells")
	delay := flag.Duration("delay", 35*time.Millisecond, "base animation delay")
	styleName := flag.String("style", "rounded", "visual preset: rounded, square, or ascii")
	indicatorName := flag.String("indicator", "theme", "multiselect marker: theme, check, dot, diamond, box, minimal, or none")
	flag.Parse()
	selectedTheme, themeOK := demoTheme(*styleName)
	if !themeOK {
		fmt.Fprintf(os.Stderr, "unknown style %q; choose rounded, square, or ascii\n", *styleName)
		os.Exit(2)
	}
	selectedIndicator, indicatorOK := demoIndicator(*indicatorName)
	if !indicatorOK {
		fmt.Fprintf(os.Stderr, "unknown indicator %q; choose theme, check, dot, diamond, box, minimal, or none\n", *indicatorName)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	available := []showcase{
		{name: "renderer", run: runRendererDemo},
		{name: "progress", run: runProgressDemo},
		{name: "spinner", run: runSpinnerDemo},
		{name: "multi", run: runMultiProgressDemo},
		{name: "stepper", run: runStepperDemo},
		{name: "text", run: runTextDemo},
		{name: "select", run: runSelectDemo},
		{name: "multiselect", run: runMultiSelectDemo},
		{name: "form", run: runFormDemo},
		{name: "tree", run: runTreeDemo},
		{name: "table", run: runTableDemo},
	}
	chosen := available
	if *selected != "all" {
		chosen = nil
		for _, demo := range available {
			if demo.name == *selected {
				chosen = append(chosen, demo)
			}
		}
		if len(chosen) == 0 {
			fmt.Fprintf(os.Stderr, "unknown demo %q; use -demo with renderer, progress, spinner, multi, stepper, text, select, multiselect, form, tree, table, or all\n", *selected)
			os.Exit(2)
		}
	}

	options := demoOptions{
		width:     outputWidth(max(*requestedWidth, 1)),
		delay:     max(*delay, time.Millisecond),
		terminal:  term.IsTerminal(int(os.Stdout.Fd())),
		automated: *selected == "all",
		theme:     selectedTheme,
		indicator: selectedIndicator,
	}
	for i, demo := range chosen {
		if i > 0 {
			fmt.Fprintln(os.Stdout)
		}
		if err := runShowcase(ctx, demo, options); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func runTreeDemo(_ context.Context, renderer *inline.Renderer, options demoOptions) error {
	tree := inline.NewTree("Release workspace",
		inline.TreeNode{Label: "cmd", Children: []inline.TreeNode{
			{Label: "dado", Detail: "CLI entrypoint"},
			{Label: "inlinerenderdemo", Detail: "showcase"},
		}},
		inline.TreeNode{Label: "inline", Children: []inline.TreeNode{
			{Label: "renderer.go", Detail: "normal-screen lifecycle"},
			{Label: "table.go", Detail: "renderer-native table"},
			{Label: "tree.go", Detail: "renderer-native hierarchy"},
		}},
		inline.TreeNode{Label: "go.mod", Detail: "module definition"},
	).SetTheme(options.theme)
	return renderer.Render(tree.Frame(options.width))
}

func runTableDemo(_ context.Context, renderer *inline.Renderer, options demoOptions) error {
	table := inline.NewTable("Release targets",
		inline.TableColumn{Header: "Target", MinWidth: 6},
		inline.TableColumn{Header: "Artifact"},
		inline.TableColumn{Header: "Size", Align: inline.AlignRight, MinWidth: 5},
	).SetTheme(options.theme).SetMaxRows(4).SetRows(
		[]string{"linux/amd64", "dado-linux", "8.4 MB"},
		[]string{"darwin/arm64", "dado-darwin", "8.1 MB"},
		[]string{"windows/amd64", "dado.exe", "8.7 MB"},
		[]string{"linux/arm64", "dado-linux-arm64", "8.2 MB"},
		[]string{"freebsd/amd64", "dado-freebsd", "8.3 MB"},
	)
	return renderer.Render(table.Frame(options.width))
}

func demoTheme(name string) (inline.InlineTheme, bool) {
	switch name {
	case "rounded":
		return inline.RoundedInlineTheme(), true
	case "square":
		return inline.SquareInlineTheme(), true
	case "ascii":
		return inline.ASCIIInlineTheme(), true
	default:
		return inline.InlineTheme{}, false
	}
}

func demoIndicator(name string) (inline.MultiSelectIndicator, bool) {
	switch name {
	case "theme":
		return inline.IndicatorTheme, true
	case "check":
		return inline.IndicatorCheck, true
	case "dot":
		return inline.IndicatorDot, true
	case "diamond":
		return inline.IndicatorDiamond, true
	case "box":
		return inline.IndicatorBox, true
	case "minimal":
		return inline.IndicatorMinimal, true
	case "none":
		return inline.IndicatorNone, true
	default:
		return inline.IndicatorTheme, false
	}
}

func runTextDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	form := inline.NewForm("Text input").SetTheme(options.theme).Add(inline.NewTextField("name", "Project name").Required().SetPlaceholder("myapp"))
	return runFormSession(ctx, renderer, options, form, "dado\r")
}

func runSelectDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	form := inline.NewForm("Select").SetTheme(options.theme).Add(inline.NewSelectField("channel", "Release channel",
		inline.NewChoice("stable", "Stable"), inline.NewChoice("preview", "Preview"), inline.NewChoice("nightly", "Nightly")).Required())
	return runFormSession(ctx, renderer, options, form, "\x1b[B\r")
}

func runMultiSelectDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	form := inline.NewForm("Multi-select").SetTheme(options.theme).Add(inline.NewMultiSelectField("targets", "Build targets",
		inline.NewChoice("linux", "Linux"), inline.NewChoice("darwin", "macOS"), inline.NewChoice("windows", "Windows")).MinSelected(1).SetIndicator(options.indicator))
	return runFormSession(ctx, renderer, options, form, " \x1b[B \r")
}

func runFormDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	stepper := inline.NewStepper(
		"",
		inline.WithStepOrientation(inline.StepHorizontal),
		inline.WithStepperTheme(options.theme.Status),
	)
	steps := []struct{ id, label string }{
		{"details", "Details"},
		{"configuration", "Configuration"},
		{"review", "Review"},
	}
	for _, step := range steps {
		if err := stepper.Add(step.id, step.label); err != nil {
			return err
		}
	}
	activeStep := -1
	form := inline.NewForm("Create release").
		SetTheme(options.theme).
		SetHeader(stepper).
		SetHeaderGap(1).
		Add(
			inline.NewTextField("name", "Release name").Required().SetPlaceholder("v1.0.0"),
			inline.NewSelectField("channel", "Channel", inline.NewChoice("stable", "Stable"), inline.NewChoice("preview", "Preview")).Required(),
			inline.NewMultiSelectField("targets", "Targets", inline.NewChoice("linux", "Linux"), inline.NewChoice("darwin", "macOS"), inline.NewChoice("windows", "Windows")).MinSelected(1).SetIndicator(options.indicator),
		).
		OnFocusChange(func(index int, _ inline.FormField) {
			if index < activeStep {
				for reset := len(steps) - 1; reset > index; reset-- {
					_ = stepper.Activate(steps[reset].id)
				}
			}
			if activeStep >= 0 && index == activeStep+1 {
				_ = stepper.Complete(steps[activeStep].id)
			}
			_ = stepper.Activate(steps[index].id)
			activeStep = index
		})
	return runFormSession(ctx, renderer, options, form, "v1.0.0\r\x1b[B\r \x1b[B \r")
}

func runFormSession(ctx context.Context, renderer *inline.Renderer, options demoOptions, form *inline.Form, scripted string) error {
	sessionOptions := []inline.SessionOption{inline.WithSessionWidth(options.width)}
	if options.automated || !options.terminal {
		sessionOptions = append(sessionOptions, inline.WithSessionInput(bytes.NewBufferString(scripted)))
	}
	_, err := inline.NewSession(renderer, sessionOptions...).Run(ctx, form)
	return err
}

func runShowcase(ctx context.Context, demo showcase, options demoOptions) error {
	renderer := inline.NewRenderer(inline.WithOutput(os.Stdout))
	err := demo.run(ctx, renderer, options)
	closeErr := renderer.Close()
	return errors.Join(err, closeErr)
}

func runRendererDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	progress := components.NewProgressBar().
		SetLabel("Compiling packages").
		SetShowPercentage(true)
	if !options.terminal {
		return renderer.Render(rendererSummaryFrame(options.width))
	}

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	logged := map[int]bool{}
	for percent := 0; percent <= 100; percent += 2 {
		if err := contextError(ctx); err != nil {
			return errors.Join(err, renderer.Println("Build interrupted"))
		}
		progress.SetProgress(float64(percent) / 100)
		frame, err := rendererDemoFrame(options.width, percent, frames[(percent/2)%len(frames)], progress)
		if err != nil {
			return err
		}
		if err := renderer.Render(frame); err != nil {
			return err
		}
		for _, milestone := range []int{25, 50, 75} {
			if percent >= milestone && !logged[milestone] {
				logged[milestone] = true
				if err := renderer.Printf("checkpoint: %d%% complete", milestone); err != nil {
					return err
				}
			}
		}
		if err := wait(ctx, options.delay); err != nil {
			return err
		}
	}
	if err := renderer.Println("compiled 18 packages"); err != nil {
		return err
	}
	return renderer.Render(rendererSummaryFrame(options.width))
}

func rendererDemoFrame(width, percent int, spinner string, progress *components.ProgressBar) (*inline.Frame, error) {
	const height = 7
	return inline.CaptureFrame(width, height, func(screen tcell.Screen) {
		title := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorFuchsia)
		muted := tcell.StyleDefault.Dim(true).Foreground(tcell.ColorSilver)
		active := tcell.StyleDefault.Foreground(tcell.ColorAqua)
		screen.PutStrStyled(0, 0, "dado inline renderer", title)
		screen.PutStrStyled(0, 1, "normal screen • scrollback preserved • tcell-backed cells", muted)
		screen.PutStrStyled(0, 3, spinner+" Building release artifacts", active)
		progress.SetRect(0, 4, width, 2)
		progress.Draw(screen)
		screen.PutStrStyled(0, 6, fmt.Sprintf("frame %02d  •  only changed rows are redrawn", percent/2+1), muted)
	})
}

func rendererSummaryFrame(width int) *inline.Frame {
	frame := inline.NewFrame(width, 3)
	frame.DrawString(0, 0, "✓ Build complete", tcell.StyleDefault.Bold(true).Foreground(tcell.ColorLime))
	frame.DrawString(0, 1, "The managed region shrank from 7 rows to 3.", tcell.StyleDefault)
	frame.DrawString(0, 2, "The final frame remains in normal terminal scrollback.", tcell.StyleDefault.Dim(true))
	return frame
}

func runProgressDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	progress := inline.NewProgress("Download release assets", 100, inline.WithProgressTheme(options.theme.Status))
	if !options.terminal {
		progress.SetDetail("42.8 MB downloaded")
		progress.Complete()
		return renderer.Render(progress.Frame(options.width))
	}
	for value := 0; value <= 100; value += 2 {
		progress.Set(int64(value))
		progress.SetDetail(fmt.Sprintf("%.1f of 42.8 MB", float64(value)*0.428))
		if err := renderer.Render(progress.Frame(options.width)); err != nil {
			return err
		}
		if err := wait(ctx, options.delay); err != nil {
			progress.Cancel("interrupted")
			return errors.Join(err, renderer.Render(progress.Frame(options.width)))
		}
	}
	progress.SetDetail("42.8 MB downloaded")
	return renderer.Render(progress.Frame(options.width))
}

func runSpinnerDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	spinner := inline.NewSpinner("Resolve dependency graph", inline.WithSpinnerTheme(options.theme.Status))
	if !options.terminal {
		spinner.Succeed("186 modules resolved")
		return renderer.Render(spinner.Frame(options.width))
	}
	spinner.Begin()
	for phase := 1; phase <= 24; phase++ {
		spinner.SetDetail(fmt.Sprintf("visiting module %d of 24", phase))
		if err := renderer.Render(spinner.Frame(options.width)); err != nil {
			return err
		}
		if err := wait(ctx, options.delay*2); err != nil {
			spinner.Cancel("interrupted")
			return errors.Join(err, renderer.Render(spinner.Frame(options.width)))
		}
	}
	spinner.Succeed("186 modules resolved")
	return renderer.Render(spinner.Frame(options.width))
}

type buildJob struct {
	id     string
	label  string
	factor int
	failAt int
	detail string
}

var buildJobs = []buildJob{
	{id: "modules", label: "Download modules", factor: 1},
	{id: "linux", label: "linux/amd64", factor: 2},
	{id: "darwin", label: "darwin/arm64", factor: 2},
	{id: "windows", label: "windows/amd64", factor: 2, failAt: 64, detail: "cross-linker unavailable"},
	{id: "lint", label: "Static analysis", factor: 1},
	{id: "unit", label: "Unit tests", factor: 1},
	{id: "integration", label: "Integration tests", factor: 3},
	{id: "docs", label: "Documentation", factor: 1},
}

func runMultiProgressDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	progress := inline.NewMultiProgress(
		"Building release targets",
		inline.WithMaxVisibleTasks(6),
		inline.WithCompletedTasksCollapsed(true),
		inline.WithMultiProgressTheme(options.theme.Status),
	)
	for _, job := range buildJobs {
		if err := progress.Add(job.id, job.label, 100); err != nil {
			return err
		}
	}
	if err := progress.Add("package", "Package artifacts", 0); err != nil {
		return err
	}
	if !options.terminal {
		finishBuildWithoutAnimation(progress)
		return renderer.Render(progress.Frame(options.width))
	}

	done := make(chan struct{})
	go runBuildJobs(ctx, progress, options.delay, done)
	ticker := time.NewTicker(max(options.delay*2, 20*time.Millisecond))
	defer ticker.Stop()
	for {
		if err := renderer.Render(progress.Frame(options.width)); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			cancelBuildTasks(progress)
			return errors.Join(ctx.Err(), renderer.Render(progress.Frame(options.width)))
		case <-done:
			return renderer.Render(progress.Frame(options.width))
		case <-ticker.C:
		}
	}
}

func runBuildJobs(ctx context.Context, progress *inline.MultiProgress, delay time.Duration, done chan<- struct{}) {
	defer close(done)
	var workers sync.WaitGroup
	for _, job := range buildJobs {
		job := job
		workers.Add(1)
		go func() {
			defer workers.Done()
			_ = progress.Start(job.id)
			for value := 4; value <= 100; value += 4 {
				if wait(ctx, delay*time.Duration(job.factor)) != nil {
					_ = progress.Cancel(job.id, "interrupted")
					return
				}
				if job.failAt > 0 && value >= job.failAt {
					_ = progress.Fail(job.id, errors.New(job.detail))
					return
				}
				_ = progress.Set(job.id, int64(value))
			}
		}()
	}
	workers.Wait()
	if ctx.Err() != nil {
		return
	}
	_ = progress.Start("package")
	_ = progress.SetDetail("package", "waiting on successful targets")
	if wait(ctx, delay*5) != nil {
		_ = progress.Cancel("package", "interrupted")
		return
	}
	_ = progress.Skip("package", "windows target failed")
}

func finishBuildWithoutAnimation(progress *inline.MultiProgress) {
	for _, job := range buildJobs {
		_ = progress.Start(job.id)
		if job.failAt > 0 {
			_ = progress.Set(job.id, int64(job.failAt))
			_ = progress.Fail(job.id, errors.New(job.detail))
			continue
		}
		_ = progress.Complete(job.id)
	}
	_ = progress.Skip("package", "one target failed")
}

func cancelBuildTasks(progress *inline.MultiProgress) {
	for _, task := range progress.Tasks() {
		if task.State == inline.TaskPending || task.State == inline.TaskRunning {
			_ = progress.Cancel(task.ID, "interrupted")
		}
	}
}

type workflowStep struct {
	id     string
	label  string
	detail string
}

var workflow = []workflowStep{
	{id: "prepare", label: "Prepare workspace", detail: "checking configuration"},
	{id: "build", label: "Build artifacts", detail: "compiling packages"},
	{id: "test", label: "Run verification", detail: "executing test suite"},
	{id: "approve", label: "Request approval", detail: "not required for preview"},
	{id: "deploy", label: "Deploy preview", detail: "uploading release"},
}

func runStepperDemo(ctx context.Context, renderer *inline.Renderer, options demoOptions) error {
	stepper := inline.NewStepper(
		"Preview deployment",
		inline.WithMaxVisibleSteps(4),
		inline.WithStepperTheme(options.theme.Status),
	)
	for _, step := range workflow {
		if err := stepper.Add(step.id, step.label); err != nil {
			return err
		}
	}
	if !options.terminal {
		finishWorkflowWithoutAnimation(stepper)
		return renderer.Render(stepper.Frame(options.width))
	}
	if err := stepper.Start(); err != nil {
		return err
	}
	for index, step := range workflow {
		if step.id == "approve" {
			if err := stepper.Skip(step.id, step.detail); err != nil {
				return err
			}
			if err := stepper.Activate("deploy"); err != nil {
				return err
			}
			continue
		}
		if err := animateStep(ctx, renderer, stepper, step, options); err != nil {
			return err
		}
		if step.id == "test" {
			if err := stepper.Fail(step.id, errors.New("flaky integration test")); err != nil {
				return err
			}
			if err := renderer.Render(stepper.Frame(options.width)); err != nil {
				return err
			}
			if err := renderer.Println("verification failed once; retrying"); err != nil {
				return err
			}
			if err := wait(ctx, options.delay*4); err != nil {
				return err
			}
			if err := stepper.Activate(step.id); err != nil {
				return err
			}
			if err := stepper.SetDetail(step.id, "retry passed"); err != nil {
				return err
			}
		}
		if err := stepper.Advance(); err != nil {
			return err
		}
		if index == len(workflow)-1 {
			break
		}
	}
	return renderer.Render(stepper.Frame(options.width))
}

func animateStep(ctx context.Context, renderer *inline.Renderer, stepper *inline.Stepper, step workflowStep, options demoOptions) error {
	for phase := 1; phase <= 3; phase++ {
		if err := stepper.SetDetail(step.id, fmt.Sprintf("%s (%d/3)", step.detail, phase)); err != nil {
			return err
		}
		if err := renderer.Render(stepper.Frame(options.width)); err != nil {
			return err
		}
		if err := wait(ctx, options.delay*3); err != nil {
			return err
		}
	}
	return nil
}

func finishWorkflowWithoutAnimation(stepper *inline.Stepper) {
	_ = stepper.Complete("prepare")
	_ = stepper.SetDetail("prepare", "workspace ready")
	_ = stepper.Complete("build")
	_ = stepper.SetDetail("build", "artifacts built")
	_ = stepper.Complete("test")
	_ = stepper.SetDetail("test", "verification passed on retry")
	_ = stepper.Skip("approve", "not required for preview")
	_ = stepper.Complete("deploy")
	_ = stepper.SetDetail("deploy", "preview deployed")
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func contextError(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func outputWidth(requested int) int {
	width := requested
	if terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = min(width, max(terminalWidth-1, 1))
	}
	return width
}
