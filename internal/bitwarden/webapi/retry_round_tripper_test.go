//go:build offline

package webapi

import (
	"context"
	"errors"
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
	testData := map[int]time.Duration{
		1: 2 * time.Second,
		2: 4 * time.Second,
		3: 8 * time.Second,
		4: 16 * time.Second,
		5: 30 * time.Second,
		6: 30 * time.Second,
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("POST", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, transport.index)
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
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

	rrt := NewRetryRoundTripper(2, 3, time.Second)
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

	rrt := NewRetryRoundTripper(1, 3, 100*time.Millisecond)
	rrt.Transport = transport

	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	// Create a context with a shorter timeout
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := rrt.RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "context deadline exceeded")
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
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

	rrt := NewRetryRoundTripper(1, 3, time.Second)
	rrt.Transport = transport

	req, err := http.NewRequest("POST", "http://example.com", nil)
	require.NoError(t, err)

	resp, err := rrt.RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 1, transport.index)
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
		return nil, errors.New("no more responses")
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
