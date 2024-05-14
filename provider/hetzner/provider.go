package hetzner

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/resource"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"golang.org/x/sync/errgroup"
)

var _ resource.Provider = Provider{}

// Provider wraps a Hetzner API client.
type Provider struct {
	hc *hcloud.Client

	cfg cfgmodel.GroupConfig

	// locations is a map from location name (e.g. "nbg1") to the location object.
	locations map[string]*hcloud.Location
}

func hetznerIDToResourceID(hetznerID int64) string {
	return fmt.Sprint(hetznerID)
}

// NewProvider creates a new Hetzner provider for a given group.
func NewProvider(ctx context.Context, cfg cfgmodel.GroupConfig) (*Provider, error) {
	if cfg.Hetzner.APIToken == "" {
		return nil, fmt.Errorf("hetzner API token is required")
	}

	hc := hcloud.NewClient(hcloud.WithToken(cfg.Hetzner.APIToken))

	// We load some data that we assume will not change during the lifetime of the client.
	// It has the added benefit of checking that the API key is valid.
	locations, err := hc.Location.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load locations: %w", err)
	}
	locationMap := make(map[string]*hcloud.Location, len(locations))
	for _, loc := range locations {
		locationMap[loc.Name] = loc
	}

	return &Provider{
		hc:        hc,
		cfg:       cfg,
		locations: locationMap,
	}, nil
}

// func (c Client) GetLocation(name string) (*hcloud.Location, error) {
// 	loc, ok := c.locations[name]
// 	if !ok {
// 		return nil, fmt.Errorf("location %s not found", name)
// 	}
// 	return loc, nil
// }

// Name returns the name of the Hetzner provider, "hetzner".
func (c Provider) Name() resource.ProviderName {
	return resource.ProviderNameHetzner
}

// Poll returns the current resources in the provider.
func (c Provider) Poll(ctx context.Context) (resource.Group, error) {
	errgp := errgroup.Group{}

	var floatingIPs []resource.FloatingIP
	var servers []resource.Server

	errgp.Go(func() error {
		flips, err := c.hc.FloatingIP.AllWithOpts(ctx, hcloud.FloatingIPListOpts{
			ListOpts: hcloud.ListOpts{LabelSelector: c.cfg.Hetzner.FloatingIPs.LabelSelector},
		})
		if err != nil {
			return fmt.Errorf("failed to list floating IPs: %w", err)
		}

		for _, flip := range flips {
			ip := flip.IP.String()
			ipParsed, parseErr := netip.ParseAddr(ip)
			if parseErr != nil { // The Hetzner API should always return a valid IP, so this is a bug if it happens.
				return fmt.Errorf("failed to parse IP %s: %w", ip, parseErr)
			}

			currentTarget := ""
			if flip.Server != nil {
				currentTarget = hetznerIDToResourceID(flip.Server.ID)
			}

			floatingIPs = append(floatingIPs, resource.FloatingIP{
				Provider:       c.Name(),
				HetznerID:      flip.ID,
				FloatingIPName: flip.Name,
				Location:       flip.HomeLocation.Name,
				NetworkZone:    string(flip.HomeLocation.NetworkZone),
				IP:             ipParsed,
				CurrentTarget:  currentTarget,
				ResourceIndex:  resourceIndexFromLabel(flip.Labels),
			})
		}
		return nil
	})

	errgp.Go(func() error {
		srvs, err := c.hc.Server.AllWithOpts(ctx, hcloud.ServerListOpts{
			ListOpts: hcloud.ListOpts{LabelSelector: c.cfg.Hetzner.Servers.LabelSelector},
		})
		if err != nil {
			return fmt.Errorf("failed to list servers: %w", err)
		}

		for _, srv := range srvs {
			ipv4Target := netip.MustParseAddr(srv.PublicNet.IPv4.IP.String())
			ipv6Target := getTargetIPv6Address(srv.PublicNet.IPv6)

			servers = append(servers, resource.Server{
				Provider:      c.Name(),
				HetznerID:     srv.ID,
				ServerName:    srv.Name,
				Location:      srv.Datacenter.Location.Name,
				NetworkZone:   string(srv.Datacenter.Location.NetworkZone),
				PublicIPv4:    ipv4Target,
				PublicIPv6:    ipv6Target,
				ResourceIndex: resourceIndexFromLabel(srv.Labels),
			})
		}
		return nil
	})

	if err := errgp.Wait(); err != nil {
		return resource.Group{}, fmt.Errorf("failed to poll Hetzner: %w", err)
	}

	return resource.Group{
		FloatingIPs: floatingIPs,
		Servers:     servers,
	}, nil
}

// AssignFloatingIP targets a floating IP at a server.
func (c Provider) AssignFloatingIP(ctx context.Context, flip resource.FloatingIP, srv resource.Server) error {
	// We check this elsewhere too, but it won't hurt to check here as well.
	if c.cfg.ReadOnly {
		return fmt.Errorf("provider is read-only")
	}

	if flip.Provider != c.Name() {
		return fmt.Errorf("floating IP is not from hetzner: %w", resource.ErrWrongProvider)
	}

	if srv.Provider != c.Name() {
		return fmt.Errorf("server is not from hetzner: %w", resource.ErrWrongProvider)
	}

	// We create these fake objects to use the hcloud-go API without first fetching the objects.
	hflip := &hcloud.FloatingIP{ID: flip.HetznerID}
	hsrv := &hcloud.Server{ID: srv.HetznerID}

	_, _, err := c.hc.FloatingIP.Assign(ctx, hflip, hsrv)
	if err != nil {
		return fmt.Errorf("failed to assign floating IP in hetzner: %w", err)
	}

	return nil
}
