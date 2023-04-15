package pkgpostgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/avast/retry-go"
	"github.com/aws/smithy-go/ptr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultRetryConnectAttempts = 10
	defaultRetryConnectDelay    = time.Second * 3
)

// Conn is a wrapper for sql.DB.
type Conn struct {
	*sql.DB
}

type Config struct {
	// Connection URL.
	// Example: postgres://user:password@localhost:5432/database
	DSN           string
	RetryAttempts *int
	RetryDelay    *time.Duration
}

func (c Config) prepare() Config {
	if c.RetryAttempts == nil {
		c.RetryAttempts = ptr.Int(defaultRetryConnectAttempts)
	}

	if c.RetryDelay == nil {
		c.RetryDelay = ptr.Duration(defaultRetryConnectDelay)
	}

	return c
}

// NewConn creates a new postgres connection.
func NewConn(ctx context.Context, cfg Config) (*Conn, error) {
	cfg.prepare()

	pgxCfg, err := pgx.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	var db *sql.DB
	err = retry.Do(func() error {
		db = stdlib.OpenDB(*pgxCfg)
		connErr := db.Ping()
		if connErr != nil {
			_ = db.Close()
			return connErr
		}

		return nil
	}, retry.Attempts(defaultRetryConnectAttempts), retry.Delay(defaultRetryConnectDelay), retry.DelayType(retry.FixedDelay), retry.Context(ctx))

	if err != nil {
		return nil, err
	}

	c := &Conn{db}

	return c, nil
}
