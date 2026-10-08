package user

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	userclient "github.com/artlink52/marketplace/internal/gateway/client/user"
	"github.com/artlink52/marketplace/internal/gateway/lib"
	"github.com/artlink52/marketplace/internal/gateway/render"
	"github.com/go-chi/chi/v5"
)

type Client interface {
	Register(ctx context.Context, input userclient.RegisterInput) (string, error)
	Authenticate(ctx context.Context, input userclient.AuthenticateInput) (string, userclient.Role, error)
	GetUser(ctx context.Context, userID string) (userclient.User, error)
}

type Handler struct {
	client    Client
	jwtSecret string
	jwtTTL    time.Duration
}

func New(client Client, jwtSecret string, jwtTTL time.Duration) *Handler {
	return &Handler{
		client:    client,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
	}
}

type registerResponse struct {
	UserID string `json:"user_id"`
}

type loginResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	res := render.New(w)

	var req userclient.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	userID, err := h.client.Register(r.Context(), req)
	if err != nil {
		res.GRPCError(err)
		return
	}

	res.JSON(http.StatusCreated, registerResponse{UserID: userID})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	res := render.New(w)

	var req userclient.AuthenticateInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Error(http.StatusBadRequest, "invalid request body")
		return
	}

	userID, role, err := h.client.Authenticate(r.Context(), req)
	if err != nil {
		res.GRPCError(err)
		return
	}

	token, err := lib.NewToken(lib.Claims{UserID: userID, Role: string(role)}, h.jwtSecret, h.jwtTTL)
	if err != nil {
		res.Error(http.StatusInternalServerError, "failed to generate token")
		return
	}

	res.JSON(http.StatusOK, loginResponse{UserID: userID, Token: token})
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	res := render.New(w)

	userID := chi.URLParam(r, "id")

	user, err := h.client.GetUser(r.Context(), userID)
	if err != nil {
		res.GRPCError(err)
		return
	}

	res.JSON(http.StatusOK, user)
}
