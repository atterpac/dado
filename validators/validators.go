package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Validator is a function that validates a value and returns an error if invalid.
type Validator func(value any) error

// Required returns a validator that checks if a value is non-empty.
func Required() Validator {
	return func(value any) error {
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return errors.New("this field is required")
			}
		case nil:
			return errors.New("this field is required")
		case []string:
			if len(v) == 0 {
				return errors.New("at least one option must be selected")
			}
		case bool:
			// Booleans are always valid (false is a valid value)
			return nil
		}
		return nil
	}
}

// MinLength returns a validator that checks minimum string length.
func MinLength(n int) Validator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if len(s) < n {
				return fmt.Errorf("must be at least %d characters", n)
			}
		}
		return nil
	}
}

// MaxLength returns a validator that checks maximum string length.
func MaxLength(n int) Validator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if len(s) > n {
				return fmt.Errorf("must be at most %d characters", n)
			}
		}
		return nil
	}
}

// Email returns a validator that checks for valid email format.
func Email() Validator {
	// Simple email regex - covers most common cases
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return func(value any) error {
		if s, ok := value.(string); ok {
			if s == "" {
				return nil // Use Required() for non-empty check
			}
			if !emailRegex.MatchString(s) {
				return errors.New("invalid email format")
			}
		}
		return nil
	}
}

// Pattern returns a validator that checks if value matches a regex pattern.
func Pattern(pattern string) Validator {
	re := regexp.MustCompile(pattern)
	return func(value any) error {
		if s, ok := value.(string); ok {
			if s == "" {
				return nil // Use Required() for non-empty check
			}
			if !re.MatchString(s) {
				return errors.New("invalid format")
			}
		}
		return nil
	}
}

// PatternWithMessage returns a validator with a custom error message.
func PatternWithMessage(pattern, message string) Validator {
	re := regexp.MustCompile(pattern)
	return func(value any) error {
		if s, ok := value.(string); ok {
			if s == "" {
				return nil
			}
			if !re.MatchString(s) {
				return errors.New(message)
			}
		}
		return nil
	}
}

// Range returns a validator that checks if a numeric value is within range.
func Range(min, max float64) Validator {
	return func(value any) error {
		var n float64
		switch v := value.(type) {
		case int:
			n = float64(v)
		case int64:
			n = float64(v)
		case float64:
			n = v
		case string:
			// Try to parse as number
			var parsed float64
			if _, err := fmt.Sscanf(v, "%f", &parsed); err != nil {
				return errors.New("must be a number")
			}
			n = parsed
		default:
			return nil
		}

		if n < min || n > max {
			return fmt.Errorf("must be between %.0f and %.0f", min, max)
		}
		return nil
	}
}

// Min returns a validator that checks minimum numeric value.
func Min(min float64) Validator {
	return func(value any) error {
		var n float64
		switch v := value.(type) {
		case int:
			n = float64(v)
		case int64:
			n = float64(v)
		case float64:
			n = v
		case string:
			var parsed float64
			if _, err := fmt.Sscanf(v, "%f", &parsed); err != nil {
				return nil // Not a number, skip
			}
			n = parsed
		default:
			return nil
		}

		if n < min {
			return fmt.Errorf("must be at least %.0f", min)
		}
		return nil
	}
}

// Max returns a validator that checks maximum numeric value.
func Max(max float64) Validator {
	return func(value any) error {
		var n float64
		switch v := value.(type) {
		case int:
			n = float64(v)
		case int64:
			n = float64(v)
		case float64:
			n = v
		case string:
			var parsed float64
			if _, err := fmt.Sscanf(v, "%f", &parsed); err != nil {
				return nil
			}
			n = parsed
		default:
			return nil
		}

		if n > max {
			return fmt.Errorf("must be at most %.0f", max)
		}
		return nil
	}
}

// OneOf returns a validator that checks if value is one of allowed options.
func OneOf(options ...string) Validator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if s == "" {
				return nil
			}
			for _, opt := range options {
				if s == opt {
					return nil
				}
			}
			return fmt.Errorf("must be one of: %s", strings.Join(options, ", "))
		}
		return nil
	}
}

// Custom returns a validator from a custom function.
func Custom(fn func(value any) error) Validator {
	return fn
}

// All returns a validator that requires all provided validators to pass.
func All(validators ...Validator) Validator {
	return func(value any) error {
		for _, v := range validators {
			if err := v(value); err != nil {
				return err
			}
		}
		return nil
	}
}

// Any returns a validator that passes if any provided validator passes.
func Any(validators ...Validator) Validator {
	return func(value any) error {
		if len(validators) == 0 {
			return nil
		}

		var lastErr error
		for _, v := range validators {
			if err := v(value); err == nil {
				return nil
			} else {
				lastErr = err
			}
		}
		return lastErr
	}
}

// WithMessage wraps a validator with a custom error message.
func WithMessage(v Validator, message string) Validator {
	return func(value any) error {
		if err := v(value); err != nil {
			return errors.New(message)
		}
		return nil
	}
}

// URL returns a validator that checks for valid URL format.
func URL() Validator {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return func(value any) error {
		if s, ok := value.(string); ok {
			if s == "" {
				return nil
			}
			if !urlRegex.MatchString(s) {
				return errors.New("invalid URL format")
			}
		}
		return nil
	}
}

// Alphanumeric returns a validator that checks for alphanumeric characters only.
func Alphanumeric() Validator {
	re := regexp.MustCompile(`^[a-zA-Z0-9]*$`)
	return func(value any) error {
		if s, ok := value.(string); ok {
			if s == "" {
				return nil
			}
			if !re.MatchString(s) {
				return errors.New("must contain only letters and numbers")
			}
		}
		return nil
	}
}

// NoWhitespace returns a validator that rejects strings with whitespace.
func NoWhitespace() Validator {
	return func(value any) error {
		if s, ok := value.(string); ok {
			if strings.ContainsAny(s, " \t\n\r") {
				return errors.New("must not contain whitespace")
			}
		}
		return nil
	}
}
