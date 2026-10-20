package ergani

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testServer is a stand in for the Ergani API. It records the requests it
// received and answers the routes the tests need.
type testServer struct {
	*httptest.Server

	mu       sync.Mutex
	requests []recordedRequest

	authCalls     atomic.Int32
	documentCalls atomic.Int32

	// documentStatus is the status the document route answers with. It is
	// 200 when it is left at zero.
	documentStatus atomic.Int32
	// documentBody is the body the document route answers with.
	documentBody atomic.Value
	// failFirstAnswers makes the document route fail the given number of
	// times before it succeeds.
	failFirstAnswers atomic.Int32
}

// recordedRequest is a request the test server received.
type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Body   string
	Auth   string
}

// newTestServer starts a test server.
func newTestServer(t *testing.T) *testServer {
	t.Helper()

	server := &testServer{}
	server.documentBody.Store(`[{"id":"92","protocol":"ΕΥΣ92","submitDate":"04/05/2022 01:13"}]`)

	mux := http.NewServeMux()
	mux.HandleFunc("/Authentication", func(w http.ResponseWriter, r *http.Request) {
		server.authCalls.Add(1)
		server.record(r)

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"accessToken":"token-1","accessTokenExpired":10800,"refreshToken":"refresh-1","refreshTokenExpired":"2030-05-10T09:43:37.5388855+03:00"}`)); err != nil {
			t.Errorf("Failed to write the authentication response: %v", err)
		}
	})
	mux.HandleFunc("/Authentication/Refresh", func(w http.ResponseWriter, r *http.Request) {
		server.record(r)

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"accessToken":"token-2","accessTokenExpired":10800,"refreshToken":"refresh-2"}`)); err != nil {
			t.Errorf("Failed to write the refresh response: %v", err)
		}
	})
	mux.HandleFunc("/Lookup/Submissions", func(w http.ResponseWriter, r *http.Request) {
		server.record(r)

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`[{"id":82,"code":"WRKCardSE","description":"Δήλωση έναρξης/λήξης εργασίας εργαζομένων"}]`)); err != nil {
			t.Errorf("Failed to write the submission types: %v", err)
		}
	})
	mux.HandleFunc("/WebServices/ServicesList", func(w http.ResponseWriter, r *http.Request) {
		server.record(r)

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`[{"name":"EX_BASE_01","description":"ΣΤΟΙΧΕΙΑ ΕΡΓΟΔΟΤΗ","parameters":[]}]`)); err != nil {
			t.Errorf("Failed to write the services: %v", err)
		}
	})
	mux.HandleFunc("/WebServices/ExecuteService", func(w http.ResponseWriter, r *http.Request) {
		server.record(r)

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"EX_BASE_01":{"Ergodotis":{"Afm":"012345678"}}}`)); err != nil {
			t.Errorf("Failed to write the service result: %v", err)
		}
	})
	mux.HandleFunc("/Documents/", func(w http.ResponseWriter, r *http.Request) {
		server.documentCalls.Add(1)
		server.record(r)

		if server.failFirstAnswers.Load() > 0 {
			server.failFirstAnswers.Add(-1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"message":"boom"}`)); err != nil {
				t.Errorf("Failed to write the error response: %v", err)
			}
			return
		}

		status := int(server.documentStatus.Load())
		if status == 0 {
			status = http.StatusOK
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if _, err := w.Write([]byte(server.documentBody.Load().(string))); err != nil {
			t.Errorf("Failed to write the document response: %v", err)
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server.record(r)
		w.WriteHeader(http.StatusNotFound)
	})

	server.Server = httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server
}

// record remembers a request.
func (s *testServer) record(r *http.Request) {
	body := ""
	if r.Body != nil {
		buf := make([]byte, 0)
		chunk := make([]byte, 4096)
		for {
			n, err := r.Body.Read(chunk)
			buf = append(buf, chunk[:n]...)
			if err != nil {
				break
			}
		}
		body = string(buf)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.requests = append(s.requests, recordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
		Body:   body,
		Auth:   r.Header.Get("Authorization"),
	})
}

