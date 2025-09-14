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
	// 1️⃣ Check if user already exists
	existingUser, err := s.uow.Store().Users().GetByPhone(req.Phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// 2️⃣ Map DTO to domain user
	user := mappers.DTOToUser(req)

	// 3️⃣ Hash the password
	user.Password = s.hasher.Hash(user.Password)

	// 4️⃣ Assign default role
	user.Roles = []domain.UserRole{
		{Role: domain.RoleCustomer},
	}

	// 5️⃣ Do registration in a transaction
	err = s.uow.DoRegistration(func(userRepo repositories.UserRepository, eventRepo repositories.EventRepository) error {
		if err := userRepo.Create(user); err != nil {
			return err
		}
		return eventRepo.Save("UserRegistered", user)
	})
	if err != nil {
		return nil, err
	}

	// 6️⃣ Map domain user back to DTO and return
	return mappers.UserToDTO(user), nil
}
