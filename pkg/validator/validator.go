package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wrap thư viện go-playground/validator
// Tương đương việc bạn dùng zod.object({...}).parse() bên Node.js
type Validator struct {
	v *validator.Validate
}

func New() *Validator {
	return &Validator{v: validator.New()}
}

// Struct validate một struct bất kỳ dựa vào tag `validate:"..."`
// Trả về error đã được format dễ đọc, ví dụ:
//   "title is required; description must be at most 1000 characters"
func (vd *Validator) Struct(s any) error {
	err := vd.v.Struct(s)
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
