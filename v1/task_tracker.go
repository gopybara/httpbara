package httpbara

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

var (
	// ErrTerminating is returned when attempting to start a new task after
	// the shutdown process has already been initiated. Once the tracker
	// begins terminating, no new tasks can be registered.
	ErrTerminating = errors.New("received terminating signal")
)

// ActiveTaskTracker is a utility for managing and monitoring a group of
// concurrent tasks. It keeps track of the number of tasks currently running,
// and provides a controlled way to initiate shutdown and wait for all tasks
// to complete.
//
// Key features:
// - Incrementing and decrementing an active task count.
// - Initiating a termination signal that prevents new tasks from starting.
// - Waiting until all currently active tasks finish once termination is signaled.
//
// This is particularly useful in servers or background processes where you
// need to gracefully shut down, ensuring all ongoing operations complete
// before the application fully stops.
type ActiveTaskTracker struct {
	// count holds the current number of active tasks.
	count atomic.Int32

	// ctx is the context used to detect termination state.
	// Once ctx is canceled, no new tasks should be started.
	ctx context.Context

	// cancelFunc is the cancellation function associated with ctx.
	// Calling it initiates the shutdown process.
	cancelFunc context.CancelFunc

	// doneCh is a channel used to signal the completion of all remaining
	// tasks once a shutdown has been initiated.
	doneCh chan bool
}

// StartTask increments the count of active tasks by one, representing the start
// of a new task. If the tracker has already begun the termination process
// (i.e., ctx is canceled), this method returns ErrTerminating instead of
// incrementing the count.
//
// Returns:
//   - nil if the task was successfully started.
//   - ErrTerminating if the tracker is in the process of shutting down.
func (t *ActiveTaskTracker) StartTask() error {
	if t.ctx.Err() != nil {
		return ErrTerminating
	}

	t.count.Add(1)
	return nil
}

// FinishTask decrements the count of active tasks by one, signaling that a
// previously started task has completed. If the termination process was
// initiated (ctx is canceled) and this results in zero remaining tasks, it
// sends a signal on the doneCh channel to unblock any waiting Shutdown()
// calls.
func (t *ActiveTaskTracker) FinishTask() {
	t.count.Add(-1)
	if t.ctx.Err() != nil && t.count.Load() == 0 {
		t.doneCh <- true
	}
}

// TaskCount returns the current number of active tasks. This can be used
// for monitoring or logging purposes, providing visibility into the number
// of tasks currently running.
func (t *ActiveTaskTracker) TaskCount() int32 {
	return t.count.Load()
}

// Shutdown initiates the graceful shutdown process. Once called, no new tasks
// can be started. The method will:
//
//  1. Call cancelFunc to signal that shutdown has begun.
//  2. If there are any active tasks, this method will block until all tasks
//     have completed and FinishTask sends a completion signal.
//  3. If no tasks are currently active, Shutdown returns immediately.
//
// Usage scenario:
// You might call Shutdown() in response to receiving a termination signal
// (e.g., SIGTERM) in a server. This ensures that all in-flight requests or
// background jobs complete before the program exits.
func (t *ActiveTaskTracker) Shutdown(ctx context.Context) error {
	t.cancelFunc()

	if t.count.Load() == 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	case <-t.doneCh:
		return nil
	}
}

// NewActiveTaskTracker creates and returns a new ActiveTaskTracker instance.
// By default, it starts with zero active tasks and a background context
// that can be canceled when shutdown is initiated.
//
// Example usage:
//
//	tracker := NewActiveTaskTracker()
//
//	// Start a new task
//	if err := tracker.StartTask(); err == nil {
//		go func() {
//			// Do some work...
//			defer tracker.FinishTask()
//		}()
//	}
//
//	// Once you decide to stop the application:
//	tracker.Shutdown() // Blocks until all tasks have finished.
func NewActiveTaskTracker() *ActiveTaskTracker {
	att := &ActiveTaskTracker{
		doneCh: make(chan bool),
	}
	att.ctx, att.cancelFunc = context.WithCancel(context.Background())
	return att
}
