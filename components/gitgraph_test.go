package components

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/atterpac/dado/theme"
)

func buildData(commits []*GitCommit, current string) *GitGraphData {
	d := NewGitGraphData()
	d.CurrentBranch = current
	for _, c := range commits {
		if len(c.Parents) > 1 {
			c.IsMerge = true
		}
		d.AddCommit(c)
	}
	d.LayoutGraph()
	return d
}

func columns(d *GitGraphData) []int {
	cols := make([]int, len(d.Commits))
	for i, c := range d.Commits {
		cols[i] = c.Column
	}
	return cols
}

// Layout must be deterministic across repeated runs (no map-iteration nondeterminism).
func TestLayoutDeterministic(t *testing.T) {
	mk := func() []*GitCommit {
		return []*GitCommit{
			{Hash: "a8", Parents: []string{"merge2"}, Refs: []string{"HEAD", "main"}},
			{Hash: "merge2", Parents: []string{"a5", "c1"}},
			{Hash: "c1", Parents: []string{"a5"}, Refs: []string{"hotfix"}},
			{Hash: "a5", Parents: []string{"merge1"}},
			{Hash: "merge1", Parents: []string{"a2", "b2"}},
			{Hash: "b2", Parents: []string{"b1"}, Refs: []string{"feature"}},
			{Hash: "b1", Parents: []string{"a2"}},
			{Hash: "a2", Parents: []string{"a1"}},
			{Hash: "a1", Parents: nil},
		}
	}
	want := columns(buildData(mk(), "main"))
	for i := 0; i < 200; i++ {
		got := columns(buildData(mk(), "main"))
		for j := range want {
			if got[j] != want[j] {
				t.Fatalf("run %d: column mismatch at %d: got %v want %v", i, j, got, want)
			}
		}
	}
	// HEAD/main tip must be column 0.
	if want[0] != 0 {
		t.Fatalf("current branch tip not in column 0: %v", want)
	}
}

// Lanes freed by terminated branches must be reused instead of growing the cap.
func TestLaneReclamation(t *testing.T) {
	// main with three sequential, fully-merged feature branches. Each feature
	// occupies lane 1 while live; after merging, lane 1 should free and be
	// reused by the next feature rather than spilling to lane 2, 3, ...
	commits := []*GitCommit{
		{Hash: "m3", Parents: []string{"a3", "f3"}}, // merge feat3
		{Hash: "f3", Parents: []string{"a3"}},
		{Hash: "a3", Parents: []string{"m2"}},
		{Hash: "m2", Parents: []string{"a2", "f2"}}, // merge feat2
		{Hash: "f2", Parents: []string{"a2"}},
		{Hash: "a2", Parents: []string{"m1"}},
		{Hash: "m1", Parents: []string{"a1", "f1"}}, // merge feat1
		{Hash: "f1", Parents: []string{"a1"}},
		{Hash: "a1", Parents: nil, Refs: []string{"HEAD", "main"}},
	}
	d := buildData(commits, "main")
	if d.MaxColumn != 1 {
		t.Fatalf("expected lanes to be reclaimed (MaxColumn==1), got MaxColumn=%d cols=%v", d.MaxColumn, columns(d))
	}

	// Each feature branch reuses lane 1, but they are distinct branches, so
	// their nodes must carry distinct color ids — otherwise the reused column
	// would render in the same color and the branches would be indistinguishable.
	g := NewGitGraph()
	g.SetGraph(d)
	rs := g.buildRowStates()
	colorOf := func(hash string) int {
		for i, c := range d.Commits {
			if c.Hash == hash {
				return rs[i][c.Column].colorID
			}
		}
		t.Fatalf("commit %s not found", hash)
		return -1
	}
	f1, f2, f3 := colorOf("f1"), colorOf("f2"), colorOf("f3")
	if f1 == f2 || f2 == f3 || f1 == f3 {
		t.Fatalf("features sharing reused lane got same color id: f1=%d f2=%d f3=%d", f1, f2, f3)
	}
}

// Octopus merge (3 parents) must place each parent in its own lane and not panic.
func TestOctopusMerge(t *testing.T) {
	commits := []*GitCommit{
		{Hash: "oct", Parents: []string{"a1", "b1", "c1"}, Refs: []string{"HEAD", "main"}},
		{Hash: "b1", Parents: []string{"a1"}},
		{Hash: "c1", Parents: []string{"a1"}},
		{Hash: "a1", Parents: nil},
	}
	d := buildData(commits, "main")
	cols := columns(d)
	// merge node at 0; the two extra parents must get distinct lanes.
	if cols[0] != 0 {
		t.Fatalf("octopus merge node not at column 0: %v", cols)
	}
	if d.MaxColumn < 2 {
		t.Fatalf("octopus merge did not allocate distinct lanes, MaxColumn=%d cols=%v", d.MaxColumn, cols)
	}
	g := NewGitGraph()
	g.SetGraph(d)
	if rs := g.buildRowStates(); len(rs) != len(commits) {
		t.Fatalf("row states length mismatch: %d", len(rs))
	}
}

