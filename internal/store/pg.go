package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the connection pool and implements every store interface.
//
// One type rather than several: the ceremonies need users, credentials and
// challenges to move together inside a single transaction, and splitting them
// across types would only make that harder to express.
type DB struct {
	pool *pgxpool.Pool
}

// Open connects and verifies the connection, so a bad URL fails at startup.
func Open(ctx context.Context, url string) (*DB, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("store: connecting: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: pinging: %w", err)
	}
	return &DB{pool: pool}, nil
}

// Close releases the pool.
func (d *DB) Close() { d.pool.Close() }

// Pool exposes the pool for tests and for the health check.
func (d *DB) Pool() *pgxpool.Pool { return d.pool }

// norm maps a pgx no-rows error onto the package's own sentinel so callers do
// not import pgx to check for a miss.
func norm(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// NewSubject returns an opaque account identifier.
//
// The value is random, not a row id: the subject is what leaves this service
// inside every token, and a sequential one would leak how many accounts exist
// and invite callers to guess their neighbors.
func NewSubject() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "usr_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// RandomToken returns n random bytes as an unpadded base64url string.
func RandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
