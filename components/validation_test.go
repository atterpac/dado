package components

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/validators"
)

// TestTextField_Validation tests TextField validation functionality.
func TestTextField_Validation(t *testing.T) {
	tests := []struct {
		name       string
		validator  func(string) error
		value      string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "no validator passes",
			validator: nil,
			value:     "anything",
			wantErr:   false,
		},
		{
			name: "required fails on empty",
			validator: func(v string) error {
				return validators.Required()(v)
			},
			value:      "",
			wantErr:    true,
			wantErrMsg: "required",
		},
		{
			name: "required passes on value",
			validator: func(v string) error {
				return validators.Required()(v)
			},
			value:   "hello",
			wantErr: false,
		},
		{
			name: "email fails on invalid",
			validator: func(v string) error {
				return validators.Email()(v)
			},
			value:      "not-an-email",
			wantErr:    true,
			wantErrMsg: "email",
		},
		{
			name: "email passes on valid",
			validator: func(v string) error {
				return validators.Email()(v)
			},
			value:   "test@example.com",
			wantErr: false,
		},
		{
			name: "custom validator",
			validator: func(v string) error {
				if len(v) < 3 {
					return errors.New("too short")
				}
				return nil
			},
			value:      "ab",
			wantErr:    true,
			wantErrMsg: "too short",
		},
		{
			name: "custom validator passes",
			validator: func(v string) error {
				if len(v) < 3 {
					return errors.New("too short")
				}
				return nil
			},
			value:   "abc",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewTextField("test")
			if tt.validator != nil {
				field.SetValidator(tt.validator)
			}
			field.SetValue(tt.value)

			err := field.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, field.HasError())
				assert.Contains(t, field.GetError(), tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.False(t, field.HasError())
				assert.Empty(t, field.GetError())
			}
		})
	}
}

// TestTextField_SetValidators tests multiple validators.
func TestTextField_SetValidators(t *testing.T) {
	field := NewTextField("email").
		SetValidators(validators.Required(), validators.Email())

	// Empty fails on required
	field.SetValue("")
	err := field.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")

	// Invalid email fails on email validation
	field.SetValue("not-email")
	err = field.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email")

	// Valid passes both
	field.SetValue("test@example.com")
	err = field.Validate()
	assert.NoError(t, err)
}

// TestTextField_ValidationOnValueChange tests that validation runs when value changes.
func TestTextField_ValidationOnValueChange(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			if v == "invalid" {
				return errors.New("invalid value")
			}
			return nil
		})

	// Setting invalid value triggers validation
	field.SetValue("invalid")
	assert.True(t, field.HasError())
	assert.Equal(t, "invalid value", field.GetError())

	// Setting valid value clears error
	field.SetValue("valid")
	assert.False(t, field.HasError())
	assert.Empty(t, field.GetError())
}

// TestTextField_ValidateMethod tests explicit Validate() call.
func TestTextField_ValidateMethod(t *testing.T) {
	field := NewTextField("test")

	// No validator - always passes
	err := field.Validate()
	assert.NoError(t, err)

	// With validator
	field.SetValidator(func(v string) error {
		if v == "" {
			return errors.New("required")
		}
		return nil
	})

	// Empty value fails
	err = field.Validate()
	assert.Error(t, err)

	// Non-empty passes
	field.SetValue("hello")
	err = field.Validate()
	assert.NoError(t, err)
}

// TestForm_Validate tests form-level validation.
func TestForm_Validate(t *testing.T) {
	form := NewForm()

	field1 := NewTextField("name").SetValidator(func(v string) error {
		if v == "" {
			return errors.New("name required")
		}
		return nil
	})

	field2 := NewTextField("email").SetValidator(func(v string) error {
		if v == "" {
			return errors.New("email required")
		}
		return nil
	})

	form.AddField(field1)
	form.AddField(field2)

	// Both empty - fails on first
	err := form.Validate()
	assert.Error(t, err)

	// Set first field
	field1.SetValue("John")
	err = form.Validate()
	assert.Error(t, err)

	// Set second field
	field2.SetValue("john@example.com")
	err = form.Validate()
	assert.NoError(t, err)
}

// TestForm_ValidateAll tests form ValidateAll method.
func TestForm_ValidateAll(t *testing.T) {
	form := NewForm()

	field1 := NewTextField("name").SetValidator(func(v string) error {
		if v == "" {
			return errors.New("name required")
		}
		return nil
	})

	field2 := NewTextField("email").SetValidator(func(v string) error {
		if v == "" {
			return errors.New("email required")
		}
		return nil
	})

	form.AddField(field1)
	form.AddField(field2)

	// Both empty - returns all errors
	result := form.ValidateAll()
	assert.True(t, result.HasErrors())
	assert.Len(t, result.Errors, 2)
	assert.Equal(t, "name", result.Errors[0].Field)
	assert.Equal(t, "email", result.Errors[1].Field)
}

