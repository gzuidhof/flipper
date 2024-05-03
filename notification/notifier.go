package notification

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/notification/mattermost"
)

// Notifier is the interface that wraps the Notify method.
type Notifier interface {
	Notify(ctx context.Context, message string) error
}

// NoopNotifier is a Notifier that does nothing.
type NoopNotifier struct{}

// Notify does nothing on the NoopNotifier.
func (n *NoopNotifier) Notify(_ context.Context, _ string) error {
	return nil
}

// NewNotifierFromConfig creates a new Notifier from the given NotificationsConfig.
//
//nolint:ireturn // This is a factory function, it's okay to return an interface.
func NewNotifierFromConfig(cfg cfgmodel.NotificationsConfig, logger *slog.Logger) (Notifier, error) {
	if !cfg.Enabled {
		return &NoopNotifier{}, nil
	}
	if len(cfg.Targets) == 0 {
		return nil, fmt.Errorf("no notification targets configured, but notifications are enabled")
	} else if len(cfg.Targets) > 1 {
		// TODO: Create a fan-out notifier that sends to multiple targets.
		return nil, fmt.Errorf("multiple notification targets are not supported yet")
	}

	targetCfg := cfg.Targets[0]
	if targetCfg.Type != "mattermost" {
		return nil, fmt.Errorf("unsupported notification target type: %s", targetCfg.Type)
	}

	client := mattermost.New(targetCfg.URL)
	return mattermost.NewNotifier(targetCfg, logger, client), nil
}
