package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type JwksHandler struct {
	jwks string
}

func NewJwksHandler(jwks string) *JwksHandler {
	return &JwksHandler{jwks: jwks}
}

func (h *JwksHandler) BindRoutes(r *gin.RouterGroup) {
	r.GET("/.well-known/jwks.json", h.JWKS)
}

func (h *JwksHandler) JWKS(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", []byte(h.jwks))
}
