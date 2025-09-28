package txmanager

import (
	"context"
	"fmt"
	"route256/loms/internal/infra/logger"

	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
)

type transactor interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type txKeyType struct{}

var TxKey = txKeyType{}

type Handler func(ctx context.Context) error

type TxManager struct {
	db transactor
}

func New(db transactor) *TxManager {
	return &TxManager{
		db: db,
	}
}

func (mgr *TxManager) transaction(ctx context.Context, opts pgx.TxOptions, fnc Handler) (err error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return fnc(ctx)
	}

	tx, err = mgr.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	ctx = context.WithValue(ctx, TxKey, tx)

	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panica recovered: %v", rec)
		}

		if err != nil {
			if errRollBack := tx.Rollback(ctx); errRollBack != nil {
				logger.Errorf(ctx, "failed to rollback transaction %v", errRollBack)
			}

			return
		}

		if err == nil {
			err = tx.Commit(ctx)
			if err != nil {
				err = fmt.Errorf("tx commit failed: %w", err)
			}
		}
	}()

	if err = fnc(ctx); err != nil {
		err = fmt.Errorf("failed executing code inside transaction: %w", err)
	}

	return err
}

func (mgr *TxManager) ReadCommitted(ctx context.Context, fnc Handler) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "txmanager.ReadCommitted")
	defer span.Finish()

	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return mgr.transaction(ctx, txOpts, fnc)
}
