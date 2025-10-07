package mappers

import (
	"app/internal/domain"
	"app/internal/dto"
)

// DTOToUser DTO → Domain
func DTOToUser(req dto.RegisterRequest) *domain.User {
	return &domain.User{
		Phone:    req.Phone,
		Password: req.Password,
		Name:     req.Name,
		Surname:  req.Surname,
	}
}

// UserToDTO Domain → DTO
func UserToDTO(user *domain.User) *dto.RegisterResponse {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Role
	}

	return &dto.RegisterResponse{
		ID:    user.ID.String(),
		Phone: user.Phone,
		Roles: roles,
	}
}
