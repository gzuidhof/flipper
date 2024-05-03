package cfgmodel

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// GroupConfig is the config for a specific group of floating IPs and servers to watch.
// The group is independent from other groups and can have its own Hetzner API token, so
// you could use it to watch resources from different accounts.
type GroupConfig struct {
	// A unique ID for the group. It will be used in metrics and logs.
	ID          string `koanf:"id"`
	DisplayName string `koanf:"display_name"`

	// ReadOnly is a flag that indicates that the group should not perform any actions.
	// This is useful for testing or for monitoring-only setups.
	ReadOnly bool `koanf:"readonly"`

	// PollInterval is the interval at which to poll the Hetzner API for changes.
	// This defaults to 1 minute.
	PollInterval time.Duration `koanf:"poll_interval"`

	// PollTimeout is the maximum time to wait for the Hetzner API to respond.
	// This defaults to 20 seconds.
	PollTimeout time.Duration `koanf:"poll_timeout"`

	// PlanApplyTimeout is the maximum time to wait for a plan to be applied.
	// This defaults to 30 seconds.
	PlanApplyTimeout time.Duration `koanf:"plan_apply_timeout"`

	// PlanApplyWithUnknownStatus is a flag that indicates that the group should apply plans
	// even if the status of one or more servers is unknown.
	PlanApplyWithUnkownStatus bool `koanf:"plan_apply_with_unknown_status"`

	// Provider is the name of the cloud provider that the group is using.
	// Currently only "hetzner" is supported.
	Provider string `koanf:"provider"`

	// Hetzner is the Hetzner-specific configuration.
	// This is only used if the provider is "hetzner".
	Hetzner HetznerProviderConfig `koanf:"hetzner"`

	// Checks is a list of health checks to perform on the servers.
	Checks []HealthCheckConfig `koanf:"checks"`
}

// Validate validates the group config.
func (c GroupConfig) Validate() error {
	// Check that all check IDs are unique.
	ids := make(map[string]struct{})
	for _, check := range c.Checks {
		if _, ok := ids[check.ID]; ok {
			return validation.NewError("duplicate_check_id", "duplicate check ID")
		}
		ids[check.ID] = struct{}{}
	}

	return validation.ValidateStruct(&c,
		validation.Field(&c.ID, validation.Required),
		validation.Field(&c.DisplayName, validation.Required),
		validation.Field(&c.Provider, validation.Required, validation.In("hetzner")),
		validation.Field(&c.Hetzner, validation.When(c.Provider == "hetzner", validation.Required)),
		validation.Field(&c.Checks),
	)
}

// PollIntervalOrDefault returns the poll interval or the default if not set.
func (c GroupConfig) PollIntervalOrDefault() time.Duration {
	if c.PollInterval == 0 {
		return time.Minute
	}
	return c.PollInterval
}

// PollTimeoutOrDefault returns the poll timeout or the default if not set.
func (c GroupConfig) PollTimeoutOrDefault() time.Duration {
	if c.PollTimeout == 0 {
		return 20 * time.Second
	}
	return c.PollTimeout
}

// PlanApplyTimeoutOrDefault returns the plan apply timeout or the default if not set.
func (c GroupConfig) PlanApplyTimeoutOrDefault() time.Duration {
	if c.PlanApplyTimeout == 0 {
		return 30 * time.Second
	}
	return c.PlanApplyTimeout
}

// ServiceConfig describes the service. The name is used in metrics and logs.
type ServiceConfig struct {
	Name string `koanf:"name"`
}

// Config is the top-level configuration for flipper.
type Config struct {
	// Version if a compatibility number for the config. It is always 1 for now.
	Version int `koanf:"version"`

	Groups []GroupConfig `koanf:"groups"`

	Server ServerConfig `koanf:"server"`

	Telemetry TelemetryConfig `koanf:"telemetry"`

	Heartbeat HeartbeatConfig `koanf:"heartbeat"`

	Service ServiceConfig `koanf:"service"`

	Notifications NotificationsConfig `koanf:"notifications"`
}

// Validate validates the config.
func (c Config) Validate() error {
	// Check that all group IDs are unique.
	ids := make(map[string]struct{})
	for _, group := range c.Groups {
		if _, ok := ids[group.ID]; ok {
			return validation.NewError("duplicate_group_id", "duplicate group ID")
		}
		ids[group.ID] = struct{}{}
	}
	return validation.ValidateStruct(&c,
		validation.Field(&c.Version, validation.Required, validation.In(1)),
		validation.Field(&c.Groups),
		validation.Field(&c.Server),
		validation.Field(&c.Telemetry),
		validation.Field(&c.Heartbeat),
	)
}
