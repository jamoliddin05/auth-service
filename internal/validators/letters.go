package validators

import (
	"github.com/go-playground/validator/v10"
	"unicode"
)

// LettersValidator checks if the field contains only letters
func LettersValidator(fl validator.FieldLevel) bool {
	for _, r := range fl.Field().String() {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
