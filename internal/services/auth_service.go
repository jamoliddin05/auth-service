package services

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/repositories"
	"app/internal/uows"
	"app/internal/utils"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
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

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
	existingUser, err := s.uow.Store().Users().GetByPhone(req.Phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	user := &domain.User{
		Phone:    req.Phone,
		Password: req.Password,
		Name:     req.Name,
		Surname:  req.Surname,
	}

	user.Password = s.hasher.Hash(user.Password)

	user.Roles = []domain.UserRole{
		{Role: domain.RoleCustomer},
	}

	err = s.uow.DoRegistration(func(userRepo repositories.UserRepository, eventRepo repositories.EventRepository) error {
		if err := userRepo.Save(user); err != nil {
			return err
		}
		return eventRepo.Save("UserRegistered", user)
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		User: *user,
	}, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.TokenResponse, error) {
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

	refreshResp, err := s.generateRefreshResponse(existingUser)
	if err != nil {
		return nil, err
	}

	err = s.uow.DoLogin(func(tokenRepo repositories.TokenRepository, eventRepo repositories.EventRepository) error {
		token := &domain.Token{
			UserID:    existingUser.ID,
			User:      existingUser,
			TokenHash: s.hasher.Hash(refreshResp.RefreshToken),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		if err := tokenRepo.Save(token); err != nil {
			return err
		}
		return eventRepo.Save("UserLoggedIn", existingUser)
	})
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  refreshResp.AccessToken,
		RefreshToken: refreshResp.RefreshToken,
	}, nil
}

func (s *AuthService) BecomeSeller(userId string) (*dto.TokenResponse, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	existingUser, err := s.uow.Store().Users().GetByID(userUUID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidCredentials
	}

	hasSeller := false
	for _, r := range existingUser.Roles {
		if strings.EqualFold(r.Role, domain.RoleSeller) {
			hasSeller = true
			break
		}
	}

	if !hasSeller {
		existingUser.Roles = append(existingUser.Roles, domain.UserRole{
			UserID: userUUID,
			Role:   domain.RoleSeller,
		})

		err = s.uow.Store().Users().Save(existingUser)
		if err != nil {
			return nil, err
		}
	}

	resp, err := s.generateRefreshResponse(existingUser)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *AuthService) Refresh(req dto.RefreshRequest, userId string) (*dto.TokenResponse, error) {
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

	valid, err := s.validateRefreshToken(existingUser.ID, req.RefreshToken)
	if err != nil || !valid {
		return nil, ErrInvalidCredentials
	}

	return s.generateRefreshResponse(existingUser)
}

func (s *AuthService) GetMe(userId string) (*dto.UserResponse, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, ErrInvalidCredentials
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

func (s *AuthService) validateRefreshToken(userID uuid.UUID, refreshToken string) (bool, error) {
	tokens, err := s.uow.Store().Tokens().GetByUserID(userID.String())
	if err != nil {
		return false, err
	}

	for _, t := range tokens {
		if s.hasher.Verify(refreshToken, t.TokenHash) && t.ExpiresAt.After(time.Now()) {
			return true, nil
		}
	}

	return false, nil
}

func (s *AuthService) generateRefreshResponse(user *domain.User) (*dto.TokenResponse, error) {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Role
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), roles)
	if err != nil {
		return nil, err
	}

	tokenString, err := utils.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}
	newTokenHash := s.hasher.Hash(tokenString)

	token := &domain.Token{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	if err := s.uow.Store().Tokens().Save(token); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: tokenString,
	}, nil
}
