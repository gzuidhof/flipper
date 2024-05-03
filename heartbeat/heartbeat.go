package heartbeat

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gzuidhof/flipper/config/cfgmodel"
)

// Heartbeat can be used to send a request to a URL at a regular interval.
type Heartbeat struct {
	cfg    cfgmodel.HeartbeatConfig
	logger *slog.Logger
}

// New creates a new heartbeat with the given configuration and logger.
func New(cfg cfgmodel.HeartbeatConfig, logger *slog.Logger) *Heartbeat {
	return &Heartbeat{cfg: cfg, logger: logger}
}

func (h *Heartbeat) do(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, h.cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.cfg.URL, nil)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to create heartbeat request.", slog.String("error", err.Error()))
		return
	}

	h.logger.DebugContext(ctx, "Sending heartbeat.", slog.String("url", h.cfg.URL))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to send heartbeat.", slog.String("error", err.Error()))
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.ErrorContext(ctx, "Heartbeat returned non-200 status.", slog.Int("status_code", resp.StatusCode))
	}

	closeErr := resp.Body.Close()
	if closeErr != nil {
		h.logger.ErrorContext(ctx, "Failed to close heartbeat response body.", slog.String("error", closeErr.Error()))
	}
}

// Start the heartbeat. It will send a request to the configured URL every interval.
// This function will block until the context is cancelled.
func (h *Heartbeat) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(h.cfg.Interval)
	defer ticker.Stop()

	// Send a heartbeat right away.
	h.do(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.do(ctx)
		}
	}
}
