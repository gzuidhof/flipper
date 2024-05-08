package cfgmodel

import (
	"math"
	"net/http"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// HealthCheckConfig describes a single health check.
type HealthCheckConfig struct {
	ID string `koanf:"id"`

	DisplayName string `koanf:"display_name"`

	// Type of check, currently only "http" or "https" is supported.
	Type string `koanf:"type"`

	// Interval for the check. Must be a `time.Duration` string like "5s" or "1m".
	// Defaults to 1 minute.
	Interval time.Duration `koanf:"interval"`
	// Timeout for waiting for a successful answer. Must be a `time.Duration` string like "5s" or "1m".
	// Defaults to 10 seconds.
	Timeout time.Duration `koanf:"timeout"`

	// Fall is the number of consecutive failures required to mark the check as down.
	// This is useful to avoid flapping. Defaults to 1.
	Fall uint64 `koanf:"fall"`

	// Rise is the number of consecutive successes required to mark the check as up.
	// This is useful to avoid flapping. Defaults to 1.
	Rise uint64 `koanf:"rise"`

	// Method is the HTTP method to use for the check. Defaults to "GET".
	Method string `koanf:"method"`

	// Host is the value of the host header to set. If empty, IP address is used.
	// For HTTPS checks, this is used for SNI: the host must match the certificate.
	Host string `koanf:"host"`

	// Port is the port to check. Must be between 1 and 65535.
	// Defaults to 80 for HTTP and 443 for HTTPS.
	Port int `koanf:"port"`

	// Path is the URL path to check. Should start with a forward slash "/".
	Path string `koanf:"path"`

	// IPVersion is the IP version to use for the check. Must be either "ipv4", "ipv6" or "both".
	// Defaults to "both".
	IPVersion string `koanf:"ip_version"`

	// TODO: Add support for headers.
	// TODO: Add support for body.
	// TODO: Add an `Expectations` field for checking for specific response body, status, etc.
}

// PortOrDefault returns the port or the default port if not set.
func (h HealthCheckConfig) PortOrDefault() uint {
	if h.Port == 0 {
		if h.Type == "https" {
			return 443
		}
		return 80
	}
	if h.Port < 0 { // Impossible with the validation.. but just in case.
		panic("check.port must be positive")
	}
	return uint(h.Port)
}

// MethodOrDefault returns the method or the default method if not set.
func (h HealthCheckConfig) MethodOrDefault() string {
	if h.Method == "" {
		return http.MethodGet
	}
	return h.Method
}

// IntervalOrDefault returns the interval for the health check or the default if not set.
func (h HealthCheckConfig) IntervalOrDefault() time.Duration {
	if h.Interval == 0 {
		return time.Minute
	}
	return h.Interval
}

// TimeoutOrDefault returns the timeout or the default if not set.
func (h HealthCheckConfig) TimeoutOrDefault() time.Duration {
	if h.Timeout == 0 {
		return 10 * time.Second
	}
	return h.Timeout
}

// IPVersionOrDefault returns the IP version or the default ("both") if not set.
func (h HealthCheckConfig) IPVersionOrDefault() string {
	if h.IPVersion == "" {
		return "both"
	}
	return h.IPVersion
}

// FallOrDefault returns the fall value or the default if not set.
func (h HealthCheckConfig) FallOrDefault() uint64 {
	if h.Fall == 0 {
		return 1
	}
	return h.Fall
}

// RiseOrDefault returns the rise value or the default if not set.
func (h HealthCheckConfig) RiseOrDefault() uint64 {
	if h.Rise == 0 {
		return 1
	}
	return h.Rise
}

// Validate validates the health check config.
func (h HealthCheckConfig) Validate() error {
	return validation.ValidateStruct(&h,
		validation.Field(&h.ID, validation.Required),
		validation.Field(&h.DisplayName, validation.Required, validation.Length(1, 128)),
		validation.Field(&h.Type, validation.Required, validation.In("http", "https")),
		validation.Field(&h.Method, validation.In(
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
			http.MethodPatch,
		)),
		validation.Field(&h.Port, validation.Min(1), validation.Max(math.MaxUint16)),
		validation.Field(&h.Path, validation.Required, validation.Match(regexp.MustCompile("^/.*$"))),
		validation.Field(&h.Host, validation.When(h.Type == "https", validation.Required)),
		validation.Field(&h.IPVersion, validation.In("ipv4", "ipv6", "both")),
	)
}
