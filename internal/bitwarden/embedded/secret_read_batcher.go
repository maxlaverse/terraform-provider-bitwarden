package embedded

import (
	"context"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	secretReadBatchWindow = 5 * time.Millisecond
	secretReadBatchSize   = 10
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

	// The batcher mutex protects the waiter counts. Once a batch is dispatched,
	// ids is immutable and the worker can read it without holding the mutex.
	ids     map[string]int
	waiters int

	// The worker publishes these before closing done.
	results map[string]*webapi.Secret
	err     error
}

func (b *secretReadBatcher) get(ctx context.Context, id string) (*webapi.Secret, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	parsedID, err := uuid.Parse(id)
	if err != nil || !strings.EqualFold(id, parsedID.String()) {
		return b.client.GetSecret(ctx, id)
	}

	canonicalID := parsedID.String()
	batch := b.enqueue(ctx, canonicalID)
	defer b.release(batch, canonicalID)

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

	if secret := batch.results[canonicalID]; secret != nil && len(secret.Projects) > 0 && secret.Projects[0].ID != "" {
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
		batch.timer = time.AfterFunc(secretReadBatchWindow, func() {
			b.flush(batch)
		})
		b.pending = batch
	}

	batch.ids[id]++
	batch.waiters++

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

	go b.readBatch(batch)
}

func (b *secretReadBatcher) release(batch *secretReadBatch, id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	batch.waiters--

	if b.pending == batch {
		batch.ids[id]--
		if batch.ids[id] == 0 {
			delete(batch.ids, id)
		}

		if batch.waiters == 0 {
			batch.timer.Stop()
			b.pending = nil
		}
	}

	if batch.waiters == 0 {
		batch.cancel()
	}
}

func (b *secretReadBatcher) readBatch(batch *secretReadBatch) {
	defer close(batch.done)
	defer batch.cancel()

	ids := slices.Collect(maps.Keys(batch.ids))

	secrets, err := b.client.GetSecretsByIDs(batch.ctx, ids)
	if err != nil {
		if !retryAsSingleReads(err) {
			batch.err = err
		}

		return
	}

	results := make(map[string]*webapi.Secret, len(secrets))

	for i := range secrets {
		secret := &secrets[i]
		id := strings.ToLower(secret.ID)

		if batch.ids[id] == 0 || results[id] != nil {
			return
		}

		results[id] = secret
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
		// Older servers can reject the endpoint or the request entirely.
		return true
	default:
		return false
	}
}
