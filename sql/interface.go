package pkgsql

import (
	"context"
	"database/sql"
)

// Beginner begins a transaction.
type Beginner interface {
	BeginTx(ctx context.Context, txOptions *sql.TxOptions) (*sql.Tx, error)
}

// Committer commits a transaction.
type Committer interface {
	Commit() error
}

// Rollbacker rollbacks a transaction.
type Rollbacker interface {
	Rollback() error
}

// Execer executes a query without returning sql rows.
type Execer interface {
	ExecContext(ctx context.Context, sql string, arguments ...any) (sql.Result, error)
}

// Querier runs a single SQL query.
type Querier interface {
	QueryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
}

// Transactor commits or rollbacks a transaction.
type Transactor interface {
	Committer
	Rollbacker
}

// TableOperator can run Exec and Query operations on database.
type TableOperator interface {
	Execer
	Querier
}

// Tx is an interface for SQL transaction.
type Tx interface {
	Transactor
	TableOperator
}

// Database represents a database connection.
type Database interface {
	Beginner
	TableOperator
}
