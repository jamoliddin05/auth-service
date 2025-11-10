package repositories

import (
	"app/internal/domain"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockery --name=UserRepository --output=../mocks --structname=UserRepositoryMock
type UserRepository interface {
	Save(u *domain.User) error
	GetByID(id uuid.UUID) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Save(u *domain.User) error {
	return r.db.Save(u).Error
}

func (r *UserRepositoryImpl) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("Roles").First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepositoryImpl) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("Roles").First(&user, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
