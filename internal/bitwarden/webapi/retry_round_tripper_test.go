//go:build offline

package webapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryRoundTripper_Backoff(t *testing.T) {
	// retryBackoffFactor is 1.5: delay = 1.5^attempt seconds, capped at 30s
	testData := map[int]time.Duration{
		1:  1 * time.Second,
		2:  2 * time.Second,
		3:  3 * time.Second,
		4:  5 * time.Second,
		5:  7 * time.Second,
		6:  11 * time.Second,
		7:  17 * time.Second,
		8:  25 * time.Second,
		9:  30 * time.Second,
		10: 30 * time.Second,
	}
	for attempt, expected := range testData {
		assert.Equal(t, expected, backoff(attempt))
	}
}

func TestRetryRoundTripper_BasicRetry(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Create a transport that fails once then succeeds
	transport := &mockTransport{
		responses: []*http.Response{
			nil,
			{StatusCode: http.StatusOK},
		},
		errors: []error{
			&net.OpError{Op: "dial", Err: errors.New("connection refused")},
			nil,
		},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.index)
}

func TestRetryRoundTripper_ServiceUnavailable(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Create a transport that returns 503 then succeeds
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable},
			{StatusCode: http.StatusOK},
		},
		errors: []error{nil, nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.index)
}

func TestRetryRoundTripper_4xxWithGET(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Create a transport that returns 429 (4xx) then succeeds
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusTooManyRequests},
			{StatusCode: http.StatusOK},
		},
		errors: []error{nil, nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.index)
}

func TestRetryRoundTripper_4xxWithPOST(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Create a transport that returns 429 (4xx) then succeeds
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusTooManyRequests},
			{StatusCode: http.StatusOK},
		},
		errors: []error{nil, nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("POST", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.index)
}

func TestRetryRoundTripper_4xxWithPOSTRewindsBody(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Transport that returns 429 then 200 and records the body received on each attempt.
	var bodies [][]byte
	var bodiesMu sync.Mutex
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusTooManyRequests},
			{StatusCode: http.StatusOK},
		},
		errors: []error{nil, nil},
	}
	recordBodyTransport := &roundTripFunc{
		fn: func(req *http.Request) (*http.Response, error) {
			var body []byte
			if req.Body != nil {
				body, _ = io.ReadAll(req.Body)
			}
			bodiesMu.Lock()
			bodies = append(bodies, body)
			bodiesMu.Unlock()
			return transport.RoundTrip(req)
		},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = recordBodyTransport

	body := []byte("grant_type=password&username=foo@example.com")
	req, err := http.NewRequest("POST", "http://example.com", bytes.NewReader(body))
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.index)
	bodiesMu.Lock()
	defer bodiesMu.Unlock()
	require.Len(t, bodies, 2, "body should have been sent on both attempts")
	assert.Equal(t, body, bodies[0], "first attempt should send full body")
	assert.Equal(t, body, bodies[1], "retry should send same body (rewound)")
}

func TestRetryRoundTripper_429RetriedIndefinitely(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// With maxRetries=3 we'd normally give up after 2 attempts. For 429 we retry
	// until success. Return 429 three times then 200; we should get 200 on attempt 4.
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusTooManyRequests},
			{StatusCode: http.StatusTooManyRequests},
			{StatusCode: http.StatusTooManyRequests},
			{StatusCode: http.StatusOK},
		},
		errors: []error{nil, nil, nil, nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("POST", "http://example.com", bytes.NewReader([]byte("body")))
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 4, transport.index, "should have retried 429 more than maxRetries times until success")
}

func TestRetryRoundTripper_4xxWithGETNotInRetryableList(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Create a transport that returns 400 (4xx, not in retryableStatusCodes) - should not retry
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusBadRequest},
		},
		errors: []error{nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, 1, transport.index)
}

func TestRetryRoundTripper_5xxWithGet(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Create a transport that returns 503 (5xx, in retryableStatusCodes) - should retry
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable},
			{StatusCode: http.StatusOK},
		},
		errors: []error{nil, nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.index)
}

func TestRetryRoundTripper_5xxWithPOST(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Create a transport that returns 503 (5xx) with POST - should not retry because POST is not retryable
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable},
		},
		errors: []error{nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("POST", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, 1, transport.index)
}

func TestRetryRoundTripper_5xxWithGETNotInRetryableList(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Create a transport that returns 500 (5xx, not in retryableStatusCodes) - should not retry
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusInternalServerError},
		},
		errors: []error{nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, 1, transport.index)
}

