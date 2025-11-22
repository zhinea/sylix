package validator

import (
	"github.com/go-playground/validator/v10"
	baseValidator "github.com/zhinea/sylix/internal/common/validator"
	pbValidation "github.com/zhinea/sylix/internal/infra/proto/common"

	"github.com/zhinea/sylix/internal/module/controlplane/entity"
)

type ServerValidator struct {
	*baseValidator.BaseValidator
}

func NewServerValidator() *ServerValidator {
	base := baseValidator.NewBaseValidator()

	_ = base.RegisterValidation("credential_required", credentialRequired)
	base.RegisterTagMessage("credential_required", func(validator.FieldError) string {
		return "either password or SSH key must be provided"
	})

	return &ServerValidator{
		BaseValidator: base,
	}
}

func (v *ServerValidator) Validate(server *entity.Server) []*pbValidation.ValidationError {
	if errors := v.ValidateStruct(server); len(errors) > 0 {
		return errors
	}

	return v.validateBusinessRules(server)
}

func (v *ServerValidator) validateBusinessRules(server *entity.Server) []*pbValidation.ValidationError {
	var errors []*pbValidation.ValidationError

	// Custom business rule: Either password or SSH key must be provided
	if server.Credential.Password == nil && server.Credential.SSHKey == nil {
		errors = append(errors, &pbValidation.ValidationError{
			Field:   "Credential",
			Message: "either password or SSH key must be provided",
		})
	}

	return errors
}

func credentialRequired(fl validator.FieldLevel) bool {
	cred, ok := fl.Field().Interface().(entity.ServerCredential)
	if !ok {
		return false
	}

	return cred.Password != nil || cred.SSHKey != nil
}
