package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tuongaz/go-saas/pkg/apierror"
	"github.com/tuongaz/go-saas/pkg/log"
)

type Response struct {
	w http.ResponseWriter
}

func (r *Response) Error(ctx context.Context, err error) {
	log.Default().ErrorContext(ctx, "internal server error", log.ErrorAttr(err))
	r.JSON(map[string]string{"message": "internal server error"}, http.StatusInternalServerError)
}

func (r *Response) JSON(body any, status ...int) {
	if body == nil {
		body = map[string]string{}
	}
	r.w.Header().Set("Content-Type", "application/json")
	if len(status) > 0 {
		r.w.WriteHeader(status[0])
	} else {
		r.w.WriteHeader(http.StatusOK)
	}

	if body != nil {
		switch b := body.(type) {
		case []byte:
			_, _ = r.w.Write(b)
			break
		case string:
			_, _ = r.w.Write([]byte(b))
			break
		case *string:
			_, _ = r.w.Write([]byte(*b))
			break
		default:
			data, err := json.Marshal(body)
			if err != nil {
				r.Error(context.Background(), err)
				return
			}
			_, _ = r.w.Write(data)
		}
	}
}

func (r *Response) NoContent() {
	r.w.WriteHeader(http.StatusNoContent)
}

func HandleResponse(ctx context.Context, w http.ResponseWriter, out any, err error, status ...int) {
	responseStatus := http.StatusOK
	if len(status) > 0 {
		responseStatus = status[0]
	}

	response := New(w)
	if err != nil {
		log.Default().ErrorContext(ctx, err.Error(), log.ErrorAttr(err))

		var unwrappedErr *apierror.APIError
		if errors.As(err, &unwrappedErr) {
			response.JSON(unwrappedErr, unwrappedErr.Code)
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
