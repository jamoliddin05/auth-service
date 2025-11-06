package services

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/mocks"
	"app/internal/repositories"
	"gorm.io/gorm"
	"testing"
	"time"

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
	mockUserRepo.On("Save", mock.Anything).Return(func(u *domain.User) error {
		u.ID = uuid.New()
		return nil
	})
	mockEventRepo.On("Save", "UserRegistered", mock.Anything).Return(nil)

	mockUow.On("DoRegistration", mock.Anything).Run(func(args mock.Arguments) {
		cb := args.Get(0).(func(repositories.UserRepository, repositories.EventRepository) error)
		_ = cb(mockUserRepo, mockEventRepo)
	}).Return(nil)

	authSvc := NewAuthService(mockUow, mockHasher, mockJwtHelper)
	req := dto.RegisterRequest{Phone: "+1234567890", Password: "password123"}

	resp, err := authSvc.Register(req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.User.ID)
	assert.Equal(t, req.Phone, resp.User.Phone)

	roles := make([]string, len(resp.User.Roles))
	for i, r := range resp.User.Roles {
		roles[i] = r.Role
	}
	assert.Contains(t, roles, "customer")
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
	mockStore.On("Tokens").Return(mockTokenRepo)
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

func TestAuthService_Refresh_Success(t *testing.T) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockTokenRepo := new(mocks.TokenRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)
	mockJwtHelper := new(mocks.JWTHelperMock)

	userID := uuid.New()
	existingUser := &domain.User{
		ID: userID,
		Roles: []domain.UserRole{
			{Role: "customer"},
		},
	}

	oldToken := &domain.Token{
		ID:        1,
		UserID:    userID,
		TokenHash: "old_hash",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	req := dto.RefreshRequest{
		RefreshToken: "valid_refresh_token",
	}

	// === Mock setup ===
	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)
	mockStore.On("Tokens").Return(mockTokenRepo)

	// user exists
	mockUserRepo.On("GetByID", userID).Return(existingUser, nil)

	// tokens exist for that user
	mockTokenRepo.On("GetByUserID", userID.String()).Return([]*domain.Token{oldToken}, nil)

	// refresh token matches existing hash
	mockHasher.On("Verify", req.RefreshToken, "old_hash").Return(true)

	// hashing new token (any string input returns "new_hashed_token")
	mockHasher.On("Hash", mock.Anything).Return("new_hashed_token")

	// generating new access token
	mockJwtHelper.On("GenerateAccessToken", userID.String(), []string{"customer"}).Return("new_access_token", nil)

	// saving updated token
	mockTokenRepo.On("Save", mock.AnythingOfType("*domain.Token")).Return(nil)

	// === Execute ===
	authSvc := NewAuthService(mockUow, mockHasher, mockJwtHelper)
	resp, err := authSvc.Refresh(req, userID.String())

	// === Assertions ===
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "new_access_token", resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	mockUow.AssertExpectations(t)
	mockStore.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
	mockJwtHelper.AssertExpectations(t)
}

func TestAuthService_Refresh_InvalidUserID(t *testing.T) {
	authSvc := NewAuthService(nil, nil, nil)
	req := dto.RefreshRequest{}
	_, err := authSvc.Refresh(req, "not-a-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestAuthService_Refresh_NoMatchingToken(t *testing.T) {
	mockUow := new(mocks.UnitOfWorkMock)
	mockStore := new(mocks.StoreMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockTokenRepo := new(mocks.TokenRepositoryMock)
	mockHasher := new(mocks.PasswordHasherMock)

	userID := uuid.New()
	existingUser := &domain.User{ID: userID}
	req := dto.RefreshRequest{RefreshToken: "badtoken"}

	mockUow.On("Store").Return(mockStore)
	mockStore.On("Users").Return(mockUserRepo)
	mockStore.On("Tokens").Return(mockTokenRepo)

	mockUserRepo.On("GetByID", userID).Return(existingUser, nil)
	mockTokenRepo.On("GetByUserID", userID.String()).Return([]*domain.Token{
		{TokenHash: "hash1"},
	}, nil)
	mockHasher.On("Verify", req.RefreshToken, "hash1").Return(false)

	authSvc := NewAuthService(mockUow, mockHasher, nil)
	_, err := authSvc.Refresh(req, userID.String())

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}
