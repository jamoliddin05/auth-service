package helpers

import (
	"app/bootstrap/configs"
	"app/internal/handlers"
	"app/internal/services"
	"app/internal/uows"
	"app/internal/utils"
	"app/internal/validators"
	"log"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// MustInitDB инициализирует базу данных или падает
func MustInitDB(cfg *configs.Config) *configs.Wrapper {
	dbWrapper, err := configs.NewDBWrapper(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	return dbWrapper
}

// MustRegisterValidators регистрирует кастомные валидаторы Gin
func MustRegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("uzphone", validators.UzPhoneValidator); err != nil {
			log.Fatalf("could not register uzphone validator: %v", err)
		}

		if err := v.RegisterValidation("letters", validators.LettersValidator); err != nil {
			log.Fatalf("could not register letters validator: %v", err)
		}
	}
}

// BuildAuthHandler строит слой AuthService + Gin handler
func BuildAuthHandler(dbWrapper *configs.Wrapper, jwtPrivateKey string, jwtPublicKey string) *handlers.AuthHandler {
	uow := uows.NewUnitOfWork(dbWrapper.DB())
	hasher := utils.NewBcryptHasher()
	jwtHelper, err := utils.NewJWTManager(jwtPrivateKey, 15*time.Minute, "my_key_id")
	if err != nil {
		log.Fatalf("could not initialize JWT manager: %v", err)
	}
	jwksStr, err := configs.LoadJWKSFromPEM(jwtPublicKey, "my_key_id")
	if err != nil {
		log.Fatalf("could not initialize JWT manager: %v", err)
	}

	usersSvc := services.NewUserService(hasher)
	tokensSvc := services.NewTokenService(hasher, jwtHelper)
	authSvc := services.NewAuthService(uow, usersSvc, tokensSvc)
	authHandler := handlers.NewAuthHandler(authSvc, jwksStr)

	return authHandler
}
