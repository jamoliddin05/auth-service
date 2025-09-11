package dto

type UserCreateRequest struct {
	Phone    string `json:"phone" binding:"required,uzphone"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserResponse struct {
	ID    string   `json:"id"`
	Phone string   `json:"phone"`
	Roles []string `json:"roles"`
}
