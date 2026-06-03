package components

import (
	"testing"
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
