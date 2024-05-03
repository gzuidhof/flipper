package plan

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/gzuidhof/flipper/resource"
)

// State represents the current state of the cloud resources.
type State struct {
	FloatingIPs map[string]resource.FloatingIP
	Servers     map[string]*resource.WithStatus[resource.Server]
}

// NewState creates a new state.
func NewState(floatingIPs []resource.FloatingIP, servers []*resource.WithStatus[resource.Server]) State {
	state := State{
		FloatingIPs: make(map[string]resource.FloatingIP, len(floatingIPs)),
		Servers:     make(map[string]*resource.WithStatus[resource.Server], len(servers)),
	}

	for _, fip := range floatingIPs {
		state.FloatingIPs[fip.ID()] = fip
	}

	for _, s := range servers {
		state.Servers[s.Resource.ID()] = s
	}

	return state
}

// NewStateFromGroup creates a new state from a group of resources.
// The resources will start with an unknown status.
func NewStateFromGroup(group resource.Group) State {
	statefulServers := make([]*resource.WithStatus[resource.Server], 0, len(group.Servers))
	for _, server := range group.Servers {
		statefulServers = append(statefulServers,
			resource.NewWithStatus(server, resource.State{Status: resource.StatusUnknown}),
		)
	}

	return NewState(group.FloatingIPs, statefulServers)
}

// CandidateServers returns all healthy servers, or all servers if there are no healthy ones.
// It returns them in a fixed order.
func (s State) CandidateServers() []*resource.WithStatus[resource.Server] {
	var candidates []*resource.WithStatus[resource.Server]
	for _, server := range s.Servers {
		if server.IsHealthy() {
			candidates = append(candidates, server)
		}
	}

	if len(candidates) == 0 {
		slog.Warn("no healthy servers found")
		// We use all servers as candidates if there are no healthy ones.
		for _, server := range s.Servers {
			candidates = append(candidates, server)
		}
	}

	// Sort by location first, and then the index (if not -1), and finally the name.
	slices.SortStableFunc(candidates, func(i, j *resource.WithStatus[resource.Server]) int {
		if i.Resource.Location != j.Resource.Location {
			return strings.Compare(i.Resource.Location, j.Resource.Location)
		}
		sij := i.Resource.ResourceIndex
		sjj := j.Resource.ResourceIndex

		if sij == -1 && sjj == -1 { // Both servers don't have an index
			return strings.Compare(i.Resource.Name(), j.Resource.Name())
		}
		if sij == -1 { // i doesn't have an index
			return 1
		}
		if sjj == -1 { // j doesn't have an index
			return -1
		}
		if sij != sjj {
			return sij - sjj
		}
		return strings.Compare(i.Resource.Name(), j.Resource.Name())
	})
	return candidates
}

// CandidateFloatingIPs returns all floating IPs that we should me managing the targets of.
func (s State) CandidateFloatingIPs() []resource.FloatingIP {
	flips := make([]resource.FloatingIP, 0, len(s.FloatingIPs))
	for _, flip := range s.FloatingIPs {
		f := flip

		// We do not touch floating IPs that are pointed to a server that is unknown to this tool.
		if _, ok := s.Servers[f.CurrentTarget]; !ok && f.CurrentTarget != "" {
			slog.Warn("floating IP points to unknown server",
				"floating_ip_id", f.HetznerID,
				"server_id", f.CurrentTarget,
			)
			continue
		}

		flips = append(flips, f)
	}

	// Sort todo to make the plan deterministic.
	// The exact order doesn't matter, as long as it's deterministic. For the tests to be simple we sort by ID.
	slices.SortFunc(flips, func(i, j resource.FloatingIP) int {
		return int(i.HetznerID - j.HetznerID)
	})

	return flips
}

// UnassignedFloatingIPs returns all floating IPs that are not assigned to a server.
func (s State) UnassignedFloatingIPs() []resource.FloatingIP {
	var unassigned []resource.FloatingIP
	for _, fip := range s.FloatingIPs {
		if fip.CurrentTarget == "" {
			unassigned = append(unassigned, fip)
		}
	}
	return unassigned
}

// UnhealthyFloatingIPs returns all floating IPs that are assigned to an unhealthy server.
func (s State) UnhealthyFloatingIPs() []resource.FloatingIP {
	var unhealthy []resource.FloatingIP
	for _, flip := range s.FloatingIPs {
		if flip.CurrentTarget == "" {
			continue // No target.
		}
		serverTarget, ok := s.Servers[flip.CurrentTarget]
		if !ok { // For now let's not reassign floating IPs that have an unknown target.
			slog.Warn("floating IP has unknown target server",
				slog.String("floating_ip_id", flip.ID()),
				slog.String("floating_ip_name", flip.FloatingIPName),
				slog.String("target", flip.CurrentTarget),
			)
			continue
		}
		if serverTarget.IsUnhealthy() {
			unhealthy = append(unhealthy, flip)
		}
	}
	return unhealthy
}

// FloatingIPsByServer returns a map of server IDs to floating IPs.
func (s State) FloatingIPsByServer() map[string]resource.FloatingIPs {
	fipsByServer := make(map[string]resource.FloatingIPs)
	for _, fip := range s.FloatingIPs {
		if fip.CurrentTarget == "" {
			continue
		}
		fipsByServer[fip.CurrentTarget] = append(fipsByServer[fip.CurrentTarget], fip)
	}

	// Sort the floating IPs by name for each server.
	for _, fips := range fipsByServer {
		fips.SortByName()
	}
	return fipsByServer
}

// HasServersWithUnknownStatus returns true if there are servers with an unknown status.
func (s State) HasServersWithUnknownStatus() bool {
	for _, server := range s.Servers {
		if server.Status() == resource.StatusUnknown {
			return true
		}
	}
	return false
}

// FloatingIPsTargetedOutsideGroup returns the floating IPs that are targeted at unknown servers.
func (s State) FloatingIPsTargetedOutsideGroup() resource.FloatingIPs {
	var unknown resource.FloatingIPs
	for _, flip := range s.FloatingIPs {
		if _, ok := s.Servers[flip.CurrentTarget]; !ok && flip.CurrentTarget != "" {
			unknown = append(unknown, flip)
		}
	}

	return unknown
}

// NumUnhealthyServers returns the number of unhealthy servers.
func (s State) NumUnhealthyServers() int {
	count := 0
	for _, server := range s.Servers {
		if server.IsUnhealthy() {
			count++
		}
	}
	return count
}
