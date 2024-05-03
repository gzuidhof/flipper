package cfgmodel

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// HeartbeatConfig is the configuration for a heartbeat signal to be sent to a remote service.
type HeartbeatConfig struct {
	// Enabled is a flag that enables or disables the heartbeat.
	Enabled bool `koanf:"enabled"`

	// URL is the URL to send the heartbeat to.
	URL string `koanf:"url"`

	// Interval is the interval at which the heartbeat should be sent.
	Interval time.Duration `koanf:"interval"`

	// Timeout is the maximum time to wait for the heartbeat to be sent.
	Timeout time.Duration `koanf:"timeout"`
}

// Validate validates the heartbeat config.
func (c HeartbeatConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	return validation.ValidateStruct(&c,
		validation.Field(&c.URL, validation.Required),
		validation.Field(&c.Interval, validation.Required),
		validation.Field(&c.Timeout, validation.Required),
	)
}
