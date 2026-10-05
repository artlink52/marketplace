package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/artlink52/marketplace/internal/user/domain"
)

type UserRepository interface {
	Register(ctx context.Context, u domain.NewUser) (uuid.UUID, error)
	Authenticate(ctx context.Context, email string) (domain.AuthUser, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

type UserService struct {
	userRepository UserRepository
}

func New(userRepository UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

var selfRegisterableRoles = map[domain.UserRole]bool{
	domain.RoleCustomer: true,
	domain.RoleSeller:   true,
}

func (s *UserService) Register(ctx context.Context, u domain.RegisterUser) (uuid.UUID, error) {
	const op = "service.Register"

	if !selfRegisterableRoles[u.Role] {
		return uuid.Nil, domain.ErrRoleNotAllowed
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	id, err := s.userRepository.Register(ctx, domain.NewUser{
		Email:        u.Email,
		PasswordHash: string(passHash),
		Role:         u.Role,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *UserService) Authenticate(ctx context.Context, u domain.LoginUser) (domain.AuthUser, error) {
	authUser, err := s.userRepository.Authenticate(ctx, u.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.AuthUser{}, domain.ErrInvalidCredentials
		}
		return domain.AuthUser{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(authUser.PasswordHash), []byte(u.Password)); err != nil {
		return domain.AuthUser{}, domain.ErrInvalidCredentials
	}

	if authUser.Status == domain.StatusBlocked {
		return domain.AuthUser{}, domain.ErrUserBlocked
	}

	return authUser, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return s.userRepository.GetUserByID(ctx, id)
}
