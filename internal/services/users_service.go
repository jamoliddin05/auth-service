package services

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/stores"
	"app/internal/utils"
	"github.com/google/uuid"
	"strings"
)

type UserService struct {
	hasher utils.PasswordHasher
}

func NewUserService(hasher utils.PasswordHasher) *UserService {
	return &UserService{hasher: hasher}
}

func (s *UserService) Register(store stores.UserTokenStore, req dto.RegisterRequest) (*domain.User, error) {
	user, err := store.Users().GetByPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, ErrUserAlreadyExists
	}

	user = &domain.User{
		Phone:    req.Phone,
		Password: s.hasher.Hash(req.Password),
		Name:     req.Name,
		Surname:  req.Surname,
		Roles: []domain.UserRole{
			{Role: domain.RoleCustomer},
		},
	}

	if err := store.Users().Save(user); err != nil {
		return nil, err
	}

	err = store.Outbox().Save("UserRegistered", user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) PromoteToSeller(store stores.UserTokenStore, userId string) (*domain.User, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := store.Users().GetByID(userUUID)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	hasSeller := false
	for _, r := range user.Roles {
		if strings.EqualFold(r.Role, domain.RoleSeller) {
			hasSeller = true
			break
		}
	}

	if !hasSeller {
		user.Roles = append(user.Roles, domain.UserRole{UserID: userUUID, Role: domain.RoleSeller})
		if err := store.Users().Save(user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) GetByID(store stores.UserTokenStore, userId string) (*domain.User, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, err
	}

	user, err := store.Users().GetByID(userUUID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) Authenticate(store stores.UserTokenStore, phone string, password string) (*domain.User, error) {
	user, err := store.Users().GetByPhone(phone)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	if !s.hasher.Verify(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
