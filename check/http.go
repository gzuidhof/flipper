package check

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gzuidhof/flipper/config/cfgmodel"
)

// getRetargetHTTPClient returns a new HTTP client that will connect to the target IP address
// regardless of the address in the URL. This is useful for health checks, where we want to
// connect to a specific IP address but still use the original address in the Host value of the
// HTTP request - as well as any checks that the Go http client does behind the scenes.
func getRetargetHTTPClient(target string, port string, timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	transport := &http.Transport{
		// We need to rewrite the address to the target IP, because the HTTP client will use the original address
		// in the Host value of the HTTP request. This is important for HTTPS checks, because the Host value must
		// match the certificate - but we want to connect to a specific IP address.
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			rewrittenAddr := net.JoinHostPort(target, port)
			return dialer.DialContext(ctx, network, rewrittenAddr)
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// HTTPCheck checks the health of a resource over HTTP or HTTPS.
type HTTPCheck struct {
	cfg    cfgmodel.HealthCheckConfig
	target string

	client *http.Client
}

// NewHTTPCheck creates a new HTTP or HTTPS health check from a config.
// The target is generally the IP address of the resource being checked, although it could also be a hostname.
func NewHTTPCheck(cfg cfgmodel.HealthCheckConfig, target string) *HTTPCheck {
	return &HTTPCheck{
		cfg:    cfg,
		target: target,
		client: getRetargetHTTPClient(target, fmt.Sprintf("%d", cfg.PortOrDefault()), cfg.TimeoutOrDefault()),
	}
}

func (h *HTTPCheck) hostValue() string {
	if h.cfg.Host != "" {
		return h.cfg.Host
	}
	return h.target
}

// Check the health of a resource at the given IP address by performing a HTTP request.
func (h *HTTPCheck) Check(ctx context.Context) HTTPCheckResult {
	result := HTTPCheckResult{}

	if h.cfg.Type != "http" && h.cfg.Type != "https" {
		return result.Errorf("unsupported health check type: %s", h.cfg.Type)
	}

	ctx, cancel := context.WithTimeout(ctx, h.cfg.TimeoutOrDefault())
	defer cancel()

	url := url.URL{
		Scheme: h.cfg.Type,
		Host:   net.JoinHostPort(h.hostValue(), fmt.Sprintf("%d", h.cfg.PortOrDefault())),
		Path:   h.cfg.Path,
	}

	req, err := http.NewRequestWithContext(ctx, h.cfg.MethodOrDefault(), url.String(), nil)
	if err != nil {
		return result.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return result.Errorf("failed to perform request: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			slog.ErrorContext(ctx, "failed to close response body", slog.String("error", closeErr.Error()))
		}
	}()

	result.StatusCode = resp.StatusCode

	tlsResult := TLSCertificatesResultFromConnectionState(resp.TLS)
	result.TLS = tlsResult
	if tlsResult.IsTLS {
		if tlsResult.ExpiresWithin(time.Hour * 24 * 14) { // Warn for certificates that expire within 14 days.
			logFunc := slog.WarnContext
			if tlsResult.ExpiresWithin(time.Hour * 24 * 7) { // Error if it's within 7 days.
				logFunc = slog.ErrorContext
			}

			logFunc(ctx, "TLS certificate expires soon",
				slog.String("host", h.hostValue()),
				slog.String("target", h.target),
				slog.String("dns_names", fmt.Sprintf("%v", tlsResult.DNSNames)),
				slog.Time("expires_at", tlsResult.NotAfter),
			)
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return result
}

// Config returns the health check configuration.
func (h *HTTPCheck) Config() cfgmodel.HealthCheckConfig {
	return h.cfg
}
