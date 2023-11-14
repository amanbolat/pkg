package pkgstore

import "context"

// AtomicOperation is a function that can be executed atomically on a store.
// This type of method is typically used in conjunction with AtomicStore.
//
// Example:
//
//	atomicStore.Exec(ctx, func(ctx context.Context, s Store) error {
//	    _ = s.DoOperationOne(ctx)
//	    _ = s.DoOperationTwo(ctx)
//	    return nil
//	})
type AtomicOperation[T any] func(ctx context.Context, store T) error

// AtomicStore is a store designed to execute operations atomically.
// Its purpose is to encapsulate the logic of transaction management for any type of database.
type AtomicStore[T any] interface {
	// Exec executes an operation atomically.
	// All operations within the function are executed as a single transaction.
	// Not all databases support transactions, so an external
	// service may be required to ensure this atomicity.
	Exec(ctx context.Context, op AtomicOperation[T]) error
}
