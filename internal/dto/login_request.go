package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (r *LoginRequest) FieldErrorCode(field string) string {
	switch field {
	case "email":
		return "ERR_INVALID_EMAIL"
	case "password":
		return "ERR_PASSWORD_SHORT"
	default:
		return "ERR"
	}
}
