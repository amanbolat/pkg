package pkgsql

import (
	"context"
	"database/sql"

	pkgstore "github.com/amanbolat/pkg/store"
)

type NewStoreFunc[T any] func(db Database, to TableOperator) T

type AtomicStore[T any] struct {
	db           Database
	newStoreFunc NewStoreFunc[T]
}

// NewAtomicStore returns new AtomicStore[T].
func NewAtomicStore[T any](db Database, newStoreFunc NewStoreFunc[T]) *AtomicStore[T] {
	return &AtomicStore[T]{
		db:           db,
		newStoreFunc: newStoreFunc,
	}
}

// Exec runs an atomic operation.
func (s *AtomicStore[T]) Exec(ctx context.Context, op pkgstore.AtomicOperation[T]) (err error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return err
	}

	defer func() {
		err = EndTx(tx, err)
	}()

	store := s.newStoreFunc(s.db, tx)
	err = op(ctx, store)
	if err != nil {
		return err
	}

	return nil
}
