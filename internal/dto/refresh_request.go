package dto

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (r *RefreshRequest) FieldErrorCode(field string) string {
	switch field {
	case "refresh_token":
		return "ERR_INVALID_REFRESH_TOKEN"
	default:
		return "ERR"
	}
}
