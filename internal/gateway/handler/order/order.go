package order

import (
	"encoding/json"
	"net/http"

	orderclient "github.com/artlink52/marketplace/internal/gateway/client/order"
	"github.com/artlink52/marketplace/internal/gateway/middleware"
	"github.com/artlink52/marketplace/internal/gateway/render"
	"github.com/google/uuid"
)

type CreateOrderBody struct {
	Items []Item `json:"items"`
}

type Item struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	res := render.New(w)

	var req CreateOrderBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		res.Error(http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		res.Error(http.StatusInternalServerError, "failed to parse user id")
		return
	}

	idempotencyKey := r.Header.Get("X-Idempotency-Key")

	items := make([]orderclient.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = orderclient.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	resp, err := h.client.CreateOrder(ctx, orderclient.CreateOrderInput{
		UserID:         userID,
		IdempotencyKey: idempotencyKey,
		Items:          items,
	})
	if err != nil {
		res.GRPCError(err)
		return
	}

	res.JSON(http.StatusCreated, resp)
}
