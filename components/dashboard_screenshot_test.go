package components

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/snapshots"
	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

// TestGenerateDashboardScreenshot renders a dense, multi-component "kitchen sink"
// dashboard populated with mock data and writes it to a single PNG. Intended for
// the website landing page.
//
// Only runs when GENERATE_DASHBOARD=1. Output defaults to
// ../website/public/images/dashboard.png; override with DASHBOARD_OUT=/abs/path.
//
//	GENERATE_DASHBOARD=1 go test ./components/ -run TestGenerateDashboardScreenshot
//	GENERATE_DASHBOARD=1 DASHBOARD_THEME=tokyonight-night DASHBOARD_OUT=/tmp/d.png \
//	    go test ./components/ -run TestGenerateDashboardScreenshot
func TestGenerateDashboardScreenshot(t *testing.T) {
	if os.Getenv("GENERATE_DASHBOARD") == "" {
		t.Skip("set GENERATE_DASHBOARD=1 to run")
	}

	themeName := os.Getenv("DASHBOARD_THEME")
	if themeName == "" {
		themeName = "tokyonight-night"
	}
	outPath := os.Getenv("DASHBOARD_OUT")
	if outPath == "" {
		outPath = filepath.Join("..", "website", "public", "images", "dashboard.png")
	}

	theme.Default().SetAutoRefresh(false)
	defer theme.Default().SetAutoRefresh(true)
	th := themes.Get(themeName)
	if th == nil {
		t.Fatalf("unknown theme %q", themeName)
	}
	theme.Default().SetTheme(th)

	const w, h = 200, 54

	root := buildDashboard()

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen init: %v", err)
	}
	screen.SetSize(w, h)

	bgStyle := tcell.StyleDefault.Background(theme.Bg())
	for y := range h {
		for x := range w {
			screen.SetContent(x, y, ' ', nil, bgStyle)
		}
	}

	root.SetRect(0, 0, w, h)
	root.Draw(screen)
	screen.Show()

	img := snapshots.RenderToPNG(screen)
	if err := snapshots.SavePNG(img, outPath); err != nil {
		t.Fatalf("save PNG: %v", err)
	}
	t.Logf("wrote %s (%dx%d cells, theme %s)", outPath, w, h, themeName)
}

// panel frames a primitive with a titled border for visual density.
func panel(title string, p core.Widget) core.Widget {
	return NewPanel().SetTitle(title).SetContent(p)
}

// barRow builds a 1-row bar with left-aligned and right-aligned dynamic-color
// text, matching the custom status/command bars used in the qry and gxt apps.
func barRow(leftText, rightText string) core.Widget {
	bg := theme.Bg()
	left := core.NewTextView()
	left.SetDynamicColors(true)
	left.SetText(leftText)
	left.Box.SetBackgroundColor(bg)
	right := core.NewTextView()
	right.SetDynamicColors(true)
	right.SetText(rightText)
	right.Box.SetBackgroundColor(bg)
	row := core.NewFlex().SetDirection(core.Row)
	row.AddItem(left, 0, 1, false)
	row.AddItem(right, 0, 1, false)
	row.Box.SetBackgroundColor(bg)
	return row
}

// Nerd Font glyphs mirroring the icons gxt uses in its status bar header.
const (
	glyphGitBranch = "" // nf-dev-git_branch
	glyphGitCommit = "" // nf-dev-git_commit
	glyphStaged    = "" // nf-fa-check
	glyphUnstaged  = "" // nf-fa-pencil
	glyphRefresh   = "" // nf-fa-refresh
)

// appStatusBar mimics the gxt application header: a bordered status bar with an
// app-name title, a connection indicator, and git-style sections (branch,
// commit, ahead/behind, staged/unstaged) separated by dot bullets, plus a
// right-aligned environment/version block.
func appStatusBar() core.Widget {
	bg := theme.Bg()
	sep := "  [#7a7665]•[-]  "

	left := core.NewTextView()
	left.SetDynamicColors(true)
	left.Box.SetBackgroundColor(bg)
	left.SetText(
		"[#9ab87a]" + theme.IconConnected + "[-] ~/src/control-plane" + sep +
			"[#c578b0]" + glyphGitBranch + " main[-]" + sep +
			"[#7a7665]" + glyphGitCommit + " 3a281b7[-]" + sep +
			"[#d9a441]" + glyphRefresh + " ↑2 ↓1[-]" + sep +
			"[#9ab87a]" + glyphStaged + " 3 +128 -14[-]" + sep +
			"[#d9a441]" + glyphUnstaged + " 5 +42 -18[-]",
	)

	right := core.NewTextView()
	right.SetDynamicColors(true)
	right.Box.SetBackgroundColor(bg)
	right.SetText("[#9ab87a]● prod[-]  [::d]4 regions  •  v2.4.1[::-] ")

	row := core.NewFlex().SetDirection(core.Row)
	row.AddItem(left, 0, 1, false)
	row.AddItem(right, 0, 1, false)
	row.Box.SetBackgroundColor(bg)

	return NewPanel().SetTitle(" dado ").SetTitleAlign(TitleAlignLeft).SetContent(row)
}

