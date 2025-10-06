package services

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/mappers"
	"app/internal/repositories"
	"app/internal/uows"
	"app/internal/utils"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	uow    uows.UnitOfWork
	hasher utils.PasswordHasher
	jwt    utils.JWTHelper
}

func NewAuthService(
	uow uows.UnitOfWork,
	hasher utils.PasswordHasher,
	jwtHelper utils.JWTHelper,
) *AuthService {
	return &AuthService{
		uow:    uow,
		hasher: hasher,
		jwt:    jwtHelper,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	existingUser, err := s.uow.Store().Users().GetByPhone(req.Phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	user := mappers.DTOToUser(req)

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

	return mappers.UserToDTO(user), nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	existingUser, err := s.uow.Store().Users().GetByPhone(req.Phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !s.hasher.Verify(req.Password, existingUser.Password) {
		return nil, ErrInvalidCredentials
	}

	roles := make([]string, len(existingUser.Roles))
	for i, r := range existingUser.Roles {
		roles[i] = r.Role
	}

	// Generating JWT access token
	accessToken, err := s.jwt.GenerateAccessToken(existingUser.ID.String(), roles)
	if err != nil {
		return nil, err
	}

	// Generating refresh token
	tokenString, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}
	tokenHash := s.hasher.Hash(tokenString)

	token := &domain.Token{
		UserID:    existingUser.ID,
		User:      existingUser,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// Saving the token and the Login event in one transaction
	err = s.uow.DoLogin(func(tokenRepo repositories.TokenRepository, eventRepo repositories.EventRepository) error {
		if err := tokenRepo.Save(token); err != nil {
			return err
		}
		return eventRepo.Save("UserLoggedIn", existingUser)
	})

	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: tokenString,
	}, err
}

func (s *AuthService) Refresh(req dto.RefreshRequest, userId string) (*dto.RefreshResponse, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	existingUser, err := s.uow.Store().Users().GetByID(userUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	tokens, err := s.uow.Store().Tokens().GetByUserID(userUUID.String())
	if err != nil {
		return nil, err
	}

	var matchedToken *domain.Token
	for _, t := range tokens {
		if s.hasher.Verify(req.RefreshToken, t.TokenHash) {
			matchedToken = t
			break
		}
	}

	if matchedToken == nil {
		return nil, ErrInvalidCredentials
	}

	roles := make([]string, len(existingUser.Roles))
	for i, r := range existingUser.Roles {
		roles[i] = r.Role
	}

	accessToken, err := s.jwt.GenerateAccessToken(existingUser.ID.String(), roles)
	if err != nil {
		return nil, err
	}

	tokenString, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}
	newTokenHash := s.hasher.Hash(tokenString)
	matchedToken.TokenHash = newTokenHash
	err = s.uow.Store().Tokens().Save(matchedToken)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: tokenString,
	}, nil
}

func (s *AuthService) GetMe(userId string) (*dto.UserResponse, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	existingUser, err := s.uow.Store().Users().GetByID(userUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	return &dto.UserResponse{
		User: *existingUser,
	}, nil
}
