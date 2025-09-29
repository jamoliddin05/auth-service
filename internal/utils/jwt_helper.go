package utils

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

//go:generate mockery --name=JWTHelper --output=../mocks --structname=JWTHelperMock
type JWTHelper interface {
	GenerateAccessToken(userID string, roles []string) (string, error)
}

type JWTManager struct {
	privateKey    *rsa.PrivateKey
	tokenDuration time.Duration
}

// NewJWTManager parses an RSA private key from a PEM string
func NewJWTManager(pemKey string, duration time.Duration) (*JWTManager, error) {
	if pemKey == "" {
		return nil, fmt.Errorf("private key PEM string is empty")
	}

	// Replace literal \n with actual newlines
	pemKey = strings.ReplaceAll(pemKey, `\n`, "\n")

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pemKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &JWTManager{
		privateKey:    privateKey,
		tokenDuration: duration,
	}, nil
}

// Claims structure
type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateAccessToken signs a JWT with RS256
func (j *JWTManager) GenerateAccessToken(userID string, roles []string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}
