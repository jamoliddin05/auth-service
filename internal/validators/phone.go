package validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// UzPhoneValidator checks if phone matches +998XXXXXXXXX format
func UzPhoneValidator(fl validator.FieldLevel) bool {
	regex := regexp.MustCompile(`^\+998\d{9}$`)
	return regex.MatchString(fl.Field().String())
}
