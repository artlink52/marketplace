package render

import (
	"encoding/json"
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
	st := status.Convert(err)

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
	default:
		code = http.StatusInternalServerError
	}

	r.Error(code, st.Message())
}

func (r *Response) Error(code int, msg string) {
	r.JSON(code, struct {
		Error string `json:"error"`
	}{Error: msg})
}
