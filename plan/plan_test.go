package plan

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/gzuidhof/flipper/resource"
	"github.com/stretchr/testify/assert"
)

// servers is helper function to create a slice of servers with the given status.
// The ID is also used as the resource index.
func servers(state resource.Status, location, networkZone string, ids ...int64) []*resource.WithStatus[resource.Server] {
	servers := make([]*resource.WithStatus[resource.Server], len(ids))
	for idx, id := range ids {
		serv := resource.Server{
			ServerName:    fmt.Sprintf("mock-server-%d", id),
			HetznerID:     id,
			Location:      location,
			NetworkZone:   networkZone,
			Provider:      resource.ProviderNameMock,
			ResourceIndex: int(id),
		}

		servers[idx] = resource.NewWithStatus(serv, resource.State{Status: state})
	}
	return servers
}

func TestPlan(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		servers      []*resource.WithStatus[resource.Server]
		floatingIPs  []resource.FloatingIP
		expectedPlan Plan
	}{
		{
			name:    "no_changes",
			servers: servers(resource.StatusHealthy, "nbg1", "eu-central", 1, 2, 3),
			floatingIPs: []resource.FloatingIP{
				{HetznerID: 1, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1", FloatingIPName: "floating-ip-1"},
				{HetznerID: 2, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "2", FloatingIPName: "floating-ip-2"},
				{HetznerID: 3, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "3", FloatingIPName: "floating-ip-3"},
			},
			expectedPlan: Plan{
				Actions: []ReassignFloatingIPAction{},
			},
		},
		{
			name:    "spread", // spread the floating IPs across the servers
			servers: servers(resource.StatusHealthy, "nbg1", "eu-central", 1, 2, 3),
			floatingIPs: []resource.FloatingIP{ // They start out all on server 1
				{HetznerID: 1, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1", FloatingIPName: "floating-ip-1"},
				{HetznerID: 2, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1", FloatingIPName: "floating-ip-2"},
				{HetznerID: 3, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1", FloatingIPName: "floating-ip-3"},
			},
			expectedPlan: Plan{
				Actions: []ReassignFloatingIPAction{
					{ServerID: "2", FloatingIPID: "2"},
					{ServerID: "3", FloatingIPID: "3"},
				},
			},
		},
		{
			name:    "spread_looparound",
			servers: servers(resource.StatusHealthy, "nbg1", "eu-central", 1, 2),
			floatingIPs: []resource.FloatingIP{
				{HetznerID: 1, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1", FloatingIPName: "floating-ip-1"},
				{HetznerID: 2, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1", FloatingIPName: "floating-ip-2"},
				{HetznerID: 3, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1", FloatingIPName: "floating-ip-3"},
			},
			expectedPlan: Plan{
				Actions: []ReassignFloatingIPAction{
					{ServerID: "2", FloatingIPID: "2"},
				},
			},
		},
		{
			name: "prefer_own_location",
			servers: append(
				servers(resource.StatusHealthy, "nbg1", "eu-central", 1, 2),
				servers(resource.StatusHealthy, "fsn1", "eu-central", 3, 4)...,
			),
			floatingIPs: []resource.FloatingIP{ // Start unassigned.
				{HetznerID: 1, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "", FloatingIPName: "floating-ip-1"},
				{HetznerID: 2, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "", FloatingIPName: "floating-ip-2"},
				{HetznerID: 3, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "", FloatingIPName: "floating-ip-3"},
			},
			expectedPlan: Plan{
				Actions: []ReassignFloatingIPAction{
					{ServerID: "1", FloatingIPID: "1"},
					{ServerID: "2", FloatingIPID: "2"},
					{ServerID: "1", FloatingIPID: "3"},
				},
			},
		},
		{
			name:    "other_location",
			servers: servers(resource.StatusHealthy, "nbg1", "eu-central", 1, 2),
			floatingIPs: []resource.FloatingIP{
				{HetznerID: 1, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "", FloatingIPName: "floating-ip-1"},
				{HetznerID: 2, Location: "fsn1", NetworkZone: "eu-central", CurrentTarget: "", FloatingIPName: "floating-ip-2"},
				{HetznerID: 3, Location: "fsn1", NetworkZone: "eu-central", CurrentTarget: "", FloatingIPName: "floating-ip-3"},
			},
			expectedPlan: Plan{
				Actions: []ReassignFloatingIPAction{
					{ServerID: "1", FloatingIPID: "1"},
					{ServerID: "2", FloatingIPID: "2"},
					{ServerID: "1", FloatingIPID: "3"},
				},
			},
		},
		{
			name:    "different_network_zone",
			servers: servers(resource.StatusHealthy, "nbg1", "eu-north", 1),
			floatingIPs: []resource.FloatingIP{
				{HetznerID: 1, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "", FloatingIPName: "floating-ip-1"},
			},
			expectedPlan: Plan{
				Actions: []ReassignFloatingIPAction{},
			},
		},
		{
			name: "unhealthy_across_regions",
			servers: append(
				servers(resource.StatusUnhealthy, "nbg1", "eu-central", 1, 2),
				servers(resource.StatusHealthy, "fsn1", "eu-central", 3, 4)...,
			),
			floatingIPs: []resource.FloatingIP{
				{HetznerID: 1, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1", FloatingIPName: "floating-ip-1"},
				{HetznerID: 2, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "2", FloatingIPName: "floating-ip-2"},
				{HetznerID: 3, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "", FloatingIPName: "floating-ip-3"},
			},
			expectedPlan: Plan{
				Actions: []ReassignFloatingIPAction{
					{ServerID: "3", FloatingIPID: "1"},
					{ServerID: "4", FloatingIPID: "2"},
					{ServerID: "3", FloatingIPID: "3"},
				},
			},
		},
		{
			name:    "assigned_to_unknown_server",
			servers: servers(resource.StatusHealthy, "nbg1", "eu-central", 1),
			floatingIPs: []resource.FloatingIP{
				{HetznerID: 1, Location: "nbg1", NetworkZone: "eu-central", CurrentTarget: "1234", FloatingIPName: "floating-ip-1"},
			},
			expectedPlan: Plan{
				Actions: []ReassignFloatingIPAction{},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := NewState(tc.floatingIPs, tc.servers)

			plan := New(state)
			// The plan ID is random, so we can't compare it.
			plan.ID = uuid.UUID{}
			assert.Equal(t, tc.expectedPlan, plan)
		})
	}
}