// buildRowStates must return the cached states without recomputation.
func TestRowStatesCached(t *testing.T) {
	commits := []*GitCommit{
		{Hash: "a2", Parents: []string{"a1"}, Refs: []string{"HEAD", "main"}},
		{Hash: "a1", Parents: nil},
	}
	d := buildData(commits, "main")
	g := NewGitGraph()
	g.SetGraph(d)
	first := g.buildRowStates()
	second := g.buildRowStates()
	if len(first) == 0 || len(first) != len(second) {
		t.Fatalf("cached states not stable")
	}
	// identity: same backing slice element pointers
	if first[0][0] != second[0][0] {
		t.Fatalf("buildRowStates recomputed instead of returning cache")
	}
}

// Long branch names in the ref decoration must be truncated with an ellipsis,
// and the cap must be configurable / disableable.
func TestRefNameTruncation(t *testing.T) {
	long := "feature/some-really-long-branch-name-that-overflows"
	commits := []*GitCommit{
		{Hash: "a1", Parents: nil, Refs: []string{long}},
	}
	g := NewGitGraph()
	g.SetGraph(buildData(commits, long))

	// Default cap (20): name shown is truncated and ends with an ellipsis.
	got := g.buildRefTexts()[0]
	if !strings.Contains(got, "…") {
		t.Fatalf("expected ellipsis in truncated ref, got %q", got)
	}
	if strings.Contains(got, long) {
		t.Fatalf("long name not truncated: %q", got)
	}

	// Custom cap.
	g.SetMaxRefLen(8)
	if got := g.buildRefTexts()[0]; !strings.Contains(got, "feature…") {
		t.Fatalf("custom cap not applied: %q", got)
	}

	// Disabled.
	g.SetMaxRefLen(-1)
	if got := g.buildRefTexts()[0]; !strings.Contains(got, long) {
		t.Fatalf("truncation should be disabled, got %q", got)
	}
}

func debounceGraph() *GitGraph {
	commits := []*GitCommit{
		{Hash: "a3", Parents: []string{"a2"}},
		{Hash: "a2", Parents: []string{"a1"}},
		{Hash: "a1", Parents: nil},
	}
	g := NewGitGraph()
	g.SetGraph(buildData(commits, "main"))
	return g
}

// A burst of selection changes must collapse into a single onChange after the
// debounce interval elapses, not one call per step.
func TestChangeDebounceCoalesces(t *testing.T) {
	g := debounceGraph()

	var mu sync.Mutex
	var calls int
	var last string
	g.SetOnChange(func(c *GitCommit) {
		mu.Lock()
		calls++
		if c != nil {
			last = c.Hash
		}
		mu.Unlock()
	})
	g.SetChangeDebounce(20 * time.Millisecond)

	// Three rapid moves within one debounce window.
	g.SetSelectedIndex(1)
	g.SetSelectedIndex(2)
	g.SetSelectedIndex(1)

	mu.Lock()
	if calls != 0 {
		mu.Unlock()
		t.Fatalf("onChange fired during burst, want debounced; got %d calls", calls)
	}
	mu.Unlock()

	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("debounced onChange fired %d times, want 1", calls)
	}
	if last != "a2" {
		t.Fatalf("debounced onChange saw stale selection %q, want final a2", last)
	}
}

// Tearing down (Subs().Release, as ComponentBase.Stop does) while a debounce is
// in flight must not race on changeTimer. The fired callback is routed through
// the theme queue exactly as the real app routes it onto the draw goroutine, so
// selection reads stay on the move goroutine; only changeTimer is contended.
// Run under -race.
func TestChangeDebounceTeardownRace(t *testing.T) {
	// Serialize queued callbacks onto a single goroutine, mirroring the app's
	// draw-goroutine contract instead of running them on the timer goroutine.
	ui := make(chan func(), 512)
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case fn := <-ui:
				fn()
			case <-stop:
				return
			}
		}
	}()
	theme.SetQueue(func(fn func()) {
		select {
		case ui <- fn:
		case <-stop:
		}
	})
	defer func() { theme.SetQueue(nil); close(stop) }()

	g := debounceGraph()
	g.SetOnChange(func(*GitCommit) {})
	g.SetChangeDebounce(2 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { // moves run on the ui goroutine, so selectedIndex stays single-threaded
		defer wg.Done()
		for i := 0; i < 200; i++ {
			idx := i % 3
			ui <- func() { g.SetSelectedIndex(idx) }
		}
	}()
	go func() { // teardown goroutine: contends changeTimer via Release
		defer wg.Done()
		for i := 0; i < 200; i++ {
			g.Subs().Release()
		}
	}()
	wg.Wait()
	time.Sleep(10 * time.Millisecond)
}
