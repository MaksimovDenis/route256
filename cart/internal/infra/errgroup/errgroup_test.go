package errgroup_test

import (
	"context"
	"errors"
	"route256/cart/internal/infra/errgroup"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestErrGroup_Success(t *testing.T) {
	t.Parallel()

	group, _ := errgroup.New(context.Background())

	for i := 0; i < 5; i++ {
		group.Go(func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}

	err := group.Wait()
	require.NoError(t, err)
}

func TestErrGroup_ContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	group, groupCtx := errgroup.New(ctx)

	group.Go(func() error {
		cancel()
		select {
		case <-groupCtx.Done():
			return groupCtx.Err()
		case <-time.After(10 * time.Millisecond):
			return nil
		}
	})

	err := group.Wait()
	require.ErrorContains(t, err, "context canceled")
}

func TestErrGroup_Error(t *testing.T) {
	t.Parallel()

	group, ctx := errgroup.New(context.Background())

	start := make(chan struct{})
	done := make(chan struct{})

	group.Go(func() error {
		<-start
		time.Sleep(10 * time.Millisecond)
		return errors.New("fail")
	})

	for i := 0; i < 3; i++ {
		group.Go(func() error {
			<-start
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-done:
				return nil
			}
		})
	}

	close(start)
	err := group.Wait()
	close(done)

	require.ErrorContains(t, err, "fail")
}

func TestErrGroup_SetLimit(t *testing.T) {
	t.Parallel()

	const limit = 2
	var maxConcurrent int32
	var current int32

	group, _ := errgroup.New(context.Background())
	group.SetLimit(limit)

	for i := 0; i < 10; i++ {
		group.Go(func() error {
			val := atomic.AddInt32(&current, 1)
			defer atomic.AddInt32(&current, -1)

			for {
				old := atomic.LoadInt32(&maxConcurrent)
				if val <= old {
					break
				}
				if atomic.CompareAndSwapInt32(&maxConcurrent, old, val) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}

	err := group.Wait()
	require.NoError(t, err)
	require.LessOrEqual(t, maxConcurrent, int32(limit), "limit was exceeded")
}