// appCommandBar mimics the gxt/qry bottom status/command line: a mode block, a
// ":" command prompt, and right-aligned key hints.
func appCommandBar() core.Widget {
	return barRow(
		"[black:#c578b0:b] NORMAL [-:-:-] [#c578b0]:[-] [::d]filter services…[::-]",
		"[::d]j/k nav  •  / search  •  : cmd  •  ? help  •  q quit[::-] ",
	)
}

func buildDashboard() core.Widget {
	t0 := time.Date(2026, 5, 31, 14, 2, 0, 0, time.UTC)
	tt := func(m int) time.Time { return t0.Add(time.Duration(m) * time.Minute) }

	// ── Top status bar (gxt app-header style) ─────────────────────────────
	header := appStatusBar()

	// ── KPI metric cards ──────────────────────────────────────────────────
	mc := func(label, val string, dir Trend, delta string, good bool, spark []float64) core.Widget {
		c := NewMetricCard().SetLabel(label).SetValue(val).SetTrend(dir, delta, good)
		c.SetCompact(true)
		c.SetSparkline(spark)
		return c
	}
	metrics := core.NewFlex().SetDirection(core.Row)
	metrics.AddItem(mc("Requests/s", "18.4k", TrendUp, "+12%", true,
		[]float64{9, 11, 10, 13, 12, 15, 14, 17, 16, 18, 17, 18}), 0, 1, false)
	metrics.AddItem(mc("p99 Latency", "82ms", TrendDown, "-8ms", true,
		[]float64{120, 110, 115, 98, 102, 90, 95, 88, 84, 86, 83, 82}), 0, 1, false)
	metrics.AddItem(mc("Error Rate", "0.41%", TrendUp, "+0.1%", false,
		[]float64{0.2, 0.25, 0.3, 0.28, 0.35, 0.31, 0.38, 0.36, 0.4, 0.39, 0.42, 0.41}), 0, 1, false)
	metrics.AddItem(mc("Active Pods", "248", TrendUp, "+6", true,
		[]float64{210, 215, 220, 228, 232, 235, 240, 242, 244, 246, 247, 248}), 0, 1, false)
	metrics.AddItem(mc("Cost / day", "$1.2k", TrendDown, "-4%", true,
		[]float64{1.4, 1.38, 1.35, 1.33, 1.3, 1.28, 1.27, 1.25, 1.24, 1.23, 1.21, 1.2}), 0, 1, false)

	// ── Left column: latency line graph + request bar chart ───────────────
	latency := NewLineGraph().
		SetSeries(
			DataSeries{Label: "p50", Values: []float64{24, 28, 22, 30, 26, 34, 29, 38, 31, 40, 35, 42, 33, 30}},
			DataSeries{Label: "p99", Values: []float64{70, 82, 66, 90, 78, 102, 88, 115, 95, 120, 100, 124, 92, 82}},
		).
		SetTitle("Request latency (ms)").
		SetShowLegend(true).
		SetShowGrid(true)

	reqBar := NewBarChart().
		SetValues(
			[]float64{42, 58, 51, 73, 66, 88, 79, 61},
			[]string{"00", "03", "06", "09", "12", "15", "18", "21"},
		).
		SetShowValues(true)

	// Schema / ERD
	users := &ERDTable{ID: "users", Name: "users", Columns: []ERDColumn{
		{Name: "id", Type: "uuid", IsPK: true},
		{Name: "email", Type: "text"},
		{Name: "name", Type: "text"},
	}}
	orders := &ERDTable{ID: "orders", Name: "orders", Columns: []ERDColumn{
		{Name: "id", Type: "uuid", IsPK: true},
		{Name: "user_id", Type: "uuid", IsFK: true, FKTarget: "users.id"},
		{Name: "total", Type: "numeric"},
		{Name: "status", Type: "text"},
	}}
	items := &ERDTable{ID: "items", Name: "line_items", Columns: []ERDColumn{
		{Name: "id", Type: "uuid", IsPK: true},
		{Name: "order_id", Type: "uuid", IsFK: true, FKTarget: "orders.id"},
		{Name: "sku", Type: "text"},
		{Name: "qty", Type: "int"},
	}}
	erd := NewERDGraph().SetNodeWidth(15).SetSpacing(4, 1).SetFit(true).SetData(
		[]*ERDTable{users, orders, items},
		[]*ERDRelation{
			{FromTable: "users", ToTable: "orders", Cardinality: OneToMany},
			{FromTable: "orders", ToTable: "items", Cardinality: OneToMany},
		},
	)

	heat := NewHeatMap().SetValues([][]float64{
		{0.1, 0.3, 0.7, 0.9, 0.6, 0.4, 0.2},
		{0.4, 0.8, 1.0, 0.7, 0.3, 0.5, 0.6},
		{0.6, 0.9, 0.6, 0.2, 0.5, 0.8, 0.4},
	}).SetCellSize(9, 1).SetColorScale(ColorScaleTheme)

	logs := NewLogViewer()
	logs.AddEntry(LogEntry{Timestamp: tt(0), Level: LogLevelInfo, Message: "deploy v2.4.1 rolled out to prod", Source: "deployer"})
	logs.AddEntry(LogEntry{Timestamp: tt(1), Level: LogLevelInfo, Message: "health check passed (248/248 pods)", Source: "k8s"})
	logs.AddEntry(LogEntry{Timestamp: tt(2), Level: LogLevelDebug, Message: "scaling cache pool 8 → 12", Source: "autoscaler"})
	logs.AddEntry(LogEntry{Timestamp: tt(3), Level: LogLevelWarn, Message: "cache hit rate dropped to 71%", Source: "cache"})
	logs.AddEntry(LogEntry{Timestamp: tt(4), Level: LogLevelInfo, Message: "connection pool resized (db)", Source: "database"})
	logs.AddEntry(LogEntry{Timestamp: tt(5), Level: LogLevelError, Message: "upstream timeout on /v1/charge (retry 1/3)", Source: "billing"})
	logs.AddEntry(LogEntry{Timestamp: tt(6), Level: LogLevelInfo, Message: "retry succeeded in 412ms", Source: "billing"})
	logs.AddEntry(LogEntry{Timestamp: tt(7), Level: LogLevelInfo, Message: "nightly backup completed (512M)", Source: "database"})

	leftCol := core.NewFlex()
	leftCol.AddItem(panel(" Latency ", latency), 0, 3, false)
	leftCol.AddItem(panel(" Activity ", heat), 6, 0, false)
	leftCol.AddItem(panel(" Live logs ", logs), 0, 3, false)

	// ── Center column: services table + log stream ────────────────────────
	services := NewTable().
		SetHeaders("Service", "Status", "CPU", "Mem", "Up").
		AddRow("api-gateway", "● Running", "12%", "128M", "14d").
		AddRow("auth-service", "● Running", "4%", "64M", "14d").
		AddRow("billing", "● Running", "22%", "256M", "9d").
		AddRow("database", "● Running", "31%", "512M", "21d").
		AddRow("cache", "◌ Degraded", "71%", "1.2G", "2d").
		AddRow("worker-01", "● Running", "8%", "96M", "5d").
		AddRow("scheduler", "● Running", "3%", "48M", "5d").
		AddRow("ingest", "○ Stopped", "0%", "0M", "—")

	centerCol := core.NewFlex()
	centerCol.AddItem(panel(" Services ", services), 0, 2, false)
	centerCol.AddItem(panel(" Schema ", erd), 0, 3, false)

	// ── Right column: gauges, deploy timeline, dependency graph ───────────
	gauges := core.NewFlex()
	for _, g := range []struct {
		label string
		v     float64
	}{{"CPU", 0.42}, {"Memory", 0.68}, {"Disk", 0.81}, {"Net", 0.37}} {
		gauges.AddItem(NewGauge().SetValue(g.v).SetLabel(g.label).SetUnit("%"), 3, 0, false)
	}

	deps := NewNodeGraphData()
	deps.AddNode(&GraphNode{ID: "api", Label: "API Gateway", Status: "running", Focused: true})
	deps.AddNode(&GraphNode{ID: "auth", Label: "Auth", Status: "running"})
	deps.AddNode(&GraphNode{ID: "db", Label: "Database", Status: "running"})
	deps.AddNode(&GraphNode{ID: "cache", Label: "Cache", Status: "degraded"})
	deps.AddEdge(&GraphEdge{From: "api", To: "auth", Label: "JWT"})
	deps.AddEdge(&GraphEdge{From: "api", To: "db"})
	deps.AddEdge(&GraphEdge{From: "auth", To: "cache"})
	depGraph := NewNodeGraph().SetFit(true).SetData(deps)

	rightCol := core.NewFlex()
	rightCol.AddItem(panel(" Resources ", gauges), 15, 0, false)
	rightCol.AddItem(panel(" Dependencies ", depGraph), 0, 1, false)

	// ── Body: three columns ───────────────────────────────────────────────
	body := core.NewFlex().SetDirection(core.Row)
	body.AddItem(leftCol, 0, 3, false)
	body.AddItem(centerCol, 0, 3, false)
	body.AddItem(rightCol, 0, 2, false)

	// ── Bottom row: deploy timeline + heatmap + git graph ─────────────────
	timeline := NewTimeline().SetLanes([]TimelineLane{
		{Name: "Checkout", StartTime: tt(0), EndTime: tp(tt(2))},
		{Name: "Build", StartTime: tt(2), EndTime: tp(tt(14))},
		{Name: "Test", StartTime: tt(14), EndTime: tp(tt(28))},
		{Name: "Lint", StartTime: tt(14), EndTime: tp(tt(22))},
		{Name: "Deploy", StartTime: tt(28), EndTime: tp(tt(36))},
		{Name: "Verify", StartTime: tt(36)},
	})

	git := NewGitGraphData()
	git.CurrentBranch = "main"
	// Interleave trunk + branch commits so feature lanes stay open across
	// several rows (forces parallel columns instead of collapsing to lane 0).
	git.AddCommit(&GitCommit{Hash: "9f3a1c2", ShortHash: "9f3a1c2", Message: "Merge PR #142: fuzzy search", Author: "alice", Parents: []string{"7b21e4d", "c0de551"}, IsMerge: true, Refs: []string{"HEAD", "main"}})
	git.AddCommit(&GitCommit{Hash: "c0de551", ShortHash: "c0de551", Message: "feat: fuzzy ranking", Author: "dave", Branch: "feat/search", Parents: []string{"a14f0b9"}})
	git.AddCommit(&GitCommit{Hash: "7b21e4d", ShortHash: "7b21e4d", Message: "fix: nil deref in cache", Author: "bob", Parents: []string{"2c9b6f3"}})
	git.AddCommit(&GitCommit{Hash: "a14f0b9", ShortHash: "a14f0b9", Message: "feat: search index", Author: "dave", Branch: "feat/search", Parents: []string{"2c9b6f3"}, Refs: []string{"feat/search"}})
	git.AddCommit(&GitCommit{Hash: "2c9b6f3", ShortHash: "2c9b6f3", Message: "refactor: extract router", Author: "alice", Parents: []string{"5e8c7a0"}})
	git.AddCommit(&GitCommit{Hash: "5e8c7a0", ShortHash: "5e8c7a0", Message: "Merge PR #139: billing", Author: "alice", Parents: []string{"3a9d2f1", "b6f4e88"}, IsMerge: true})
	git.AddCommit(&GitCommit{Hash: "b6f4e88", ShortHash: "b6f4e88", Message: "test: webhook retries", Author: "carol", Branch: "feat/billing", Parents: []string{"d1e2f3a"}})
	git.AddCommit(&GitCommit{Hash: "3a9d2f1", ShortHash: "3a9d2f1", Message: "chore: bump deps", Author: "carol", Parents: []string{"e4f5a6b"}})
	git.AddCommit(&GitCommit{Hash: "d1e2f3a", ShortHash: "d1e2f3a", Message: "feat: stripe webhooks", Author: "carol", Branch: "feat/billing", Parents: []string{"e4f5a6b"}, Refs: []string{"feat/billing"}})
	git.AddCommit(&GitCommit{Hash: "e4f5a6b", ShortHash: "e4f5a6b", Message: "feat: add rate limiter", Author: "alice", Parents: []string{"a1b2c3d"}})
	git.AddCommit(&GitCommit{Hash: "a1b2c3d", ShortHash: "a1b2c3d", Message: "fix: token expiry off-by-one", Author: "bob"})
	git.LayoutGraph()
	gitGraph := NewGitGraph().SetGraph(git).SetShowAuthor(false).SetShowHash(true)

	bottom := core.NewFlex().SetDirection(core.Row)
	bottom.AddItem(panel(" Deploy pipeline ", timeline), 0, 3, false)
	bottom.AddItem(panel(" Requests by hour ", reqBar), 0, 2, false)
	bottom.AddItem(panel(" Recent commits ", gitGraph), 0, 5, false)

	// ── Bottom status / command bar (qry/gxt style) ───────────────────────
	footer := appCommandBar()

	// ── Assemble ──────────────────────────────────────────────────────────
	rootFlex := core.NewFlex()
	rootFlex.AddItem(header, 3, 0, false)
	rootFlex.AddItem(metrics, 3, 0, false)
	rootFlex.AddItem(body, 0, 3, false)
	rootFlex.AddItem(bottom, 15, 0, false)
	rootFlex.AddItem(footer, 1, 0, false)
	return rootFlex
}

func tp(t time.Time) *time.Time { return &t }
