package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrParentNotFound        = errors.New("parent category not found")
	ErrInvalidCategoryName   = errors.New("invalid category name")
	ErrInvalidCategorySlug   = errors.New("invalid category slug")

	ErrProductAlreadyExists      = errors.New("product already exists")
	ErrInvalidProductTitle       = errors.New("invalid product title")
	ErrInvalidProductDescription = errors.New("invalid product description")
	ErrInvalidProductPrice       = errors.New("invalid product price")
	ErrProductNotFound           = errors.New("product not found")

	ErrInvalidPageToken = errors.New("invalid page token")
	ErrTooManyIDs       = errors.New("too many ids")
)

type ProductStatus string

const (
	StatusDraft    ProductStatus = "draft"
	StatusActive   ProductStatus = "active"
	StatusArchived ProductStatus = "archived"
)

type Category struct {
	ID            uuid.UUID
	Name          string
	Slug          string
	ParentID      *uuid.UUID
	Subcategories []Category
}

type Money struct {
	Amount   int64
	Currency string
}

type Product struct {
	ID          uuid.UUID
	SellerID    uuid.UUID
	CategoryID  uuid.UUID
	Title       string
	Description *string
	Price       Money
	Status      ProductStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductFilter — необязательные фильтры витрины; nil означает «не фильтровать».
type ProductFilter struct {
	CategoryID *uuid.UUID
	SellerID   *uuid.UUID
}

// ProductCursor — позиция последнего товара на странице для keyset-пагинации.
type ProductCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}
