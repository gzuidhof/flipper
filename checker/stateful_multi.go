package checker

import (
	"context"
)

// StatefulMulti is a health checker that combines multiple stateful health checks.
// These checks can run in parallel, even at different intervals.
// The state of the multi check is the worst state of all the checks.
type StatefulMulti[ResultType Result] struct {
	checks []*Stateful[ResultType]

	// checkStates stores the latest update of each check.
	// It's a map from the ID of the check to the last state of every check.
	lastUpdates map[string]StatefulUpdate[ResultType]
}

// StatefulMultiUpdate is an update from a stateful multi check.
type StatefulMultiUpdate[ResultType Result] struct {
	// HealthState is the current state of the multi check.
	// This will be the worst state of all the checks.
	HealthState State

	// lastUpdatedID is the index of the last check that was updated.
	lastUpdatedID string

	// Updates is a map from the ID of the check to the last update of every check.
	updates map[string]StatefulUpdate[ResultType]
}

// LastUpdate returns the last update of the multi check.
func (u StatefulMultiUpdate[ResultType]) LastUpdate() StatefulUpdate[ResultType] {
	return u.updates[u.lastUpdatedID]
}

// UnhealthyChecks returns the current unhealthy checks.
func (u StatefulMultiUpdate[ResultType]) UnhealthyChecks() []ResultType {
	unhealthyChecks := make([]ResultType, 0)
	for _, update := range u.updates {
		if update.State == StateUnhealthy {
			unhealthyChecks = append(unhealthyChecks, update.ResultType)
		}
	}
	return unhealthyChecks
}

// NewStatefulMultiCheck creates a new stateful multi check.
func NewStatefulMultiCheck[ResultType Result](
	checks []Check[ResultType],
) *StatefulMulti[ResultType] {
	statefulChecks := make([]*Stateful[ResultType], len(checks))
	for i, check := range checks {
		statefulChecks[i] = newStateful(check)
	}

	return &StatefulMulti[ResultType]{
		checks:      statefulChecks,
		lastUpdates: make(map[string]StatefulUpdate[ResultType]),
	}
}

func (c *StatefulMulti[ResultType]) worstState() State {
	worstState := StateHealthy
	for _, update := range c.lastUpdates {
		if update.State == StateUnknown {
			worstState = StateUnknown
		}
		if update.State == StateUnhealthy {
			worstState = StateUnhealthy
			break
		}
	}
	return worstState
}

// Start the stateful multi check. It will run the checks until the context is cancelled.
// It will send updates to the update channel on every individual check update.
func (c *StatefulMulti[ResultType]) Start(
	ctx context.Context,
	updateChan chan<- StatefulMultiUpdate[ResultType],
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	recvUpdateChan := make(chan StatefulUpdate[ResultType], 16)
	defer close(recvUpdateChan)

	// Start the checks.
	for _, check := range c.checks {
		check := check
		go check.Start(ctx, recvUpdateChan)
	}

	// We receive updates from all the checks.
	for {
		select {
		case <-ctx.Done():
			return
		case recvUpdate := <-recvUpdateChan:
			// Update the last state of the check.
			c.lastUpdates[recvUpdate.ID] = recvUpdate

			// Send the update.
			update := StatefulMultiUpdate[ResultType]{
				HealthState:   c.worstState(),
				lastUpdatedID: recvUpdate.ID,
				updates:       c.lastUpdates,
			}
			if ctx.Err() != nil {
				return
			}
			updateChan <- update
		}
	}
}
