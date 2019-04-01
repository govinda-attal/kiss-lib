package pgxy

import (
	"context"

	"github.com/jackc/pgx"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
)

// Tx is transaction interface for sql.Tx
type Tx interface {
	Exec(sql string, arguments ...interface{}) (commandTag pgx.CommandTag, err error)
	ExecEx(ctx context.Context, sql string, options *pgx.QueryExOptions, arguments ...interface{}) (commandTag pgx.CommandTag, err error)

	Prepare(name, sql string) (*pgx.PreparedStatement, error)
	PrepareEx(ctx context.Context, name, sql string, opts *pgx.PrepareExOptions) (*pgx.PreparedStatement, error)

	Query(sql string, args ...interface{}) (*pgx.Rows, error)
	QueryEx(ctx context.Context, sql string, options *pgx.QueryExOptions, args ...interface{}) (*pgx.Rows, error)

	QueryRow(sql string, args ...interface{}) *pgx.Row
	QueryRowEx(ctx context.Context, sql string, options *pgx.QueryExOptions, args ...interface{}) *pgx.Row
}

// TxFunc function signature when using pgx.Tx for transaction.
type TxFunc func(ctx context.Context, tx Tx) error

// WithTx calls the TxFunc with a new sql.Tx transaction.
// It simplifies implementation of Transactional SQLDB integration as there is
// no need to explicitly begin and commit the transactions within TxFunc.
// It will rollback the transaction if the TxFunc returns an error.
func WithTx(ctx context.Context, pool *pgx.ConnPool, fn TxFunc, opts *pgx.TxOptions) (err error) {
	tx, err := pool.BeginEx(ctx, opts)
	if err != nil {
		return status.ErrInternal().WithError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if errc := tx.Commit(); errc != nil {
				err = status.ErrInternal().WithError(errc)
			}
		}
	}()
	err = fn(ctx, tx)
	if err != nil {
		if _, ok := err.(status.ErrServiceStatus); !ok {
			return status.ErrInternal().WithError(err)
		}
	}
	return err
}
