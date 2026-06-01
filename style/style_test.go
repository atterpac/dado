package style

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRender_EmptyStyleNoTags(t *testing.T) {
	got := New().Render("hello")
	if got != "hello" {
		t.Fatalf("got %q want %q", got, "hello")
	}
}

func TestRender_AttrsTag(t *testing.T) {
	got := New().Bold().Underline().Render("x")
	if !strings.HasPrefix(got, "[::bu]") {
		t.Fatalf("missing bold+underline tag: %q", got)
	}
	if !strings.HasSuffix(got, "[-:-:-]") {
		t.Fatalf("missing close tag: %q", got)
	}
}

func TestRender_ForegroundStaticHexTag(t *testing.T) {
	got := New().Foreground(tcell.NewRGBColor(255, 0, 0)).Render("x")
	if !strings.HasPrefix(got, "[#FF0000::]") {
		t.Fatalf("got %q want hex foreground tag", got)
	}
}

func TestRender_EscapesUserInput(t *testing.T) {
	got := New().Bold().Render("a[red]b")
	// Escaping replaces "[" with "[[" — verify the inner '[' is escaped.
	if strings.Contains(got, "[red]") {
		t.Fatalf("user '[red]' tag not escaped: %q", got)
	}
}

func TestRender_PaddingX(t *testing.T) {
	got := New().PaddingX(2).Render("x")
	if got != "  x  " {
		t.Fatalf("got %q want '  x  '", got)
	}
}

func TestRender_PaddingY(t *testing.T) {
	got := New().PaddingY(1).Render("x")
	if got != "\nx\n" {
		t.Fatalf("got %q want '\\nx\\n'", got)
	}
}

func TestTcellStyle_Attributes(t *testing.T) {
	s := New().Bold().Italic().Foreground(tcell.ColorRed).Background(tcell.ColorBlue)
	st := s.TcellStyle()
	fg, bg, attrs := st.Decompose()
	if fg != tcell.ColorRed {
		t.Fatalf("fg=%v want red", fg)
	}
	if bg != tcell.ColorBlue {
		t.Fatalf("bg=%v want blue", bg)
	}
	if attrs&tcell.AttrBold == 0 {
		t.Fatal("bold not set")
	}
	if attrs&tcell.AttrItalic == 0 {
		t.Fatal("italic not set")
	}
}

func TestImmutableBuilders(t *testing.T) {
	a := New()
	b := a.Bold()
	if a.bold {
		t.Fatal("Bold mutated original Style")
	}
	if !b.bold {
		t.Fatal("Bold did not set on returned Style")
	}
}

func TestColorTagDefaultEmpty(t *testing.T) {
	if got := colorTag(tcell.ColorDefault); got != "" {
		t.Fatalf("ColorDefault should produce empty tag, got %q", got)
	}
}
