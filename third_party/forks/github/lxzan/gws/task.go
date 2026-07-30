package gws

import (
	"sync"

	"github.com/icha-senpai/note/third_party/forks/github/lxzan/gws/internal"
)

type (
	// Task queue
	workerQueue struct {
		// mutex
		mu sync.Mutex

		// double-ended queue to store asynchronous jobs
		q internal.Deque[asyncJob]

		// maximum concurrency
		maxConcurrency int32

		// current concurrency
		curConcurrency int32
	}

	// Asynchronous job
	asyncJob func()
)

// Creates a task queue
func newWorkerQueue(maxConcurrency int32) *workerQueue {
	c := &workerQueue{
		mu:             sync.Mutex{},
		maxConcurrency: maxConcurrency,
		curConcurrency: 0,
	}
	return c
}

// Retrieves a job from the worker queue
func (c *workerQueue) getJob(newJob asyncJob, delta int32) asyncJob {
	c.mu.Lock()
	defer c.mu.Unlock()

	if newJob != nil {
		c.q.PushBack(newJob)
	}
	c.curConcurrency += delta
	if c.curConcurrency >= c.maxConcurrency {
		return nil
	}
	var job = c.q.PopFront()
	if job == nil {
		return nil
	}
	c.curConcurrency++
	return job
}

// Do continuously executes jobs in the worker queue
func (c *workerQueue) do(job asyncJob) {
	for job != nil {
		job()
		job = c.getJob(nil, -1)
	}
}

// Adds a job to the queue and executes it immediately if resources are available
func (c *workerQueue) Push(job asyncJob) {
	if nextJob := c.getJob(job, 0); nextJob != nil {
		go c.do(nextJob)
	}
}

type channel chan struct{}

func (c channel) add() { c <- struct{}{} }

func (c channel) done() { <-c }

func (c channel) Go(m *Message, f func(*Message) error) error {
	c.add()
	go func() {
		_ = f(m)
		c.done()
	}()
	return nil
}
