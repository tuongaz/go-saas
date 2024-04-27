package sqlite

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/autopus/bootstrap/store/persistence"
)

func (s *SQL) GetAccount(ctx context.Context, accountID string) (*persistence.AccountRow, error) {
	var row persistence.AccountRow
	err := s.getContext(
		ctx,
		&row,
		"SELECT * FROM account WHERE id = ?",
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	return &row, nil
}

func (s *SQL) GetAccountRoles(ctx context.Context, accountID string) ([]*persistence.AccountRoleRow, error) {
	var rows []*persistence.AccountRoleRow
	err := s.selectContext(
		ctx,
		&rows,
		"SELECT * FROM account_role WHERE account_id = ?",
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("get account roles: %w", err)
	}

	return rows, nil
}

func (s *SQL) GetAccountRoleByID(ctx context.Context, accountRoleID string) (*persistence.AccountRoleRow, error) {
	var row persistence.AccountRoleRow
	err := s.getContext(
		ctx,
		&row,
		"SELECT * FROM account_role WHERE id = ?",
		accountRoleID,
	)
	if err != nil {
		return nil, fmt.Errorf("get account role by id: %w", err)
	}

	return &row, nil
}

func (s *SQL) GetAccountRole(ctx context.Context, organisationID, accountID string) (*persistence.AccountRoleRow, error) {
	var row persistence.AccountRoleRow
	err := s.getContext(
		ctx,
		&row,
		"SELECT * FROM account_role WHERE organisation_id = ? AND account_id = ?",
		organisationID,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("get account role: %w", err)
	}

	return &row, nil
}

func (s *SQL) GetOrganisationByID(ctx context.Context, id string) (*persistence.OrganisationRow, error) {
	var row persistence.OrganisationRow
	err := s.getContext(
		ctx,
		&row,
		"SELECT * FROM organisation WHERE id = ?",
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get organisation by id: %w", err)
	}

	return &row, nil
}

func (s *SQL) GetDefaultOwnerAccountByProvider(ctx context.Context, provider string, providerUserID string) (*persistence.AccountRow, *persistence.OrganisationRow, error) {
	var accountRow persistence.AccountRow
	err := s.getContext(
		ctx,
		&accountRow,
		"SELECT * FROM account WHERE id = (SELECT account_id FROM auth_provider WHERE provider = ? AND provider_user_id = ?)",
		provider,
		providerUserID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("get default owner account by provider: %w", err)
	}

	var organisationRow persistence.OrganisationRow
	err = s.getContext(
		ctx,
		&organisationRow,
		"SELECT * FROM organisation WHERE id = (SELECT organisation_id FROM account_role WHERE account_id = ? AND role = ?)",
		accountRow.ID,
		"OWNER",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("get default owner organisation by provider: %w", err)
	}

	return &accountRow, &organisationRow, nil
}

func (s *SQL) CreateOwnerAccount(
	ctx context.Context,
	accountRow persistence.AccountRow,
	orgRow persistence.OrganisationRow,
	providerRow persistence.AuthProviderRow,
	accountRole persistence.AccountRoleRow,
	userRow *persistence.UserRow,
) (err error) {
	tx, err := s.conn.Beginx()
	if err != nil {
		return fmt.Errorf("starting a transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				err = fmt.Errorf("rolling back transaction: %w", err)
			}
		}
	}()

	err = s.createOrganisation(ctx, tx, orgRow)
	if err != nil {
		return err
	}

	err = s.createAccount(ctx, tx, accountRow)
	if err != nil {
		return err
	}

	err = s.createProvider(ctx, tx, providerRow)
	if err != nil {
		return err
	}

	err = s.createAccountRole(ctx, tx, accountRole)
	if err != nil {
		return err
	}

	if userRow != nil {
		err = s.createUser(ctx, tx, userRow)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQL) createUser(ctx context.Context, tx *sqlx.Tx, row *persistence.UserRow) error {
	_, err := tx.NamedExecContext(
		ctx,
		"INSERT INTO auth_user (id, email, name, password, reset_password_code, reset_password_code_expired_timestamp, created_at, updated_at) VALUES (:id, :email, :name, :password, :reset_password_code, :reset_password_code_expired_timestamp, :created_at, :updated_at)",
		row,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (s *SQL) createOrganisation(ctx context.Context, tx *sqlx.Tx, row persistence.OrganisationRow) error {
	_, err := tx.NamedExecContext(
		ctx,
		"INSERT INTO organisation (id, created_at, updated_at) VALUES (:id, :created_at, :updated_at)",
		row,
	)
	if err != nil {
		return fmt.Errorf("insert organisation: %w", err)
	}

	return nil
}

func (s *SQL) createAccount(ctx context.Context, tx *sqlx.Tx, accRow persistence.AccountRow) error {
	_, err := tx.NamedExecContext(
		ctx,
		"INSERT INTO account (id, name, first_name, last_name, avatar, communication_email, created_at, updated_at) VALUES (:id, :name, :first_name, :last_name, :avatar, :communication_email, :created_at, :updated_at)",
		accRow,
	)
	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}

	return nil
}

func (s *SQL) createAccountRole(ctx context.Context, tx *sqlx.Tx, row persistence.AccountRoleRow) error {
	_, err := tx.NamedExecContext(
		ctx,
		"INSERT INTO account_role (id, account_id, organisation_id, role, created_at, updated_at) VALUES (:id, :account_id, :organisation_id, :role, :created_at, :updated_at)",
		row,
	)
	if err != nil {
		return fmt.Errorf("insert account role: %w", err)
	}

	return nil
}
