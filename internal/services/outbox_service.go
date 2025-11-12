package services

import (
	"app/internal/stores"
)

type UserTokenOutboxService struct {
}

func NewOutboxService() *UserTokenOutboxService {
	return &UserTokenOutboxService{}
}

func (s *UserTokenOutboxService) SaveUserRegisteredEvent(store *stores.GormUserTokenOutboxStore, payload interface{}) error {
	return store.Outbox().Save(UserRegistered, payload)
}

func (s *UserTokenOutboxService) SaveUserLoggedInEvent(store *stores.GormUserTokenOutboxStore, payload interface{}) error {
	return store.Outbox().Save(UserLoggedIn, payload)
}
