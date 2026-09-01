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

func TestTextFieldEditsGraphemeClusters(t *testing.T) {
	t.Parallel()
	field := NewTextField("name", "Name").SetValue("a\u0301👨‍👩‍👧‍👦")
	assert.Equal(t, 2, field.cursor)
	assert.Equal(t, 3, displayWidth(field.Value().(string)))

	field.handle(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
	assert.Equal(t, "a\u0301", field.Value())
	field.handle(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
	assert.Empty(t, field.Value())
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

func TestSelectUsesNoDefaultIndicators(t *testing.T) {
	t.Parallel()
	field := NewSelectField("choice", "Choice", NewChoice("a", "Alpha"), NewChoice("b", "Beta"))
	plain := strings.Join(plainFrameLines(fieldFrame(field, 36)), "\n")
	assert.Contains(t, plain, "Alpha")
	assert.NotContains(t, plain, "◇")
	assert.NotContains(t, plain, "·")
	assert.NotContains(t, plain, "✓ Alpha")
	field.handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	plain = strings.Join(plainFrameLines(fieldFrame(field, 36)), "\n")
	assert.Contains(t, plain, "Alpha")
	assert.NotContains(t, plain, "✓ Alpha")
}

func TestSelectChoiceToneUsesSemanticThemeStyle(t *testing.T) {
	t.Parallel()
	theme := RoundedInlineTheme()
	field := NewSelectField("choice", "Choice",
		NewChoice("create", "+ Create source").WithTone(ChoiceToneSuccess),
	)
	frame := NewFrame(36, field.height(true))
	field.draw(frame, 0, 36, true, theme)

	_, style, _ := frame.Cell(2, 2)
	foreground, _, _ := style.Decompose()
	wantForeground, _, _ := theme.Success.Decompose()
	assert.Equal(t, wantForeground, foreground)
}

func TestMultiSelectSetSelectedValues(t *testing.T) {
	t.Parallel()
	field := NewMultiSelectField("targets", "Targets",
		NewChoice("linux", "Linux"),
		NewChoice("darwin", "macOS"),
	).SetSelectedValues("darwin", "missing")
	assert.Equal(t, []string{"darwin"}, field.Value())
	field.SetSelectedValues("linux")
	assert.Equal(t, []string{"linux"}, field.Value())
}

func TestFilterableSelectNarrowsAndSelectsChoices(t *testing.T) {
	t.Parallel()
	field := NewSelectField("connector", "Connector",
		Choice{Value: "sample", Label: "Sample"},
		Choice{Value: "postgres", Label: "PostgreSQL", Description: "Database connector"},
		Choice{Value: "stdout", Label: "Standard output"},
	).Filterable(true).Required()
	for _, value := range "post" {
		field.handle(tcell.NewEventKey(tcell.KeyRune, value, tcell.ModNone))
	}
	assert.Equal(t, []int{1}, field.visibleChoices())
	field.handle(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	assert.Equal(t, "postgres", field.Value())
	plain := strings.Join(plainFrameLines(fieldFrame(field, 48)), "\n")
	assert.Contains(t, plain, "/ post")
	assert.Contains(t, plain, "PostgreSQL Database connector")
	assert.NotContains(t, plain, "Sample")
}

func TestFilterableSelectBackspaceRestoresChoices(t *testing.T) {
	t.Parallel()
	field := NewSelectField("connector", "Connector",
		NewChoice("sample", "Sample"),
		NewChoice("postgres", "PostgreSQL"),
	).Filterable(true)
	for _, value := range "post" {
		field.handle(tcell.NewEventKey(tcell.KeyRune, value, tcell.ModNone))
	}
	for range 4 {
		field.handle(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
	}
	assert.Len(t, field.visibleChoices(), 2)
	assert.Equal(t, 1, field.cursor)
}

func TestSelectFieldsSupportVimMovementWithoutCapturingSearchInput(t *testing.T) {
	t.Parallel()
	choices := []Choice{NewChoice("a", "Alpha"), NewChoice("b", "Beta"), NewChoice("c", "Gamma")}
	field := NewSelectField("choice", "Choice", choices...)
	for _, movement := range []struct {
		key  rune
		want int
	}{{'j', 1}, {'l', 2}, {'k', 1}, {'h', 0}} {
		field.handle(tcell.NewEventKey(tcell.KeyRune, movement.key, tcell.ModNone))
		assert.Equal(t, movement.want, field.cursor)
	}

	filterable := NewSelectField("choice", "Choice", choices...).Filterable(true)
	for _, key := range "hjkl" {
		filterable.handle(tcell.NewEventKey(tcell.KeyRune, key, tcell.ModNone))
	}
	assert.Equal(t, "hjkl", strings.Join(filterable.query, ""))
}

func TestMultiSelectSupportsVimMovement(t *testing.T) {
	t.Parallel()
	field := NewMultiSelectField("choice", "Choice",
		NewChoice("a", "Alpha"),
		NewChoice("b", "Beta"),
		NewChoice("c", "Gamma"),
	)
	for _, movement := range []struct {
		key  rune
		want int
	}{{'j', 1}, {'l', 2}, {'k', 1}, {'h', 0}} {
		field.handle(tcell.NewEventKey(tcell.KeyRune, movement.key, tcell.ModNone))
		assert.Equal(t, movement.want, field.cursor)
	}
}

func fieldFrame(field *SelectField, width int) *Frame {
	frame := NewFrame(width, field.height(true))
	field.draw(frame, 0, width, true, RoundedInlineTheme())
	return frame
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

func TestFormRendersStepperHeader(t *testing.T) {
	t.Parallel()
	stepper := NewStepper("Configure", WithStepOrientation(StepHorizontal))
	require.NoError(t, stepper.Add("source", "Source"))
	require.NoError(t, stepper.Add("sink", "Sink"))
	require.NoError(t, stepper.Activate("source"))
	form := NewForm("Connection").
		SetHeader(stepper).
		SetHeaderGap(1).
		Add(NewTextField("name", "Name"))
	plain := strings.Join(plainFrameLines(form.Frame(48)), "\n")
	assert.Contains(t, plain, "Configure")
	assert.Contains(t, plain, "Source")
	assert.Contains(t, plain, "Connection")
}

func TestFormReportsFocusChanges(t *testing.T) {
	t.Parallel()
	form := NewForm("Focus").Add(
		NewTextField("one", "One"),
		NewTextField("two", "Two"),
		NewTextField("three", "Three"),
	)
	var focused []string
	form.OnFocusChange(func(_ int, field FormField) {
		focused = append(focused, field.ID())
	})

	form.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModShift))

	assert.Equal(t, []string{"one", "two", "three", "two"}, focused)
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

func TestFormRejectsMalformedFields(t *testing.T) {
	t.Parallel()
	assert.ErrorIs(t, NewForm("Empty").validateStructure(), ErrFormHasNoFields)
	assert.ErrorIs(t, NewForm("Nil").Add((*TextField)(nil)).validateStructure(), ErrNilFormField)
	assert.ErrorIs(t, NewForm("Blank").Add(NewTextField("", "Name")).validateStructure(), ErrInvalidFieldID)
	assert.ErrorIs(t, NewForm("Duplicate").Add(
		NewTextField("name", "Name"),
		NewSelectField("name", "Other", NewChoice("x", "X")),
	).validateStructure(), ErrDuplicateField)
}
