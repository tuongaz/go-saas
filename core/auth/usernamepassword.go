package auth

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/tuongaz/go-saas/pkg/uid"
	"github.com/tuongaz/go-saas/service/emailer"
	"golang.org/x/crypto/bcrypt"

	model2 "github.com/tuongaz/go-saas/core/auth/model"
	"github.com/tuongaz/go-saas/core/auth/store"
	"github.com/tuongaz/go-saas/pkg/apierror"
	coreStore "github.com/tuongaz/go-saas/store"
)

type SignupInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ResetPasswordRequestInput struct {
	Email string `json:"email"`
}

type ResetPasswordConfirmInput struct {
	Code     string `json:"code"`
	Password string `json:"password"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *service) signupUsernamePasswordAccount(
	ctx context.Context,
	input *SignupInput,
) (*model2.AuthenticatedInfo, error) {
	found, err := s.store.LoginCredentialsUserEmailExists(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("check email exists: %w", err)
	}

	if found {
		return nil, apierror.NewValidationError("Email already exists", nil, nil)
	}

	hashedPw, err := s.hashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	firstName, lastName := splitName(input.Name)

	ownerAcc, org, loginProvider, accountRole, err := s.store.CreateOwnerAccount(ctx, store.CreateOwnerAccountInput{
		Email:     strings.TrimSpace(strings.ToLower(input.Email)),
		Name:      input.Name,
		FirstName: firstName,
		LastName:  lastName,
		Provider:  model2.AuthProviderUsernamePassword,
		Password:  hashedPw,
	})
	if err != nil {
		return nil, fmt.Errorf("create owner account: %w", err)
	}

	if _, err := s.CreateAccessToken(ctx, accountRole.ID, loginProvider.ProviderUserID, DeviceFromCtx(ctx)); err != nil {
		return nil, fmt.Errorf("create auth token: %w", err)
	}

	if err := s.OnAccountCreated().Trigger(ctx, &OnAccountCreatedEvent{
		AccountID:      ownerAcc.ID,
		OrganisationID: org.ID,
	}); err != nil {
		return nil, fmt.Errorf("trigger on account created: %w", err)
	}

	out, err := s.getAuthenticatedInfo(ctx, accountRole, loginProvider.ProviderUserID, DeviceFromCtx(ctx))
	if err != nil {
		return nil, fmt.Errorf("get authenticated info: %w", err)
	}

	return out, nil
}

func (s *service) resetPasswordConfirm(ctx context.Context, input *ResetPasswordConfirmInput) error {
	req, err := s.store.GetResetPasswordRequest(ctx, input.Code)
	if err != nil {
		if coreStore.IsNotFoundError(err) {
			return apierror.NewValidationError("reset password request not found", nil)
		}

		return fmt.Errorf("auth: reset password confirm - GetResetPasswordRequest: %w", err)
	}

	if req.IsExpired(s.cfg.ResetPasswordRequestExpiryMinutes) {
		return apierror.NewValidationError("reset password request expired", nil)
	}

	hashedPw, err := s.hashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("auth: reset password confirm - hash password: %w", err)
	}

	if err := s.store.UpdateLoginCredentialsUserPassword(ctx, req.UserID, hashedPw); err != nil {
		return fmt.Errorf("auth: reset password confirm - UpdateLoginCredentialsUserPassword: %w", err)
	}

	if err := s.store.DeleteResetPasswordRequest(ctx, req.ID); err != nil {
		return fmt.Errorf("auth: reset password confirm - DeleteResetPasswordRequest: %w", err)
	}

	return nil
}

func (s *service) resetPasswordRequest(ctx context.Context, input *ResetPasswordRequestInput) error {
	user, err := s.store.GetLoginCredentialsUserByEmail(ctx, strings.TrimSpace(strings.ToLower(input.Email)))
	if err != nil {
		if !coreStore.IsNotFoundError(err) {
			return fmt.Errorf("auth: reset password request - GetLoginCredentialsUserByEmail: %w", err)
		}

		// yes, we don't want to return error to frontend if email not found
		return nil
	}

	pwRequest, err := s.store.CreateResetPasswordRequest(ctx, user.ID, uid.ID())
	if err != nil {
		return fmt.Errorf("auth: reset password request - CreateResetPasswordRequest: %w", err)
	}

	// send email with code
	const emailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Reset Your Password</title>
</head>
<body>
    <p>Hi {{.name}},</p>
    <p>You requested to reset your password. Please click the link below to reset it:</p>
    <p><a href="{{.reset_link}}">Reset Password</a></p>
    <p>If you didn't request a password reset, please ignore this email.</p>
</body>
</html>
`
	tmpl, err := template.New("resetPasswordEmail").Parse(emailTemplate)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, map[string]any{
		"name":       user.Name,
		"reset_link": fmt.Sprintf("%s/auth/reset-password-confirm?code=%s", s.cfg.BaseURL, pwRequest.Code),
	}); err != nil {
		return fmt.Errorf("auth: reset password request - execute email template: %w", err)
	}

	out, err := s.emailer.Send(ctx, emailer.SendEmailInput{
		From:    s.cfg.EmailFrom,
		To:      []string{user.Email},
		HTML:    body.String(),
		Subject: "Reset Your Password",
	})
	if err != nil {
		return fmt.Errorf("auth: reset password request - send email: %w", err)
	}

	if err := s.store.UpdateResetPasswordReceipt(ctx, pwRequest.ID, out.ID); err != nil {
		return fmt.Errorf("auth: reset password request - UpdateResetPasswordReceipt: %w", err)
	}

	return nil
}

