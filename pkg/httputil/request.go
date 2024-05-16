package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tuongaz/go-saas/pkg/apierror"
	"github.com/tuongaz/go-saas/pkg/log"
)

func ParseRequestBody[T any](r *http.Request) (*T, error) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Default().ErrorContext(r.Context(), "failed to close request body", log.ErrorAttr(err))
		}
	}()

	if r.Body == nil {
		return nil, nil
	}
	target := new(T)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}

	if len(body) == 0 {
		return target, nil
	}

	if err := json.Unmarshal(body, target); err != nil {
		return nil, apierror.NewValidationError("Invalid data structure", fmt.Errorf("unmarshal request body: %w", err))
	}

	return target, nil
}
