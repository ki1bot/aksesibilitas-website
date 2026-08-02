package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Queries struct {
	db DBTX
}

func New(database DBTX) *Queries {
	return &Queries{
		db: database,
	}
}

func (queries *Queries) WithTx(transaction pgx.Tx) *Queries {
	return &Queries{
		db: transaction,
	}
}
