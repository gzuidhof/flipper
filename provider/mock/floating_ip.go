package mock

import (
	"math/rand"
	"net/netip"

	"github.com/gzuidhof/flipper/resource"
)

// NewFloatingIP creates a new mock floating IP.
func NewFloatingIP(name, location, networkZone string, ip netip.Addr) resource.FloatingIP {
	return resource.FloatingIP{
		Provider: resource.ProviderNameMock,
		//nolint:gosec // This is a mock provider, so we don't need to worry about cryptographic security.
		HetznerID:      rand.Int63(),
		FloatingIPName: name,
		Location:       location,
		NetworkZone:    networkZone,
		IP:             ip,
		CurrentTarget:  "",
		URL:            "",
	}
}

// NewFloatingIPv4 creates a new mock floating IP with an unspecified IPv4 address.
func NewFloatingIPv4(name, location, networkZone string) resource.FloatingIP {
	return NewFloatingIP(name, location, networkZone, netip.IPv4Unspecified())
}

// NewFloatingIPv6 creates a new mock floating IP with an unspecified IPv6 address.
func NewFloatingIPv6(name, location, networkZone string) resource.FloatingIP {
	return NewFloatingIP(name, location, networkZone, netip.IPv6Unspecified())
}
