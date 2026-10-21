package store

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUnavailable(t *testing.T) {
	refused := &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}
	for _, tt := range []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"other error", errors.New("boom"), false},
		{"connection failure", &pgconn.PgError{Code: "08006"}, true},
		{"server shutting down", fmt.Errorf("insert activity: %w", &pgconn.PgError{Code: "57P01"}), true},
		{"server starting up", &pgconn.PgError{Code: "57P03"}, true},
		{"unique violation", &pgconn.PgError{Code: "23505"}, false},
		{"invalid coordinate", &pgconn.PgError{Code: "XX000"}, false},
		{"connection refused", fmt.Errorf("begin import: %w", refused), true},
	} {
		if got := IsUnavailable(tt.err); got != tt.want {
			t.Errorf("%s: IsUnavailable(%v) = %v, want %v", tt.name, tt.err, got, tt.want)
		}
	}
}

func TestIsUnavailableForUnreachableDatabase(t *testing.T) {
	// Nothing listens on port 1, so this needs no database.
	_, err := New(context.Background(), "postgres://u:p@127.0.0.1:1/db?connect_timeout=2")
	if err == nil {
		t.Fatal("New succeeded without a database")
	}
	if !IsUnavailable(err) {
		t.Errorf("IsUnavailable(%v) = false, want true", err)
	}
}
