package services

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/stores"
	"app/internal/uows"
	"errors"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	uow    uows.UserTokenUnitOfWork
	users  *UserService
	tokens *TokenService
}

func NewAuthService(
	uow uows.UserTokenUnitOfWork,
	users *UserService,
	tokens *TokenService,
) *AuthService {
	return &AuthService{
		uow:    uow,
		users:  users,
		tokens: tokens,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.UserResponse, error) {
	var user *domain.User

	err := s.withinTx(func(store stores.UserTokenStore) error {
		var err error
		user, err = s.users.Register(store, req)
		if err != nil {
			return err
		}
		return store.Outbox().Save("UserRegistered", user)
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{User: *user}, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.TokenResponse, error) {
	tokens := &dto.TokenResponse{}

	err := s.withinTx(func(store stores.UserTokenStore) error {
		user, err := s.users.Authenticate(store, req.Phone, req.Password)
		if err != nil {
			return err
		}

		accessToken, refreshToken, err := s.tokens.IssueTokenForUser(store, user)
		tokens.AccessToken = accessToken
		tokens.RefreshToken = refreshToken
		if err != nil {
			return err
		}

		return store.Outbox().Save("UserLoggedIn", user)
	})
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) PromoteToSeller(userId string) (*dto.TokenResponse, error) {
	tokens := &dto.TokenResponse{}

	err := s.withinTx(func(store stores.UserTokenStore) error {
		user, err := s.users.PromoteToSeller(store, userId)
		if err != nil {
			return err
		}

		accessToken, refreshToken, err := s.tokens.IssueTokenForUser(store, user)
		tokens.AccessToken = accessToken
		tokens.RefreshToken = refreshToken
		if err != nil {
			return err
		}

		return store.Outbox().Save("UserBecameSeller", user)
	})
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) Refresh(req dto.RefreshRequest, userId string) (*dto.TokenResponse, error) {
	tokens := &dto.TokenResponse{}

	err := s.withinTx(func(store stores.UserTokenStore) error {
		user, err := s.users.GetByID(store, userId)
		if err != nil {
			return err
		}

		ok, err := s.tokens.VerifyRefreshToken(store, user.ID, req.RefreshToken)
		if err != nil || !ok {
			return ErrInvalidCredentials
		}

		accessToken, refreshToken, err := s.tokens.IssueTokenForUser(store, user)
		tokens.AccessToken = accessToken
		tokens.RefreshToken = refreshToken
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) GetMe(userId string) (*dto.UserResponse, error) {
	var user *domain.User

	err := s.withinTx(func(store stores.UserTokenStore) error {
		var err error
		user, err = s.users.GetByID(store, userId)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrInvalidCredentials
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{User: *user}, nil
}

func (s *AuthService) withinTx(fn func(txStore stores.UserTokenStore) error) error {
	return s.uow.DoTransaction(fn)
}
