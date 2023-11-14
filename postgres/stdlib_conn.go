package pkgpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	pkgptr "github.com/amanbolat/pkg/ptr"
	"github.com/avast/retry-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	DefaultRetryConnectAttempts = 10
	DefaultRetryConnectDelay    = time.Second * 3
	DefaultMaxOpenConns         = 15
	DefaultMaxIdleConns         = 15
	DefaultConnMaxLifetime      = time.Minute * 5
)

// SQLConn is a wrapper around sql.DB that uses pgx driver under the hood.
type SQLConn struct {
	*sql.DB
}

type SQLConnConfig struct {
	DSN                  string
	RetryConnectAttempts *uint
	RetryConnectDelay    *time.Duration
	MaxOpenConns         *int
	MaxIdleConns         *int
	ConnMaxLifetime      *time.Duration
	ConnOptions          []stdlib.OptionOpenDB
}

func (c *SQLConnConfig) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("dsn is required")
	}

	if c.RetryConnectAttempts == nil {
		c.RetryConnectAttempts = pkgptr.Ptr(uint(DefaultRetryConnectAttempts))
	}

	if c.RetryConnectDelay == nil {
		c.RetryConnectDelay = pkgptr.Ptr(DefaultRetryConnectDelay)
	}

	if c.MaxOpenConns == nil {
		c.MaxOpenConns = pkgptr.Ptr(DefaultMaxOpenConns)
	}

	if c.MaxIdleConns == nil {
		c.MaxIdleConns = pkgptr.Ptr(DefaultMaxIdleConns)
	}

	if c.ConnMaxLifetime == nil {
		c.ConnMaxLifetime = pkgptr.Ptr(DefaultConnMaxLifetime)
	}

	return nil
}

// NewSQLConn creates and returns a new postgres connection.
// The function ensures that the connection is established correctly
// by send a ping request to the database.
func NewSQLConn(ctx context.Context, cfg SQLConnConfig) (*SQLConn, error) {
	err := cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid SQLConn config: %w", err)
	}

	pgxCfg, err := pgx.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	var db *sql.DB
	err = retry.Do(func() error {
		db = stdlib.OpenDB(*pgxCfg, cfg.ConnOptions...)
		connErr := db.Ping()
		if connErr != nil {
			_ = db.Close()
			return connErr
		}

		return nil
	}, retry.Attempts(*cfg.RetryConnectAttempts), retry.Delay(*cfg.RetryConnectDelay), retry.DelayType(retry.FixedDelay), retry.Context(ctx))

	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(*cfg.MaxOpenConns)
	db.SetMaxIdleConns(*cfg.MaxIdleConns)
	db.SetConnMaxLifetime(*cfg.ConnMaxLifetime)

	c := &SQLConn{db}

	return c, nil
}
