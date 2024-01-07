package main

import (
	"context"
	"runtime"
	"sync"

	"gitm/workerpool"
)

type Worker struct {
	wg     sync.WaitGroup
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	queue []func(context.Context) error

	pool *workerpool.WorkerPool[struct{}]
}

func NewWorker(max int, failFast bool) *Worker {
	if max <= 0 {
		// Use number of CPUs
		max = runtime.NumCPU()
	}

	return &Worker{
		pool: workerpool.Go[struct{}](max, failFast),
	}
}

func (p *Worker) Run(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)

	go func() {
		// Add jobs to the worker pool.
		for i, _ := range p.queue {
			if p.ctx.Err() != nil {
				break
			}

			fn := p.queue[i]
			p.pool.AddJob(func() (struct{}, error) {
				if p.ctx.Err() != nil {
					return struct{}{}, p.ctx.Err()
				}

				err := fn(p.ctx)
				return struct{}{}, err
			})
		}

		p.pool.Done()
	}()

	// Process results
	var err error
	for result := range p.pool.Results() {
		if result.Err != nil {
			if nil == err {
				err = result.Err
			}
		}
	}
	p.pool.Wait()
	return err
}

func (p *Worker) Stop() {
	p.cancel()
}

func (p *Worker) Add(f func(context.Context) error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue = append(p.queue, f)
}
