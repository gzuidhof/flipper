// Package server implements the HTTP server for flipper.
package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gzuidhof/flipper/view/static"
	"github.com/gzuidhof/flipper/view/template"
)

// Server is a simple HTTP server.
type Server struct {
	addr            string
	shutdownTimeout time.Duration
	logger          *slog.Logger

	mux *http.ServeMux

	staticFS fs.FS
	template *template.Engine
}

// Option is a functional option for the server.
type Option func(s *Server) error

// New creates a new server with the given options.
func New(opts ...Option) (*Server, error) {
	s := &Server{
		shutdownTimeout: 5 * time.Second,
		mux:             http.NewServeMux(),
		logger:          slog.Default(),
		staticFS:        static.EmbeddedFS(),
		template:        template.NewEmbedded(),
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	s.registerRoutes()

	return s, nil
}

// WithAddr sets the address the server listens on.
func WithAddr(addr string) Option {
	return func(s *Server) error {
		s.addr = addr
		return nil
	}
}

// WithStaticFS sets the static file system for the server.
func WithStaticFS(staticFS fs.FS) Option {
	return func(s *Server) error {
		s.staticFS = staticFS
		return nil
	}
}

// WithLogger sets the logger for the server.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) error {
		s.logger = logger
		return nil
	}
}

// WithShutdownTimeout sets the shutdown timeout for the server.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) error {
		s.shutdownTimeout = timeout
		return nil
	}
}

// WithTemplateEngine sets the template engine for the server.
func WithTemplateEngine(t *template.Engine) Option {
	return func(s *Server) error {
		s.template = t
		return nil
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe listens on the configured address and serves requests.
// When the context is cancelled, it attempts to gracefully shut down.
func (s *Server) ListenAndServe(ctx context.Context) error {
	server := &http.Server{
		Addr:              s.addr,
		Handler:           s,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       120 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	serveErr := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	s.logger.InfoContext(ctx, "Server listening.", slog.String("addr", s.addr))

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		s.logger.DebugContext(ctx, "Gracefully shutting down server.")

		//nolint:contextcheck // Using a non-inherited context is intentional.
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
		return nil
	case err := <-serveErr:
		return err
	}
}
