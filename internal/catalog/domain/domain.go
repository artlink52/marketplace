package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrParentNotFound        = errors.New("parent category not found")
	ErrInvalidCategoryName   = errors.New("invalid category name")
	ErrInvalidCategorySlug   = errors.New("invalid category slug")
	ErrNoPrivacy             = errors.New("no privacy")
)

type Category struct {
	ID       uuid.UUID
	Name     string
	Slug     string
	ParentID *uuid.UUID
}
