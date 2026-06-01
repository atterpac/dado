package components_test

// Kitchen-sink usage examples. Each Example<Type> exercises the main options of
// a component and is compiled (and run) by `go test`, so the rendered "Usage"
// block on the docs site cannot drift from the API.

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func ExampleBarChart() {
	chart := components.NewBarChart().
		SetTitle("Requests by region").
		SetOrientation(components.BarVertical).
		SetItems(
			components.BarItem{Label: "us-east", Value: 1420, Color: tcell.ColorGreen},
			components.BarItem{Label: "eu-west", Value: 890},
			components.BarItem{Label: "ap-south", Value: 320},
		).
		SetShowValues(true).
		SetValueFormat("%.0f")

	fmt.Println(chart != nil)
	// Output: true
}

func ExampleLineGraph() {
	graph := components.NewLineGraph().
		SetTitle("CPU usage").
		SetStyle(components.LineGraphSolid).
		SetSeries(
			components.DataSeries{Label: "core0", Values: []float64{12, 18, 9, 22, 30}, Color: tcell.ColorGreen},
			components.DataSeries{Label: "core1", Values: []float64{5, 8, 14, 11, 19}},
		).
		SetRange(0, 100).
		SetShowLegend(true).
		SetYAxis(components.AxisConfig{Show: true, Format: "%.0f%%"})

	fmt.Println(graph != nil)
	// Output: true
}

func ExampleSparkline() {
	spark := components.NewSparkline().
		SetValues([]float64{12, 18, 15, 22, 19, 25, 30}).
		SetLabel("req/s")

	fmt.Println(spark != nil)
	// Output: true
}

func ExampleMetricCard() {
	card := components.NewMetricCard().
		SetLabel("Requests / sec").
		SetNumericValue(1420, "%.0f").
		SetUnit("rps").
		SetTrend(components.TrendUp, "+12%", true).
		SetSparkline([]float64{980, 1100, 1250, 1420}).
		SetShowSpark(true)

	fmt.Println(card != nil)
	// Output: true
}

func ExampleProgressBar() {
	bar := components.NewProgressBar().
		SetLabel("Uploading").
		SetProgress(0.75).
		SetShowPercentage(true)

	fmt.Println(bar != nil)
	// Output: true
}

func ExampleSpinner() {
	spinner := components.NewSpinner().
		SetStyle(components.SpinnerBraille).
		SetLabel("Loading").
		SetInterval(100 * time.Millisecond)

	fmt.Println(spinner != nil)
	// Output: true
}

func ExampleGauge() {
	gauge := components.NewGauge().
		SetLabel("CPU").
		SetValue(0.72).
		SetUnit("%")

	fmt.Println(gauge != nil)
	// Output: true
}

func ExampleBadge() {
	badge := components.NewBadge("42").
		SetVariant(components.BadgeSuccess).
		SetIcon("✓").
		SetPill(true)

	fmt.Println(badge != nil)
	// Output: true
}

func ExampleChip() {
	chip := components.NewChip("go").
		SetRemovable(true)

	fmt.Println(chip != nil)
	// Output: true
}

func ExampleDivider() {
	divider := components.NewDivider().
		SetLabel("Section").
		SetStyle('─')

	fmt.Println(divider != nil)
	// Output: true
}

func ExampleSkeleton() {
	skeleton := components.NewSkeleton().
		SetVariant(components.SkeletonText).
		SetLines(3).
		SetAnimated(true)

	fmt.Println(skeleton != nil)
	// Output: true
}

func ExampleButton() {
	btn := components.NewButton("Save").
		SetVariant(components.ButtonPrimary).
		SetDisabled(false)
	btn.SetOnClick(func() {
		fmt.Println("clicked")
	})

	// Click is also invoked by Enter or Space when focused.
	btn.Click()
	// Output: clicked
}

func ExampleLabel() {
	label := components.NewLabel("Ready").
		SetAlign(components.AlignLeft).
		SetColor(tcell.ColorGreen).
		SetBold(true).
		SetWordWrap(true)

	fmt.Println(label != nil)
	// Output: true
}

func ExampleTextArea() {
	area := components.NewTextArea("notes").
		SetLabel("Notes").
		SetPlaceholder("Write a description...").
		SetMaxLines(10).
		SetValue("first line\nsecond line")
	area.SetOnChange(func(e *components.ChangeEvent[string]) {
		fmt.Println("changed:", e.NewValue)
	})

	fmt.Println(area.Value())
	// Output: first line
	// second line
}

