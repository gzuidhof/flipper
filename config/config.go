// Package config provides the logic around the configuration of the flipper service.
package config

import (
	"fmt"

	_ "embed"

	"github.com/gzuidhof/ckoanf"
	"github.com/gzuidhof/flipper/config/cfgmodel"
)

//go:embed defaults.yaml
var defaults []byte

const envPrefix = "FLIPPER_"

// Init initializes the configuration from defaults, the given file and environment variables.
// If configFilepath is empty, only the built-in defaults and environment variables are used.
func Init(configFilepath string) (*cfgmodel.Config, error) {
	sources := []ckoanf.SourceFunc[*cfgmodel.Config]{
		ckoanf.EmbeddedDefaults[*cfgmodel.Config](
			defaults,
			ckoanf.FileTypeYAML,
		),
	}
	if configFilepath != "" {
		sources = append(sources, ckoanf.LocalFile[*cfgmodel.Config](configFilepath))
	}
	sources = append(sources, ckoanf.Env[*cfgmodel.Config](envPrefix))

	config, err := ckoanf.Init(
		&cfgmodel.Config{},
		ckoanf.WithSource(
			sources...,
		),
		ckoanf.WithValidation[*cfgmodel.Config](true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// For now we just return the config struct itself. We don't use the key-based lookup or mutate
	// the config at any point, so this will do.
	return config.Model(), nil
}

// Default returns the default, baked-in configuration.
// It panics if the default configuration is not valid.
func Default() *cfgmodel.Config {
	config, err := ckoanf.Init(
		&cfgmodel.Config{},
		ckoanf.WithSource(
			ckoanf.EmbeddedDefaults[*cfgmodel.Config](
				defaults,
				ckoanf.FileTypeYAML,
			),
		),
		ckoanf.WithValidation[*cfgmodel.Config](true),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize default configuration: %v", err))
	}
	return config.Model()
}
