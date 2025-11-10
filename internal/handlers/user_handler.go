package handlers

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/middlewares"
	"app/internal/services"
	"app/internal/stores"
	"app/internal/uows"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	uow              uows.UserTokeOutboxUnitOfWork
	requestValidator *middlewares.RequestValidator
	userService      *services.UserService
	outboxService    *services.OutboxService
}

func NewUserHandler(
	uow uows.UserTokeOutboxUnitOfWork,
	requestValidator *middlewares.RequestValidator,
	userService *services.UserService,
	outboxService *services.OutboxService,
) *UserHandler {
	return &UserHandler{
		requestValidator: requestValidator,
		uow:              uow,
		userService:      userService,
		outboxService:    outboxService,
	}
}

func (h *UserHandler) BindRoutes(r *gin.RouterGroup) {
	r.POST("/register", h.Register)
	r.POST("/promote-to-seller", h.PromoteToSeller)
	r.GET("/me", h.GetMe)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !h.requestValidator.ValidateRequest(c, &req) {
		return
	}
	var user *domain.User
	var err error
	err = h.uow.DoTransaction(func(txStore stores.UserTokenOutboxStore) error {
		user, err = h.userService.Register(
			txStore,
			req.Name,
			req.Surname,
			req.Email,
			req.Password,
		)
		if err != nil {
			return err
		}

		return h.outboxService.SaveUserRegisteredEvent(txStore, user)
	})

	resp := dto.APIResponse{
		Errors: make(map[string]string),
	}
	status := http.StatusCreated

	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserAlreadyExists):
			resp.Errors["error"] = "ERR_USER_EXISTS"
			status = http.StatusConflict
		default:
			resp.Errors["error"] = "ERR_INTERNAL"
			status = http.StatusInternalServerError
		}
		c.JSON(status, resp)
		return
	}

	resp.Success = true
	resp.Data = dto.UserResponse{
		User: user,
	}

	c.JSON(status, resp)
}

func (h *UserHandler) PromoteToSeller(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	var user *domain.User
	var err error
	err = h.uow.DoTransaction(func(txStore stores.UserTokenOutboxStore) error {
		user, err = h.userService.PromoteToSeller(
			txStore,
			userID,
		)
		if err != nil {
			return err
		}

		return err
	})

	resp := dto.APIResponse{
		Errors: make(map[string]string),
	}
	status := http.StatusOK

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			resp.Errors["error"] = "ERR_BAD_REQUEST"
			status = http.StatusUnauthorized
		default:
			resp.Errors["error"] = "ERR_INTERNAl"
			status = http.StatusInternalServerError
		}
		c.JSON(status, resp)
		return
	}

	resp.Success = true
	resp.Data = dto.UserResponse{
		User: user,
	}

	c.JSON(status, resp)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")

	var user *domain.User
	var err error
	err = h.uow.DoTransaction(func(txStore stores.UserTokenOutboxStore) error {
		user, err = h.userService.GetByID(
			txStore,
			userID,
		)
		if err != nil {
			return err
		}

		return err
	})

	resp := dto.APIResponse{
		Errors: make(map[string]string),
	}
	status := http.StatusOK

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			resp.Errors["error"] = "ERR_INVALID_CREDENTIALS"
			status = http.StatusUnauthorized
		default:
			resp.Errors["error"] = "ERR_INTERNAL"
			status = http.StatusInternalServerError
		}
		c.JSON(status, resp)
		return
	}

	resp.Success = true
	resp.Data = dto.UserResponse{
		User: user,
	}

	c.JSON(status, resp)
}
