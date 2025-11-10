package services

import (
	"app/internal/stores"
)

type OutboxService struct {
}

var (
	UserRegistered = "UserRegistered"
	UserLoggedIn   = "UserLoggedIn"
)

func NewOutboxService() *OutboxService {
	return &OutboxService{}
}

func (s *OutboxService) SaveUserRegisteredEvent(store stores.UserTokenOutboxStore, payload interface{}) error {
	return store.Outbox().Save(UserRegistered, payload)
}

func (s *OutboxService) SaveUserLoggedInEvent(store stores.UserTokenOutboxStore, payload interface{}) error {
	return store.Outbox().Save(UserLoggedIn, payload)
}
