package errgroup

import (
	"context"
	"fmt"
	"sync"
)

type token struct{}

type ErrGroup struct {
	cancel func(error)
	wg     sync.WaitGroup

	sem chan token

	errOnce sync.Once
	err     error
}

func New(ctx context.Context) (*ErrGroup, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &ErrGroup{cancel: cancel}, ctx
}

func (eg *ErrGroup) Go(action func() error) {
	if eg.sem != nil {
		eg.sem <- token{}
	}

	eg.wg.Add(1)
	go func() {
		defer eg.wg.Done()
		if eg.sem != nil {
			defer func() {
				<-eg.sem
			}()
		}
		if err := action(); err != nil {
			eg.errOnce.Do(func() {
				eg.err = err
				eg.cancel(err)
			})
		}
	}()
}

func (eg *ErrGroup) Wait() error {
	eg.wg.Wait()
	return eg.err
}

func (eg *ErrGroup) SetLimit(n int) {
	if n < 0 {
		eg.sem = nil
		return
	}
	if len(eg.sem) > 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(eg.sem)))
	}
	eg.sem = make(chan token, n)
}
