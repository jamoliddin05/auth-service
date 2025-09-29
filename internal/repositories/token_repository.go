package repositories

import (
	"app/internal/domain"
	"gorm.io/gorm"
)

type TokenRepository interface {
	Save(token *domain.Token) error
	GetByID(id uint) (*domain.Token, error)
	GetByHash(hash string) (*domain.Token, error)
	Delete(id uint) error
	DeleteByUser(userID string) error
}

type TokenRepositoryImpl struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &TokenRepositoryImpl{db: db}
}

func (r *TokenRepositoryImpl) Save(token *domain.Token) error {
	return r.db.Create(token).Error
}

func (r *TokenRepositoryImpl) GetByID(id uint) (*domain.Token, error) {
	var token domain.Token
	if err := r.db.First(&token, id).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepositoryImpl) GetByHash(hash string) (*domain.Token, error) {
	var token domain.Token
	if err := r.db.Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&domain.Token{}, id).Error
}

func (r *TokenRepositoryImpl) DeleteByUser(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.Token{}).Error
}
