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
		validation.Field(&c.LabelSelector, validation.Required),
	)
}

// MockProviderConfig is the config for the mock provider used for testing.
type MockProviderConfig struct {
	// Servers is a list of mock servers to create.
	Servers []MockServerConfig `koanf:"servers"`
	// FloatingIPs is a list of mock floating IPs to create.
	FloatingIPs []MockFloatingIPConfig `koanf:"floating_ips"`
}

// Validate validates the mock provider config.
func (c MockProviderConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Servers),
		validation.Field(&c.FloatingIPs),
	)
}

// MockServerConfig is the config for a mock server.
type MockServerConfig struct {
	ID            int64  `koanf:"id"`
	Name          string `koanf:"name"`
	Location      string `koanf:"location"`
	NetworkZone   string `koanf:"network_zone"`
	ResourceIndex int    `koanf:"resource_index"`
	PublicIPv4    string `koanf:"public_ipv4"`
	PublicIPv6    string `koanf:"public_ipv6"`
}

// Validate validates the mock server config.
func (c MockServerConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ID, validation.Required, validation.Min(int64(1))),
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.PublicIPv4, validation.Required),
	)
}

// MockFloatingIPConfig is the config for a mock floating IP.
type MockFloatingIPConfig struct {
	ID            int64  `koanf:"id"`
	Name          string `koanf:"name"`
	Location      string `koanf:"location"`
	NetworkZone   string `koanf:"network_zone"`
	ResourceIndex int    `koanf:"resource_index"`
	IP            string `koanf:"ip"`
	// CurrentTarget is the ID of the server this floating IP is assigned to.
	// Use 0 or omit for unassigned.
	CurrentTarget int64 `koanf:"current_target"`
}

// Validate validates the mock floating IP config.
func (c MockFloatingIPConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ID, validation.Required, validation.Min(int64(1))),
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.IP, validation.Required),
	)
}
