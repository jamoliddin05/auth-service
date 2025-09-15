package services

import (
	"app/internal/repositories"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"app/internal/domain"
	"app/internal/dto"
	"app/internal/mocks"
)

// buildMocks bootstraps needed mocks
func buildMocks() (*mocks.UnitOfWorkMock, *mocks.StoreMock, *mocks.UserRepositoryMock, *mocks.EventRepositoryMock, *mocks.PasswordHasherMock) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventRepo := new(mocks.EventRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)

	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)
	mockStore.On("Outbox").Return(mockEventRepo)

	return mockUow, mockStore, mockUserRepo, mockEventRepo, mockHasher
}

// Successfull Registration
func TestAuthService_Register_Success(t *testing.T) {
	mockUow, _, mockUserRepo, mockEventRepo, mockHasher := buildMocks()

	mockHasher.On("Hash", "password123").Return("hashed_password")
	mockUserRepo.On("GetByPhone", "+1234567890").Return(nil, nil)
	mockUserRepo.On("Create", mock.Anything).Return(nil)
	mockEventRepo.On("Save", "UserRegistered", mock.Anything).Return(nil)

	// DoRegistration
	mockUow.On("DoRegistration", mock.Anything).
		Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(repositories.UserRepository, repositories.EventRepository) error)
			_ = cb(mockUserRepo, mockEventRepo)
		}).
		Return(nil)

	authSvc := NewAuthService(mockUow, mockHasher)
	req := dto.RegisterRequest{Phone: "+1234567890", Password: "password123"}
	resp, err := authSvc.Register(req)

	assert.NoError(t, err)
	assert.Equal(t, req.Phone, resp.Phone)
	assert.NotEmpty(t, resp.ID)
	assert.Contains(t, resp.Roles, "customer")

	mockHasher.AssertCalled(t, "Hash", "password123")
	mockUserRepo.AssertCalled(t, "Create", mock.Anything)
	mockEventRepo.AssertCalled(t, "Save", "UserRegistered", mock.Anything)
	mockUserRepo.AssertCalled(t, "GetByPhone", "+1234567890")
	mockUow.AssertCalled(t, "DoRegistration", mock.Anything)
}

// User exists
func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	mockUow, _, mockUserRepo, _, mockHasher := buildMocks()

	existingUser := &domain.User{ID: uuid.New()}
	mockUserRepo.On("GetByPhone", "+1234567890").Return(existingUser, nil)

	authSvc := NewAuthService(mockUow, mockHasher)
	req := dto.RegisterRequest{Phone: "+1234567890", Password: "password123"}
	_, err := authSvc.Register(req)

	assert.ErrorIs(t, err, ErrUserAlreadyExists)
	mockUserRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// DB error on checking whether user exists
func TestAuthService_Register_GetByPhoneError(t *testing.T) {
	mockUow, _, mockUserRepo, _, mockHasher := buildMocks()

	mockUserRepo.On("GetByPhone", "+1234567890").Return(nil, errors.New("db error"))

	authSvc := NewAuthService(mockUow, mockHasher)
	req := dto.RegisterRequest{Phone: "+1234567890", Password: "password123"}
	_, err := authSvc.Register(req)

	assert.Error(t, err)
	mockUserRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// Error in the DoRegistration's transaction
func TestAuthService_Register_DoRegistrationError(t *testing.T) {
	mockUow, _, mockUserRepo, mockEventRepo, mockHasher := buildMocks()

	mockHasher.On("Hash", "password123").Return("hashed_password")
	mockUserRepo.On("GetByPhone", "+1234567890").Return(nil, nil)
	mockUserRepo.On("Create", mock.Anything).Return(nil)
	mockEventRepo.On("Save", "UserRegistered", mock.Anything).Return(nil)

	// DoRegistration calls the given callback, but returns error at the end
	mockUow.On("DoRegistration", mock.Anything).
		Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(repositories.UserRepository, repositories.EventRepository) error)
			_ = cb(mockUserRepo, mockEventRepo)
		}).
		Return(errors.New("transaction failed"))

	authSvc := NewAuthService(mockUow, mockHasher)
	req := dto.RegisterRequest{Phone: "+1234567890", Password: "password123"}
	_, err := authSvc.Register(req)

	assert.Error(t, err)
}

// Error when saving the event
func TestAuthService_Register_EventSaveError(t *testing.T) {
	mockUow, _, mockUserRepo, mockEventRepo, mockHasher := buildMocks()

	mockHasher.On("Hash", "password123").Return("hashed_password")
	mockUserRepo.On("GetByPhone", "+1234567890").Return(nil, nil)
	mockUserRepo.On("Create", mock.Anything).Return(nil)
	mockEventRepo.On("Save", "UserRegistered", mock.Anything).Return(errors.New("save failed"))

	// DoRegistration вызывает callback с моками
	mockUow.On("DoRegistration", mock.Anything).
		Run(func(args mock.Arguments) {
			cb := args.Get(0).(func(repositories.UserRepository, repositories.EventRepository) error)
			_ = cb(mockUserRepo, mockEventRepo) // Save will return error
		}).
		Return(errors.New("save failed"))

	authSvc := NewAuthService(mockUow, mockHasher)
	req := dto.RegisterRequest{Phone: "+1234567890", Password: "password123"}
	_, err := authSvc.Register(req)

	assert.Error(t, err)
}
