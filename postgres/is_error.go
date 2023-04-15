package pkgpostgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorCode is a Postgres error code.
type ErrorCode string

// IsError checks if the error is a Postgres error with the given code
// and constraint name if one was provided.
func IsError(err error, code string, constraintName *string) bool {
	if err == nil {
		return false
	}

	if code == "" {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	if pgErr.Code != code {
		return false
	}

	if constraintName != nil && *constraintName != pgErr.ConstraintName {
		return false
	}

	return true
}
