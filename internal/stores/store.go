package stores

import (
	"app/internal/repositories"
	"gorm.io/gorm"
)

//go:generate mockery --name=Store --output=../mocks --structname=StoreMock
type UserTokenStore interface {
	Users() repositories.UserRepository
	Tokens() repositories.TokenRepository
	Outbox() repositories.EventRepository
}

type UserTokenStoreImpl struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) UserTokenStore {
	return &UserTokenStoreImpl{db: db}
}

func (s *UserTokenStoreImpl) Users() repositories.UserRepository {
	return repositories.NewUserRepository(s.db)
}
func (s *UserTokenStoreImpl) Tokens() repositories.TokenRepository {
	return repositories.NewTokenRepository(s.db)
}
func (s *UserTokenStoreImpl) Outbox() repositories.EventRepository {
	return repositories.NewEventRepository(s.db)
}
