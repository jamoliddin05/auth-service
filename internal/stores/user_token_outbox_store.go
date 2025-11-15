package stores

import (
	"app/internal/repositories"
	"gorm.io/gorm"
)

type UserTokenOutboxStore struct {
	db *gorm.DB
}

func NewUserTokenOutboxStore(db *gorm.DB) *UserTokenOutboxStore {
	return &UserTokenOutboxStore{db: db}
}

func (s *UserTokenOutboxStore) Users() repositories.UserRepository {
	return repositories.NewUserRepository(s.db)
}
func (s *UserTokenOutboxStore) Tokens() repositories.TokenRepository {
	return repositories.NewTokenRepository(s.db)
}
func (s *UserTokenOutboxStore) Outbox() repositories.EventRepository {
	return repositories.NewEventRepository(s.db)
}
