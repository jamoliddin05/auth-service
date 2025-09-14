package services

import (
    "app/internal/repositories"
    "app/internal/uows"
    "app/internal/utils"
    "app/internal/domain"
)

type AuthService struct {
    uow uows.UnitOfWork
    hasher utils.PasswordHasher
}

func NewAuthService(uow uows.UnitOfWork, hasher PasswordHasher) *AuthService {
    return &AuthService{uow: uow, hasher: hasher}
}

func (s *AuthService) Register(dto RegisterDTO) (*UserResponse, error) {
	user := &domain.User{
		Phone:    dto.Phone,
		Password: s.hasher.Hash(dto.Password),
		Roles: []domain.UserRole{
			{Role: domain.RoleCustomer},
		},
	}

	err := s.uow.DoRegistration(func(userRepo repositories.UserRepository, eventRepo repositories.EventRepository) error {
		if err := userRepo.Create(user); err != nil {
			return err
		}
		return eventRepo.Save("UserRegistered", user)
	})
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:    user.ID,
		Phone: user.Phone,
		Roles: user.Roles,
	}, nil
}