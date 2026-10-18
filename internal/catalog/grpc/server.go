package grpc

import (
	"context"
	"errors"

	catalogv1 "github.com/artlink52/marketplace/gen/go/catalog/v1"
	commonv1 "github.com/artlink52/marketplace/gen/go/common/v1"
	"github.com/artlink52/marketplace/internal/catalog/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CatalogService interface {
	CreateCategory(ctx context.Context, c domain.Category) (domain.Category, error)
	ListCategories(ctx context.Context) ([]domain.Category, error)
	CreateProduct(ctx context.Context, p domain.Product) (uuid.UUID, error)
	GetProduct(ctx context.Context, id uuid.UUID) (domain.Product, error)
	ListProducts(ctx context.Context, filter domain.ProductFilter, pageSize int32, pageToken string) ([]domain.Product, string, error)
	BatchGetProducts(ctx context.Context, ids []uuid.UUID) ([]domain.Product, error)
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

func (s *CatalogServer) ListCategories(ctx context.Context, _ *catalogv1.ListCategoriesRequest) (*catalogv1.ListCategoriesResponse, error) {
	categories, err := s.catalogService.ListCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list categories")
	}

	resp := &catalogv1.ListCategoriesResponse{}
	for _, c := range categories {
		resp.Categories = append(resp.Categories, toProtoCategory(c))
	}
	return resp, nil
}

func (s *CatalogServer) CreateProduct(ctx context.Context, req *catalogv1.CreateProductRequest) (*catalogv1.CreateProductResponse, error) {

	sellerID, err := uuid.Parse(req.GetSellerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid seller_id")
	}

	categoryID, err := uuid.Parse(req.GetCategoryId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid category_id")
	}

	productID, err := s.catalogService.CreateProduct(ctx, domain.Product{
		SellerID:    sellerID,
		CategoryID:  categoryID,
		Title:       req.GetTitle(),
		Description: req.Description,
		Price: domain.Money{
			Amount:   req.GetPrice().GetAmount(),
			Currency: req.GetPrice().GetCurrency(),
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidProductTitle):
			return nil, status.Error(codes.InvalidArgument, "max title length is 255 symbols and not empty")
		case errors.Is(err, domain.ErrInvalidProductDescription):
			return nil, status.Error(codes.InvalidArgument, "max description length is 500 symbols")
		case errors.Is(err, domain.ErrInvalidProductPrice):
			return nil, status.Error(codes.InvalidArgument, "price must be >0")
		case errors.Is(err, domain.ErrCategoryNotFound):
			return nil, status.Error(codes.InvalidArgument, "category not found")
		default:
			return nil, status.Error(codes.Internal, "failed to create product")
		}
	}

	return &catalogv1.CreateProductResponse{
		ProductId: productID.String(),
	}, nil
}

func (s *CatalogServer) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	id, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	product, err := s.catalogService.GetProduct(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, "product not found")
		default:
			return nil, status.Error(codes.Internal, "failed to get product")
		}
	}

	return &catalogv1.GetProductResponse{Product: toProtoProduct(product)}, nil
}

func (s *CatalogServer) ListProducts(ctx context.Context, req *catalogv1.ListProductsRequest) (*catalogv1.ListProductsResponse, error) {
	categoryID, err := parseOptionalUUID(req.GetCategoryId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid category_id")
	}

	sellerID, err := parseOptionalUUID(req.GetSellerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid seller_id")
	}

	products, nextPageToken, err := s.catalogService.ListProducts(
		ctx,
		domain.ProductFilter{CategoryID: categoryID, SellerID: sellerID},
		req.GetPage().GetPageSize(),
		req.GetPage().GetPageToken(),
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidPageToken):
			return nil, status.Error(codes.InvalidArgument, "invalid page_token")
		default:
			return nil, status.Error(codes.Internal, "failed to list products")
		}
	}

	resp := &catalogv1.ListProductsResponse{
		Products: make([]*catalogv1.Product, 0, len(products)),
		Page:     &commonv1.PageResponse{NextPageToken: nextPageToken},
	}
	for _, p := range products {
		resp.Products = append(resp.Products, toProtoProduct(p))
	}
	return resp, nil
}

func (s *CatalogServer) BatchGetProducts(ctx context.Context, req *catalogv1.BatchGetProductsRequest) (*catalogv1.BatchGetProductsResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.GetProductIds()))
	for _, raw := range req.GetProductIds() {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid product_id: %q", raw)
		}
		ids = append(ids, id)
	}

	products, err := s.catalogService.BatchGetProducts(ctx, ids)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTooManyIDs):
			return nil, status.Error(codes.InvalidArgument, "too many product_ids")
		default:
			return nil, status.Error(codes.Internal, "failed to get products")
		}
	}

	resp := &catalogv1.BatchGetProductsResponse{
		Products: make([]*catalogv1.Product, 0, len(products)),
	}
	for _, p := range products {
		resp.Products = append(resp.Products, toProtoProduct(p))
	}
	return resp, nil
}

// parseOptionalUUID: пустая строка — фильтр не задан (nil), иначе строка должна быть валидным UUID.
func parseOptionalUUID(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func toProtoProduct(p domain.Product) *catalogv1.Product {
	return &catalogv1.Product{
		Id:          p.ID.String(),
		SellerId:    p.SellerID.String(),
		Title:       p.Title,
		Description: p.Description,
		CategoryId:  p.CategoryID.String(),
		Price: &commonv1.Money{
			Amount:   p.Price.Amount,
			Currency: p.Price.Currency,
		},
		Status:    toProtoProductStatus(p.Status),
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}
}

func toProtoProductStatus(st domain.ProductStatus) catalogv1.ProductStatus {
	switch st {
	case domain.StatusDraft:
		return catalogv1.ProductStatus_PRODUCT_STATUS_DRAFT
	case domain.StatusActive:
		return catalogv1.ProductStatus_PRODUCT_STATUS_ACTIVE
	case domain.StatusArchived:
		return catalogv1.ProductStatus_PRODUCT_STATUS_ARCHIVED
	default:
		return catalogv1.ProductStatus_PRODUCT_STATUS_UNSPECIFIED
	}
}

func toProtoCategory(c domain.Category) *catalogv1.Category {
	cat := &catalogv1.Category{
		Id:   c.ID.String(),
		Name: c.Name,
		Slug: c.Slug,
	}
	if c.ParentID != nil {
		s := c.ParentID.String()
		cat.ParentId = new(s)
	}
	for _, sub := range c.Subcategories {
		cat.Subcategories = append(cat.Subcategories, toProtoCategory(sub))
	}
	return cat
}