func TestRetryRoundTripper_ConcurrentRequests(t *testing.T) {
	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusOK},
			{StatusCode: http.StatusOK},
			{StatusCode: http.StatusOK},
		},
		errors: []error{nil, nil, nil},
	}

	rrt := NewRetryRoundTripper(2, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	// Create a channel to coordinate the requests
	start := make(chan struct{})
	done := make(chan struct{})

	// Launch 3 concurrent requests
	for i := 0; i < 3; i++ {
		go func() {
			<-start
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			rrt.RoundTrip(req)
			done <- struct{}{}
		}()
	}

	// Start all requests at once
	close(start)

	// Wait for all requests to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify that all requests were processed
	assert.Equal(t, 3, transport.index)
}

func TestRetryRoundTripper_RequestTimeout(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	transport := &mockTransport{
		responses: []*http.Response{nil},
		errors:    []error{context.DeadlineExceeded},
	}

	rrt := NewRetryRoundTripper(1, 3, 100*time.Millisecond, 100*time.Millisecond, 100*time.Millisecond, 100*time.Millisecond)
	rrt.Transport = transport

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)
	req = req.WithContext(t.Context())

	resp, err := rrt.RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, errNoMoreResponses))
	// Context timeout is retried for GET; we retried once and the second attempt hit the mock with no more responses.
	assert.Equal(t, 2, transport.index)
}

func TestRetryRoundTripper_DisableRetries(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	transport := &mockTransport{
		responses: []*http.Response{
			{StatusCode: http.StatusServiceUnavailable},
		},
		errors: []error{nil},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport
	rrt.DisableRetries = true

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, 1, transport.index)
}

func TestRetryRoundTripper_NonGetRequest(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	transport := &mockTransport{
		responses: []*http.Response{nil},
		errors:    []error{&net.OpError{Op: "dial", Err: errors.New("connection refused")}},
	}

	rrt := NewRetryRoundTripper(1, 3, time.Second, time.Second, time.Second, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("POST", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	// Dial errors are retried for all methods; we retry once then hit "no more responses".
	assert.Equal(t, 2, transport.index)
}

func TestRetryRoundTripper_IsConnectTimeout(t *testing.T) {
	lowerBackoffFactor()
	defer lowerBackoffFactor()

	// Test with a temporary error
	tempErr := &net.OpError{
		Op:  "dial",
		Err: &os.SyscallError{Syscall: "connect", Err: syscall.EAGAIN},
	}
	assert.True(t, isDialError(tempErr))

	// Test with a non-temporary error
	nonTempErr := &net.OpError{
		Op:  "dial",
		Err: &os.SyscallError{Syscall: "connect", Err: syscall.ECONNREFUSED},
	}
	assert.True(t, isDialError(nonTempErr))

	// Test with a non-dial error
	readErr := &net.OpError{
		Op:  "read",
		Err: &os.SyscallError{Syscall: "read", Err: syscall.EAGAIN},
	}
	assert.False(t, isDialError(readErr))

	// Test with nil
	assert.False(t, isDialError(nil))
}

func TestRetryRoundTripper_ReadWaitDurationFromHeaders(t *testing.T) {

	testData := []struct {
		header   http.Header
		duration time.Duration
	}{
		{
			header:   http.Header{"X-Retry-After": []string{"1"}},
			duration: 1 * time.Minute,
		},
		{
			header:   http.Header{"X-Rate-Limit-Reset": []string{time.Now().Add(15 * time.Second).UTC().Format(time.RFC3339)}, "X-Retry-After": []string{"1"}},
			duration: 15 * time.Second,
		},
	}

	for _, test := range testData {
		duration := tryToReadWaitDurationFromHeaders(&http.Response{StatusCode: http.StatusTooManyRequests, Header: test.header})
		assert.True(t, math.Abs(float64(duration-test.duration)) <= float64(time.Second),
			"Duration %v should be within 1 second of expected %v", duration, test.duration)
	}
}

type roundTripFunc struct {
	fn func(*http.Request) (*http.Response, error)
}

func (r *roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r.fn(req)
}

var errNoMoreResponses = errors.New("no more responses")

type mockTransport struct {
	responses []*http.Response
	errors    []error
	index     int
	mu        sync.Mutex
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.index >= len(m.responses) {
		m.index++
		return nil, errNoMoreResponses
	}
	resp := m.responses[m.index]
	err := m.errors[m.index]
	m.index++
	return resp, err
}

func lowerBackoffFactor() func() {
	originalBackoffFactor := retryBackoffFactor
	retryBackoffFactor = 0
	return func() {
		retryBackoffFactor = originalBackoffFactor
	}
}
