package embedded

import (
	"context"
	"maps"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	// In testing, 5 ms grouped concurrent reads while keeping the wait for an
	// isolated read short. A longer window only helps a partial batch if
	// another read can start before the timer fires.
	secretReadDebounceWindow = 5 * time.Millisecond

	// Terraform defaults to ten concurrent operations. When ten distinct reads
	// are waiting, one must finish before the graph can start another.
	secretReadBatchSize = 10
)

// A batcher belongs to one authenticated Secrets Manager client. It only
// groups concurrent reads, and retains no results for subsequent batches.
type secretReadBatcher struct {
	client webapi.Client

	mu      sync.Mutex
	pending *secretReadBatch
}

type secretReadBatch struct {
	ctx    context.Context
	cancel context.CancelFunc
	timer  *time.Timer
	done   chan struct{}

	// Number of callers waiting for each ID, protected by the batcher mutex.
	ids map[string]int

	// The worker publishes these before closing done.
	results map[string]*webapi.Secret
	err     error
}

func (b *secretReadBatcher) get(ctx context.Context, id string) (*webapi.Secret, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	batch := b.enqueue(ctx, id)
	defer b.release(batch, id)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-batch.done:
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if batch.err != nil {
		return nil, batch.err
	}

	if secret := batch.results[id]; secret != nil && len(secret.Projects) > 0 && secret.Projects[0].ID != "" {
		return secret, nil
	}

	// The single endpoint also handles a lone ID, unsupported batch endpoints,
	// and records whose metadata is incomplete in the batch response.
	return b.client.GetSecret(ctx, id)
}

func (b *secretReadBatcher) enqueue(ctx context.Context, id string) *secretReadBatch {
	b.mu.Lock()
	defer b.mu.Unlock()

	batch := b.pending
	if batch == nil {
		// Keep Terraform's log context, but end the shared call only when its
		// last waiter leaves. Each waiter still observes its own deadline.
		batchCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
		batch = &secretReadBatch{
			ctx:    batchCtx,
			cancel: cancel,
			done:   make(chan struct{}),
			ids:    make(map[string]int),
		}
		batch.timer = time.AfterFunc(secretReadDebounceWindow, func() {
			b.flush(batch)
		})
		b.pending = batch
	}

	batch.ids[id]++

	if len(batch.ids) == secretReadBatchSize {
		b.dispatchLocked(batch)
	}

	return batch
}

func (b *secretReadBatcher) flush(batch *secretReadBatch) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.pending == batch {
		b.dispatchLocked(batch)
	}
}

func (b *secretReadBatcher) dispatchLocked(batch *secretReadBatch) {
	batch.timer.Stop()
	b.pending = nil

	if len(batch.ids) == 1 {
		batch.cancel()
		close(batch.done)

		return
	}

	ids := slices.Collect(maps.Keys(batch.ids))
	go b.readBatch(batch, ids)
}

func (b *secretReadBatcher) release(batch *secretReadBatch, id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	batch.ids[id]--
	if batch.ids[id] == 0 {
		delete(batch.ids, id)
	}

	if len(batch.ids) > 0 {
		return
	}

	if b.pending == batch {
		batch.timer.Stop()
		b.pending = nil
	}

	batch.cancel()
}

func (b *secretReadBatcher) readBatch(batch *secretReadBatch, ids []string) {
	defer close(batch.done)
	defer batch.cancel()

	page, err := b.client.GetSecretsByIDs(batch.ctx, ids)
	if err != nil {
		if !retryAsSingleReads(err) {
			batch.err = err
		}

		return
	}

	if page.ContinuationToken != nil && *page.ContinuationToken != "" {
		tflog.Warn(batch.ctx, "Retrying paginated Bitwarden get-by-ids response with individual GETs")
		return
	}

	results := make(map[string]*webapi.Secret, len(page.Data))

	for i := range page.Data {
		secret := &page.Data[i]
		if !slices.Contains(ids, secret.ID) || results[secret.ID] != nil {
			return
		}

		results[secret.ID] = secret
	}

	batch.results = results
}

func retryAsSingleReads(err error) bool {
	httpErr, ok := webapi.IsHTTPError(err)
	if !ok {
		return false
	}

	switch httpErr.GetStatusCode() {
	case http.StatusBadRequest, http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusNotImplemented:
		// Bitwarden returns 404 for the entire batch if any ID cannot be read.
		// Servers predating Bitwarden v2023.9.0 may not have the endpoint.
		return true
	default:
		return false
	}
}
