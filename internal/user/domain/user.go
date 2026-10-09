package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleSeller   UserRole = "seller"
	RoleAdmin    UserRole = "admin"
)

type UserStatus string

const (
	StatusActive              UserStatus = "active"
	StatusBlocked             UserStatus = "blocked"
	StatusPendingVerification UserStatus = "pending_verification"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrRoleNotAllowed     = errors.New("role not allowed for self-registration")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserBlocked        = errors.New("user is blocked")
)

type NewUser struct {
	Email        string
	PasswordHash string
	Role         UserRole
}

type User struct {
	ID        uuid.UUID
	Email     string
	Role      UserRole
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AuthUser struct {
	ID           uuid.UUID
	PasswordHash string
	Role         UserRole
	Status       UserStatus
}

type RegisterUser struct {
	Email    string
	Password string
	Role     UserRole
}

type LoginUser struct {
	Email    string
	Password string
}
