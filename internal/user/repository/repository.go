package repository

import (
	"context"
	"errors"

	"github.com/artlink52/marketplace/internal/user/domain"
	dbuser "github.com/artlink52/marketplace/internal/user/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	q *dbuser.Queries
}

func New(q *dbuser.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) Register(ctx context.Context, u domain.NewUser) (uuid.UUID, error) {
	id, err := r.q.CreateUser(ctx, dbuser.CreateUserParams{
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         dbuser.UserRole(u.Role),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return uuid.Nil, domain.ErrEmailAlreadyExists
		}
		return uuid.Nil, err
	}

	return id, nil
}

func (r *UserRepository) Authenticate(ctx context.Context, email string) (domain.AuthUser, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AuthUser{}, domain.ErrUserNotFound
		}
		return domain.AuthUser{}, err
	}

	return domain.AuthUser{
		ID:           row.ID,
		PasswordHash: row.PasswordHash,
		Role:         domain.UserRole(row.Role),
		Status:       domain.UserStatus(row.Status),
	}, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return domain.User{
		ID:        id,
		Email:     row.Email,
		Role:      domain.UserRole(row.Role),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
