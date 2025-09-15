package stores

import (
	"app/internal/repositories"
	"gorm.io/gorm"
)

//go:generate mockery --name=Store --output=../mocks --structname=StoreMock
type Store interface {
	Users() repositories.UserRepository
	Outbox() repositories.EventRepository
}

type gormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) Store {
	return &gormStore{db: db}
}

func (s *gormStore) Users() repositories.UserRepository { return repositories.NewUserRepository(s.db) }
func (s *gormStore) Outbox() repositories.EventRepository {
	return repositories.NewEventRepository(s.db)
}
