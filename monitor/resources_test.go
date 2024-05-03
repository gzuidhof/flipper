package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/provider/mock"
	"github.com/gzuidhof/flipper/resource"
	"github.com/stretchr/testify/assert"
)

func TestResourcesWatcher(t *testing.T) {
	t.Parallel()
	cfg := cfgmodel.GroupConfig{}

	provider := mock.NewProvider()

	provider.Servers = append(provider.Servers, resource.Server{
		Provider:   resource.ProviderNameMock,
		ServerName: "mock-server-1",
		HetznerID:  1,
	})

	watcher := NewResourcesWatcher(cfg, slog.Default(), provider)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, cs, err := watcher.Update(ctx)
	assert.NoError(t, err)
	assert.Len(t, r.Servers, 1)
	assert.Len(t, cs.Servers.Added, 1)

	// Update again, no changes
	r, cs, err = watcher.Update(ctx)
	assert.NoError(t, err)

	assert.Len(t, r.Servers, 1)
	assert.Len(t, cs.Servers.Added, 0)
	assert.True(t, cs.Empty())

	// Add a server
	provider.Servers = append(provider.Servers, resource.Server{
		Provider:   resource.ProviderNameMock,
		ServerName: "mock-server-2",
		HetznerID:  2,
	})

	r, cs, err = watcher.Update(ctx)
	assert.NoError(t, err)

	assert.Len(t, r.Servers, 2)
	assert.Len(t, cs.Servers.Added, 1)
	fmt.Printf("%s", cs)
}
