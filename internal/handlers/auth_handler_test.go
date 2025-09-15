package handlers

import (
	"app/internal/domain"
	"app/internal/mocks"
	"app/internal/repositories"
	"app/internal/services"
	"errors"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func registerTestValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("uzphone", func(fl validator.FieldLevel) bool {
			val := fl.Field().String()
			return strings.HasPrefix(val, "+998") && len(val) == 13
		})
	}
}

func buildHandlerWithMocks() (*GinAuthHandler, *mocks.UnitOfWorkMock, *mocks.UserRepositoryMock, *mocks.EventRepositoryMock, *mocks.PasswordHasherMock) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventRepo := new(mocks.EventRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)

	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)
	mockStore.On("Outbox").Return(mockEventRepo)

	authSvc := services.NewAuthService(mockUow, mockHasher)
	handler := &GinAuthHandler{authService: authSvc}

	return handler, mockUow, mockUserRepo, mockEventRepo, mockHasher
}

func TestGinAuthHandler_Register(t *testing.T) {
	registerTestValidators()

	t.Run("Invalid JSON", func(t *testing.T) {
		handler := &GinAuthHandler{authService: nil}
		r := gin.Default()
		r.POST("/register", handler.Register)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"phone": "123"`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})

	t.Run("Invalid phone", func(t *testing.T) {
		handler := &GinAuthHandler{authService: nil}
		r := gin.Default()
		r.POST("/register", handler.Register)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"phone": "123", "password": "password123"}`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "phone")
	})

	t.Run("User already exists", func(t *testing.T) {
		handler, _, mockUserRepo, _, _ := buildHandlerWithMocks()
		mockUserRepo.On("GetByPhone", "+998901234567").Return(&domain.User{}, nil)

		r := gin.Default()
		r.POST("/register", handler.Register)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"phone": "+998901234567", "password": "password123"}`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already exists")
	})

	t.Run("Internal service error", func(t *testing.T) {
		handler, mockUow, mockUserRepo, _, mockHasher := buildHandlerWithMocks()

		mockUserRepo.On("GetByPhone", "+998901234567").Return(nil, nil)
		mockHasher.On("Hash", "password123").Return("hashed_password") // <== добавлено
		mockUow.On("DoRegistration", mock.Anything).Return(errors.New("internal error"))

		r := gin.Default()
		r.POST("/register", handler.Register)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"phone": "+998901234567", "password": "password123"}`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal error")
	})

	t.Run("Success", func(t *testing.T) {
		handler, mockUow, mockUserRepo, mockEventRepo, mockHasher := buildHandlerWithMocks()
		mockUserRepo.On("GetByPhone", "+998901234567").Return(nil, nil)
		mockUserRepo.On("Create", mock.Anything).Return(nil)
		mockEventRepo.On("Save", "UserRegistered", mock.Anything).Return(nil)
		mockHasher.On("Hash", "password123").Return("hashed_password")

		mockUow.On("DoRegistration", mock.Anything).Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(repositories.UserRepository, repositories.EventRepository) error)
			_ = cb(mockUserRepo, mockEventRepo)
		}).Return(nil)

		r := gin.Default()
		r.POST("/register", handler.Register)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"phone": "+998901234567", "password": "password123"}`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "+998901234567")
	})
}