// recorded returns the requests the server received.
func (s *testServer) recorded() []recordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]recordedRequest(nil), s.requests...)
}

// newTestClient returns a client that talks to the test server, with a sleep
// that records the delays instead of waiting.
func newTestClient(t *testing.T, server *testServer) (*Client, *[]time.Duration) {
	t.Helper()

	client, err := New("user", "password", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	delays := &[]time.Duration{}
	client.sleep = func(ctx context.Context, d time.Duration) error {
		*delays = append(*delays, d)
		return ctx.Err()
	}

	return client, delays
}

func TestNew_Defaults(t *testing.T) {
	client, err := New("user", "password")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client.BaseURL() != TrialBaseURL {
		t.Errorf("New() base URL = %q, want %q", client.BaseURL(), TrialBaseURL)
	}
	if client.maxRetries != DefaultMaxRetries {
		t.Errorf("New() max retries = %d, want %d", client.maxRetries, DefaultMaxRetries)
	}
	if client.retryBackoff != DefaultRetryBackoff {
		t.Errorf("New() retry backoff = %s, want %s", client.retryBackoff, DefaultRetryBackoff)
	}
	if client.userType != UserTypeErgani {
		t.Errorf("New() user type = %q, want %q", client.userType, UserTypeErgani)
	}
	if client.httpClient == nil {
		t.Error("New() did not create an HTTP client")
	}
}

func TestNew_Options(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second}

	client, err := New("user", "password",
		WithBaseURL(ProductionBaseURL),
		WithHTTPClient(httpClient),
		WithMaxRetries(5),
		WithRetryBackoff(time.Second),
		WithUserAgent("test/1.0"),
		WithUserType(UserTypeExternal),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client.BaseURL() != ProductionBaseURL {
		t.Errorf("New() base URL = %q, want %q", client.BaseURL(), ProductionBaseURL)
	}
	if client.httpClient != httpClient {
		t.Error("New() did not use the given HTTP client")
	}
	if client.maxRetries != 5 {
		t.Errorf("New() max retries = %d, want 5", client.maxRetries)
	}
	if client.retryBackoff != time.Second {
		t.Errorf("New() retry backoff = %s, want 1s", client.retryBackoff)
	}
	if client.userAgent != "test/1.0" {
		t.Errorf("New() user agent = %q, want %q", client.userAgent, "test/1.0")
	}
	if client.userType != UserTypeExternal {
		t.Errorf("New() user type = %q, want %q", client.userType, UserTypeExternal)
	}
}

func TestNew_InvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  Option
	}{
		{"relative base URL", WithBaseURL("/relative")},
		{"nil HTTP client", WithHTTPClient(nil)},
		{"negative timeout", WithTimeout(-time.Second)},
		{"negative retries", WithMaxRetries(-1)},
		{"negative backoff", WithRetryBackoff(-time.Second)},
		{"empty user agent", WithUserAgent("")},
		{"invalid user type", WithUserType(UserType("99"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New("user", "password", tt.opt); err == nil {
				t.Error("New() error = nil, want an error")
			}
		})
	}
}

