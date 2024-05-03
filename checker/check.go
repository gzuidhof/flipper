package checker

import (
	"context"

	"github.com/gzuidhof/flipper/config/cfgmodel"
)

// Result is the result of a health check.
type Result interface {
	// Healthy returns true if the check is healthy.
	Healthy() bool
}

// Check is a common interface for health checks.
type Check[ResultType Result] interface {
	// Check runs the health check and returns the result.
	Check(ctx context.Context) ResultType

	// Config returns the configuration of the health check.
	Config() cfgmodel.HealthCheckConfig
}
