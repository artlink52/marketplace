package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/artlink52/marketplace/internal/catalog/domain"
	"github.com/google/uuid"
)

const (
	maxCategoryNameLen = 100
	maxCategorySlugLen = 64

	maxProductTitleLen       = 255
	maxProductDescriptionLen = 5000

	maxBatchSize = 100
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type CatalogRepository interface {
	CreateCategory(ctx context.Context, c domain.Category) (domain.Category, error)
	ListCategories(ctx context.Context) ([]domain.Category, error)
	CreateProduct(ctx context.Context, p domain.Product) (uuid.UUID, error)
	GetProduct(ctx context.Context, id uuid.UUID) (domain.Product, error)
	ListProducts(ctx context.Context, filter domain.ProductFilter, cursor *domain.ProductCursor, limit int32) ([]domain.Product, error)
	BatchGetProducts(ctx context.Context, ids []uuid.UUID) ([]domain.Product, error)
}

type CatalogService struct {
	CatalogRepository CatalogRepository
}

func New(catalogRepository CatalogRepository) *CatalogService {
	return &CatalogService{
		CatalogRepository: catalogRepository,
	}
}

func (s *CatalogService) CreateCategory(ctx context.Context, c domain.Category) (domain.Category, error) {
	const op = "service.CreateCategory"

	c.Name = strings.TrimSpace(c.Name)
	c.Slug = strings.ToLower(strings.TrimSpace(c.Slug))

	if c.Name == "" || utf8.RuneCountInString(c.Name) > maxCategoryNameLen {
		return domain.Category{}, domain.ErrInvalidCategoryName
	}

	if len(c.Slug) > maxCategorySlugLen || !slugRe.MatchString(c.Slug) {
		return domain.Category{}, domain.ErrInvalidCategorySlug
	}

	category, err := s.CatalogRepository.CreateCategory(ctx, c)
	if err != nil {
		return domain.Category{}, fmt.Errorf("%s: %w", op, err)
	}

	return category, nil
}

func (s *CatalogService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	const op = "service.ListCategories"
	list, err := s.CatalogRepository.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return buildTree(list), nil
}

func buildTree(flat []domain.Category) []domain.Category {
	children := make(map[uuid.UUID][]domain.Category)
	var roots []domain.Category

	for _, c := range flat {
		if c.ParentID == nil {
			roots = append(roots, c)
			continue
		}
		children[*c.ParentID] = append(children[*c.ParentID], c)
	}

	var attach func(nodes []domain.Category) []domain.Category
	attach = func(nodes []domain.Category) []domain.Category {
		for i := range nodes {
			nodes[i].Subcategories = attach(children[nodes[i].ID])
		}
		return nodes
	}

	return attach(roots)
}

func (s *CatalogService) CreateProduct(ctx context.Context, p domain.Product) (uuid.UUID, error) {
	const op = "service.CreateProduct"

	p.Title = strings.TrimSpace(p.Title)
	if p.Description != nil {
		trimmed := strings.TrimSpace(*p.Description)
		if trimmed == "" {
			p.Description = nil
		} else {
			if utf8.RuneCountInString(trimmed) > maxProductDescriptionLen {
				return uuid.Nil, domain.ErrInvalidProductDescription
			}
			p.Description = &trimmed
		}
	}

	if p.Title == "" || utf8.RuneCountInString(p.Title) > maxProductTitleLen {
		return uuid.Nil, domain.ErrInvalidProductTitle
	}

	if p.Price.Amount < 0 || utf8.RuneCountInString(p.Price.Currency) > 3 {
		return uuid.Nil, domain.ErrInvalidProductPrice
	}

	productID, err := s.CatalogRepository.CreateProduct(ctx, p)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return productID, nil
}

func (s *CatalogService) GetProduct(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	const op = "service.GetProduct"
	product, err := s.CatalogRepository.GetProduct(ctx, id)
	if err != nil {
		return domain.Product{}, fmt.Errorf("%s: %w", op, err)
	}
	return product, nil
}

// ListProducts возвращает страницу активных товаров и токен следующей страницы
// (пустой, если страница последняя).
func (s *CatalogService) ListProducts(
	ctx context.Context,
	filter domain.ProductFilter,
	pageSize int32,
	pageToken string,
) ([]domain.Product, string, error) {
	const op = "service.ListProducts"

	pageSize = normalizePageSize(pageSize)

	var cursor *domain.ProductCursor
	if pageToken != "" {
		c, err := decodeCursor(pageToken)
		if err != nil {
			return nil, "", domain.ErrInvalidPageToken
		}
		cursor = &c
	}

	// Берём на один товар больше: если он пришёл — значит, есть следующая страница.
	products, err := s.CatalogRepository.ListProducts(ctx, filter, cursor, pageSize+1)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	var nextPageToken string
	if len(products) > int(pageSize) {
		products = products[:pageSize]
		last := products[len(products)-1]
		nextPageToken = encodeCursor(domain.ProductCursor{CreatedAt: last.CreatedAt, ID: last.ID})
	}

	return products, nextPageToken, nil
}

// BatchGetProducts возвращает найденные товары в любом статусе. Ненайденные id
// просто отсутствуют в ответе — вызывающий сервис сам решает, что с этим делать.
func (s *CatalogService) BatchGetProducts(ctx context.Context, ids []uuid.UUID) ([]domain.Product, error) {
	const op = "service.BatchGetProducts"

	if len(ids) == 0 {
		return []domain.Product{}, nil
	}
	if len(ids) > maxBatchSize {
		return nil, domain.ErrTooManyIDs
	}

	products, err := s.CatalogRepository.BatchGetProducts(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return products, nil
}
