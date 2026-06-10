package components

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// makeGraph builds a synthetic commit graph: a linear mainline of `n` commits
// with a feature branch forked and merged back every `branchEvery` commits.
// This exercises lane allocation, reclamation, and merge handling in
// LayoutGraph, the allocation-heavy path run on every graph refresh.
func makeGraph(n, branchEvery int) []*GitCommit {
	commits := make([]*GitCommit, 0, n)
	hash := func(i int) string { return fmt.Sprintf("c%06d", i) }

	for i := 0; i < n; i++ {
		c := &GitCommit{
			Hash:      hash(i),
			ShortHash: hash(i)[:7],
			Message:   fmt.Sprintf("commit %d", i),
			Author:    "Bench",
		}
		if i+1 < n {
			c.Parents = []string{hash(i + 1)}
		}
		// Inject a merge of a short feature branch periodically.
		if branchEvery > 0 && i%branchEvery == 0 && i+2 < n {
			feat := &GitCommit{
				Hash:      hash(i) + "f",
				ShortHash: (hash(i) + "f")[:7],
				Message:   fmt.Sprintf("feature %d", i),
				Parents:   []string{hash(i + 2)},
			}
			c.Parents = []string{hash(i + 1), feat.Hash}
			c.IsMerge = true
			commits = append(commits, c, feat)
			i++ // feature consumes a slot in the visible list
			continue
		}
		commits = append(commits, c)
	}
	if len(commits) > 0 {
		commits[0].Refs = []string{"HEAD", "main"}
	}
	return commits
}

func BenchmarkLayoutGraph(b *testing.B) {
	cases := []struct {
		name        string
		n           int
		branchEvery int
	}{
		{"linear_100", 100, 0},
		{"linear_500", 500, 0},
		{"branchy_100", 100, 5},
		{"branchy_500", 500, 5},
	}
	for _, c := range cases {
		commits := makeGraph(c.n, c.branchEvery)
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				d := NewGitGraphData()
				d.CurrentBranch = "main"
				for _, c := range commits {
					d.AddCommit(c)
				}
				d.LayoutGraph()
			}
		})
	}
}

// BenchmarkLayoutGraphOnly isolates LayoutGraph from graph construction so the
// lane-layout allocations are measured on their own.
func BenchmarkLayoutGraphOnly(b *testing.B) {
	commits := makeGraph(500, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := NewGitGraphData()
		d.CurrentBranch = "main"
		for _, c := range commits {
			d.AddCommit(c)
		}
		b.StartTimer()
		d.LayoutGraph()
		b.StopTimer()
	}
}

// buildGraphData constructs and lays out a graph ready to render.
func buildGraphData(n, branchEvery int) *GitGraphData {
	d := NewGitGraphData()
	d.CurrentBranch = "main"
	for _, c := range makeGraph(n, branchEvery) {
		d.AddCommit(c)
	}
	d.LayoutGraph()
	return d
}

// BenchmarkGitGraphDraw measures the per-frame render cost, the hot path that
// runs on every keystroke/scroll, far more often than LayoutGraph. rowStates is
// cached by LayoutGraph, so a clean Draw should allocate little; this guards
// against per-cell allocations creeping into the render loop.
func BenchmarkGitGraphDraw(b *testing.B) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(120, 40)

	for _, c := range []struct {
		name      string
		n, branch int
	}{
		{"linear_500", 500, 0},
		{"branchy_500", 500, 5},
	} {
		g := NewGitGraph().SetGraph(buildGraphData(c.n, c.branch))
		g.SetRect(0, 0, 120, 40)
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				screen.Clear()
				g.Draw(screen)
				screen.Show()
			}
		})
	}
}

// noAllocScreen wraps a SimulationScreen but makes SetContent a no-op. The
// real per-cell SetContent cost is a property of the screen backend (and is
// allocation-free on a live terminal's persistent CellBuffer), so benchmarking
// Draw against a real SimulationScreen measures ~3 allocs/cell of harness
// overhead that swamps Draw's own logic. This isolates Draw's own allocations.
type noAllocScreen struct{ tcell.SimulationScreen }

func (noAllocScreen) SetContent(x, y int, mainc rune, combc []rune, style tcell.Style) {}

// BenchmarkGitGraphDrawOwnAllocs measures only Draw's own per-frame allocations
// (excluding the screen backend's per-cell cost). This is the number that
// matters on a real terminal and the regression guard for keeping Draw
// allocation-free.
func BenchmarkGitGraphDrawOwnAllocs(b *testing.B) {
	sim := tcell.NewSimulationScreen("UTF-8")
	_ = sim.Init()
	sim.SetSize(120, 40)
	screen := noAllocScreen{sim}

	g := NewGitGraph().SetGraph(buildGraphData(500, 5)).
		SetShowHash(true).SetShowAuthor(true).SetShowRefs(true).SetShowDate(true)
	g.SetRect(0, 0, 120, 40)
	g.Draw(screen) // warm caches and scratch buffers

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Draw(screen)
	}
}

// BenchmarkGitGraphScroll measures Draw while the selection moves every frame,
// the realistic "hold j to scroll" path; exercises offset/viewport handling on
// top of the render cost.
func BenchmarkGitGraphScroll(b *testing.B) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(120, 40)

	g := NewGitGraph().SetGraph(buildGraphData(500, 5))
	g.SetRect(0, 0, 120, 40)
	n := len(g.GetGraph().Commits)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.SetSelectedIndex(i % n)
		screen.Clear()
		g.Draw(screen)
		screen.Show()
	}
}

func BenchmarkAddCommit(b *testing.B) {
	commits := makeGraph(1000, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := NewGitGraphData()
		for _, c := range commits {
			d.AddCommit(c)
		}
	}
}
