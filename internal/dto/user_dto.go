package dto

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,uzphone"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterResponse struct {
	ID    string   `json:"id"`
	Phone string   `json:"phone"`
	Roles []string `json:"roles"`
}
