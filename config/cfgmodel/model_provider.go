package cfgmodel

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// HetznerProviderConfig is the config for authenticating with Hetzner.
type HetznerProviderConfig struct {
	// API token to use to authenticate with Hetzner
	APIToken string `koanf:"api_token"`

	// ProjectID is the ID of the project to use, you can find this in the URL of the Hetzner Cloud Console.
	// For example, if the URL is https://console.hetzner.cloud/projects/123456, the project ID is 123456.
	//
	// Hetzner does not provide a way to list projects or check the ID, so you will need to know this in advance,
	// see https://github.com/hetznercloud/hcloud-go/issues/451.
	ProjectID string `koanf:"project_id"`

	FloatingIPs HetznerSelector `koanf:"floating_ips"`
	Servers     HetznerSelector `koanf:"servers"`
}

// Validate validates the Hetzner config.
func (c HetznerProviderConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.APIToken, validation.Required),
		validation.Field(&c.ProjectID, validation.Required),
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
