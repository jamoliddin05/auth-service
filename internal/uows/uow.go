package uows

import (
	"app/internal/repositories"
	"app/internal/stores"
)

type UnitOfWork interface {
    Store() stores.Store
    DoRegistration(fn func(userRepo repositories.UserRepository, eventRepo repositories.EventRepository) error) error
    // DoLogin(fn func(tokenRepo TokenRepository, eventRepo EventRepository) error) error
}

type gormUnitOfWork struct {
    db *gorm.DB
    store stores.Store
}

func NewUnitOfWork(db *gorm.DB) UnitOfWork {
    return &gormUnitOfWork{
        db: db,
        store: stores.NewStore(db),
    }
}

// plain store (без транзакций)
func (u *gormUnitOfWork) Store() Store {
    return u.store
}

// транзакция для регистрации
func (u *gormUnitOfWork) DoRegistration(fn func(UserRepository, EventRepository) error) error {
    return u.db.Transaction(func(tx *gorm.DB) error {
        txStore := stores.NewStore(tx)
        return fn(txStore.Users(), txStore.Outbox())
    })
}
