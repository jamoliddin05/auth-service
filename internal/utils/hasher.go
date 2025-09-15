package utils

import (
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockery --name=PasswordHasher --output=../mocks --structname=PasswordHasherMock
type PasswordHasher interface {
	Hash(password string) string
	Verify(password, hash string) bool
}

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (h *BcryptHasher) Hash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func (h *BcryptHasher) Verify(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
