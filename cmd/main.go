package main

import (
	"app/bootstrap"
	"app/bootstrap/db"
	"app/internal/handlers"
	"app/internal/services"
	"app/internal/uows"
	"app/internal/utils"
	"app/internal/validators"
	"context"
	"errors"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configs
	cfg := db.LoadConfig()

	// ---- 1. Initialize DB ----
	dbWrapper, err := db.NewDBWrapper(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("uzphone", validators.UzPhoneValidator)
		if err != nil {
			log.Fatalf("could not register uzphone validator: %v", err)
		}
	}

	// ---- 2. Build layers (repo → service → handler) ----
	uow := uows.NewUnitOfWork(dbWrapper.DB())
	hasher := utils.NewBcryptHasher()
	authSvc := services.NewAuthService(uow, hasher)
	userHandler := handlers.NewGinAuthHandler(authSvc)

	// ---- 3. Setup Gin router ----
	r := gin.Default()
	userHandler.BindRoutes(r)

	// ---- 4. Build App ----
	app := bootstrap.NewApp(r, ":8080")
	app.RegisterCloser(dbWrapper) // register DB for shutdown

	// ---- 5. Run server in background ----
	go func() {
		if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()
	log.Println("server running on :8080")

	// ---- 6. Listen for termination signals ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received")

	// ---- 7. Graceful shutdown ----
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown gracefully: %v", err)
	}
	log.Println("server exited cleanly")
}
