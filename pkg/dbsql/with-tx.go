package dbsql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
)

// Txx is transaction interface for sql.Tx
type Tx interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)

	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Txx is transaction interface for sqlx.Tx
type Txx interface {
	Tx
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// TxFunc function signature when using sql.Tx for transaction.
type TxFunc func(ctx context.Context, tx Tx) error

// TxxFunc function signature when using sqlx.Tx for transaction.
type TxxFunc func(ctx context.Context, txx Txx) error

// WithTx calls the TxFunc with a new sql.Tx transaction.
// It simplifies implementation of Transactional SQLDB integration as there is
// no need to explicitly begin and commit the transactions within TxFunc.
// It will rollback the transaction if the TxFunc returns an error.
func WithTx(ctx context.Context, db *sql.DB, fn TxFunc, opts *sql.TxOptions) (err error) {
	tx, err := db.BeginTx(ctx, opts)
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

// WithTxx calls the TxxFunc with a new sqlx.Tx transaction.
// It simplifies implementation of Transactional SQLDB integration as there is
// no need to explicitly begin and commit the transactions within TxxFunc.
// It will rollback the transaction if the TxxFunc returns an error.
func WithTxx(ctx context.Context, db *sqlx.DB, fn TxxFunc, opts *sql.TxOptions) (err error) {
	txx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return status.ErrInternal().WithError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			txx.Rollback()
			panic(p)
		} else if err != nil {
			txx.Rollback()
		} else {
			if errc := txx.Commit(); errc != nil {
				err = status.ErrInternal().WithError(errc)
			}
		}
	}()
	err = fn(ctx, txx)
	if err != nil {
		if _, ok := err.(status.ErrServiceStatus); !ok {
			return status.ErrInternal().WithError(err)
		}
	}
	return err
}
