package inline

import (
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextFieldEditingAndValidation(t *testing.T) {
	t.Parallel()
	field := NewTextField("name", "Name").Required().SetPlaceholder("project")
	for _, r := range []rune("ac") {
		field.handle(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	field.handle(tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone))
	field.handle(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	assert.Equal(t, "abc", field.Value())
	assert.NoError(t, field.validate())
	field.SetValue("")
	assert.EqualError(t, field.validate(), "required")
}

func TestTextFieldScrollsToKeepCursorVisible(t *testing.T) {
	t.Parallel()
	field := NewTextField("name", "Name").SetValue("abcdefghijklmnop")
	frame := NewFrame(10, 4)
	field.draw(frame, 0, 10, true, RoundedInlineTheme())
	require.NotNil(t, frame.cursor)
	assert.Less(t, frame.cursor.X, 10)
	assert.Contains(t, strings.Join(plainFrameLines(frame), "\n"), "p")
}

func TestSelectAndMultiSelectFields(t *testing.T) {
	t.Parallel()
	choices := []Choice{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta", Disabled: true}, {Value: "c", Label: "Gamma"}}
	selectField := NewSelectField("one", "One", choices...).Required()
	selectField.handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	selectField.handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	assert.Equal(t, "c", selectField.Value())

	multi := NewMultiSelectField("many", "Many", choices...).MinSelected(1).MaxSelected(2)
	multi.handle(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	multi.handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	multi.handle(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	assert.Equal(t, []string{"a", "c"}, multi.Value())
	assert.NoError(t, multi.validate())
}

func TestMultiSelectIndicatorPresets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		indicator            MultiSelectIndicator
		selected, unselected string
	}{{"theme", IndicatorTheme, "✓", "○"}, {"check", IndicatorCheck, "✓", "○"}, {"dot", IndicatorDot, "●", "○"}, {"diamond", IndicatorDiamond, "◆", "◇"}, {"box", IndicatorBox, "▣", "□"}, {"minimal", IndicatorMinimal, "✓", "·"}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := NewMultiSelectField("targets", "Targets", NewChoice("linux", "Linux"), NewChoice("darwin", "macOS")).SetIndicator(test.indicator)
			field.handle(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
			frame := NewFrame(36, 10)
			field.draw(frame, 0, 36, true, RoundedInlineTheme())
			plain := strings.Join(plainFrameLines(frame), "\n")
			assert.Contains(t, plain, test.selected+" Linux")
			assert.Contains(t, plain, test.unselected+" macOS")
			assert.NotContains(t, plain, "[")
		})
	}
}

func TestMultiSelectCustomAndMarkerlessIndicators(t *testing.T) {
	t.Parallel()
	choices := []Choice{NewChoice("linux", "Linux"), NewChoice("darwin", "macOS"), {Value: "windows", Label: "Windows", Disabled: true}}
	custom := NewMultiSelectField("targets", "Targets", choices...).SetIndicatorGlyphs("Y", "N", "D")
	custom.handle(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	frame := NewFrame(36, 10)
	custom.draw(frame, 0, 36, true, RoundedInlineTheme())
	plain := strings.Join(plainFrameLines(frame), "\n")
	assert.Contains(t, plain, "Y Linux")
	assert.Contains(t, plain, "N macOS")
	assert.Contains(t, plain, "D Windows")

	markerless := NewMultiSelectField("targets", "Targets", NewChoice("linux", "Linux"), NewChoice("darwin", "macOS")).SetIndicator(IndicatorNone)
	markerless.handle(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	frame = NewFrame(36, 10)
	markerless.draw(frame, 0, 36, true, RoundedInlineTheme())
	plain = strings.Join(plainFrameLines(frame), "\n")
	assert.NotContains(t, plain, "✓")
	assert.NotContains(t, plain, "○")
	_, style, _ := frame.Cell(2, 2)
	_, _, attrs := style.Decompose()
	assert.NotZero(t, attrs&tcell.AttrReverse)
}

func TestFormSubmissionAndSummary(t *testing.T) {
	t.Parallel()
	form := NewForm("Create").Add(
		NewTextField("name", "Name").Required(),
		NewSelectField("channel", "Channel", NewChoice("stable", "Stable"), NewChoice("preview", "Preview")).Required(),
		NewMultiSelectField("targets", "Targets", NewChoice("linux", "Linux"), NewChoice("darwin", "macOS")).MinSelected(1),
	)
	for _, r := range []rune("demo") {
		form.HandleKey(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
	form.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	require.True(t, form.Done())
	result := form.Result()
	assert.Equal(t, "demo", result["name"])
	assert.Equal(t, "stable", result["channel"])
	assert.Equal(t, []string{"linux"}, result["targets"])
	plain := strings.Join(plainFrameLines(form.Frame(48)), "\n")
	assert.Contains(t, plain, "✓ Create")
	assert.Contains(t, plain, "Name")
	assert.Contains(t, plain, "demo")
}

func TestFormValidationFocusesFirstInvalidField(t *testing.T) {
	t.Parallel()
	name := NewTextField("name", "Name").Required()
	choice := NewSelectField("choice", "Choice", NewChoice("x", "X")).Required()
	form := NewForm("Validate").Add(name, choice)
	form.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	assert.False(t, form.Done())
	assert.Equal(t, 0, form.focus)
	assert.Equal(t, errors.New("required").Error(), name.err.Error())
}

func TestFormNarrowFramesStayWithinWidth(t *testing.T) {
	t.Parallel()
	form := NewForm("Long form title").Add(NewTextField("name", "A very long field label").SetPlaceholder("A very long placeholder"))
	for width := range 30 {
		for _, line := range plainFrameLines(form.Frame(width)) {
			assert.LessOrEqual(t, displayWidth(line), width)
		}
	}
}

func TestFormFocusedControlUsesThemeChrome(t *testing.T) {
	t.Parallel()
	form := NewForm("Styled").Add(NewTextField("name", "Name").Required().SetPlaceholder("project"))
	plain := strings.Join(plainFrameLines(form.Frame(36)), "\n")
	assert.Contains(t, plain, "Styled ─")
	assert.Contains(t, plain, "Name *")
	assert.Contains(t, plain, "╭")
	assert.Contains(t, plain, "╰")
	assert.Contains(t, plain, "[Tab]")
}

func TestFormThemePresets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, prefix string
		theme        InlineTheme
	}{{"rounded", "╭", RoundedInlineTheme()}, {"square", "┌", SquareInlineTheme()}, {"ascii", "+", ASCIIInlineTheme()}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			form := NewForm("Preset").SetTheme(test.theme).Add(NewTextField("name", "Name").SetValue("value"))
			form.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
			assert.True(t, strings.HasPrefix(plainFrameLines(form.Frame(30))[0], test.prefix))
		})
	}
}