func ExampleRadioGroup() {
	radio := components.NewRadioGroup("plan").
		SetLabel("Plan").
		SetOptions([]string{"Free", "Pro", "Enterprise"}).
		SetSelected(1)
	radio.SetOnChange(func(e *components.ChangeEvent[string]) {
		fmt.Println("selected:", e.NewValue)
	})

	fmt.Println(radio.Value())
	// Output: Pro
}

func ExampleCheckbox() {
	check := components.NewCheckbox("notify").
		SetLabel("Enable notifications").
		SetChecked(true)
	check.SetOnChange(func(e *components.ChangeEvent[bool]) {
		fmt.Println("checked:", e.NewValue)
	})

	fmt.Println(check.Value())
	// Output: true
}

func ExampleMultiSelect() {
	ms := components.NewMultiSelect("tags").
		SetLabel("Tags").
		SetOptionsWithValues([]components.SelectOption{
			{Label: "Backend", Value: "be"},
			{Label: "Frontend", Value: "fe"},
			{Label: "Infra", Value: "infra"},
		})
	_ = ms.SetSelectedValues([]string{"be", "infra"})
	ms.SetOnChange(func(e *components.ChangeEvent[[]components.SelectOption]) {
		fmt.Println("count:", len(e.NewValue))
	})

	fmt.Println(len(ms.Values()))
	// Output: 2
}

func ExampleSearchBar() {
	bar := components.NewSearchBar().
		SetPlaceholder("Search pods...").
		SetIcon("").
		SetMaxResults(8)
	bar.SetOnChange(func(query string) {
		// Recompute results as the user types.
		bar.SetResults([]components.SearchResult{
			{Text: query + "-1"},
			{Text: query + "-2"},
		})
	})
	bar.SetOnSelect(func(r components.SearchResult) {
		fmt.Println("selected:", r.Text)
	})

	fmt.Println(bar != nil)
	// Output: true
}

func ExampleTabs() {
	tabs := components.NewTabs().
		AddTab("Overview", new(core.Box)).
		AddTabWithIcon("Logs", "", new(core.Box)).
		AddTab("Settings", new(core.Box)).
		SetShowIcons(true).
		SetClosable(true)
	tabs.SetBadge("Logs", 3)
	tabs.SetActiveByName("Logs")

	fmt.Println(tabs != nil)
	// Output: true
}

func ExampleSplit() {
	split := components.NewSplit().
		SetDirection(components.SplitHorizontal).
		SetFirst(new(core.Box)).
		SetSecond(new(core.Box)).
		SetRatio(0.3).
		SetResizable(true).
		SetShowDivider(true)

	fmt.Println(split != nil)
	// Output: true
}

func ExampleDrawer() {
	drawer := components.NewDrawer(components.DrawerConfig{
		Title:    "Details",
		Width:    50,
		Position: components.DrawerRight,
		Backdrop: true,
	}).
		SetContent(new(core.Box)).
		SetDismissOnEsc(true).
		SetHints([]components.KeyHint{
			{Key: "Esc", Description: "Close"},
		})

	fmt.Println(drawer != nil)
	// Output: true
}

func ExampleBottomSheet() {
	sheet := components.NewBottomSheet(components.BottomSheetConfig{
		Title:    "Quick Actions",
		Height:   12,
		Backdrop: true,
	}).
		SetContent(new(core.Box)).
		SetHints([]components.KeyHint{
			{Key: "Esc", Description: "Close"},
		})

	fmt.Println(sheet != nil)
	// Output: true
}

func ExampleContextMenu() {
	menu := components.NewContextMenu().
		SetItems([]components.MenuItem{
			{ID: "edit", Label: "Edit", Shortcut: "e", Handler: func() {}},
			{ID: "dup", Label: "Duplicate", Handler: func() {}},
			{ID: "del", Label: "Delete", Danger: true, Handler: func() {}},
		})
	menu.AddDivider()
	menu.AddSubmenu("export", "Export", []components.MenuItem{
		{ID: "csv", Label: "CSV", Handler: func() {}},
		{ID: "json", Label: "JSON", Handler: func() {}},
	})

	// Display at a cursor position or centered:
	//   menu.ShowAt(x, y)
	//   menu.ShowCentered(screenW, screenH)
	fmt.Println(menu != nil)
	// Output: true
}

