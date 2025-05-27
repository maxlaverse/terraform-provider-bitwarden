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
	"slices"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/sync/semaphore"
)

type RetryRoundTripper struct {
	DisableRetries bool
	Transport      http.RoundTripper

	concurrentRequestsSem *semaphore.Weighted
	maxRetries            int
	requestTimeout        time.Duration
}

// List of status codes that are considered retryable depending on the method.
var retryableStatusCodes = []int{
	http.StatusTooManyRequests,
	http.StatusServiceUnavailable,
}

var retryBackoffFactor = 2

// NewRetryRoundTripper creates a new retry-capable HTTP transport.
//
// Parameters:
//   - maxConcurrentRequests: Maximum number of concurrent requests allowed
//   - maxRetries: Maximum number of retries for low-level errors (e.g. timeouts) and
//     bad status codes (e.g. 503). A value of 0 means no retries.
//   - requestTimeout: Timeout duration that applies to individual request attempts
func NewRetryRoundTripper(maxConcurrentRequests int, maxRetries int, requestTimeout time.Duration) *RetryRoundTripper {
	return &RetryRoundTripper{
		Transport:             http.DefaultTransport,
		concurrentRequestsSem: semaphore.NewWeighted(int64(maxConcurrentRequests)),
		maxRetries:            maxRetries,
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

		if !shouldRetry {
			return resp, nil
		}
	}
}

func (rrt *RetryRoundTripper) doRequest(originalCtx context.Context, httpReq *http.Request, attemptNumber int) (*http.Response, bool, error) {
	reqCtx, cancel := context.WithTimeout(originalCtx, rrt.requestTimeout)
	defer cancel()

	resp, err := rrt.Transport.RoundTrip(httpReq.WithContext(reqCtx))

	isSuccessful := err == nil && !slices.Contains(retryableStatusCodes, resp.StatusCode)

	// If the request was successful, we return without additional logging.
	// We preserve the response body as exiting the method cancels the context.
	if isSuccessful {
		return resp, false, preserveResponseBody(resp)
	}

	// At this point, there was a problem. We need to check if it's retryable,
	// and we want to log information in any case.
	isDialError := err != nil && isDialError(err)
	isRetriableHttpStatusCode := httpReq.Method == http.MethodGet && err == nil && slices.Contains(retryableStatusCodes, resp.StatusCode)
	isRetriableReadTimeout := httpReq.Method == http.MethodGet && (isReadTimeout(err))
	isLastPossibleAttempt := attemptNumber >= rrt.maxRetries-1 || rrt.DisableRetries

	debugInfo := map[string]interface{}{
		"url":                           httpReq.URL.RequestURI(),
		"method":                        httpReq.Method,
		"attempt_number":                attemptNumber,
		"is_last_attempt":               isLastPossibleAttempt,
		"is_retriable_http_status_code": isRetriableHttpStatusCode,
		"is_retriable_read_timeout":     isRetriableReadTimeout,
	}

	// If the request is not retryable, we preserve the response body, log and return.
	if isLastPossibleAttempt || (!isDialError && !isRetriableHttpStatusCode && !isRetriableReadTimeout) {
		tflog.Info(originalCtx, "retry_round_tripper", debugInfo)
		readErr := preserveResponseBody(resp)
		if readErr != nil && err != nil {
			err = fmt.Errorf("%w (and additionally %w)", err, readErr)
		} else if readErr != nil {
			err = readErr
		}
		return resp, false, err
	}

	// We're going to retry the request, and therefore should throw away the
	// response body of the previous attempt if it exists.
	if resp != nil && resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	if err != nil {
		debugInfo["error"] = err
	}
	if resp != nil {
		debugInfo["status_code"] = resp.StatusCode
		debugInfo["status_message"] = resp.Status
	}

	// Try to find the best waiting duration, either from the response headers
	// or from the backoff function.
	var waitDuration time.Duration
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		waitDuration = tryToReadWaitDurationFromHeaders(resp)
	} else {
		waitDuration = backoff(attemptNumber)
	}

	debugInfo["wait_duration_sec"] = waitDuration.Seconds()
	tflog.Info(originalCtx, "retry_round_tripper", debugInfo)

	return nil, true, sleepWithContext(originalCtx, waitDuration)
}

func tryToReadWaitDurationFromHeaders(resp *http.Response) time.Duration {
	rateLimitResetRaw := resp.Header.Get("X-Rate-Limit-Reset")

	if len(rateLimitResetRaw) != 0 {
		resetTime, err := time.Parse(time.RFC3339Nano, rateLimitResetRaw)
		if err == nil {
			waitDuration := time.Until(resetTime)
			if waitDuration > 0 {
				return waitDuration
			}
		}
	}

	retryAfterRaw := resp.Header.Get("X-Retry-After")
	if len(retryAfterRaw) != 0 {
		retryAfter, err := strconv.ParseInt(retryAfterRaw, 10, 64)
		if err == nil {
			return time.Minute * time.Duration(retryAfter)
		}
	}
	return 0
}

func isDialError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
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
	delay := time.Duration(math.Pow(float64(retryBackoffFactor), float64(attempt))) * time.Second
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

// preserveResponseBody reads the response body and creates a new reader for it.
// This is necessary because the response body can only be read once.
func preserveResponseBody(resp *http.Response) error {
	if resp == nil || resp.Body == nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response body in round tripper: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return nil
}
