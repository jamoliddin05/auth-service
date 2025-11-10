package dto

import "app/internal/domain"

type UserResponse struct {
	User *domain.User `json:"user"`
}
