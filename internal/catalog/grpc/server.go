package grpc

import (
	"context"
	"errors"

	catalogv1 "github.com/artlink52/marketplace/gen/go/catalog/v1"
	"github.com/artlink52/marketplace/internal/catalog/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogService interface {
	CreateCategory(ctx context.Context, c domain.Category) (domain.Category, error)
}

type CatalogServer struct {
	catalogv1.UnimplementedCatalogServiceServer
	catalogService CatalogService
}

func RegisterCatalogServer(server *grpc.Server, catalogService CatalogService) {
	catalogv1.RegisterCatalogServiceServer(server, &CatalogServer{catalogService: catalogService})
}

func (s *CatalogServer) CreateCategory(
	ctx context.Context,
	req *catalogv1.CreateCategoryRequest,
) (*catalogv1.CreateCategoryResponse, error) {
	var parentID *uuid.UUID

	if req.ParentId != nil {
		id, err := uuid.Parse(req.GetParentId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid parent_id")
		}
		parentID = &id
	}

	category, err := s.catalogService.CreateCategory(ctx, domain.Category{
		Name:     req.GetName(),
		Slug:     req.GetSlug(),
		ParentID: parentID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCategoryName):
			return nil, status.Error(codes.InvalidArgument, "invalid category name")
		case errors.Is(err, domain.ErrInvalidCategorySlug):
			return nil, status.Error(codes.InvalidArgument, "slug must contain only a-z, 0-9 and single hyphens")
		case errors.Is(err, domain.ErrParentNotFound):
			return nil, status.Error(codes.InvalidArgument, "parent category not found")
		case errors.Is(err, domain.ErrCategoryAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "category with this slug already exists")
		default:
			return nil, status.Error(codes.Internal, "failed to create category")
		}
	}

	return &catalogv1.CreateCategoryResponse{
		Category: toProtoCategory(category),
	}, nil
}

func toProtoCategory(c domain.Category) *catalogv1.Category {
	cat := &catalogv1.Category{
		Id:   c.ID.String(),
		Name: c.Name,
		Slug: c.Slug,
	}
	if c.ParentID != nil {
		s := c.ParentID.String()
		cat.ParentId = &s
	}
	return cat
}
