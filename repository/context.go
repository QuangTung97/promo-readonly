package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
)

// GetTx get Transaction from context
func GetTx(ctx context.Context) Transaction {
	tx, ok := ctx.Value(ctxTxKey).(ctxTxValue)
	if !ok {
		panic("Not found transaction")
	}
	return tx.tx
}

// GetReadonly get Readonly from context
func GetReadonly(ctx context.Context) Readonly {
	db, ok := ctx.Value(ctxReadonlyKey).(ctxReadonlyValue)
	if !ok {
		panic("Not found readonly repository")
	}
	return db.db
}

type ctxTxKeyType struct {
}

type ctxReadonlyKeyType struct {
}

var ctxTxKey = ctxTxKeyType{}
var ctxReadonlyKey = ctxReadonlyKeyType{}

type ctxTxValue struct {
	tx *sqlx.Tx
}

type ctxReadonlyValue struct {
	db *sqlx.DB
}
