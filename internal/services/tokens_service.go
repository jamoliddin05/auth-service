package services

import (
	"app/internal/domain"
	"app/internal/stores"
	"app/internal/utils"
	"github.com/google/uuid"
	"time"
)

type TokenService struct {
	hasher utils.PasswordHasher
	jwt    utils.JWTHelper
}

func NewTokenService(hasher utils.PasswordHasher, jwt utils.JWTHelper) *TokenService {
	return &TokenService{hasher: hasher, jwt: jwt}
}

func (s *TokenService) IssueTokenForUser(store stores.UserTokenStore, user *domain.User) (string, string, error) {
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

func (s *TokenService) VerifyRefreshToken(store stores.UserTokenStore, userID uuid.UUID, refreshToken string) (bool, error) {
	token, err := store.Tokens().GetByUserID(userID)
	if err != nil {
		return false, err
	}

	if token == nil || !s.hasher.Verify(refreshToken, token.TokenHash) {
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

	refreshToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *TokenService) saveRefreshToken(store stores.UserTokenStore, userID uuid.UUID, refreshToken string) error {
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
