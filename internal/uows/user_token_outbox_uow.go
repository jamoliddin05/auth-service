package uows

import (
	"app/internal/stores"
	"gorm.io/gorm"
)

//go:generate mockery --name=UnitOfWork --output=../mocks --structname=UnitOfWorkMock
type UserTokeOutboxUnitOfWork interface {
	DoTransaction(fn func(txStores stores.UserTokenOutboxStore) error) error
	Do(fn func(txStores stores.UserTokenOutboxStore) error) error
}

type GormUserTokeOutboxUnitOfWork struct {
	db *gorm.DB
}

func NewGormUserTokeOutboxUnitOfWork(db *gorm.DB) UserTokeOutboxUnitOfWork {
	return &GormUserTokeOutboxUnitOfWork{
		db: db,
	}
}

func (u *GormUserTokeOutboxUnitOfWork) DoTransaction(fn func(txStores stores.UserTokenOutboxStore) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txStore := stores.NewUserTokenOutboxStore(tx)
		return fn(txStore)
	})
}

func (u *GormUserTokeOutboxUnitOfWork) Do(fn func(store stores.UserTokenOutboxStore) error) error {
	store := stores.NewUserTokenOutboxStore(u.db)
	return fn(store)
}
