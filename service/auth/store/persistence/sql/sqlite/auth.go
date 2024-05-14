package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/tuongaz/go-saas/service/auth/store/persistence"
)

func (s *SQL) CreateAuthToken(ctx context.Context, row persistence.AuthTokenRow) (sql.Result, error) {
	result, err := s.namedExecContext(
		ctx,
		"INSERT INTO auth_token (id, account_role_id, refresh_token, created_at, updated_at) VALUES (:id, :account_role_id, :refresh_token, :created_at, :updated_at)",
		row,
	)
	if err != nil {
		return nil, fmt.Errorf("create auth token: %w", err)
	}

	return result, nil
}

func (s *SQL) UpdateAuthToken(ctx context.Context, id string, input persistence.UpdateAuthTokenInput) (sql.Result, error) {
	result, err := s.namedExecContext(
		ctx,
		"UPDATE auth_token SET refresh_token = :refresh_token, updated_at = :updated_at WHERE id = :id",
		map[string]interface{}{
			"id":            id,
			"refresh_token": input.RefreshToken,
			"updated_at":    input.UpdatedAt,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("update auth token: %w", err)
	}

	return result, nil
}

func (s *SQL) GetAuthTokenByAccountRoleID(ctx context.Context, accountRoleID string) (*persistence.AuthTokenRow, error) {
	var row persistence.AuthTokenRow
	err := s.getContext(
		ctx,
		&row,
		"SELECT * FROM auth_token WHERE account_role_id = ?",
		accountRoleID,
	)
	if err != nil {
		return nil, fmt.Errorf("get auth token by account role id: %w", err)
	}

	return &row, nil
}

func (s *SQL) GetUserByEmail(ctx context.Context, email string) (*persistence.UserRow, error) {
	var row persistence.UserRow
	err := s.getContext(
		ctx,
		&row,
		"SELECT * FROM auth_user WHERE email = ?",
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &row, nil
}

func (s *SQL) GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*persistence.AuthTokenRow, error) {
	var row persistence.AuthTokenRow
	err := s.getContext(
		ctx,
		&row,
		"SELECT * FROM auth_token WHERE refresh_token = ?",
		refreshToken,
	)
	if err != nil {
		return nil, fmt.Errorf("get auth token by refresh token: %w", err)
	}

	return &row, nil
}

func (s *SQL) createProvider(ctx context.Context, tx *sqlx.Tx, providerRow persistence.AuthProviderRow) error {
	_, err := tx.NamedExecContext(
		ctx,
		"INSERT INTO auth_provider (id, name, provider, provider_user_id, email, avatar, account_id, last_login, created_at, updated_at) VALUES (:id, :name, :provider, :provider_user_id, :email, :avatar, :account_id, :last_login, :created_at, :updated_at)",
		providerRow,
	)
	if err != nil {
		return fmt.Errorf("insert auth provider: %w", err)
	}

	return nil
}
