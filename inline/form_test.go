package inline

import (
	"errors"
	"fmt"
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

func TestSelectFieldBoundsVisibleChoicesAndReportsRange(t *testing.T) {
	t.Parallel()
	field := NewSelectField("choice", "Choice", numberedChoices(6)...).MaxVisibleChoices(3)
	frame := fieldFrame(field, 36)
	plain := strings.Join(plainFrameLines(frame), "\n")

	assert.Equal(t, 7, frame.Height())
	assert.Contains(t, plain, "Choice 1")
	assert.Contains(t, plain, "Choice 3")
	assert.NotContains(t, plain, "Choice 4")
	assert.Contains(t, plain, "1–3 of 6")

	field.MaxVisibleChoices(0)
	frame = fieldFrame(field, 36)
	plain = strings.Join(plainFrameLines(frame), "\n")
	assert.Equal(t, 9, frame.Height())
	assert.Contains(t, plain, "Choice 6")
	assert.NotContains(t, plain, "of 6")

	field.MaxVisibleChoices(-2)
	assert.Equal(t, 9, field.height(true))
}

func TestBoundedSelectScrollsBothDirectionsAndSkipsDisabledChoices(t *testing.T) {
	t.Parallel()
	choices := numberedChoices(6)
	choices[1].Disabled = true
	field := NewSelectField("choice", "Choice", choices...).MaxVisibleChoices(2)
	stableHeight := field.height(true)

	field.handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	assert.Equal(t, 2, field.cursor)
	assert.Equal(t, 1, field.viewport.offset)
	plain := strings.Join(plainFrameLines(fieldFrame(field, 36)), "\n")
	assert.Contains(t, plain, "Choice 2")
	assert.Contains(t, plain, "Choice 3")
	assert.NotContains(t, plain, "Choice 1")

	field.handle(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	assert.Equal(t, 0, field.cursor)
	assert.Zero(t, field.viewport.offset)
	field.handle(tcell.NewEventKey(tcell.KeyRune, 'k', tcell.ModNone))
	assert.Equal(t, 5, field.cursor)
	assert.Equal(t, 4, field.viewport.offset)
	plain = strings.Join(plainFrameLines(fieldFrame(field, 36)), "\n")
	assert.Contains(t, plain, "Choice 6")
	assert.Contains(t, plain, "5–6 of 6")

	field.handle(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))
	assert.Equal(t, 0, field.cursor)
	assert.Zero(t, field.viewport.offset)
	assert.Equal(t, stableHeight, field.height(true))
}

func TestBoundedMultiSelectPreservesOffscreenSelections(t *testing.T) {
	t.Parallel()
	field := NewMultiSelectField("choices", "Choices", numberedChoices(5)...).MaxVisibleChoices(2)
	stableHeight := field.height(true)
	field.handle(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	field.handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	field.handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	field.handle(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	field.handle(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	field.handle(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))

	assert.Equal(t, []string{"choice-1", "choice-3"}, field.Value())
	assert.Equal(t, 3, field.viewport.offset)
	plain := strings.Join(plainFrameLines(multiSelectFieldFrame(field, 36)), "\n")
	assert.NotContains(t, plain, "Choice 1")
	assert.Contains(t, plain, "Choice 5")
	assert.Contains(t, plain, "4–5 of 5")
	assert.Equal(t, stableHeight, field.height(true))

	for range 4 {
		field.handle(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	}
	assert.Zero(t, field.viewport.offset)
	assert.Equal(t, []string{"choice-1", "choice-3"}, field.Value())
}

func TestBoundedMultiSelectSkipsDisabledChoices(t *testing.T) {
	t.Parallel()
	choices := numberedChoices(4)
	choices[1].Disabled = true
	field := NewMultiSelectField("choices", "Choices", choices...).MaxVisibleChoices(2)
	field.handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	assert.Equal(t, 2, field.cursor)
	field.handle(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	assert.Equal(t, []string{"choice-3"}, field.Value())

	frame := multiSelectFieldFrame(field, 36)
	_, disabledStyle, _ := frame.Cell(2, 2)
	wantForeground, _, _ := RoundedInlineTheme().Muted.Decompose()
	foreground, _, _ := disabledStyle.Decompose()
	assert.Equal(t, wantForeground, foreground)
}

func TestFilterableSelectClampsViewportAsMatchesChange(t *testing.T) {
	t.Parallel()
	field := NewSelectField("choice", "Choice",
		NewChoice("alpha", "Alpha"),
		NewChoice("alpine", "Alpine"),
		NewChoice("beta", "Beta"),
		NewChoice("blue", "Blue"),
		NewChoice("blush", "Blush"),
	).MaxVisibleChoices(2).Filterable(true)
	for range 4 {
		field.handle(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	}
	assert.Equal(t, 3, field.viewport.offset)

	for _, value := range "blue" {
		field.handle(tcell.NewEventKey(tcell.KeyRune, value, tcell.ModNone))
	}
	assert.Equal(t, []int{3}, field.visibleChoices())
	assert.Equal(t, 3, field.cursor)
	assert.Zero(t, field.viewport.offset)
	plain := strings.Join(plainFrameLines(fieldFrame(field, 36)), "\n")
	assert.Contains(t, plain, "Blue")
	assert.NotContains(t, plain, "Blush")
	assert.NotContains(t, plain, " of ")

	for range 4 {
		field.handle(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
	}
	assert.Len(t, field.visibleChoices(), 5)
	assert.Equal(t, 2, field.viewport.offset)
	plain = strings.Join(plainFrameLines(fieldFrame(field, 36)), "\n")
	assert.Contains(t, plain, "3–4 of 5")
	assert.Contains(t, plain, "Blue")
	assert.NotContains(t, plain, "Alpha")
}

func TestBoundedChoiceRenderingSupportsThemesAndDescriptions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		theme InlineTheme
	}{{"boxed", RoundedInlineTheme()}, {"column", ASCIIInlineTheme()}} {
		t.Run(test.name, func(t *testing.T) {
			field := NewSelectField("choice", "Choice",
				Choice{Value: "a", Label: "Alpha", Description: "First column"},
				Choice{Value: "b", Label: "Beta", Description: "Second column"},
				Choice{Value: "c", Label: "Gamma", Description: "Third column"},
			).MaxVisibleChoices(2)
			frame := NewFrame(44, field.height(true))
			field.draw(frame, 0, 44, true, test.theme)
			plain := strings.Join(plainFrameLines(frame), "\n")
			assert.Contains(t, plain, test.theme.Borders.TopLeft)
			assert.Contains(t, plain, "Alpha First column")
			assert.Contains(t, plain, "1–2 of 3")
			assert.NotContains(t, plain, "Gamma")
		})
	}
}

func fieldFrame(field *SelectField, width int) *Frame {
	frame := NewFrame(width, field.height(true))
	field.draw(frame, 0, width, true, RoundedInlineTheme())
	return frame
}

func multiSelectFieldFrame(field *MultiSelectField, width int) *Frame {
	frame := NewFrame(width, field.height(true))
	field.draw(frame, 0, width, true, RoundedInlineTheme())
	return frame
}

func numberedChoices(count int) []Choice {
	choices := make([]Choice, count)
	for index := range choices {
		choices[index] = NewChoice(fmt.Sprintf("choice-%d", index+1), fmt.Sprintf("Choice %d", index+1))
	}
	return choices
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

func TestFormBoundaryTabAdvancesAndSubmits(t *testing.T) {
	t.Parallel()
	form := NewForm("Boundary").NavigateAtBoundaries(true).Add(
		NewTextField("first", "First"),
		NewTextField("second", "Second"),
	)
	form.HandleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	assert.Equal(t, 1, form.focus)
	assert.False(t, form.Done())
	form.HandleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))

	require.True(t, form.Done())
	assert.True(t, form.submitted)
	assert.Equal(t, "a", form.Result()["first"])
	assert.Equal(t, "b", form.Result()["second"])
}

func TestFormBoundaryShiftTabSignalsPrevious(t *testing.T) {
	t.Parallel()
	form := NewForm("Boundary").NavigateAtBoundaries(true).Add(NewTextField("name", "Name"))
	form.HandleKey(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModShift))
	assert.True(t, form.Done())
	assert.True(t, form.previous)
	assert.False(t, form.cancelled)
}

func TestFormBoundaryNavigationAcceptsShiftModifiedTab(t *testing.T) {
	t.Parallel()
	form := NewForm("Boundary").NavigateAtBoundaries(true).Add(NewTextField("name", "Name"))
	form.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModShift))
	assert.True(t, form.previous)
}

func TestFormNavigationDefaultsStillWrap(t *testing.T) {
	t.Parallel()
	form := NewForm("Default").Add(NewTextField("first", "First"), NewTextField("second", "Second"))
	form.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	form.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	assert.Zero(t, form.focus)
	assert.False(t, form.Done())

	form.HandleKey(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModShift))
	assert.Equal(t, 1, form.focus)
	assert.False(t, form.Done())
	assert.False(t, form.previous)
}

