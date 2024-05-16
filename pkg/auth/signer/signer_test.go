package signer

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tuongaz/go-saas/service/auth/model"
)

func TestSignSuccess(t *testing.T) {
	signer := NewHS256Signer([]byte("secret"))
	claims := model.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Hour)},
		},
		// Add additional fields to CustomClaims if needed
	}

	tokenString, err := signer.SignCustomClaims(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
}

func TestSignError(t *testing.T) {
	signer := NewHS256Signer(nil) // Passing nil to simulate an error
	claims := model.CustomClaims{}

	_, err := signer.SignCustomClaims(claims)
	assert.Error(t, err)
}

func TestParseSuccess(t *testing.T) {
	secretKey := []byte("secret")
	signer := NewHS256Signer(secretKey)
	claims := model.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Hour)},
		},
	}

	tokenString, _ := signer.SignCustomClaims(claims)

	parsedClaims, err := signer.ParseCustomClaims(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, parsedClaims)
}

func TestParseInvalidToken(t *testing.T) {
	signer := NewHS256Signer([]byte("secret"))

	_, err := signer.ParseCustomClaims("invalidToken")
	assert.Error(t, err)
}
