//go:build offline

package embedded

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/uuid"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
	"github.com/stretchr/testify/require"
)

const (
	firstSecretID  = "aaaaaaaa-0000-4000-8000-000000000001"
	secondSecretID = "bbbbbbbb-0000-4000-8000-000000000002"
)

func TestSecretReadBatcherGroupsConcurrentReads(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		batchIDs := make(chan []string, 1)
		client := &testSecretReadClient{
			batch: func(_ context.Context, ids []string) ([]webapi.Secret, error) {
				batchIDs <- ids

				return []webapi.Secret{testSecret(secondSecretID), testSecret(firstSecretID)}, nil
			},
			single: func(_ context.Context, id string) (*webapi.Secret, error) {
				secret := testSecret(id)
				secret.Value = "updated"

				return &secret, nil
			},
		}
		batcher := &secretReadBatcher{client: client}

		first := readSecretAsync(t.Context(), batcher, firstSecretID)
		duplicate := readSecretAsync(t.Context(), batcher, strings.ToUpper(firstSecretID))
		second := readSecretAsync(t.Context(), batcher, secondSecretID)

		require.Equal(t, testSecret(firstSecretID), takeSecret(t, first))
		require.Equal(t, testSecret(firstSecretID), takeSecret(t, duplicate))
		require.Equal(t, testSecret(secondSecretID), takeSecret(t, second))
		require.ElementsMatch(t, []string{firstSecretID, secondSecretID}, <-batchIDs)

		updated, err := batcher.get(t.Context(), firstSecretID)
		require.NoError(t, err)
		require.Equal(t, "updated", updated.Value, "a later read must reach the server")
	})
}

func TestSecretReadBatcherDispatchesFullBatchImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dispatchedAt := make(chan time.Time, 1)
		client := &testSecretReadClient{
			batch: func(_ context.Context, ids []string) ([]webapi.Secret, error) {
				dispatchedAt <- time.Now()

				secrets := make([]webapi.Secret, 0, len(ids))
				for _, id := range ids {
					secrets = append(secrets, testSecret(id))
				}

				return secrets, nil
			},
		}
		batcher := &secretReadBatcher{client: client}

		start := time.Now()
		reads := make([]<-chan secretReadResult, 0, secretReadBatchSize)

		for range secretReadBatchSize {
			reads = append(reads, readSecretAsync(t.Context(), batcher, uuid.NewString()))
		}

		require.Equal(t, start, <-dispatchedAt)

		for _, read := range reads {
			takeSecret(t, read)
		}
	})
}

func TestSecretReadBatcherFallsBackAfterNotFound(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		notFound := &webapi.HTTPError{StatusCode: http.StatusNotFound}
		calls := make(chan string, 3)
		client := &testSecretReadClient{
			batch: func(_ context.Context, _ []string) ([]webapi.Secret, error) {
				calls <- "batch"

				return nil, notFound
			},
			single: func(_ context.Context, id string) (*webapi.Secret, error) {
				calls <- id

				if id == secondSecretID {
					return nil, notFound
				}

				secret := testSecret(id)
				return &secret, nil
			},
		}
		batcher := &secretReadBatcher{client: client}

		present := readSecretAsync(t.Context(), batcher, firstSecretID)
		missing := readSecretAsync(t.Context(), batcher, secondSecretID)

		require.Equal(t, testSecret(firstSecretID), takeSecret(t, present))
		require.ErrorIs(t, (<-missing).err, notFound)

		require.Len(t, calls, 3)
		require.Equal(t, "batch", <-calls)
		require.ElementsMatch(t, []string{firstSecretID, secondSecretID}, []string{<-calls, <-calls})
	})
}

