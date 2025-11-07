package repositories

import (
	"app/internal/domain"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//go:generate mockery --name=TokenRepository --output=../mocks --structname=TokenRepositoryMock
type TokenRepository interface {
	Save(token *domain.Token) error
	GetByID(id uint) (*domain.Token, error)
	GetByHash(hash string) (*domain.Token, error)
	Delete(id uint) error
	DeleteByUser(userID uuid.UUID) error
	GetByUserID(userID uuid.UUID) (*domain.Token, error)
}

type TokenRepositoryImpl struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &TokenRepositoryImpl{db: db}
}

func (r *TokenRepositoryImpl) Save(token *domain.Token) error {
	return r.db.Save(token).Error
}

func (r *TokenRepositoryImpl) GetByID(id uint) (*domain.Token, error) {
	var token domain.Token
	err := r.db.First(&token, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepositoryImpl) GetByHash(hash string) (*domain.Token, error) {
	var token domain.Token
	err := r.db.Where("token_hash = ?", hash).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&domain.Token{}, id).Error
}

func (r *TokenRepositoryImpl) DeleteByUser(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.Token{}).Error
}

func (r *TokenRepositoryImpl) GetByUserID(userID uuid.UUID) (*domain.Token, error) {
	var token domain.Token
	err := r.db.Where("user_id = ?", userID).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}
