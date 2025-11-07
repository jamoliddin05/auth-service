package uows

import (
	"app/internal/stores"
	"gorm.io/gorm"
)

//go:generate mockery --name=UnitOfWork --output=../mocks --structname=UnitOfWorkMock
type UserTokenUnitOfWork interface {
	DoTransaction(fn func(txStores stores.UserTokenStore) error) error
}

type UserTokenUnitOfWorkImpl struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) UserTokenUnitOfWork {
	return &UserTokenUnitOfWorkImpl{
		db: db,
	}
}

// DoTransaction runs a function inside a transaction with a transactional store
func (u *UserTokenUnitOfWorkImpl) DoTransaction(fn func(txStores stores.UserTokenStore) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txStore := stores.NewStore(tx)
		return fn(txStore)
	})
}
