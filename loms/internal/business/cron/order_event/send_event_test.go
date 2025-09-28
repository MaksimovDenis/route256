package orderevent_test

import (
	"context"
	"fmt"
	"route256/loms/internal/domain"
	txmanager "route256/loms/internal/infra/tx_manager"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestHandlePendingEvents_SuccessAllSent(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{ID: 1, Key: "123", Payload: []byte("payload1")},
		{ID: 2, Key: "456", Payload: []byte("payload2")},
	}

	f := setUp(t)
	ctx := context.Background()

	f.txManager.ReadCommittedMock.Set(func(ctx context.Context, fn txmanager.Handler) error {
		return fn(ctx)
	})

	f.eventRepository.FetchNextMessagesMock.
		Expect(minimock.AnyContext, 100).
		Return(events, nil)

	f.producerKafka.SendOrderEventsBatchMock.Set(func(_ context.Context, _ []domain.Event) ([]int64, []int64, error) {
		return []int64{1, 2}, nil, nil
	})

	f.eventRepository.MarkAsSentMock.Set(func(_ context.Context, ids []int64) error {
		expected := []int64{1, 2}
		if !f.ElementsMatch(ids, expected) {
			t.Errorf("MarkAsSent got IDs %v, want %v", ids, expected)
		}
		return nil
	})

	err := f.executor.Do(ctx)
	f.NoError(err)
}

func TestHandlePendingEvents_OneMessageFails(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{ID: 1, Key: "123", Payload: []byte("payload1")},
		{ID: 2, Key: "456", Payload: []byte("payload2")},
	}

	f := setUp(t)
	ctx := context.Background()

	f.txManager.ReadCommittedMock.Set(func(ctx context.Context, fn txmanager.Handler) error {
		return fn(ctx)
	})

	f.eventRepository.FetchNextMessagesMock.
		Expect(minimock.AnyContext, 100).
		Return(events, nil)

	f.producerKafka.SendOrderEventsBatchMock.Set(func(_ context.Context, _ []domain.Event) ([]int64, []int64, error) {
		return []int64{2}, []int64{1}, nil
	})

	f.eventRepository.MarkAsSentMock.Set(func(_ context.Context, ids []int64) error {
		expected := []int64{2}
		if !f.ElementsMatch(ids, expected) {
			t.Errorf("MarkAsSent got IDs %v, want %v", ids, expected)
		}
		return nil
	})

	f.eventRepository.MarkAsErrorMock.Set(func(_ context.Context, ids []int64) error {
		expected := []int64{1}
		if !f.ElementsMatch(ids, expected) {
			t.Errorf("MarkAsError got IDs %v, want %v", ids, expected)
		}
		return nil
	})

	err := f.executor.Do(ctx)
	f.NoError(err)
}

func TestHandlePendingEvents_FetchNextMessagesError(t *testing.T) {
	t.Parallel()

	f := setUp(t)
	ctx := context.Background()

	f.txManager.ReadCommittedMock.Set(func(ctx context.Context, fn txmanager.Handler) error {
		return fn(ctx)
	})

	f.eventRepository.FetchNextMessagesMock.
		Expect(minimock.AnyContext, 100).
		Return(nil, fmt.Errorf("fetch error"))

	err := f.executor.Do(ctx)
	f.NoError(err)
}

func TestHandlePendingEvents_MarkAsSentError(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{ID: 1, Key: "123", Payload: []byte("payload1")},
		{ID: 2, Key: "456", Payload: []byte("payload2")},
	}

	f := setUp(t)
	ctx := context.Background()

	f.txManager.ReadCommittedMock.Set(func(ctx context.Context, fn txmanager.Handler) error {
		return fn(ctx)
	})

	f.eventRepository.FetchNextMessagesMock.
		Expect(minimock.AnyContext, 100).
		Return(events, nil)

	f.producerKafka.SendOrderEventsBatchMock.Set(func(_ context.Context, _ []domain.Event) ([]int64, []int64, error) {
		return []int64{1, 2}, nil, nil
	})

	f.eventRepository.MarkAsSentMock.Set(func(_ context.Context, _ []int64) error {
		return fmt.Errorf("mark sent error")
	})

	err := f.executor.Do(ctx)
	f.NoError(err)
}

func TestHandlePendingEvents_SendOrderEventsBatchError(t *testing.T) {
	t.Parallel()

	events := []domain.Event{
		{ID: 1, Key: "123", Payload: []byte("payload1")},
		{ID: 2, Key: "456", Payload: []byte("payload2")},
	}

	f := setUp(t)
	ctx := context.Background()

	f.txManager.ReadCommittedMock.Set(func(ctx context.Context, fn txmanager.Handler) error {
		return fn(ctx)
	})

	f.eventRepository.FetchNextMessagesMock.
		Expect(minimock.AnyContext, 100).
		Return(events, nil)

	f.producerKafka.SendOrderEventsBatchMock.Set(func(_ context.Context, _ []domain.Event) ([]int64, []int64, error) {
		return nil, nil, fmt.Errorf("send batch error")
	})

	err := f.executor.Do(ctx)
	f.NoError(err)
}
