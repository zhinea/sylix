package validator

import (
	pbValidation "github.com/zhinea/sylix/internal/infra/proto/common"
)

func ToProto(err []ValidationError) []pbValidation.ValidationError {
	// format to pbValidation
	var pbErrors []pbValidation.ValidationError
	for _, e := range err {
		pbErrors = append(pbErrors, pbValidation.ValidationError{
			Field:   e.Field,
			Message: e.Message,
		})
	}

	return pbErrors
}
