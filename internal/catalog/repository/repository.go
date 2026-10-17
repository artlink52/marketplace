package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/artlink52/marketplace/internal/catalog/domain"
	dbcatalog "github.com/artlink52/marketplace/internal/catalog/repository/sqlc"
	"github.com/google/uuid"
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

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	const op = "repository.ListCategories"
	rows, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var result []domain.Category
	for _, row := range rows {
		result = append(result, domain.Category{
			ID:       row.ID,
			Name:     row.Name,
			Slug:     row.Slug,
			ParentID: row.ParentID,
		})
	}
	return result, nil
}

func (r *CatalogRepository) CreateProduct(ctx context.Context, p domain.Product) (uuid.UUID, error) {
	id, err := r.q.CreateProduct(ctx, dbcatalog.CreateProductParams{
		SellerID:      p.SellerID,
		Title:         p.Title,
		Description:   p.Description,
		CategoryID:    p.CategoryID,
		PriceAmount:   p.Price.Amount,
		PriceCurrency: p.Price.Currency,
	})
	if err != nil {
		switch {
		case isUniqueViolation(err):
			return uuid.Nil, domain.ErrProductAlreadyExists
		case isForeignKeyViolation(err):
			return uuid.Nil, domain.ErrCategoryNotFound
		}
	}

	return id, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
