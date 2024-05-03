package checker

import (
	"context"
	"time"
)

type periodic[T Result] struct {
	checker Check[T]
}

type periodicUpdate[T Result] struct {
	Result T

	// Timestamp is the time when the check was started.
	Timestamp time.Time
	// Duration is the time it took to run the check.
	Duration time.Duration
}

// newPeriodic creates a new periodic checker.
func newPeriodic[T Result](checker Check[T]) *periodic[T] {
	return &periodic[T]{
		checker: checker,
	}
}

// Check the health of the resource once.
func (c *periodic[T]) Check(ctx context.Context) periodicUpdate[T] {
	t := time.Now()
	result := c.checker.Check(ctx)
	return periodicUpdate[T]{
		Result:    result,
		Timestamp: t,
		Duration:  time.Since(t),
	}
}

// Start the periodic checker. It will run the check every interval until the context is cancelled.
// It will send updates to the update channel.
func (c *periodic[T]) Start(ctx context.Context, updateChan chan<- periodicUpdate[T]) {
	ticker := time.NewTicker(c.checker.Config().IntervalOrDefault())
	defer ticker.Stop()
	defer close(updateChan)

	// Run the first check immediately.
	// Note: Will this ever be a problem if we have too many checks? I guess we don't have to worry about that for now.
	update := c.Check(ctx)
	updateChan <- update

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			update = c.Check(ctx)
			updateChan <- update
		}
	}
}
