package services

import (
	"app/internal/domain"
	"app/internal/dto"
	"app/internal/mappers"
	"app/internal/repositories"
	"github.com/google/uuid"
)

// UserService handles business logic for users
type UserService struct {
	repo repositories.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser creates a new user and returns a DTO
func (s *UserService) CreateUser(req dto.UserCreateRequest) (*dto.UserResponse, error) {
	user := mappers.DTOToUser(req)
	user.Roles = []domain.UserRole{
		{Role: domain.RoleCustomer},
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	resp := mappers.UserToDTO(user)

	return resp, nil
}

// GetUserByID fetches a user and returns a DTO
func (s *UserService) GetUserByID(id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	resp := mappers.UserToDTO(user)

	return resp, nil
}
