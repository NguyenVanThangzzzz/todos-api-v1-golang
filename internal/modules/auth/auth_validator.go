package auth

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func validateStruct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}
	msgs := make([]string, 0, len(validationErrs))
	for _, fe := range validationErrs {
		msgs = append(msgs, formatFieldError(fe))
	}
	return fmt.Errorf("%s", strings.Join(msgs, "; "))
}

func formatFieldError(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	default:
		return fmt.Sprintf("%s is invalid (%s)", field, fe.Tag())
	}
}