func TestClient_SubmitWorkCards(t *testing.T) {
	server := newTestServer(t)
	client, _ := newTestClient(t, server)

	submissions, err := client.SubmitWorkCards(context.Background(), []WorkCardDeclaration{
		{
			EmployerTaxID: "111111111",
			BranchNumber:  "0",
			Cards: []WorkCard{
				{
					EmployeeTaxID:    "999999999",
					MovementType:     MovementArrival,
					ReferenceDate:    time.Date(2022, 4, 21, 0, 0, 0, 0, time.UTC),
					MovementDateTime: time.Date(2022, 4, 21, 9, 0, 0, 0, time.UTC),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitWorkCards() error = %v", err)
	}

	if len(submissions) != 1 {
		t.Fatalf("SubmitWorkCards() returned %d submissions, want 1", len(submissions))
	}
	if submissions[0].ID != "92" {
		t.Errorf("SubmitWorkCards() submission ID = %q, want %q", submissions[0].ID, "92")
	}
	if submissions[0].Protocol != "ΕΥΣ92" {
		t.Errorf("SubmitWorkCards() protocol = %q, want %q", submissions[0].Protocol, "ΕΥΣ92")
	}
	if submissions[0].SubmittedAt.IsZero() {
		t.Error("SubmitWorkCards() did not report the moment of the submission")
	}

	requests := server.recorded()
	if len(requests) != 2 {
		t.Fatalf("the server received %d requests, want 2", len(requests))
	}
	if requests[0].Path != "/Authentication" {
		t.Errorf("the first request went to %q, want %q", requests[0].Path, "/Authentication")
	}
	if requests[1].Path != "/Documents/WRKCardSE" {
		t.Errorf("the second request went to %q, want %q", requests[1].Path, "/Documents/WRKCardSE")
	}
	if requests[1].Auth != "Bearer token-1" {
		t.Errorf("the second request carried %q, want %q", requests[1].Auth, "Bearer token-1")
	}
}

func TestClient_RetriesTransientFailures(t *testing.T) {
	server := newTestServer(t)
	server.failFirstAnswers.Store(2)

	client, delays := newTestClient(t, server)

	if _, err := client.SubmitWorkCards(context.Background(), nil); err != nil {
		t.Fatalf("SubmitWorkCards() error = %v", err)
	}

	if got := server.documentCalls.Load(); got != 3 {
		t.Errorf("the document route was called %d times, want 3", got)
	}
	if len(*delays) != 2 {
		t.Fatalf("the client waited %d times, want 2", len(*delays))
	}
	if (*delays)[0] != DefaultRetryBackoff {
		t.Errorf("the first delay was %s, want %s", (*delays)[0], DefaultRetryBackoff)
	}
	if (*delays)[1] != 2*DefaultRetryBackoff {
		t.Errorf("the second delay was %s, want %s", (*delays)[1], 2*DefaultRetryBackoff)
	}
}

func TestClient_RetriesAreBounded(t *testing.T) {
	server := newTestServer(t)
	server.failFirstAnswers.Store(10)

	client, delays := newTestClient(t, server)

	_, err := client.SubmitWorkCards(context.Background(), nil)
	if err == nil {
		t.Fatal("SubmitWorkCards() error = nil, want an error")
	}
	if !errors.Is(err, ErrServer) {
		t.Errorf("SubmitWorkCards() error = %v, want it to match ErrServer", err)
	}
	if got := server.documentCalls.Load(); got != int32(DefaultMaxRetries)+1 {
		t.Errorf("the document route was called %d times, want %d", got, DefaultMaxRetries+1)
	}
	if len(*delays) != DefaultMaxRetries {
		t.Errorf("the client waited %d times, want %d", len(*delays), DefaultMaxRetries)
	}
}

func TestClient_ReauthenticatesAfterUnauthorized(t *testing.T) {
	server := newTestServer(t)

	client, _ := newTestClient(t, server)

	// Authenticate once, so that the client holds a token.
	if _, err := client.Authenticate(context.Background()); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	// The API now rejects the token.
	server.documentStatus.Store(http.StatusUnauthorized)
	if _, err := client.SubmitWorkCards(context.Background(), nil); err == nil {
		t.Fatal("SubmitWorkCards() error = nil, want an error")
	}

	// The client renewed the access token with the refresh token instead of
	// sending the credentials again.
	var refreshed bool
	for _, request := range server.recorded() {
		if request.Path == "/Authentication/Refresh" {
			refreshed = true
		}
	}
	if !refreshed {
		t.Error("the client did not renew the access token")
	}
	if got := server.authCalls.Load(); got != 1 {
		t.Errorf("the authentication route was called %d times, want 1", got)
	}
}

func TestClient_ConcurrentUse(t *testing.T) {
	server := newTestServer(t)
	client, _ := newTestClient(t, server)

	var group sync.WaitGroup
	errs := make(chan error, 50)

	for i := 0; i < 50; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := client.SubmitWorkCards(context.Background(), nil); err != nil {
				errs <- err
			}
		}()
	}

	group.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("SubmitWorkCards() error = %v", err)
	}

	// The token is shared, so the concurrent callers trigger a single
	// authentication rather than one each.
	if got := server.authCalls.Load(); got != 1 {
		t.Errorf("the authentication route was called %d times, want 1", got)
	}
}

func TestClient_ListSubmissions(t *testing.T) {
	server := newTestServer(t)
	client, _ := newTestClient(t, server)

	types, err := client.ListSubmissions(context.Background())
	if err != nil {
		t.Fatalf("ListSubmissions() error = %v", err)
	}

	if len(types) != 1 {
		t.Fatalf("ListSubmissions() returned %d types, want 1", len(types))
	}
	if types[0].Code != CodeWorkCard {
		t.Errorf("ListSubmissions() code = %q, want %q", types[0].Code, CodeWorkCard)
	}
}

func TestClient_Services(t *testing.T) {
	server := newTestServer(t)
	client, _ := newTestClient(t, server)

	services, err := client.Services(context.Background())
	if err != nil {
		t.Fatalf("Services() error = %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("Services() returned %d services, want 1", len(services))
	}
	if services[0].Name != "EX_BASE_01" {
		t.Errorf("Services() name = %q, want %q", services[0].Name, "EX_BASE_01")
	}
}

func TestClient_ExecuteService(t *testing.T) {
	server := newTestServer(t)
	client, _ := newTestClient(t, server)

	result, err := client.ExecuteService(context.Background(), ServiceRequest{
		ServiceCode: "EX_BASE_01",
	})
	if err != nil {
		t.Fatalf("ExecuteService() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("ExecuteService() returned an invalid document: %v", err)
	}

	requests := server.recorded()
	var found bool
	for _, request := range requests {
		if request.Path == "/WebServices/ExecuteService" {
			found = true
			if !json.Valid([]byte(request.Body)) {
				t.Errorf("ExecuteService() sent an invalid body: %s", request.Body)
			}
		}
	}
	if !found {
		t.Error("ExecuteService() did not call the route of the service")
	}
}

func TestClient_ValidationErrors(t *testing.T) {
	server := newTestServer(t)
	client, _ := newTestClient(t, server)

	tests := []struct {
		name  string
		field string
		call  func() error
	}{
		{
			name:  "missing employer tax ID",
			field: "declarations[0].employerTaxID",
			call: func() error {
				_, err := client.SubmitWorkCards(context.Background(), []WorkCardDeclaration{{BranchNumber: "0"}})
				return err
			},
		},
		{
			name:  "invalid movement type",
			field: "declarations[0].cards[0].movementType",
			call: func() error {
				_, err := client.SubmitWorkCards(context.Background(), []WorkCardDeclaration{{
					EmployerTaxID: "111111111",
					BranchNumber:  "0",
					Cards: []WorkCard{{
						EmployeeTaxID:    "999999999",
						MovementType:     MovementType("SLEEP"),
						ReferenceDate:    time.Now(),
						MovementDateTime: time.Now(),
					}},
				}})
				return err
			},
		},
		{
			name:  "missing branch number",
			field: "declarations[0].branchNumber",
			call: func() error {
				_, err := client.SubmitDailySchedules(context.Background(), []DailyScheduleDeclaration{{}})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if err == nil {
				t.Fatal("the call did not fail")
			}

			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("the error is %T, want a *ValidationError", err)
			}
			if validationErr.Field != tt.field {
				t.Errorf("the error points at %q, want %q", validationErr.Field, tt.field)
			}
		})
	}
}

func TestClient_ValidationHappensBeforeTheRequest(t *testing.T) {
	server := newTestServer(t)
	client, _ := newTestClient(t, server)

	if _, err := client.SubmitWorkCards(context.Background(), []WorkCardDeclaration{{}}); err == nil {
		t.Fatal("SubmitWorkCards() error = nil, want an error")
	}

	if got := server.documentCalls.Load(); got != 0 {
		t.Errorf("the document route was called %d times, want 0", got)
	}
}
