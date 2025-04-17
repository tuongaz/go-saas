package signer

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tuongaz/go-saas/core/auth/model"
)

var _ Interface = (*SecretKeySigner)(nil)

type Interface interface {
	SignCustomClaims(claims model.CustomClaims) (string, error)
	ParseCustomClaims(tokenString string) (*model.CustomClaims, error)
	SignRegisteredClaims(claims jwt.RegisteredClaims) (string, error)
	ParseRegisteredClaims(tokenString string) (*jwt.RegisteredClaims, error)
}

type SecretKeySigner struct {
	secretKey []byte
}

func (h SecretKeySigner) SignRegisteredClaims(claims jwt.RegisteredClaims) (string, error) {
	if h.secretKey == nil {
		return "", fmt.Errorf("secret key is nil")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(h.secretKey)
	if err != nil {
		return "", fmt.Errorf("HS512 sign token: %w", err)
	}

	return tokenString, nil
}

func (h SecretKeySigner) SignCustomClaims(claims model.CustomClaims) (string, error) {
	if h.secretKey == nil {
		return "", fmt.Errorf("secret key is nil")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(h.secretKey)
	if err != nil {
		return "", fmt.Errorf("HS512 sign token: %w", err)
	}

	return tokenString, nil
}

func (h SecretKeySigner) ParseRegisteredClaims(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return h.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func (h SecretKeySigner) ParseCustomClaims(tokenString string) (*model.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return h.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse custom claims: %w", err)
	}

	if claims, ok := token.Claims.(*model.CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func NewHS512Signer(secretKey []byte) *SecretKeySigner {
	return &SecretKeySigner{
		secretKey: secretKey,
	}
}
