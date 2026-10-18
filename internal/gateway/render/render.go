package render

import (
	"encoding/json"
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Response struct {
	rw http.ResponseWriter
}

func New(rw http.ResponseWriter) *Response {
	return &Response{rw: rw}
}

func (r *Response) JSON(code int, v any) {
	r.rw.Header().Set("Content-Type", "application/json")
	r.rw.WriteHeader(code)
	if err := json.NewEncoder(r.rw).Encode(v); err != nil {
		panic(err)
	}
}

func (r *Response) NoContent() {
	r.rw.WriteHeader(http.StatusNoContent)
}

func (r *Response) GRPCError(err error) {
	// Клиенты оборачивают ошибки (fmt.Errorf("...: %w")), а status.Convert для
	// обёрнутой ошибки кладёт в message весь текст обёртки. Достаём исходный статус.
	st := status.New(codes.Unknown, err.Error())
	var gs interface{ GRPCStatus() *status.Status }
	if errors.As(err, &gs) {
		st = gs.GRPCStatus()
	}

	var code int

	switch st.Code() {
	case codes.NotFound:
		code = http.StatusNotFound
	case codes.AlreadyExists:
		code = http.StatusConflict
	case codes.InvalidArgument:
		code = http.StatusBadRequest
	case codes.Unauthenticated:
		code = http.StatusUnauthorized
	case codes.PermissionDenied:
		code = http.StatusForbidden
	case codes.Unavailable:
		code = http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		code = http.StatusGatewayTimeout
	default:
		code = http.StatusInternalServerError
	}

	// Текст внутренних ошибок наружу не отдаём: там могут быть детали БД и сети.
	if code >= http.StatusInternalServerError {
		r.Error(code, http.StatusText(code))
		return
	}

	r.Error(code, st.Message())
}

func (r *Response) Error(code int, msg string) {
	r.JSON(code, struct {
		Error string `json:"error"`
	}{Error: msg})
}
