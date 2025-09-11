package handlers

import (
	"net/http"

	"app/internal/dto"
	"app/internal/services"

	"github.com/gin-gonic/gin"
)

// GinUserHandler is the Gin adapter for UserService
type GinUserHandler struct {
	userService *services.UserService
}

// NewGinUserHandler creates a new Gin adapter
func NewGinUserHandler(userService *services.UserService) *GinUserHandler {
	return &GinUserHandler{userService: userService}
}

// BindRoutes registers the routes with Gin
func (h *GinUserHandler) BindRoutes(r *gin.Engine) {
	r.POST("/auth/register", h.Register)
}

// Register handles the /register route
func (h *GinUserHandler) Register(c *gin.Context) {
	var req dto.UserCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.userService.CreateUser(req)
	if err != nil {
		JSONError(c, err.Error(), http.StatusInternalServerError)
		return
	}

	JSONSuccess(c, resp, http.StatusCreated)
}
