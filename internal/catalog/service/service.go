package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/artlink52/marketplace/internal/catalog/domain"
)

type CatalogRepository interface {
	CreateCategory(ctx context.Context, c domain.Category) (domain.Category, error)
}
type CatalogService struct {
	CatalogRepository CatalogRepository
}

func New(catalogRepository CatalogRepository) *CatalogService {
	return &CatalogService{
		CatalogRepository: catalogRepository,
	}
}

const (
	maxCategoryNameLen = 100
	maxCategorySlugLen = 64
)

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

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