func (s *service) loginUsernamePasswordAccount(
	ctx context.Context,
	input *LoginInput,
) (*model2.AuthenticatedInfo, error) {
	user, err := s.store.GetLoginCredentialsUserByEmail(ctx, input.Email)
	if err != nil {
		if coreStore.IsNotFoundError(err) {
			return nil, apierror.NewUnauthorizedErr("invalid credentials", nil)
		}

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if !s.isPasswordMatched(input.Password, user.Password) {
		return nil, apierror.NewUnauthorizedErr("invalid credentials", nil)
	}

	acc, err := s.store.GetAccountByLoginProvider(ctx, model2.AuthProviderUsernamePassword, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get account by auth provider: %w", err)
	}

	org, err := s.store.GetOrganisationByAccountIDAndRole(ctx, acc.ID, string(model2.RoleOwner))
	if err != nil {
		return nil, fmt.Errorf("get default owner account by provider: %w", err)
	}

	accountRole, err := s.store.GetAccountRoleByOrgAndAccountID(ctx, org.ID, acc.ID)
	if err != nil {
		return nil, fmt.Errorf("get account role: %w", err)
	}

	return s.getAuthenticatedInfo(ctx, accountRole, user.ID, DeviceFromCtx(ctx))
}

func (s *service) hashPassword(password string) (string, error) {
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}

	return string(hashedPw), nil
}

func (s *service) isPasswordMatched(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func (s *service) changePassword(ctx context.Context, accountID string, input *ChangePasswordInput) error {
	// Find login providers for this account with username_password provider
	// Query login_provider table for the account record
	loginProviders, err := s.store.GetLoginProviderByAccountID(ctx, accountID, model2.AuthProviderUsernamePassword)
	if err != nil {
		return fmt.Errorf("auth: change password - get login providers: %w", err)
	}

	// Check if username_password provider exists
	if loginProviders == nil {
		return apierror.NewValidationError("account does not have a username/password login method", nil)
	}

	// Get the login credentials user using the provider user ID
	user, err := s.store.GetLoginCredentialsUserByEmail(ctx, loginProviders.Email)
	if err != nil {
		return fmt.Errorf("auth: change password - get user: %w", err)
	}

	// Verify current password
	if !s.isPasswordMatched(input.CurrentPassword, user.Password) {
		return apierror.NewValidationError("current password is incorrect", nil)
	}

	// Hash and update new password
	hashedPw, err := s.hashPassword(input.NewPassword)
	if err != nil {
		return fmt.Errorf("auth: change password - hash password: %w", err)
	}

	if err := s.store.UpdateLoginCredentialsUserPassword(ctx, user.ID, hashedPw); err != nil {
		return fmt.Errorf("auth: change password - update password: %w", err)
	}

	return nil
}

func splitName(name string) (string, string) {
	var firstName, lastName string
	if name != "" {
		names := strings.Split(name, " ")
		firstName = names[0]
		lastName = names[len(names)-1]
	}

	return firstName, lastName
}
