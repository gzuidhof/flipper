package resource

// Group is a set of resources. It generally created as the result of a Provider's poll call.
type Group struct {
	FloatingIPs []FloatingIP
	Servers     []Server
}

// FloatingIPsByID returns a map of FloatingIPs by their ID.
func (g *Group) FloatingIPsByID() map[string]FloatingIP {
	fips := make(map[string]FloatingIP, len(g.FloatingIPs))
	for _, fip := range g.FloatingIPs {
		fips[fip.ID()] = fip
	}
	return fips
}

// ServersByID returns a map of Servers by their ID.
func (g *Group) ServersByID() map[string]Server {
	servers := make(map[string]Server, len(g.Servers))
	for _, server := range g.Servers {
		servers[server.ID()] = server
	}
	return servers
}

// Empty returns true if the group is empty.
func (g *Group) Empty() bool {
	return len(g.FloatingIPs) == 0 && len(g.Servers) == 0
}
