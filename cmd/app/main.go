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

	userHandler := helpers.BuildUserHandler(dbWrapper)
	jwksHandler := helpers.BuildJwksHandler(cfg.JWTPublicKey)
	authHandler := helpers.BuildAuthHandler(dbWrapper, cfg.JWTPrivateKey)

	r := gin.Default()
	auth := r.Group("/auth")
	userHandler.BindRoutes(auth)
	jwksHandler.BindRoutes(auth)
	authHandler.BindRoutes(auth)

	app := bootstrap.NewApp(r, ":8080")
	app.RegisterCloser(dbWrapper)

	app.RunWithGracefulShutdown()
}
