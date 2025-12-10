package mock

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/gzuidhof/flipper/config/cfgmodel"
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

// NewProviderFromConfig creates a new mock provider from config.
func NewProviderFromConfig(cfg cfgmodel.MockProviderConfig) (*Provider, error) {
	p := NewProvider()

	for _, s := range cfg.Servers {
		var ipv6 netip.Addr
		if s.PublicIPv6 != "" {
			var err error
			ipv6, err = netip.ParseAddr(s.PublicIPv6)
			if err != nil {
				return nil, fmt.Errorf("invalid public_ipv6 %q for server %q: %w", s.PublicIPv6, s.Name, err)
			}
		} else {
			ipv6 = netip.IPv6Unspecified()
		}
		ipv4, err := netip.ParseAddr(s.PublicIPv4)
		if err != nil {
			return nil, fmt.Errorf("invalid public_ipv4 %q for server %q: %w", s.PublicIPv4, s.Name, err)
		}
		p.Servers = append(p.Servers, resource.Server{
			Provider:      resource.ProviderNameMock,
			MockID:        s.ID,
			ServerName:    s.Name,
			Location:      s.Location,
			NetworkZone:   s.NetworkZone,
			ResourceIndex: s.ResourceIndex,
			PublicIPv4:    ipv4,
			PublicIPv6:    ipv6,
		})
	}

	for _, f := range cfg.FloatingIPs {
		currentTarget := ""
		if f.CurrentTarget != 0 {
			currentTarget = fmt.Sprint(f.CurrentTarget)
		}
		ip, err := netip.ParseAddr(f.IP)
		if err != nil {
			return nil, fmt.Errorf("invalid ip %q for floating_ip %q: %w", f.IP, f.Name, err)
		}
		p.FloatingIPs = append(p.FloatingIPs, resource.FloatingIP{
			Provider:       resource.ProviderNameMock,
			MockID:         f.ID,
			FloatingIPName: f.Name,
			Location:       f.Location,
			NetworkZone:    f.NetworkZone,
			ResourceIndex:  f.ResourceIndex,
			IP:             ip,
			CurrentTarget:  currentTarget,
		})
	}

	return p, nil
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
