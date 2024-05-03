package monitor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gzuidhof/flipper/checker"
	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/notification"
	"github.com/gzuidhof/flipper/notification/notificationtemplate"
	"github.com/gzuidhof/flipper/plan"
	"github.com/gzuidhof/flipper/resource"
)

// HealthKeeper keeps track of the health of the resources in a group and generates actions to keep them healthy.
// It is responsible for monitoring the health of the resources.
type HealthKeeper struct {
	cfg      cfgmodel.GroupConfig
	provider resource.Provider
	logger   *slog.Logger
	notifier notification.Notifier

	didReceiveInitialResourcesUpdate bool

	serverWatcherCancel map[string]context.CancelFunc

	state             plan.State
	resourcesSequence uint64
}

// HealthKeeperAction is an action that the healthkeeper requests to be executed.
type HealthKeeperAction struct {
	// ResourcesSequence is the sequence number of the resources when the action was generated.
	ResourcesSequence uint64
	State             plan.State
	Plan              plan.Plan
}

// NewHealthKeeper creates a new healthkeeper for a group.
func NewHealthKeeper(
	cfg cfgmodel.GroupConfig,
	logger *slog.Logger,
	provider resource.Provider,
	notifier notification.Notifier,
) *HealthKeeper {
	return &HealthKeeper{
		cfg:      cfg,
		provider: provider,
		logger:   logger,
		notifier: notifier,

		serverWatcherCancel: make(map[string]context.CancelFunc),
		state:               plan.NewStateFromGroup(resource.Group{}),
	}
}

func (h *HealthKeeper) updateResourceChanges(
	ctx context.Context,
	rup ResourceUpdate,
	recvUpdateChan chan<- checker.ServerCheckUpdate,
) {
	h.resourcesSequence = rup.Sequence
	changeset := rup.Changeset

	// A guard: I don't think we ever call this function with an empty changeset, but just in case.
	if changeset.Empty() {
		h.logger.DebugContext(ctx, "No changes.")
		return
	}

	// We notify when there were substantial changes to the resources, e.g. servers or floating IPs being
	// added or removed. That shouldn't really happen unexpectedly.
	//
	// We expect this to fire once when the healthkeeper starts up: then let's not notify.
	if h.didReceiveInitialResourcesUpdate && !changeset.IsUpdatesOnly() {
		msg := fmt.Sprintf("⚡️ Resources in group **%s** (`%s`) being watched changed substantially changed.\n",
			h.cfg.DisplayName, h.cfg.ID,
		) + fmt.Sprintf("```\n%s\n```", changeset) // TODO: prettyprint this through some template.
		_ = h.notifier.Notify(ctx, msg)
	}
	h.didReceiveInitialResourcesUpdate = true

	startServerChecker := func(ctx context.Context, server resource.Server) {
		serverWithStatus := resource.NewWithStatus(server, resource.State{Status: resource.StatusUnknown})
		serverChecker := checker.NewServerChecker(h.cfg.Checks, serverWithStatus)
		ctx, cancel := context.WithCancel(ctx)
		h.serverWatcherCancel[server.ID()] = cancel
		h.state.Servers[server.ID()] = serverWithStatus
		go serverChecker.Start(ctx, recvUpdateChan)
	}

	for _, fip := range changeset.FloatingIPs.Added {
		h.state.FloatingIPs[fip.ID()] = fip
	}
	for _, fip := range changeset.FloatingIPs.Updated {
		h.state.FloatingIPs[fip.ID()] = fip
	}
	for _, fip := range changeset.FloatingIPs.Removed {
		delete(h.state.FloatingIPs, fip.ID())
	}
	for _, server := range changeset.Servers.Added {
		startServerChecker(ctx, server)
	}
	for _, server := range changeset.Servers.Updated {
		cancelServerWatcher, ok := h.serverWatcherCancel[server.ID()]
		if ok {
			cancelServerWatcher()
		}
		delete(h.state.Servers, server.ID())
		startServerChecker(ctx, server)
	}
	for _, server := range changeset.Servers.Removed {
		delete(h.state.Servers, server.ID())
	}
}

