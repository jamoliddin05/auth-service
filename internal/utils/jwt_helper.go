package utils

import (
	"crypto/rsa"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

//go:generate mockery --name=JWTHelper --output=../mocks --structname=JWTHelperMock
type JWTHelper interface {
	GenerateAccessToken(userID string, roles []string) (string, error)
}

type JWTManager struct {
	privateKey    *rsa.PrivateKey
	tokenDuration time.Duration
}

// NewJWTManager loads an RSA private key from file
func NewJWTManager(privateKeyPath string, duration time.Duration) *JWTManager {
	privBytes, _ := os.ReadFile(privateKeyPath)
	privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM(privBytes)

	return &JWTManager{
		privateKey:    privateKey,
		tokenDuration: duration,
	}
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
