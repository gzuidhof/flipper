package hetzner

import (
	"net/netip"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// getTargetIPv6Address returns the target IPv6 address for a given IPv6 network, which is the second address in the
// network range. This is necessary in Hetzner,
// see https://docs.hetzner.com/cloud/servers/getting-started/connecting-to-the-server/
func getTargetIPv6Address(netIPv6 hcloud.ServerPublicNetIPv6) netip.Addr {
	if netIPv6.IsUnspecified() {
		return netip.Addr{}
	}
	// Prefix is something like "2a01:4f8:1c17:1d1::/64"
	prefix := netip.MustParsePrefix(netIPv6.Network.String())

	if prefix.IsSingleIP() {
		return prefix.Addr()
	}
	// Will return "2a01:4f8:1c17:1d1::1"
	return prefix.Addr().Next()
}
