package handlers

import (
	"net/http"

	"app/internal/dto"
	"app/internal/services"

	"github.com/gin-gonic/gin"
)

// GinAuthHandler is the Gin adapter for UserService
type GinAuthHandler struct {
	authService *services.AuthService
}

// NewGinAuthHandler creates a new Gin adapter
func NewGinAuthHandler(authService *services.AuthService) *GinAuthHandler {
	return &GinAuthHandler{authService: authService}
}

// BindRoutes registers the routes with Gin
func (h *GinAuthHandler) BindRoutes(r *gin.Engine) {
	r.POST("/register", h.Register)
}

// Register handles the /register route
func (h *GinAuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		JSONError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	JSONSuccess(c, resp, http.StatusCreated)
}
