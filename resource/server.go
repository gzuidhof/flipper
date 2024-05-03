package resource

import (
	"fmt"
	"net/netip"
	"sort"
)

// Servers is a list of servers.
type Servers []Server

// Server is a physical or virtual server that can be assigned a floating IP.
type Server struct {
	// Provider is name of the cloud provider where the server is located.
	Provider ProviderName

	// HetznerID is the unique ID of the server in Hetzner.
	HetznerID int64

	// ServerName is the name of the server.
	ServerName string

	// Location is the datacenter where the server is located.
	// This is generally `fsn1` or `nbg1`.
	Location string

	// NetworkZone is the network zone where the server is located.
	// This is generally `eu-central`.
	NetworkZone string

	// ResourceIndex is more or less the number of a server within a deployment
	// within a zone with the same role.
	//
	// In other words: if there are say 3 API servers, they would probably have
	// index 0 through 2. This is used to somewhat consistently map
	// load balancers to servers.
	//
	// This should be `-1` if unknown (or invalid).
	ResourceIndex int

	// PublicIPv4 is the public IPv4 address of the server.
	PublicIPv4 netip.Addr

	// PublicIPv6 is the public IPv6 address of the server.
	PublicIPv6 netip.Addr
}

// ID returns the unique identifier of the server.
func (s Server) ID() string {
	return fmt.Sprint(s.HetznerID)
}

// Name returns the name of the server.
func (s Server) Name() string {
	return s.ServerName
}

// Equal returns true if the two servers are equal.
func (s Server) Equal(other Resource) bool {
	otherServer, ok := other.(Server)
	if !ok {
		return false
	}

	return s.Provider == otherServer.Provider &&
		s.HetznerID == otherServer.HetznerID &&
		s.ServerName == otherServer.ServerName &&
		s.Location == otherServer.Location &&
		s.NetworkZone == otherServer.NetworkZone &&
		s.ResourceIndex == otherServer.ResourceIndex &&
		s.PublicIPv4 == otherServer.PublicIPv4 &&
		s.PublicIPv6 == otherServer.PublicIPv6
}

// String returns a string representation of the server.
func (s Server) String() string {
	//nolint:lll // Splitting it doesn't make it more readable.
	return fmt.Sprintf("Server{Provider: %s, ID: %s, Name: %s, Location: %s, NetworkZone: %s, ResourceIndex: %d, IPv4: %s, IPv6: %s}",
		s.Provider, s.ID(), s.ServerName, s.Location, s.NetworkZone, s.ResourceIndex, s.PublicIPv4, s.PublicIPv6)
}

// SortByName sorts the servers by name.
func (servs Servers) SortByName() {
	sort.Slice(servs, func(i, j int) bool {
		return servs[i].ServerName < servs[j].ServerName
	})
}