func TestSecretReadBatcherFallsBackForIncompleteProjectMetadata(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		firstSecret := testSecret(firstSecretID)
		secondSecret := testSecret(secondSecretID)
		secondSecret.Projects = nil

		singleIDs := make(chan string, 2)
		client := &testSecretReadClient{
			batch: func(_ context.Context, _ []string) ([]webapi.Secret, error) {
				return []webapi.Secret{secondSecret, firstSecret}, nil
			},
			single: func(_ context.Context, id string) (*webapi.Secret, error) {
				singleIDs <- id

				secret := testSecret(id)
				return &secret, nil
			},
		}
		batcher := &secretReadBatcher{client: client}

		first := readSecretAsync(t.Context(), batcher, firstSecretID)
		second := readSecretAsync(t.Context(), batcher, secondSecretID)

		require.Equal(t, firstSecret, takeSecret(t, first))
		require.Equal(t, testSecret(secondSecretID), takeSecret(t, second))
		require.Len(t, singleIDs, 1)
		require.Equal(t, secondSecretID, <-singleIDs)
	})
}

func TestSecretReadBatcherDoesNotFanOutRateLimitErrors(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		throttled := &webapi.HTTPError{StatusCode: http.StatusTooManyRequests}
		client := &testSecretReadClient{
			batch: func(_ context.Context, _ []string) ([]webapi.Secret, error) {
				return nil, throttled
			},
		}
		batcher := &secretReadBatcher{client: client}

		first := readSecretAsync(t.Context(), batcher, firstSecretID)
		second := readSecretAsync(t.Context(), batcher, secondSecretID)

		require.ErrorIs(t, (<-first).err, throttled)
		require.ErrorIs(t, (<-second).err, throttled)
	})
}

func TestSecretReadBatcherCancelsSharedReadWhenLastCallerLeaves(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		batchContext := make(chan context.Context, 1)
		client := &testSecretReadClient{
			batch: func(ctx context.Context, _ []string) ([]webapi.Secret, error) {
				batchContext <- ctx
				<-ctx.Done()

				return nil, ctx.Err()
			},
		}
		batcher := &secretReadBatcher{client: client}

		firstCtx, cancelFirst := context.WithCancel(t.Context())
		defer cancelFirst()

		secondCtx, cancelSecond := context.WithCancel(t.Context())
		defer cancelSecond()

		first := readSecretAsync(firstCtx, batcher, firstSecretID)
		synctest.Wait()
		second := readSecretAsync(secondCtx, batcher, secondSecretID)

		shared := <-batchContext
		cancelFirst()

		require.ErrorIs(t, (<-first).err, context.Canceled)
		require.NoError(t, shared.Err())

		cancelSecond()

		require.ErrorIs(t, (<-second).err, context.Canceled)
		require.ErrorIs(t, shared.Err(), context.Canceled)
	})
}

type testSecretReadClient struct {
	webapi.Client
	batch  func(context.Context, []string) ([]webapi.Secret, error)
	single func(context.Context, string) (*webapi.Secret, error)
}

func (c *testSecretReadClient) GetSecret(ctx context.Context, id string) (*webapi.Secret, error) {
	return c.single(ctx, id)
}

func (c *testSecretReadClient) GetSecretsByIDs(ctx context.Context, ids []string) ([]webapi.Secret, error) {
	return c.batch(ctx, ids)
}

type secretReadResult struct {
	secret *webapi.Secret
	err    error
}

func readSecretAsync(ctx context.Context, batcher *secretReadBatcher, id string) <-chan secretReadResult {
	result := make(chan secretReadResult, 1)

	go func() {
		secret, err := batcher.get(ctx, id)
		result <- secretReadResult{secret: secret, err: err}
	}()

	return result
}

func takeSecret(t *testing.T, read <-chan secretReadResult) webapi.Secret {
	t.Helper()

	result := <-read
	require.NoError(t, result.err)
	require.NotNil(t, result.secret)

	return *result.secret
}

func testSecret(id string) webapi.Secret {
	return webapi.Secret{
		SecretSummary: webapi.SecretSummary{
			ID:       id,
			Projects: []models.Project{{ID: "project"}},
		},
		Value: id,
	}
}
