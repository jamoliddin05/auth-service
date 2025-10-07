package dto

import "app/internal/domain"

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,uzphone"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Surname  string `json:"surname" binding:"required"`
}

type RegisterResponse struct {
	ID    string   `json:"id"`
	Phone string   `json:"phone"`
	Roles []string `json:"roles"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required,uzphone"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	User domain.User `json:"user"`
}
