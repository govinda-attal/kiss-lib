package jwtkit

import (
	"github.com/govinda-attal/kiss-lib/pkg/core/status"
)

// ErrTokenInValid represents an invalid JWT token.
func ErrTokenInValid() status.ErrServiceStatus {
	return status.ErrUnauthorized().WithMessage("Token is not valid")
}
