package resource

import "errors"

// ErrNilFloatingIP is returned when a nil floating IP is used.
var ErrNilFloatingIP = errors.New("floating IP is nil")

// FloatingIPPair is a pair of floating IPs, one IPv4 and one IPv6.
type FloatingIPPair struct {
	v4 FloatingIP
	v6 FloatingIP
}

// NewFloatingIPPair creates a new FloatingIPPair.
func NewFloatingIPPair(v4, v6 FloatingIP) (FloatingIPPair, error) {
	if !v4.IP.Is4() {
		return FloatingIPPair{}, errors.New("v4 is not an IPv4 address")
	}

	if !v6.IP.Is6() {
		return FloatingIPPair{}, errors.New("v6 is not an IPv6 address")
	}

	return FloatingIPPair{v4: v4, v6: v6}, nil
}
