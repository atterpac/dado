package components

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/bus"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/snapshots"
	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

// imageCase mirrors snapCase but with a richer make() that includes sample data.
type imageCase struct {
	name string
	w, h int
	make func() core.Widget
}

// imageSnapshotCases defines all component image cases with realistic sample data.
// Where a component is identical to the draw snapshot case (already has data), the
// snapshot case is reused. Empty/viz components have dedicated populated variants.
func imageSnapshotCases() []imageCase {
	t0 := time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 3, 1, 9, 30, 0, 0, time.UTC)
	t2 := time.Date(2024, 3, 2, 12, 0, 0, 0, time.UTC)

	return []imageCase{
		// ── primitives / simple widgets ──────────────────────────────────────
		{"badge", 16, 1, func() core.Widget { return NewBadge("NEW").SetVariant(BadgeSuccess) }},
		{"chip", 18, 1, func() core.Widget { return NewChip("tag").SetRemovable(true) }},
		{"divider", 24, 1, func() core.Widget { return NewDivider().SetLabel("Section") }},
		{"button", 12, 1, func() core.Widget { return NewButton("OK") }},
		{"progress-bar", 30, 1, func() core.Widget {
			return NewProgressBar().SetProgress(0.42).SetLabel("Load").SetShowPercentage(true)
		}},
		{"gauge", 24, 15, func() core.Widget {
			gauges := []struct {
				label string
				value float64
			}{
				{"CPU", 0.32},
				{"Memory", 0.71},
				{"Disk", 0.88},
				{"Network", 0.45},
			}
			flex := core.NewFlex()
			for _, g := range gauges {
				flex.AddItem(NewGauge().SetValue(g.value).SetLabel(g.label).SetUnit("%"), 3, 0, false)
			}
			return flex
		}},
		{"sparkline", 44, 4, func() core.Widget {
			flex := core.NewFlex()
			flex.AddItem(NewSparkline().SetLabel("CPU %").SetValues(
				[]float64{12, 18, 15, 42, 38, 55, 61, 48, 35, 29, 44, 52, 67, 71, 58, 43, 38, 50, 62, 55, 40, 35, 28, 32, 45, 51, 63, 58, 44, 37, 30, 42, 55, 61, 49, 38, 44, 52, 60, 48},
			), 2, 0, false)
			flex.AddItem(NewSparkline().SetLabel("Memory %").SetValues(
				[]float64{55, 56, 57, 58, 60, 61, 63, 62, 64, 65, 64, 66, 67, 68, 67, 69, 70, 71, 72, 71, 73, 74, 75, 74, 76, 77, 78, 77, 79, 80, 79, 81, 82, 83, 82, 84, 85, 86, 85, 87},
			), 2, 0, false)
			return flex
		}},
		{"checkbox", 24, 5, func() core.Widget {
			cb := NewCheckbox("agree").SetLabel("Agree").SetChecked(true)
			flex := core.NewFlex()
			flex.AddItem(new(core.Box), 0, 1, false)
			flex.AddItem(cb, 1, 0, true)
			flex.AddItem(new(core.Box), 0, 1, false)
			return NewPanel().SetTitle("Preferences").SetContent(flex)
		}},
		{"radio-group", 24, 8, func() core.Widget {
			rg := NewRadioGroup("pick").SetLabel("Pick").
				SetOptions([]string{"One", "Two", "Three"}).SetSelected(1)
			flex := core.NewFlex()
			flex.AddItem(new(core.Box), 0, 1, false)
			flex.AddItem(rg, 4, 0, true)
			flex.AddItem(new(core.Box), 0, 1, false)
			return NewPanel().SetTitle("Options").SetContent(flex)
		}},
		{"metric-card", 26, 6, func() core.Widget {
			return NewMetricCard().SetLabel("Requests").SetValue("1.2k").SetTrend(TrendUp, "+5%", true)
		}},
		{"bar-chart", 32, 10, func() core.Widget {
			return NewBarChart().
				SetValues([]float64{3, 7, 5, 9, 4, 8}, []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}).
				SetShowValues(true)
		}},
		{"tabs", 30, 1, func() core.Widget {
			return NewTabs().AddTab("Overview", nil).AddTab("Logs", nil).AddTab("Metrics", nil).SetActive(0)
		}},
		{"context-menu", 36, 12, func() core.Widget {
			m := NewContextMenu().
				AddItemWithShortcut("copy", "Copy", "Ctrl+C", nil).
				AddItemWithShortcut("cut", "Cut", "Ctrl+X", nil).
				AddItemWithShortcut("paste", "Paste", "Ctrl+V", nil).
				AddDivider().
				AddItem("rename", "Rename", nil).
				AddItem("delete", "Delete", nil)
			m.SetDanger("delete", true)
			m.ShowCentered(36, 12)
			return m
		}},
		{"key-hint-bar", 40, 1, func() core.Widget {
			return NewKeyHintBar().SetHints([]KeyHint{
				{Key: "Enter", Description: "Select"},
				{Key: "Esc", Description: "Back"},
				{Key: "?", Description: "Help"},
			})
		}},
		{"hint-grid", 40, 3, func() core.Widget {
			return NewHintGrid().SetHints([]KeyHint{
				{Key: "j/k", Description: "Navigate"},
				{Key: "Enter", Description: "Select"},
				{Key: "/", Description: "Filter"},
				{Key: "r", Description: "Refresh"},
			})
		}},
		{"panel", 24, 6, func() core.Widget { return NewPanel().SetTitle("Details") }},
		{"select", 28, 8, func() core.Widget {
			return NewSelect("s").SetLabel("Region").
				SetOptions([]string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}).
				SetSelected(0).
				SetExpanded(true)
		}},
		{"multi-select", 32, 9, func() core.Widget {
			m := NewMultiSelect("m").SetLabel("Tags").
				SetOptions([]string{"production", "staging", "database", "backend", "frontend"})
			m.SetSelected([]int{0, 2}).SetExpanded(true)
			return m
		}},
		{"text-field", 30, 4, func() core.Widget {
			return NewTextField("t").SetLabel("Name").SetValue("John Doe")
		}},
		{"text-area", 30, 4, func() core.Widget {
			return NewTextArea("ta").SetLabel("Notes").SetValue("First line\nSecond line")
		}},
		{"autocomplete-input", 44, 12, func() core.Widget {
			ai := NewAutocompleteInput().
				SetText("badge").
				SetPlaceholder("Type to search...").
				SetSuggestionProvider(StaticSuggestions([]Suggestion{
					{Text: "badge", Description: "Status badge indicator", Category: "Widget"},
					{Text: "barchart", Description: "Vertical bar chart", Category: "Chart"},
					{Text: "button", Description: "Clickable button", Category: "Widget"},
					{Text: "checkbox", Description: "Toggle checkbox", Category: "Input"},
				}))
			return ai
		}},
		{"status-bar", 40, 1, func() core.Widget {
			return NewStatusBar().
				SetLeft(StatusSection{Text: "● Connected"}).
				SetRight(StatusSection{Text: "v2.1.0"})
		}},
		{"search-bar", 40, 8, func() core.Widget {
			return NewSearchBar().
				SetIcon("").
				SetPlaceholder("Search components...").
				SetQuery("badge").
				SetResults([]SearchResult{
					{Text: "badge", Description: "Status badge indicator"},
					{Text: "barchart", Description: "Vertical bar chart"},
					{Text: "button", Description: "Clickable button"},
				})
		}},
		{"progress-modal", 44, 12, func() core.Widget {
			return NewProgressModal().
				SetWidth(40).
				SetTitle("Deploying").
				SetMessage("Uploading container image...").
				SetSubMessage("Step 2 of 5").
				SetProgress(0.42).
				SetShowBackdrop(false)
		}},

		// ── data-rich / viz components ────────────────────────────────────────
		{"finder", 40, 10, func() core.Widget {
			f := NewFinder().SetPrompt("> ").SetPlaceholder("Search...")
			f.SetItems([]FinderItem{
				{ID: "1", Label: "badge", Description: "Small label indicator"},
				{ID: "2", Label: "button", Description: "Clickable action element"},
				{ID: "3", Label: "chart", Description: "Data visualization"},
				{ID: "4", Label: "dialog", Description: "Modal overlay"},
			})
			return f
		}},
		{"virtual-list", 30, 8, func() core.Widget {
			return NewVirtualList().SetItems([]VirtualListItem{
				{ID: "1", Data: "  api-gateway          Running"},
				{ID: "2", Data: "  auth-service         Running"},
				{ID: "3", Data: "  database             Running"},
				{ID: "4", Data: "  cache                Stopped"},
				{ID: "5", Data: "  worker               Running"},
				{ID: "6", Data: "  scheduler            Running"},
			})
		}},
		{"log-viewer", 40, 10, func() core.Widget {
			v := NewLogViewer()
			v.AddEntry(LogEntry{Timestamp: t0, Level: LogLevelInfo, Message: "Server started on :8080", Source: "main"})
			v.AddEntry(LogEntry{Timestamp: t0, Level: LogLevelDebug, Message: "Loading config from /etc/app.yaml", Source: "config"})
			v.AddEntry(LogEntry{Timestamp: t1, Level: LogLevelInfo, Message: "Connected to database", Source: "db"})
			v.AddEntry(LogEntry{Timestamp: t1, Level: LogLevelWarn, Message: "Slow query detected (523ms)", Source: "db"})
			v.AddEntry(LogEntry{Timestamp: t2, Level: LogLevelError, Message: "Failed to reach upstream service", Source: "http"})
			v.AddEntry(LogEntry{Timestamp: t2, Level: LogLevelInfo, Message: "Retrying in 5s", Source: "http"})
			return v
		}},
		{"diff-viewer", 40, 10, func() core.Widget {
			d := NewDiffViewer()
			d.SetDiff(
				"func greet(name string) string {\n\treturn \"Hello, \" + name\n}\n\nfunc main() {\n\tfmt.Println(greet(\"world\"))\n}",
				"func greet(name, title string) string {\n\treturn \"Hello, \" + title + \" \" + name\n}\n\nfunc main() {\n\tfmt.Println(greet(\"world\", \"Dr.\"))\n}",
			)
			return d
		}},
		{"data-grid", 40, 10, func() core.Widget {
			cols := []GridColumn{
				{Name: "Name", Width: 14},
				{Name: "Status", Width: 10},
				{Name: "CPU", Width: 6, Align: AlignRight},
				{Name: "Mem", Width: 6, Align: AlignRight},
			}
			rows := [][]GridCell{
				{{Value: "api-gateway"}, {Value: "Running"}, {Value: "12%"}, {Value: "128M"}},
				{{Value: "auth-service"}, {Value: "Running"}, {Value: "4%"}, {Value: "64M"}},
				{{Value: "database"}, {Value: "Running"}, {Value: "31%"}, {Value: "512M"}},
				{{Value: "cache"}, {Value: "Stopped"}, {Value: "0%"}, {Value: "0M"}},
				{{Value: "worker"}, {Value: "Running"}, {Value: "8%"}, {Value: "96M"}},
			}
			return NewDataGrid().SetSource(NewSliceSource(cols, rows))
		}},
		{"line-graph", 50, 16, func() core.Widget {
			return NewLineGraph().
				SetSeries(
					DataSeries{Label: "p50", Values: []float64{12, 18, 14, 22, 19, 28, 24, 31, 27, 35, 30, 38}},
					DataSeries{Label: "p99", Values: []float64{28, 34, 30, 42, 38, 51, 46, 58, 50, 65, 55, 70}},
				).
				SetTitle("Latency (ms)").
				SetShowLegend(true).
				SetShowGrid(true)
		}},
		{"heat-map", 30, 10, func() core.Widget {
			vals := [][]float64{
				{0.1, 0.3, 0.7, 0.9, 0.6},
				{0.4, 0.8, 1.0, 0.7, 0.3},
				{0.2, 0.5, 0.8, 0.4, 0.1},
				{0.6, 0.9, 0.6, 0.2, 0.5},
				{0.3, 0.4, 0.3, 0.8, 0.9},
			}
			return NewHeatMap().SetValues(vals).SetCellSize(6, 2).SetColorScale(ColorScaleTheme)
		}},
		{"tree", 30, 8, func() core.Widget {
			root := &TreeNode{Label: "src/", Expanded: true}
			components := &TreeNode{Label: "components/", Expanded: true}
			components.AddChild(&TreeNode{Label: "button.go"})
			components.AddChild(&TreeNode{Label: "table.go"})
			components.AddChild(&TreeNode{Label: "modal.go"})
			root.AddChild(components)
			root.AddChild(&TreeNode{Label: "main.go"})
			root.AddChild(&TreeNode{Label: "go.mod"})
			return NewTree().SetRoot(root).SetShowLines(true).ExpandAll()
		}},
		{"timeline", 56, 10, func() core.Widget {
			base := time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC)
			m := func(mins int) time.Time { return base.Add(time.Duration(mins) * time.Minute) }
			e := func(mins int) *time.Time { t := m(mins); return &t }
			return NewTimeline().SetLanes([]TimelineLane{
				{Name: "Checkout", StartTime: m(0), EndTime: e(2)},
				{Name: "Build",    StartTime: m(2), EndTime: e(18)},
				{Name: "Test",     StartTime: m(18), EndTime: e(34)},
				{Name: "Lint",     StartTime: m(18), EndTime: e(26)},
				{Name: "Deploy",   StartTime: m(34), EndTime: e(42)},
				{Name: "Verify",   StartTime: m(42)},
			})
		}},
		{"git-graph", 40, 10, func() core.Widget {
			data := NewGitGraphData()
			data.CurrentBranch = "main"
			data.AddCommit(&GitCommit{Hash: "a1b2c3d", ShortHash: "a1b2c3d", Message: "feat: add auth middleware", Author: "alice", Refs: []string{"HEAD", "main"}})
			data.AddCommit(&GitCommit{Hash: "e4f5a6b", ShortHash: "e4f5a6b", Message: "fix: correct token expiry", Author: "bob", Parents: []string{"a1b2c3d"}})
			data.AddCommit(&GitCommit{Hash: "c7d8e9f", ShortHash: "c7d8e9f", Message: "chore: update deps", Author: "alice", Branch: "feature/deps", Parents: []string{"a1b2c3d"}, Refs: []string{"feature/deps"}})
			data.AddCommit(&GitCommit{Hash: "f0a1b2c", ShortHash: "f0a1b2c", Message: "Merge branch 'feature/deps'", Author: "alice", Parents: []string{"e4f5a6b", "c7d8e9f"}, IsMerge: true})
			data.LayoutGraph()
			return NewGitGraph().SetGraph(data).SetShowAuthor(true).SetShowHash(true)
		}},
		{"node-graph", 40, 15, func() core.Widget {
			data := NewNodeGraphData()
			data.AddNode(&GraphNode{ID: "api", Label: "API Gateway", Status: "running", Focused: true})
			data.AddNode(&GraphNode{ID: "auth", Label: "Auth Service", Status: "running"})
			data.AddNode(&GraphNode{ID: "db", Label: "Database", Status: "running"})
			data.AddNode(&GraphNode{ID: "cache", Label: "Cache", Status: "degraded"})
			data.AddEdge(&GraphEdge{From: "api", To: "auth", Label: "JWT"})
			data.AddEdge(&GraphEdge{From: "api", To: "db"})
			data.AddEdge(&GraphEdge{From: "auth", To: "cache"})
			return NewNodeGraph().SetFit(true).SetData(data)
		}},
		{"graph-tree", 40, 12, func() core.Widget {
			data := NewGraphTreeData()
			data.RootID = "deploy"
			data.AddNode(&GraphTreeNode{ID: "deploy", Label: "Deploy Pipeline", Sublabel: "main", Status: "Running", NodeType: GraphNodePrimary, Expanded: true, Children: []string{"build", "test"}})
			data.AddNode(&GraphTreeNode{ID: "build", Label: "Build", Status: "Completed", NodeType: GraphNodeSecondary, Expanded: true, Children: []string{"lint"}})
			data.AddNode(&GraphTreeNode{ID: "test", Label: "Test Suite", Status: "Running", NodeType: GraphNodeSecondary})
			data.AddNode(&GraphTreeNode{ID: "lint", Label: "Lint", Status: "Completed", NodeType: GraphNodeSecondary})
			data.AddEdge(&GraphTreeEdge{From: "deploy", To: "build"})
			data.AddEdge(&GraphTreeEdge{From: "deploy", To: "test"})
			data.AddEdge(&GraphTreeEdge{From: "build", To: "lint"})
			return NewGraphTree().SetData(data)
		}},
		{"e-r-d-graph", 50, 13, func() core.Widget {
			users := &ERDTable{ID: "users", Name: "users", Columns: []ERDColumn{
				{Name: "id", Type: "uuid", IsPK: true},
				{Name: "email", Type: "text"},
				{Name: "name", Type: "text"},
			}}
			posts := &ERDTable{ID: "posts", Name: "posts", Columns: []ERDColumn{
				{Name: "id", Type: "uuid", IsPK: true},
				{Name: "user_id", Type: "uuid", IsFK: true, FKTarget: "users.id"},
				{Name: "title", Type: "text"},
			}}
			rel := &ERDRelation{FromTable: "users", ToTable: "posts", Cardinality: OneToMany}
			return NewERDGraph().SetNodeWidth(16).SetSpacing(4, 1).SetFit(true).SetData([]*ERDTable{users, posts}, []*ERDRelation{rel})
		}},
		// ── missing-coverage components ──────────────────────────────────────
		{"label", 20, 1, func() core.Widget {
			return NewLabel("Hello, dado!")
		}},
		{"spinner", 12, 3, func() core.Widget {
			return NewSpinner().SetLabel("Loading")
		}},
		{"skeleton", 30, 4, func() core.Widget {
			return NewSkeleton().SetLines(3)
		}},
		{"empty-state", 30, 8, func() core.Widget {
			return NewEmptyState().Configure("⊘", "No Results", "Try adjusting your filters")
		}},
		{"list", 28, 8, func() core.Widget {
			return NewList().
				AddItemWithSecondary("nginx", "port 80").
				AddItemWithSecondary("postgres", "port 5432").
				AddItemWithSecondary("redis", "port 6379").
				AddItemWithSecondary("rabbitmq", "port 5672")
		}},
		{"table", 40, 8, func() core.Widget {
			return NewTable().
				SetHeaders("Service", "Status", "Uptime").
				AddRow("nginx", "running", "14d").
				AddRow("postgres", "running", "14d").
				AddRow("redis", "running", "3d").
				AddRow("cache", "stopped", "—")
		}},
		{"modal", 44, 14, func() core.Widget {
			m := NewModal(ModalConfig{Title: "Confirm Delete", Width: 38, Height: 8})
			inner := core.NewFlex()
			inner.AddItem(core.NewTextView().SetText("Are you sure you want to delete\nthis resource? This cannot be undone."), 0, 1, false)
			m.SetContent(inner)
			return m
		}},
		{"form", 40, 14, func() core.Widget {
			return NewForm().
				AddTextField("name", "Name", "John Doe").
				AddTextField("email", "Email", "user@example.com").
				AddSelect("role", "Role", []string{"Admin", "Editor", "Viewer"}).
				AddCheckbox("notify", "Email notifications")
		}},
		{"form-builder", 40, 12, func() core.Widget {
			return NewFormBuilder().
				Text("host", "Host").Placeholder("localhost").Done().
				Text("port", "Port").Placeholder("5432").Done().
				Select("driver", "Driver", []string{"postgres", "mysql", "sqlite"}).Done().
				Build()
		}},
		{"bottom-sheet", 44, 14, func() core.Widget {
			bs := NewBottomSheet(BottomSheetConfig{Title: "Filter", Height: 10})
			inner := core.NewFlex()
			inner.AddItem(core.NewTextView().SetText("Status: Running"), 1, 0, false)
			inner.AddItem(core.NewTextView().SetText("Region: us-east-1"), 1, 0, false)
			bs.SetContent(inner)
			return bs
		}},
		{"drawer", 44, 12, func() core.Widget {
			d := NewDrawer(DrawerConfig{Title: "Details", Width: 20})
			inner := core.NewFlex()
			inner.AddItem(core.NewTextView().SetText("Name: api-gateway\nStatus: Running\nCPU: 12%\nMem: 128M"), 0, 1, false)
			d.SetContent(inner)
			return d
		}},
		{"split", 44, 12, func() core.Widget {
			left := NewList().AddItems("nginx", "postgres", "redis", "cache", "worker")
			right := core.NewTextView().SetText("Service: nginx\nStatus: Running\nUptime: 14 days")
			return NewSplit().
				SetDirection(SplitHorizontal).
				SetRatio(0.35).
				SetFirst(left).
				SetSecond(right)
		}},
		{"master-detail-view", 50, 14, func() core.Widget {
			list := NewList().
				AddItemWithSecondary("api-gateway", "Running").
				AddItemWithSecondary("auth-service", "Running").
				AddItemWithSecondary("database", "Running")
			detail := core.NewTextView().SetText("Service: api-gateway\nStatus: Running\nCPU: 12%\nMemory: 128M\nUptime: 14 days")
			return NewMasterDetailView().
				SetMasterTitle("Services").
				SetDetailTitle("Details").
				SetMasterContent(list).
				SetDetailContent(detail).
				SetDetailVisible(true).
				SetRatio(0.35)
		}},
		{"splash", 44, 14, func() core.Widget {
			return NewSplash().
				SetLogo("dado").
				SetStatus("Loading...").
				Build()
		}},
		{"toast-manager", 44, 14, func() core.Widget {
			// ToastManager requires a live app; render a panel placeholder instead.
			return NewPanel().SetTitle("Toast Notifications").
				SetContent(core.NewTextView().SetText(
					"  ✓ Deployment successful\n  ✓ Config reloaded\n  ⚠ Cache miss rate high",
				))
		}},
		{"debug-overlay", 44, 14, func() core.Widget {
			// Seed the bus with sample events so the overlay has data to display.
			bus.SetEnabled(true)
			for _, e := range []bus.Event{
				{Source: bus.SourceNav, Payload: "push: dashboard"},
				{Source: bus.SourceBinding, Payload: "key: ctrl+k"},
				{Source: bus.SourceTheme, Payload: "set: nord"},
				{Source: bus.SourceAsync, Payload: "task: fetch-data started"},
				{Source: bus.SourceInput, Payload: "focus: data-grid"},
				{Source: bus.SourceEffect, Payload: "render: badge"},
			} {
				bus.Publish(e)
			}
			d := NewDebugOverlay(50)
			d.Start()
			return d
		}},

		{"code-view", 40, 10, func() core.Widget {
			return NewCodeView().SetLanguage(LangGo).SetShowLineNumbers(true).SetCode(
				`func NewServer(cfg *Config) *Server {
    return &Server{
        addr:    cfg.Addr,
        timeout: cfg.Timeout,
        router:  http.NewServeMux(),
    }
}`,
			)
		}},
	}
}

