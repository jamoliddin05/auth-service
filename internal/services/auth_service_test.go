package services

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/mocks"
	"app/internal/repositories"
	"gorm.io/gorm"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Register_Success(t *testing.T) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventRepo := new(mocks.EventRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)
	mockJwtHelper := new(mocks.JWTHelperMock)

	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)
	mockStore.On("Outbox").Return(mockEventRepo)

	mockHasher.On("Hash", "password123").Return("hashed_password")
	mockUserRepo.On("GetByPhone", "+1234567890").Return(nil, nil)
	mockUserRepo.On("Create", mock.Anything).Return(nil)
	mockEventRepo.On("Save", "UserRegistered", mock.Anything).Return(nil)

	mockUow.On("DoRegistration", mock.Anything).Run(func(args mock.Arguments) {
		cb := args.Get(0).(func(repositories.UserRepository, repositories.EventRepository) error)
		_ = cb(mockUserRepo, mockEventRepo)
	}).Return(nil)

	authSvc := NewAuthService(mockUow, mockHasher, mockJwtHelper)
	req := dto.RegisterRequest{Phone: "+1234567890", Password: "password123"}

	resp, err := authSvc.Register(req)

	assert.NoError(t, err)
	assert.Equal(t, req.Phone, resp.Phone)
	assert.NotEmpty(t, resp.ID)
	assert.Contains(t, resp.Roles, "customer")
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)
	mockJwtHelper := new(mocks.JWTHelperMock)

	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)

	existingUser := &domain.User{ID: uuid.New()}
	mockUserRepo.On("GetByPhone", "+1234567890").Return(existingUser, nil)

	authSvc := NewAuthService(mockUow, mockHasher, mockJwtHelper)
	req := dto.RegisterRequest{Phone: "+1234567890", Password: "password123"}

	_, err := authSvc.Register(req)
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockTokenRepo := new(mocks.TokenRepositoryMock)
	mockEventRepo := new(mocks.EventRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)
	mockJwtHelper := new(mocks.JWTHelperMock)

	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)
	mockStore.On("Outbox").Return(mockEventRepo)

	existingUser := &domain.User{
		ID:       uuid.New(),
		Phone:    "+1234567890",
		Password: "hashed_password",
		Roles:    []domain.UserRole{{Role: "customer"}},
	}

	mockTokenRepo.On("Save", mock.Anything).Return(nil)
	mockEventRepo.On("Save", "UserLoggedIn", existingUser).Return(nil)
	mockUserRepo.On("GetByPhone", "+1234567890").Return(existingUser, nil)
	mockHasher.On("Verify", "password123", "hashed_password").Return(true)
	mockHasher.On("Hash", mock.Anything).Return("hashed_token")
	mockJwtHelper.On("GenerateAccessToken", existingUser.ID.String(), mock.Anything).Return("jwt_token", nil)

	mockUow.On("DoLogin", mock.Anything).Run(func(args mock.Arguments) {
		cb := args.Get(0).(func(repositories.TokenRepository, repositories.EventRepository) error)
		_ = cb(mockTokenRepo, mockEventRepo)
	}).Return(nil)

	authSvc := NewAuthService(mockUow, mockHasher, mockJwtHelper)
	req := dto.LoginRequest{Phone: "+1234567890", Password: "password123"}

	resp, err := authSvc.Login(req)
	assert.NoError(t, err)
	assert.Equal(t, "jwt_token", resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)
	mockJwtHelper := new(mocks.JWTHelperMock)

	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)

	existingUser := &domain.User{ID: uuid.New(), Password: "hashed_password"}
	mockUserRepo.On("GetByPhone", "+1234567890").Return(existingUser, nil)
	mockHasher.On("Verify", "wrong_password", "hashed_password").Return(false)

	authSvc := NewAuthService(mockUow, mockHasher, mockJwtHelper)
	req := dto.LoginRequest{Phone: "+1234567890", Password: "wrong_password"}

	_, err := authSvc.Login(req)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)
	mockJwtHelper := new(mocks.JWTHelperMock)

	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)

	mockUserRepo.On("GetByPhone", "+1234567890").Return(nil, gorm.ErrRecordNotFound)

	authSvc := NewAuthService(mockUow, mockHasher, mockJwtHelper)
	req := dto.LoginRequest{Phone: "+1234567890", Password: "password123"}

	_, err := authSvc.Login(req)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}
