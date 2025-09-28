package stock

import (
	"context"
	"errors"
	"fmt"
	sqlc "route256/loms/internal/adapter/repository/postgtres/queries_sqlc_generated"
	"route256/loms/internal/domain"
	"route256/loms/internal/infra/metrics"
	txmanager "route256/loms/internal/infra/tx_manager"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opentracing/opentracing-go"
)

const batchSize = 1000

type postgresPools interface {
	GetWriteReplica() *pgxpool.Pool
	GetReadReplica() *pgxpool.Pool
}

type Repository struct {
	connPools postgresPools
}

func New(connPools postgresPools) *Repository {
	return &Repository{
		connPools: connPools,
	}
}

func (r *Repository) getMasterQuerier(ctx context.Context) *sqlc.Queries {
	tx, ok := ctx.Value(txmanager.TxKey).(pgx.Tx)
	if ok {
		return sqlc.New(tx)
	}

	return sqlc.New(r.connPools.GetWriteReplica())
}

func (r *Repository) getReplicaQuerier(ctx context.Context) *sqlc.Queries {
	tx, ok := ctx.Value(txmanager.TxKey).(pgx.Tx)
	if ok {
		return sqlc.New(tx)
	}

	return sqlc.New(r.connPools.GetReadReplica())
}

func (r *Repository) GetStockBySku(ctx context.Context, sku domain.Sku) (stock domain.Stock, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "stockRepository.GetStockBySku")
	defer func(now time.Time) {
		status := string(metrics.DBQueryStatusOK)
		if err != nil {
			status = string(metrics.DBQueryStatusError)
		}

		metrics.IncDBQueryCounter(string(metrics.Select), status)
		metrics.DBQueryDurationHistogram(string(metrics.Select), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	querier := r.getReplicaQuerier(ctx)

	stockSqlc, err := querier.GetStockBySku(ctx, int64(sku))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Stock{}, domain.ErrStockNotFound
		}

		return domain.Stock{}, fmt.Errorf("querier.GetStockBySku: %w", err)
	}

	return domain.Stock{
		TotalCount: stockSqlc.TotalCount,
		Reserved:   stockSqlc.Reserved,
	}, nil
}

func (r *Repository) GetStocksBySkuForUpdate(ctx context.Context, items []domain.Item) (result map[domain.Sku]domain.Stock, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "stockRepository.GetStocksBySkuForUpdate")

	defer func(now time.Time) {
		status := string(metrics.DBQueryStatusOK)
		if err != nil {
			status = string(metrics.DBQueryStatusError)
		}

		metrics.IncDBQueryCounter(string(metrics.Select), status)
		metrics.DBQueryDurationHistogram(string(metrics.Select), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	querier := r.getMasterQuerier(ctx)

	if len(items) == 0 {
		return nil, nil
	}

	domainSku := make([]int64, len(items))
	for idx, value := range items {
		domainSku[idx] = int64(value.Sku)
	}

	stocks, err := querier.GetStocksBySkuForUpdate(ctx, domainSku)
	if err != nil {
		return nil, fmt.Errorf("querier.GetStockBySkuForUpdate: %w", err)
	}

	if len(stocks) == 0 {
		return nil, domain.ErrStockNotFound
	}

	result = make(map[domain.Sku]domain.Stock, len(stocks))
	for _, value := range stocks {
		result[domain.Sku(value.Sku)] = domain.Stock{
			TotalCount: value.TotalCount,
			Reserved:   value.Reserved,
		}
	}

	return result, nil
}

func (r *Repository) UpdateStocks(ctx context.Context, stocks map[domain.Sku]domain.Stock) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "stockRepository.UpdateStocks")
	defer func(now time.Time) {
		status := string(metrics.DBQueryStatusOK)
		if err != nil {
			status = string(metrics.DBQueryStatusError)
		}

		metrics.IncDBQueryCounter(string(metrics.Update), status)
		metrics.DBQueryDurationHistogram(string(metrics.Update), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	querier := r.getMasterQuerier(ctx)

	if len(stocks) == 0 {
		return nil
	}

	type pair struct {
		Sku   int64
		Stock domain.Stock
	}

	pairs := make([]pair, 0, len(stocks))
	for sku, stock := range stocks {
		pairs = append(pairs, pair{
			Sku:   int64(sku),
			Stock: stock,
		})
	}

	for start := 0; start < len(pairs); start += batchSize {
		end := start + batchSize
		if end > len(pairs) {
			end = len(pairs)
		}
		batch := pairs[start:end]

		skus := make([]int64, 0, len(batch))
		totalCounts := make([]int64, 0, len(batch))
		reservedCounts := make([]int64, 0, len(batch))

		for _, p := range batch {
			skus = append(skus, p.Sku)
			totalCounts = append(totalCounts, p.Stock.TotalCount)
			reservedCounts = append(reservedCounts, p.Stock.Reserved)
		}

		err := querier.UpdateStocks(ctx, &sqlc.UpdateStocksParams{
			Sku:        skus,
			TotalCount: totalCounts,
			Reserved:   reservedCounts,
		})
		if err != nil {
			return fmt.Errorf("querier.UpdateStocks failed: %w", err)
		}
	}

	return nil
}
