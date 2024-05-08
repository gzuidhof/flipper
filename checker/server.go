package checker

import (
	"context"
	"time"

	"github.com/gzuidhof/flipper/check"
	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/resource"
)

// Server checks the health of a server's public interfaces.
type Server struct {
	cfgs   []cfgmodel.HealthCheckConfig
	server *resource.WithStatus[resource.Server]

	multichecker *StatefulMulti[check.HTTPCheckResult]
}

// ServerCheckUpdate is an update to a server's health check.
type ServerCheckUpdate struct {
	PreviousServerState resource.State
	ServerState         resource.State
	ServerStateChanged  bool

	// The server that was checked.
	Server *resource.Server

	// The result of the check.
	Result StatefulMultiUpdate[check.HTTPCheckResult]
}

// UnhealthyChecks returns the current unhealthy checks.
func (u ServerCheckUpdate) UnhealthyChecks() []check.HTTPCheckResult {
	return u.Result.UnhealthyChecks()
}

// NewServerChecker creates a new server checker.
func NewServerChecker(
	cfgs []cfgmodel.HealthCheckConfig,
	serverWithStatus *resource.WithStatus[resource.Server],
) *Server {
	checker := &Server{
		cfgs:   cfgs,
		server: serverWithStatus,
	}

	server := serverWithStatus.Resource
	checks := make([]Check[check.HTTPCheckResult], 0)

	for _, c := range cfgs {
		if server.PublicIPv4.IsValid() && c.IPVersion != "ipv6" {
			ipv4c := c
			ipv4c.ID += "__ipv4"
			checks = append(checks, check.NewHTTPCheck(ipv4c, server.PublicIPv4.String()))
		}
		if server.PublicIPv6.IsValid() && c.IPVersion != "ipv4" {
			ipv6c := c
			ipv6c.ID += "__ipv6"
			checks = append(checks, check.NewHTTPCheck(c, server.PublicIPv6.String()))
		}
	}

	checker.multichecker = NewStatefulMultiCheck(checks)
	return checker
}

// Start periodic checks of the server's health, changing the server's status accordingly.
// This function blocks until the context is cancelled.
// The caller is responsible for closing the onUpdate channel.
func (c *Server) Start(ctx context.Context, onUpdate chan<- ServerCheckUpdate) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	updateChan := make(chan StatefulMultiUpdate[check.HTTPCheckResult], 16)
	defer close(updateChan)

	lastState := resource.State{
		Status: resource.StatusUnknown,
	}

	go c.multichecker.Start(ctx, updateChan)

	for {
		select {
		case <-ctx.Done():
			return
		case recvUpdate := <-updateChan:
			state := resource.State{
				LastUpdated: time.Now(),
				Status:      resource.StatusUnhealthy,
			}
			if recvUpdate.HealthState == StateHealthy {
				state.Status = resource.StatusHealthy
			} else if recvUpdate.HealthState == StateUnknown {
				state.Status = resource.StatusUnknown
			}

			update := ServerCheckUpdate{
				ServerState:         state,
				PreviousServerState: lastState,
				ServerStateChanged:  state.Status != lastState.Status,
				Server:              &c.server.Resource,
				Result:              recvUpdate,
			}

			if update.ServerStateChanged {
				lastState = state
				c.server.SetState(state)
			}

			onUpdate <- update
		}
	}
}
