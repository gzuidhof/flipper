package resource

import "context"

// ProviderName is the name of a cloud provider.
type ProviderName string

const (
	// ProviderNameHetzner is the name of the Hetzner cloud provider.
	ProviderNameHetzner ProviderName = "hetzner"

	// ProviderNameMock is the name of the mock cloud provider used for testing.
	ProviderNameMock ProviderName = "mock"
)

// Provider is a cloud provider that can provide resources.
type Provider interface {
	// Name uniquely identifies the provider.
	Name() ProviderName

	// Poll returns the current resources in the provider.
	Poll(ctx context.Context) (Group, error)

	// AssignFloatingIP targets a floating IP at a server.
	AssignFloatingIP(ctx context.Context, flip FloatingIP, srv Server) error
}
