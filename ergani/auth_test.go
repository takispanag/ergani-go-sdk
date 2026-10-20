package ergani

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient builds a client with a short backoff so that the retry tests
// stay fast.
func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()

	client, err := NewClientWithConfig(Config{
		Username:     "testuser",
		Password:     "testpass",
		BaseURL:      baseURL,
		Timeout:      5 * time.Second,
		MaxRetries:   2,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client
}

func TestRequest_ReauthenticatesOnUnauthorized(t *testing.T) {
	var authCalls int32
	var documentCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			atomic.AddInt32(&authCalls, 1)
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "token-1"}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		atomic.AddInt32(&documentCalls, 1)
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"message": "token expired"}`)); err != nil {
			t.Errorf("Failed to write error response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	_, err := client.SubmitWorkCard(context.Background(), []CompanyWorkCard{})
	if err == nil {
		t.Fatal("Expected an error after the retries were exhausted, got nil")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("Expected the error to match ErrUnauthorized, got %v", err)
	}
	if !IsStatusCode(err, http.StatusUnauthorized) {
		t.Errorf("Expected the error to carry status 401, got %v", err)
	}

	// One authentication for the initial request plus one for every retry.
	if got := atomic.LoadInt32(&authCalls); got != 3 {
		t.Errorf("Expected 3 authentication calls, got %d", got)
	}
	if got := atomic.LoadInt32(&documentCalls); got != 3 {
		t.Errorf("Expected 3 document calls, got %d", got)
	}
}

func TestRequest_ReusesValidToken(t *testing.T) {
	var authCalls int32
	var documentCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			atomic.AddInt32(&authCalls, 1)
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "token-1", "expiresIn": 3600}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		atomic.AddInt32(&documentCalls, 1)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	for i := 0; i < 3; i++ {
		if _, err := client.SubmitWorkCard(context.Background(), []CompanyWorkCard{}); err != nil {
			t.Fatalf("SubmitWorkCard failed on attempt %d: %v", i, err)
		}
	}

	if got := atomic.LoadInt32(&authCalls); got != 1 {
		t.Errorf("Expected the token to be reused, got %d authentication calls", got)
	}
	if got := atomic.LoadInt32(&documentCalls); got != 3 {
		t.Errorf("Expected 3 document calls, got %d", got)
	}
}

func TestRequest_RetriesServerErrors(t *testing.T) {
	var documentCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "token-1"}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		if atomic.AddInt32(&documentCalls, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"message": "boom"}`)); err != nil {
				t.Errorf("Failed to write error response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[{"id": "sub1", "protocol": "proto1", "submitDate": "10/07/2025 14:56"}]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	responses, err := client.SubmitWorkCard(context.Background(), []CompanyWorkCard{})
	if err != nil {
		t.Fatalf("Expected the request to succeed after the retries, got: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("Expected 1 submission response, got %d", len(responses))
	}
	if got := atomic.LoadInt32(&documentCalls); got != 3 {
		t.Errorf("Expected 3 document calls, got %d", got)
	}
}

func TestRequest_RetriesTooManyRequests(t *testing.T) {
	var documentCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "token-1"}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		if atomic.AddInt32(&documentCalls, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write([]byte(`{"message": "slow down"}`)); err != nil {
				t.Errorf("Failed to write error response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	if _, err := client.SubmitWorkCard(context.Background(), []CompanyWorkCard{}); err != nil {
		t.Fatalf("Expected the request to succeed after the retry, got: %v", err)
	}
	if got := atomic.LoadInt32(&documentCalls); got != 2 {
		t.Errorf("Expected 2 document calls, got %d", got)
	}
}

func TestRequest_DoesNotRetryClientErrors(t *testing.T) {
	var documentCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "token-1"}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		atomic.AddInt32(&documentCalls, 1)
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"msg": "Invalid data provided"}`)); err != nil {
			t.Errorf("Failed to write error response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	_, err := client.SubmitWorkCard(context.Background(), []CompanyWorkCard{})
	if err == nil {
		t.Fatal("Expected an API error, got nil")
	}
	if !IsStatusCode(err, http.StatusBadRequest) {
		t.Errorf("Expected the error to carry status 400, got %v", err)
	}
	if got := atomic.LoadInt32(&documentCalls); got != 1 {
		t.Errorf("Expected a single document call, got %d", got)
	}
}

// flakyClient fails the first call at the transport level and then delegates
// to the default client.
type flakyClient struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (f *flakyClient) Do(req *http.Request) (*http.Response, error) {
	f.mu.Lock()
	f.calls++
	calls := f.calls
	f.mu.Unlock()

	if calls == 1 {
		return nil, f.err
	}

	return http.DefaultClient.Do(req)
}

func TestRequest_RetriesTransportErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "token-1"}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	flaky := &flakyClient{err: errors.New("connection reset")}

	client, err := NewClientWithConfig(Config{
		Username:     "testuser",
		Password:     "testpass",
		BaseURL:      server.URL,
		HTTPClient:   flaky,
		MaxRetries:   2,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if _, err := client.SubmitWorkCard(context.Background(), []CompanyWorkCard{}); err != nil {
		t.Fatalf("Expected the request to succeed after the retry, got: %v", err)
	}
}

func TestRequest_RespectsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "token-1"}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte(`{"message": "unavailable"}`)); err != nil {
			t.Errorf("Failed to write error response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClientWithConfig(Config{
		Username:     "testuser",
		Password:     "testpass",
		BaseURL:      server.URL,
		MaxRetries:   5,
		RetryBackoff: time.Hour,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = client.SubmitWorkCard(ctx, []CompanyWorkCard{})
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected the error to match context.DeadlineExceeded, got %v", err)
	}
}

func TestClient_ConcurrentUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Authentication" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"accessToken": "token-1"}`)); err != nil {
				t.Errorf("Failed to write auth response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`[]`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := client.SubmitWorkCard(context.Background(), []CompanyWorkCard{}); err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("Concurrent request failed: %v", err)
	}
}

func TestNewClientWithConfig_Defaults(t *testing.T) {
	client, err := NewClientWithConfig(Config{Username: "user", Password: "pass"})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.baseURL.String() != defaultBaseURL {
		t.Errorf("Expected the default base URL %q, got %q", defaultBaseURL, client.baseURL.String())
	}
	if client.maxRetries != DefaultMaxRetries {
		t.Errorf("Expected %d retries by default, got %d", DefaultMaxRetries, client.maxRetries)
	}
	if client.retryBackoff != DefaultRetryBackoff {
		t.Errorf("Expected a backoff of %s by default, got %s", DefaultRetryBackoff, client.retryBackoff)
	}
	if client.userAgent != defaultUserAgent {
		t.Errorf("Expected the user agent %q, got %q", defaultUserAgent, client.userAgent)
	}
}

func TestNewClientWithConfig_RetriesDisabled(t *testing.T) {
	client, err := NewClientWithConfig(Config{Username: "user", Password: "pass", MaxRetries: -1})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.maxRetries != 0 {
		t.Errorf("Expected the retries to be disabled, got %d", client.maxRetries)
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{"empty", "", 0},
		{"seconds", "3", 3 * time.Second},
		{"negative", "-1", 0},
		{"garbage", "later", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRetryAfter(tt.value); got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}
