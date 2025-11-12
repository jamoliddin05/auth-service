package services

import (
	"app/internal/domain"
	"app/internal/stores"
	"app/internal/utils"
	"github.com/google/uuid"
	"time"
)

var (
	UserLoggedIn = "UserLoggedIn"
)

type TokenService struct {
	hasher         utils.PasswordHasher
	tokenGenerator utils.TokenGenerator
	jwt            utils.JWTHelper
}

func NewTokenService(hasher utils.PasswordHasher, tokenGenerator utils.TokenGenerator, jwt utils.JWTHelper) *TokenService {
	return &TokenService{
		hasher:         hasher,
		tokenGenerator: tokenGenerator,
		jwt:            jwt,
	}
}

func (s *TokenService) IssueTokenForUser(
	store *stores.GormUserTokenOutboxStore,
	user *domain.User,
) (string, string, error) {
	accessToken, refreshToken, err := s.generateTokens(user)
	if err != nil {
		return "", "", err
	}

	err = s.saveRefreshToken(store, user.ID, refreshToken)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *TokenService) VerifyRefreshToken(
	store *stores.GormUserTokenOutboxStore,
	userID uuid.UUID,
	refreshToken string,
) (bool, error) {
	token, err := store.Tokens().GetByUserID(userID)
	if err != nil {
		return false, err
	}

	if token == nil || !s.hasher.Verify(refreshToken, token.TokenHash) || token.ExpiresAt.Before(time.Now()) {
		return false, ErrInvalidCredentials
	}

	return true, nil
}

func (s *TokenService) generateTokens(user *domain.User) (string, string, error) {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Role
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID.String(), roles)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.tokenGenerator.GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *TokenService) saveRefreshToken(
	store *stores.GormUserTokenOutboxStore,
	userID uuid.UUID,
	refreshToken string,
) error {
	token, err := store.Tokens().GetByUserID(userID)
	if err != nil {
		return err
	}

	if token == nil {
		token = &domain.Token{
			UserID:    userID,
			TokenHash: s.hasher.Hash(refreshToken),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
	} else {
		token.TokenHash = s.hasher.Hash(refreshToken)
		token.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	return store.Tokens().Save(token)
}
