package main

import (
	"app/bootstrap"
	"app/bootstrap/db"
	"app/bootstrap/helpers"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := db.LoadConfig()

	dbWrapper := helpers.MustInitDB(cfg)
	helpers.MustRegisterValidators()

	authHandler := helpers.BuildAuthHandler(dbWrapper)

	r := gin.Default()
	authHandler.BindRoutes(r)

	app := bootstrap.NewApp(r, ":8080")
	app.RegisterCloser(dbWrapper)

	// запуск с graceful shutdown
	app.RunWithGracefulShutdown()
}
