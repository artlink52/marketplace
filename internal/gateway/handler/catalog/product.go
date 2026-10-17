package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/artlink52/marketplace/internal/gateway/client/catalog"
	"github.com/artlink52/marketplace/internal/gateway/middleware"
	"github.com/artlink52/marketplace/internal/gateway/render"
)

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

	res.JSON(http.StatusCreated, productID)
}
