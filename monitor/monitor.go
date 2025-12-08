package monitor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/notification"
	"github.com/gzuidhof/flipper/provider/hetzner"
	"github.com/gzuidhof/flipper/provider/mock"
	"github.com/gzuidhof/flipper/resource"
	"golang.org/x/sync/errgroup"
)

//nolint:ireturn,nolintlint // This is a factory function.
func buildProvider(ctx context.Context, group cfgmodel.GroupConfig) (resource.Provider, error) {
	switch group.Provider {
	case string(resource.ProviderNameHetzner):
		provider, err := hetzner.NewProvider(ctx, group)
		if err != nil {
			return nil, fmt.Errorf("failed to create hetzner provider: %w", err)
		}
		return provider, nil
	case string(resource.ProviderNameMock):
		if group.Mock == nil {
			return nil, fmt.Errorf("mock config is required")
		}
		provider, err := mock.NewProviderFromConfig(*group.Mock)
		if err != nil {
			return nil, fmt.Errorf("failed to create mock provider: %w", err)
		}
		return provider, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", group.Provider)
	}
}

// Monitor watches resources. It supports watching multiple groups of resources in parallel.
type Monitor struct {
	didStart bool
	groups   []*Group
}

// New creates a new monitor from a config.
func New(
	ctx context.Context,
	cfg *cfgmodel.Config,
	logger *slog.Logger,
	notifier notification.Notifier,
) (*Monitor, error) {
	groups := make([]*Group, len(cfg.Groups))

	if len(cfg.Groups) == 0 {
		return nil, fmt.Errorf("no groups to monitor, check your configuration")
	}

	for i, group := range cfg.Groups {
		provider, err := buildProvider(ctx, group)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider: %w", err)
		}

		groups[i] = NewGroup(group, provider, logger, notifier)
	}

	return &Monitor{
		groups: groups,
	}, nil
}

// Watch starts monitoring the resources. To gracefully stop watching, cancel the context.
func (w *Monitor) Watch(ctx context.Context) error {
	if w.didStart {
		return fmt.Errorf("watcher already started")
	}
	w.didStart = true

	errgrp, ctx := errgroup.WithContext(ctx)

	for _, group := range w.groups {
		group := group
		errgrp.Go(func() error {
			return group.Start(ctx)
		})
	}

	//nolint:wrapcheck // we wrap the error in the subroutines.
	return errgrp.Wait()
}
