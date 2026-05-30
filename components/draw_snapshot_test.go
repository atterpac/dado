package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

// Golden snapshot tests pin the exact rendered grid (runes + fg + bg per cell)
// of each widget so the Draw-boilerplate sweep (manual SetContent loops -> the
// draw.go helpers) can be proven byte-for-byte behavior preserving. Regenerate
// after an intentional rendering change with: UPDATE_GOLDEN=1 go test ./components/ -run Snapshot
//
// These render against a fixed theme so colors are deterministic.

type snapCase struct {
	name string
	w, h int
	make func() tview.Primitive
}

func drawSnapshotCases() []snapCase {
	return []snapCase{
		{"badge", 16, 1, func() tview.Primitive { return NewBadge("NEW").SetVariant(BadgeSuccess) }},
		{"chip", 18, 1, func() tview.Primitive { return NewChip("tag").SetRemovable(true) }},
		{"divider_label", 24, 1, func() tview.Primitive { return NewDivider().SetLabel("Section") }},
		{"button", 12, 1, func() tview.Primitive { return NewButton("OK") }},
		{"progressbar", 30, 1, func() tview.Primitive {
			return NewProgressBar().SetProgress(0.42).SetLabel("Load").SetShowPercentage(true)
		}},
		{"gauge", 22, 3, func() tview.Primitive { return NewGauge().SetValue(63).SetLabel("CPU").SetUnit("%") }},
		{"sparkline", 20, 1, func() tview.Primitive {
			return NewSparkline().SetValues([]float64{1, 3, 2, 5, 4, 6, 2, 7})
		}},
		{"checkbox", 24, 1, func() tview.Primitive { return NewCheckbox("agree").SetLabel("Agree").SetChecked(true) }},
		{"radiogroup", 24, 4, func() tview.Primitive {
			return NewRadioGroup("pick").SetLabel("Pick").SetOptions([]string{"One", "Two", "Three"})
		}},
		{"metriccard", 26, 6, func() tview.Primitive {
			return NewMetricCard().SetLabel("Requests").SetValue("1.2k").SetTrend(TrendUp, "+5%", true)
		}},
		{"barchart", 32, 10, func() tview.Primitive {
			return NewBarChart().SetValues([]float64{3, 7, 5, 9}, []string{"a", "b", "c", "d"}).SetShowValues(true)
		}},
		{"tabs", 30, 1, func() tview.Primitive {
			return NewTabs().AddTab("One", nil).AddTab("Two", nil).SetActive(0)
		}},
		{"contextmenu", 22, 6, func() tview.Primitive {
			return NewContextMenu().AddItem("a", "Copy", nil).AddItem("b", "Paste", nil)
		}},
		{"keyhintbar", 40, 1, func() tview.Primitive {
			return NewKeyHintBar().SetHints([]KeyHint{{Key: "Enter", Description: "Select"}, {Key: "Esc", Description: "Close"}})
		}},
		{"hintgrid", 40, 3, func() tview.Primitive {
			return NewHintGrid().SetHints([]KeyHint{{Key: "j", Description: "Down"}, {Key: "k", Description: "Up"}})
		}},
		{"panel", 24, 6, func() tview.Primitive { return NewPanel().SetTitle("Title") }},
		{"select", 28, 1, func() tview.Primitive {
			return NewSelect("s").SetLabel("Pick").SetOptions([]string{"Alpha", "Beta", "Gamma"})
		}},
		{"multiselect", 28, 4, func() tview.Primitive {
			return NewMultiSelect("m").SetLabel("Pick").SetOptions([]string{"Alpha", "Beta"})
		}},
		{"textfield", 30, 1, func() tview.Primitive {
			return NewTextField("t").SetLabel("Name").SetValue("Ada")
		}},
		{"textfield_placeholder", 30, 1, func() tview.Primitive {
			return NewTextField("t").SetLabel("Name").SetPlaceholder("enter name")
		}},
		{"textarea", 30, 4, func() tview.Primitive {
			return NewTextArea("ta").SetLabel("Bio").SetValue("hello")
		}},
		{"autocomplete", 30, 1, func() tview.Primitive {
			return NewAutocompleteInput().SetText("que")
		}},
		{"statusbar", 40, 1, func() tview.Primitive {
			return NewStatusBar().SetLeft(StatusSection{Text: "Ready"}).SetRight(StatusSection{Text: "v1"})
		}},
		{"searchbar", 40, 1, func() tview.Primitive {
			return NewSearchBar().SetPlaceholder("Search").SetQuery("foo")
		}},
		{"finder", 40, 10, func() tview.Primitive {
			return NewFinder().SetPrompt("> ").SetPlaceholder("Find")
		}},
		{"progressmodal", 44, 12, func() tview.Primitive {
			return NewProgressModal().SetMessage("Working")
		}},
		// Viz/container widgets rendered empty: the background-clear loops run
		// regardless of data, so these pin those fills even without content.
		{"codeview", 40, 10, func() tview.Primitive { return NewCodeView() }},
		{"gitgraph", 40, 10, func() tview.Primitive { return NewGitGraph() }},
		{"heatmap", 30, 10, func() tview.Primitive { return NewHeatMap() }},
		{"nodegraph", 40, 12, func() tview.Primitive { return NewNodeGraph() }},
		{"tree", 30, 8, func() tview.Primitive { return NewTree() }},
		{"linegraph", 32, 10, func() tview.Primitive { return NewLineGraph() }},
		{"graphtree", 40, 12, func() tview.Primitive { return NewGraphTree() }},
		{"timeline", 40, 8, func() tview.Primitive { return NewTimeline() }},
		{"virtuallist", 30, 8, func() tview.Primitive { return NewVirtualList() }},
		{"logviewer", 40, 10, func() tview.Primitive { return NewLogViewer() }},
		{"diffviewer", 40, 10, func() tview.Primitive { return NewDiffViewer() }},
		{"datagrid", 40, 10, func() tview.Primitive { return NewDataGrid() }},
		{"erdgraph", 40, 12, func() tview.Primitive { return NewERDGraph() }},
	}
}

