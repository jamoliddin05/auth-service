package handlers

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"

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
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			errorsMap := make(map[string]string)
			for _, fe := range ve {
				errorsMap[strings.ToLower(fe.Field())] = msgForTag(fe)
			}
			JSONError(c, "Validation failed", http.StatusBadRequest, errorsMap)
			return
		}

		JSONError(c, err.Error(), http.StatusBadRequest, nil)
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserAlreadyExists):
			JSONError(c, err.Error(), http.StatusConflict, nil)
		default:
			JSONError(c, err.Error(), http.StatusInternalServerError, nil)
		}
		return
	}

	JSONSuccess(c, resp, http.StatusCreated)
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "uzphone":
		return "Phone number must be a valid Uzbekistan phone number"
	default:
		return fmt.Sprintf("%s is not valid", fe.Field())
	}
}
