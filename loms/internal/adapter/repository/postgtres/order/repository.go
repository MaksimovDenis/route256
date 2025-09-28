package order

import (
	"context"
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

type outboxRepository interface {
	CreateEvent(ctx context.Context, events domain.Event) error
}

type postgresPools interface {
	GetWriteReplica() *pgxpool.Pool
	GetReadReplica() *pgxpool.Pool
}

type Repository struct {
	connPools        postgresPools
	outboxRepository outboxRepository
}

func New(connPools postgresPools, outboxRepository outboxRepository) *Repository {
	return &Repository{
		connPools:        connPools,
		outboxRepository: outboxRepository,
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

func (r *Repository) GetByOrderID(ctx context.Context, orderID int64) (order domain.Order, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRepository.GetByOrderID")
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

	rows, err := querier.GetByOrderID(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("querier.GetByOrderID: %w", err)
	}
	if len(rows) == 0 {
		return domain.Order{}, domain.ErrOrderNotFound
	}

	order = domain.Order{
		UserID: rows[0].UserID,
		Status: domain.OrderStatus(rows[0].Status),
		Items:  make([]domain.Item, 0, len(rows)),
	}

	for _, row := range rows {
		item := domain.Item{
			Sku:   domain.Sku(row.Sku),
			Count: row.Count,
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *Repository) GetByOrderIDForUpdate(ctx context.Context, orderID int64) (order domain.Order, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRepository.GetByOrderIDForUpdate")
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

	rows, err := querier.GetByOrderIDForUpdate(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("querier.GetByOrderID: %w", err)
	}
	if len(rows) == 0 {
		return domain.Order{}, domain.ErrOrderNotFound
	}

	order = domain.Order{
		UserID: rows[0].UserID,
		Status: domain.OrderStatus(rows[0].Status),
		Items:  make([]domain.Item, 0, len(rows)),
	}

	for _, row := range rows {
		item := domain.Item{
			Sku:   domain.Sku(row.Sku),
			Count: row.Count,
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (r *Repository) SetStatus(ctx context.Context, orderID int64, status domain.OrderStatus) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRepository.SetStatus")
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

	args := &sqlc.SetStatusParams{
		Status: sqlc.OrderStatus(status),
		ID:     orderID,
	}

	if err := querier.SetStatus(ctx, args); err != nil {
		return fmt.Errorf("querier.SetStatus: %w", err)
	}

	return nil
}

func (r *Repository) CreateOrder(ctx context.Context, userID int64) (orderID int64, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRepository.CreateOrder")
	defer func(now time.Time) {
		status := string(metrics.DBQueryStatusOK)
		if err != nil {
			status = string(metrics.DBQueryStatusError)
		}

		metrics.IncDBQueryCounter(string(metrics.Create), status)
		metrics.DBQueryDurationHistogram(string(metrics.Create), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	querier := r.getMasterQuerier(ctx)

	orderID, err = querier.CreateOrder(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("querier.CreateOrder: failed to create order %w", err)
	}

	return orderID, nil
}

func (r *Repository) CreateOrderItems(ctx context.Context, orderID int64, items []domain.Item) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRepository.CreateOrderItems")
	defer func(now time.Time) {
		status := string(metrics.DBQueryStatusOK)
		if err != nil {
			status = string(metrics.DBQueryStatusError)
		}

		metrics.IncDBQueryCounter(string(metrics.Create), status)
		metrics.DBQueryDurationHistogram(string(metrics.Create), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	querier := r.getMasterQuerier(ctx)

	if len(items) == 0 {
		return nil
	}

	for start := 0; start < len(items); start += batchSize {
		end := start + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[start:end]

		params := &sqlc.CreateOrderItemsParams{
			OrderIds: make([]int64, len(batch)),
			Skus:     make([]int64, len(batch)),
			Counts:   make([]int64, len(batch)),
		}

		for i, item := range batch {
			params.OrderIds[i] = orderID
			params.Skus[i] = int64(item.Sku)
			params.Counts[i] = item.Count
		}

		if err := querier.CreateOrderItems(ctx, params); err != nil {
			return fmt.Errorf("CreateOrderItems sqlc failed: %w", err)
		}
	}

	return nil
}

func (r *Repository) SetStatusAndCreateEvent(
	ctx context.Context,
	orderID int64,
	status domain.OrderStatus,
	event domain.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderRepository.SetStatusAndCreateEvent")
	defer span.Finish()

	err := r.SetStatus(ctx, orderID, status)
	if err != nil {
		return fmt.Errorf("SetStatus sqlc failed: %w", err)
	}

	err = r.outboxRepository.CreateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("CreateEvent sqlc failed: %w", err)
	}

	return nil
}
