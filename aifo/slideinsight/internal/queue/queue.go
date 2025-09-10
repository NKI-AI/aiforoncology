// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file in the root of this repository.
package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/ratelimit"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusRetrying   TaskStatus = "retrying"
)

// Task represents a generic task that can be executed by the queue
type Task interface {
	// ID returns a unique identifier for the task
	ID() string

	// Type returns the type of task (e.g., "email", "algorithm", etc.)
	Type() string

	// Execute performs the actual work of the task
	Execute(ctx context.Context) error

	// MaxRetries returns the maximum number of retry attempts
	MaxRetries() int

	// RetryDelay returns the delay between retry attempts
	RetryDelay() time.Duration

	// String returns a human-readable description of the task
	String() string
}

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID    string        `json:"task_id"`
	TaskType  string        `json:"task_type"`
	Status    TaskStatus    `json:"status"`
	Error     error         `json:"error,omitempty"`
	Attempts  int           `json:"attempts"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

// QueueConfig holds configuration for the queue
type QueueConfig struct {
	// DefaultRateLimit is used for any task type not specifically overridden.
	DefaultRateLimit int `json:"default_rate_limit"`
	// RateLimits allows you to throttle individual task types:
	// e.g. {"email": 2, "algorithm": 10}
	RateLimits      map[string]int `json:"rate_limits"`
	Workers         int            `json:"workers"`
	BufferSize      int            `json:"buffer_size"`
	ShutdownTimeout time.Duration  `json:"shutdown_timeout"`
}

// DefaultQueueConfig returns a default queue configuration
func DefaultQueueConfig() QueueConfig {
	return QueueConfig{
		DefaultRateLimit: 10,
		RateLimits:       map[string]int{},
		Workers:          3,
		BufferSize:       100,
		ShutdownTimeout:  30 * time.Second,
	}
}

// Queue represents a rate-limited task queue with basic metrics.
type Queue struct {
	config         QueueConfig
	rateLimiters   map[string]ratelimit.Limiter
	taskChan       chan Task
	resultChan     chan TaskResult
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	running        atomic.Bool
	totalTasks     atomic.Int64
	completedTasks atomic.Int64
	failedTasks    atomic.Int64

	// per-type metrics
	typeCompleted sync.Map // map[string]*atomic.Int64
	typeFailed    sync.Map // map[string]*atomic.Int64
	typeDurSum    sync.Map // map[string]*atomic.Int64 (ms)

	mu sync.RWMutex
}

func NewQueue(config QueueConfig) *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &Queue{
		config:       config,
		rateLimiters: make(map[string]ratelimit.Limiter),
		taskChan:     make(chan Task, config.BufferSize),
		resultChan:   make(chan TaskResult, config.BufferSize),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Initialize rate limiters for each task type
	for typ, rl := range config.RateLimits {
		q.rateLimiters[typ] = ratelimit.New(rl)
	}

	return q
}

// IsRunning returns whether the queue is currently running
func (q *Queue) IsRunning() bool {
	return q.running.Load()
}

// Start starts the queue workers
func (q *Queue) Start() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running.Load() {
		return fmt.Errorf("queue is already running")
	}

	log.Info("Starting queue",
		"workers", q.config.Workers,
		"default_rate_limit", q.config.DefaultRateLimit,
		"buffer_size", q.config.BufferSize)

	// Start workers
	for i := 0; i < q.config.Workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	// Start result processor
	q.wg.Add(1)
	go q.resultProcessor()

	q.running.Store(true)
	log.Info("Queue started successfully")

	return nil
}

// Stop gracefully stops the queue
func (q *Queue) Stop() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running.Load() {
		return fmt.Errorf("queue is not running")
	}

	log.Info("Stopping queue", "shutdown_timeout", q.config.ShutdownTimeout)

	// Cancel context to signal workers to stop
	q.cancel()

	// Close task channel to prevent new tasks
	close(q.taskChan)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("Queue stopped gracefully")
	case <-time.After(q.config.ShutdownTimeout):
		log.Warn("Queue shutdown timeout reached")
	}

	close(q.resultChan)
	q.running.Store(false)

	return nil
}

// Submit adds a task to the queue
func (q *Queue) Submit(task Task) error {
	if !q.running.Load() {
		return fmt.Errorf("queue is not running")
	}

	select {
	case q.taskChan <- task:
		q.totalTasks.Add(1)
		log.Info("Task submitted to queue",
			"task_id", task.ID(),
			"task_type", task.Type(),
			"description", task.String())
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("queue is shutting down")
	default:
		return fmt.Errorf("queue is full")
	}
}

// worker processes tasks from the queue
func (q *Queue) worker(workerID int) {
	defer q.wg.Done()

	log.Info("Worker started", "worker_id", workerID)

	for {
		select {
		case task, ok := <-q.taskChan:
			if !ok {
				log.Info("Worker stopping - task channel closed", "worker_id", workerID)
				return
			}

			q.processTask(workerID, task)

		case <-q.ctx.Done():
			log.Info("Worker stopping - context cancelled", "worker_id", workerID)
			return
		}
	}
}

// resultProcessor processes task results
func (q *Queue) resultProcessor() {
	defer q.wg.Done()

	log.Info("Result processor started")

	for {
		select {
		case result, ok := <-q.resultChan:
			if !ok {
				log.Info("Result processor stopping - result channel closed")
				return
			}

			// Log result based on status
			switch result.Status {
			case TaskStatusCompleted:
				log.Info("Task result: completed",
					"task_id", result.TaskID,
					"task_type", result.TaskType,
					"duration", result.Duration,
					"attempts", result.Attempts)

			case TaskStatusFailed:
				log.Error("Task result: failed",
					"task_id", result.TaskID,
					"task_type", result.TaskType,
					"duration", result.Duration,
					"attempts", result.Attempts,
					"error", result.Error)

			case TaskStatusRetrying:
				log.Debug("Task result: retrying",
					"task_id", result.TaskID,
					"task_type", result.TaskType,
					"attempts", result.Attempts,
					"error", result.Error)
			}

		case <-q.ctx.Done():
			log.Info("Result processor stopping - context cancelled")
			return
		}
	}
}

// sendResult sends a task result to the result channel
func (q *Queue) sendResult(result TaskResult) {
	select {
	case q.resultChan <- result:
		// Result sent successfully
	default:
		// Result channel is full, log and drop
		log.Warn("Result channel full, dropping result",
			"task_id", result.TaskID,
			"status", result.Status)
	}
}

func (q *Queue) processTask(workerID int, task Task) {
	limiter, ok := q.rateLimiters[task.Type()]
	if !ok {
		// Use default rate limiter if no specific one is configured
		limiter = ratelimit.New(q.config.DefaultRateLimit)
	}
	limiter.Take()
	startTime := time.Now()
	attempts := 0
	maxRetries := task.MaxRetries()

	log.Info("Processing task",
		"worker_id", workerID,
		"task_id", task.ID(),
		"type", task.Type(),
		"max_retries", maxRetries,
	)

	for attempts <= maxRetries {
		attempts++
		taskCtx, cancel := context.WithTimeout(q.ctx, 5*time.Minute)
		err := task.Execute(taskCtx)
		cancel()

		endTime := time.Now()
		duration := endTime.Sub(startTime)
		result := TaskResult{
			TaskID:    task.ID(),
			TaskType:  task.Type(),
			Attempts:  attempts,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  duration,
		}

		if err == nil {
			result.Status = TaskStatusCompleted
			q.completedTasks.Add(1)
			// record per-type completed
			c, _ := q.typeCompleted.LoadOrStore(task.Type(), &atomic.Int64{})
			c.(*atomic.Int64).Add(1)
			// record duration
			sum, _ := q.typeDurSum.LoadOrStore(task.Type(), &atomic.Int64{})
			sum.(*atomic.Int64).Add(duration.Milliseconds())

			log.Info("Task completed",
				"worker_id", workerID,
				"task_id", task.ID(),
				"attempts", attempts,
				"duration", duration,
			)
			q.sendResult(result)
			return
		}

		// failure case
		result.Error = err
		if attempts > maxRetries {
			result.Status = TaskStatusFailed
			q.failedTasks.Add(1)
			f, _ := q.typeFailed.LoadOrStore(task.Type(), &atomic.Int64{})
			f.(*atomic.Int64).Add(1)
			sum, _ := q.typeDurSum.LoadOrStore(task.Type(), &atomic.Int64{})
			sum.(*atomic.Int64).Add(duration.Milliseconds())

			log.Error("Task failed after retries",
				"worker_id", workerID,
				"task_id", task.ID(),
				"attempts", attempts,
				"error", err,
				"duration", duration,
			)
			q.sendResult(result)
			return
		}

		// will retry
		result.Status = TaskStatusRetrying
		log.Warn("Task failed — retrying",
			"worker_id", workerID,
			"task_id", task.ID(),
			"attempt", attempts,
			"error", err,
			"delay", task.RetryDelay(),
		)
		q.sendResult(result)

		select {
		case <-time.After(task.RetryDelay()):
		case <-q.ctx.Done():
			log.Info("Retry cancelled due to shutdown",
				"worker_id", workerID, "task_id", task.ID(),
			)
			return
		}
	}
}

func (q *Queue) Stats() map[string]interface{} {
	stats := map[string]interface{}{
		"running":            q.running.Load(),
		"workers":            q.config.Workers,
		"default_rate_limit": q.config.DefaultRateLimit,
		"pending":            len(q.taskChan),
		"total":              q.totalTasks.Load(),
		"completed":          q.completedTasks.Load(),
		"failed":             q.failedTasks.Load(),
	}

	byType := make(map[string]map[string]interface{})
	q.typeCompleted.Range(func(key, val any) bool {
		typ := key.(string)
		c := val.(*atomic.Int64).Load()
		fVal, _ := q.typeFailed.LoadOrStore(typ, &atomic.Int64{})
		f := fVal.(*atomic.Int64).Load()
		sumVal, _ := q.typeDurSum.LoadOrStore(typ, &atomic.Int64{})
		sum := sumVal.(*atomic.Int64).Load()
		avg := int64(0)
		if c > 0 {
			avg = sum / c
		}
		byType[typ] = map[string]interface{}{
			"completed": c,
			"failed":    f,
			"avg_ms":    avg,
		}
		return true
	})
	stats["by_type"] = byType
	return stats
}
