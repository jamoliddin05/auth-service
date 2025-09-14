package stores

import (
	"app/internal/repositories"
)

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

func (s *gormStore) Users() repositories.UserRepository   { return repositories.NewUserRepo(s.db) }
func (s *gormStore) Outbox() repositories.EventRepository { return repositories.NewEventRepo(s.db) }