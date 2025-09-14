package services

import (
	"app/internal/dto"
	"app/internal/mappers"
	"app/internal/repositories"
	"app/internal/uows"
	"app/internal/utils"
)

type AuthService struct {
	uow    uows.UnitOfWork
	hasher utils.PasswordHasher
}

func NewAuthService(uow uows.UnitOfWork, hasher utils.PasswordHasher) *AuthService {
	return &AuthService{uow: uow, hasher: hasher}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	user := mappers.DTOToUser(req)

	err := s.uow.DoRegistration(func(userRepo repositories.UserRepository, eventRepo repositories.EventRepository) error {
		if err := userRepo.Create(user); err != nil {
			return err
		}
		return eventRepo.Save("UserRegistered", user)
	})
	if err != nil {
		return nil, err
	}

	return mappers.UserToDTO(user), nil
}
