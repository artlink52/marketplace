package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/artlink52/marketplace/internal/gateway/client/catalog"
	"github.com/artlink52/marketplace/internal/gateway/render"
)

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	res := render.New(w)

	var req catalog.CreateCategoryInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.categoryClient.CreateCategory(ctx, req)
	if err != nil {
		res.GRPCError(err)
		return
	}

	res.JSON(http.StatusCreated, resp)
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	res := render.New(w)

	resp, err := h.categoryClient.ListCategories(ctx)
	if err != nil {
		res.GRPCError(err)
		return
	}

	res.JSON(http.StatusOK, resp)
}
