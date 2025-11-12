package helpers

import (
	"app/bootstrap/configs"
	"app/internal/handlers"
	"app/internal/middlewares"
	"app/internal/services"
	"app/internal/stores"
	"app/internal/uows"
	"app/internal/utils"
	"app/internal/validators"
	"github.com/go-playground/validator/v10"
	"log"
	"time"
)

func MustInitDB(cfg *configs.Config) *configs.Wrapper {
	dbWrapper, err := configs.NewDBWrapper(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	return dbWrapper
}

func BuildAuthHandler(dbWrapper *configs.Wrapper, jwtPrivateKey string) *handlers.AuthHandler {
	jwtHelper, err := utils.NewJWTManager(jwtPrivateKey, 15*time.Minute, "my_key_id")
	if err != nil {
		log.Fatalf("could not initialize JWT manager: %v", err)
	}

	uow := uows.NewGormUnitOfWork[*stores.GormUserTokenOutboxStore](dbWrapper.DB(), stores.NewUserTokenOutboxStore)
	hasher := utils.NewBcryptHasher()
	tokenGenerator := utils.NewTokenGenerator()
	val := validators.NewValidator(validator.New())
	middleware := middlewares.NewRequestValidator(val)

	usersSvc := services.NewUserService(hasher)
	tokensSvc := services.NewTokenService(hasher, tokenGenerator, jwtHelper)
	outboxSvc := services.NewOutboxService()
	authHandler := handlers.NewAuthHandler(uow, middleware, usersSvc, tokensSvc, outboxSvc)

	return authHandler
}

func BuildUserHandler(dbWrapper *configs.Wrapper) *handlers.UserHandler {
	uow := uows.NewGormUnitOfWork[*stores.GormUserTokenOutboxStore](dbWrapper.DB(), stores.NewUserTokenOutboxStore)
	hasher := utils.NewBcryptHasher()
	val := validators.NewValidator(validator.New())
	middleware := middlewares.NewRequestValidator(val)

	usersSvc := services.NewUserService(hasher)
	outboxSvc := services.NewOutboxService()

	return handlers.NewUserHandler(uow, middleware, usersSvc, outboxSvc)
}

func BuildJwksHandler(jwtPublicKey string) *handlers.JwksHandler {
	jwksStr, err := configs.LoadJWKSFromPEM(jwtPublicKey, "my_key_id")
	if err != nil {
		log.Fatalf("could not initialize JWT manager: %v", err)
	}

	return handlers.NewJwksHandler(jwksStr)
}
