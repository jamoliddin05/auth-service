package handlers

import (
	"errors"
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
	jwksStr     string
}

func NewGinAuthHandler(authService *services.AuthService, jwksStr string) *GinAuthHandler {
	return &GinAuthHandler{
		authService: authService,
		jwksStr:     jwksStr,
	}
}

func (h *GinAuthHandler) BindRoutes(r *gin.Engine) {
	auth := r.Group("/auth")

	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/become-seller", h.BecomeSeller)
	auth.GET("/me", h.GetMe)
	auth.GET("/.well-known/jwks.json", h.JWKS)

}

func (h *GinAuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	vr := h.bindAndValidate(c, &req)
	if !vr.Valid {
		h.respondValidationError(c, vr)
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserAlreadyExists):
			JSONError(c, http.StatusConflict, map[string]string{"code": string(ErrUserExists)})
		default:
			JSONError(c, http.StatusInternalServerError, map[string]string{"code": string(ErrInternal)})
		}
		return
	}

	JSONSuccess(c, resp, http.StatusCreated)

}

func (h *GinAuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	vr := h.bindAndValidate(c, &req)
	if !vr.Valid {
		h.respondValidationError(c, vr)
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			JSONError(c, http.StatusUnauthorized, map[string]string{"code": string(ErrInvalidCredentials)})
		default:
			JSONError(c, http.StatusInternalServerError, map[string]string{"code": string(ErrInternal)})
		}
		return
	}

	JSONSuccess(c, resp, http.StatusOK)

}

func (h *GinAuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	vr := h.bindAndValidate(c, &req)
	if !vr.Valid {
		h.respondValidationError(c, vr)
		return
	}

	userID := c.GetHeader("X-User-Id")
	resp, err := h.authService.Refresh(req, userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			JSONError(c, http.StatusUnauthorized, map[string]string{"code": string(ErrInvalidCredentials)})
		default:
			JSONError(c, http.StatusInternalServerError, map[string]string{"code": string(ErrInternal)})
		}
		return
	}

	JSONSuccess(c, resp, http.StatusOK)

}

func (h *GinAuthHandler) BecomeSeller(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	resp, err := h.authService.BecomeSeller(userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			JSONError(c, http.StatusUnauthorized, map[string]string{"code": string(ErrInvalidCredentials)})
		default:
			JSONError(c, http.StatusInternalServerError, map[string]string{"code": string(ErrInternal)})
		}
		return
	}

	JSONSuccess(c, resp, http.StatusOK)

}

func (h *GinAuthHandler) GetMe(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")

	resp, err := h.authService.GetMe(userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			JSONError(c, http.StatusUnauthorized, map[string]string{"code": string(ErrInvalidCredentials)})
		default:
			JSONError(c, http.StatusInternalServerError, map[string]string{"code": string(ErrInternal)})
		}
		return
	}

	JSONSuccess(c, resp, http.StatusOK)

}

func (h *GinAuthHandler) JWKS(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", []byte(h.jwksStr))
}

func (h *GinAuthHandler) bindAndValidate(c *gin.Context, req any) ValidationResult {
	if err := c.ShouldBindJSON(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			errorsMap := make(map[string]string)
			for _, fe := range ve {
				field := strings.ToLower(fe.Field())
				errorsMap[field] = codeForTag(fe)
			}
			return ValidationResult{Valid: false, Errors: errorsMap}
		}
		return ValidationResult{Valid: false, Err: err}
	}
	return ValidationResult{Valid: true}
}

func (h *GinAuthHandler) respondValidationError(c *gin.Context, vr ValidationResult) {
	errorsMap := make(map[string]string)

	// Use field-specific errors if available
	if len(vr.Errors) > 0 {
		for k, v := range vr.Errors {
			errorsMap[k] = v
		}
	}

	// Fallback for general errors if no field-specific ones exist
	if vr.Err != nil || len(errorsMap) == 0 {
		errorsMap["code"] = string(ErrBadRequest)
	}

	JSONError(c, http.StatusBadRequest, errorsMap)
}

// --- Error Code Mapper ---
func codeForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return string(ErrRequired)
	case "min":
		return string(ErrMinLength)
	case "max":
		return string(ErrMaxLength)
	case "uzphone":
		return string(ErrInvalidPhone)
	default:
		return string(ErrInvalidField)
	}
}
