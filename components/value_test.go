package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTextField_ValueProvider tests TextField's ValueProvider[string] implementation.
func TestTextField_ValueProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		setValue string
		wantVal  string
		wantHas  bool
	}{
		{
			name:     "empty initial",
			initial:  "",
			setValue: "",
			wantVal:  "",
			wantHas:  false,
		},
		{
			name:     "set value",
			initial:  "",
			setValue: "hello",
			wantVal:  "hello",
			wantHas:  true,
		},
		{
			name:     "overwrite value",
			initial:  "old",
			setValue: "new",
			wantVal:  "new",
			wantHas:  true,
		},
		{
			name:     "clear to empty",
			initial:  "value",
			setValue: "",
			wantVal:  "",
			wantHas:  false,
		},
		{
			name:     "unicode value",
			initial:  "",
			setValue: "こんにちは",
			wantVal:  "こんにちは",
			wantHas:  true,
		},
		{
			name:     "whitespace only",
			initial:  "",
			setValue: "   ",
			wantVal:  "   ",
			wantHas:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewTextField("test")
			if tt.initial != "" {
				field.SetValue(tt.initial)
			}
			field.SetValue(tt.setValue)

			assert.Equal(t, tt.wantVal, field.Value())
			assert.Equal(t, tt.wantVal, field.GetValue())
			assert.Equal(t, tt.wantHas, field.HasValue())
		})
	}
}

// TestTextField_Clear tests TextField Clear method.
func TestTextField_Clear(t *testing.T) {
	field := NewTextField("test").SetValue("hello")
	require.True(t, field.HasValue())

	field.Clear()

	assert.False(t, field.HasValue())
	assert.Equal(t, "", field.Value())
}

// TestTextField_GetName tests TextField name accessor.
func TestTextField_GetName(t *testing.T) {
	field := NewTextField("my-field")
	assert.Equal(t, "my-field", field.GetName())
}

// TestCheckbox_ValueProvider tests Checkbox's ValueProvider[bool] implementation.
func TestCheckbox_ValueProvider(t *testing.T) {
	tests := []struct {
		name    string
		checked bool
		toggle  bool
		wantVal bool
	}{
		{
			name:    "default unchecked",
			checked: false,
			toggle:  false,
			wantVal: false,
		},
		{
			name:    "set checked",
			checked: true,
			toggle:  false,
			wantVal: true,
		},
		{
			name:    "toggle from unchecked",
			checked: false,
			toggle:  true,
			wantVal: true,
		},
		{
			name:    "toggle from checked",
			checked: true,
			toggle:  true,
			wantVal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := NewCheckbox("test").SetChecked(tt.checked)
			if tt.toggle {
				cb.Toggle()
			}
			assert.Equal(t, tt.wantVal, cb.Value())
			assert.Equal(t, tt.wantVal, cb.Checked())
			assert.Equal(t, tt.wantVal, cb.GetValue())
		})
	}
}

// TestCheckbox_HasValue tests that Checkbox always reports having a value.
func TestCheckbox_HasValue(t *testing.T) {
	cb := NewCheckbox("test")
	assert.True(t, cb.HasValue(), "checkbox should always have a value")

	cb.SetChecked(true)
	assert.True(t, cb.HasValue())

	cb.SetChecked(false)
	assert.True(t, cb.HasValue())
}

// TestCheckbox_Clear tests Checkbox Clear method.
func TestCheckbox_Clear(t *testing.T) {
	cb := NewCheckbox("test").SetChecked(true)
	require.True(t, cb.Value())

	cb.Clear()
	assert.False(t, cb.Value())
}

// TestSelect_ValueProvider tests Select's IndexedValueProvider[string] implementation.
func TestSelect_ValueProvider(t *testing.T) {
	tests := []struct {
		name       string
		options    []string
		setDefault string
		setIndex   int
		wantVal    string
		wantIdx    int
		wantHas    bool
	}{
		{
			name:     "no selection",
			options:  []string{"a", "b", "c"},
			setIndex: -2, // don't set index
			wantVal:  "",
			wantIdx:  -1,
			wantHas:  false,
		},
		{
			name:       "explicit default",
			options:    []string{"a", "b", "c"},
			setDefault: "b",
			setIndex:   -2,
			wantVal:    "b",
			wantIdx:    1,
			wantHas:    true,
		},
		{
			name:     "set by index",
			options:  []string{"a", "b", "c"},
			setIndex: 2,
			wantVal:  "c",
			wantIdx:  2,
			wantHas:  true,
		},
		{
			name:     "first option by index",
			options:  []string{"first", "second"},
			setIndex: 0,
			wantVal:  "first",
			wantIdx:  0,
			wantHas:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel := NewSelect("test").SetOptions(tt.options)
			if tt.setDefault != "" {
				sel.SetDefault(tt.setDefault)
			}
			if tt.setIndex >= -1 {
				sel.SetSelected(tt.setIndex)
			}

			assert.Equal(t, tt.wantVal, sel.Value())
			assert.Equal(t, tt.wantVal, sel.GetValue())
			assert.Equal(t, tt.wantIdx, sel.SelectedIndex())
			assert.Equal(t, tt.wantHas, sel.HasValue())
		})
	}
}

