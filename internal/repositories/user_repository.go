package repositories

import (
	"app/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(u *domain.User) error
	GetByID(id uuid.UUID) (*domain.User, error)
}
