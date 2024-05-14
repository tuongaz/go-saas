package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tuongaz/go-saas/pkg/auth/signer"
)

func TestMiddlewareWithInvalidToken(t *testing.T) {
	secretKey := []byte("secret")
	sign := signer.NewHS256Signer(secretKey)
	service := &Service{signer: sign}

	middleware := service.NewMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := PrincipalFromCtx(r.Context())
		assert.Equal(t, "org123", principal.OrganisationID)
		assert.Equal(t, "sub123", principal.AccountID)
		assert.Equal(t, "type123", principal.AccountType)
		assert.Equal(t, "role123", principal.Role)
	})

	testServer := httptest.NewServer(middleware(handler))
	defer testServer.Close()

	client := testServer.Client()
	req, _ := http.NewRequest("GET", testServer.URL, nil)
	req.Header.Add("Authorization", "Bearer "+"hello")

	resp, _ := client.Do(req)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
