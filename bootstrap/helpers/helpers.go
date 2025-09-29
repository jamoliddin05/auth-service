package helpers

import (
	"app/bootstrap/db"
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
func MustInitDB(cfg *db.Config) *db.Wrapper {
	dbWrapper, err := db.NewDBWrapper(cfg.DSN())
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
	}
}

// BuildAuthHandler строит слой AuthService + Gin handler
func BuildAuthHandler(dbWrapper *db.Wrapper, jwtKey string) *handlers.GinAuthHandler {
	uow := uows.NewUnitOfWork(dbWrapper.DB())
	hasher := utils.NewBcryptHasher()
	jwtHelper, err := utils.NewJWTManager(jwtKey, 15*time.Minute)
	if err != nil {
		log.Fatalf("could not initialize JWT manager: %v", err)
	}
	authSvc := services.NewAuthService(uow, hasher, jwtHelper)
	authHandler := handlers.NewGinAuthHandler(authSvc)
	return authHandler
}
