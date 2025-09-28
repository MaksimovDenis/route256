package daemon

import (
	"context"
	"route256/cart/internal/infra/logger"
	"sync"
	"time"
)

type cartCronProcessor interface {
	Do(ctx context.Context) error
}

type Daemon struct {
	cartCronProcessor cartCronProcessor
	interval          time.Duration
	startOnce         sync.Once
}

func New(cartCronProcessor cartCronProcessor, interval time.Duration) *Daemon {
	return &Daemon{
		cartCronProcessor: cartCronProcessor,
		interval:          interval,
	}
}

func (d *Daemon) Start(ctx context.Context) {
	d.startOnce.Do(func() {
		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Infof(ctx, "Daemon stopped")
				return
			case <-ticker.C:
				if err := d.cartCronProcessor.Do(ctx); err != nil {
					logger.Errorf(ctx, "cronProcessor Do error: %v", err)
				}
			}
		}
	})
}
