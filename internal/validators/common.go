package validators

import (
	"app/internal/dto"
	"errors"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationResult struct {
	Valid  bool
	Errors map[string]string
}

type Validator interface {
	Validate(req dto.Request) ValidationResult
}

type GoValidator struct {
	validator *validator.Validate
}

func NewValidator(validator *validator.Validate) Validator {
	return &GoValidator{validator: validator}
}

func (v *GoValidator) Validate(req dto.Request) ValidationResult {
	log.Printf("Test: %+v", req)
	err := v.validator.Struct(req)
	if err == nil {
		return ValidationResult{Valid: true}
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errorsMap := make(map[string]string)
		for _, fe := range ve {
			field := strings.ToLower(fe.Field())
			errorsMap[field] = req.FieldErrorCode(field)
		}
		return ValidationResult{
			Valid:  false,
			Errors: errorsMap,
		}
	}

	return ValidationResult{
		Valid:  false,
		Errors: map[string]string{"_error": "validation_failed"},
	}
}
