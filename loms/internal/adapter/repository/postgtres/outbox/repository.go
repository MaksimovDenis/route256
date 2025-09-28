package outbox

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

const queueSize = 100

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

func (r *Repository) CreateEvent(ctx context.Context, events domain.Event) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "outboxRepository.CreateEvent")
	defer span.Finish()
	defer func(now time.Time) {
		status := string(metrics.DBQueryStatusOK)
		if err != nil {
			status = string(metrics.DBQueryStatusError)
		}

		metrics.IncDBQueryCounter(string(metrics.Create), status)
		metrics.DBQueryDurationHistogram(string(metrics.Create), status, time.Since(now).Seconds())
	}(time.Now())

	querier := r.getMasterQuerier(ctx)

	arg := &sqlc.CreateEventParams{
		Topic:   events.Topic,
		Key:     &events.Key,
		Payload: events.Payload,
	}

	if err := querier.CreateEvent(ctx, arg); err != nil {
		fmt.Println(err)
		return fmt.Errorf("querier.CreateEvent sqlc failed: %w", err)
	}

	return nil
}

func (r *Repository) FetchNextMessages(ctx context.Context, limit int32) (domainEvents []domain.Event, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "outboxRepository.FetchNextMessages")
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

	eventsSqlc, err := querier.FetchNextMessages(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("querier.FetchNextMessage sqlc failed: %w", err)
	}

	if len(eventsSqlc) == 0 {
		return nil, nil
	}

	domainEvents = make([]domain.Event, len(eventsSqlc))

	for idx, event := range eventsSqlc {
		domainEvents[idx] = domain.Event{
			ID:      event.ID,
			Topic:   event.Topic,
			Key:     *event.Key,
			Payload: event.Payload,
			Status:  domain.EventStatus(event.Status),
		}
	}

	return domainEvents, nil
}

func (r *Repository) MarkAsSent(ctx context.Context, orderIDs []int64) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "outboxRepository.MarkAsSent")

	defer func(now time.Time) {
		status := string(metrics.DBQueryStatusOK)
		if err != nil {
			status = string(metrics.DBQueryStatusError)
		}

		metrics.IncDBQueryCounter(string(metrics.Update), status)
		metrics.DBQueryDurationHistogram(string(metrics.Update), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	if len(orderIDs) == 0 {
		return nil
	}

	querier := r.getMasterQuerier(ctx)

	for start := 0; start < len(orderIDs); start += queueSize {
		end := start + queueSize
		if end > len(orderIDs) {
			end = len(orderIDs)
		}

		queue := orderIDs[start:end]

		err := querier.MarkAsSent(ctx, queue)
		if err != nil {
			return fmt.Errorf("querier.MarkAsSent sqlc failed: %w", err)
		}
	}

	return nil
}

func (r *Repository) MarkAsError(ctx context.Context, orderIDs []int64) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "outboxRepository.MarkAsError")

	defer func(now time.Time) {
		status := string(metrics.DBQueryStatusOK)
		if err != nil {
			status = string(metrics.DBQueryStatusError)
		}

		metrics.IncDBQueryCounter(string(metrics.Update), status)
		metrics.DBQueryDurationHistogram(string(metrics.Update), status, time.Since(now).Seconds())

		span.Finish()
	}(time.Now())

	if len(orderIDs) == 0 {
		return nil
	}

	querier := r.getMasterQuerier(ctx)

	for start := 0; start < len(orderIDs); start += queueSize {
		end := start + queueSize
		if end > len(orderIDs) {
			end = len(orderIDs)
		}

		queue := orderIDs[start:end]

		err := querier.MarkAsError(ctx, queue)
		if err != nil {
			return fmt.Errorf("querier.MarkAsError sqlc failed: %w", err)
		}
	}

	return nil
}
