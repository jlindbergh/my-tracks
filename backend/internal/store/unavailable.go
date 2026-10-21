package store

import (
	"errors"
	"net"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// unavailableCodes are SQLSTATE codes, besides class 08 (connection
// exception), meaning the server cannot serve queries right now.
var unavailableCodes = []string{
	"57P01", // admin_shutdown
	"57P02", // crash_shutdown
	"57P03", // cannot_connect_now, e.g. while starting up
	"53300", // too_many_connections
}

// IsUnavailable reports whether err means the database cannot be reached or
// cannot serve queries right now, rather than that one query failed.
func IsUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "08") || slices.Contains(unavailableCodes, pgErr.Code)
	}
	var connectErr *pgconn.ConnectError
	if errors.As(err, &connectErr) {
		return true
	}
	// A connection that was refused, reset or timed out.
	var netErr net.Error
	return errors.As(err, &netErr)
}