// TestSelect_OptionsWithValues tests Select with custom label/value pairs.
func TestSelect_OptionsWithValues(t *testing.T) {
	sel := NewSelect("status").
		SetOptionsWithValues([]SelectOption{
			{Label: "Active", Value: "active"},
			{Label: "Inactive", Value: "inactive"},
			{Label: "Pending", Value: "pending"},
		}).
		SetDefault("active")

	assert.Equal(t, "active", sel.Value())
	assert.Equal(t, 0, sel.SelectedIndex())

	opt := sel.SelectedOption()
	assert.Equal(t, "Active", opt.Label)
	assert.Equal(t, "active", opt.Value)
}

// TestSelect_SetSelectedIndex tests Select SetSelectedIndex with error handling.
func TestSelect_SetSelectedIndex(t *testing.T) {
	sel := NewSelect("test").SetOptions([]string{"a", "b", "c"})

	// Valid indices
	err := sel.SetSelectedIndex(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, sel.SelectedIndex())

	err = sel.SetSelectedIndex(2)
	assert.NoError(t, err)
	assert.Equal(t, 2, sel.SelectedIndex())

	// Invalid indices
	err = sel.SetSelectedIndex(10)
	assert.Error(t, err)

	err = sel.SetSelectedIndex(-5)
	assert.Error(t, err)

	// -1 clears selection
	err = sel.SetSelectedIndex(-1)
	assert.NoError(t, err)
	assert.Equal(t, -1, sel.SelectedIndex())
}

// TestSelect_SetSelectedValue tests Select SetSelectedValue.
func TestSelect_SetSelectedValue(t *testing.T) {
	sel := NewSelect("test").SetOptionsWithValues([]SelectOption{
		{Label: "One", Value: "1"},
		{Label: "Two", Value: "2"},
	})

	err := sel.SetSelectedValue("2")
	assert.NoError(t, err)
	assert.Equal(t, "2", sel.Value())

	err = sel.SetSelectedValue("nonexistent")
	assert.Error(t, err)
}

// TestSelect_Clear tests Select Clear method.
func TestSelect_Clear(t *testing.T) {
	sel := NewSelect("test").SetOptions([]string{"a", "b"}).SetDefault("a")
	require.True(t, sel.HasValue())

	sel.Clear()
	assert.False(t, sel.HasValue())
	assert.Equal(t, -1, sel.SelectedIndex())
}

// TestRadioGroup_ValueProvider tests RadioGroup's IndexedValueProvider[string] implementation.
func TestRadioGroup_ValueProvider(t *testing.T) {
	tests := []struct {
		name     string
		options  []string
		selected int
		wantVal  string
		wantHas  bool
	}{
		{
			name:     "no selection",
			options:  []string{"opt1", "opt2"},
			selected: -1,
			wantVal:  "",
			wantHas:  false,
		},
		{
			name:     "first selected",
			options:  []string{"opt1", "opt2"},
			selected: 0,
			wantVal:  "opt1",
			wantHas:  true,
		},
		{
			name:     "second selected",
			options:  []string{"opt1", "opt2"},
			selected: 1,
			wantVal:  "opt2",
			wantHas:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rg := NewRadioGroup("test").SetOptions(tt.options)
			if tt.selected >= 0 {
				rg.SetSelected(tt.selected)
			}

			assert.Equal(t, tt.wantVal, rg.Value())
			assert.Equal(t, tt.wantVal, rg.GetValue())
			assert.Equal(t, tt.wantHas, rg.HasValue())
			if tt.selected >= 0 {
				assert.Equal(t, tt.selected, rg.SelectedIndex())
			}
		})
	}
}

// TestRadioGroup_SetSelectedValue tests RadioGroup SetSelectedValue.
func TestRadioGroup_SetSelectedValue(t *testing.T) {
	rg := NewRadioGroup("test").SetOptions([]string{"red", "green", "blue"})

	err := rg.SetSelectedValue("green")
	assert.NoError(t, err)
	assert.Equal(t, "green", rg.Value())
	assert.Equal(t, 1, rg.SelectedIndex())

	err = rg.SetSelectedValue("yellow")
	assert.Error(t, err)
}

// TestRadioGroup_Clear tests RadioGroup Clear method.
func TestRadioGroup_Clear(t *testing.T) {
	rg := NewRadioGroup("test").SetOptions([]string{"a", "b"}).SetSelected(0)
	require.True(t, rg.HasValue())

	rg.Clear()
	assert.False(t, rg.HasValue())
	assert.Equal(t, -1, rg.SelectedIndex())
}

