package services

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/mappers"
	"app/internal/repositories"
	"app/internal/uows"
	"app/internal/utils"
	"errors"
	"gorm.io/gorm"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type AuthService struct {
	uow    uows.UnitOfWork
	hasher utils.PasswordHasher
}

func NewAuthService(uow uows.UnitOfWork, hasher utils.PasswordHasher) *AuthService {
	return &AuthService{uow: uow, hasher: hasher}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// Check if user already exists
	existingUser, err := s.uow.Store().Users().GetByPhone(req.Phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Map DTO to domain user
	user := mappers.DTOToUser(req)

	// Hash the password
	user.Password = s.hasher.Hash(user.Password)

	// Assign default role
	user.Roles = []domain.UserRole{
		{Role: domain.RoleCustomer},
	}

	// Do registration in a transaction
	err = s.uow.DoRegistration(func(userRepo repositories.UserRepository, eventRepo repositories.EventRepository) error {
		if err := userRepo.Create(user); err != nil {
			return err
		}
		return eventRepo.Save("UserRegistered", user)
	})
	if err != nil {
		return nil, err
	}

	// Map domain user back to DTO and return
	return mappers.UserToDTO(user), nil
}
