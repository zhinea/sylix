package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a common validation error payload.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// messageFunc defines how to build an error message for a validation tag.
type messageFunc func(validator.FieldError) string

// BaseValidator provides reusable helpers for struct validation and error formatting.
type BaseValidator struct {
	validate    *validator.Validate
	tagMessages map[string]messageFunc
}

// NewBaseValidator returns a BaseValidator pre-configured with default messages.
func NewBaseValidator() *BaseValidator {
	return &BaseValidator{
		validate:    validator.New(),
		tagMessages: defaultTagMessages(),
	}
}

// ValidateStruct validates a struct and returns formatted errors when present.
func (b *BaseValidator) ValidateStruct(value interface{}) []ValidationError {
	if err := b.validate.Struct(value); err != nil {
		return b.formatErrors(err)
	}

	return nil
}

// RegisterValidation exposes validator.RegisterValidation to consumers.
func (b *BaseValidator) RegisterValidation(tag string, fn validator.Func) error {
	return b.validate.RegisterValidation(tag, fn)
}

// RegisterTagMessage allows components to override or add tag-specific error messages.
func (b *BaseValidator) RegisterTagMessage(tag string, fn messageFunc) {
	b.tagMessages[tag] = fn
}

// formatErrors converts validator.ValidationErrors into the common payload format.
func (b *BaseValidator) formatErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   fieldErr.Field(),
				Message: b.getErrorMessage(fieldErr),
			})
		}
	}

	return errors
}

func (b *BaseValidator) getErrorMessage(fieldErr validator.FieldError) string {
	if msgBuilder, ok := b.tagMessages[fieldErr.Tag()]; ok {
		return msgBuilder(fieldErr)
	}

	return fmt.Sprintf("%s is invalid", fieldErr.Field())
}

func defaultTagMessages() map[string]messageFunc {
	return map[string]messageFunc{
		"required": func(e validator.FieldError) string {
			return fmt.Sprintf("%s is required", e.Field())
		},
		"min": func(e validator.FieldError) string {
			return fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
		},
		"max": func(e validator.FieldError) string {
			return fmt.Sprintf("%s must be at most %s characters", e.Field(), e.Param())
		},
		"ip": func(e validator.FieldError) string {
			return fmt.Sprintf("%s must be a valid IP address", e.Field())
		},
		"oneof": func(e validator.FieldError) string {
			return fmt.Sprintf("%s must be one of: %s", e.Field(), e.Param())
		},
	}
}
