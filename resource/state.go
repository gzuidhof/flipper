package resource

import (
	"sort"
	"sync/atomic"
	"time"
)

// State is the health state of a resource.
type State struct {
	// LastUpdated is the time when the status was last updated.
	LastUpdated time.Time

	// Status is the status code of the resource, e.g. healthy, unhealthy, unknown.
	Status Status
}

// Status is the status code of a resource.
type Status string

const (
	// StatusHealthy is used when the resource is known to be healthy.
	StatusHealthy Status = "healthy"
	// StatusUnhealthy is used when the resource is known to be unhealthy.
	StatusUnhealthy Status = "unhealthy"
	// StatusUnknown is used when the resource status is not known, e.g. when the resource is not yet checked.
	StatusUnknown Status = "unknown"
)

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// WithStatusSlice is a list of resources with associated statuses.
type WithStatusSlice[R Resource] []*WithStatus[R]

// WithStatus is a resource with an associated status, which can be updated atomically.
type WithStatus[R Resource] struct {
	Resource     R
	atomicStatus atomic.Pointer[State]
}

// NewWithStatus creates a new server with a given (initial) status.
func NewWithStatus[R Resource](s R, status State) *WithStatus[R] {
	ss := &WithStatus[R]{
		Resource: s,
	}

	ss.atomicStatus.Store(&status)
	return ss
}

// SetState sets the state of the resource and returns the previous status.
func (s *WithStatus[R]) SetState(state State) State {
	prev := s.atomicStatus.Swap(&state)
	if prev == nil {
		return State{Status: StatusUnknown}
	}
	return *prev
}

// State returns the current health state of the resource.
func (s *WithStatus[R]) State() State {
	v := s.atomicStatus.Load()
	if v == nil {
		return State{Status: StatusUnknown}
	}
	return *v
}

// Status returns the current status code of the resource.
func (s *WithStatus[R]) Status() Status {
	return s.State().Status
}

// IsUnhealthy returns true if the resource is known to be unhealthy.
func (s *WithStatus[R]) IsUnhealthy() bool {
	return s.Status() == StatusUnhealthy
}

// IsHealthy returns true if the resource is known to be healthy.
func (s *WithStatus[R]) IsHealthy() bool {
	return s.Status() == StatusHealthy
}

// SortByName sorts the resources by name.
func (servs WithStatusSlice[R]) SortByName() {
	sort.Slice(servs, func(i, j int) bool {
		return servs[i].Resource.Name() < servs[j].Resource.Name()
	})
}