func ExampleEmptyState() {
	empty := components.NewEmptyState().
		SetIcon("").
		SetTitle("No workflows").
		SetMessage("Create a workflow to get started")

	fmt.Println(empty != nil)
	// Output: true
}

func ExampleSplash() {
	splash := components.NewSplash().
		SetLogo("  DADO  ").
		SetStatus("Loading...").
		SetColors([]string{"#89b4fa", "#cba6f7"}).
		SetAutoDismiss(3 * time.Second).
		SetDismissKeys([]components.DismissKey{
			components.DismissEnter,
			components.DismissAnyKey,
		})

	fmt.Println(splash != nil)
	// Output: true
}

func ExampleMasterDetailView() {
	view := components.NewMasterDetailView().
		SetMasterTitle("Deployments").
		SetDetailTitle("Details").
		SetMasterContent(new(core.Box)).
		SetDetailContent(new(core.Box)).
		SetRatio(0.4).
		SetResizable(true)
	view.SetEmptyTitle("Nothing selected")

	fmt.Println(view != nil)
	// Output: true
}

func ExampleKeyHintBar() {
	bar := components.NewKeyHintBar().
		SetHints([]components.KeyHint{
			{Key: "Enter", Description: "Select"},
			{Key: "j/k", Description: "Navigate"},
			{Key: "Esc", Description: "Back"},
		})

	fmt.Println(bar != nil)
	// Output: true
}

func ExampleHintGrid() {
	grid := components.NewHintGrid().
		SetHints([]components.KeyHint{
			{Key: "j/k", Description: "Navigate"},
			{Key: "Enter", Description: "Select"},
			{Key: "Space", Description: "Toggle"},
			{Key: "Ctrl+S", Description: "Save"},
			{Key: "Esc", Description: "Close"},
		})

	fmt.Println(grid != nil)
	// Output: true
}

func ExampleTree() {
	root := &components.TreeNode{ID: "root", Label: "Project", Expanded: true}
	root.AddChild(&components.TreeNode{ID: "src", Label: "src/", Icon: ""})
	root.AddChild(&components.TreeNode{ID: "readme", Label: "README.md"})

	tree := components.NewTree().
		SetRoot(root).
		SetShowLines(true).
		SetShowIcons(true).
		SetIndentSize(2).
		SetMultiSelect(false)
	tree.SetOnSelect(func(n *components.TreeNode) {
		fmt.Println("selected:", n.Label)
	})

	fmt.Println(tree != nil)
	// Output: true
}

func ExampleList() {
	list := components.NewList().
		AddItem("First").
		AddItemWithSecondary("Second", "with detail").
		AddItems("Third", "Fourth").
		SetShowSecondary(true).
		SetWrapAround(true).
		SetHighlightFullLine(true)
	list.SetOnSelect(func(index int, item components.ListItem) {
		fmt.Println("selected:", item.Text)
	})

	fmt.Println(list != nil)
	// Output: true
}

func ExampleVirtualList() {
	items := make([]string, 10000)
	for i := range items {
		items[i] = fmt.Sprintf("row %d", i)
	}

	vl := components.NewVirtualList().
		SetTotalCount(len(items)).
		SetShowScrollbar(true).
		SetShowIndex(true).
		SetOverscan(4)
	// Only visible rows are fetched/rendered.
	vl.SetRenderFunc(func(index int, item components.VirtualListItem, width int, selected bool) string {
		return items[index]
	})

	fmt.Println(vl != nil)
	// Output: true
}

func ExampleLogViewer() {
	log := components.NewLogViewer().
		SetMaxEntries(5000).
		SetShowTimestamp(true).
		SetShowLevel(true).
		SetTimestampFormat("15:04:05").
		SetFollow(true).
		SetMinLevel(components.LogLevelInfo)

	// Append from any goroutine; the viewer redraws itself.
	log.AddEntry(components.LogEntry{Level: components.LogLevelInfo, Message: "server started"})
	log.AddEntry(components.LogEntry{Level: components.LogLevelError, Message: "connection lost"})

	fmt.Println(log != nil)
	// Output: true
}

