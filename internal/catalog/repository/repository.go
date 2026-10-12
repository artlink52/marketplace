package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/artlink52/marketplace/internal/catalog/domain"
	dbcatalog "github.com/artlink52/marketplace/internal/catalog/repository/sqlc"
	"github.com/jackc/pgx/v5/pgconn"
)

type CatalogRepository struct {
	q *dbcatalog.Queries
}

func New(q *dbcatalog.Queries) *CatalogRepository {
	return &CatalogRepository{q: q}
}

func (r *CatalogRepository) CreateCategory(ctx context.Context, c domain.Category) (domain.Category, error) {
	const op = "repository.CreateCategory"

	row, err := r.q.CreateCategory(ctx, dbcatalog.CreateCategoryParams{
		Name:     c.Name,
		Slug:     c.Slug,
		ParentID: c.ParentID,
	})
	if err != nil {
		switch {
		case isUniqueViolation(err):
			return domain.Category{}, domain.ErrCategoryAlreadyExists
		case isForeignKeyViolation(err):
			return domain.Category{}, domain.ErrParentNotFound
		default:
			return domain.Category{}, fmt.Errorf("%s: %w", op, err)
		}
	}
	return domain.Category{
		ID:       row.ID,
		Name:     row.Name,
		Slug:     row.Slug,
		ParentID: row.ParentID,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
