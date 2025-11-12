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

	jwksHandler := helpers.BuildJwksHandler(cfg.JWTPublicKey)
	userHandler := helpers.BuildUserHandler(dbWrapper)
	authHandler := helpers.BuildAuthHandler(dbWrapper, cfg.JWTPrivateKey)

	r := gin.Default()
	auth := r.Group("/auth")
	jwksHandler.BindRoutes(auth)
	userHandler.BindRoutes(auth)
	authHandler.BindRoutes(auth)

	app := bootstrap.NewApp(r, ":8080")
	app.RegisterCloser(dbWrapper)

	app.RunWithGracefulShutdown()
}