func TestFormContextualAndConfigurableKeyHints(t *testing.T) {
	t.Parallel()
	selectForm := NewForm("Select").Add(NewSelectField("choice", "Choice", numberedChoices(2)...))
	plain := strings.Join(plainFrameLines(selectForm.Frame(80)), "\n")
	assert.Contains(t, plain, "[↑/↓] Move")
	assert.NotContains(t, plain, "[Space]")
	assert.Contains(t, plain, "[Enter] Submit")
	assert.Contains(t, plain, "[Esc] Cancel")

	multiForm := NewForm("Multi").NavigateAtBoundaries(true).Add(
		NewMultiSelectField("choices", "Choices", numberedChoices(2)...),
		NewTextField("name", "Name"),
	)
	plain = strings.Join(plainFrameLines(multiForm.Frame(120)), "\n")
	assert.Contains(t, plain, "[↑/↓] Move")
	assert.Contains(t, plain, "[Space] Toggle")
	assert.Contains(t, plain, "[Tab] Next")
	assert.Contains(t, plain, "[Shift+Tab] Previous")
	assert.Contains(t, plain, "[Enter] Next field")

	multiForm.SetKeyHints(FormKeyHint{Key: "?", Label: "Help"})
	plain = strings.Join(plainFrameLines(multiForm.Frame(60)), "\n")
	assert.Contains(t, plain, "[?] Help")
	assert.NotContains(t, plain, "[Esc]")
	multiForm.SetKeyHints()
	plain = strings.Join(plainFrameLines(multiForm.Frame(120)), "\n")
	assert.Contains(t, plain, "[Esc] Cancel")
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
	assert.Contains(t, plain, "[Enter]")
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
