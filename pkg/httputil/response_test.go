package httputil

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/autopus/bootstrap/pkg/errors"
)

type mockResponseWriter struct {
	header     http.Header
	body       bytes.Buffer
	statusCode int
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{header: make(http.Header)}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write(bytes []byte) (int, error) {
	return m.body.Write(bytes)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestJSON(t *testing.T) {
	mrw := newMockResponseWriter()
	resp := New(mrw)

	body := map[string]string{"hello": "world"}
	resp.JSON(body, http.StatusOK)

	assert.Equal(t, http.StatusOK, mrw.statusCode)
	assert.Contains(t, mrw.body.String(), "\"hello\":\"world\"")
	assert.Equal(t, "application/json", mrw.header.Get("Content-Type"))
}

// Tests for other methods can follow a similar pattern.
// For example, testing the Response method:
func TestResponse(t *testing.T) {
	mrw := newMockResponseWriter()
	resp := New(mrw)

	resp.Response([]byte("response body"), http.StatusAccepted)

	assert.Equal(t, http.StatusAccepted, mrw.statusCode)
	assert.Equal(t, "response body", mrw.body.String())
}

func TestBadRequest(t *testing.T) {
	mrw := newMockResponseWriter()
	resp := New(mrw)

	err := errors.New("invalid input")
	resp.BadRequest(context.Background(), err)

	assert.Equal(t, http.StatusBadRequest, mrw.statusCode)
	assert.Contains(t, mrw.body.String(), "invalid input")
	assert.Equal(t, "application/json", mrw.header.Get("Content-Type"))
}

func TestNotFound(t *testing.T) {
	mrw := newMockResponseWriter()
	resp := New(mrw)

	err := errors.New("not found error")
	resp.NotFound(context.Background(), err)

	assert.Equal(t, http.StatusNotFound, mrw.statusCode)
	assert.Contains(t, mrw.body.String(), "not found error")
	assert.Equal(t, "application/json", mrw.header.Get("Content-Type"))
}

func TestError(t *testing.T) {
	mrw := newMockResponseWriter()
	resp := New(mrw)

	err := errors.New("internal error")
	resp.Error(context.Background(), err)

	assert.Equal(t, http.StatusInternalServerError, mrw.statusCode)
	assert.Equal(t, "{\"error\":\"internal server error\"}\n", mrw.body.String())
	assert.Equal(t, "application/json", mrw.header.Get("Content-Type"))
}

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{"No Error", nil, http.StatusOK, "{}\n"},
		{"Validation Error", errors.NewValidationError(errors.New("validation error")), http.StatusBadRequest, "{\"error\":\"validation error\"}\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := context.Background()

			HandleResponse(ctx, w, struct{}{}, tt.err)

			res := w.Result()
			defer res.Body.Close()

			body, _ := io.ReadAll(res.Body)
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			assert.Equal(t, tt.expectedBody, string(body))
		})
	}
}
