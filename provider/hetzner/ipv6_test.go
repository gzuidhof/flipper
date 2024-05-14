package hetzner

import (
	"net"
	"testing"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/stretchr/testify/require"
)

func TestGetIPv6AddressFromNetwork(t *testing.T) {
	ip, net, err := net.ParseCIDR("2a01:4f8:1c17:1d1::/64")
	require.NoError(t, err)

	publicNet := hcloud.ServerPublicNetIPv6{
		ID:      0,
		Network: net,
		IP:      ip,
		Blocked: false,
	}

	addr := getTargetIPv6Address(publicNet)
	require.Equal(t, "2a01:4f8:1c17:1d1::1", addr.String())
}
