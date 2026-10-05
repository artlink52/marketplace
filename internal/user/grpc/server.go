package grpc

import (
	"context"
	"errors"
	"net/mail"

	userv1 "github.com/artlink52/marketplace/gen/go/user/v1"
	"github.com/artlink52/marketplace/internal/user/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	minPasswordLen = 6
)

type UserService interface {
	Register(ctx context.Context, u domain.RegisterUser) (uuid.UUID, error)
	Authenticate(ctx context.Context, u domain.LoginUser) (domain.AuthUser, error)
	GetUser(ctx context.Context, id uuid.UUID) (domain.User, error)
}

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	userService UserService
}

func RegisterUserServer(server *grpc.Server, userService UserService) {
	userv1.RegisterUserServiceServer(server, &UserServer{userService: userService})
}

func (s *UserServer) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.userService.Register(ctx, domain.RegisterUser{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Role:     roleFromProto(req.GetRole()),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		case errors.Is(err, domain.ErrRoleNotAllowed):
			return nil, status.Error(codes.InvalidArgument, "role not allowed for self-registration")
		default:
			return nil, status.Error(codes.Internal, "failed to register user.go")
		}
	}

	return &userv1.RegisterResponse{UserId: userID.String()}, nil
}

func (s *UserServer) Authenticate(ctx context.Context, req *userv1.AuthenticateRequest) (*userv1.AuthenticateResponse, error) {
	if err := validateAuthenticate(req); err != nil {
		return nil, err
	}

	user, err := s.userService.Authenticate(ctx, domain.LoginUser{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		case errors.Is(err, domain.ErrUserBlocked):
			return nil, status.Error(codes.PermissionDenied, "user.go is blocked")
		default:
			return nil, status.Error(codes.Internal, "failed to authenticate user.go")
		}
	}

	return &userv1.AuthenticateResponse{
		UserId: user.ID.String(),
		Role:   roleToProto(user.Role),
	}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	user, err := s.userService.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user.go not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user.go")
	}

	return &userv1.GetUserResponse{
		User: &userv1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			Role:      roleToProto(user.Role),
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func roleToProto(r domain.UserRole) userv1.UserRole {
	switch r {
	case domain.RoleSeller:
		return userv1.UserRole_USER_ROLE_SELLER
	case domain.RoleAdmin:
		return userv1.UserRole_USER_ROLE_ADMIN
	default:
		return userv1.UserRole_USER_ROLE_CUSTOMER
	}
}

func roleFromProto(r userv1.UserRole) domain.UserRole {
	switch r {
	case userv1.UserRole_USER_ROLE_SELLER:
		return domain.RoleSeller
	case userv1.UserRole_USER_ROLE_ADMIN:
		return domain.RoleAdmin
	default:
		return domain.RoleCustomer
	}
}

func validateAuthenticate(req *userv1.AuthenticateRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if _, err := mail.ParseAddress(req.GetEmail()); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if len(req.GetPassword()) < minPasswordLen {
		return status.Error(codes.InvalidArgument, "password must be at least 6 characters")
	}

	return nil
}

func validateRegister(req *userv1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if _, err := mail.ParseAddress(req.GetEmail()); err != nil {
		return status.Error(codes.InvalidArgument, "invalid email format")
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}

	if len(req.GetPassword()) < minPasswordLen {
		return status.Error(codes.InvalidArgument, "password must be at least 6 characters")
	}

	return nil
}
