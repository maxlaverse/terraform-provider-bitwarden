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
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/sync/semaphore"
)

type RetryRoundTripper struct {
	DisableRetries bool
	Transport      http.RoundTripper

	concurrentRequestsSem *semaphore.Weighted
	maxRetries            int
	attemptTimeout        time.Duration
}

// List of status codes that are considered retryable depending on the method.
var retryableStatusCodes = []int{
	http.StatusTooManyRequests,
	http.StatusServiceUnavailable,
	http.StatusInternalServerError,
}

var retryBackoffFactor = 1.5

// NewRetryRoundTripper creates a new retry-capable HTTP transport.
//
// Parameters:
//   - maxConcurrentRequests: Maximum number of concurrent requests allowed
//   - maxRetries: Maximum number of retries for low-level errors (e.g. timeouts) and
//     bad status codes (e.g. 503). A value of 0 means no retries.
//   - attemptTimeout: Timeout duration that applies to the per-attempt context
//     (overall bound for a single attempt, includes everything).
//   - dialTimeout: Timeout for DNS + TCP connect. 0 means no timeout (infinite).
//   - tlsHandshakeTimeout: Timeout for TLS handshake. 0 means no timeout (infinite).
//   - responseHeaderTimeout: Timeout for reading response headers. 0 means no timeout (infinite).
func NewRetryRoundTripper(maxConcurrentRequests int, maxRetries int, attemptTimeout, dialTimeout, tlsHandshakeTimeout, responseHeaderTimeout time.Duration) *RetryRoundTripper {
	return &RetryRoundTripper{
		Transport:             newBaseTransport(dialTimeout, tlsHandshakeTimeout, responseHeaderTimeout),
		concurrentRequestsSem: semaphore.NewWeighted(int64(maxConcurrentRequests)),
		maxRetries:            maxRetries,
		attemptTimeout:        attemptTimeout,
	}
}

