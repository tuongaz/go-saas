package httputil

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tuongaz/go-saas/pkg/errors/apierror"
	"github.com/tuongaz/go-saas/pkg/log"
)

type Response struct {
	w http.ResponseWriter
}

func (r *Response) Error(ctx context.Context, err error) {
	log.Default().ErrorContext(ctx, "internal server error", log.ErrorAttr(err))
	r.JSON(map[string]string{"error": "internal server error"}, http.StatusInternalServerError)
}

func (r *Response) Response(body []byte, statuses ...int) {
	if len(statuses) > 0 {
		r.w.WriteHeader(statuses[0])
	} else {
		r.w.WriteHeader(http.StatusOK)
	}
	if body != nil {
		_, _ = r.w.Write(body)
	}
}

func (r *Response) Unauthorized(ctx context.Context, err error) {
	r.JSON(map[string]string{"error": err.Error()}, http.StatusUnauthorized)
}

func (r *Response) Forbidden(ctx context.Context, err error) {
	r.JSON(map[string]string{"error": err.Error()}, http.StatusForbidden)
}

func (r *Response) NotFound(ctx context.Context, err error) {
	r.JSON(map[string]string{"error": err.Error()}, http.StatusNotFound)
}

func (r *Response) BadRequest(ctx context.Context, err error) {
	r.JSON(map[string]string{"error": err.Error()}, http.StatusBadRequest)
}

func (r *Response) JSON(body any, status ...int) {
	if body != nil {
		r.w.Header().Set("Content-Type", "application/json")
	}
	if len(status) > 0 {
		r.w.WriteHeader(status[0])
	} else {
		r.w.WriteHeader(http.StatusOK)
	}

	if body != nil {
		switch b := body.(type) {
		case []byte:
			_, _ = r.w.Write(b)
		case string:
			_, _ = r.w.Write([]byte(b))
		default:
			_ = json.NewEncoder(r.w).Encode(body)
		}
	}
}

func HandleResponse(ctx context.Context, w http.ResponseWriter, out any, err error, status ...int) {
	responseStatus := http.StatusOK
	if len(status) > 0 {
		responseStatus = status[0]
	}

	response := New(w)
	if err != nil {
		log.Default().ErrorContext(ctx, "request error", log.ErrorAttr(err))
		if apierror.IsValidation(err) {
			response.BadRequest(ctx, err)
			return
		}

		if apierror.IsNotFound1(err) {
			response.NotFound(ctx, err)
			return
		}

		if apierror.IsForbidden(err) {
			response.Forbidden(ctx, err)
			return
		}

		if apierror.IsUnauthorized(err) {
			response.Unauthorized(ctx, err)
			return
		}

		response.Error(ctx, err)
		return
	}

	response.JSON(out, responseStatus)
}

func New(w http.ResponseWriter) *Response {
	return &Response{w: w}
}
