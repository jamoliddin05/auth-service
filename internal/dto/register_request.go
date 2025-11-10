package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
	Surname  string `json:"surname" validate:"required"`
}

func (r *RegisterRequest) FieldErrorCode(field string) string {
	switch field {
	case "email":
		return "ERR_INVALID_EMAIL"
	case "password":
		return "ERR_PASSWORD_SHORT"
	case "name":
		return "ERR_INVALID_NAME"
	case "surname":
		return "ERR_INVALID_SURNAME"
	default:
		return "ERR"
	}
}
