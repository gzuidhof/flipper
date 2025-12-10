package monitor

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/notification"
	"github.com/gzuidhof/flipper/provider/mock"
	"github.com/gzuidhof/flipper/resource"
)

func TestMonitorWatch_BlocksUntilContextCancelled(t *testing.T) {
	t.Parallel()

	provider := mock.NewProvider()
	provider.Servers = append(provider.Servers, resource.Server{
		Provider:   resource.ProviderNameMock,
		ServerName: "mock-server",
		MockID:     1,
	})

	cfg := cfgmodel.GroupConfig{
		ID:           "test-group",
		DisplayName:  "Test Group",
		Provider:     "mock",
		PollInterval: 100 * time.Millisecond,
	}

	group := NewGroup(cfg, provider, slog.Default(), &notification.NoopNotifier{})
	m := &Monitor{groups: []*Group{group}}

	ctx, cancel := context.WithCancel(context.Background())

	watchReturned := make(chan struct{})
	go func() {
		_ = m.Watch(ctx)
		close(watchReturned)
	}()

	// Watch should not return before context is cancelled.
	time.Sleep(50 * time.Millisecond)
	select {
	case <-watchReturned:
		t.Fatal("Watch() returned before context was cancelled")
	default:
	}

	cancel()

	select {
	case <-watchReturned:
	case <-time.After(5 * time.Second):
		t.Fatal("Watch() did not return after context cancellation")
	}
}
