package cfgmodel

import (
	"math"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// ServerConfig is the config for the built-in HTTP server.
type ServerConfig struct {
	Enabled bool `koanf:"enabled"`
	// Host to bind to.
	Host string `koanf:"host"`
	// Port to bind to.
	Port int `koanf:"port"`

	// Timeout for graceful shutdown.
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`

	// Assets is the path to serve assets from. If empty, the server will use the embedded files.
	// Generally you will only need to specify this during development for quick iteration.
	Assets string `koanf:"assets"`
	// Templates is the path to serve templates from. If empty, the server will use the embedded files.
	// Generally you will only need to specify this during development for quick iteration.
	Templates string `koanf:"templates"`
}

// Validate validates the server config.
func (c ServerConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	return validation.ValidateStruct(&c,
		validation.Field(&c.Port, validation.Required, validation.Min(1), validation.Max(math.MaxUint16)),
		validation.Field(&c.ShutdownTimeout, validation.Required),
	)
}
