package cfgmodel

import validation "github.com/go-ozzo/ozzo-validation/v4"

// LoggingConfig is the logging configuration.
type LoggingConfig struct {
	// Level is the log level to use. Valid values are: "debug", "info", "warn", "error".
	// Default is "info".
	Level string `koanf:"level"`
	// Format is the log format to use. Valid values are: "json", "text". Default is "json".
	Format string `koanf:"format"`
}

// Validate validates the logging config.
func (c LoggingConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Level, validation.Required, validation.In("debug", "info", "warn", "error")),
		validation.Field(&c.Format, validation.In("json", "text")),
	)
}

// TelemetryConfig is the telemetry configuration.
type TelemetryConfig struct {
	Logging LoggingConfig `koanf:"logging"`
}

// Validate validates the telemetry config.
func (c TelemetryConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Logging),
	)
}