// TestMultiSelect_ValueProvider tests MultiSelect's MultiValueProvider[string] implementation.
func TestMultiSelect_ValueProvider(t *testing.T) {
	tests := []struct {
		name        string
		options     []string
		selected    []int
		wantValues  []string
		wantIndices []int
		wantHas     bool
	}{
		{
			name:        "no selection",
			options:     []string{"a", "b", "c"},
			selected:    nil,
			wantValues:  nil,
			wantIndices: []int{},
			wantHas:     false,
		},
		{
			name:        "single selection",
			options:     []string{"a", "b", "c"},
			selected:    []int{1},
			wantValues:  []string{"b"},
			wantIndices: []int{1},
			wantHas:     true,
		},
		{
			name:        "multiple selections",
			options:     []string{"a", "b", "c"},
			selected:    []int{0, 2},
			wantValues:  []string{"a", "c"},
			wantIndices: []int{0, 2},
			wantHas:     true,
		},
		{
			name:        "all selected",
			options:     []string{"x", "y"},
			selected:    []int{0, 1},
			wantValues:  []string{"x", "y"},
			wantIndices: []int{0, 1},
			wantHas:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMultiSelect("test").SetOptions(tt.options)
			if tt.selected != nil {
				ms.SetSelected(tt.selected)
			}

			assert.Equal(t, tt.wantValues, ms.Values())
			assert.Equal(t, tt.wantIndices, ms.SelectedIndices())
			assert.Equal(t, tt.wantHas, ms.HasValue())
		})
	}
}

// TestMultiSelect_SetSelectedValues tests MultiSelect SetSelectedValues.
func TestMultiSelect_SetSelectedValues(t *testing.T) {
	ms := NewMultiSelect("test").SetOptionsWithValues([]SelectOption{
		{Label: "One", Value: "1"},
		{Label: "Two", Value: "2"},
		{Label: "Three", Value: "3"},
	})

	err := ms.SetSelectedValues([]string{"1", "3"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "3"}, ms.Values())

	err = ms.SetSelectedValues([]string{"nonexistent"})
	assert.Error(t, err)
}

// TestMultiSelect_SetSelectedIndices tests MultiSelect SetSelectedIndices with error handling.
func TestMultiSelect_SetSelectedIndices(t *testing.T) {
	ms := NewMultiSelect("test").SetOptions([]string{"a", "b", "c"})

	err := ms.SetSelectedIndices([]int{0, 2})
	assert.NoError(t, err)
	assert.Equal(t, []int{0, 2}, ms.SelectedIndices())

	err = ms.SetSelectedIndices([]int{10})
	assert.Error(t, err)

	err = ms.SetSelectedIndices([]int{-1})
	assert.Error(t, err)
}

// TestMultiSelect_Clear tests MultiSelect Clear method.
func TestMultiSelect_Clear(t *testing.T) {
	ms := NewMultiSelect("test").SetOptions([]string{"a", "b"}).SetSelected([]int{0, 1})
	require.True(t, ms.HasValue())

	ms.Clear()
	assert.False(t, ms.HasValue())
	assert.Equal(t, []int{}, ms.SelectedIndices())
}

// TestMultiSelect_SelectedOptions tests MultiSelect SelectedOptions method.
func TestMultiSelect_SelectedOptions(t *testing.T) {
	ms := NewMultiSelect("test").SetOptionsWithValues([]SelectOption{
		{Label: "Apples", Value: "apple"},
		{Label: "Bananas", Value: "banana"},
		{Label: "Cherries", Value: "cherry"},
	}).SetSelected([]int{0, 2})

	opts := ms.SelectedOptions()
	require.Len(t, opts, 2)
	assert.Equal(t, "Apples", opts[0].Label)
	assert.Equal(t, "apple", opts[0].Value)
	assert.Equal(t, "Cherries", opts[1].Label)
	assert.Equal(t, "cherry", opts[1].Value)
}

// TestTextArea_ValueProvider tests TextArea's ValueProvider[string] implementation.
func TestTextArea_ValueProvider(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		wantVal  string
		wantHas  bool
	}{
		{
			name:     "empty",
			setValue: "",
			wantVal:  "",
			wantHas:  false,
		},
		{
			name:     "single line",
			setValue: "hello",
			wantVal:  "hello",
			wantHas:  true,
		},
		{
			name:     "multiple lines",
			setValue: "line1\nline2\nline3",
			wantVal:  "line1\nline2\nline3",
			wantHas:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := NewTextArea("test")
			ta.SetValue(tt.setValue)

			assert.Equal(t, tt.wantVal, ta.Value())
			assert.Equal(t, tt.wantVal, ta.GetValue())
			assert.Equal(t, tt.wantHas, ta.HasValue())
		})
	}
}

// TestTextArea_Clear tests TextArea Clear method.
func TestTextArea_Clear(t *testing.T) {
	ta := NewTextArea("test").SetValue("some text")
	require.True(t, ta.HasValue())

	ta.Clear()
	assert.False(t, ta.HasValue())
	assert.Equal(t, "", ta.Value())
}
