package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gzuidhof/flipper/config/cfgmodel"
	"github.com/gzuidhof/flipper/resource"
)

// ResourceUpdate contains the current resources and the changeset since the last update.
type ResourceUpdate struct {
	// Sequence is the sequence number of the update, it can be used to detect stale updates.
	Sequence uint64
	// Resources is the current state of the resources.
	Resources resource.Group
	Changeset resource.GroupChangeset
}

// ResourcesWatcher watches for changes of the resources in a resource group itself.
// It polls the resources from the provider and sends updates over a channel.
type ResourcesWatcher struct {
	sync.Mutex

	updateMutex sync.Mutex

	cfg      cfgmodel.GroupConfig
	provider resource.Provider
	logger   *slog.Logger

	resources resource.Group

	currentSequence uint64
}

// NewResourcesWatcher creates a new resource watcher for a group.
func NewResourcesWatcher(cfg cfgmodel.GroupConfig, logger *slog.Logger, provider resource.Provider) *ResourcesWatcher {
	return &ResourcesWatcher{
		cfg:      cfg,
		provider: provider,
		logger:   logger,
	}
}

// poll the resources from the provider with the configured timeout.
func (w *ResourcesWatcher) poll(ctx context.Context) (resource.Group, error) {
	ctx, cancel := context.WithTimeout(ctx, w.cfg.PollTimeoutOrDefault())
	defer cancel()

	g, err := w.provider.Poll(ctx)
	if err != nil {
		return resource.Group{}, fmt.Errorf("resources watched failed to to poll resources: %w", err)
	}
	return g, nil
}

// Update the resources by polling the provider. This updates the internal state
// and returns the resources and changeset from the last call to Update.
func (w *ResourcesWatcher) Update(ctx context.Context) (resource.Group, resource.GroupChangeset, error) {
	newResources, err := w.poll(ctx)
	if err != nil {
		return resource.Group{}, resource.GroupChangeset{}, fmt.Errorf("watcher failed to poll resources: %w", err)
	}
	// Check if the context was canceled in the meantime.
	if ctx.Err() != nil {
		return resource.Group{}, resource.GroupChangeset{}, fmt.Errorf("context canceled")
	}

	w.Lock()
	cs := resource.NewGroupChangeset(w.resources, newResources)
	w.resources = newResources
	w.Unlock()

	return newResources, cs, nil
}

func (w *ResourcesWatcher) bumpSequence() uint64 {
	w.Lock()
	defer w.Unlock()

	w.currentSequence++
	return w.currentSequence
}

// performUpdate is a helper function to perform an update, sending the result on the given channels.
func (w *ResourcesWatcher) performUpdate(
	ctx context.Context,
	onChange chan<- ResourceUpdate,
	onError chan<- error,
	forceSendUpdate bool,
) uint64 {
	w.updateMutex.Lock()
	defer w.updateMutex.Unlock()

	seq := w.bumpSequence()

	r, cs, err := w.Update(ctx)
	if ctx.Err() != nil {
		return seq
	}
	if err != nil {
		onError <- err
		return seq
	}
	if forceSendUpdate || !cs.Empty() {
		onChange <- ResourceUpdate{
			Resources: r,
			Changeset: cs,
			Sequence:  seq,
		}
	}
	return seq
}

// Start watching the resources and send updates to the onChange channel.
// If an error occurs, it is sent to the onError channel.
// The context can be used to stop the watcher.
// This function blocks until the context is canceled.
func (w *ResourcesWatcher) Start(ctx context.Context, onChange chan<- ResourceUpdate, onError chan<- error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(w.cfg.PollIntervalOrDefault())
	defer ticker.Stop()
	defer close(onChange)
	defer close(onError)

	slog.DebugContext(ctx, "Watcher performing initial resources update.")
	w.performUpdate(ctx, onChange, onError, false)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.DebugContext(ctx, "Watcher polling resources.")
			w.performUpdate(ctx, onChange, onError, false)
		}
	}
}
