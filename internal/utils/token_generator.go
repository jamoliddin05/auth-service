package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes), nil
}