func ExampleCodeView() {
	code := components.NewCodeView().
		SetCode("package main\n\nfunc main() {}\n").
		SetLanguage(components.LangGo).
		SetShowLineNumbers(true).
		SetTabWidth(4).
		SetHighlightLine(3)

	fmt.Println(code != nil)
	// Output: true
}

func ExampleAutocompleteInput() {
	suggestions := []components.Suggestion{
		{Text: "status", Category: "Field"},
		{Text: "service", Category: "Field"},
		{Text: "started", Category: "Value"},
	}

	input := components.NewAutocompleteInput().
		SetPrompt("> ").
		SetPlaceholder("Type to search...").
		SetMaxSuggestions(8).
		SetSuggestionProvider(components.FuzzyMatcher(suggestions))
	input.SetOnSubmit(func(text string) {
		fmt.Println("submitted:", text)
	})

	fmt.Println(input != nil)
	// Output: true
}

func ExampleDiffViewer() {
	dv := components.NewDiffViewer().
		SetTitle("main.go").
		SetDiff("line one\nline two\n", "line one\nline two changed\n").
		SetSideBySide(false).
		SetShowLineNumbers(true).
		SetWordDiff(true)

	fmt.Println(dv != nil)
	// Output: true
}

func ExampleTimeline() {
	start := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	mid := start.Add(2 * time.Hour)
	end := start.Add(4 * time.Hour)

	tl := components.NewTimeline().
		SetLanes([]components.TimelineLane{
			{ID: "build", Name: "Build", StartTime: start, EndTime: &mid},
			{ID: "test", Name: "Test", StartTime: mid, EndTime: &end},
		}).
		SetLabelWidth(12).
		SetShowLegend(true)
	tl.SetTimeRange(start, end)

	fmt.Println(tl != nil)
	// Output: true
}

func ExampleHeatMap() {
	hm := components.NewHeatMap().
		SetTitle("Activity").
		SetValues([][]float64{
			{0, 3, 8, 2},
			{5, 9, 1, 6},
		}).
		SetColLabels([]string{"Mon", "Tue", "Wed", "Thu"}).
		SetRowLabels([]string{"AM", "PM"}).
		SetColorScale(components.ColorScaleTheme).
		SetShowValues(true).
		SetValueFormat("%.0f")

	fmt.Println(hm != nil)
	// Output: true
}

func ExampleProgressModal() {
	modal := components.NewProgressModal().
		SetTitle("Deploying").
		SetMessage("Uploading image...").
		SetCancelable(true).
		SetProgress(0.4) // determinate; omit for an indeterminate spinner
	modal.SetOnCancel(func() {
		fmt.Println("cancelled")
	})

	fmt.Println(modal != nil)
	// Output: true
}

func ExampleForm() {
	form := components.NewForm().
		AddTextField("name", "Name", "Your name").
		AddSelect("role", "Role", []string{"Admin", "Editor", "Viewer"}).
		AddCheckbox("active", "Active")
	form.SetOnSubmit(func(values map[string]any) {
		fmt.Println("name:", values["name"])
	})

	fmt.Println(form != nil)
	// Output: true
}

func ExampleDataGrid() {
	source := components.NewSliceSource(
		[]components.GridColumn{
			{Name: "Name", MinWidth: 10},
			{Name: "Status", Width: 12},
			{Name: "Age", Align: components.AlignRight},
		},
		[][]components.GridCell{
			{{Value: "Alice"}, {Value: "Active"}, {Value: "32"}},
			{{Value: "Bob"}, {Value: "Inactive"}, {Value: "29"}},
		},
	)

	grid := components.NewDataGrid().
		SetSource(source).
		SetShowRowNumbers(true).
		SetShowHeader(true)
	grid.SetOnSubmit(func(cs *components.Changeset) {
		fmt.Println("pending changes:", cs.HasChanges())
	})

	fmt.Println(grid != nil)
	// Output: true
}

func ExampleToastManager() {
	toasts := components.NewToastManager()
	toasts.SetPosition(components.ToastTopRight)
	toasts.SetMaxVisible(5)

	toasts.Show("Saved successfully", components.ToastSuccess)
	toasts.ShowPersistent("Connection lost", components.ToastError)

	fmt.Println(toasts != nil)
	// Output: true
}

