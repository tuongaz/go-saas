package persistence

import (
	"context"
	"database/sql"
	"time"
)

type AuthInterface interface {
	GetAccount(ctx context.Context, accountID string) (*AccountRow, error)
	GetDefaultOwnerAccountByProvider(ctx context.Context, provider string, providerUserID string) (*AccountRow, *OrganisationRow, error)
	CreateAuthToken(ctx context.Context, row AuthTokenRow) (sql.Result, error)
	UpdateAuthToken(ctx context.Context, id string, input UpdateAuthTokenInput) (sql.Result, error)
	GetAuthTokenByRefreshToken(ctx context.Context, refreshToken string) (*AuthTokenRow, error)
	GetAuthTokenByAccountRoleID(ctx context.Context, accountRoleID string) (*AuthTokenRow, error)
	CreateOwnerAccount(
		ctx context.Context,
		accountRow AccountRow,
		orgRow OrganisationRow,
		providerRow AuthProviderRow,
		accRole AccountRoleRow,
		userRow *UserRow,
	) (err error)
	GetUserByEmail(ctx context.Context, email string) (*UserRow, error)
	GetAccountRole(ctx context.Context, organisationID, accountID string) (*AccountRoleRow, error)
	GetAccountRoleByID(ctx context.Context, accountRoleID string) (*AccountRoleRow, error)
	GetAccountRoles(ctx context.Context, accountID string) ([]*AccountRoleRow, error)
	GetOrganisationByID(ctx context.Context, id string) (*OrganisationRow, error)
}

type UserRow struct {
	ID                                string    `db:"id"`
	Email                             string    `db:"email"`
	Name                              string    `db:"name"`
	Password                          string    `db:"password"`
	ResetPasswordCode                 string    `db:"reset_password_code"`
	ResetPasswordCodeExpiredTimestamp int64     `db:"reset_password_code_expired_timestamp"`
	CreatedAt                         time.Time `db:"created_at"`
	UpdatedAt                         time.Time `db:"updated_at"`
}

type AuthTokenRow struct {
	ID            string    `db:"id"`
	AccountRoleID string    `db:"account_role_id"`
	RefreshToken  string    `db:"refresh_token"`
	ExpiresAt     time.Time `db:"expires_at"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type UpdateAuthTokenInput struct {
	RefreshToken string    `db:"refresh_token"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type AuthProviderRow struct {
	ID             string    `db:"id"`
	Name           string    `db:"name"`
	Provider       string    `db:"provider"`
	ProviderUserID string    `db:"provider_user_id"`
	Email          string    `db:"email"`
	Avatar         string    `db:"avatar"`
	AccountID      string    `db:"account_id"`
	LastLogin      time.Time `db:"last_login"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func GetPageSize(size int) int {
	if size == 0 {
		return 10
	}

	return size
}

type AccountRow struct {
	ID                 string    `db:"id"`
	Name               string    `db:"name"`
	FirstName          string    `db:"first_name"`
	LastName           string    `db:"last_name"`
	Avatar             string    `db:"avatar"`
	CommunicationEmail string    `db:"communication_email"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

type AccountRoleRow struct {
	ID             string    `db:"id"`
	AccountID      string    `db:"account_id"`
	OrganisationID string    `db:"organisation_id"`
	Role           string    `db:"role"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type OrganisationRow struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