// Start monitoring the given resources, stopping anything it was watching before.
// This function blocks until the context is cancelled.
// It will send actions to the actionChan when it detects that an action is required to keep the resources healthy.
// The caller is responsible for closing the actionChan.
func (h *HealthKeeper) Start(
	ctx context.Context,
	resourceChanges <-chan ResourceUpdate,
	actionChan chan<- HealthKeeperAction,
) {
	// We use a single channel to receive updates from all the server checkers.
	serverCheckUpdateChan := make(chan checker.ServerCheckUpdate, 32)

	for {
		select {
		case <-ctx.Done():
			return
		case resourcesUpdate := <-resourceChanges:
			h.logger.InfoContext(ctx, "Resources changed, updating healthkeeper state.")
			h.updateResourceChanges(ctx, resourcesUpdate, serverCheckUpdateChan)
		case update := <-serverCheckUpdateChan:
			// The actual most recent check's update that triggered this event.
			mostRecentUpdate := update.Result.LastUpdate()
			serverID := update.Server.ID()

			if h.state.Servers[serverID] == nil {
				// Unlikely, but it's possible there could be updates buffered that have since been removed.
				h.logger.WarnContext(ctx, "Server not found in state, skipping update.",
					slog.String("server_id", serverID),
				)
				continue
			}

			logger := h.logger.With(
				slog.String("server_id", serverID),
				slog.String("server_name", update.Server.Name()),
				slog.String("check_id", mostRecentUpdate.ID),
				slog.String("server_state", string(update.ServerState.Status)),
			)

			logger.DebugContext(ctx, "Server health check completed.",
				slog.Uint64("rise", mostRecentUpdate.Rise),
				slog.Uint64("fall", mostRecentUpdate.Fall),
				slog.Duration("duration", mostRecentUpdate.Duration),
			)

			if update.ServerStateChanged {
				logger.InfoContext(ctx, "Server state changed.")
				if update.ServerState.Status == resource.StatusUnhealthy {
					// TODO improve the formatting here.. This is a bit of a mess.
					// We should use a template for this (and the below healthy message).
					_ = h.notifier.Notify(ctx,
						fmt.Sprintf(":fire: Server **`%s`** (%s) in location `%s` became **_unhealthy_**.\n",
							update.Server.Name(),
							update.Server.ID(),
							update.Server.Location,
						)+
							fmt.Sprintf("```\n%+v\n```\n", update.UnhealthyChecks())+notificationtemplate.RenderState(h.cfg, h.state),
					)

					h.logger.ErrorContext(ctx, "Server became unhealthy.",
						slog.String("unhealthy_checks", fmt.Sprintf("%+v", update.UnhealthyChecks())), // TODO: improve.
					)
				}

				if update.ServerState.Status == resource.StatusHealthy &&
					update.PreviousServerState.Status != resource.StatusUnknown {
					_ = h.notifier.Notify(ctx,
						fmt.Sprintf(
							"✅ Server **`%s`** (%s) in location `%s` became **_healthy_** again.\n",
							update.Server.Name(),
							update.Server.ID(),
							update.Server.Location,
						)+notificationtemplate.RenderState(h.cfg, h.state),
					)
				}
			}

			actionPlan := plan.New(h.state)
			if actionPlan.Empty() {
				h.logger.DebugContext(ctx, "No actions required.")
				continue
			}

			if !h.cfg.PlanApplyWithUnkownStatus && h.state.HasServersWithUnknownStatus() {
				// It's best if we wait until all servers have a known status before considering applying any plan.
				// Especially during startup, when would end up creating a lot of unnecessary actions.
				h.logger.DebugContext(ctx, "There are still servers with unknown status, deferring plan.")
				continue
			}

			h.logger.ErrorContext(ctx, "Actions required.",
				slog.String("plan", actionPlan.String()),
				slog.String("plan_id", actionPlan.ID.String()),
				slog.String("unhealthy_floating_ips", fmt.Sprintf("%+v", h.state.UnhealthyFloatingIPs())),
			)

			actionChan <- HealthKeeperAction{
				State:             h.state,
				Plan:              actionPlan,
				ResourcesSequence: h.resourcesSequence,
			}
		}
	}
}
