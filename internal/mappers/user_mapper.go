package mappers

import (
	"app/internal/domain"
	"app/internal/dto"
)

// DTOToUser DTO → Domain
func DTOToUser(req dto.UserCreateRequest) *domain.User {
	return &domain.User{
		Phone:    req.Phone,
		Password: req.Password,
	}
}

// UserToDTO Domain → DTO
func UserToDTO(user *domain.User) *dto.UserResponse {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Role
	}

	return &dto.UserResponse{
		ID:    user.ID.String(),
		Phone: user.Phone,
		Roles: roles,
	}
}