// TestGenerateComponentImages renders every component to a standardised PNG
// (snapshots.CanvasW × snapshots.CanvasH cells) with the component centered and
// the theme background filling the surrounding padding.
//
// Runs once per built-in theme, writing to <outDir>/<theme>/<component>.png.
// Only runs when GENERATE_IMAGES=1. Output dir defaults to ../docs/images/components.
//
//	GENERATE_IMAGES=1 go test ./components/ -run TestGenerateComponentImages
//	GENERATE_IMAGES=1 IMAGE_OUT=/abs/path go test ./components/ -run TestGenerateComponentImages
func TestGenerateComponentImages(t *testing.T) {
	if os.Getenv("GENERATE_IMAGES") == "" {
		t.Skip("set GENERATE_IMAGES=1 to run")
	}

	outDir := os.Getenv("IMAGE_OUT")
	if outDir == "" {
		outDir = filepath.Join("..", "docs", "images", "components")
	}

	theme.Default().SetAutoRefresh(false)
	defer theme.Default().SetAutoRefresh(true)

	allThemes := themes.All()
	// Sort for deterministic output order.
	themeNames := make([]string, 0, len(allThemes))
	for name := range allThemes {
		themeNames = append(themeNames, name)
	}
	sort.Strings(themeNames)

	for _, themeName := range themeNames {
		t.Run(themeName, func(t *testing.T) {
			theme.Default().SetTheme(allThemes[themeName])

			bgTcell := theme.Bg()
			r, g, b := bgTcell.RGB()
			canvasBg := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
			bgStyle := tcell.StyleDefault.Background(bgTcell)

			themeDir := filepath.Join(outDir, themeName)

			for _, c := range imageSnapshotCases() {
				t.Run(c.name, func(t *testing.T) {
					screen := tcell.NewSimulationScreen("UTF-8")
					if err := screen.Init(); err != nil {
						t.Fatalf("screen init: %v", err)
					}
					screen.SetSize(c.w, c.h)

					// Pre-fill all cells with the theme bg so components that
					// only paint a sub-region (e.g. ContextMenu) don't leave
					// black uninitialized cells in the snapshot.
					for y := range c.h {
						for x := range c.w {
							screen.SetContent(x, y, ' ', nil, bgStyle)
						}
					}

					p := c.make()
					p.SetRect(0, 0, c.w, c.h)
					p.Draw(screen)
					screen.Show()

					img := snapshots.RenderCentered(screen, c.w, c.h, canvasBg)
					path := filepath.Join(themeDir, fmt.Sprintf("%s.png", c.name))
					if err := snapshots.SavePNG(img, path); err != nil {
						t.Fatalf("save PNG: %v", err)
					}
					t.Logf("wrote %s", path)
				})
			}
		})
	}
}
