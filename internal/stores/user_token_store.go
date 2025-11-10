package stores

import (
	"app/internal/repositories"
	"gorm.io/gorm"
)

//go:generate mockery --name=Store --output=../mocks --structname=StoreMock
type UserTokenOutboxStore interface {
	Users() repositories.UserRepository
	Tokens() repositories.TokenRepository
	Outbox() repositories.EventRepository
}

type GormUserTokenOutboxStore struct {
	db *gorm.DB
}

func NewUserTokenOutboxStore(db *gorm.DB) UserTokenOutboxStore {
	return &GormUserTokenOutboxStore{db: db}
}

func (s *GormUserTokenOutboxStore) Users() repositories.UserRepository {
	return repositories.NewUserRepository(s.db)
}
func (s *GormUserTokenOutboxStore) Tokens() repositories.TokenRepository {
	return repositories.NewTokenRepository(s.db)
}
func (s *GormUserTokenOutboxStore) Outbox() repositories.EventRepository {
	return repositories.NewEventRepository(s.db)
}
