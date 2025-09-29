package handlers

import (
	"app/internal/domain"
	"app/internal/mocks"
	"app/internal/repositories"
	"app/internal/services"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func registerTestValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("uzphone", func(fl validator.FieldLevel) bool {
			val := fl.Field().String()
			return strings.HasPrefix(val, "+998") && len(val) == 13
		})
	}
}

func TestGinAuthHandler_Register(t *testing.T) {
	registerTestValidators()

	t.Run("Invalid JSON", func(t *testing.T) {
		r := gin.Default()
		handler := &GinAuthHandler{authService: nil}
		r.POST("/register", handler.Register)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register", strings.NewReader(`{"phone": "123"`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
	})

	t.Run("User already exists", func(t *testing.T) {
		mockUow := new(mocks.UnitOfWorkMock)
		mockStore := new(mocks.StoreMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockEventRepo := new(mocks.EventRepositoryMock)
		mockHasher := new(mocks.PasswordHasherMock)
		mockJwtHelper := new(mocks.JWTHelperMock)

		mockUow.On("Store").Return(mockStore)
		mockStore.On("Users").Return(mockUserRepo)
		mockStore.On("Outbox").Return(mockEventRepo)

		handler := &GinAuthHandler{authService: services.NewAuthService(mockUow, mockHasher, mockJwtHelper)}

		mockUserRepo.On("GetByPhone", "+998901234567").Return(&domain.User{}, nil)

		r := gin.Default()
		r.POST("/register", handler.Register)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register",
			strings.NewReader(`{"phone": "+998901234567", "password": "password123"}`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "already exists")
	})

	t.Run("Success", func(t *testing.T) {
		mockUow := new(mocks.UnitOfWorkMock)
		mockStore := new(mocks.StoreMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockEventRepo := new(mocks.EventRepositoryMock)
		mockHasher := new(mocks.PasswordHasherMock)
		mockJwtHelper := new(mocks.JWTHelperMock)

		mockUow.On("Store").Return(mockStore)
		mockStore.On("Users").Return(mockUserRepo)
		mockStore.On("Outbox").Return(mockEventRepo)

		handler := &GinAuthHandler{authService: services.NewAuthService(mockUow, mockHasher, mockJwtHelper)}

		mockUserRepo.On("GetByPhone", "+998901234567").Return(nil, nil)
		mockUserRepo.On("Create", mock.Anything).Return(nil)
		mockEventRepo.On("Save", "UserRegistered", mock.Anything).Return(nil)
		mockHasher.On("Hash", "password123").Return("hashed_password")
		mockUow.On("DoRegistration", mock.Anything).Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(repositories.UserRepository, repositories.EventRepository) error)
			_ = cb(mockUserRepo, mockEventRepo) // pass mocks as interface
		}).Return(nil)

		r := gin.Default()
		r.POST("/register", handler.Register)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/register",
			strings.NewReader(`{"phone": "+998901234567", "password": "password123"}`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "+998901234567")
	})
}

func TestGinAuthHandler_Login(t *testing.T) {
	registerTestValidators()

	t.Run("Invalid credentials", func(t *testing.T) {
		mockUow := new(mocks.UnitOfWorkMock)
		mockStore := new(mocks.StoreMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockHasher := new(mocks.PasswordHasherMock)
		mockJwtHelper := new(mocks.JWTHelperMock)

		mockUow.On("Store").Return(mockStore)
		mockStore.On("Users").Return(mockUserRepo)

		handler := &GinAuthHandler{authService: services.NewAuthService(mockUow, mockHasher, mockJwtHelper)}

		// User not found
		mockUserRepo.On("GetByPhone", "+998901234567").Return(nil, gorm.ErrRecordNotFound)

		r := gin.Default()
		r.POST("/login", handler.Login)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login",
			strings.NewReader(`{"phone": "+998901234567", "password": "password123"}`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid credentials")
	})

	t.Run("Success", func(t *testing.T) {
		mockUow := new(mocks.UnitOfWorkMock)
		mockStore := new(mocks.StoreMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockEventRepo := new(mocks.EventRepositoryMock)
		mockTokenRepo := new(mocks.TokenRepositoryMock)
		mockHasher := new(mocks.PasswordHasherMock)
		mockJwtHelper := new(mocks.JWTHelperMock)

		mockUow.On("Store").Return(mockStore)
		mockStore.On("Users").Return(mockUserRepo)
		mockStore.On("Outbox").Return(mockEventRepo)

		handler := &GinAuthHandler{
			authService: services.NewAuthService(mockUow, mockHasher, mockJwtHelper),
		}

		user := &domain.User{
			ID:       userID(),
			Phone:    "+998901234567",
			Password: "hashed_password",
			Roles:    []domain.UserRole{{Role: "customer"}},
		}

		mockUserRepo.On("GetByPhone", "+998901234567").Return(user, nil)
		mockHasher.On("Verify", "password123", "hashed_password").Return(true)
		mockJwtHelper.On("GenerateAccessToken", user.ID.String(), mock.Anything).Return("jwt_token", nil)
		mockHasher.On("Hash", mock.Anything).Return("hashed_refresh_token")

		// Mock Save on tokenRepo and EventRepo
		mockTokenRepo.On("Save", mock.AnythingOfType("*domain.Token")).Return(nil)
		mockEventRepo.On("Save", "UserLoggedIn", user).Return(nil)

		// Use interface types in callback
		mockUow.On("DoLogin", mock.Anything).Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(repositories.TokenRepository, repositories.EventRepository) error)
			_ = cb(mockTokenRepo, mockEventRepo)
		}).Return(nil)

		r := gin.Default()
		r.POST("/login", handler.Login)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login",
			strings.NewReader(`{"phone": "+998901234567", "password": "password123"}`))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "jwt_token")
	})
}

// helper to generate a fixed UUID
func userID() uuid.UUID {
	id, _ := uuid.Parse("11111111-1111-1111-1111-111111111111")
	return id
}
