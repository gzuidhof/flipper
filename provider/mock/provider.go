package mock

import (
	"context"
	"time"

	"github.com/gzuidhof/flipper/resource"
)

var _ resource.Provider = (*Provider)(nil)

// Provider is a mock provider for testing.
type Provider struct {
	FloatingIPs []resource.FloatingIP
	Servers     []resource.Server

	PollDelay             time.Duration
	AssignFloatingIPDelay time.Duration

	PollError             error
	AssignFloatingIPError error
}

// NewProvider creates a new mock provider.
func NewProvider() *Provider {
	return &Provider{
		FloatingIPs: []resource.FloatingIP{},
		Servers:     []resource.Server{},
	}
}

// Name returns the name of the mock provider, "mock".
func (p *Provider) Name() resource.ProviderName {
	return resource.ProviderNameMock
}

// Poll returns the current state of the cloud resources.
func (p *Provider) Poll(_ context.Context) (resource.Group, error) {
	if p.PollDelay > 0 {
		time.Sleep(p.PollDelay)
	}
	if p.PollError != nil {
		return resource.Group{}, p.PollError
	}

	result := resource.Group{}

	result.FloatingIPs = append(result.FloatingIPs, p.FloatingIPs...)
	result.Servers = append(result.Servers, p.Servers...)

	return result, nil
}

// AssignFloatingIP assigns a floating IP to a server.
func (p *Provider) AssignFloatingIP(_ context.Context, fip resource.FloatingIP, server resource.Server) error {
	if p.AssignFloatingIPDelay > 0 {
		time.Sleep(p.AssignFloatingIPDelay)
	}
	if p.AssignFloatingIPError != nil {
		return p.AssignFloatingIPError
	}

	for i, f := range p.FloatingIPs {
		if f.ID() == fip.ID() {
			p.FloatingIPs[i].CurrentTarget = server.ID()
			break
		}
	}

	return nil
}
