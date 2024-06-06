package resource

import (
	"fmt"
	"net/netip"
	"sort"
)

// FloatingIPs is a list of floating IPs.
type FloatingIPs []FloatingIP

// FloatingIP is a reassignable IP that can be moved between servers.
type FloatingIP struct {
	// Provider is the name of the cloud provider that the floating IP is from.
	Provider ProviderName

	// HetznerID is the unique identifier of the floating IP in Hetzner.
	HetznerID int64

	// FloatingIPName is the name of the floating IP.
	FloatingIPName string

	// Location is the datacenter where the floating IP is located.
	// This is generally `fsn1` or `nbg1`.
	Location string

	// NetworkZone is the network zone where the floating IP is located.
	NetworkZone string

	// IP is the IP address of the floating IP.
	IP netip.Addr

	// CurrentTarget is the ID of the server that the floating IP is currently assigned to.
	// Empty if the floating IP is not currently assigned to a server.
	CurrentTarget string

	// ResourceIndex is more or less the number of a server within a deployment
	// within a zone with the same role.
	//
	// In other words: if there are say 3 API servers, they would probably have
	// index 0 through 2. This is used to somewhat consistently map
	// load balancers to servers.
	//
	// This should be `-1` if unknown (or invalid).
	ResourceIndex int

	// URL is the URL of the floating IP in the cloud provider's web interface.
	URL string
}

// ID returns the unique identifier of the floating IP.
func (f FloatingIP) ID() string {
	return fmt.Sprint(f.HetznerID)
}

// Name returns the name of the floating IP.
func (f FloatingIP) Name() string {
	return f.FloatingIPName
}

// Equal returns true if the two floating IPs are equal.
func (f FloatingIP) Equal(other Resource) bool {
	otherFloatingIP, ok := other.(FloatingIP)
	if !ok {
		return false
	}

	return f.Provider == otherFloatingIP.Provider &&
		f.HetznerID == otherFloatingIP.HetznerID &&
		f.FloatingIPName == otherFloatingIP.FloatingIPName &&
		f.Location == otherFloatingIP.Location &&
		f.NetworkZone == otherFloatingIP.NetworkZone &&
		f.IP == otherFloatingIP.IP &&
		f.ResourceIndex == otherFloatingIP.ResourceIndex &&
		f.CurrentTarget == otherFloatingIP.CurrentTarget
}

// String returns a string representation of the floating IP.
func (f FloatingIP) String() string {
	//nolint:lll // Splitting this line would make it less readable.
	return fmt.Sprintf("FloatingIP{Provider: %s, ID: %s, Name: %s, Location: %s, NetworkZone: %s, IP: %s, CurrentTarget: %s, ResourceIndex: %d}",
		f.Provider, f.ID(), f.FloatingIPName, f.Location, f.NetworkZone, f.IP.String(), f.CurrentTarget, f.ResourceIndex)
}

// SortByName sorts the floating IPs by name.
func (flips FloatingIPs) SortByName() {
	sort.Slice(flips, func(i, j int) bool {
		return flips[i].FloatingIPName < flips[j].FloatingIPName
	})
}
