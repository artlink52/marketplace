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
