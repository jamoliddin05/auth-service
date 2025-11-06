package dto

import "app/internal/domain"

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,uzphone"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,letters"`
	Surname  string `json:"surname" binding:"required,letters"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required,uzphone"`
	Password string `json:"password" binding:"required,min=6"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	User domain.User `json:"user"`
}
