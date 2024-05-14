package monitor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gzuidhof/flipper/buildinfo"
	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/notification"
	"github.com/gzuidhof/flipper/notification/notificationtemplate"
	"github.com/gzuidhof/flipper/plan"
	"github.com/gzuidhof/flipper/resource"
)

// Group monitors a group of resources and actions on them if necessary.
type Group struct {
	cfg      cfgmodel.GroupConfig
	logger   *slog.Logger
	provider resource.Provider
	notifier notification.Notifier

	watcher      *ResourcesWatcher
	healthkeeper *HealthKeeper
}

// NewGroup creates a new monitor group from a config and a provider.
func NewGroup(
	cfg cfgmodel.GroupConfig,
	provider resource.Provider,
	logger *slog.Logger,
	notifier notification.Notifier,
) *Group {
	logger = logger.With(
		slog.String("group", cfg.DisplayName),
		slog.String("group_id", cfg.ID),
		slog.String("provider", cfg.Provider),
	)

	watcher := NewResourcesWatcher(cfg, logger, provider)
	healthkeeper := NewHealthKeeper(cfg, logger, provider, notifier)

	return &Group{
		cfg:      cfg,
		watcher:  watcher,
		logger:   logger,
		notifier: notifier,

		healthkeeper: healthkeeper,
		provider:     provider,
	}
}

func (g *Group) executePlan(ctx context.Context, logger *slog.Logger, state plan.State, actionPlan plan.Plan) error {
	ctx, cancel := context.WithTimeout(ctx, g.cfg.PlanApplyTimeoutOrDefault())
	defer cancel()

	for _, action := range actionPlan.Actions {
		if ctx.Err() != nil {
			logger.ErrorContext(ctx, "Context cancelled, aborting plan.")
			return fmt.Errorf("plan cancelled because of context: %w", ctx.Err())
		}
		flip, ok := state.FloatingIPs[action.FloatingIPID]
		if !ok {
			logger.ErrorContext(ctx, "Floating IP not found, plan will be aborted.",
				slog.String("floating_ip_id", action.FloatingIPID),
			)
			return fmt.Errorf("floating IP not found for %s", action.FloatingIPID)
		}
		serverWithStatus, ok := state.Servers[action.ServerID]
		if !ok {
			logger.ErrorContext(ctx, "Server not found, plan will be aborted.",
				slog.String("server_id", action.ServerID),
			)
			return fmt.Errorf("server not found for %s", action.ServerID)
		}

		err := g.provider.AssignFloatingIP(ctx, flip, serverWithStatus.Resource)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to assign floating IP",
				slog.String("error", err.Error()),
				slog.String("floating_ip_id", flip.ID()),
				slog.String("server_id", serverWithStatus.Resource.ID()),
			)
			return fmt.Errorf("failed to assign floating IP: %w", err)
		}
		logger.InfoContext(ctx, "Floating IP assigned.",
			slog.String("floating_ip_id", flip.ID()),
			slog.String("server_id", serverWithStatus.Resource.ID()),
		)
	}
	return nil
}

// Start watching the resources and performing health checks.
// This function blocks until the context is cancelled.
func (g *Group) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msg := fmt.Sprintf(":eyes: Starting monitor for group **%s** (`%s`). Flipper version `%s`. ",
		g.cfg.DisplayName, g.cfg.ID, buildinfo.Version())

	if g.cfg.ReadOnly {
		msg += "\n:lock: **Read-only mode** enabled, no actions will be taken. Only unhealthy/healthy notifications" +
			" will be sent."
	}

	_ = g.notifier.Notify(ctx, msg)

	updateChan := make(chan ResourceUpdate, 16)
	errChan := make(chan error, 16)

	actionChan := make(chan HealthKeeperAction)
	defer close(actionChan)

	g.logger.DebugContext(ctx, "Starting resource watcher.")
	go g.watcher.Start(ctx, updateChan, errChan)
	go g.healthkeeper.Start(ctx, updateChan, actionChan)

	minSequence := uint64(0)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			if err == nil {
				// This should not happen, but it seems like it can happen during a graceful shutdown?!
				// Can anybody explain how this is possible?
				// g.logger.InfoContext(ctx, "Received nil error from resources watcher.")
				continue
			}
			// If this happens once or twice, it's not a big deal. It only means that if there were any changes
			// to the resources being watched (servers, floating IPs) they would not be picked up.
			g.logger.ErrorContext(ctx, "Error in resources watcher update.",
				slog.String("error", err.Error()),
			)
		case action := <-actionChan:
			// During the time that we apply the plan we might have received more updates that are no longer
			// relevant. We ignore them by checking the sequence number.
			if action.ResourcesSequence < minSequence {
				g.logger.DebugContext(ctx, "Ignoring stale action with sequence lower than min sequence.",
					slog.Uint64("sequence", action.ResourcesSequence),
					slog.Uint64("min_sequence", minSequence),
				)
				continue
			}

			logger := g.logger.With(
				slog.String("plan_id", action.Plan.ID.String()),
			)

			if g.cfg.ReadOnly {
				logger.InfoContext(ctx, "Read-only group, not executing plan.")
				continue
			}

			logger.InfoContext(ctx, "Executing plan.")
			_ = g.notifier.Notify(ctx,
				notificationtemplate.RenderPlanExecution(g.cfg, action.State, action.Plan),
			)

			err := g.executePlan(ctx, logger, action.State, action.Plan)
			if err != nil {
				_ = g.notifier.Notify(ctx,
					fmt.Sprintf("💥 Failed to execute plan `%s` for group **%s** (`%s`).\n",
						action.Plan.ID,
						g.cfg.DisplayName,
						g.cfg.ID,
					)+
						fmt.Sprintf("Error: `%s`\n", err.Error()),
				)
				logger.ErrorContext(ctx, "Failed to execute plan.",
					slog.String("error", err.Error()),
				)
			} else {
				msg := fmt.Sprintf("🚀 **Plan** `%s` **executed successfully** for group **%s** (`%s`).",
					action.Plan.ID,
					g.cfg.DisplayName,
					g.cfg.ID,
				)
				numUnhealthy := action.State.NumUnhealthyServers()

				if numUnhealthy > 0 {
					msg += fmt.Sprintf("\n**Note there are still %d _unhealthy_ servers 🔥.**", numUnhealthy)
				} else {
					msg += "\n No servers are **_unhealthy_**."
				}

				_ = g.notifier.Notify(ctx, msg)
				logger.InfoContext(ctx, "Plan executed successfully.",
					slog.Int("num_unhealthy_servers", numUnhealthy))
			}

			minSequence = g.watcher.performUpdate(ctx, updateChan, errChan)
		}
	}
}