func ExampleNodeGraph() {
	data := components.NewNodeGraphData()
	data.AddNode(&components.GraphNode{ID: "api", Label: "API", Status: "running"})
	data.AddNode(&components.GraphNode{ID: "db", Label: "Database", Status: "running"})
	data.AddEdge(&components.GraphEdge{From: "api", To: "db", Label: "queries"})
	data.FocusID = "api"

	graph := components.NewNodeGraph().
		SetData(data).
		SetShowEdgeLabels(true).
		SetFit(true)

	fmt.Println(graph != nil)
	// Output: true
}

func ExampleERDGraph() {
	users := &components.ERDTable{
		ID:   "users",
		Name: "users",
		Columns: []components.ERDColumn{
			{Name: "id", Type: "bigint", IsPK: true},
			{Name: "email", Type: "varchar"},
		},
	}
	orders := &components.ERDTable{
		ID:   "orders",
		Name: "orders",
		Columns: []components.ERDColumn{
			{Name: "id", Type: "bigint", IsPK: true},
			{Name: "user_id", Type: "bigint", IsFK: true, FKTarget: "users.id"},
		},
	}

	erd := components.NewERDGraph().
		SetData(
			[]*components.ERDTable{users, orders},
			[]*components.ERDRelation{
				{FromTable: "orders", FromColumn: "user_id", ToTable: "users", ToColumn: "id"},
			},
		).
		SetFit(true)

	fmt.Println(erd != nil)
	// Output: true
}

func ExampleGitGraph() {
	data := components.NewGitGraphData()
	data.AddCommit(&components.GitCommit{
		Hash: "a1b2c3d", ShortHash: "a1b2c3d",
		Message: "Initial commit", Author: "atterpac",
	})
	data.AddCommit(&components.GitCommit{
		Hash: "e4f5g6h", ShortHash: "e4f5g6h",
		Message: "Add feature", Author: "atterpac",
		Parents: []string{"a1b2c3d"},
	})
	data.LayoutGraph()

	graph := components.NewGitGraph().
		SetGraph(data).
		SetShowRefs(true).
		SetShowAuthor(true)

	fmt.Println(graph != nil)
	// Output: true
}

func ExampleGraphTree() {
	data := components.NewGraphTreeData()
	data.AddNode(&components.GraphTreeNode{ID: "root", Label: "Workflow", Status: "Running", Children: []string{"step1"}, Expanded: true})
	data.AddNode(&components.GraphTreeNode{ID: "step1", Label: "Validate", Status: "Completed"})

	tree := components.NewGraphTree().
		SetData(data).
		SetShowEdgeLabels(true)

	fmt.Println(tree != nil)
	// Output: true
}

func ExampleFinder() {
	finder := components.NewFinder().
		SetPlaceholder("Search commands...").
		SetPrompt("> ").
		SetItems([]components.FinderItem{
			{ID: "deploy", Label: "Deploy", Description: "Ship to production", Category: "Actions"},
			{ID: "rollback", Label: "Rollback", Description: "Revert last deploy", Category: "Actions"},
		}).
		SetShowDescription(true).
		SetVimMode(false)
	finder.SetPreview(func(item components.FinderItem) string {
		return "Preview of " + item.Label
	})
	finder.SetOnSelect(func(item components.FinderItem) {
		fmt.Println("ran:", item.ID)
	})

	fmt.Println(finder != nil)
	// Output: true
}

func ExampleStatusBar() {
	bar := components.NewStatusBar().
		SetLeft(
			components.StatusSection{Icon: "", Text: "main"},
			components.StatusSection{Text: "12 pods"},
		).
		SetRight(
			components.StatusSection{Icon: "", Text: "Connected", Color: tcell.ColorGreen},
		).
		SetShowBorder(true)

	fmt.Println(bar != nil)
	// Output: true
}

// ExampleDebugOverlay constructs the bus-event inspector directly. In a real
// app you rarely do this — set AppConfig{Debug: true} and press Ctrl+D, which
// builds and wires the overlay (including SetOnClose to pop the page) for you.
func ExampleDebugOverlay() {
	overlay := components.NewDebugOverlay(0) // 0 → default capacity (500 events)
	overlay.SetOnClose(func() {
		fmt.Println("closed")
	})

	fmt.Println(overlay != nil)
	// Output: true
}