func (rrt *RetryRoundTripper) RoundTrip(httpReq *http.Request) (*http.Response, error) {
	err := rrt.concurrentRequestsSem.Acquire(httpReq.Context(), 1)
	if err == nil {
		defer rrt.concurrentRequestsSem.Release(1)
	}

	ctx := httpReq.Context()

	// Buffer request body for potential retries. POST requests can get 429/503
	// and be retried; once the body is read by the transport it cannot be read again.
	var bodyBuf []byte
	if httpReq.Body != nil && httpReq.Method == http.MethodPost {
		var readErr error
		bodyBuf, readErr = io.ReadAll(httpReq.Body)
		httpReq.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("buffer request body for retry: %w", readErr)
		}
		httpReq.Body = io.NopCloser(bytes.NewReader(bodyBuf))
	}

	attemptNumber := 0
	for {
		attemptNumber += 1

		// Restore body before each retry so the transport can read it again.
		if attemptNumber > 1 && len(bodyBuf) > 0 {
			httpReq.Body = io.NopCloser(bytes.NewReader(bodyBuf))
		}

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
	reqCtx, cancel := context.WithTimeout(originalCtx, rrt.attemptTimeout)
	defer cancel()

	resp, err := rrt.Transport.RoundTrip(httpReq.WithContext(reqCtx))

	// A request is successful if there's no error, we got a response, and the status code is not retryable
	isSuccessful := err == nil && resp != nil && !isRetriableStatusCode(httpReq.Method, resp.StatusCode)

	// Success: return response, no retry.
	if isSuccessful {
		return resp, false, preserveResponseBody(resp)
	}

	// At this point, there was a problem. We need to check if it's retryable,
	// and we want to log information in any case.
	is429 := resp != nil && resp.StatusCode == http.StatusTooManyRequests
	// 429 is retried indefinitely (rate limits eventually reset). Other retryable cases use maxRetries.
	isLastPossibleAttempt := (attemptNumber >= rrt.maxRetries-1 || rrt.DisableRetries) && !is429
	reasons := retriableReasons(err, httpReq, resp)
	isRetriable := len(reasons) > 0
	giveUp := isLastPossibleAttempt || !isRetriable

	debugInfo := map[string]interface{}{
		"url":               httpReq.URL.RequestURI(),
		"method":            httpReq.Method,
		"attempt_number":    attemptNumber,
		"is_last_attempt":   isLastPossibleAttempt,
		"retriable_reasons": reasons,
	}

	// Give up: not retriable or last attempt; log and return.
	if giveUp {
		tflog.Info(originalCtx, "retry_round_tripper", debugInfo)
		readErr := preserveResponseBody(resp)
		if readErr != nil && err != nil {
			err = fmt.Errorf("%w (and additionally %w)", err, readErr)
		} else if readErr != nil {
			err = readErr
		}
		return resp, false, err
	}

	// Retry: discard response body, enrich debugInfo for retry log, then sleep before next attempt.
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
	waitDuration := backoff(attemptNumber)
	if is429 {
		if d := tryToReadWaitDurationFromHeaders(resp); d > 0 {
			waitDuration = d
		}
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

// isTLSHandshakeTimeout reports whether the error is from the transport's
// TLSHandshakeTimeout being exceeded (transient; retriable). Distinct from
// other TLS handshake failures (e.g. certificate verification, server alert)
// which are not retriable.
func isTLSHandshakeTimeout(err error) bool {
	for ; err != nil; err = errors.Unwrap(err) {
		if strings.Contains(err.Error(), "TLS handshake timeout") {
			return true
		}
	}
	return false
}

// isResponseHeaderTimeout reports whether the error is from the transport's
// ResponseHeaderTimeout being exceeded (e.g. "http2: timeout awaiting response headers").
// Transient and retriable for GET requests.
func isResponseHeaderTimeout(err error) bool {
	for ; err != nil; err = errors.Unwrap(err) {
		if strings.Contains(err.Error(), "timeout awaiting response headers") {
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

// isContextTimeout reports whether the error is due to the request context
// deadline being exceeded (e.g. our per-attempt request timeout).
func isContextTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

// isRetriableForResponsePhase reports whether response-phase timeouts (headers,
// read, or context) should be retried: GET always, or POST when unauthenticated
// (no Authorization header).
func isRetriableForResponsePhase(httpReq *http.Request) bool {
	if httpReq == nil {
		return false
	}
	if httpReq.Method == http.MethodGet {
		return true
	}
	if httpReq.Method == http.MethodPost && httpReq.Header.Get("Authorization") == "" {
		return true
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

// isRetriableStatusCode determines if a status code is retryable based on the HTTP method.
// - 4xx status codes: retryable for both GET and POST
// - 5xx status codes: retryable for GET only if the code is in retryableStatusCodes
func isRetriableStatusCode(method string, statusCode int) bool {
	if !slices.Contains(retryableStatusCodes, statusCode) {
		return false
	}

	switch {
	case statusCode >= 400 && statusCode < 500:
		return method == http.MethodGet || method == http.MethodPost
	case statusCode >= 500 && statusCode < 600:
		return method == http.MethodGet
	}
	return false
}

// retriableReasons returns the reasons why a request is retriable (e.g. dial_error, read_timeout).
// Empty slice means the request is not retriable.
func retriableReasons(err error, httpReq *http.Request, resp *http.Response) []string {
	var reasons []string
	if err != nil && isDialError(err) {
		reasons = append(reasons, "dial_error")
	}
	if err != nil && isTLSHandshakeTimeout(err) {
		reasons = append(reasons, "tls_handshake_timeout")
	}
	if isRetriableForResponsePhase(httpReq) && isReadTimeout(err) {
		reasons = append(reasons, "read_timeout")
	}
	if isRetriableForResponsePhase(httpReq) && isResponseHeaderTimeout(err) {
		reasons = append(reasons, "response_header_timeout")
	}
	if resp != nil && isRetriableStatusCode(httpReq.Method, resp.StatusCode) {
		reasons = append(reasons, "retriable_http_status_code")
	}
	if isRetriableForResponsePhase(httpReq) && isContextTimeout(err) {
		reasons = append(reasons, "context_timeout")
	}
	return reasons
}

// newBaseTransport returns an *http.Transport with the given timeouts.
// A zero value for any timeout means no timeout (infinite) for that phase.
func newBaseTransport(dialTimeout, tlsHandshakeTimeout, responseHeaderTimeout time.Duration) *http.Transport {
	base := http.DefaultTransport.(*http.Transport).Clone()
	if dialTimeout > 0 {
		base.DialContext = (&net.Dialer{Timeout: dialTimeout}).DialContext
	}
	if tlsHandshakeTimeout > 0 {
		base.TLSHandshakeTimeout = tlsHandshakeTimeout
	}
	if responseHeaderTimeout > 0 {
		base.ResponseHeaderTimeout = responseHeaderTimeout
	}
	return base
}
