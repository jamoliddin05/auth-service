package main

import (
	"app/bootstrap"
	"app/bootstrap/configs"
	"app/bootstrap/helpers"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := configs.LoadConfig()

	dbWrapper := helpers.MustInitDB(cfg)
	helpers.MustRegisterValidators()

	authHandler := helpers.BuildAuthHandler(dbWrapper, cfg.JWTPrivateKey, cfg.JWTPublicKey)

	r := gin.Default()
	authHandler.BindRoutes(r)

	app := bootstrap.NewApp(r, ":8080")
	app.RegisterCloser(dbWrapper)

	app.RunWithGracefulShutdown()
}
