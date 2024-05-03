package entry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/gzuidhof/flipper/buildinfo"
	"github.com/gzuidhof/flipper/config"
	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/heartbeat"
	"github.com/gzuidhof/flipper/monitor"
	"github.com/gzuidhof/flipper/notification"
	"github.com/gzuidhof/flipper/server"
	"github.com/gzuidhof/flipper/telemetry"
	"github.com/gzuidhof/flipper/view/template"
	"golang.org/x/sync/errgroup"
)

func setupServer(cfg cfgmodel.ServerConfig, logger *slog.Logger) (*server.Server, error) {
	opts := []server.Option{
		server.WithAddr(net.JoinHostPort(cfg.Host, fmt.Sprint(cfg.Port))),
		server.WithShutdownTimeout(cfg.ShutdownTimeout),
		server.WithLogger(logger),
	}
	if cfg.Assets != "" {
		logger.Info("Using assets from disk.")
		opts = append(opts, server.WithStaticFS(os.DirFS(cfg.Assets)))
	}
	if cfg.Templates != "" {
		logger.Info("Using templates from disk.")
		opts = append(opts, server.WithTemplateEngine(template.New(os.DirFS(cfg.Templates))))
	}

	server, err := server.New(
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	return server, nil
}

// Monitor starts watching the resources and acts to keep the floating IPs pointed to healthy servers.
func Monitor(ctx context.Context, configFilepath string) error {
	cfg, cfgErr := config.Init(configFilepath)
	if cfgErr != nil {
		return fmt.Errorf("failed to initialize config: %w", cfgErr)
	}

	logger := telemetry.SetupLogger(cfg.Telemetry.Logging, os.Stdout)
	logger = logger.With(slog.String("version", buildinfo.Version()), slog.String("service", cfg.Service.Name))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	notifier, err := notification.NewNotifierFromConfig(cfg.Notifications, logger)
	if err != nil {
		return fmt.Errorf("failed to create notifier: %w", err)
	}

	grp, ctx := errgroup.WithContext(ctx)
	w, err := monitor.New(ctx, cfg, logger, notifier)
	if err != nil {
		return fmt.Errorf("failed to start monitors: %w", err)
	}
	grp.Go(func() error {
		if watchErr := w.Watch(ctx); watchErr != nil {
			cancel()
			return fmt.Errorf("monitor watch failed: %w", watchErr)
		}
		return nil
	})

	if cfg.Server.Enabled {
		server, setupErr := setupServer(cfg.Server, logger)
		if setupErr != nil {
			return fmt.Errorf("failed to create server: %w", setupErr)
		}

		grp.Go(func() error {
			if serveErr := server.ListenAndServe(ctx); serveErr != nil {
				if errors.Is(serveErr, context.Canceled) {
					return nil
				}
				return fmt.Errorf("server listen and serve failed: %w", serveErr)
			}

			return nil
		})
	}

	if cfg.Heartbeat.Enabled {
		h := heartbeat.New(cfg.Heartbeat, logger)
		grp.Go(func() error {
			logger.InfoContext(ctx, "Starting heartbeat.")
			h.Start(ctx)
			return nil
		})
	}

	//nolint:wrapcheck // Not desired to wrap the error here.
	return grp.Wait()
}
