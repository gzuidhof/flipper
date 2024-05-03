package plan

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/gzuidhof/flipper/resource"
)

// ReassignFloatingIPAction is a plan to reassign a floating IP to a different server.
type ReassignFloatingIPAction struct {
	// FloatingIPID is the ID of the floating IP to be reassigned.
	FloatingIPID string

	// ServerID string is the ID of the server to which the floating IP should be reassigned.
	ServerID string
}

// String returns a string representation of the action.
func (a ReassignFloatingIPAction) String() string {
	return a.FloatingIPID + " -> " + a.ServerID
}

// Plan describes the changeset to be applied to the cloud resources.
type Plan struct {
	// ID is a random UUID to identify the plan.
	ID      uuid.UUID
	Actions []ReassignFloatingIPAction
}

// AddAction adds an action to the plan.
func (p *Plan) AddAction(action ReassignFloatingIPAction) {
	p.Actions = append(p.Actions, action)
}

// New takes the current state of the cloud resources and returns a plan for reassigning floating IPs.
func New(s State) Plan {
	// First, we find all candidate servers.
	allCandidates := s.CandidateServers()
	todo := s.CandidateFloatingIPs()

	// Floating IP ID to server ID
	proposal := map[string]string{}
	// Server ID to floating IP count
	assignCount := map[string]int{}

	// Locations that contain at least one candidate server.
	candidatesPerLocation := map[string][]*resource.WithStatus[resource.Server]{}
	// Network zones that contain at least one candidate server.
	availableRegions := map[string]bool{}
	for _, server := range allCandidates {
		loc := server.Resource.Location
		candidatesPerLocation[loc] = append(candidatesPerLocation[loc], server)
		availableRegions[server.Resource.NetworkZone] = true
	}

	// Floating IPs that are unassignable.
	unassignable := map[string]bool{}

	// First we try to assign floating IPs to servers in the same location with the same index.
	for _, flip := range todo {
		if !availableRegions[flip.NetworkZone] {
			unassignable[flip.ID()] = true
			slog.Warn("no candidate servers in network zone for floating IP",
				slog.String("network_zone", flip.NetworkZone),
				slog.String("floating_ip_id", flip.ID()),
			)
			continue
		}

		candidates, sameLocationPossible := candidatesPerLocation[flip.Location]
		if !sameLocationPossible {
			continue
		}

		// The first choice is the server that is in the same location and shares the index with the floating IP.
		for _, server := range candidates {
			if server.Resource.ResourceIndex == flip.ResourceIndex {
				proposal[flip.ID()] = server.Resource.ID()
				assignCount[server.Resource.ID()]++
				break
			}
		}
	}

	// Then we assign the remaining floating IPs.
	for _, flip := range todo {
		if _, ok := proposal[flip.ID()]; ok { // Already assigned.. O(2nm) is fine
			continue
		}
		if unassignable[flip.ID()] {
			continue
		}

		candidates, sameLocationPossible := candidatesPerLocation[flip.Location]
		if !sameLocationPossible {
			candidates = allCandidates
		}

		// The second choice is the server that has the least floating IPs assigned to it.
		var minCount int
		var minServerID string
		for _, server := range candidates {
			if count := assignCount[server.Resource.ID()]; count < minCount || minServerID == "" {
				minCount = count
				minServerID = server.Resource.ID()
			}
		}
		proposal[flip.ID()] = minServerID
		assignCount[minServerID]++
	}

	// Now we remove any assignments that are already in place.
	for _, flip := range s.FloatingIPs {
		if flip.CurrentTarget == "" {
			continue
		}
		if flip.CurrentTarget == proposal[flip.ID()] {
			delete(proposal, flip.ID())
		}
	}

	return planFromProposal(proposal)
}

// planFromProposal creates a plan from a proposal.
// The proposal is a map of floating IP IDs to server IDs.
func planFromProposal(proposal map[string]string) Plan {
	plan := Plan{
		ID:      uuid.New(),
		Actions: make([]ReassignFloatingIPAction, 0, len(proposal)),
	}

	// Let's order by floating IP ID to make the plan deterministic.
	proposalKeys := make([]string, 0, len(proposal))
	for flipID := range proposal {
		proposalKeys = append(proposalKeys, flipID)
	}
	slices.Sort(proposalKeys)

	for _, k := range proposalKeys {
		plan.AddAction(ReassignFloatingIPAction{
			ServerID:     proposal[k],
			FloatingIPID: k,
		})
	}

	return plan
}

// String returns a string representation of the plan.
func (p Plan) String() string {
	actionStrings := make([]string, 0, len(p.Actions))
	for _, action := range p.Actions {
		actionStrings = append(actionStrings, action.String())
	}
	res := "Plan{"
	if len(actionStrings) > 0 {
		res += " " + strings.Join(actionStrings, ", ") + " "
	}
	res += "}"
	return res
}

// Empty returns true if the plan is empty.
func (p Plan) Empty() bool {
	return len(p.Actions) == 0
}

// ToBeReassignedMap returns the floating IP IDs that will be reassigned.
func (p Plan) ToBeReassignedMap() map[string]bool {
	m := make(map[string]bool, len(p.Actions))
	for _, action := range p.Actions {
		m[action.FloatingIPID] = true
	}
	return m
}
