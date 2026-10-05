package userclient

import (
	"context"
	"fmt"
	"time"

	userv1 "github.com/artlink52/marketplace/gen/go/user/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	Role   string
	Status string
)

const (
	RoleCustomer Role = "customer"
	RoleSeller   Role = "seller"
	RoleAdmin    Role = "admin"

	StatusActive              Status = "active"
	StatusBlocked             Status = "blocked"
	StatusPendingVerification Status = "pending_verification"
)

type Client struct {
	user userv1.UserServiceClient
	conn *grpc.ClientConn
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string
	Role      Role
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     Role   `json:"role"`
}

type AuthenticateInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to user.go-service: %w", err)
	}

	return &Client{
		user: userv1.NewUserServiceClient(conn),
		conn: conn,
	}, nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close user.go-service: %w", err)
	}

	return nil
}

func (c *Client) Register(ctx context.Context, input RegisterInput) (string, error) {
	resp, err := c.user.Register(ctx, &userv1.RegisterRequest{
		Email:    input.Email,
		Password: input.Password,
		Role:     toGrpcRole(input.Role),
	})
	if err != nil {
		return "", fmt.Errorf("register user: %w", err)
	}

	return resp.GetUserId(), nil
}

func (c *Client) Authenticate(ctx context.Context, input AuthenticateInput) (string, Role, error) {
	resp, err := c.user.Authenticate(ctx, &userv1.AuthenticateRequest{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return "", "", fmt.Errorf("authenticate user: %w", err)
	}

	return resp.UserId, toGatewayRole(resp.Role), nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (User, error) {
	resp, err := c.user.GetUser(ctx, &userv1.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}

	id, err := uuid.Parse(resp.User.Id)
	if err != nil {
		return User{}, fmt.Errorf("parse user id: %w", err)
	}

	return User{
		ID:        id,
		Email:     resp.User.Email,
		Role:      toGatewayRole(resp.User.Role),
		Status:    StatusActive,
		CreatedAt: resp.User.CreatedAt.AsTime(),
		UpdatedAt: resp.User.UpdatedAt.AsTime(),
	}, nil
}

func toGrpcRole(role Role) userv1.UserRole {
	switch role {
	case RoleCustomer:
		return userv1.UserRole_USER_ROLE_CUSTOMER
	case RoleSeller:
		return userv1.UserRole_USER_ROLE_SELLER
	case RoleAdmin:
		return userv1.UserRole_USER_ROLE_ADMIN
	default:
		return userv1.UserRole_USER_ROLE_CUSTOMER
	}
}

func toGatewayRole(userRole userv1.UserRole) Role {
	switch userRole {
	case userv1.UserRole_USER_ROLE_CUSTOMER:
		return RoleCustomer
	case userv1.UserRole_USER_ROLE_ADMIN:
		return RoleAdmin
	case userv1.UserRole_USER_ROLE_SELLER:
		return RoleSeller
	default:
		return RoleCustomer
	}
}
