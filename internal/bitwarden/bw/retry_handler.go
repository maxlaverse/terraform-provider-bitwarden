package bw

import (
	"math"
	"strings"
	"time"
)

const (
	rateLimitExceededError = "Rate limit exceeded."
)

type retryHandler struct {
	disableRetryBackoff bool
}

func (r *retryHandler) IsRetryable(err error, attempt int) bool {
	return strings.Contains(err.Error(), rateLimitExceededError) && attempt < 3
}

func (r *retryHandler) Backoff(attempt int) time.Duration {
	if r.disableRetryBackoff {
		return 0
	}

	maxInterval := 30 * time.Second
	delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	if delay > maxInterval {
		delay = maxInterval
	}

	time.Sleep(delay)
	return delay
}
