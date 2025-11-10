package handlers

import (
	"app/internal/middlewares"
	"app/internal/stores"
	"app/internal/uows"
	"errors"
	"net/http"

	"app/internal/dto"
	"app/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	uow              uows.UserTokeOutboxUnitOfWork
	requestValidator *middlewares.RequestValidator
	users            *services.UserService
	tokens           *services.TokenService
	outbox           *services.OutboxService
}

func NewAuthHandler(
	uow uows.UserTokeOutboxUnitOfWork,
	requestValidator *middlewares.RequestValidator,
	users *services.UserService,
	tokens *services.TokenService,
	outbox *services.OutboxService,
) *AuthHandler {
	return &AuthHandler{
		uow:              uow,
		requestValidator: requestValidator,
		users:            users,
		tokens:           tokens,
		outbox:           outbox,
	}
}

func (h *AuthHandler) BindRoutes(r *gin.RouterGroup) {
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !h.requestValidator.ValidateRequest(c, &req) {
		return
	}
	var accessToken string
	var refreshToken string
	err := h.uow.DoTransaction(func(store stores.UserTokenOutboxStore) error {
		user, err := h.users.Authenticate(store, req.Email, req.Password)
		if err != nil {
			return err
		}

		accessToken, refreshToken, err = h.tokens.IssueTokenForUser(store, user)
		if err != nil {
			return err
		}

		return h.outbox.SaveUserLoggedInEvent(store, user)
	})

	resp := dto.APIResponse{
		Errors: make(map[string]string),
	}
	status := http.StatusOK

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			resp.Errors["error"] = "ERR_INVALID_CREDENTIALS"
			status = http.StatusConflict
		default:
			resp.Errors["error"] = "ERR_INTERNAL"
			status = http.StatusInternalServerError
		}
		c.JSON(status, resp)
		return
	}

	resp.Success = true
	resp.Data = dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(status, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if !h.requestValidator.ValidateRequest(c, &req) {
		return
	}
	userID := c.GetHeader("X-User-Id")
	var accessToken string
	var refreshToken string
	err := h.uow.DoTransaction(func(store stores.UserTokenOutboxStore) error {
		user, err := h.users.GetByID(store, userID)
		if err != nil {
			return err
		}

		ok, err := h.tokens.VerifyRefreshToken(store, user.ID, req.RefreshToken)
		if err != nil || !ok {
			return services.ErrInvalidCredentials
		}

		accessToken, refreshToken, err = h.tokens.IssueTokenForUser(store, user)
		if err != nil {
			return err
		}

		return nil
	})

	resp := dto.APIResponse{
		Errors: make(map[string]string),
	}
	status := http.StatusOK

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			resp.Errors["error"] = "ERR_INVALID_CREDENTIALS"
			status = http.StatusConflict
		default:
			resp.Errors["error"] = "ERR_INTERNAL"
			status = http.StatusInternalServerError
		}
		c.JSON(status, resp)
		return
	}

	resp.Success = true
	resp.Data = dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(status, resp)
}
