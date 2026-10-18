package catalog

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/artlink52/marketplace/internal/gateway/client/catalog"
	"github.com/artlink52/marketplace/internal/gateway/middleware"
	"github.com/artlink52/marketplace/internal/gateway/render"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createProductResponse struct {
	ID uuid.UUID `json:"id"`
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	res := render.New(w)

	var req catalog.CreateProductInput

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		res.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	productID, err := h.categoryClient.CreateProduct(ctx, req, claims.UserID)
	if err != nil {
		res.GRPCError(err)
		return
	}

	res.JSON(http.StatusCreated, createProductResponse{ID: productID})
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	res := render.New(w)

	product, err := h.categoryClient.GetProduct(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		res.GRPCError(err)
		return
	}

	// Ручка публичная: черновик видит только продавец, остальным — как будто товара нет.
	// Архивные отдаём: по ним можно показать «снят с продажи».
	if product.Status == "draft" {
		res.Error(http.StatusNotFound, "product not found")
		return
	}

	res.JSON(http.StatusOK, product)
}

// ListProducts: GET /products?category_id=&seller_id=&page_size=&page_token=
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	res := render.New(w)
	q := r.URL.Query()

	var pageSize int32
	if v := q.Get("page_size"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			res.Error(http.StatusBadRequest, "page_size must be a number")
			return
		}
		pageSize = int32(n)
	}

	page, err := h.categoryClient.ListProducts(r.Context(), catalog.ListProductsInput{
		CategoryID: q.Get("category_id"),
		SellerID:   q.Get("seller_id"),
		PageSize:   pageSize,
		PageToken:  q.Get("page_token"),
	})
	if err != nil {
		res.GRPCError(err)
		return
	}

	res.JSON(http.StatusOK, page)
}
