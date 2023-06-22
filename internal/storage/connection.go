package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/kkereziev/notifier/v2/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"

	_ "github.com/lib/pq"
)

// Querier is an interface for the DB operations.
type Querier interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

// Transactioner is an interface for the DB transaction operations.
type Transactioner interface {
	Tx(ctx context.Context, txFunc TxFunc) error
}

type contextKey string

func (c contextKey) String() string {
	return fmt.Sprintf("context key: %s", string(c))
}

var txKey = contextKey("txKey")

// TxFunc is a callback function, which holds all the operations of the transaction.
type TxFunc func(ctx context.Context) error

// TxPair is a pair of db transaction and connection.
type TxPair struct {
	tx *sqlx.Tx
	db *sqlx.DB
}

// Connection is a wrapper around sqlx.DB.
type Connection struct {
	db *sqlx.DB
}

// NewConnection is a constructor function for Connection.
func NewConnection(config *config.Config) (*Connection, error) {
	db, err := sql.Open("postgres", config.Database.DSN())
	if err != nil {
		return nil, err
	}

	// initiate zerolog
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zlogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// prepare logger
	loggerOptions := []sqldblogger.Option{
		sqldblogger.WithSQLQueryFieldname("sql"),
		sqldblogger.WithWrapResult(false),
		sqldblogger.WithExecerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithQueryerLevel(sqldblogger.LevelDebug),
		sqldblogger.WithPreparerLevel(sqldblogger.LevelDebug),
	}

	// wrap *sql.DB to transparent logger
	db = sqldblogger.OpenDriver(config.Database.DSN(), db.Driver(), zerologadapter.New(zlogger), loggerOptions...)

	// pass it sqlx
	sqlxDB := sqlx.NewDb(db, "postgres")

	// use sqlxDB as usual, no need any more change
	if err := sqlxDB.Ping(); err != nil {
		return nil, err
	}

	return &Connection{db: sqlxDB}, nil
}

// DB returns db instance, if there is a transaction going it returns the transaction.
func (c *Connection) DB(ctx context.Context) Querier {
	if dtp, ok := GetTransactionFromContext(ctx); ok {
		return dtp.tx
	}

	return c.db
}

// Tx starts a new db transaction. It can be used in the business logic
// to wrap many DB(Repository) operations and treat them a single unit of work.
func (c *Connection) Tx(ctx context.Context, txFunc TxFunc) error {
	return c.execTransaction(ctx, txFunc)
}

func (c *Connection) execTransaction(ctx context.Context, txFunc TxFunc) error {
	if _, ok := GetTransactionFromContext(ctx); ok {
		return txFunc(ctx)
	}

	errRun := c.runInTx(ctx, nil, func(ctx context.Context, tx *sqlx.Tx) error {
		return txFunc(setTransactionToContext(ctx, c.db, tx))
	})

	return errors.WithStack(errRun)
}

// GetTransactionFromContext retrieves the db transaction from the context.
func GetTransactionFromContext(ctx context.Context) (TxPair, bool) {
	dtp, ok := ctx.Value(txKey).(TxPair)

	return dtp, ok
}

func setTransactionToContext(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey, TxPair{
		db: db,
		tx: tx,
	})
}

func (c *Connection) runInTx(
	ctx context.Context,
	opts *sql.TxOptions,
	fn func(ctx context.Context, tx *sqlx.Tx) error,
) error {
	tx, err := c.db.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}

	var done bool

	defer func() {
		if !done {
			//nolint: errcheck
			_ = tx.Rollback()
		}
	}()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	done = true

	return tx.Commit()
}
