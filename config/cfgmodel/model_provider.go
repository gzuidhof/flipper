package cfgmodel

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// HetznerProviderConfig is the config for authenticating with Hetzner.
type HetznerProviderConfig struct {
	// API token to use to authenticate with Hetzner
	APIToken string `koanf:"api_token"`

	FloatingIPs HetznerSelector `koanf:"floating_ips"`
	Servers     HetznerSelector `koanf:"servers"`
}

// Validate validates the Hetzner config.
func (c HetznerProviderConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.APIToken, validation.Required),
		validation.Field(&c.FloatingIPs, validation.Required),
		validation.Field(&c.Servers, validation.Required),
	)
}

// HetznerSelector is a selector for a group of resources on Hetzner.
type HetznerSelector struct {
	LabelSelector string `koanf:"label_selector"`
	// In the future we could add more fields here, like a list of IDs if we ever need that.
}

// Validate validates the Hetzner selector.
func (c HetznerSelector) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.LabelSelector, validation.Required), // This disallows an empty selector.
	)
}
