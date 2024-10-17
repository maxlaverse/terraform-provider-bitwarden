package webapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/sync/semaphore"
)

type RetryRoundTripper struct {
	DisableRetries bool
	Transport      http.RoundTripper

	concurrentRequestsSem *semaphore.Weighted
	maxLowLevelRetries    int
	requestTimeout        time.Duration
}

// maxLowLevelRetries is the maximum number of retries for low-level errors (e.g. timeouts).
// A value of 0 means no retries.
func NewRetryRoundTripper(maxConcurrentRequests int, maxLowLevelRetries int, requestTimeout time.Duration) *RetryRoundTripper {
	return &RetryRoundTripper{
		Transport:             http.DefaultTransport,
		concurrentRequestsSem: semaphore.NewWeighted(int64(maxConcurrentRequests)),
		maxLowLevelRetries:    maxLowLevelRetries,
		requestTimeout:        requestTimeout,
	}
}

func (rrt *RetryRoundTripper) RoundTrip(httpReq *http.Request) (*http.Response, error) {
	err := rrt.concurrentRequestsSem.Acquire(httpReq.Context(), 1)
	if err == nil {
		defer rrt.concurrentRequestsSem.Release(1)
	}

	ctx := httpReq.Context()
	attemptNumber := 0
	for {
		attemptNumber += 1

		resp, shouldRetry, err := rrt.doRequest(ctx, httpReq, attemptNumber)
		if err != nil {
			return nil, err
		}

		if !shouldRetry || rrt.DisableRetries {
			return resp, nil
		}
	}
}

func (rrt *RetryRoundTripper) doRequest(ctx context.Context, httpReq *http.Request, attemptNumber int) (*http.Response, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, rrt.requestTimeout)
	defer cancel()

	resp, err := rrt.Transport.RoundTrip(httpReq.WithContext(ctx))

	// Successfully got an HTTP response that is not a 429
	if err == nil && resp.StatusCode != http.StatusTooManyRequests {
		// We read the body as we're cancelling the context when leaving the function.
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return resp, false, fmt.Errorf("failed to read response body in round tripper: %w", err)
		}

		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, false, nil
	}

	// Got a low-level error, or a 429 HTTP response. We're returning the error
	// so let's not log anything additional.
	if !rrt.isRetryableError(err, attemptNumber, httpReq.Method) {
		return nil, false, err
	}

	// Retryable request that had a response. A body is present, throw it away.
	io.ReadAll(resp.Body)
	resp.Body.Close()

	var waitDuration time.Duration
	debugInfo := map[string]interface{}{
		"url":            httpReq.URL.RequestURI(),
		"method":         httpReq.Method,
		"attempt_number": attemptNumber,
		"is_retryable":   true,
	}

	if err != nil {
		debugInfo["error"] = err
		waitDuration = backoff(attemptNumber)
	} else if resp.StatusCode == http.StatusTooManyRequests {
		debugInfo["status_code"] = resp.StatusCode
		debugInfo["status_message"] = resp.Status
		waitDuration = tryToReadWaitDurationFromHeaders(resp)
	}

	debugInfo["wait_duration_sec"] = waitDuration.Seconds()
	tflog.Info(ctx, "retry_round_tripper", debugInfo)

	return resp, true, sleepWithContext(ctx, waitDuration)
}

func (rrt *RetryRoundTripper) isRetryableError(err error, attemptNumber int, httpMethod string) bool {
	if err != nil {
		return false
	}

	if attemptNumber >= rrt.maxLowLevelRetries-1 {
		return false
	}
	if isConnectTimeout(err) {
		return true
	}
	if isReadTimeout(err) && httpMethod == http.MethodGet {
		return true
	}
	return false
}

func tryToReadWaitDurationFromHeaders(resp *http.Response) time.Duration {
	retryAfterRaw := resp.Header.Get("X-Retry-After")
	if len(retryAfterRaw) != 0 {
		retryAfter, err := strconv.ParseInt(retryAfterRaw, 10, 64)
		if err == nil {
			return time.Minute * time.Duration(retryAfter)
		}
	}
	return 0
}

func isConnectTimeout(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		opErr, ok := netErr.(*net.OpError)
		if ok && opErr.Op == "dial" {
			return true
		}
	}
	return false
}

func isReadTimeout(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		opErr, ok := netErr.(*net.OpError)
		if ok && opErr.Op == "read" {
			return true
		}
	}
	return false
}

func backoff(attempt int) time.Duration {
	maxInterval := 30 * time.Second
	delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	if delay > maxInterval {
		delay = maxInterval
	}

	return delay
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("sleep cancelled: %v", ctx.Err())
	case <-time.After(duration):
		return nil
	}
}