func renderSnapshot(p tview.Primitive, w, h int) string {
	ts := newTestScreen(w, h)
	p.SetRect(0, 0, w, h)
	p.Draw(ts.SimulationScreen)
	ts.Show()

	var b strings.Builder
	b.WriteString("RUNES\n")
	for y := 0; y < h; y++ {
		b.WriteString(strings.TrimRight(ts.getContent(0, y, w), " "))
		b.WriteByte('\n')
	}
	b.WriteString("STYLE\n")
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			_, _, style, _ := ts.SimulationScreen.GetContent(x, y)
			fg, bg, _ := style.Decompose()
			fmt.Fprintf(&b, "%d:%d ", int64(fg.Hex()), int64(bg.Hex()))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestDrawSnapshots(t *testing.T) {
	// Deterministic colors for the whole test.
	theme.Default().SetAutoRefresh(false)
	defer theme.Default().SetAutoRefresh(true)
	theme.Default().SetTheme(themes.Nord)

	dir := filepath.Join("testdata", "draw_golden")
	update := os.Getenv("UPDATE_GOLDEN") != ""
	if update {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	for _, c := range drawSnapshotCases() {
		t.Run(c.name, func(t *testing.T) {
			got := renderSnapshot(c.make(), c.w, c.h)
			path := filepath.Join(dir, c.name+".golden")
			if update {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("missing golden %s (run UPDATE_GOLDEN=1 to create): %v", path, err)
			}
			if string(want) != got {
				t.Errorf("render changed for %s.\n--- want ---\n%s\n--- got ---\n%s", c.name, want, got)
			}
		})
	}
}

// TestScopedThemeChangesColors proves base-routed theming: a widget reads its
// colors through w.th() at draw time, so SetTheme(scopedProvider) takes effect
// without touching the global default, and an unscoped widget still falls back
// to it. (The golden tests above pin the unscoped == default case byte-for-byte;
// this pins the it-actually-rescopes case.)
func TestScopedThemeChangesColors(t *testing.T) {
	theme.Default().SetAutoRefresh(false)
	defer theme.Default().SetAutoRefresh(true)
	theme.Default().SetTheme(themes.Nord)

	scoped := func(th theme.Theme) *theme.Provider {
		p := theme.NewProvider()
		p.SetAutoRefresh(false)
		p.SetTheme(th)
		return p
	}
	render := func(p *theme.Provider) string {
		b := NewButton("OK")
		if p != nil {
			b.SetTheme(p)
		}
		return renderSnapshot(b, 12, 1)
	}

	nord := render(scoped(themes.Nord))
	dracula := render(scoped(themes.Dracula))
	unscoped := render(nil)

	// Same widget, two different scoped themes -> the rendered cell styles must
	// differ. If they match, th() is not actually consulting the scoped provider.
	if nord == dracula {
		t.Fatal("scoped theme had no effect: Nord-scoped and Dracula-scoped renders are identical")
	}
	// An unscoped widget reads the package default (Nord here), so it must match
	// the explicitly Nord-scoped render exactly.
	if unscoped != nord {
		t.Errorf("unscoped widget did not fall back to the default theme.\n--- default ---\n%s\n--- nord-scoped ---\n%s", unscoped, nord)
	}
}
