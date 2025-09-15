package uows

import (
	"app/internal/repositories"
	"app/internal/stores"
	"gorm.io/gorm"
)

//go:generate mockery --name=UnitOfWork --output=../mocks --structname=UnitOfWorkMock
type UnitOfWork interface {
	Store() stores.Store
	DoRegistration(fn func(userRepo repositories.UserRepository, eventRepo repositories.EventRepository) error) error
	// DoLogin(fn func(tokenRepo TokenRepository, eventRepo EventRepository) error) error
}

type gormUnitOfWork struct {
	db    *gorm.DB
	store stores.Store
}

func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &gormUnitOfWork{
		db:    db,
		store: stores.NewStore(db),
	}
}

// Store plain store
func (u *gormUnitOfWork) Store() stores.Store {
	return u.store
}

// DoRegistration with tx
func (u *gormUnitOfWork) DoRegistration(fn func(repositories.UserRepository, repositories.EventRepository) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txStore := stores.NewStore(tx)
		return fn(txStore.Users(), txStore.Outbox())
	})
}
