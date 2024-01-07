package workerpool

import (
	"sync"
)

// JobResult represents the result of a job.
type JobResult[T any] struct {
	Result T
	Err    error
}

// WorkerPool represents a pool of workers.
type WorkerPool[T any] struct {
	workers         []*worker[T]
	jobs            chan jobFunc[T]
	results         chan JobResult[T]
	failFast        bool
	hasFailed       bool
	jobWaitGroup    sync.WaitGroup
	workerWaitGroup sync.WaitGroup
	mu              sync.Mutex
}

// jobFunc represents a function that returns an error.
type jobFunc[T any] func() (T, error)

// worker represents a worker routine.
type worker[T any] struct {
	id   int
	pool *WorkerPool[T]
}

// New creates a new worker pool with the specified number of workers and failFast option.
func New[T any](numWorkers int, failFast bool) *WorkerPool[T] {
	pool := &WorkerPool[T]{
		workers:  make([]*worker[T], numWorkers),
		jobs:     make(chan jobFunc[T]),
		results:  make(chan JobResult[T]),
		failFast: failFast,
	}

	for i := 0; i < numWorkers; i++ {
		w := &worker[T]{
			id:   i + 1,
			pool: pool,
		}
		pool.workers[i] = w
	}

	return pool
}

// Start starts the worker pool.
func (wp *WorkerPool[T]) Start() {
	for i := 0; i < len(wp.workers); i++ {
		wp.workerWaitGroup.Add(1)
		go wp.workers[i].start()
	}
}

// Go creates a new worker pool with the specified number of workers and failFast option
// and starts it.
func Go[T any](numWorkers int, failFast bool) *WorkerPool[T] {
	pool := New[T](numWorkers, failFast)
	pool.Start()
	return pool
}

// Results returns a channel of job results.
func (wp *WorkerPool[T]) Results() <-chan JobResult[T] {
	return wp.results
}

// AddJob adds a job function to the worker pool.
func (wp *WorkerPool[T]) AddJob(job jobFunc[T]) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.failFast && wp.hasFailed {
		// If failFast is enabled and any job has already failed, do not accept new jobs.
		return
	}

	wp.jobWaitGroup.Add(1)
	wp.jobs <- job
}

// Done waits for all jobs to complete and closes the channels.
func (wp *WorkerPool[T]) Done() {
	wp.jobWaitGroup.Wait()
	close(wp.jobs)
	close(wp.results)
	wp.workerWaitGroup.Wait()
}

// Wait waits for all jobs and workers to complete.
func (wp *WorkerPool[T]) Wait() {
	wp.jobWaitGroup.Wait()
	wp.workerWaitGroup.Wait()
}

// worker.start starts the worker routine.
func (w *worker[T]) start() {
	defer w.pool.workerWaitGroup.Done()

	for job := range w.pool.jobs {
		if w.pool.failFast && w.pool.hasFailed {
			w.pool.jobWaitGroup.Done()
			continue
		}
		result, err := job()

		w.pool.results <- JobResult[T]{
			Result: result,
			Err:    err,
		}

		w.pool.jobWaitGroup.Done()

		if w.pool.failFast && err != nil {
			w.pool.hasFailed = true
		}
	}
}
