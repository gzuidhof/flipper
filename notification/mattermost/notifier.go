package mattermost

import (
	"context"
	"log/slog"

	"github.com/gzuidhof/flipper/config/cfgmodel"
)

// Notifier sends notifications to a Mattermost webhook.
type Notifier struct {
	cfg    cfgmodel.NotificationTargetConfig
	logger *slog.Logger
	client *Client
}

// NewNotifier creates a new MattermostNotifier.
func NewNotifier(
	cfg cfgmodel.NotificationTargetConfig,
	logger *slog.Logger,
	client *Client,
) *Notifier {
	logger = logger.With("type", cfg.Type, "url", cfg.URL)

	return &Notifier{
		client: client,
		logger: logger,
		cfg:    cfg,
	}
}

// Notify sends a notification to the Mattermost webhook.
func (n *Notifier) Notify(ctx context.Context, message string) error {
	n.logger.DebugContext(ctx, "Sending Mattermost notification")

	msg := Message{
		Text:      message,
		Channel:   n.cfg.Channel,
		Username:  n.cfg.Username,
		IconEmoji: n.cfg.IconEmojiOrDefault(),
	}

	err := n.client.Send(ctx, msg)
	if err != nil {
		n.logger.ErrorContext(ctx, "Failed to send Mattermost notification.", slog.String("error", err.Error()))
		return err
	}

	return nil
}
