package catalog

import (
	"context"
	"fmt"

	catalogv1 "github.com/artlink52/marketplace/gen/go/catalog/v1"
)

type Category struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	ParentID      *string    `json:"parent_id"`
	Subcategories []Category `json:"subcategories,omitempty"`
}

type CreateCategoryInput struct {
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ParentID *string `json:"parent_id"`
}

func (c *Client) CreateCategory(ctx context.Context, input CreateCategoryInput) (*Category, error) {
	resp, err := c.catalog.CreateCategory(ctx, &catalogv1.CreateCategoryRequest{
		Name:     input.Name,
		Slug:     input.Slug,
		ParentId: input.ParentID,
	})
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	return &Category{
		ID:       resp.Category.Id,
		Name:     resp.Category.Name,
		Slug:     resp.Category.Slug,
		ParentID: resp.Category.ParentId,
	}, nil
}

func (c *Client) ListCategories(ctx context.Context) ([]Category, error) {
	resp, err := c.catalog.ListCategories(ctx, &catalogv1.ListCategoriesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	categories := make([]Category, 0, len(resp.GetCategories()))
	for _, pc := range resp.GetCategories() {
		categories = append(categories, fromProtoCategory(pc))
	}
	return categories, nil
}

func fromProtoCategory(c *catalogv1.Category) Category {
	cat := Category{
		ID:       c.GetId(),
		Name:     c.GetName(),
		Slug:     c.GetSlug(),
		ParentID: c.ParentId,
	}
	for _, sub := range c.GetSubcategories() {
		cat.Subcategories = append(cat.Subcategories, fromProtoCategory(sub))
	}
	return cat
}
