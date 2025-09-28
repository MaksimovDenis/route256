package daemon

import (
	"context"
	"route256/loms/internal/infra/logger"
	"sync"
	"time"
)

type eventCronProcessor interface {
	Do(ctx context.Context) error
}

type Daemon struct {
	eventCronProcessor eventCronProcessor
	interval           time.Duration
	startOnce          sync.Once
}

func New(eventCronProcessor eventCronProcessor, interval time.Duration) *Daemon {
	return &Daemon{
		eventCronProcessor: eventCronProcessor,
		interval:           interval,
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
				if err := d.eventCronProcessor.Do(ctx); err != nil {
					logger.Errorf(ctx, "cronProcessor Do error: %v", err)
				}
			}
		}
	})
}
