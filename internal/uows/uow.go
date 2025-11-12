package uows

import (
	"gorm.io/gorm"
)

type UnitOfWork[T any] interface {
	DoTransaction(fn func(store T) error) error
	Do(fn func(store T) error) error
}

type GormUnitOfWork[T any] struct {
	db           *gorm.DB
	storeFactory func(tx *gorm.DB) T
}

func NewGormUnitOfWork[T any](db *gorm.DB, factory func(tx *gorm.DB) T) *GormUnitOfWork[T] {
	return &GormUnitOfWork[T]{db: db, storeFactory: factory}
}

func (u *GormUnitOfWork[T]) DoTransaction(fn func(store T) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txStore := u.storeFactory(tx)
		return fn(txStore)
	})
}

func (u *GormUnitOfWork[T]) Do(fn func(store T) error) error {
	store := u.storeFactory(u.db)
	return fn(store)
}
