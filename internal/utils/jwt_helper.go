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
	kid           string
}

func NewJWTManager(pemKey string, duration time.Duration, kid string) (*JWTManager, error) {
	if pemKey == "" {
		return nil, fmt.Errorf("private key PEM string is empty")
	}

	pemKey = strings.ReplaceAll(pemKey, `\n`, "\n")

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pemKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &JWTManager{
		privateKey:    privateKey,
		tokenDuration: duration,
		kid:           kid,
	}, nil
}

type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

func (j *JWTManager) GenerateAccessToken(userID string, roles []string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = j.kid

	return token.SignedString(j.privateKey)
}
