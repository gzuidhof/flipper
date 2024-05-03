package checker

import (
	"context"
	"time"
)

// State is the state of a health check.
type State string

// Possible states of a health check.
const (
	// StateUnknown is the initial state of a health check.
	StateUnknown State = "unknown"
	// StateHealthy means past health checks indicate the service should be considered healthy.
	StateHealthy State = "healthy"
	// StateUnhealthy means past health checks indicate the service should be considered unhealthy.
	StateUnhealthy State = "unhealthy"
)

// Stateful is a health checker that runs periodically and carries internal state between checks.
type Stateful[ResultType Result] struct {
	checker *periodic[ResultType]

	// fallCount is the number of successive failures.
	fallCount uint64
	// riseCount is the number of successive successes.
	riseCount uint64

	id            string
	riseThreshold uint64
	fallThreshold uint64

	currentState State
}

// StatefulUpdate is an update from stateful health checker.
type StatefulUpdate[ResultType Result] struct {
	ResultType ResultType

	// Timestamp is the time when the check was started.
	Timestamp time.Time
	// Duration is the time it took to run the check.
	Duration time.Duration

	// State is the state of the health check.
	State State

	// ID is the check ID that the update is for.
	ID string

	// Rise is the number of consecutive health checks that were successful.
	Rise uint64
	// Fall is the number of consecutive health checks that failed.
	Fall uint64
}

// newStateful creates a new stateful health check.
func newStateful[ResultType Result](checker Check[ResultType]) *Stateful[ResultType] {
	cfg := checker.Config()
	return &Stateful[ResultType]{
		checker:       newPeriodic(checker),
		riseThreshold: cfg.RiseOrDefault(),
		fallThreshold: cfg.FallOrDefault(),
		id:            cfg.ID,
		currentState:  StateUnknown,
	}
}

// updateCheckerState updates the state of the health check based on the number of successive successes and failures.
func (c *Stateful[ResultType]) updateCheckerState() State {
	if c.riseCount >= c.riseThreshold {
		c.currentState = StateHealthy
		return c.currentState
	} else if c.fallCount >= c.fallThreshold {
		c.currentState = StateUnhealthy
		return c.currentState
	}
	return c.currentState
}

func (c *Stateful[ResultType]) updateRiseAndFall(success bool) {
	if success {
		c.riseCount++
		c.fallCount = 0
	} else {
		c.fallCount++
		c.riseCount = 0
	}
}

// Start the stateful health check. It will run the check every interval until the context is cancelled.
// The index can be used to identify this check in a multi-check setup (which is the most common use case for this).
// The caller is responsible for closing the update channel.
func (c *Stateful[ResultType]) Start(ctx context.Context, updateChan chan<- StatefulUpdate[ResultType]) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	periodicUpdateChan := make(chan periodicUpdate[ResultType], 16)
	go c.checker.Start(ctx, periodicUpdateChan)

	for {
		select {
		case <-ctx.Done():
			return
		case periodicUpdate := <-periodicUpdateChan:
			c.updateRiseAndFall(periodicUpdate.Result.Healthy())
			newState := c.updateCheckerState()

			update := StatefulUpdate[ResultType]{
				ResultType: periodicUpdate.Result,
				Timestamp:  periodicUpdate.Timestamp,
				Duration:   periodicUpdate.Duration,
				State:      newState,
				ID:         c.id,

				Rise: c.riseCount,
				Fall: c.fallCount,
			}

			if ctx.Err() != nil {
				return
			}
			updateChan <- update
		}
	}
}
