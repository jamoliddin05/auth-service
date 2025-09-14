package repositories

import (
	"app/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(u *domain.User) error
	GetByID(id uuid.UUID) (*domain.User, error)
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(u *domain.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepositoryImpl) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.Preload("Roles").First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
