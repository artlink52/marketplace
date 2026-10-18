package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/artlink52/marketplace/internal/catalog/domain"
	dbcatalog "github.com/artlink52/marketplace/internal/catalog/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	const op = "repository.CreateProduct"
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
		default:
			return uuid.Nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return id, nil
}

func (r *CatalogRepository) BatchGetProducts(ctx context.Context, ids []uuid.UUID) ([]domain.Product, error) {
	const op = "repository.BatchGetProducts"
	rows, err := r.q.BatchGetProducts(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainProduct(row))
	}
	return result, nil
}

func (r *CatalogRepository) GetProduct(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	const op = "repository.GetProduct"
	row, err := r.q.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}
		return domain.Product{}, fmt.Errorf("%s: %w", op, err)
	}
	return toDomainProduct(row), nil
}

func (r *CatalogRepository) ListProducts(
	ctx context.Context,
	filter domain.ProductFilter,
	cursor *domain.ProductCursor,
	limit int32,
) ([]domain.Product, error) {
	const op = "repository.ListProducts"

	params := dbcatalog.ListProductsParams{
		CategoryID: filter.CategoryID,
		SellerID:   filter.SellerID,
		Limit:      limit,
	}
	if cursor != nil {
		params.CursorCreatedAt = &cursor.CreatedAt
		params.CursorID = &cursor.ID
	}

	rows, err := r.q.ListProducts(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		result = append(result, toDomainProduct(row))
	}
	return result, nil
}

func toDomainProduct(row dbcatalog.Product) domain.Product {
	return domain.Product{
		ID:          row.ID,
		SellerID:    row.SellerID,
		CategoryID:  row.CategoryID,
		Title:       row.Title,
		Description: row.Description,
		Price: domain.Money{
			Amount:   row.PriceAmount,
			Currency: row.PriceCurrency,
		},
		Status:    domain.ProductStatus(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
