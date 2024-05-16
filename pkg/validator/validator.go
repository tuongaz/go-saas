package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/tuongaz/go-saas/pkg/errors/apierror"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

func Validate(input any) error {
	if err := validate.Struct(input); err != nil {
		return apierror.NewValidationError(err)
	}

	return nil
}