// TestForm_IsValid tests form IsValid helper.
func TestForm_IsValid(t *testing.T) {
	form := NewForm()

	field := NewTextField("test").SetValidator(func(v string) error {
		if v == "" {
			return errors.New("required")
		}
		return nil
	})
	form.AddField(field)

	assert.False(t, form.IsValid())

	field.SetValue("value")
	assert.True(t, form.IsValid())
}

// TestValidators_MinLength tests MinLength validator.
func TestValidators_MinLength(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.MinLength(5)(v)
		})

	field.SetValue("abc")
	assert.Error(t, field.Validate())

	field.SetValue("abcde")
	assert.NoError(t, field.Validate())

	field.SetValue("abcdef")
	assert.NoError(t, field.Validate())
}

// TestValidators_MaxLength tests MaxLength validator.
func TestValidators_MaxLength(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.MaxLength(5)(v)
		})

	field.SetValue("abc")
	assert.NoError(t, field.Validate())

	field.SetValue("abcde")
	assert.NoError(t, field.Validate())

	field.SetValue("abcdef")
	assert.Error(t, field.Validate())
}

// TestValidators_Pattern tests Pattern validator.
func TestValidators_Pattern(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.Pattern(`^\d{3}-\d{4}$`)(v)
		})

	field.SetValue("123-4567")
	assert.NoError(t, field.Validate())

	field.SetValue("12-4567")
	assert.Error(t, field.Validate())

	field.SetValue("abc-defg")
	assert.Error(t, field.Validate())
}

// TestValidators_URL tests URL validator.
func TestValidators_URL(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.URL()(v)
		})

	field.SetValue("https://example.com")
	assert.NoError(t, field.Validate())

	field.SetValue("http://localhost:8080/path")
	assert.NoError(t, field.Validate())

	field.SetValue("not-a-url")
	assert.Error(t, field.Validate())
}

// TestValidators_Alphanumeric tests Alphanumeric validator.
func TestValidators_Alphanumeric(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.Alphanumeric()(v)
		})

	field.SetValue("abc123")
	assert.NoError(t, field.Validate())

	field.SetValue("ABC")
	assert.NoError(t, field.Validate())

	field.SetValue("abc-123")
	assert.Error(t, field.Validate())

	field.SetValue("abc 123")
	assert.Error(t, field.Validate())
}

// TestValidators_NoWhitespace tests NoWhitespace validator.
func TestValidators_NoWhitespace(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.NoWhitespace()(v)
		})

	field.SetValue("nospaces")
	assert.NoError(t, field.Validate())

	field.SetValue("has space")
	assert.Error(t, field.Validate())

	field.SetValue("has\ttab")
	assert.Error(t, field.Validate())

	field.SetValue("has\nnewline")
	assert.Error(t, field.Validate())
}

// TestValidators_OneOf tests OneOf validator.
func TestValidators_OneOf(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.OneOf("red", "green", "blue")(v)
		})

	field.SetValue("red")
	assert.NoError(t, field.Validate())

	field.SetValue("green")
	assert.NoError(t, field.Validate())

	field.SetValue("yellow")
	assert.Error(t, field.Validate())
}

// TestValidators_Range tests Range validator.
func TestValidators_Range(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.Range(1, 100)(v)
		})

	field.SetValue("50")
	assert.NoError(t, field.Validate())

	field.SetValue("1")
	assert.NoError(t, field.Validate())

	field.SetValue("100")
	assert.NoError(t, field.Validate())

	field.SetValue("0")
	assert.Error(t, field.Validate())

	field.SetValue("101")
	assert.Error(t, field.Validate())
}

// TestValidators_Composite tests All and Any validators.
func TestValidators_Composite(t *testing.T) {
	t.Run("All requires all to pass", func(t *testing.T) {
		field := NewTextField("test").
			SetValidator(func(v string) error {
				return validators.All(
					validators.Required(),
					validators.MinLength(3),
					validators.MaxLength(10),
				)(v)
			})

		field.SetValue("hello")
		assert.NoError(t, field.Validate())

		field.SetValue("ab")
		assert.Error(t, field.Validate())

		field.SetValue("")
		assert.Error(t, field.Validate())

		field.SetValue("verylongstring")
		assert.Error(t, field.Validate())
	})

	t.Run("Any passes if one passes", func(t *testing.T) {
		field := NewTextField("test").
			SetValidator(func(v string) error {
				return validators.Any(
					validators.Pattern(`^\d+$`),
					validators.Pattern(`^[a-z]+$`),
				)(v)
			})

		field.SetValue("123")
		assert.NoError(t, field.Validate())

		field.SetValue("abc")
		assert.NoError(t, field.Validate())

		field.SetValue("ABC")
		assert.Error(t, field.Validate())
	})
}

// TestValidators_WithMessage tests custom error messages.
func TestValidators_WithMessage(t *testing.T) {
	field := NewTextField("test").
		SetValidator(func(v string) error {
			return validators.WithMessage(
				validators.MinLength(5),
				"Please enter at least 5 characters",
			)(v)
		})

	field.SetValue("ab")
	err := field.Validate()
	assert.Error(t, err)
	assert.Equal(t, "Please enter at least 5 characters", err.Error())
}
